package render

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFontPropertiesHandlesListsAndPaths(t *testing.T) {
	props := ParseFontProperties(`["DejaVu Sans", "Arial", sans-serif]`)
	if len(props.Families) != 3 {
		t.Fatalf("family count = %d, want 3 (%+v)", len(props.Families), props)
	}
	if props.Families[0] != "DejaVu Sans" || props.Families[2] != "sans-serif" {
		t.Fatalf("families = %#v", props.Families)
	}

	path := filepath.Join(t.TempDir(), "ExampleFont.ttf")
	if err := os.WriteFile(path, []byte("not a real font"), 0o644); err != nil {
		t.Fatalf("write font placeholder: %v", err)
	}
	props = ParseFontProperties(path)
	if props.File != path || len(props.Families) != 0 {
		t.Fatalf("path properties = %+v", props)
	}
}

func TestFontManagerResolvesDirectFontFileAndCaches(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ExampleFont.ttf")
	if err := os.WriteFile(path, []byte("not a real font"), 0o644); err != nil {
		t.Fatalf("write font placeholder: %v", err)
	}

	manager := NewFontManager()
	face, ok := manager.FindFont(FontProperties{File: path})
	if !ok || face.Path != path {
		t.Fatalf("FindFont direct path = %+v, %v; want %q", face, ok, path)
	}

	if got := manager.FindFontPath(path); got != path {
		t.Fatalf("FindFontPath direct path = %q, want %q", got, path)
	}
}

func TestFontManagerScansAddedDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Example Sans.ttf")
	if err := os.WriteFile(path, []byte("not a real font"), 0o644); err != nil {
		t.Fatalf("write font placeholder: %v", err)
	}

	manager := NewFontManager()
	manager.AddFontDir(dir)
	face, ok := manager.FindFont(FontProperties{Families: []string{"Example Sans"}})
	if !ok || face.Path != path {
		t.Fatalf("FindFont scanned path = %+v, %v; want %q", face, ok, path)
	}
}

func TestEmbeddedFontFaceProvidesDefaultFamilies(t *testing.T) {
	tests := []struct {
		family string
		want   string
	}{
		{family: "sans-serif", want: "DejaVu Sans"},
		{family: "serif", want: "DejaVu Serif"},
		{family: "monospace", want: "DejaVu Sans Mono"},
	}

	for _, tt := range tests {
		face, ok := embeddedFontFace(tt.family, FontProperties{Style: FontStyleNormal, Weight: 400})
		if !ok {
			t.Fatalf("embeddedFontFace(%q) failed", tt.family)
		}
		if face.Family != tt.want {
			t.Fatalf("embeddedFontFace(%q).Family = %q, want %q", tt.family, face.Family, tt.want)
		}
		if len(face.Data) == 0 {
			t.Fatalf("embeddedFontFace(%q) returned no data", tt.family)
		}
		if face.Path != "" {
			t.Fatalf("embeddedFontFace(%q).Path = %q, want empty", tt.family, face.Path)
		}
	}
}

func TestEmbeddedFontFaceDoesNotSatisfyStyledRequests(t *testing.T) {
	if face, ok := embeddedFontFace("DejaVu Sans", FontProperties{Style: FontStyleItalic, Weight: 400}); ok {
		t.Fatalf("embedded regular face satisfied italic request: %+v", face)
	}
	if face, ok := embeddedFontFace("DejaVu Sans", FontProperties{Style: FontStyleNormal, Weight: 700}); ok {
		t.Fatalf("embedded regular face satisfied bold request: %+v", face)
	}
}

func TestFontStyleMatchesRequestedRejectsRegularForItalic(t *testing.T) {
	if fontStyleMatchesRequested("Book", FontStyleItalic) {
		t.Fatal("regular Book style should not satisfy italic request")
	}
	if !fontStyleMatchesRequested("Oblique", FontStyleItalic) {
		t.Fatal("oblique style should satisfy italic request")
	}
	if !fontStyleMatchesRequested("Italic", FontStyleOblique) {
		t.Fatal("italic style should satisfy oblique request")
	}
}

func TestCSSFontFamilyVariants(t *testing.T) {
	tests := map[string]string{
		"serif":       "DejaVu Serif, serif",
		"sans-serif":  "DejaVu Sans, Arial, sans-serif",
		"monospace":   "DejaVu Sans Mono, monospace",
		"mono_space":  "DejaVu Sans Mono, monospace",
		"custom-font": "DejaVu Sans, Arial, sans-serif",
	}

	for key, want := range tests {
		if got := CSSFontFamily(key); got != want {
			t.Fatalf("CSSFontFamily(%q) = %q, want %q", key, got, want)
		}
	}
}
