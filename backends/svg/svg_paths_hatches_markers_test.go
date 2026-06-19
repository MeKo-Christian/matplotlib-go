package svg

import (
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestPathSerializesStrokeFillOpacityAndDashes(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 10, Y: 15})
		path.LineTo(geom.Pt{X: 60, Y: 15})
		path.LineTo(geom.Pt{X: 60, Y: 45})
		path.Close()

		r.Path(path, &render.Paint{
			LineWidth:  2.5,
			LineJoin:   render.JoinRound,
			LineCap:    render.CapSquare,
			MiterLimit: 7,
			Stroke:     render.Color{R: 1, G: 0, B: 0, A: 0.25},
			Fill:       render.Color{G: 1, A: 0.5},
			Dashes:     []float64{4, 2, 1, 3},
		})
	})

	for _, want := range []string{
		`fill="rgb(0,255,0)"`,
		`fill-opacity="0.5"`,
		`stroke="rgb(255,0,0)"`,
		`stroke-opacity="0.25"`,
		`stroke-width="2.5"`,
		`stroke-linejoin="round"`,
		`stroke-linecap="square"`,
		`stroke-miterlimit="7"`,
		`stroke-dasharray="4,2,1,3"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected serialized path attribute %q in %q", want, content)
		}
	}
}

func TestDrawMarkersEmitsDefAndUseElements(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		ok := r.DrawMarkers(render.MarkerBatch{
			Marker: circleMarkerPath(),
			Items: []render.MarkerItem{
				{
					Offset:    geom.Pt{X: 10, Y: 20},
					Transform: geom.Identity(),
					Paint:     render.Paint{Fill: render.Color{R: 1, A: 1}, Stroke: render.Color{A: 1}, LineWidth: 1},
				},
				{
					Offset:    geom.Pt{X: 50, Y: 60},
					Transform: geom.Identity(),
					Paint:     render.Paint{Fill: render.Color{B: 1, A: 1}, Stroke: render.Color{A: 1}, LineWidth: 1},
				},
			},
		})
		if !ok {
			t.Fatal("DrawMarkers returned false")
		}
	})

	if !strings.Contains(content, `<path id="marker1" d="M -1 0 L 1 0 L 0 1 Z" />`) {
		t.Fatalf("expected marker path def, got %q", content)
	}
	if got := strings.Count(content, `<use href="#marker1"`); got != 2 {
		t.Fatalf("expected 2 use references, got %d in %q", got, content)
	}
	if !strings.Contains(content, `transform="matrix(1 0 0 -1 10 100)"`) {
		t.Fatalf("expected first item transform, got %q", content)
	}
	if !strings.Contains(content, `transform="matrix(1 0 0 -1 50 60)"`) {
		t.Fatalf("expected second item transform, got %q", content)
	}
	if !strings.Contains(content, `fill="rgb(255,0,0)"`) || !strings.Contains(content, `fill="rgb(0,0,255)"`) {
		t.Fatalf("expected per-item fills, got %q", content)
	}
}

func TestDrawMarkersDedupesIdenticalMarkerGeometry(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		batch := render.MarkerBatch{
			Marker: circleMarkerPath(),
			Items: []render.MarkerItem{
				{
					Offset: geom.Pt{X: 0, Y: 0}, Transform: geom.Identity(),
					Paint: render.Paint{Fill: render.Color{A: 1}},
				},
			},
		}
		r.DrawMarkers(batch)
		r.DrawMarkers(batch)
	})

	if got := strings.Count(content, `<path id="marker`); got != 1 {
		t.Fatalf("identical marker geometry should share one def, got %d defs in %q", got, content)
	}
	if got := strings.Count(content, `<use href="#marker1"`); got != 2 {
		t.Fatalf("expected 2 use references sharing marker1, got %d in %q", got, content)
	}
}

func TestDrawMarkersHonorsActiveClip(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.ClipRect(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}})
		r.DrawMarkers(render.MarkerBatch{
			Marker: circleMarkerPath(),
			Items: []render.MarkerItem{
				{
					Offset: geom.Pt{X: 50, Y: 50}, Transform: geom.Identity(),
					Paint: render.Paint{Fill: render.Color{A: 1}},
				},
			},
		})
	})

	if !strings.Contains(content, `<g clip-path="url(#clip1)"><use href="#marker1"`) {
		t.Fatalf("markers should be wrapped in the active clip group, got %q", content)
	}
}

func TestDrawMarkersRejectsEmptyBatchOrInvalidMarker(t *testing.T) {
	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	if ok := r.DrawMarkers(render.MarkerBatch{}); ok {
		t.Error("DrawMarkers should return false for empty batch")
	}
	if ok := r.DrawMarkers(render.MarkerBatch{
		Marker: circleMarkerPath(),
		Items:  nil,
	}); ok {
		t.Error("DrawMarkers should return false when Items is empty")
	}

	// Items with no fill and no stroke should produce no output but still
	// register the marker def (caller asked for a marker batch with that
	// geometry) — implementation returns true to signal it took ownership.
	ok := r.DrawMarkers(render.MarkerBatch{
		Marker: circleMarkerPath(),
		Items: []render.MarkerItem{
			{Offset: geom.Pt{}, Transform: geom.Identity(), Paint: render.Paint{}},
		},
	})
	if !ok {
		t.Error("DrawMarkers should return true even when all items are invisible")
	}

	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}
	content := r.renderSVG()
	if strings.Contains(content, "<use ") {
		t.Fatalf("invisible markers should not emit use elements, got %q", content)
	}
}

func TestDrawPathCollectionEmitsDefsAndUseElements(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		itemPath := circleMarkerPath()
		ok := r.DrawPathCollection(render.PathCollectionBatch{
			Items: []render.PathCollectionItem{
				{
					Path:  itemPath,
					Paint: render.Paint{Fill: render.Color{R: 1, A: 1}, Stroke: render.Color{A: 1}, LineWidth: 1},
				},
				{
					Path:  itemPath,
					Paint: render.Paint{Fill: render.Color{B: 1, A: 1}, Stroke: render.Color{A: 1}, LineWidth: 1},
				},
			},
		})
		if !ok {
			t.Fatal("DrawPathCollection returned false")
		}
	})

	if !strings.Contains(content, `<path id="pathcoll1" d="M -1 120 L 1 120 L 0 119 Z" />`) {
		t.Fatalf("expected path collection path def, got %q", content)
	}
	if got := strings.Count(content, `<use href="#pathcoll1"`); got != 2 {
		t.Fatalf("expected 2 path collection use references, got %d in %q", got, content)
	}
	if !strings.Contains(content, `fill="rgb(255,0,0)"`) || !strings.Contains(content, `fill="rgb(0,0,255)"`) {
		t.Fatalf("expected per-item fills, got %q", content)
	}
}

func TestDrawPathCollectionHonorsActiveClipAndRejectsEmptyBatch(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		if ok := r.DrawPathCollection(render.PathCollectionBatch{}); ok {
			t.Fatal("DrawPathCollection should reject an empty batch")
		}
		r.ClipRect(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}})
		ok := r.DrawPathCollection(render.PathCollectionBatch{
			Items: []render.PathCollectionItem{
				{
					Path:  circleMarkerPath(),
					Paint: render.Paint{Fill: render.Color{A: 1}},
				},
			},
		})
		if !ok {
			t.Fatal("DrawPathCollection returned false")
		}
	})

	if !strings.Contains(content, `<g clip-path="url(#clip1)"><use href="#pathcoll1"`) {
		t.Fatalf("path collection should be wrapped in the active clip group, got %q", content)
	}
}

func TestPathWithHatchEmitsPatternFill(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 10, Y: 10})
		path.LineTo(geom.Pt{X: 70, Y: 10})
		path.LineTo(geom.Pt{X: 70, Y: 50})
		path.Close()
		r.Path(path, &render.Paint{
			Fill:           render.Color{G: 1, A: 0.5},
			Hatch:          "/",
			HatchColor:     render.Color{R: 1, A: 0.75},
			HatchLineWidth: 2,
			HatchSpacing:   12,
		})
	})

	for _, want := range []string{
		`<pattern id="hatch1" patternUnits="userSpaceOnUse" width="72" height="72">`,
		`<rect x="0" y="0" width="72" height="72" fill="rgb(0,255,0)" fill-opacity="0.5" />`,
		`stroke="rgb(255,0,0)"`,
		`stroke-opacity="0.75"`,
		`stroke-width="2"`,
		`fill="url(#hatch1)"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected hatch pattern fragment %q in %q", want, content)
		}
	}
}

func TestPathWithShapeHatchEmitsPatternGeometry(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 10, Y: 10})
		path.LineTo(geom.Pt{X: 70, Y: 10})
		path.LineTo(geom.Pt{X: 70, Y: 50})
		path.Close()
		r.Path(path, &render.Paint{
			Hatch:          "oO.*",
			HatchColor:     render.Color{A: 1},
			HatchLineWidth: 1,
			HatchSpacing:   12,
		})
	})

	for _, want := range []string{
		`<pattern id="hatch1" patternUnits="userSpaceOnUse" width="72" height="72">`,
		` C `,
		` Z`,
		`fill="rgb(0,0,0)"`,
		`stroke="rgb(0,0,0)"`,
		`fill="url(#hatch1)"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected shape hatch pattern fragment %q in %q", want, content)
		}
	}
}

func TestHatchPatternDefsAreReused(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 0, Y: 0})
		path.LineTo(geom.Pt{X: 10, Y: 0})
		path.LineTo(geom.Pt{X: 10, Y: 10})
		path.Close()
		paint := &render.Paint{
			Fill:           render.Color{A: 1},
			Hatch:          "x",
			HatchColor:     render.Color{A: 1},
			HatchLineWidth: 1,
		}
		r.Path(path, paint)
		r.Path(path, paint)
	})

	if got := strings.Count(content, `<pattern id="hatch`); got != 1 {
		t.Fatalf("identical hatch metadata should share one pattern def, got %d in %q", got, content)
	}
	if got := strings.Count(content, `fill="url(#hatch1)"`); got != 2 {
		t.Fatalf("both paths should reference shared hatch pattern, got %d refs in %q", got, content)
	}
}

func TestPathForcedAlphaUsesElementOpacity(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 0, Y: 0})
		path.LineTo(geom.Pt{X: 10, Y: 0})
		path.LineTo(geom.Pt{X: 10, Y: 10})
		path.Close()
		r.Path(path, &render.Paint{
			Fill:       render.Color{R: 1, A: 0.2},
			Stroke:     render.Color{B: 1, A: 0.2},
			LineWidth:  1,
			ForceAlpha: true,
			Alpha:      0.2,
		})
	})

	if !strings.Contains(content, `opacity="0.2"`) {
		t.Fatalf("forced-alpha path should carry element opacity, got %q", content)
	}
	if strings.Contains(content, `fill-opacity=`) || strings.Contains(content, `stroke-opacity=`) {
		t.Fatalf("forced-alpha path should not double-apply per-color opacity, got %q", content)
	}
}

func TestSupportsNativeHatch(t *testing.T) {
	r := mustNewRenderer(t)
	if !r.SupportsNativeHatch() {
		t.Fatal("SVG renderer should advertise native hatch consumption")
	}
}

func TestBuildPathDataSupportsQuadraticCubicAndClose(t *testing.T) {
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.QuadTo, geom.CubicTo, geom.ClosePath},
		V: []geom.Pt{
			{X: 1, Y: 2},
			{X: 3, Y: 4},
			{X: 5, Y: 6},
			{X: 7, Y: 8},
			{X: 9, Y: 10},
			{X: 11, Y: 12},
		},
	}

	got := buildPathData(path)
	want := "M 1 2 Q 3 4 5 6 C 7 8 9 10 11 12 Z"
	if got != want {
		t.Fatalf("unexpected path data:\nwant %q\ngot  %q", want, got)
	}
}

func TestBuildPathDataRejectsMalformedCommands(t *testing.T) {
	tests := []geom.Path{
		{C: []geom.Cmd{geom.MoveTo}},
		{C: []geom.Cmd{geom.LineTo}},
		{C: []geom.Cmd{geom.QuadTo}, V: []geom.Pt{{X: 1, Y: 2}}},
		{C: []geom.Cmd{geom.CubicTo}, V: []geom.Pt{{X: 1, Y: 2}, {X: 3, Y: 4}}},
		{C: []geom.Cmd{geom.Cmd(99)}, V: []geom.Pt{{X: 1, Y: 2}}},
	}

	for _, path := range tests {
		if got := buildPathData(path); got != "" {
			t.Fatalf("malformed path should serialize to empty data, got %q for %+v", got, path)
		}
	}
}
