package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/optarg"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// AnchoredDrawingAreaOptions configures an anchored fixed-size drawing area.
type AnchoredDrawingAreaOptions struct {
	Location LegendLocation
	Locator  AnchoredBoxLocator
	Padding  float64
	Inset    float64
	FrameOn  *bool
	Clip     bool

	BackgroundColor render.Color
	BorderColor     render.Color
	BorderWidth     float64
}

// AnchoredDrawingArea draws local display-space paths inside an anchored box.
type AnchoredDrawingArea struct {
	Width  float64
	Height float64

	Location LegendLocation
	Locator  AnchoredBoxLocator
	Padding  float64
	Inset    float64
	FrameOn  bool
	Clip     bool

	BackgroundColor render.Color
	BorderColor     render.Color
	BorderWidth     float64

	paths []anchoredDrawingAreaPath
	z     float64
}

type anchoredDrawingAreaPath struct {
	path  geom.Path
	paint render.Paint
}

// AddAnchoredDrawingArea appends a fixed-size drawing area inside an axes.
func (a *Axes) AddAnchoredDrawingArea(width, height float64, opts ...AnchoredDrawingAreaOptions) *AnchoredDrawingArea {
	rc := style.CurrentDefaults()
	if a != nil {
		rc = a.resolvedRC()
	}
	area := newAnchoredDrawingArea(width, height, rc, opts...)
	if a != nil {
		a.Add(area)
	}
	return area
}

func newAnchoredDrawingArea(width, height float64, rc style.RC, opts ...AnchoredDrawingAreaOptions) *AnchoredDrawingArea {
	cfg := AnchoredDrawingAreaOptions{
		Location:        LegendUpperRight,
		Padding:         -1,
		Inset:           -1,
		BackgroundColor: rc.LegendBackground,
		BorderColor:     rc.LegendBorderColor,
		BorderWidth:     1,
	}
	frameOn := true
	cfg.FrameOn = &frameOn
	if opt, ok := optarg.Optional("anchored drawing area", opts); ok {
		cfg.Location = opt.Location
		cfg.Locator = opt.Locator
		if opt.Padding >= 0 {
			cfg.Padding = opt.Padding
		}
		if opt.Inset >= 0 {
			cfg.Inset = opt.Inset
		}
		if opt.FrameOn != nil {
			cfg.FrameOn = cloneBool(opt.FrameOn)
		}
		cfg.Clip = opt.Clip
		if opt.BackgroundColor != (render.Color{}) {
			cfg.BackgroundColor = opt.BackgroundColor
		}
		if opt.BorderColor != (render.Color{}) {
			cfg.BorderColor = opt.BorderColor
		}
		if opt.BorderWidth > 0 {
			cfg.BorderWidth = opt.BorderWidth
		}
	}
	return &AnchoredDrawingArea{
		Width:           width,
		Height:          height,
		Location:        cfg.Location,
		Locator:         cfg.Locator,
		Padding:         cfg.Padding,
		Inset:           cfg.Inset,
		FrameOn:         cfg.FrameOn == nil || *cfg.FrameOn,
		Clip:            cfg.Clip,
		BackgroundColor: cfg.BackgroundColor,
		BorderColor:     cfg.BorderColor,
		BorderWidth:     cfg.BorderWidth,
		z:               950,
	}
}

// AddPath appends a local path to the drawing area. Local coordinates use the
// Matplotlib DrawingArea convention: (0,0) is the lower-left corner.
func (a *AnchoredDrawingArea) AddPath(path geom.Path, paint render.Paint) *AnchoredDrawingArea {
	if a == nil {
		return nil
	}
	a.paths = append(a.paths, anchoredDrawingAreaPath{
		path:  clonePath(path),
		paint: clonePaint(paint),
	})
	return a
}

func (a *AnchoredDrawingArea) Draw(r render.Renderer, ctx *DrawContext) {
	if a == nil || r == nil || ctx == nil {
		return
	}
	layout := a.layout(ctx)
	if layout.empty {
		return
	}
	if a.FrameOn {
		r.Path(pixelRectPath(layout.frame), &render.Paint{
			Fill:      a.BackgroundColor,
			Stroke:    a.BorderColor,
			LineWidth: pointsToPixels(ctx.RC, a.BorderWidth),
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
			Snap:      render.SnapAuto,
		})
	}
	drawChildren := func() {
		for _, child := range a.paths {
			paint := child.paint
			r.Path(localDrawingAreaPath(child.path, layout.content, layout.scale), &paint)
		}
	}
	if a.Clip {
		r.Save()
		r.ClipRect(layout.content)
		drawChildren()
		r.Restore()
		return
	}
	drawChildren()
}

// Bounds returns an empty rect so anchored drawing areas do not affect autoscaling.
func (a *AnchoredDrawingArea) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the anchored drawing area z-order.
func (a *AnchoredDrawingArea) Z() float64 { return a.z }

func (a *AnchoredDrawingArea) boxRect(_ render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	layout := a.layout(ctx)
	if layout.empty {
		return geom.Rect{}, false
	}
	return layout.frame, true
}

type anchoredDrawingAreaLayout struct {
	frame   geom.Rect
	content geom.Rect
	scale   float64
	empty   bool
}

func (a *AnchoredDrawingArea) layout(ctx *DrawContext) anchoredDrawingAreaLayout {
	if a == nil || ctx == nil || a.Width <= 0 || a.Height <= 0 {
		return anchoredDrawingAreaLayout{empty: true}
	}
	padding := a.resolvedPadding(ctx)
	scale := pointsToPixels(ctx.RC, 1)
	width := a.Width * scale
	height := a.Height * scale
	frame := resolveAnchoredBoxRect(a.Locator, ctx.Clip, width+padding*2, height+padding*2, a.Location, a.resolvedInset(ctx))
	return anchoredDrawingAreaLayout{
		frame: frame,
		content: geom.Rect{
			Min: geom.Pt{X: frame.Min.X + padding, Y: frame.Min.Y + padding},
			Max: geom.Pt{X: frame.Max.X - padding, Y: frame.Max.Y - padding},
		},
		scale: scale,
	}
}

func (a *AnchoredDrawingArea) resolvedPadding(ctx *DrawContext) float64 {
	if a != nil && a.Padding >= 0 {
		return a.Padding
	}
	return pointsToPixels(ctx.RC, 0.1*ctx.RC.LegendSize())
}

func (a *AnchoredDrawingArea) resolvedInset(ctx *DrawContext) float64 {
	if a != nil && a.Inset >= 0 {
		return a.Inset
	}
	return pointsToPixels(ctx.RC, 0.1*ctx.RC.LegendSize())
}

func localDrawingAreaPath(path geom.Path, box geom.Rect, scale float64) geom.Path {
	if scale <= 0 {
		scale = 1
	}
	out := clonePath(path)
	for i, pt := range out.V {
		// Display space is y-up: the drawing area's local origin is its lower-left
		// (Matplotlib DrawingArea convention), so local Y grows upward from box.Min.Y.
		out.V[i] = geom.Pt{
			X: box.Min.X + pt.X*scale,
			Y: box.Min.Y + pt.Y*scale,
		}
	}
	return out
}

func clonePath(path geom.Path) geom.Path {
	return geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: append([]geom.Pt(nil), path.V...),
	}
}

func clonePaint(paint render.Paint) render.Paint {
	paint.Dashes = append([]float64(nil), paint.Dashes...)
	paint.PathEffects = cloneRenderPathEffects(paint.PathEffects)
	return paint
}
