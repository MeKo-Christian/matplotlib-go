package webdemo

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"slices"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	"github.com/cwbudde/matplotlib-go/core"
	animation_gallery "github.com/cwbudde/matplotlib-go/examples/animation_gallery"
	annotation_legend_offsetbox_gallery "github.com/cwbudde/matplotlib-go/examples/annotation_legend_offsetbox_gallery"
	arrays_showcase "github.com/cwbudde/matplotlib-go/examples/arrays_showcase"
	axes_control_surface "github.com/cwbudde/matplotlib-go/examples/axes_control_surface"
	bar_variants "github.com/cwbudde/matplotlib-go/examples/bar_variants"
	errorbar_basic "github.com/cwbudde/matplotlib-go/examples/errorbar_basic"
	figure_labels_composition "github.com/cwbudde/matplotlib-go/examples/figure_labels_composition"
	fill_variants "github.com/cwbudde/matplotlib-go/examples/fill_variants"
	geo_mollweide_axes "github.com/cwbudde/matplotlib-go/examples/geo_mollweide_axes"
	gridspec_composition "github.com/cwbudde/matplotlib-go/examples/gridspec_composition"
	histogram_variants "github.com/cwbudde/matplotlib-go/examples/histogram_variants"
	image_heatmap "github.com/cwbudde/matplotlib-go/examples/image_heatmap"
	lines_markers_gallery "github.com/cwbudde/matplotlib-go/examples/lines_markers_gallery"
	mesh_contour_tri "github.com/cwbudde/matplotlib-go/examples/mesh_contour_tri"
	patch_showcase "github.com/cwbudde/matplotlib-go/examples/patch_showcase"
	plot_variants "github.com/cwbudde/matplotlib-go/examples/plot_variants"
	polar_axes "github.com/cwbudde/matplotlib-go/examples/polar_axes"
	scatter_gallery "github.com/cwbudde/matplotlib-go/examples/scatter_gallery"
	specialty_artists "github.com/cwbudde/matplotlib-go/examples/specialty_artists"
	stat_variants "github.com/cwbudde/matplotlib-go/examples/stat_variants"
	units_overview "github.com/cwbudde/matplotlib-go/examples/units_overview"
	vector_fields "github.com/cwbudde/matplotlib-go/examples/vector_fields"
	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	DefaultWidth  = 960
	DefaultHeight = 540
)

type Descriptor struct {
	ID          string
	Title       string
	Description string
}

type BackendDescriptor struct {
	ID          string
	Name        string
	Description string
}

var descriptors = webDescriptorsFromCatalog()

var backendDescriptors = append(baseBackendDescriptors(), optionalBackendDescriptors()...)

// showcaseBuilders maps a web-demo ID to the showcase package's Plot() function.
// The Plot() functions are backend-agnostic and produce a *core.Figure with
// dimensions baked in by the showcase author. The width/height arguments to
// Build are advisory only — the figure's own SizePx wins at render time.
var showcaseBuilders = map[string]func() *core.Figure{
	"lines":       lines_markers_gallery.Plot,
	"scatter":     scatter_gallery.Plot,
	"bars":        bar_variants.Plot,
	"fills":       fill_variants.Plot,
	"errorbars":   errorbar_basic.Plot,
	"histogram":   histogram_variants.Plot,
	"heatmap":     image_heatmap.Plot,
	"axes":        axes_control_surface.Plot,
	"composition": gridspec_composition.Plot,
	"subplots":    figure_labels_composition.Plot,
	"annotations": annotation_legend_offsetbox_gallery.Plot,
	"patches":     patch_showcase.Plot,
	"mesh":        mesh_contour_tri.Plot,
	"variants":    plot_variants.Plot,
	"statistics":  stat_variants.Plot,
	"specialty":   specialty_artists.Plot,
	"units":       units_overview.Plot,
	"vectors":     vector_fields.Plot,
	"polar":       polar_axes.Plot,
	"projections": geo_mollweide_axes.Plot,
	"matrix":      arrays_showcase.Plot,
	"animation":   animation_gallery.Plot,
}

type rasterRenderer interface {
	render.Renderer
	GetImage() *image.RGBA
}

func webDescriptorsFromCatalog() []Descriptor {
	cases := examplecatalog.WebDemos()
	out := make([]Descriptor, 0, len(cases))
	for _, c := range cases {
		out = append(out, Descriptor{
			ID:          c.WebDemoID,
			Title:       c.Title,
			Description: c.Description,
		})
	}
	return out
}

func Catalog() []Descriptor {
	out := make([]Descriptor, len(descriptors))
	copy(out, descriptors)
	return out
}

func Backends() []BackendDescriptor {
	out := make([]BackendDescriptor, len(backendDescriptors))
	copy(out, backendDescriptors)
	return out
}

// Build returns the showcase figure for the given web-demo ID. The width and
// height arguments are kept for API stability but the figure's intrinsic
// SizePx (set by the showcase Plot()) is what drives rendering downstream.
func Build(id string, width, height int) (*core.Figure, Descriptor, error) {
	_ = width
	_ = height

	for _, descriptor := range descriptors {
		if descriptor.ID != id {
			continue
		}
		builder, ok := showcaseBuilders[id]
		if !ok {
			return nil, Descriptor{}, fmt.Errorf("webdemo: no showcase builder for demo %q", id)
		}
		return builder(), descriptor, nil
	}

	return nil, Descriptor{}, fmt.Errorf("webdemo: unknown demo %q", id)
}

func Render(id string, width, height int) (*image.RGBA, Descriptor, error) {
	return RenderWithBackend(id, DefaultBackendID(), width, height)
}

func RenderWithBackend(id, backendID string, width, height int) (*image.RGBA, Descriptor, error) {
	fig, descriptor, err := Build(id, width, height)
	if err != nil {
		return nil, Descriptor{}, err
	}

	renderWidth := width
	renderHeight := height
	if renderWidth <= 0 {
		renderWidth = DefaultWidth
	}
	if renderHeight <= 0 {
		renderHeight = DefaultHeight
	}
	if fig != nil {
		if w := int(fig.SizePx.X); w > 0 {
			renderWidth = w
		}
		if h := int(fig.SizePx.Y); h > 0 {
			renderHeight = h
		}
	}

	r, err := newRasterRenderer(backendID, renderWidth, renderHeight, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		return nil, Descriptor{}, err
	}

	core.DrawFigure(fig, r)
	return r.GetImage(), descriptor, nil
}

func RenderPNG(id string, width, height int) ([]byte, Descriptor, error) {
	return RenderPNGWithBackend(id, DefaultBackendID(), width, height)
}

func RenderPNGWithBackend(id, backendID string, width, height int) ([]byte, Descriptor, error) {
	img, descriptor, err := RenderWithBackend(id, backendID, width, height)
	if err != nil {
		return nil, Descriptor{}, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, Descriptor{}, fmt.Errorf("webdemo: encode PNG: %w", err)
	}
	return buf.Bytes(), descriptor, nil
}

func DefaultDemoID() string {
	return descriptors[0].ID
}

func DefaultBackendID() string {
	return "agg"
}

func ValidDemoID(id string) bool {
	return slices.ContainsFunc(descriptors, func(descriptor Descriptor) bool {
		return descriptor.ID == id
	})
}

func ValidBackendID(id string) bool {
	return slices.ContainsFunc(backendDescriptors, func(descriptor BackendDescriptor) bool {
		return descriptor.ID == id
	})
}

func newRasterRenderer(backendID string, width, height int, bg render.Color) (rasterRenderer, error) {
	if backendID == "" {
		backendID = DefaultBackendID()
	}
	switch backendID {
	case "gobasic":
		r := gobasic.New(width, height, bg)
		if r == nil {
			return nil, errors.New("webdemo: failed to create gobasic renderer")
		}
		return r, nil
	case "agg":
		r, err := agg.New(width, height, bg)
		if err != nil {
			return nil, fmt.Errorf("webdemo: create agg renderer: %w", err)
		}
		return r, nil
	default:
		if r, ok, err := newOptionalRasterRenderer(backendID, width, height, bg); ok || err != nil {
			return r, err
		}
		return nil, fmt.Errorf("webdemo: unknown backend %q", backendID)
	}
}

func baseBackendDescriptors() []BackendDescriptor {
	return []BackendDescriptor{
		{
			ID:          "gobasic",
			Name:        "GoBasic",
			Description: "Pure-Go raster backend available for browser rendering.",
		},
		{
			ID:          "agg",
			Name:        "AGG",
			Description: "Anti-Grain Geometry raster backend via github.com/cwbudde/agg_go.",
		},
	}
}
