package pdf

import (
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestDrawMarkersEmitsReusableFormXObject(t *testing.T) {
	r := newTestRenderer(t)
	drawer, ok := any(r).(render.MarkerDrawer)
	if !ok {
		t.Fatal("PDF renderer should implement render.MarkerDrawer")
	}
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})

	ok = drawer.DrawMarkers(render.MarkerBatch{
		Marker: pdfTestTrianglePath(),
		Items: []render.MarkerItem{
			{
				Offset: geom.Pt{X: 20, Y: 30},
				Paint: render.Paint{
					Fill:      render.Color{R: 1, A: 1},
					Stroke:    render.Color{A: 1},
					LineWidth: 2,
				},
				Antialiased: true,
			},
			{
				Offset: geom.Pt{X: 40, Y: 50},
				Paint: render.Paint{
					Fill:      render.Color{G: 1, A: 1},
					Stroke:    render.Color{A: 1},
					LineWidth: 2,
				},
				Antialiased: true,
			},
		},
	})
	if !ok {
		t.Fatal("DrawMarkers returned false")
	}
	if got := strings.Count(r.content.String(), "/M1 Do"); got != 2 {
		t.Fatalf("expected two marker form invocations, got %d in %q", got, r.content.String())
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	if !pdfDocumentBodyContains(doc, "/XObject << /M1") {
		t.Fatalf("page resources should reference marker form M1; objects: %#v", doc.Objects)
	}
	formBody := pdfDocumentObjectBodyContaining(doc, "/Subtype /Form")
	for _, want := range []string{"/Type /XObject", "/Subtype /Form", "/BBox", " m ", " l ", "B"} {
		if !strings.Contains(formBody, want) {
			t.Fatalf("marker form object missing %q:\n%s", want, formBody)
		}
	}
}

func TestDrawPathCollectionEmitsFormXObjects(t *testing.T) {
	r := newTestRenderer(t)
	drawer, ok := any(r).(render.PathCollectionDrawer)
	if !ok {
		t.Fatal("PDF renderer should implement render.PathCollectionDrawer")
	}
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})

	ok = drawer.DrawPathCollection(render.PathCollectionBatch{Items: []render.PathCollectionItem{
		{
			Path: pdfTestRectPath(10, 10, 30, 25),
			Paint: render.Paint{
				Fill: render.Color{B: 1, A: 1},
			},
			Antialiased: true,
		},
		{
			Path: pdfTestRectPath(40, 10, 65, 25),
			Paint: render.Paint{
				Stroke:    render.Color{R: 1, A: 1},
				LineWidth: 1,
			},
			Antialiased: true,
		},
	}})
	if !ok {
		t.Fatal("DrawPathCollection returned false")
	}
	if !strings.Contains(r.content.String(), "/P1 Do") || !strings.Contains(r.content.String(), "/P2 Do") {
		t.Fatalf("expected path collection form invocations, got %q", r.content.String())
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	if !pdfDocumentBodyContains(doc, "/XObject << /P1") || !pdfDocumentBodyContains(doc, "/P2") {
		t.Fatalf("page resources should reference path collection forms; objects: %#v", doc.Objects)
	}
	if got := pdfDocumentObjectCountContaining(doc, "/Subtype /Form"); got < 2 {
		t.Fatalf("expected at least two form XObjects, got %d; objects: %#v", got, doc.Objects)
	}
}
