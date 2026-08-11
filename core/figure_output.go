package core

import (
	"fmt"
	"image"
	"image/draw"
	"io"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/cwbudde/matplotlib-go/render"
)

// FigureOutputRendererFactory constructs a renderer for Figure output.
type FigureOutputRendererFactory func(width, height int, background render.Color) (render.Renderer, error)

var figureOutputRenderers = struct {
	sync.RWMutex
	factories map[string]FigureOutputRendererFactory
}{
	factories: make(map[string]FigureOutputRendererFactory),
}

type nrgbaExporter interface {
	ImageNRGBA() *image.NRGBA
}

// RegisterFigureOutputRenderer registers a renderer factory for a file format.
// Backend packages call this during initialization; applications can import
// backends/all to register all built-in output formats.
func RegisterFigureOutputRenderer(format string, factory FigureOutputRendererFactory) {
	ext := normalizeOutputFormat(format)
	if ext == "" || factory == nil {
		return
	}
	figureOutputRenderers.Lock()
	defer figureOutputRenderers.Unlock()
	figureOutputRenderers.factories[ext] = factory
}

// Save renders the figure and writes it to path. The output format is selected
// from render.WithSaveFormat, the figure's savefig.format setting, or the path
// extension, in that order.
func (f *Figure) Save(path string, opts ...render.SaveOption) error {
	if f == nil {
		return fmt.Errorf("figure save: nil figure")
	}
	ext := resolveSaveFormat(f, path, opts)
	r, err := f.newOutputRenderer(ext, opts...)
	if err != nil {
		return err
	}
	return SaveFig(f, r, path, opts...)
}

// WriteTo renders the figure in format and writes the encoded output to w.
//
// Supported formats are png, svg, pdf, ps, eps, and pgf.
func (f *Figure) WriteTo(w io.Writer, format string, opts ...render.SaveOption) error {
	if f == nil {
		return fmt.Errorf("figure write: nil figure")
	}
	if w == nil {
		return fmt.Errorf("figure write: nil writer")
	}
	ext := normalizeOutputFormat(format)
	if _, ok := supportedSaveExtensions[ext]; !ok {
		return fmt.Errorf("figure write: unsupported format %q; supported: %s", format, supportedExtensionsList())
	}

	tmp, err := os.CreateTemp("", "matplotlib-go-figure-*"+ext)
	if err != nil {
		return fmt.Errorf("figure write: create temporary output: %w", err)
	}
	tmpPath := tmp.Name()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("figure write: close temporary output: %w", err)
	}
	defer os.Remove(tmpPath)

	saveOpts := append(append([]render.SaveOption(nil), opts...), render.WithSaveFormat(strings.TrimPrefix(ext, ".")))
	if err := f.Save(tmpPath, saveOpts...); err != nil {
		return err
	}
	file, err := os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("figure write: reopen encoded output: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(w, file); err != nil {
		return fmt.Errorf("figure write: copy encoded output: %w", err)
	}
	return nil
}

// Image renders the figure with the registered PNG raster backend and returns
// a detached, correctly premultiplied RGBA image. Applications can import
// backends/all (or backends/agg directly) to register the built-in raster
// backend. Mutating the returned image does not affect later renders.
func (f *Figure) Image() (*image.RGBA, error) {
	if f == nil {
		return nil, fmt.Errorf("figure image: nil figure")
	}
	r, err := f.newOutputRenderer(".png")
	if err != nil {
		return nil, fmt.Errorf("figure image: %w", err)
	}
	DrawFigureWithOptions(f, r, DrawOptions{Transparent: !f.RC.Figure.FrameOn})

	if exporter, ok := r.(nrgbaExporter); ok {
		src := exporter.ImageNRGBA()
		if src == nil {
			return nil, fmt.Errorf("figure image: raster backend returned no image")
		}
		return detachedRGBA(src), nil
	}

	exporter, ok := r.(render.RGBAExporter)
	if !ok {
		return nil, fmt.Errorf("figure image: PNG renderer does not expose an RGBA image")
	}
	src := exporter.Image()
	if src == nil {
		return nil, fmt.Errorf("figure image: raster backend returned no image")
	}
	return detachedRGBA(src), nil
}

func detachedRGBA(src image.Image) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
	return dst
}

func normalizeOutputFormat(format string) string {
	format = strings.TrimSpace(strings.ToLower(format))
	if format == "" {
		return ""
	}
	if !strings.HasPrefix(format, ".") {
		format = "." + format
	}
	return format
}

func (f *Figure) newOutputRenderer(ext string, opts ...render.SaveOption) (render.Renderer, error) {
	width, height := f.outputCanvasSize(ext, opts...)
	background := f.RC.FigureBackground()
	saveOptions := render.ResolveSaveOptions(opts...)
	resolved := resolveSaveFigureOptions(f, &saveOptions.Figure)
	switch {
	case !f.RC.Figure.FrameOn || resolved.transparent:
		background = render.Color{}
	case resolved.hasFacecolor:
		background = resolved.facecolor
	}
	ext = normalizeOutputFormat(ext)
	figureOutputRenderers.RLock()
	factory := figureOutputRenderers.factories[ext]
	figureOutputRenderers.RUnlock()
	if factory == nil && ext == ".eps" {
		figureOutputRenderers.RLock()
		factory = figureOutputRenderers.factories[".ps"]
		figureOutputRenderers.RUnlock()
	}
	if factory == nil {
		return nil, fmt.Errorf(
			"figure save: no renderer registered for %q; import github.com/cwbudde/matplotlib-go/backends/all or the desired backend package",
			ext,
		)
	}
	return factory(width, height, background)
}

func (f *Figure) outputCanvasSize(ext string, opts ...render.SaveOption) (width, height int) {
	if f == nil {
		return 0, 0
	}
	dpi := f.RC.DPI
	saveOptions := render.ResolveSaveOptions(opts...)
	resolved := resolveSaveFigureOptions(f, &saveOptions.Figure)
	if resolved.dpi > 0 {
		dpi = resolved.dpi
	}
	if isVectorSaveExtension(ext) {
		dpi = 72
	}
	if f.RC.DPI <= 0 || dpi <= 0 {
		return f.CanvasSize()
	}
	return int(math.Round(f.SizePx.X * dpi / f.RC.DPI)), int(math.Round(f.SizePx.Y * dpi / f.RC.DPI))
}

func isVectorSaveExtension(ext string) bool {
	switch normalizeOutputFormat(ext) {
	case ".pdf", ".svg", ".ps", ".eps", ".pgf":
		return true
	default:
		return false
	}
}
