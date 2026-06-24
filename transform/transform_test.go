package transform

import (
	"math"
	"math/rand"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func approx(a, b, eps float64) bool {
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
		minVal := r.Float64()*2e6 - 1e6
		span := r.Float64()*2e6 + 1e-6 // ensure >0
		maxVal := minVal + span
		s := NewLinear(minVal, maxVal)
		for j := 0; j < 10; j++ {
			x := minVal + r.Float64()*(maxVal-minVal)
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
		minVal := math.Exp(r.Float64()*10 - 2) // ~[e^-2, e^8]
		span := r.Float64()*5 + 0.1
		maxVal := minVal * (1 + span)
		base := bases[i%len(bases)]
		s := NewLog(minVal, maxVal, base)
		for j := 0; j < 10; j++ {
			// pick x in (min,max]
			u := r.Float64()
			x, ok := s.Inv(u)
			if !ok {
				t.Fatalf("log inv failed")
			}
			ur := s.Fwd(x)
			if !math.IsNaN(ur) && !approx(u, ur, 1e-9) {
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

// affineProviderStub is a third-party transform (defined outside the known set
// of the transform package) that declares its exact affine representation via
// the AffineProvider capability interface.
type affineProviderStub struct {
	m      geom.Affine
	affine bool
}

func (s affineProviderStub) Apply(p geom.Pt) geom.Pt { return s.m.Apply(p) }

func (s affineProviderStub) Invert(p geom.Pt) (geom.Pt, bool) {
	inv, ok := s.m.Invert()
	if !ok {
		return geom.Pt{}, false
	}
	return inv.Apply(p), true
}

func (s affineProviderStub) AsAffine() (geom.Affine, bool) { return s.m, s.affine }

// opaqueT is a third-party transform that does not implement AffineProvider.
type opaqueT struct{}

func (opaqueT) Apply(p geom.Pt) geom.Pt          { return p }
func (opaqueT) Invert(p geom.Pt) (geom.Pt, bool) { return p, true }

func TestAsAffine_ThirdPartyProvider(t *testing.T) {
	m := geom.Affine{A: 2, D: 3, E: 5, F: 7}

	got, ok := AsAffine(affineProviderStub{m: m, affine: true})
	if !ok {
		t.Fatalf("expected third-party AffineProvider to flatten to an affine")
	}
	if got != m {
		t.Fatalf("affine mismatch: got %+v want %+v", got, m)
	}
}

func TestAsAffine_ThirdPartyProviderNonAffine(t *testing.T) {
	if _, ok := AsAffine(affineProviderStub{affine: false}); ok {
		t.Fatalf("provider reporting non-affine must make AsAffine return false")
	}
}

func TestAsAffine_OpaqueThirdPartyStaysNonAffine(t *testing.T) {
	if _, ok := AsAffine(opaqueT{}); ok {
		t.Fatalf("transform without AffineProvider must stay non-affine")
	}
}

func TestAsAffine_ProviderInsideChain(t *testing.T) {
	base := geom.Affine{A: 2, D: 2, E: 1, F: 1}
	got, ok := AsAffine(Chain{
		A: affineProviderStub{m: base, affine: true},
		B: NewAffine(geom.Affine{A: 1, D: 1, E: 10, F: 20}),
	})
	if !ok {
		t.Fatalf("chain containing an AffineProvider should flatten")
	}
	want := geom.Affine{A: 1, D: 1, E: 10, F: 20}.Mul(base)
	if got != want {
		t.Fatalf("chained affine mismatch: got %+v want %+v", got, want)
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
