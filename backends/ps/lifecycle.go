package ps

import (
	"errors"
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends/internal/mixedraster"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// New creates a PostScript renderer with a point-sized page.
func New(width, height int, background render.Color) (*Renderer, error) {
	if width <= 0 {
		width = 640
	}
	if height <= 0 {
		height = 480
	}
	if background == (render.Color{}) {
		background = render.Color{R: 1, G: 1, B: 1, A: 1}
	}
	return &Renderer{
		width:      width,
		height:     height,
		background: background,
		resolution: 72,
		psOpts:     render.DefaultPSOptions(),
	}, nil
}

// SetResolution implements render.DPIAware.
func (r *Renderer) SetResolution(dpi uint) {
	if dpi == 0 {
		dpi = 72
	}
	r.resolution = dpi
}

// SetPSOptions implements render.PSOptionSetter.
func (r *Renderer) SetPSOptions(opts render.PSOptions) {
	r.psOpts = opts
}

// Begin starts a drawing session for the given viewport.
func (r *Renderer) Begin(viewport geom.Rect) error {
	if r.began {
		return errors.New("ps: Begin called twice")
	}
	r.began = true
	r.viewport = viewport
	r.content.Reset()
	r.document = nil
	r.stack = r.stack[:0]
	r.clipRect = nil
	r.clipPaths = nil
	r.raster = nil
	r.markerIDs = map[string]string{}
	r.imageIDs = map[string]string{}
	r.lastFontKey = ""

	// PostScript is natively y-up (origin bottom-left), matching the matplotlib-go
	// y-up display space exactly. Like Matplotlib's PS backend (flipy() is False),
	// no device flip is emitted; draws use display coordinates directly. The
	// gsave/grestore pair brackets the page content.
	r.content.WriteString("gsave\n")
	if r.background.A > 0 {
		writeFillColor(&r.content, r.background)
		fmt.Fprintf(
			&r.content, "newpath 0 0 moveto %s 0 lineto %s %s lineto 0 %s lineto closepath fill\n",
			shortFloat(float64(r.width)),
			shortFloat(float64(r.width)),
			shortFloat(float64(r.height)),
			shortFloat(float64(r.height)),
		)
	}
	return nil
}

// End finalizes the current drawing session.
func (r *Renderer) End() error {
	if !r.began {
		return errors.New("ps: End called before Begin")
	}
	r.began = false
	r.content.WriteString("grestore\n")
	r.document = buildDocument(r.width, r.height, r.content.String(), false)
	return nil
}

// Save pushes graphics state.
func (r *Renderer) Save() {
	if rr := r.activeRaster(); rr != nil {
		rr.Save()
		return
	}
	r.stack = append(r.stack, state{
		inContent: r.began,
		clipRect:  cloneRectPtr(r.clipRect),
		clipPaths: mixedraster.ClonePaths(r.clipPaths),
	})
	if r.began {
		r.content.WriteString("gsave\n")
	}
}

// Restore pops graphics state.
func (r *Renderer) Restore() {
	if rr := r.activeRaster(); rr != nil {
		rr.Restore()
		return
	}
	if len(r.stack) == 0 {
		return
	}
	top := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	r.clipRect = top.clipRect
	r.clipPaths = mixedraster.ClonePaths(top.clipPaths)
	if top.inContent && r.began {
		r.content.WriteString("grestore\n")
	}
}

// ClipRect installs a rectangular clip.
func (r *Renderer) ClipRect(rect geom.Rect) {
	if rr := r.activeRaster(); rr != nil {
		rr.ClipRect(rect)
		return
	}
	if !r.began {
		return
	}
	normalized := normalizeRect(rect)
	if r.clipRect == nil {
		r.clipRect = &normalized
	} else {
		intersected := r.clipRect.Intersect(normalized)
		r.clipRect = &intersected
	}
	fmt.Fprintf(
		&r.content, "newpath %s %s moveto %s %s lineto %s %s lineto %s %s lineto closepath clip newpath\n",
		shortFloat(rect.Min.X), shortFloat(rect.Min.Y),
		shortFloat(rect.Max.X), shortFloat(rect.Min.Y),
		shortFloat(rect.Max.X), shortFloat(rect.Max.Y),
		shortFloat(rect.Min.X), shortFloat(rect.Max.Y),
	)
}

// ClipPath installs an arbitrary path clip.
func (r *Renderer) ClipPath(p geom.Path) {
	if rr := r.activeRaster(); rr != nil {
		rr.ClipPath(p)
		return
	}
	if !r.began {
		return
	}
	if !writePathOps(&r.content, p) {
		return
	}
	r.clipPaths = append(r.clipPaths, mixedraster.ClonePath(p))
	r.content.WriteString("clip newpath\n")
}
