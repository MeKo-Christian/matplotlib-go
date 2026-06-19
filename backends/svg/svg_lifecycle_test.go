package svg

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestNewInvalidDimensions(t *testing.T) {
	r, err := New(0, 10, render.Color{})
	if err == nil || r != nil {
		t.Fatal("expected error for non-positive dimensions")
	}
}

func TestSerializationDeterministic(t *testing.T) {
	draw := func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 10, Y: 10})
		path.LineTo(geom.Pt{X: 170, Y: 110})
		path.Close()
		r.Path(path, &render.Paint{
			Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
			Fill:      render.Color{R: 0.5, G: 0.25, B: 0.1, A: 0.5},
			LineWidth: 2,
		})
		r.DrawText("hello", geom.Pt{X: 20, Y: 30}, 14, render.Color{R: 0, G: 0, B: 0, A: 1})
		r.DrawTextRotated("rot", geom.Pt{X: 80, Y: 60}, 12, 0.5, render.Color{R: 0, G: 0, B: 0, A: 1})
	}

	first := renderSVGDocument(t, draw)
	second := renderSVGDocument(t, draw)

	if first != second {
		t.Fatalf("SVG output not deterministic across renders:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestSaveSVG(t *testing.T) {
	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	var path geom.Path
	path.MoveTo(geom.Pt{X: 10, Y: 10})
	path.LineTo(geom.Pt{X: 170, Y: 110})
	r.Path(path, &render.Paint{
		Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth: 2,
	})

	r.DrawText("line", geom.Pt{X: 20, Y: 30}, 14, render.Color{R: 0, G: 0, B: 0, A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	tmp, err := os.CreateTemp("", "matplotlib-go-svg-*.svg")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	t.Cleanup(func() { _ = os.Remove(tmpPath) })

	if err := r.SaveSVG(tmpPath); err != nil {
		t.Fatalf("SaveSVG failed: %v", err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "<svg") || !strings.Contains(content, "</svg>") {
		t.Fatal("SVG output missing root element")
	}
	if !strings.Contains(content, "<path") {
		t.Fatal("SVG output missing path node")
	}
	if !strings.Contains(content, "<text") || !strings.Contains(content, ">line<") {
		t.Fatal("SVG output missing text node")
	}
}

func TestRasterizedArtistEmbedsImageWhileKeepingTextVector(t *testing.T) {
	fig := core.NewFigure(180, 120)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.2, Y: 0.2}, Max: geom.Pt{X: 0.8, Y: 0.75}})
	line := ax.Plot([]float64{0, 0.5, 1}, []float64{0, 1, 0})
	line.SetRasterized(true)
	ax.SetTitle("Vector title")

	r := mustNewRenderer(t)
	core.DrawFigure(fig, r)

	tmp, err := os.CreateTemp("", "matplotlib-go-mixed-svg-*.svg")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	t.Cleanup(func() { _ = os.Remove(tmpPath) })

	if err := r.SaveSVG(tmpPath); err != nil {
		t.Fatalf("SaveSVG failed: %v", err)
	}
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "<image") {
		t.Fatalf("rasterized artist did not emit an SVG image: %s", content)
	}
	if !strings.Contains(content, `<g clip-path="url(#`) || !strings.Contains(content, `><image`) {
		t.Fatalf("rasterized image should preserve the active axes clip: %s", content)
	}
	if !strings.Contains(content, "Vector title") {
		t.Fatalf("surrounding title text was not preserved as vector SVG text: %s", content)
	}
}

func TestPathEffectFilterEmitsNativeSVGFilter(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 20, Y: 70})
		path.LineTo(geom.Pt{X: 160, Y: 70})
		r.Path(path, &render.Paint{
			Stroke:    render.Color{B: 1, A: 1},
			LineWidth: 2,
			PathEffects: []render.PathEffect{
				render.FilterPathEffect(
					render.Color{},
					render.Color{R: 1, A: 0.6},
					8,
					"blur",
					3,
					geom.Pt{X: 2, Y: 2},
				),
				render.NormalPathEffect(),
			},
		})
	})

	if strings.Contains(content, "<image") {
		t.Fatalf("native SVG filter path effect should stay vector, got %q", content)
	}
	for _, want := range []string{
		`<filter id="filter1"`,
		`<feGaussianBlur stdDeviation="3"`,
		`filter="url(#filter1)"`,
		`stroke="rgb(255,0,0)"`,
		`stroke="rgb(0,0,255)"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected SVG fragment %q in %q", want, content)
		}
	}
}

func TestRenderSVGEmitsDefaultEmptyMetadata(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.DrawText("plain", geom.Pt{X: 10, Y: 20}, 12, render.Color{A: 1})
	})

	if !strings.Contains(content, "  <metadata></metadata>\n") {
		t.Fatalf("default SVG should contain an empty metadata block, got %q", content)
	}
	if strings.Contains(content, `name="Date"`) {
		t.Fatalf("default SVG metadata should not include a generated date, got %q", content)
	}
}

func TestRenderSVGMetadataOptionAndSourceDateEpoch(t *testing.T) {
	t.Setenv("SOURCE_DATE_EPOCH", "0")
	content := renderSVGDocument(t, func(r *Renderer) {
		r.SetSVGOptions(render.ResolveSVGOptions(
			render.WithSVGMetadata(map[string]string{
				"Title":  "Plot",
				"Author": "Codex",
			}),
		))
	})

	for _, want := range []string{
		`<meta name="Author" content="Codex" />`,
		`<meta name="Date" content="1970-01-01T00:00:00Z" />`,
		`<meta name="Title" content="Plot" />`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected metadata fragment %q in %q", want, content)
		}
	}
	if strings.Index(content, `name="Author"`) > strings.Index(content, `name="Title"`) {
		t.Fatalf("metadata entries should serialize in deterministic key order, got %q", content)
	}
}

func TestHashSaltUsesContentDerivedIDs(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.SetSVGOptions(render.ResolveSVGOptions(render.WithSVGHashSalt("stable")))
		r.ClipRect(geom.Rect{Min: geom.Pt{X: 1, Y: 2}, Max: geom.Pt{X: 11, Y: 12}})
		r.DrawText("clipped", geom.Pt{X: 5, Y: 6}, 12, render.Color{A: 1})
	})

	if strings.Contains(content, `id="clip1"`) || strings.Contains(content, `url(#clip1)`) {
		t.Fatalf("hash-salted SVG should not use sequential clip IDs, got %q", content)
	}
	re := regexp.MustCompile(`id="clip[0-9a-f]{10}"`)
	if !re.MatchString(content) {
		t.Fatalf("hash-salted clip ID should use a stable content hash, got %q", content)
	}
}

func TestSetResolutionIgnoresZeroAndStoresPositiveDPI(t *testing.T) {
	r := mustNewRenderer(t)

	r.SetResolution(0)
	if got := r.resolution; got != 72 {
		t.Fatalf("zero DPI should be ignored, got %d", got)
	}

	r.SetResolution(144)
	if got := r.resolution; got != 144 {
		t.Fatalf("positive DPI should be stored, got %d", got)
	}
}

func TestHelperFormattingBranches(t *testing.T) {
	if got := dashedArray([]float64{5}); got != "" {
		t.Fatalf("single dash segment should not emit dash array, got %q", got)
	}
	if got := dashedArray([]float64{5, 2, 9}); got != "5,2" {
		t.Fatalf("odd dash lists should ignore trailing value, got %q", got)
	}

	if got := mapLineJoin(render.JoinBevel); got != "bevel" {
		t.Fatalf("expected bevel join mapping, got %q", got)
	}
	if got := mapLineJoin(render.LineJoin(99)); got != "miter" {
		t.Fatalf("unknown join should fall back to miter, got %q", got)
	}
	if got := mapLineCap(render.CapRound); got != "round" {
		t.Fatalf("expected round cap mapping, got %q", got)
	}
	if got := mapLineCap(render.LineCap(99)); got != "butt" {
		t.Fatalf("unknown cap should fall back to butt, got %q", got)
	}

	if got := clamp01(-0.1); got != 0 {
		t.Fatalf("negative colors should clamp to 0, got %v", got)
	}
	if got := clamp01(1.1); got != 1 {
		t.Fatalf("oversaturated colors should clamp to 1, got %v", got)
	}
	if got := clampFloat(1.25); got != 1.25 {
		t.Fatalf("finite float should pass through, got %v", got)
	}
}

func TestSaveSVGErrorPaths(t *testing.T) {
	r := mustNewRenderer(t)

	if err := r.SaveSVG(""); err == nil {
		t.Fatal("empty output path should return an error")
	}

	dir := t.TempDir()
	if err := r.SaveSVG(filepath.Join(dir, ".")); err == nil {
		t.Fatal("writing SVG to a directory should return an error")
	}
}
