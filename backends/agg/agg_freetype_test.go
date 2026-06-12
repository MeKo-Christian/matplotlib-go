//go:build cgo && !purego

package agg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestUsesDejaVuSansWithoutFallback(t *testing.T) {
	r := mustNew(t, 200, 100)
	if fontReference(r.defaultFontFace) == "" {
		t.Fatal("expected DejaVu Sans font resource to be configured")
	}
	if want := localMatplotlibDejaVuSansPath(); want != "" && r.defaultFontFace.Path != want {
		t.Fatalf("font path = %q, want vendored matplotlib font %q", r.defaultFontFace.Path, want)
	}
	if r.defaultFontFace.Path == "" && len(r.defaultFontFace.Data) == 0 {
		t.Fatalf("default font face has no path or embedded data: %+v", r.defaultFontFace)
	}
	if r.ctx.textForceAuto {
		t.Fatal("expected current x/image raster text path not to enable the unused agg_go force-autohint flag")
	}

	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 200, Y: 100}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	samples := []struct {
		text string
		size float64
		pos  geom.Pt
	}{
		{text: "Text Labels", size: 12, pos: geom.Pt{X: 10, Y: 24}},
		{text: "Group", size: 11.64, pos: geom.Pt{X: 10, Y: 44}},
		{text: "0.0", size: 11.64, pos: geom.Pt{X: 10, Y: 64}},
	}
	for _, sample := range samples {
		metrics := r.MeasureText(sample.text, sample.size, "")
		if metrics.W <= 0 || metrics.Ascent <= 0 || metrics.H <= 0 {
			t.Fatalf("invalid metrics for %q: %+v", sample.text, metrics)
		}
		r.DrawText(sample.text, sample.pos, sample.size, white)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}
	if r.fallback {
		t.Fatal("expected native FreeType text to be used without falling back to GSV")
	}
}

func TestRasterTextWidthTracksRendererDPI(t *testing.T) {
	r := mustNew(t, 200, 100)

	r.SetResolution(72)
	width72 := r.MeasureText("Basic Bars", 12, "").W

	r.SetResolution(96)
	width96 := r.MeasureText("Basic Bars", 12, "").W

	if width72 <= 0 || width96 <= 0 {
		t.Fatalf("expected positive widths, got 72dpi=%v 96dpi=%v", width72, width96)
	}
	if width96 <= width72 {
		t.Fatalf("expected width to increase with DPI, got 72dpi=%v 96dpi=%v", width72, width96)
	}

	gotRatio := width96 / width72
	wantRatio := 96.0 / 72.0
	if math.Abs(gotRatio-wantRatio) > 0.15 {
		t.Fatalf("unexpected DPI scaling ratio: got=%v want=%v", gotRatio, wantRatio)
	}
}

func TestRasterTextKerningMatchesSharedGlyphLayout(t *testing.T) {
	r := mustNew(t, 260, 120)
	fontKey := fontReference(r.defaultFontFace)

	for _, dpi := range []uint{72, 96, 144} {
		r.SetResolution(dpi)
		for _, size := range []float64{12, 24, 72} {
			for _, text := range []string{"Tr", "Te", "To", "Ta", "AV", "WA", "Yo"} {
				t.Run(text, func(t *testing.T) {
					metrics := r.MeasureText(text, size, "")
					layout, ok := render.LayoutTextGlyphs(text, geom.Pt{}, r.fontPixelSize(size), fontKey)
					if !ok {
						t.Fatalf("LayoutTextGlyphs(%q) failed", text)
					}
					// The FreeType build measures with Matplotlib's hinting-factor
					// transform, so it should track but not exactly equal the
					// renderer-neutral sfnt layout.
					if math.Abs(metrics.W-quantize(layout.Advance)) > 0.5 {
						t.Fatalf("MeasureText(%q).W=%v, layout advance=%v glyphs=%+v", text, metrics.W, layout.Advance, layout.Glyphs)
					}
					first := r.MeasureText(text[:1], size, "").W
					second := r.MeasureText(text[1:], size, "").W
					if layout.Glyphs[1].Kern < -0.5 && metrics.W >= first+second-0.5 {
						t.Fatalf("kerned pair %q should be narrower than separate glyphs: pair=%v singles=%v kern=%v", text, metrics.W, first+second, layout.Glyphs[1].Kern)
					}
				})
			}
		}
	}
}

func TestRasterTextBoundsMatchSharedGlyphLayout(t *testing.T) {
	r := mustNew(t, 260, 120)
	r.SetResolution(96)
	fontKey := fontReference(r.defaultFontFace)

	for _, text := range []string{"Tr", "Te", "AV"} {
		t.Run(text, func(t *testing.T) {
			const size = 48.0
			bounds, ok := r.MeasureTextBounds(text, size, "")
			if !ok {
				t.Fatalf("MeasureTextBounds(%q) failed", text)
			}
			layout, ok := render.LayoutTextGlyphs(text, geom.Pt{}, r.fontPixelSize(size), fontKey)
			if !ok {
				t.Fatalf("LayoutTextGlyphs(%q) failed", text)
			}
			if math.Abs(bounds.X-layout.Bounds.X) > 0.5 ||
				math.Abs(bounds.Y-layout.Bounds.Y) > 0.5 ||
				math.Abs(bounds.W-layout.Bounds.W) > 0.5 ||
				math.Abs(bounds.H-layout.Bounds.H) > 0.5 {
				t.Fatalf("bounds mismatch for %q: renderer=%+v layout=%+v", text, bounds, layout.Bounds)
			}
		})
	}
}

func TestKerningMetricsMatchMatplotlibRendererAgg(t *testing.T) {
	r := mustNew(t, 260, 120)
	if r.fontPath == "" {
		t.Skip("Matplotlib comparison needs a path-backed DejaVu Sans font")
	}

	cases := []mplTextMetricCase{}
	for _, dpi := range []uint{72, 96, 144} {
		for _, size := range []float64{12, 24, 72} {
			for _, text := range []string{"Tr", "Te", "To", "Ta", "AV", "WA", "Yo"} {
				cases = append(cases, mplTextMetricCase{Text: text, Size: size, DPI: dpi})
			}
		}
	}

	mplMetrics := runMatplotlibTextMetrics(t, r.fontPath, cases)
	if len(mplMetrics) != len(cases) {
		t.Fatalf("matplotlib returned %d metrics, want %d", len(mplMetrics), len(cases))
	}

	for i, tc := range cases {
		tc := tc
		mpl := mplMetrics[i]
		t.Run(tc.name(), func(t *testing.T) {
			r.SetResolution(tc.DPI)
			goMetrics := r.MeasureText(tc.Text, tc.Size, "")
			goBounds, ok := r.MeasureTextBounds(tc.Text, tc.Size, "")
			if !ok {
				t.Fatalf("MeasureTextBounds(%q) failed", tc.Text)
			}

			goDescent := math.Max(0, goBounds.Y+goBounds.H)
			assertClose(t, "advance", goMetrics.W, mpl.Width, 1.0)
			assertClose(t, "ink height", goBounds.H, mpl.Height, 1.5)
			assertClose(t, "ink descent", goDescent, mpl.Descent, 1.5)
		})
	}
}

func TestAggTextSingleBaselineDiagnostic(t *testing.T) {
	r := mustNew(t, 420, 120)
	if r.fontPath == "" {
		t.Skip("Matplotlib text baseline diagnostic needs a path-backed DejaVu Sans font")
	}
	nativeFreetypeVersion := nativeFreetypeVersion()

	cases := []aggTextBaselineCase{
		{Name: "basic_bars", Text: "Basic Bars", Size: 12, X: 140, Y: 36, DPI: 100, Width: 420, Height: 120},
		{Name: "basic_bars_title_origin", Text: "Basic Bars", Size: 12, X: 276.625, Y: 27.666666666666657, DPI: 100, Width: 640, Height: 360},
		{Name: "zero_tick", Text: "0", Size: 11.64, X: 48, Y: 90, DPI: 100, Width: 160, Height: 120},
	}
	artifactDir := filepath.Join("testdata", "_artifacts", "agg_text_baseline")
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		t.Fatalf("create artifact dir: %v", err)
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			variants := []struct {
				name             string
				goForceAuto      bool
				mplHinting       string
				mplHintingFactor int
				native           bool
				nativeRun        bool
				nativeFactor     int
			}{
				{name: "current_vs_mpl_force_autohint", goForceAuto: false, mplHinting: "force_autohint"},
				{name: "current_vs_mpl_no_hinting", goForceAuto: false, mplHinting: "no_hinting"},
				{name: "forceauto_vs_mpl_force_autohint", goForceAuto: true, mplHinting: "force_autohint"},
				{name: "native_vs_mpl_force_autohint", mplHinting: "force_autohint", native: true},
				{name: "native_run_vs_mpl_force_autohint", mplHinting: "force_autohint", nativeRun: true},
				{name: "native_run_vs_mpl_force_autohint_factor1", mplHinting: "force_autohint", mplHintingFactor: 1, nativeRun: true, nativeFactor: 1},
				{name: "native_run_factor8_vs_mpl_force_autohint", mplHinting: "force_autohint", nativeRun: true, nativeFactor: 8},
			}
			for _, variant := range variants {
				goImg := renderAggTextBaseline(t, r.fontPath, tc, variant.goForceAuto)
				if variant.native {
					goImg = renderNativeFreetypeTextBaseline(t, r.fontPath, tc)
				}
				if variant.nativeRun {
					goImg = renderNativeFreetypeRunTextBaseline(t, r.fontPath, tc, variant.nativeFactor)
				}
				mplImg, mplMetrics, mplFreetypeVersion, mplHintingFactor := runMatplotlibTextBaseline(t, r.fontPath, tc, variant.mplHinting, variant.mplHintingFactor)
				diff := compareRGBA(goImg, mplImg, 0, 0)
				best := bestIntegerOffset(goImg, mplImg, 2)
				goBounds, goCount, _ := darkPixelBounds(goImg)
				mplBounds, mplCount, _ := darkPixelBounds(mplImg)

				prefix := tc.Name + "_" + variant.name
				goPath := filepath.Join(artifactDir, prefix+"_go.png")
				mplPath := filepath.Join(artifactDir, prefix+"_mpl.png")
				diffPath := filepath.Join(artifactDir, prefix+"_diff.png")
				savePNG(t, goPath, goImg)
				savePNG(t, mplPath, mplImg)
				savePNG(t, diffPath, diffImage(goImg, mplImg))

				t.Logf("%s artifacts: go=%s mpl=%s diff=%s", variant.name, goPath, mplPath, diffPath)
				t.Logf("%s matplotlib metrics: width=%.3f height=%.3f descent=%.3f freetype=%s hinting_factor=%d native_freetype=%s native_factor=%d", variant.name, mplMetrics.Width, mplMetrics.Height, mplMetrics.Descent, mplFreetypeVersion, mplHintingFactor, nativeFreetypeVersion, variant.nativeFactor)
				t.Logf("%s dark bounds: go=%v count=%d mpl=%v count=%d", variant.name, goBounds, goCount, mplBounds, mplCount)
				t.Logf("%s direct diff: mean_abs=%.4f max=%d psnr=%.2f differing=%d/%d", variant.name, diff.MeanAbs, diff.MaxAbs, diff.PSNR, diff.DifferingPixels, diff.TotalPixels)
				t.Logf("%s best integer offset within +/-2px: dx=%d dy=%d mean_abs=%.4f max=%d psnr=%.2f differing=%d/%d", variant.name, best.DX, best.DY, best.MeanAbs, best.MaxAbs, best.PSNR, best.DifferingPixels, best.TotalPixels)
			}
		})
	}
}

func TestMeasureTextUsesStableFontLineMetrics(t *testing.T) {
	r := mustNew(t, 200, 100)

	caps := r.MeasureText("H", 24, "")
	descender := r.MeasureText("g", 24, "")
	fontHeights, ok := r.MeasureFontHeights(24, "")
	if !ok {
		t.Fatal("expected renderer to expose font height metrics")
	}

	if caps.Ascent <= 0 || descender.Ascent <= 0 {
		t.Fatalf("expected positive ascent, got caps=%+v descender=%+v", caps, descender)
	}
	if caps.Ascent != descender.Ascent || caps.Descent != descender.Descent {
		t.Fatalf("expected font line metrics to stay stable across strings, got caps=%+v descender=%+v", caps, descender)
	}
	if math.Abs(caps.Ascent-fontHeights.Ascent) > 1e-6 || math.Abs(caps.Descent-fontHeights.Descent) > 1e-6 {
		t.Fatalf("MeasureText should use font height metrics, got caps=%+v font=%+v", caps, fontHeights)
	}
	if fontHeights.LineGap < 0 {
		t.Fatalf("unexpected negative line gap: %+v", fontHeights)
	}
}

type mplTextMetricCase struct {
	Text string  `json:"text"`
	Size float64 `json:"size"`
	DPI  uint    `json:"dpi"`
}

func (c mplTextMetricCase) name() string {
	return c.Text + "_" + strings.TrimRight(strings.TrimRight(jsonFloat(c.Size), "0"), ".") + "pt_" + jsonUint(c.DPI) + "dpi"
}

type mplTextMetric struct {
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	Descent float64 `json:"descent"`
}

type aggTextBaselineCase struct {
	Name   string  `json:"name"`
	Text   string  `json:"text"`
	Size   float64 `json:"size"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	DPI    uint    `json:"dpi"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
}

type mplTextBaselineOutput struct {
	PNG             string        `json:"png"`
	Metrics         mplTextMetric `json:"metrics"`
	FreetypeVersion string        `json:"freetype_version"`
	HintingFactor   int           `json:"hinting_factor"`
}

type imageDiffStats struct {
	DX              int
	DY              int
	MeanAbs         float64
	MaxAbs          uint8
	PSNR            float64
	DifferingPixels int
	TotalPixels     int
}

func runMatplotlibTextMetrics(t *testing.T, fontPath string, cases []mplTextMetricCase) []mplTextMetric {
	t.Helper()

	python, err := matplotlibPythonPath(t)
	if err != nil {
		t.Skipf("Matplotlib Python unavailable; skipping RendererAgg text metric parity test: %v", err)
	}

	payload, err := json.Marshal(cases)
	if err != nil {
		t.Fatalf("marshal metric cases: %v", err)
	}

	script := `
import json
import sys

try:
    import matplotlib
    matplotlib.use("Agg")
    import matplotlib as mpl
    from matplotlib.backends.backend_agg import FigureCanvasAgg
    from matplotlib.figure import Figure
    from matplotlib.font_manager import FontProperties
except Exception as exc:
    print(f"matplotlib import failed: {exc}", file=sys.stderr)
    sys.exit(78)

font_path = sys.argv[1]
cases = json.loads(sys.argv[2])
out = []
with mpl.rc_context({"text.hinting": "no_hinting"}):
    for case in cases:
        fig = Figure(figsize=(2, 2), dpi=case["dpi"])
        canvas = FigureCanvasAgg(fig)
        renderer = canvas.get_renderer()
        prop = FontProperties(fname=font_path, size=case["size"])
        width, height, descent = renderer.get_text_width_height_descent(case["text"], prop, False)
        out.append({"width": width, "height": height, "descent": descent})
print(json.dumps(out))
`

	cmd := exec.Command(python, "-c", script, fontPath, string(payload))
	cmd.Env = append(os.Environ(), "MPLCONFIGDIR="+t.TempDir())
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 78 {
			t.Skip(strings.TrimSpace(string(exitErr.Stderr)))
		}
		t.Fatalf("run Matplotlib RendererAgg text metric helper: %v", err)
	}

	var metrics []mplTextMetric
	if err := json.Unmarshal(out, &metrics); err != nil {
		t.Fatalf("decode Matplotlib text metrics: %v\n%s", err, out)
	}
	return metrics
}

func renderAggTextBaseline(t *testing.T, fontPath string, tc aggTextBaselineCase, forceAuto bool) *image.RGBA {
	t.Helper()

	r := mustNew(t, tc.Width, tc.Height)
	r.SetResolution(tc.DPI)
	r.fontPath = fontPath
	r.defaultFontFace = render.FontFace{Path: fontPath, Family: "DejaVu Sans", Style: render.FontStyleNormal, Weight: 400}
	r.ctx.textForceAuto = forceAuto
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: float64(tc.Width), Y: float64(tc.Height)}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	r.DrawText(tc.Text, geom.Pt{X: tc.X, Y: tc.Y}, tc.Size, render.Color{R: 0, G: 0, B: 0, A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}
	return r.GetImage()
}

func renderNativeFreetypeTextBaseline(t *testing.T, fontPath string, tc aggTextBaselineCase) *image.RGBA {
	t.Helper()

	r := mustNew(t, tc.Width, tc.Height)
	r.SetResolution(tc.DPI)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: float64(tc.Width), Y: float64(tc.Height)}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	ok := r.drawNativeFreetypeText(
		tc.Text,
		render.FontFace{Path: fontPath, Family: "DejaVu Sans"},
		geom.Pt{X: tc.X, Y: tc.Y},
		tc.Size,
		render.Color{R: 0, G: 0, B: 0, A: 1},
	)
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}
	if !ok {
		t.Fatalf("native FreeType diagnostic render failed for %q", tc.Text)
	}
	return r.GetImage()
}

func renderNativeFreetypeRunTextBaseline(t *testing.T, fontPath string, tc aggTextBaselineCase, hintingFactor int) *image.RGBA {
	t.Helper()

	r := mustNew(t, tc.Width, tc.Height)
	r.SetResolution(tc.DPI)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: float64(tc.Width), Y: float64(tc.Height)}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	ok := r.drawNativeFreetypeRunText(
		tc.Text,
		render.FontFace{Path: fontPath, Family: "DejaVu Sans"},
		geom.Pt{X: tc.X, Y: tc.Y},
		tc.Size,
		render.Color{R: 0, G: 0, B: 0, A: 1},
		hintingFactor,
	)
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}
	if !ok {
		t.Fatalf("native FreeType run diagnostic render failed for %q", tc.Text)
	}
	return r.GetImage()
}

func runMatplotlibTextBaseline(t *testing.T, fontPath string, tc aggTextBaselineCase, hinting string, hintingFactor int) (*image.RGBA, mplTextMetric, string, int) {
	t.Helper()

	python, err := matplotlibPythonPath(t)
	if err != nil {
		t.Skipf("Matplotlib Python unavailable; skipping RendererAgg text baseline diagnostic: %v", err)
	}

	payload, err := json.Marshal(tc)
	if err != nil {
		t.Fatalf("marshal baseline case: %v", err)
	}
	rcParams := map[string]any{}
	if hinting != "" {
		rcParams["text.hinting"] = hinting
	}
	if hintingFactor > 0 {
		rcParams["text.hinting_factor"] = hintingFactor
	}
	rcPayload, err := json.Marshal(rcParams)
	if err != nil {
		t.Fatalf("marshal rc params: %v", err)
	}

	script := `
import base64
import io
import json
import sys

try:
    import matplotlib
    matplotlib.use("Agg")
    import matplotlib as mpl
    import matplotlib.ft2font as ft2font
    from matplotlib.backends.backend_agg import FigureCanvasAgg
    from matplotlib.figure import Figure
    from matplotlib.font_manager import FontProperties
    from PIL import Image
except Exception as exc:
    print(f"matplotlib import failed: {exc}", file=sys.stderr)
    sys.exit(78)

font_path = sys.argv[1]
case = json.loads(sys.argv[2])
rc_params = json.loads(sys.argv[3])
with mpl.rc_context(rc_params):
    fig = Figure(figsize=(case["width"] / case["dpi"], case["height"] / case["dpi"]), dpi=case["dpi"])
    fig.patch.set_facecolor("white")
    canvas = FigureCanvasAgg(fig)
    canvas.draw()
    renderer = canvas.get_renderer()
    gc = renderer.new_gc()
    gc.set_foreground((0, 0, 0, 1), isRGBA=True)
    prop = FontProperties(fname=font_path, size=case["size"])
    renderer.draw_text(gc, case["x"], case["y"], case["text"], prop, 0, False)
    width, height, descent = renderer.get_text_width_height_descent(case["text"], prop, False)
    image = Image.frombuffer("RGBA", (case["width"], case["height"]), renderer.buffer_rgba(), "raw", "RGBA", 0, 1)
    buf = io.BytesIO()
    image.save(buf, format="PNG")
    print(json.dumps({
        "png": base64.b64encode(buf.getvalue()).decode("ascii"),
        "metrics": {"width": width, "height": height, "descent": descent},
        "freetype_version": ft2font.__freetype_version__,
        "hinting_factor": mpl.rcParams["text.hinting_factor"],
    }))
`

	cmd := exec.Command(python, "-c", script, fontPath, string(payload), string(rcPayload))
	cmd.Env = append(os.Environ(), "MPLCONFIGDIR="+t.TempDir())
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 78 {
			t.Skip(strings.TrimSpace(string(exitErr.Stderr)))
		}
		t.Fatalf("run Matplotlib RendererAgg text baseline helper: %v", err)
	}

	var result mplTextBaselineOutput
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("decode Matplotlib text baseline output: %v\n%s", err, out)
	}
	pngBytes, err := base64.StdEncoding.DecodeString(result.PNG)
	if err != nil {
		t.Fatalf("decode Matplotlib text baseline PNG: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("decode Matplotlib text baseline image: %v", err)
	}
	rgba := image.NewRGBA(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	return rgba, result.Metrics, result.FreetypeVersion, result.HintingFactor
}

func matplotlibPythonPath(t *testing.T) (string, error) {
	t.Helper()

	candidates := []string{}
	if env := os.Getenv("MATPLOTLIB_GO_PYTHON"); env != "" {
		candidates = append(candidates, env)
	}
	candidates = append(candidates, "/usr/bin/python3")
	if pyPath, err := exec.LookPath("python3"); err == nil {
		candidates = append(candidates, pyPath)
	}

	seen := map[string]bool{}
	for _, candidate := range candidates {
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		cmd := exec.Command(candidate, "-c", "import matplotlib.ft2font")
		cmd.Env = append(os.Environ(), "MPLCONFIGDIR="+t.TempDir())
		if err := cmd.Run(); err == nil {
			return candidate, nil
		}
	}
	return "", errors.New("no Python interpreter with matplotlib found; set MATPLOTLIB_GO_PYTHON")
}

func savePNG(t *testing.T, path string, img image.Image) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
}

func darkPixelBounds(img image.Image) (image.Rectangle, int, bool) {
	bounds := img.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X-1, bounds.Min.Y-1
	count := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r16, g16, b16, a16 := img.At(x, y).RGBA()
			if a16 == 0 {
				continue
			}
			if uint32(r16)+uint32(g16)+uint32(b16) >= 3*0xf000 {
				continue
			}
			count++
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	if count == 0 {
		return image.Rectangle{}, 0, false
	}
	return image.Rect(minX, minY, maxX+1, maxY+1), count, true
}

func bestIntegerOffset(got, want image.Image, radius int) imageDiffStats {
	best := imageDiffStats{MeanAbs: math.Inf(1)}
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			diff := compareRGBA(got, want, dx, dy)
			if diff.MeanAbs < best.MeanAbs {
				best = diff
			}
		}
	}
	return best
}

func compareRGBA(got, want image.Image, dx, dy int) imageDiffStats {
	bounds := got.Bounds()
	wantBounds := want.Bounds()
	if bounds.Dx() != wantBounds.Dx() || bounds.Dy() != wantBounds.Dy() {
		panic(fmt.Sprintf("image size mismatch: got=%v want=%v", bounds, wantBounds))
	}

	var totalAbs float64
	var totalSquared float64
	var maxAbs uint8
	differing := 0
	total := bounds.Dx() * bounds.Dy()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gotR, gotG, gotB := shiftedRGB(got, x, y, dx, dy)
			wantR, wantG, wantB := rgb8(want.At(x, y))
			channels := [3]int{
				int(gotR) - int(wantR),
				int(gotG) - int(wantG),
				int(gotB) - int(wantB),
			}
			pixelDiffers := false
			for _, channelDiff := range channels {
				if channelDiff < 0 {
					channelDiff = -channelDiff
				}
				abs := uint8(channelDiff)
				if abs > maxAbs {
					maxAbs = abs
				}
				if abs > 0 {
					pixelDiffers = true
				}
				absFloat := float64(abs)
				totalAbs += absFloat
				totalSquared += absFloat * absFloat
			}
			if pixelDiffers {
				differing++
			}
		}
	}

	meanAbs := totalAbs / float64(total*3)
	mse := totalSquared / float64(total*3)
	psnr := math.Inf(1)
	if mse > 0 {
		psnr = 20 * math.Log10(255/math.Sqrt(mse))
	}
	return imageDiffStats{
		DX:              dx,
		DY:              dy,
		MeanAbs:         meanAbs,
		MaxAbs:          maxAbs,
		PSNR:            psnr,
		DifferingPixels: differing,
		TotalPixels:     total,
	}
}

func shiftedRGB(img image.Image, x, y, dx, dy int) (uint8, uint8, uint8) {
	bounds := img.Bounds()
	sx, sy := x-dx, y-dy
	if sx < bounds.Min.X || sx >= bounds.Max.X || sy < bounds.Min.Y || sy >= bounds.Max.Y {
		return 255, 255, 255
	}
	return rgb8(img.At(sx, sy))
}

func rgb8(c color.Color) (uint8, uint8, uint8) {
	r16, g16, b16, _ := c.RGBA()
	return uint8(r16 >> 8), uint8(g16 >> 8), uint8(b16 >> 8)
}

func diffImage(got, want image.Image) *image.RGBA {
	bounds := got.Bounds()
	out := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gotR, gotG, gotB := rgb8(got.At(x, y))
			wantR, wantG, wantB := rgb8(want.At(x, y))
			dr := absDiff8(gotR, wantR)
			dg := absDiff8(gotG, wantG)
			db := absDiff8(gotB, wantB)
			out.SetRGBA(x, y, color.RGBA{R: dr, G: dg, B: db, A: 255})
		}
	}
	return out
}

func absDiff8(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func assertClose(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Fatalf("%s mismatch: got=%v want=%v tolerance=%v", name, got, want, tol)
	}
}

func jsonFloat(v float64) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func jsonUint(v uint) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func TestTrailingSpaceDoesNotRenderDuplicateGlyph(t *testing.T) {
	textColor := render.Color{R: 0, G: 0, B: 0, A: 1}

	renderText := func(text string) image.Image {
		r := mustNew(t, 160, 80)
		viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 160, Y: 80}}
		if err := r.Begin(viewport); err != nil {
			t.Fatalf("Begin failed: %v", err)
		}
		r.DrawText(text, geom.Pt{X: 20, Y: 42}, 24, textColor)
		if err := r.End(); err != nil {
			t.Fatalf("End failed: %v", err)
		}
		return r.GetImage()
	}

	withoutTrailingSpace := renderText("x")
	withTrailingSpace := renderText("x ")
	withoutBounds, withoutCount, withoutOK := darkPixelBounds(withoutTrailingSpace)
	withBounds, withCount, withOK := darkPixelBounds(withTrailingSpace)
	if !withoutOK || !withOK {
		t.Fatalf("expected rendered ink: without=%v with=%v", withoutOK, withOK)
	}
	if withoutCount != withCount || withoutBounds.Dx() != withBounds.Dx() || withoutBounds.Dy() != withBounds.Dy() {
		t.Fatalf(
			"expected trailing space to add no ink: without_bounds=%v without_count=%d with_bounds=%v with_count=%d",
			withoutBounds, withoutCount, withBounds, withCount,
		)
	}
}

func TestInternalSpaceDoesNotReplayPreviousGlyph(t *testing.T) {
	textColor := render.Color{R: 0, G: 0, B: 0, A: 1}

	renderText := func(text string) []byte {
		r := mustNew(t, 320, 80)
		viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 320, Y: 80}}
		if err := r.Begin(viewport); err != nil {
			t.Fatalf("Begin failed: %v", err)
		}
		r.DrawText(text, geom.Pt{X: 20, Y: 42}, 24, textColor)
		if err := r.End(); err != nil {
			t.Fatalf("End failed: %v", err)
		}
		return r.GetImage().Pix
	}

	withSingleSpace := append([]byte(nil), renderText("Histogram Strategies")...)
	withoutSpace := append([]byte(nil), renderText("HistogramStrategies")...)
	withDoubleLetter := append([]byte(nil), renderText("HistogrammStrategies")...)

	eqNoSpace := bytes.Equal(withSingleSpace, withoutSpace)
	eqDoubleLetter := bytes.Equal(withSingleSpace, withDoubleLetter)
	if eqNoSpace || eqDoubleLetter {
		t.Fatalf(
			"internal-space rendering collapsed unexpectedly: equals_no_space=%v equals_double_letter=%v",
			eqNoSpace, eqDoubleLetter,
		)
	}
}
