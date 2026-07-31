package imageinterp

import "testing"

func TestResolveFollowsMatplotlibAntialiasedPolicy(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		srcW, srcH, dstW, dstH float64
		want                   string
	}{
		// A 24x7 heatmap on a 1218x883 axes: past the 3x cutoff both ways.
		{name: "heavy upscale", srcW: 24, srcH: 7, dstW: 1218, dstH: 883, want: Nearest},
		{name: "unscaled", srcW: 24, srcH: 7, dstW: 24, dstH: 7, want: Nearest},
		{name: "exactly doubled", srcW: 24, srcH: 7, dstW: 48, dstH: 14, want: Nearest},
		{name: "mild upscale", srcW: 24, srcH: 7, dstW: 60, dstH: 17, want: Hanning},
		{name: "downscale", srcW: 240, srcH: 70, dstW: 60, dstH: 17, want: Hanning},
		// Both axes must qualify: a wide stretch with an untouched height is
		// still filtered, matching _ImageBase._make_image.
		{name: "one axis only", srcW: 24, srcH: 7, dstW: 1218, dstH: 20, want: Hanning},
		{name: "degenerate destination", srcW: 24, srcH: 7, dstW: 0, dstH: 0, want: Hanning},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Resolve("antialiased", tc.srcW, tc.srcH, tc.dstW, tc.dstH); got != tc.want {
				t.Fatalf("Resolve(antialiased, %v/%v -> %v/%v) = %q, want %q",
					tc.srcW, tc.srcH, tc.dstW, tc.dstH, got, tc.want)
			}
		})
	}
}

// A fixed filter name is the caller's explicit choice and must survive the
// policy untouched, whatever the scale.
func TestResolveLeavesFixedFiltersAlone(t *testing.T) {
	for _, name := range []string{"", "nearest", "bilinear", "lanczos"} {
		if got := Resolve(name, 24, 7, 1218, 883); got != name {
			t.Fatalf("Resolve(%q) = %q, want it unchanged", name, got)
		}
	}
}

func TestIsNearestTreatsTheDefaultAsUnfiltered(t *testing.T) {
	for name, want := range map[string]bool{
		"":          true, // renderer default
		"none":      true,
		"NEAREST":   true,
		" nearest ": true,
		"bilinear":  false,
		"hanning":   false,
	} {
		if got := IsNearest(name); got != want {
			t.Fatalf("IsNearest(%q) = %v, want %v", name, got, want)
		}
	}
}
