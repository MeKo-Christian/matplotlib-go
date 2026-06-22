package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestStemHorizontalOrientationSwapsAxes(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})

	x := []float64{0, 1, 2} // locs along y for horizontal
	y := []float64{3, 4, 5} // heads along x
	baseline := 0.0

	c := ax.Stem(x, y, StemOptions{Orientation: "horizontal", Baseline: &baseline})
	if c == nil {
		t.Fatal("Stem returned nil")
	}

	segs := c.StemLines.Segments
	if len(segs) != len(x) {
		t.Fatalf("got %d segments, want %d", len(segs), len(x))
	}
	for i := range x {
		foot, head := segs[i][0], segs[i][1]
		if foot.X != baseline || foot.Y != x[i] {
			t.Errorf("segment %d foot = %+v, want {X:%v Y:%v}", i, foot, baseline, x[i])
		}
		if head.X != y[i] || head.Y != x[i] {
			t.Errorf("segment %d head = %+v, want {X:%v Y:%v}", i, head, y[i], x[i])
		}
	}

	offs := c.MarkerCollection.Offsets
	for i := range x {
		if offs[i].X != y[i] || offs[i].Y != x[i] {
			t.Errorf("offset %d = %+v, want {X:%v Y:%v}", i, offs[i], y[i], x[i])
		}
	}

	bl := c.Baseline.XY
	if bl[0].X != baseline || bl[1].X != baseline {
		t.Errorf("baseline X = %v,%v, want vertical at %v", bl[0].X, bl[1].X, baseline)
	}
	if bl[0].Y != x[0] || bl[1].Y != x[len(x)-1] {
		t.Errorf("baseline Y span = %v..%v, want %v..%v", bl[0].Y, bl[1].Y, x[0], x[len(x)-1])
	}
}

func TestStemVerticalOrientationDefault(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})

	x := []float64{0, 1}
	y := []float64{2, 3}
	c := ax.Stem(x, y) // default vertical
	if c == nil {
		t.Fatal("Stem returned nil")
	}
	seg := c.StemLines.Segments[1]
	if seg[0].X != x[1] || seg[0].Y != 0 || seg[1].X != x[1] || seg[1].Y != y[1] {
		t.Errorf("vertical stem seg = %+v, want foot {%v,0} head {%v,%v}", seg, x[1], x[1], y[1])
	}
	bl := c.Baseline.XY
	if bl[0].Y != 0 || bl[1].Y != 0 {
		t.Errorf("vertical baseline should be horizontal at Y=0, got %v,%v", bl[0].Y, bl[1].Y)
	}
}

func TestStemInvalidOrientationFallsBackToVertical(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})

	x := []float64{0, 1}
	y := []float64{2, 3}
	c := ax.Stem(x, y, StemOptions{Orientation: "diagonal"})
	if c == nil {
		t.Fatal("Stem returned nil")
	}
	// Same geometry as vertical default.
	seg := c.StemLines.Segments[0]
	if seg[0].X != x[0] || seg[1].Y != y[0] {
		t.Errorf("invalid orientation should fall back to vertical, got seg %+v", seg)
	}
}
