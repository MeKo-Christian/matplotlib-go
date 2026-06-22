package style

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func TestWithXkcdSetsPyplotDefaults(t *testing.T) {
	rc := Apply(Default, WithXkcd())
	want := render.SketchParams{Scale: 1, Length: 100, Randomness: 2}
	if rc.PathSketch != want {
		t.Fatalf("WithXkcd PathSketch = %+v, want %+v", rc.PathSketch, want)
	}
}

func TestWithPathSketchOverridesAndDefaultIsZero(t *testing.T) {
	if Default.PathSketch != (render.SketchParams{}) {
		t.Fatalf("default PathSketch should be zero, got %+v", Default.PathSketch)
	}
	params := render.SketchParams{Scale: 2, Length: 50, Randomness: 3}
	rc := Apply(Default, WithPathSketch(params))
	if rc.PathSketch != params {
		t.Fatalf("WithPathSketch = %+v, want %+v", rc.PathSketch, params)
	}
}

func TestMPLStyleParsesPathSketch(t *testing.T) {
	cases := map[string]render.SketchParams{
		"(1, 100, 2)": {Scale: 1, Length: 100, Randomness: 2},
		"3, 40, 5":    {Scale: 3, Length: 40, Randomness: 5},
		"none":        {},
		"None":        {},
	}
	for value, want := range cases {
		rc, report, err := applyMPLStyleParams(Default, Params{"path.sketch": value})
		if err != nil {
			t.Fatalf("path.sketch=%q: %v", value, err)
		}
		if rc.PathSketch != want {
			t.Fatalf("path.sketch=%q -> %+v, want %+v", value, rc.PathSketch, want)
		}
		if len(report.Unsupported) != 0 {
			t.Fatalf("path.sketch=%q reported unsupported: %+v", value, report.Unsupported)
		}
	}
}

func TestMPLStylePathSketchInvalid(t *testing.T) {
	if _, _, err := applyMPLStyleParams(Default, Params{"path.sketch": "1, 2"}); err == nil {
		t.Fatal("expected error for 2-tuple path.sketch")
	}
}
