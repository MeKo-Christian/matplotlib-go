package style

import (
	"reflect"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func TestMPLFontRCParsesPropertiesAndOrderedFallbacks(t *testing.T) {
	theme, report, err := ParseMPLStyle("fonts", `
font.family: serif, Custom Face
font.serif: First Serif, Second Serif, serif
font.sans-serif: First Sans, sans-serif
font.cursive: First Script, cursive
font.fantasy: First Fantasy, fantasy
font.monospace: First Mono, monospace
font.style: italic
font.variant: small-caps
font.weight: semibold
font.stretch: condensed
`)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unsupported entries = %#v", report.Unsupported)
	}

	rc := theme.RC
	if got, want := rc.Font.Family, []string{"serif", "Custom Face"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("font family = %#v, want %#v", got, want)
	}
	props := render.ParseFontProperties(rc.FontKey)
	if got, want := props.Families, []string{"First Serif", "Second Serif", "serif", "Custom Face"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("resolved families = %#v, want %#v", got, want)
	}
	if props.Style != render.FontStyleItalic || props.Variant != "small-caps" ||
		props.Weight != 600 || props.Stretch != "condensed" {
		t.Fatalf("resolved properties = %#v", props)
	}
}

func TestMPLFontRCRuntimeRoundTrip(t *testing.T) {
	ResetDefaults()
	t.Cleanup(ResetDefaults)

	_, err := UpdateParams(Params{
		"font.family":     "monospace, Backup",
		"font.monospace":  "Mono One, Mono Two",
		"font.style":      "oblique",
		"font.variant":    "small-caps",
		"font.weight":     "700",
		"font.stretch":    "semi-expanded",
		"font.serif":      "Serif One, Serif Two",
		"font.sans-serif": "Sans One, Sans Two",
		"font.cursive":    "Script One, Script Two",
		"font.fantasy":    "Fantasy One, Fantasy Two",
	})
	if err != nil {
		t.Fatal(err)
	}
	params := CurrentParams()
	for key, want := range (Params{
		"font.family":     "monospace, Backup",
		"font.monospace":  "Mono One, Mono Two",
		"font.style":      "oblique",
		"font.variant":    "small-caps",
		"font.weight":     "700",
		"font.stretch":    "semi-expanded",
		"font.serif":      "Serif One, Serif Two",
		"font.sans-serif": "Sans One, Sans Two",
		"font.cursive":    "Script One, Script Two",
		"font.fantasy":    "Fantasy One, Fantasy Two",
	}) {
		if got := params[key]; got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestMPLFontRCRejectsInvalidProperties(t *testing.T) {
	for _, src := range []string{
		"font.family: []",
		"font.style: sideways",
		"font.variant: petite-caps",
		"font.weight: 2000",
		"font.stretch: elastic",
	} {
		if _, _, err := ParseMPLStyle("bad-font", src); err == nil {
			t.Errorf("ParseMPLStyle(%q) unexpectedly succeeded", src)
		}
	}
}

func TestFontRCCopyDoesNotAliasFallbackLists(t *testing.T) {
	rc := Apply(Default)
	rc.Font.SansSerif[0] = "Changed"
	if Default.Font.SansSerif[0] == "Changed" {
		t.Fatal("Apply aliased the default sans-serif fallback list")
	}
}
