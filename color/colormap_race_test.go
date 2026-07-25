package color

import (
	"fmt"
	"sync"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

// TestColormapRegistryConcurrentAccess guards the registry against the data
// race found in the 2026-07-01 audit (REVIEW.md third review §6):
// RegisterColormap wrote the package-global colormaps map with no mutex while
// draw paths read it via LookupColormap/LookupColormapStrict/ColormapNames. Run with
// -race; without the registry lock the race detector fails this test.
func TestColormapRegistryConcurrentAccess(t *testing.T) {
	const iterations = 100

	cmap := NewListedColormap("race-probe", []render.Color{
		{R: 0, G: 0, B: 0, A: 1},
		{R: 1, G: 1, B: 1, A: 1},
	})

	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		for i := range iterations {
			RegisterColormap(fmt.Sprintf("race-probe-%d", i), cmap)
		}
	}()
	go func() {
		defer wg.Done()
		for range iterations {
			LookupColormap("viridis")
			LookupColormap("no-such-colormap") // exercises the fallback read
		}
	}()
	go func() {
		defer wg.Done()
		for range iterations {
			if _, err := LookupColormapStrict("blues_r"); err != nil { // exercises the _r suffix read
				t.Errorf("LookupColormapStrict(blues_r): %v", err)
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for range iterations {
			ColormapNames()
		}
	}()
	wg.Wait()

	if _, err := LookupColormapStrict("race-probe-0"); err != nil {
		t.Fatalf("registered colormap not retrievable: %v", err)
	}
}
