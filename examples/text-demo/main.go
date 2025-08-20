package main

import (
	"fmt"
	"log"

	"matplotlib-go/backends/gobasic"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func main() {
	// Create a new renderer with white background
	width, height := 400, 200
	bgColor := render.Color{R: 1, G: 1, B: 1, A: 1} // white
	renderer := gobasic.New(width, height, bgColor)

	// Begin rendering
	viewport := geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: float64(width), Y: float64(height)},
	}
	if err := renderer.Begin(viewport); err != nil {
		log.Fatalf("Failed to begin rendering: %v", err)
	}
	defer renderer.End()

	// Test text measurement
	text := "Hello, matplotlib-go!"
	fontSize := 13.0
	metrics := renderer.MeasureText(text, fontSize, "default")
	fmt.Printf("Text metrics for '%s':\n", text)
	fmt.Printf("  Width: %.2f pixels\n", metrics.W)
	fmt.Printf("  Height: %.2f pixels\n", metrics.H)
	fmt.Printf("  Ascent: %.2f pixels\n", metrics.Ascent)
	fmt.Printf("  Descent: %.2f pixels\n", metrics.Descent)

	// Draw some text
	textColor := render.Color{R: 0, G: 0, B: 0, A: 1} // black

	// Draw text at different positions
	renderer.DrawText("matplotlib-go Text Rendering Demo", geom.Pt{X: 20, Y: 30}, 13, textColor)
	renderer.DrawText("Built with basicfont.Face7x13", geom.Pt{X: 20, Y: 60}, 13, textColor)
	renderer.DrawText("Supports basic text positioning", geom.Pt{X: 20, Y: 90}, 13, textColor)

	// Draw text with different "sizes" (scaling)
	renderer.DrawText("Small text (size 10)", geom.Pt{X: 20, Y: 120}, 10, textColor)
	renderer.DrawText("Large text (size 16)", geom.Pt{X: 20, Y: 150}, 16, textColor)

	// Draw colored text
	redColor := render.Color{R: 1, G: 0, B: 0, A: 1}
	blueColor := render.Color{R: 0, G: 0, B: 1, A: 1}
	renderer.DrawText("Red text", geom.Pt{X: 250, Y: 120}, 13, redColor)
	renderer.DrawText("Blue text", geom.Pt{X: 250, Y: 150}, 13, blueColor)

	// Save the result as PNG
	if err := renderer.SavePNG("text-demo.png"); err != nil {
		log.Fatalf("Failed to save PNG: %v", err)
	}

	fmt.Println("Text rendering demo saved as 'text-demo.png'")
}