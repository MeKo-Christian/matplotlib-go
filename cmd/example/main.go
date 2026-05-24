// Command example renders any cataloged showcase to a PNG file. It is the
// unified entry point that replaces the per-example runner files.
//
// Usage:
//
//	go run ./cmd/example -list
//	go run ./cmd/example -name basic_line -o /tmp/basic_line.png
//	MATPLOTLIB_BACKEND=agg go run ./cmd/example -name polar_axes -o polar.png
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	annotation_composition "github.com/cwbudde/matplotlib-go/examples/annotation_composition"
	arrays_showcase "github.com/cwbudde/matplotlib-go/examples/arrays_showcase"
	axes_control_surface "github.com/cwbudde/matplotlib-go/examples/axes_control_surface"
	axes_grid1_showcase "github.com/cwbudde/matplotlib-go/examples/axes_grid1_showcase"
	axisartist_showcase "github.com/cwbudde/matplotlib-go/examples/axisartist_showcase"
	bar_basic "github.com/cwbudde/matplotlib-go/examples/bar_basic"
	basic_line "github.com/cwbudde/matplotlib-go/examples/basic_line"
	boxplot_basic "github.com/cwbudde/matplotlib-go/examples/boxplot_basic"
	colorbar_composition "github.com/cwbudde/matplotlib-go/examples/colorbar_composition"
	dashes "github.com/cwbudde/matplotlib-go/examples/dashes"
	errorbar_basic "github.com/cwbudde/matplotlib-go/examples/errorbar_basic"
	figure_labels_composition "github.com/cwbudde/matplotlib-go/examples/figure_labels_composition"
	fill_basic "github.com/cwbudde/matplotlib-go/examples/fill_basic"
	geo_aitoff_axes "github.com/cwbudde/matplotlib-go/examples/geo_aitoff_axes"
	geo_mollweide_axes "github.com/cwbudde/matplotlib-go/examples/geo_mollweide_axes"
	gridspec_composition "github.com/cwbudde/matplotlib-go/examples/gridspec_composition"
	hist_basic "github.com/cwbudde/matplotlib-go/examples/hist_basic"
	image_heatmap "github.com/cwbudde/matplotlib-go/examples/image_heatmap"
	mesh_contour_tri "github.com/cwbudde/matplotlib-go/examples/mesh_contour_tri"
	mplot3d_terrain "github.com/cwbudde/matplotlib-go/examples/mplot3d_terrain"
	multi_series_basic "github.com/cwbudde/matplotlib-go/examples/multi_series_basic"
	plot_variants "github.com/cwbudde/matplotlib-go/examples/plot_variants"
	polar_axes "github.com/cwbudde/matplotlib-go/examples/polar_axes"
	radar_basic "github.com/cwbudde/matplotlib-go/examples/radar_basic"
	scatter_basic "github.com/cwbudde/matplotlib-go/examples/scatter_basic"
	skewt_basic "github.com/cwbudde/matplotlib-go/examples/skewt_basic"
	specialty_artists "github.com/cwbudde/matplotlib-go/examples/specialty_artists"
	stat_variants "github.com/cwbudde/matplotlib-go/examples/stat_variants"
	units_overview "github.com/cwbudde/matplotlib-go/examples/units_overview"
	unstructured_showcase "github.com/cwbudde/matplotlib-go/examples/unstructured_showcase"
	vector_fields "github.com/cwbudde/matplotlib-go/examples/vector_fields"
	widgets_gallery "github.com/cwbudde/matplotlib-go/examples/widgets_gallery"
	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
	"github.com/cwbudde/matplotlib-go/render"
)

// registry maps a catalog showcase ID to the Plot() function that builds the
// corresponding *core.Figure. Keep in sync with the Showcase: true rows in
// internal/examplecatalog/catalog.go.
var registry = map[string]func() *core.Figure{
	"annotation_composition":    annotation_composition.Plot,
	"arrays_showcase":           arrays_showcase.Plot,
	"axes_control_surface":      axes_control_surface.Plot,
	"axes_grid1_showcase":       axes_grid1_showcase.Plot,
	"axisartist_showcase":       axisartist_showcase.Plot,
	"bar_basic":                 bar_basic.Plot,
	"basic_line":                basic_line.Plot,
	"boxplot_basic":             boxplot_basic.Plot,
	"colorbar_composition":      colorbar_composition.Plot,
	"dashes":                    dashes.Plot,
	"errorbar_basic":            errorbar_basic.Plot,
	"figure_labels_composition": figure_labels_composition.Plot,
	"fill_basic":                fill_basic.Plot,
	"geo_aitoff_axes":           geo_aitoff_axes.Plot,
	"geo_mollweide_axes":        geo_mollweide_axes.Plot,
	"gridspec_composition":      gridspec_composition.Plot,
	"hist_basic":                hist_basic.Plot,
	"image_heatmap":             image_heatmap.Plot,
	"mesh_contour_tri":          mesh_contour_tri.Plot,
	"mplot3d_terrain":           mplot3d_terrain.Plot,
	"multi_series_basic":        multi_series_basic.Plot,
	"plot_variants":             plot_variants.Plot,
	"polar_axes":                polar_axes.Plot,
	"radar_basic":               radar_basic.Plot,
	"scatter_basic":             scatter_basic.Plot,
	"skewt_basic":               skewt_basic.Plot,
	"specialty_artists":         specialty_artists.Plot,
	"stat_variants":             stat_variants.Plot,
	"units_overview":            units_overview.Plot,
	"unstructured_showcase":     unstructured_showcase.Plot,
	"vector_fields":             vector_fields.Plot,
	"widgets_gallery":           widgets_gallery.Plot,
}

func main() {
	name := flag.String("name", "", "Catalog ID of the showcase to render")
	out := flag.String("o", "", "Output path (default: <name>.<format>)")
	format := flag.String("format", "", "Output format: png, svg, pdf, ps, eps, or pgf (default: inferred from -o or png)")
	pdfTitle := flag.String("pdf-title", "", "PDF metadata Title entry")
	pdfAuthor := flag.String("pdf-author", "", "PDF metadata Author entry")
	pdfSubject := flag.String("pdf-subject", "", "PDF metadata Subject entry")
	pdfCreationDate := flag.String("pdf-creation-date", "", "PDF CreationDate as RFC3339")
	pdfFontPolicy := flag.String("pdf-font-policy", "", "PDF font policy: path or embed")
	list := flag.Bool("list", false, "List all available showcase IDs and exit")
	flag.Parse()

	if *list {
		listShowcases()
		return
	}

	if *name == "" {
		fmt.Fprintln(os.Stderr, "error: -name is required (or use -list)")
		flag.Usage()
		os.Exit(2)
	}

	plot, ok := registry[*name]
	if !ok {
		log.Fatalf("unknown showcase %q (run with -list to see available IDs)", *name)
	}

	output := *out
	ext, err := outputExtension(*name, output, *format)
	if err != nil {
		log.Fatalf("format: %v", err)
	}
	if output == "" {
		output = *name + ext
	} else if filepath.Ext(output) == "" {
		output += ext
	} else if strings.TrimSpace(*format) != "" && strings.ToLower(filepath.Ext(output)) != ext {
		log.Fatalf("format: -format %s conflicts with output extension %s", *format, filepath.Ext(output))
	}
	saveOptions, err := exampleSaveOptions(*pdfTitle, *pdfAuthor, *pdfSubject, *pdfCreationDate, *pdfFontPolicy)
	if err != nil {
		log.Fatalf("options: %v", err)
	}

	fig := plot()
	w := int(fig.SizePx.X)
	h := int(fig.SizePx.Y)
	backend, err := selectExampleBackend(ext, exampleRequiredCapabilities())
	if err != nil {
		log.Fatalf("renderer: %v", err)
	}
	r, err := backends.Create(backend, backends.Config{
		Width:      w,
		Height:     h,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        fig.RC.DPI,
	})
	if err != nil {
		log.Fatalf("renderer: %v", err)
	}
	applyExampleSaveOptions(r, render.ResolveSaveOptions(saveOptions...))
	core.DrawFigure(fig, r)
	if err := backends.DefaultRegistry.SaveViaExtension(backend, r, output, saveOptions...); err != nil {
		log.Fatalf("save: %v", err)
	}
	log.Printf("saved %s", output)
}

func exampleSaveOptions(title, author, subject, creationDate, fontPolicy string) ([]render.SaveOption, error) {
	var opts []render.SaveOption
	metadata := map[string]string{}
	if strings.TrimSpace(title) != "" {
		metadata["Title"] = title
	}
	if strings.TrimSpace(author) != "" {
		metadata["Author"] = author
	}
	if strings.TrimSpace(subject) != "" {
		metadata["Subject"] = subject
	}
	if len(metadata) > 0 {
		opts = append(opts, render.WithPDFMetadata(metadata))
	}
	if strings.TrimSpace(creationDate) != "" {
		t, err := time.Parse(time.RFC3339, creationDate)
		if err != nil {
			return nil, fmt.Errorf("parse -pdf-creation-date: %w", err)
		}
		opts = append(opts, render.WithPDFCreationDate(t))
	}
	switch strings.ToLower(strings.TrimSpace(fontPolicy)) {
	case "":
	case string(render.PDFFontPolicyPath):
		opts = append(opts, render.WithPDFFontPolicy(render.PDFFontPolicyPath))
	case string(render.PDFFontPolicyEmbed):
		opts = append(opts, render.WithPDFFontPolicy(render.PDFFontPolicyEmbed))
	default:
		return nil, fmt.Errorf("unknown -pdf-font-policy %q", fontPolicy)
	}
	return opts, nil
}

func applyExampleSaveOptions(renderer render.Renderer, opts render.SaveOptions) {
	if setter, ok := renderer.(render.SVGOptionSetter); ok {
		setter.SetSVGOptions(opts.SVG)
	}
	if setter, ok := renderer.(render.PDFOptionSetter); ok {
		setter.SetPDFOptions(opts.PDF)
	}
	if setter, ok := renderer.(render.PSOptionSetter); ok {
		setter.SetPSOptions(opts.PS)
	}
	if setter, ok := renderer.(render.PGFOptionSetter); ok {
		setter.SetPGFOptions(opts.PGF)
	}
}

func outputExtension(name, output, format string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(format))
	if normalized != "" {
		if strings.HasPrefix(normalized, ".") {
			return normalized, nil
		}
		return "." + normalized, nil
	}
	if ext := strings.ToLower(filepath.Ext(output)); ext != "" {
		return ext, nil
	}
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("missing showcase name")
	}
	return ".png", nil
}

func selectExampleBackend(ext string, required []backends.Capability) (backends.Backend, error) {
	choice := strings.TrimSpace(os.Getenv("MATPLOTLIB_BACKEND"))
	backend, err := backends.SelectBackendForExtension(choice, ext, required)
	if err == nil {
		return backend, nil
	}
	if choice != "" {
		if fallback, fallbackErr := backends.SelectBackendForExtension("", ext, required); fallbackErr == nil {
			return fallback, nil
		}
	}
	return "", err
}

func exampleRequiredCapabilities() []backends.Capability {
	return []backends.Capability{backends.TextShaping}
}

func listShowcases() {
	cases := examplecatalog.Cases()
	type row struct {
		id, title string
	}
	var rows []row
	for i := range cases {
		c := cases[i]
		if !c.Showcase {
			continue
		}
		if _, ok := registry[c.ID]; !ok {
			continue
		}
		rows = append(rows, row{id: c.ID, title: c.Title})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].id < rows[j].id })

	maxID := 0
	for _, r := range rows {
		if len(r.id) > maxID {
			maxID = len(r.id)
		}
	}
	for _, r := range rows {
		fmt.Printf("%-*s  %s\n", maxID, r.id, r.title)
	}
}
