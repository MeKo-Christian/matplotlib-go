package main

import (
	"flag"
	"image/png"
	"log"
	"os"

	gallery "github.com/cwbudde/matplotlib-go/examples/plot2d_gallery"
)

func main() {
	output := flag.String("output", "plot2d_gallery.png", "output PNG path")
	flag.Parse()

	file, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(file, gallery.Render()); err != nil {
		_ = file.Close()
		log.Fatal(err)
	}
	if err := file.Close(); err != nil {
		log.Fatal(err)
	}
}
