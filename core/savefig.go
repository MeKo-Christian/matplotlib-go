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
var supportedSaveExtensions = map[string]func(*Figure, render.Renderer, string, ...render.SaveOption) error{
	".eps": SavePS,
	".pdf": SavePDF,
	".pgf": SavePGF,
	".png": SavePNG,
	".ps":  SavePS,
	".svg": SaveSVG,
}

// SaveFig draws the figure and writes it to path using the appropriate exporter
// inferred from the file extension (e.g. .png, .svg).
//
// The renderer must implement the corresponding capability interface
// (render.PNGExporter for .png, render.SVGExporter for .svg).
func SaveFig(fig *Figure, r render.Renderer, path string, opts ...render.SaveOption) error {
	ext := resolveSaveFormat(fig, path, opts)
	if ext == "" {
		supported := supportedExtensionsList()
		return fmt.Errorf("savefig: path %q has no extension and no savefig.format set; supported: %s", path, supported)
	}
	handler, ok := supportedSaveExtensions[ext]
	if !ok {
		supported := supportedExtensionsList()
		return fmt.Errorf("savefig: unsupported format %q for %q; supported: %s", ext, path, supported)
	}
	saveOptions := render.ResolveSaveOptions(opts...)
	if err := saveOptions.ValidateForExtension(ext); err != nil {
		return fmt.Errorf("savefig: %w", err)
	}
	return handler(fig, r, path, opts...)
}

// resolveSaveFormat picks the output format key (a leading-dot extension) from
// the per-call savefig.format option, then the figure's savefig.format rcParam,
// then the path extension.
func resolveSaveFormat(fig *Figure, path string, opts []render.SaveOption) string {
	format := strings.TrimSpace(strings.ToLower(render.ResolveSaveOptions(opts...).Figure.Format))
	if format == "" && fig != nil {
		format = strings.TrimSpace(strings.ToLower(fig.RC.Savefig.Format))
	}
	if format == "" {
		return strings.ToLower(filepath.Ext(path))
	}
	if !strings.HasPrefix(format, ".") {
		format = "." + format
	}
	return format
}

func supportedExtensionsList() string {
	keys := make([]string, 0, len(supportedSaveExtensions))
	for k := range supportedSaveExtensions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
