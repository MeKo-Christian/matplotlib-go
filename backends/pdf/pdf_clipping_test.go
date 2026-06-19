package pdf

import (
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestRendererClipRectEmitsRectangleClip(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	r.ClipRect(geom.Rect{Min: geom.Pt{X: 5, Y: 5}, Max: geom.Pt{X: 95, Y: 95}})
	raw := r.content.String()
	if !strings.Contains(raw, "5 5 90 90 re W n") {
		t.Errorf("expected rectangle clip operator in %q", raw)
	}
}

func TestRendererClipPathEmitsClipOperators(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: 0})
	p.LineTo(geom.Pt{X: 10, Y: 0})
	p.LineTo(geom.Pt{X: 10, Y: 10})
	p.Close()
	r.ClipPath(p)
	raw := r.content.String()
	if !strings.Contains(raw, "W n\n") {
		t.Errorf("expected clip operator W n in %q", raw)
	}
}
