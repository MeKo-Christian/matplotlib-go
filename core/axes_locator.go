package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// AxesLocator resolves an axes rectangle in figure-fraction coordinates at draw time.
type AxesLocator interface {
	Rect(fig *Figure, ax *Axes, r render.Renderer) geom.Rect
}

// AxesLocatorFunc adapts a function into an AxesLocator.
type AxesLocatorFunc func(fig *Figure, ax *Axes, r render.Renderer) geom.Rect

func (f AxesLocatorFunc) Rect(fig *Figure, ax *Axes, r render.Renderer) geom.Rect {
	if f == nil {
		if ax == nil {
			return geom.Rect{}
		}
		return ax.RectFraction
	}
	return f(fig, ax, r)
}

// SetAxesLocator attaches a draw-time layout locator to the axes.
func (a *Axes) SetAxesLocator(locator AxesLocator) {
	if a == nil {
		return
	}
	a.axesLocator = locator
}

// AxesLocator reports the draw-time locator attached to this axes, if any.
func (a *Axes) AxesLocator() AxesLocator {
	if a == nil {
		return nil
	}
	return a.axesLocator
}

type insetAxesOptions struct {
	style      []style.Option
	projection string
	shareX     *Axes
	shareY     *Axes
}

// InsetAxesOption configures inset axes creation.
type InsetAxesOption func(*insetAxesOptions)

// WithInsetStyle applies style options to an inset axes.
func WithInsetStyle(opts ...style.Option) InsetAxesOption {
	return func(cfg *insetAxesOptions) {
		cfg.style = append(cfg.style, opts...)
	}
}

// WithInsetProjection selects a named projection for an inset axes.
func WithInsetProjection(name string) InsetAxesOption {
	return func(cfg *insetAxesOptions) {
		cfg.projection = name
	}
}

// WithInsetSharedX makes an inset axes share x-axis state with peer.
func WithInsetSharedX(peer *Axes) InsetAxesOption {
	return func(cfg *insetAxesOptions) {
		cfg.shareX = peer
	}
}

// WithInsetSharedY makes an inset axes share y-axis state with peer.
func WithInsetSharedY(peer *Axes) InsetAxesOption {
	return func(cfg *insetAxesOptions) {
		cfg.shareY = peer
	}
}

// WithInsetSharedAxes makes an inset axes share both axis states with peer.
func WithInsetSharedAxes(peer *Axes) InsetAxesOption {
	return func(cfg *insetAxesOptions) {
		cfg.shareX = peer
		cfg.shareY = peer
	}
}

// InsetAxes creates an axes whose bounds are expressed in the parent axes'
// normalized coordinate system.
func (a *Axes) InsetAxes(bounds geom.Rect, opts ...InsetAxesOption) *Axes {
	if a == nil || a.figure == nil || bounds.Max.X <= bounds.Min.X || bounds.Max.Y <= bounds.Min.Y {
		return nil
	}

	cfg := insetAxesOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	locator := &axesFractionLocator{
		parent: a,
		bounds: bounds,
	}
	initial := locator.Rect(a.figure, nil, nil)

	var inset *Axes
	if cfg.projection != "" {
		var err error
		inset, err = a.figure.AddAxesProjection(initial, cfg.projection, cfg.style...)
		if err != nil {
			return nil
		}
	} else {
		inset = a.figure.AddAxes(initial, cfg.style...)
	}
	if inset == nil {
		return nil
	}
	inset.SetAxesLocator(locator)
	applyInsetSharing(inset, cfg)
	return inset
}

// ZoomedInset creates an inset axes with the requested data limits and adds an
// indicator linking the parent data region to the inset.
func (a *Axes) ZoomedInset(bounds geom.Rect, xlim, ylim [2]float64, opts ...InsetAxesOption) (*Axes, *InsetIndicator) {
	inset := a.InsetAxes(bounds, opts...)
	if inset == nil {
		return nil, nil
	}
	inset.SetXLim(xlim[0], xlim[1])
	inset.SetYLim(ylim[0], ylim[1])

	indicator := NewInsetIndicator(inset, xlim, ylim)
	a.Add(indicator)
	return inset, indicator
}

type axesFractionLocator struct {
	parent *Axes
	bounds geom.Rect
}

func (l *axesFractionLocator) Rect(fig *Figure, _ *Axes, _ render.Renderer) geom.Rect {
	if fig == nil || l == nil || l.parent == nil {
		return geom.Rect{}
	}
	parentPx := l.parent.adjustedLayout(fig)
	childPx := composeRect(parentPx, l.bounds)
	return pixelRectToFigureFraction(childPx, fig.DisplayRect())
}

func applyInsetSharing(ax *Axes, cfg insetAxesOptions) {
	if ax == nil {
		return
	}
	if cfg.shareX != nil {
		root := cfg.shareX.xScaleRoot()
		ax.shareX = root
		ax.XAxis = root.XAxis
	}
	if cfg.shareY != nil {
		root := cfg.shareY.yScaleRoot()
		ax.shareY = root
		ax.YAxis = root.YAxis
	}
}

func syncAxesLocators(fig *Figure, r render.Renderer) {
	if fig == nil {
		return
	}
	for _, ax := range fig.Children {
		if ax == nil || ax.axesLocator == nil {
			continue
		}
		rect := ax.axesLocator.Rect(fig, ax, r)
		if rect.Max.X <= rect.Min.X || rect.Max.Y <= rect.Min.Y {
			continue
		}
		ax.RectFraction = rect
	}
}

func pixelRectToFigureFraction(px, vp geom.Rect) geom.Rect {
	if vp.W() == 0 || vp.H() == 0 {
		return geom.Rect{}
	}
	return geom.Rect{
		Min: geom.Pt{
			X: (px.Min.X - vp.Min.X) / vp.W(),
			Y: (px.Min.Y - vp.Min.Y) / vp.H(),
		},
		Max: geom.Pt{
			X: (px.Max.X - vp.Min.X) / vp.W(),
			Y: (px.Max.Y - vp.Min.Y) / vp.H(),
		},
	}
}

// InsetIndicator draws a data-space zoom rectangle and connector lines to an inset axes.
type InsetIndicator struct {
	Inset     *Axes
	XLim      [2]float64
	YLim      [2]float64
	Color     render.Color
	LineWidth float64
	z         float64
}

// NewInsetIndicator creates a zoom indicator for an inset axes.
func NewInsetIndicator(inset *Axes, xlim, ylim [2]float64) *InsetIndicator {
	return &InsetIndicator{
		Inset:     inset,
		XLim:      xlim,
		YLim:      ylim,
		Color:     render.Color{R: 0.15, G: 0.15, B: 0.15, A: 0.9},
		LineWidth: 1,
		z:         900,
	}
}

func (i *InsetIndicator) Draw(render.Renderer, *DrawContext) {}

// DrawOverlay draws the indicator after the parent axes clip has been removed.
func (i *InsetIndicator) DrawOverlay(r render.Renderer, ctx *DrawContext) {
	if i == nil || i.Inset == nil || i.Inset.figure == nil || ctx == nil {
		return
	}
	rect := normalizedLimitRect(i.XLim, i.YLim)
	corners := []geom.Pt{
		ctx.DataToPixel.Apply(geom.Pt{X: rect.Min.X, Y: rect.Min.Y}),
		ctx.DataToPixel.Apply(geom.Pt{X: rect.Max.X, Y: rect.Min.Y}),
		ctx.DataToPixel.Apply(geom.Pt{X: rect.Max.X, Y: rect.Max.Y}),
		ctx.DataToPixel.Apply(geom.Pt{X: rect.Min.X, Y: rect.Max.Y}),
	}

	paint := render.Paint{
		Stroke:    i.Color,
		LineWidth: i.LineWidth,
		LineCap:   render.CapButt,
		LineJoin:  render.JoinMiter,
	}

	zoomPath := geom.Path{}
	zoomPath.MoveTo(corners[0])
	for _, pt := range corners[1:] {
		zoomPath.LineTo(pt)
	}
	zoomPath.Close()
	r.Path(zoomPath, &paint)

	insetRect := i.Inset.adjustedLayout(i.Inset.figure)
	insetCorners := []geom.Pt{
		{X: insetRect.Min.X, Y: insetRect.Max.Y},
		{X: insetRect.Max.X, Y: insetRect.Max.Y},
		{X: insetRect.Max.X, Y: insetRect.Min.Y},
		{X: insetRect.Min.X, Y: insetRect.Min.Y},
	}
	for _, pair := range bestInsetConnectorPairs(corners, insetCorners) {
		path := geom.Path{}
		path.MoveTo(corners[pair[0]])
		path.LineTo(insetCorners[pair[1]])
		r.Path(path, &paint)
	}
}

func (i *InsetIndicator) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

func (i *InsetIndicator) Z() float64 {
	if i == nil {
		return 0
	}
	return i.z
}

func normalizedLimitRect(xlim, ylim [2]float64) geom.Rect {
	x0, x1 := xlim[0], xlim[1]
	y0, y1 := ylim[0], ylim[1]
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	if y1 < y0 {
		y0, y1 = y1, y0
	}
	return geom.Rect{Min: geom.Pt{X: x0, Y: y0}, Max: geom.Pt{X: x1, Y: y1}}
}

func bestInsetConnectorPairs(zoomCorners, insetCorners []geom.Pt) [][2]int {
	if len(zoomCorners) != 4 || len(insetCorners) != 4 {
		return nil
	}
	best := [][2]int{{0, 0}, {2, 2}}
	bestScore := math.Inf(1)
	candidates := [][][2]int{
		{{0, 0}, {2, 2}},
		{{1, 1}, {3, 3}},
		{{0, 3}, {2, 1}},
		{{1, 2}, {3, 0}},
	}
	for _, candidate := range candidates {
		score := 0.0
		for _, pair := range candidate {
			score += pointDistanceSquared(zoomCorners[pair[0]], insetCorners[pair[1]])
		}
		if score < bestScore {
			bestScore = score
			best = candidate
		}
	}
	return best
}

func pointDistanceSquared(a, b geom.Pt) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return dx*dx + dy*dy
}
