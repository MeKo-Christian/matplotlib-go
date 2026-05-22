package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/webdemo"
)

func main() {
	outputDir := flag.String("output-dir", filepath.Join("testdata", "_artifacts", "webdemo", "go"), "Directory to write PNG files")
	demos := flag.String("demos", "all", "Comma-separated demo IDs, or all")
	backend := flag.String("backend", "", "Web demo backend ID to render; default uses the web demo default backend")
	width := flag.Int("width", webdemo.DefaultWidth, "Rendered width in pixels")
	height := flag.Int("height", webdemo.DefaultHeight, "Rendered height in pixels")
	list := flag.Bool("list", false, "List available demo IDs and exit")
	listBackends := flag.Bool("list-backends", false, "List available backend IDs and exit")
	flag.Parse()

	if *listBackends {
		for _, descriptor := range webdemo.Backends() {
			fmt.Println(descriptor.ID)
		}
		return
	}

	if *list {
		for _, descriptor := range webdemo.Catalog() {
			fmt.Println(descriptor.ID)
		}
		return
	}

	ids, err := selectedDemoIDs(*demos)
	if err != nil {
		exitf("%v", err)
	}
	backendID, err := selectedBackendID(*backend)
	if err != nil {
		exitf("%v", err)
	}
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		exitf("create output dir %s: %v", *outputDir, err)
	}

	for _, id := range ids {
		pngBytes, descriptor, err := webdemo.RenderPNGWithBackend(id, backendID, *width, *height)
		if err != nil {
			exitf("render %s: %v", id, err)
		}
		path := filepath.Join(*outputDir, descriptor.ID+".png")
		if err := os.WriteFile(path, pngBytes, 0o644); err != nil {
			exitf("write %s: %v", path, err)
		}
		fmt.Printf("wrote %s\n", path)
	}
}

func selectedDemoIDs(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "all" {
		catalog := webdemo.Catalog()
		ids := make([]string, len(catalog))
		for i, descriptor := range catalog {
			ids[i] = descriptor.ID
		}
		return ids, nil
	}

	seen := make(map[string]bool)
	var ids []string
	for _, part := range strings.Split(raw, ",") {
		id := strings.TrimSpace(part)
		if id == "" || seen[id] {
			continue
		}
		if !webdemo.ValidDemoID(id) {
			return nil, fmt.Errorf("unknown web demo %q", id)
		}
		ids = append(ids, id)
		seen[id] = true
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no demos selected")
	}
	return ids, nil
}

func selectedBackendID(raw string) (string, error) {
	id := strings.TrimSpace(raw)
	if id == "" {
		id = webdemo.DefaultBackendID()
	}
	if !webdemo.ValidBackendID(id) {
		return "", fmt.Errorf("unknown web demo backend %q", id)
	}
	return id, nil
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
