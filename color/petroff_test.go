package color

import (
	"math"
	"sync"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

// matplotlib._cm._petroff10_data (Matplotlib 3.10.9), as float RGB triples.
var petroff10Reference = [10][3]float64{
	{0.24705882352941178, 0.5647058823529412, 0.8549019607843137},
	{1.0, 0.6627450980392157, 0.054901960784313725},
	{0.7411764705882353, 0.12156862745098039, 0.00392156862745098},
	{0.5803921568627451, 0.6431372549019608, 0.6352941176470588},
	{0.5137254901960784, 0.17647058823529413, 0.7137254901960784},
	{0.6627450980392157, 0.4196078431372549, 0.34901960784313724},
	{0.9058823529411765, 0.38823529411764707, 0.0},
	{0.7254901960784313, 0.6745098039215687, 0.4392156862745098},
	{0.44313725490196076, 0.4588235294117647, 0.5058823529411764},
	{0.5725490196078431, 0.8549019607843137, 0.8666666666666667},
}

func TestPetroff10MatchesMatplotlibData(t *testing.T) {
	if len(Petroff10) != 10 {
		t.Fatalf("Petroff10 length = %d, want 10", len(Petroff10))
	}
	for i, want := range petroff10Reference {
		got := Petroff10[i]
		if !floatClose(float64(got.R), want[0]) ||
			!floatClose(float64(got.G), want[1]) ||
			!floatClose(float64(got.B), want[2]) {
			t.Errorf("Petroff10[%d] = (%v, %v, %v), want (%v, %v, %v)",
				i, got.R, got.G, got.B, want[0], want[1], want[2])
		}
		if got.A != 1 {
			t.Errorf("Petroff10[%d].A = %v, want 1", i, got.A)
		}
	}
}

func TestColorSequenceRegistry(t *testing.T) {
	seq, ok := ColorSequence("petroff10")
	if !ok {
		t.Fatal("ColorSequence(\"petroff10\") not registered")
	}
	if len(seq) != 10 {
		t.Fatalf("petroff10 sequence length = %d, want 10", len(seq))
	}
	// Returned palette must be a defensive copy: mutating it must not affect the
	// package-level Petroff10.
	orig := Petroff10[0]
	seq[0] = render.Color{R: 0, G: 0, B: 0, A: 1}
	if Petroff10[0] != orig {
		t.Fatal("ColorSequence returned a shared (non-copied) palette")
	}
	if _, ok := ColorSequence("does-not-exist"); ok {
		t.Fatal("unknown sequence reported as registered")
	}
}

func TestColorSequenceRegistryConcurrentAccess(t *testing.T) {
	const name = "test-concurrent-sequence"
	palettes := []Palette{
		{{R: 1, A: 1}},
		{{G: 1, A: 1}},
	}
	RegisterColorSequence(name, palettes[0])

	start := make(chan struct{})
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			<-start
			for i := 0; i < 500; i++ {
				if worker%2 == 0 {
					RegisterColorSequence(name, palettes[i%len(palettes)])
					continue
				}
				seq, ok := ColorSequence(name)
				if !ok || len(seq) != 1 {
					t.Errorf("ColorSequence(%q) = %v, %v; want one color", name, seq, ok)
					return
				}
				_ = ColorSequenceNames()
			}
		}(worker)
	}
	close(start)
	wg.Wait()
}

func TestPetroff10ResolvableAsColormap(t *testing.T) {
	cmap := LookupColormap("petroff10")
	if cmap.Name() != "petroff10" {
		t.Fatalf("LookupColormap(\"petroff10\").Name() = %q, want \"petroff10\"", cmap.Name())
	}
	// First listed color should match the first sequence entry.
	c := cmap.At(0)
	if !floatClose(float64(c.R), petroff10Reference[0][0]) {
		t.Fatalf("petroff10 colormap At(0).R = %v, want %v", c.R, petroff10Reference[0][0])
	}
}

func floatClose(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}
