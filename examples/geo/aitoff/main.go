// CLI runner for the examples/geo_aitoff_axes showcase. Renders to
// geo_aitoff_axes.png.
package main

import (
	"log"

	_ "github.com/cwbudde/matplotlib-go/backends/all"
	example "github.com/cwbudde/matplotlib-go/examples/geo_aitoff_axes"
)

func main() {
	fig, _ := example.Plot()
	if err := fig.Save("geo_aitoff_axes.png"); err != nil {
		log.Fatalf("save: %v", err)
	}
	log.Println("saved geo_aitoff_axes.png")
}
