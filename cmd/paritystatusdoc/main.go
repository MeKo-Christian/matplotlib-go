package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
)

type publicSurfaceArtifact struct {
	Rows []examplecatalog.PublicSurfaceRow `json:"rows"`
}

func main() {
	data, err := os.ReadFile("test/testdata/parity_surface/upstream_public_surface.json")
	if err != nil {
		fatal(err)
	}
	var artifact publicSurfaceArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		fatal(err)
	}
	fmt.Print(examplecatalog.MatplotlibParityStatusMarkdown(artifact.Rows))
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "paritystatusdoc: %v\n", err)
	os.Exit(1)
}
