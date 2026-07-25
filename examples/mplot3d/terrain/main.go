// CLI runner for the examples/mplot3d_terrain showcase.
package main

import (
	"log"

	_ "github.com/cwbudde/matplotlib-go/backends/all"
	example "github.com/cwbudde/matplotlib-go/examples/mplot3d_terrain"
)

func main() {
	fig, _ := example.Plot()
	if err := fig.Save("mplot3d_terrain.png"); err != nil {
		log.Fatalf("save: %v", err)
	}
	log.Println("saved mplot3d_terrain.png")
}
