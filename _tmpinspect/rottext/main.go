package main

import (
	"fmt"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	white := render.Color{R: 1, G: 1, B: 1, A: 1}
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	degs := []float64{0, 45, 90, 180, 270}
	for _, deg := range degs {
		ang := deg * math.Pi / 180.0
		r, err := agg.New(200, 200, white)
		if err != nil {
			panic(err)
		}
		r.DrawTextRotatedWithFont("R", geom.Pt{X: 100, Y: 100}, 80, ang, black, "DejaVu Sans")
		path := fmt.Sprintf("/tmp/rottext/R_ang_%03.0f.png", deg)
		if err := r.SavePNG(path); err != nil {
			panic(err)
		}
		fmt.Println("wrote", path)
	}
}
