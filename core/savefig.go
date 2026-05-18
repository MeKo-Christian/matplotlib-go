package core

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cwbudde/matplotlib-go/render"
)

// supportedSaveExtensions is the registry of extensions handled by SaveFig.
//
// Adding a new exporter (e.g. PostScript) means appending to this map and
// implementing the corresponding render-side capability interface.
var supportedSaveExtensions = map[string]func(*Figure, render.Renderer, string, ...render.SVGOption) error{
	".eps": SavePS,
	".png": SavePNG,
	".ps":  SavePS,
	".svg": SaveSVG,
}

// SaveFig draws the figure and writes it to path using the appropriate exporter
// inferred from the file extension (e.g. .png, .svg).
//
// The renderer must implement the corresponding capability interface
// (render.PNGExporter for .png, render.SVGExporter for .svg).
func SaveFig(fig *Figure, r render.Renderer, path string, opts ...render.SVGOption) error {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		supported := supportedExtensionsList()
		return fmt.Errorf("savefig: path %q has no extension; supported: %s", path, supported)
	}
	handler, ok := supportedSaveExtensions[ext]
	if !ok {
		supported := supportedExtensionsList()
		return fmt.Errorf("savefig: unsupported extension %q for %q; supported: %s", ext, path, supported)
	}
	return handler(fig, r, path, opts...)
}

func supportedExtensionsList() string {
	keys := make([]string, 0, len(supportedSaveExtensions))
	for k := range supportedSaveExtensions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
