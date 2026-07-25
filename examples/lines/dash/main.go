// Package demonstrates dash patterns with matplotlib-go.
package main

import (
	"log"

	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/examples/dashes"
)

func main() {
	fig := dashes.Plot()
	if err := fig.Save("dashes.png"); err != nil {
		log.Fatalf("Failed to save PNG: %v", err)
	}

	log.Println("saved dashes.png")
}
