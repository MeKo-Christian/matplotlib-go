package pgf

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cwbudde/matplotlib-go/backends/internal/mixedraster"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// New creates a PGF renderer with a point-sized canvas.
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
		width:      float64(width),
		height:     float64(height),
		background: background,
		resolution: 72,
		pgfOpts:    render.DefaultPGFOptions(),
	}, nil
}

// SetResolution implements render.DPIAware.
func (r *Renderer) SetResolution(dpi uint) {
	if dpi == 0 {
		dpi = 72
	}
	r.resolution = dpi
}

// Clear replaces the page background used by the next drawing session.
func (r *Renderer) Clear(c render.Color) {
	if r != nil {
		r.background = c
	}
}

// SetPageSize sets the physical PGF page in PostScript points.
func (r *Renderer) SetPageSize(widthPoints, heightPoints float64) {
	if r != nil && widthPoints > 0 && heightPoints > 0 {
		r.width = widthPoints
		r.height = heightPoints
	}
}

// Begin starts a drawing session.
func (r *Renderer) Begin(viewport geom.Rect) error {
	if r.began {
		return errors.New("pgf: Begin called twice")
	}
	r.began = true
	r.viewport = viewport
	r.content.Reset()
	r.colorDefs.Reset()
	r.document = nil
	r.stack = r.stack[:0]
	r.clipRect = nil
	r.clipPaths = nil
	r.raster = nil
	r.colorNames = map[string]string{}
	r.pathIDs = map[string]string{}
	r.shadingCounter = 0

	r.writeDocumentComments(&r.content)
	r.content.WriteString("\\begingroup\n")
	r.content.WriteString("\\begin{pgfpicture}\n")
	// Display coordinates use PostScript points (72 per inch), while TeX's pt
	// is 1/72.27 inch. Scale geometry once so the page remains physically exact.
	r.content.WriteString("\\pgftransformscale{1.00375}\n")
	fmt.Fprintf(&r.content, "\\pgfpathrectangle{\\pgfpoint{0pt}{0pt}}{\\pgfpoint{%spt}{%spt}}\n",
		shortFloat(r.width), shortFloat(r.height))
	r.content.WriteString("\\pgfusepath{use as bounding box}\n")
	// All \definecolor declarations are injected here, at the pgfpicture's
	// outermost group, so every nested \pgfscope can see them.
	r.content.WriteString(colorDefsPlaceholder)
	// PGF/TeX is natively y-up (origin bottom-left), matching the matplotlib-go
	// y-up display space exactly. Like Matplotlib's PGF backend (flipy() is
	// False), no device flip is emitted; draws use display coordinates directly.
	if r.background.A > 0 {
		writeFillOpacity(&r.content, r.background.A)
		writeFillColor(&r.content, r.colorName(r.background))
		fmt.Fprintf(&r.content, "\\pgfpathrectangle{\\pgfpoint{0pt}{0pt}}{\\pgfpoint{%spt}{%spt}}\n",
			shortFloat(r.width), shortFloat(r.height))
		r.content.WriteString("\\pgfusepath{fill}\n")
	}
	return nil
}

// End finalizes the PGF document.
func (r *Renderer) End() error {
	if !r.began {
		return errors.New("pgf: End called before Begin")
	}
	r.began = false
	r.content.WriteString("\\end{pgfpicture}\n")
	r.content.WriteString("\\endgroup\n")
	doc := strings.Replace(r.content.String(), colorDefsPlaceholder, r.colorDefs.String(), 1)
	r.document = []byte(doc)
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
		r.content.WriteString("\\pgfscope\n")
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
		r.content.WriteString("\\endpgfscope\n")
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
	fmt.Fprintf(&r.content, "\\pgfpathrectangle{\\pgfpoint{%spt}{%spt}}{\\pgfpoint{%spt}{%spt}}\n",
		shortFloat(rect.Min.X), shortFloat(rect.Min.Y), shortFloat(rect.W()), shortFloat(rect.H()))
	r.content.WriteString("\\pgfusepath{clip}\n")
}

// ClipPath installs an arbitrary path clip.
func (r *Renderer) ClipPath(path geom.Path) {
	if rr := r.activeRaster(); rr != nil {
		rr.ClipPath(path)
		return
	}
	if !r.began {
		return
	}
	if !writePathOps(&r.content, path) {
		return
	}
	r.clipPaths = append(r.clipPaths, mixedraster.ClonePath(path))
	r.content.WriteString("\\pgfusepath{clip}\n")
}

// ClipPathTransformed installs a path clip after applying an affine transform
// to the path, matching Matplotlib's PGF clip handling (the path is converted
// through its transform before \pgfusepath{clip}). Implements
// render.ClipPathTransformer.
func (r *Renderer) ClipPathTransformed(path geom.Path, transform geom.Affine) {
	if rr := r.activeRaster(); rr != nil {
		if ct, ok := rr.(render.ClipPathTransformer); ok {
			ct.ClipPathTransformed(path, transform)
		} else {
			rr.ClipPath(path.Transformed(normalizedAffine(transform)))
		}
		return
	}
	if !r.began {
		return
	}
	transformed := path.Transformed(normalizedAffine(transform))
	if !writePathOps(&r.content, transformed) {
		return
	}
	r.clipPaths = append(r.clipPaths, mixedraster.ClonePath(transformed))
	r.content.WriteString("\\pgfusepath{clip}\n")
}
