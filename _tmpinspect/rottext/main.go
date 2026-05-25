package main

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	white := render.Color{R: 1, G: 1, B: 1, A: 1}
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	angles := []float64{0, 45, 90, 180, 270}
	for _, ang := range angles {
		r, err := agg.New(200, 200, white)
		if err != nil {
			panic(err)
		}
		r.DrawTextRotatedWithFont("R", geom.Pt{X: 100, Y: 100}, 80, ang, black, "DejaVu Sans")
		path := fmt.Sprintf("/tmp/rottext/R_ang_%03.0f.png", ang)
		if err := r.SavePNG(path); err != nil {
			panic(err)
		}
		fmt.Println("wrote", path)
	}
}
