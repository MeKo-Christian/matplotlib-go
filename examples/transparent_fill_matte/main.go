// Command transparent_fill_matte showcases a known AGG compositing bug
// (matplotlib-go PLAN "Phase 21: Transparent-surface fill matte").
//
// A semi-transparent fill (alpha < 1) drawn through the full figure pipeline
// onto a transparent figure surface comes out matted toward the clear color's
// RGB instead of keeping its pure straight color. With the default white
// transparent clear, a 0.8-alpha pink bar (#E91E63) is saved as
// (237,75,130,204) = 0.8*color + 0.2*white, where matplotlib/Agg produces the
// pure (233,30,99,204).
//
// The same fill drawn directly on a bare agg.Renderer stays pure (see
// backends/agg/save_alpha_test.go), so the trigger is internal agg2d painter
// blend state reached only via the figure pipeline. The correct fix lives in
// the agg2d fill blender: weight the destination RGB by destination alpha so a
// fully transparent destination contributes no color.
//
// Run: go run ./examples/transparent_fill_matte
package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"os"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

const (
	width  = 600
	height = 400
)

func main() {
	output := flag.String("out", "transparent_fill_matte.png", "output PNG file")
	flag.Parse()

	// Transparent figure + axes background, exactly as a transparent savefig.
	transparent := render.Color{R: 1, G: 1, B: 1, A: 0}
	fig := core.NewFigure(
		width, height,
		style.WithBackground(1, 1, 1, 0),
		style.WithAxesBackground(transparent),
	)
	grid := fig.Subplots(1, 1)
	ax := grid[0][0]
	ax.SetTitle("D5b: semi-transparent fill matte")
	ax.SetXLim(0, 1.1)
	ax.SetYLim(-0.5, 1.0)

	// Pink #E91E63 at alpha 0.8 -> pure straight (233,30,99,204).
	pink := render.Color{R: 233.0 / 255, G: 30.0 / 255, B: 99.0 / 255, A: 0.8}
	orient := core.BarHorizontal
	bw := 0.5
	_, _ = ax.Bar([]float64{0.5}, []float64{0.9}, core.BarOptions{Color: &pink, Width: &bw, Orientation: &orient})

	r, _, err := backends.NewRenderer("agg", backends.Config{
		Width:       width,
		Height:      height,
		Background:  transparent,
		DPI:         100,
		Transparent: true,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("create renderer: %v\n", err)
		os.Exit(1)
	}
	if err := core.SavePNG(fig, r, *output); err != nil {
		fmt.Printf("save: %v\n", err)
		os.Exit(1)
	}

	got, ok := sampleBar(*output, [3]int{233, 30, 99}, [3]int{237, 75, 130})
	fmt.Printf("Wrote %s\n", *output)
	switch {
	case !ok:
		fmt.Println("No bar pixel found — example may need adjusting.")
	case near(got, [3]uint8{233, 30, 99}):
		fmt.Printf("PURE: bar = %v (matches matplotlib) — D5b appears FIXED.\n", got)
	case near(got, [3]uint8{237, 75, 130}):
		fmt.Printf("MATTED: bar = %v, want (233,30,99) — D5b reproduced "+
			"(0.8*color + 0.2*white leaked from the transparent clear).\n", got)
	default:
		fmt.Printf("bar = %v (neither pure (233,30,99) nor matted (237,75,130))\n", got)
	}
}

// sampleBar returns the most common bar-ish pixel (matching either target),
// reading straight (non-premultiplied) RGB.
func sampleBar(path string, targets ...[3]int) ([3]uint8, bool) {
	f, err := os.Open(path)
	if err != nil {
		return [3]uint8{}, false
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		return [3]uint8{}, false
	}
	nr := image.NewNRGBA(src.Bounds())
	draw.Draw(nr, nr.Bounds(), src, src.Bounds().Min, draw.Src)
	b := nr.Bounds()
	counts := map[[3]uint8]int{}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := nr.NRGBAAt(x, y)
			if c.A == 0 {
				continue
			}
			for _, t := range targets {
				if near([3]uint8{c.R, c.G, c.B}, [3]uint8{uint8(t[0]), uint8(t[1]), uint8(t[2])}) {
					counts[[3]uint8{c.R, c.G, c.B}]++
				}
			}
		}
	}
	var best [3]uint8
	bestN := 0
	for k, n := range counts {
		if n > bestN {
			best, bestN = k, n
		}
	}
	return best, bestN > 0
}

func near(a, b [3]uint8) bool {
	d := func(x, y uint8) int {
		if x > y {
			return int(x - y)
		}
		return int(y - x)
	}
	return d(a[0], b[0]) <= 8 && d(a[1], b[1]) <= 8 && d(a[2], b[2]) <= 8
}
