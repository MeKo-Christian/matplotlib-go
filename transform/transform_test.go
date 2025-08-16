package transform

import (
	"math"
	"math/rand"
	"testing"

	"matplotlib-go/internal/geom"
)

func approx(a, b float64, eps float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}

func approxPt(a, b geom.Pt, eps float64) bool { return approx(a.X, b.X, eps) && approx(a.Y, b.Y, eps) }

func TestLinearScale_RoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	for i := 0; i < 200; i++ {
		min := r.Float64()*2e6 - 1e6
		span := r.Float64()*2e6 + 1e-6 // ensure >0
		max := min + span
		s := NewLinear(min, max)
		for j := 0; j < 10; j++ {
			x := min + r.Float64()*(max-min)
			u := s.Fwd(x)
			xr, ok := s.Inv(u)
			if !ok {
				t.Fatalf("linear inv failed")
			}
			if !approx(x, xr, 1e-9*(1+math.Abs(x))) {
				t.Fatalf("roundtrip mismatch: x=%v xr=%v", x, xr)
			}
		}
	}
}

func TestLogScale_RoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	bases := []float64{2, math.E, 10}
	for i := 0; i < 100; i++ {
		min := math.Exp(r.Float64()*10 - 2) // ~[e^-2, e^8]
		span := r.Float64()*5 + 0.1
		max := min * (1 + span)
		base := bases[i%len(bases)]
		s := NewLog(min, max, base)
		for j := 0; j < 10; j++ {
			// pick x in (min,max]
			u := r.Float64()
			x, ok := s.Inv(u)
			if !ok {
				t.Fatalf("log inv failed")
			}
			ur := s.Fwd(x)
			if !(math.IsNaN(ur)) && !approx(u, ur, 1e-9) {
				t.Fatalf("log roundtrip mismatch: u=%v ur=%v", u, ur)
			}
		}
	}
}

func TestAxes2D_RoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(4))
	for i := 0; i < 100; i++ {
		// linear scales
		xmin := r.Float64()*200 - 100
		xmax := xmin + r.Float64()*100 + 1e-3
		ymin := r.Float64()*200 - 100
		ymax := ymin + r.Float64()*100 + 1e-3
		xs := NewLinear(xmin, xmax)
		ys := NewLinear(ymin, ymax)

		// random invertible affine for axes->pixel
		var M geom.Affine
		for {
			M = geom.Affine{
				A: r.Float64()*4 - 2,
				B: r.Float64()*4 - 2,
				C: r.Float64()*4 - 2,
				D: r.Float64()*4 - 2,
				E: r.Float64()*100 - 50,
				F: r.Float64()*100 - 50,
			}
			det := M.A*M.D - M.C*M.B
			if det > 1e-6 || det < -1e-6 {
				break
			}
		}
		t2 := NewAxes2D(xs, ys, NewAffine(M))

		for j := 0; j < 10; j++ {
			p := geom.Pt{X: xmin + r.Float64()*(xmax-xmin), Y: ymin + r.Float64()*(ymax-ymin)}
			q := t2.Apply(p)
			pr, ok := t2.Invert(q)
			if !ok {
				t.Fatalf("axes2d invert failed")
			}
			if !approxPt(p, pr, 1e-9) {
				t.Fatalf("axes2d roundtrip mismatch: p=%+v pr=%+v", p, pr)
			}
		}
	}
}

func TestEdgeCases(t *testing.T) {
	// Degenerate linear domain
	s := NewLinear(1, 1)
	if _, ok := s.Inv(0.5); ok {
		t.Fatalf("expected inv=false for degenerate linear domain")
	}

	// Invalid log params
	badBase := NewLog(1, 10, 1)
	if _, ok := badBase.Inv(0.5); ok {
		t.Fatalf("expected inv=false for base<=1")
	}
	badMin := NewLog(0, 10, 10)
	if _, ok := badMin.Inv(0.5); ok {
		t.Fatalf("expected inv=false for min<=0")
	}
	badRange := NewLog(5, 5, 10)
	if _, ok := badRange.Inv(0.5); ok {
		t.Fatalf("expected inv=false for min==max")
	}
}
