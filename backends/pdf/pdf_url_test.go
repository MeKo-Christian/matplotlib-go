package pdf

import (
	"bytes"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func renderPDFBytes(t *testing.T, draw func(*Renderer)) []byte {
	t.Helper()
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	draw(r)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	return r.document
}

func TestPDFPathEmitsLinkAnnotation(t *testing.T) {
	doc := renderPDFBytes(t, func(r *Renderer) {
		r.SetURL("https://example.com/x")
		r.Path(pdfTestRectPath(10, 10, 40, 40), &render.Paint{Fill: render.Color{R: 0, G: 0, B: 0, A: 1}})
	})

	if !bytes.Contains(doc, []byte("/Subtype /Link")) {
		t.Fatalf("expected /Link annotation in PDF:\n%s", doc)
	}
	if !bytes.Contains(doc, []byte("/URI (https://example.com/x)")) {
		t.Fatalf("expected /URI action with url:\n%s", doc)
	}
	if !bytes.Contains(doc, []byte("/Annots [")) {
		t.Fatalf("expected /Annots array on page:\n%s", doc)
	}
}

func TestPDFNoURLNoAnnotation(t *testing.T) {
	doc := renderPDFBytes(t, func(r *Renderer) {
		r.Path(pdfTestRectPath(10, 10, 40, 40), &render.Paint{Fill: render.Color{R: 0, G: 0, B: 0, A: 1}})
	})
	if bytes.Contains(doc, []byte("/Annots")) {
		t.Fatalf("did not expect /Annots when no url set:\n%s", doc)
	}
	if bytes.Contains(doc, []byte("/Subtype /Link")) {
		t.Fatalf("did not expect link annotation when no url set:\n%s", doc)
	}
}

func TestPDFLinkURLEscaped(t *testing.T) {
	doc := renderPDFBytes(t, func(r *Renderer) {
		r.SetURL("https://example.com/(paren)")
		r.Path(pdfTestRectPath(10, 10, 40, 40), &render.Paint{Fill: render.Color{R: 0, G: 0, B: 0, A: 1}})
	})
	// Parentheses must be backslash-escaped inside a PDF literal string.
	if !bytes.Contains(doc, []byte(`https://example.com/\(paren\)`)) {
		t.Fatalf("expected escaped parens in literal string:\n%s", doc)
	}
}
