package core

import (
	"reflect"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestLineCollectionLineStyleStringToDash(t *testing.T) {
	lc := &LineCollection{
		Collection: Collection{Coords: Coords(CoordData), Alpha: 1},
		Segments:   [][]geom.Pt{{{X: 0, Y: 0}, {X: 5, Y: 5}}},
		Color:      render.Color{A: 1},
		LineWidth:  2,
		LineStyle:  "--",
	}
	r := &recordingRenderer{}
	lc.Draw(r, createTestDrawContext())
	if len(r.pathCalls) != 1 {
		t.Fatalf("got %d path calls, want 1", len(r.pathCalls))
	}
	want := lineStyleToDashes("--", 2)
	if !reflect.DeepEqual(r.pathCalls[0].paint.Dashes, want) {
		t.Fatalf("dashes = %v, want %v", r.pathCalls[0].paint.Dashes, want)
	}
}

func TestLineCollectionExplicitDashesOverrideLineStyle(t *testing.T) {
	lc := &LineCollection{
		Collection: Collection{Coords: Coords(CoordData), Alpha: 1},
		Segments:   [][]geom.Pt{{{X: 0, Y: 0}, {X: 5, Y: 5}}},
		Color:      render.Color{A: 1},
		LineWidth:  2,
		LineStyle:  "--",
		Dashes:     []float64{4, 4},
	}
	r := &recordingRenderer{}
	lc.Draw(r, createTestDrawContext())
	if len(r.pathCalls) != 1 {
		t.Fatalf("got %d path calls, want 1", len(r.pathCalls))
	}
	if !reflect.DeepEqual(r.pathCalls[0].paint.Dashes, []float64{4, 4}) {
		t.Fatalf("explicit dashes should win, got %v", r.pathCalls[0].paint.Dashes)
	}
}

func TestLineCollectionPerItemLineStyles(t *testing.T) {
	lc := &LineCollection{
		Collection: Collection{Coords: Coords(CoordData), Alpha: 1},
		Segments: [][]geom.Pt{
			{{X: 0, Y: 0}, {X: 5, Y: 5}},
			{{X: 0, Y: 1}, {X: 5, Y: 6}},
		},
		Color:      render.Color{A: 1},
		LineWidth:  1,
		LineStyles: []string{"-", ":"},
	}
	r := &recordingRenderer{}
	lc.Draw(r, createTestDrawContext())
	if len(r.pathCalls) != 2 {
		t.Fatalf("got %d path calls, want 2", len(r.pathCalls))
	}
	if r.pathCalls[0].paint.Dashes != nil {
		t.Errorf("solid linestyle should have nil dashes, got %v", r.pathCalls[0].paint.Dashes)
	}
	want := lineStyleToDashes(":", 1)
	if !reflect.DeepEqual(r.pathCalls[1].paint.Dashes, want) {
		t.Fatalf("dotted dashes = %v, want %v", r.pathCalls[1].paint.Dashes, want)
	}
}
