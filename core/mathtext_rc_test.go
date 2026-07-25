package core

import (
	"path/filepath"
	"strings"
	"testing"

	mt "github.com/cwbudde/mathtext"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

type mathTextVariableMetricsRenderer struct {
	render.NullRenderer
	width float64
}

func (r *mathTextVariableMetricsRenderer) MeasureText(string, float64, string) render.TextMetrics {
	return render.TextMetrics{W: r.width, H: 10, Ascent: 8, Descent: 2}
}

func withMathTextParams(t *testing.T, params style.Params, fn func()) {
	t.Helper()
	restore, _, err := style.PushContext(params)
	if err != nil {
		t.Fatalf("PushContext() error = %v", err)
	}
	defer restore()
	fn()
}

func mathRunFontBase(t *testing.T, layout MathTextLayout, text string) string {
	t.Helper()
	for _, run := range layout.Runs {
		if run.Text == text {
			return filepath.Base(run.FontKey)
		}
	}
	t.Fatalf("missing MathText run %q in %+v", text, layout.Runs)
	return ""
}

func TestMathTextRCDefaultSelectsImplicitClass(t *testing.T) {
	var renderer textRecordingRenderer
	layoutForDefault := func(class string) MathTextLayout {
		t.Helper()
		var layout MathTextLayout
		withMathTextParams(t, style.Params{
			"mathtext.fontset": "custom",
			"mathtext.default": class,
			"mathtext.rm":      "DejaVu Serif",
			"mathtext.it":      "DejaVu Sans:italic",
		}, func() {
			var ok bool
			layout, ok = LayoutMathText(&renderer, "x", 20, "DejaVu Sans")
			if !ok {
				t.Fatal("LayoutMathText returned !ok")
			}
		})
		return layout
	}

	if got := mathRunFontBase(t, layoutForDefault("rm"), "x"); got != "DejaVuSerif.ttf" {
		t.Fatalf("mathtext.default=rm selected %q, want DejaVuSerif.ttf", got)
	}
	if got := mathRunFontBase(t, layoutForDefault("it"), "x"); got != "DejaVuSans-Oblique.ttf" {
		t.Fatalf("mathtext.default=it selected %q, want DejaVuSans-Oblique.ttf", got)
	}
}

func TestMathTextRCCustomClassPatternsResolveExactFaces(t *testing.T) {
	var renderer textRecordingRenderer
	withMathTextParams(t, style.Params{
		"mathtext.fontset": "custom",
		"mathtext.bf":      "DejaVu Sans:bold",
		"mathtext.bfit":    "DejaVu Serif:italic:bold",
		"mathtext.cal":     "cmsy10",
		"mathtext.it":      "cmmi10",
		"mathtext.rm":      "cmr10",
		"mathtext.sf":      "cmss10",
		"mathtext.tt":      "cmtt10",
	}, func() {
		layout, ok := LayoutMathText(
			&renderer,
			`\mathbf{b}\boldsymbol{i}\mathcal{c}\mathit{x}\mathrm{r}\mathsf{s}\mathtt{t}`,
			20,
			"DejaVu Sans",
		)
		if !ok {
			t.Fatal("LayoutMathText returned !ok")
		}
		want := map[string]string{
			"b": "DejaVuSans-Bold.ttf",
			"i": "DejaVuSerif-BoldItalic.ttf",
			"c": "cmsy10.ttf",
			"x": "cmmi10.ttf",
			"r": "cmr10.ttf",
			"s": "cmss10.ttf",
			"t": "cmtt10.ttf",
		}
		for text, filename := range want {
			if got := mathRunFontBase(t, layout, text); got != filename {
				t.Errorf("class run %q selected %q, want %q", text, got, filename)
			}
		}
	})
}

func TestMathTextRCFontsetsAndFallbackChangeResolvedGlyphs(t *testing.T) {
	var renderer textRecordingRenderer
	type result struct {
		text, font string
	}
	renderOne := func(params style.Params, expr string) result {
		t.Helper()
		var got result
		withMathTextParams(t, params, func() {
			layout, ok := LayoutMathText(&renderer, expr, 20, "DejaVu Sans")
			if !ok || len(layout.Runs) == 0 {
				t.Fatalf("LayoutMathText(%q) = %+v, %v", expr, layout, ok)
			}
			got = result{text: layout.Runs[0].Text, font: filepath.Base(layout.Runs[0].FontKey)}
		})
		return got
	}

	cm := renderOne(style.Params{"mathtext.fontset": "cm"}, `\alpha`)
	stix := renderOne(style.Params{"mathtext.fontset": "stix"}, `\alpha`)
	if cm == stix || cm.font != "cmmi10.ttf" || !strings.HasPrefix(stix.font, "STIXGeneral") {
		t.Fatalf("fontsets did not select distinct glyph profiles: cm=%+v stix=%+v", cm, stix)
	}

	fallback := renderOne(style.Params{
		"mathtext.fontset":  "custom",
		"mathtext.rm":       "DejaVu Sans",
		"mathtext.fallback": "stix",
	}, `\mathbb{R}`)
	noFallback := renderOne(style.Params{
		"mathtext.fontset":  "custom",
		"mathtext.rm":       "DejaVu Sans",
		"mathtext.fallback": "None",
	}, `\mathbb{R}`)
	if fallback.text != "ℝ" || !strings.HasPrefix(fallback.font, "STIXGeneral") {
		t.Fatalf("custom virtual glyph did not use configured STIX fallback: %+v", fallback)
	}
	if noFallback.text != "¤" || noFallback.font != "DejaVuSans.ttf" {
		t.Fatalf("mathtext.fallback=None did not use the custom roman dummy glyph: %+v", noFallback)
	}
}

func TestMathTextExplicitFontPropertiesOverrideRCFontset(t *testing.T) {
	var renderer textRecordingRenderer
	withMathTextParams(t, style.Params{"mathtext.fontset": "cm"}, func() {
		fontKey := render.FontPropertiesKey(render.FontProperties{
			Families:       []string{"DejaVu Sans"},
			MathFontFamily: "dejavuserif",
		})
		layout, ok := LayoutMathText(&renderer, `x+\mathrm{r}`, 20, fontKey)
		if !ok {
			t.Fatal("LayoutMathText returned !ok")
		}
		for _, text := range []string{"x", "r"} {
			if got := mathRunFontBase(t, layout, text); !strings.HasPrefix(got, "DejaVuSerif") {
				t.Errorf("explicit dejavuserif %q run selected %q", text, got)
			}
		}
	})
}

func TestMathTextProfileResolverConcretizesArbitraryGlyphFallbackCandidates(t *testing.T) {
	profile := mt.NewMathFontProfile(mt.MathFontSetCustom)
	profile.Fallback = mt.MathFontSetSTIX
	profile.Fonts[mt.FontClassItalic] = "DejaVu Sans:italic"
	resolver := mathTextProfileResolver{profile: profile}

	candidates := resolver.ResolveMathGlyphCandidates("DejaVu Sans", mt.GlyphRequest{
		Text:   "⫅",
		Symbol: `\subseteqq`,
		Class:  mt.FontClassItalic,
	})
	if len(candidates) < 2 {
		t.Fatalf("fallback candidates = %+v, want custom primary and STIX fallback", candidates)
	}
	if got := filepath.Base(candidates[0].FontKey); got != "DejaVuSans-Oblique.ttf" {
		t.Fatalf("primary candidate = %q, want DejaVuSans-Oblique.ttf", got)
	}
	if got := filepath.Base(candidates[1].FontKey); !strings.HasPrefix(got, "STIXGeneral") {
		t.Fatalf("fallback candidate = %q, want STIXGeneral face", got)
	}
}

func TestResolveMathTextFontPatternPreservesStructuredKeys(t *testing.T) {
	key := render.FontPropertiesKey(render.FontProperties{
		Families: []string{"Definitely Missing Math Face"},
		Style:    render.FontStyleItalic,
		Weight:   700,
	})
	if got := resolveMathTextFontPattern(key); got != key {
		t.Fatalf("structured key changed: got %q want %q", got, key)
	}
}

func TestMathTextMeasurementKeyIncludesEveryRCValue(t *testing.T) {
	base := style.Default.Mathtext
	keys := []func(*style.MathtextRC){
		func(rc *style.MathtextRC) { rc.Fontset = "cm" },
		func(rc *style.MathtextRC) { rc.Default = "rm" },
		func(rc *style.MathtextRC) { rc.Fallback = "stix" },
		func(rc *style.MathtextRC) { rc.BF += "-changed" },
		func(rc *style.MathtextRC) { rc.BFit += "-changed" },
		func(rc *style.MathtextRC) { rc.Cal += "-changed" },
		func(rc *style.MathtextRC) { rc.It += "-changed" },
		func(rc *style.MathtextRC) { rc.RM += "-changed" },
		func(rc *style.MathtextRC) { rc.SF += "-changed" },
		func(rc *style.MathtextRC) { rc.TT += "-changed" },
	}
	want := mathTextMeasurementKey(base)
	for i, mutate := range keys {
		changed := base
		mutate(&changed)
		if got := mathTextMeasurementKey(changed); got == want {
			t.Errorf("field mutation %d did not change measurement key", i)
		}
	}
}

func TestMathTextLayoutCacheSeparatesSameTypeRendererInstances(t *testing.T) {
	narrow := &mathTextVariableMetricsRenderer{width: 7}
	wide := &mathTextVariableMetricsRenderer{width: 29}
	narrowLayout, ok := LayoutMathText(narrow, "q", 12, "DejaVu Sans")
	if !ok {
		t.Fatal("narrow LayoutMathText returned !ok")
	}
	wideLayout, ok := LayoutMathText(wide, "q", 12, "DejaVu Sans")
	if !ok {
		t.Fatal("wide LayoutMathText returned !ok")
	}
	if narrowLayout.Width != 7 || wideLayout.Width != 29 {
		t.Fatalf("same-type renderer metrics collided in layout cache: narrow=%v wide=%v",
			narrowLayout.Width, wideLayout.Width)
	}
	if narrowKey, wideKey := mathTextOptions(narrow, "").MeasurementKey, mathTextOptions(wide, "").MeasurementKey; narrowKey == wideKey {
		t.Fatalf("same-type renderer instances share measurement key %q", narrowKey)
	}
}
