package agg

import (
	"image/color"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// TestTranslucentFillMatchesMatplotlibBlender pins the AGG surface to
// Matplotlib's compositing arithmetic. Matplotlib does not use stock AGG here:
// _backend_agg.h builds its pixel format from fixed_blender_rgba_plain
// (src/agg_workaround.h), which restores AGG 2.4's high-precision plain blend
// because the agg24-svn rewrite routes it through multiply -> lerp ->
// demultiply and loses an LSB. newAggSurface asks agg_go for that blender; if
// the dependency ever drops back to the agg24-svn one, these values shift by
// one on most channels and roughly four million pixels across the parity suite
// go with them.
//
// Every want below was read out of Matplotlib 3.10.9 rendering the same fill
// over an opaque white canvas.
func TestTranslucentFillMatchesMatplotlibBlender(t *testing.T) {
	tests := []struct {
		name string
		fill render.Color
		want color.RGBA
	}{
		{
			name: "step filled blue",
			fill: render.Color{R: 0.42, G: 0.62, B: 0.90, A: 0.55},
			want: color.RGBA{R: 173, G: 201, B: 241, A: 255},
		},
		{
			name: "stacked orange",
			fill: render.Color{R: 0.86, G: 0.42, B: 0.19, A: 0.8},
			want: color.RGBA{R: 226, G: 136, B: 89, A: 255},
		},
		{
			name: "stackplot blue",
			fill: render.Color{R: 0.20, G: 0.55, B: 0.75, A: 0.76},
			want: color.RGBA{R: 100, G: 167, B: 206, A: 255},
		},
		{
			name: "straight alpha blue",
			fill: render.Color{R: 0.36, G: 0.56, B: 0.92, A: 0.2},
			want: color.RGBA{R: 222, G: 232, B: 251, A: 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := mustNew(t, 100, 100)
			viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}
			if err := r.Begin(viewport); err != nil {
				t.Fatal(err)
			}
			r.Path(fullRectPath(100, 100), &render.Paint{Fill: tt.fill})
			if err := r.End(); err != nil {
				t.Fatal(err)
			}

			if got := r.Image().RGBAAt(50, 50); got != tt.want {
				t.Fatalf("interior pixel = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestOpaqueFillIsCopiedNotBlended guards the short-circuit the blender needs:
// C++ AGG stores an opaque source at full coverage rather than running it
// through blend_pix, and the high-precision formula is off by one if that
// short-circuit is missing.
func TestOpaqueFillIsCopiedNotBlended(t *testing.T) {
	r := mustNew(t, 100, 100)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}
	if err := r.Begin(viewport); err != nil {
		t.Fatal(err)
	}
	r.Path(fullRectPath(100, 100), &render.Paint{
		Fill: render.Color{R: 0.36, G: 0.56, B: 0.92, A: 1},
	})
	if err := r.End(); err != nil {
		t.Fatal(err)
	}

	want := color.RGBA{R: 92, G: 143, B: 235, A: 255}
	if got := r.Image().RGBAAt(50, 50); got != want {
		t.Fatalf("interior pixel = %+v, want %+v", got, want)
	}
}
