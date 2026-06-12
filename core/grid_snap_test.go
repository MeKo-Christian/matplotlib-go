package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

// snapPathToPixels ports matplotlib's PathSnapper SNAP_AUTO behavior
// (third_party/matplotlib/src/path_converters.h): paths made solely of
// horizontal/vertical segments snap each vertex to floor(v+0.5)+s, with
// s=0.5 for odd rounded stroke widths and s=0 otherwise.
func TestSnapPathToPixels(t *testing.T) {
	horizontal := func() geom.Path {
		p := geom.Path{}
		p.MoveTo(geom.Pt{X: 10.2, Y: 49.8})
		p.LineTo(geom.Pt{X: 20.7, Y: 49.8})
		p.LineTo(geom.Pt{X: 30.1, Y: 49.8})
		return p
	}

	t.Run("horizontal polyline odd width snaps to half pixel", func(t *testing.T) {
		got := snapPathToPixels(horizontal(), 0.8) // round(0.8)=1, odd
		want := []geom.Pt{{X: 10.5, Y: 50.5}, {X: 21.5, Y: 50.5}, {X: 30.5, Y: 50.5}}
		for i, v := range got.V {
			if v != want[i] {
				t.Fatalf("vertex %d = %v, want %v", i, v, want[i])
			}
		}
	})

	t.Run("even width snaps to integer", func(t *testing.T) {
		got := snapPathToPixels(horizontal(), 2.0)
		want := []geom.Pt{{X: 10, Y: 50}, {X: 21, Y: 50}, {X: 30, Y: 50}}
		for i, v := range got.V {
			if v != want[i] {
				t.Fatalf("vertex %d = %v, want %v", i, v, want[i])
			}
		}
	})

	t.Run("diagonal segment defeats snapping", func(t *testing.T) {
		p := geom.Path{}
		p.MoveTo(geom.Pt{X: 10.2, Y: 49.8})
		p.LineTo(geom.Pt{X: 20.7, Y: 60.3})
		got := snapPathToPixels(p, 0.8)
		if got.V[0] != (geom.Pt{X: 10.2, Y: 49.8}) || got.V[1] != (geom.Pt{X: 20.7, Y: 60.3}) {
			t.Fatalf("diagonal path was modified: %v", got.V)
		}
	})

	t.Run("mixed horizontal and vertical segments snap", func(t *testing.T) {
		p := geom.Path{}
		p.MoveTo(geom.Pt{X: 10.2, Y: 49.8})
		p.LineTo(geom.Pt{X: 20.7, Y: 49.8})
		p.LineTo(geom.Pt{X: 20.7, Y: 60.3})
		got := snapPathToPixels(p, 1.0)
		want := []geom.Pt{{X: 10.5, Y: 50.5}, {X: 21.5, Y: 50.5}, {X: 21.5, Y: 60.5}}
		for i, v := range got.V {
			if v != want[i] {
				t.Fatalf("vertex %d = %v, want %v", i, v, want[i])
			}
		}
	})

	t.Run("curves defeat snapping", func(t *testing.T) {
		p := geom.Path{}
		p.MoveTo(geom.Pt{X: 10.2, Y: 49.8})
		p.QuadTo(geom.Pt{X: 15, Y: 49.8}, geom.Pt{X: 20.7, Y: 49.8})
		got := snapPathToPixels(p, 0.8)
		if got.V[0] != (geom.Pt{X: 10.2, Y: 49.8}) {
			t.Fatalf("curved path was modified: %v", got.V)
		}
	})

	t.Run("vertex sub-1e-4 jitter still snaps", func(t *testing.T) {
		p := geom.Path{}
		p.MoveTo(geom.Pt{X: 10.2, Y: 49.8})
		p.LineTo(geom.Pt{X: 20.7, Y: 49.8 + 5e-5})
		got := snapPathToPixels(p, 0.8)
		want := []geom.Pt{{X: 10.5, Y: 50.5}, {X: 21.5, Y: 50.5}}
		for i, v := range got.V {
			if v != want[i] {
				t.Fatalf("vertex %d = %v, want %v", i, v, want[i])
			}
		}
	})

	t.Run("more than 1024 vertices defeats snapping", func(t *testing.T) {
		p := geom.Path{}
		p.MoveTo(geom.Pt{X: 0.2, Y: 49.8})
		for i := 0; i < 1100; i++ {
			p.LineTo(geom.Pt{X: float64(i) + 1.2, Y: 49.8})
		}
		got := snapPathToPixels(p, 0.8)
		if got.V[0] != (geom.Pt{X: 0.2, Y: 49.8}) {
			t.Fatalf("oversized path was modified")
		}
	})

	t.Run("close command is tolerated", func(t *testing.T) {
		p := geom.Path{}
		p.MoveTo(geom.Pt{X: 10.2, Y: 49.8})
		p.LineTo(geom.Pt{X: 20.7, Y: 49.8})
		p.LineTo(geom.Pt{X: 20.7, Y: 60.3})
		p.Close()
		got := snapPathToPixels(p, 1.0)
		if math.Abs(got.V[0].X-10.5) > 1e-12 {
			t.Fatalf("closed rectilinear path did not snap: %v", got.V)
		}
	})
}
