package svg

import (
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func rectPath() geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 40, Y: 10})
	p.LineTo(geom.Pt{X: 40, Y: 40})
	p.LineTo(geom.Pt{X: 10, Y: 40})
	p.Close()
	return p
}

func filledPaint() *render.Paint {
	return &render.Paint{Fill: render.Color{R: 0, G: 0, B: 1, A: 1}}
}

func TestSVGPathEmitsURLAndGID(t *testing.T) {
	out := renderSVGDocument(t, func(r *Renderer) {
		r.SetURL("https://example.com/a")
		r.SetGID("node-1")
		r.Path(rectPath(), filledPaint())
	})

	if !strings.Contains(out, `<a xlink:href="https://example.com/a">`) {
		t.Fatalf("expected hyperlink wrapper, got:\n%s", out)
	}
	if !strings.Contains(out, `<g id="node-1">`) {
		t.Fatalf("expected id group, got:\n%s", out)
	}
	// The path itself must live inside the anchor and id group.
	aIdx := strings.Index(out, "<a xlink:href")
	pathIdx := strings.Index(out, "<path")
	closeIdx := strings.Index(out, "</a>")
	if aIdx < 0 || aIdx >= pathIdx || pathIdx >= closeIdx {
		t.Fatalf("path not wrapped by anchor; a=%d path=%d /a=%d\n%s", aIdx, pathIdx, closeIdx, out)
	}
}

func TestSVGNoURLNoAnchor(t *testing.T) {
	out := renderSVGDocument(t, func(r *Renderer) {
		r.Path(rectPath(), filledPaint())
	})
	if strings.Contains(out, "<a ") {
		t.Fatalf("did not expect anchor when no url set:\n%s", out)
	}
}

func TestSVGURLEscaped(t *testing.T) {
	out := renderSVGDocument(t, func(r *Renderer) {
		r.SetURL("https://example.com/?a=1&b=2")
		r.Path(rectPath(), filledPaint())
	})
	if !strings.Contains(out, `href="https://example.com/?a=1&amp;b=2"`) {
		t.Fatalf("expected escaped ampersand in href, got:\n%s", out)
	}
	if strings.Contains(out, "a=1&b=2") {
		t.Fatalf("raw ampersand leaked into XML:\n%s", out)
	}
}

// TestSVGURLSnapshotRestore verifies that clearing the marker (as core.drawArtist
// does after a nested child) stops applying the url to later draws.
func TestSVGURLSnapshotRestore(t *testing.T) {
	out := renderSVGDocument(t, func(r *Renderer) {
		r.SetURL("https://outer")
		r.Path(rectPath(), filledPaint()) // linked

		// Simulate a nested child with no url: snapshot, clear, draw, restore.
		prev := r.URL()
		r.SetURL("")
		r.Path(rectPath(), filledPaint()) // not linked
		r.SetURL(prev)

		r.Path(rectPath(), filledPaint()) // linked again
	})

	if got := strings.Count(out, `<a xlink:href="https://outer">`); got != 2 {
		t.Fatalf("expected 2 linked paths, got %d:\n%s", got, out)
	}
}
