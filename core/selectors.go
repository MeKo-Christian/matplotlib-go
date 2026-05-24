package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// SpanSelectorCallback receives the committed span in data coordinates.
type SpanSelectorCallback func(*SpanSelector, float64, float64)

// RectangleSelectorCallback receives the committed rectangle bounds in data coordinates.
type RectangleSelectorCallback func(*RectangleSelector, geom.Rect)

// EllipseSelectorCallback receives the ellipse bounding box in data coordinates.
type EllipseSelectorCallback func(*EllipseSelector, geom.Rect)

// PolygonSelectorCallback receives the polygon vertices in data coordinates.
type PolygonSelectorCallback func(*PolygonSelector, []geom.Pt)

// LassoSelectorCallback receives the lasso path in data coordinates.
type LassoSelectorCallback func(*LassoSelector, []geom.Pt)

// SpanSelectorOptions configures a span selector.
type SpanSelectorOptions struct {
	Orientation string
	Color       render.Color
	FillColor   render.Color
	LineWidth   float64
	Min         *float64
	Max         *float64
}

// RectangleSelectorOptions configures a rectangle selector.
type RectangleSelectorOptions struct {
	EdgeColor render.Color
	FillColor render.Color
	LineWidth float64
}

// EllipseSelectorOptions configures an ellipse selector.
type EllipseSelectorOptions struct {
	EdgeColor render.Color
	FillColor render.Color
	LineWidth float64
}

// PolygonSelectorOptions configures a polygon selector.
type PolygonSelectorOptions struct {
	EdgeColor render.Color
	FillColor render.Color
	LineWidth float64
}

// LassoSelectorOptions configures a lasso selector.
type LassoSelectorOptions struct {
	LineColor render.Color
	LineWidth float64
}

// CursorOptions configures a hover cursor helper.
type CursorOptions struct {
	Color      render.Color
	LineWidth  float64
	HorizOn    *bool
	VerticalOn *bool
}

// MultiCursorOptions configures a multi-axis hover cursor helper.
type MultiCursorOptions struct {
	Color      render.Color
	LineWidth  float64
	HorizOn    *bool
	VerticalOn *bool
}

// SpanSelector lets users select a data span along one axis.
type SpanSelector struct {
	Orientation string
	Start       float64
	End         float64
	Color       render.Color
	FillColor   render.Color
	LineWidth   float64
	Active      bool

	onSelect widgetCallbackRegistry[SpanSelectorCallback]
	z        float64
}

// RectangleSelector stores a selected rectangular region in data coordinates.
type RectangleSelector struct {
	Min       geom.Pt
	Max       geom.Pt
	EdgeColor render.Color
	FillColor render.Color
	LineWidth float64
	Active    bool

	onSelect widgetCallbackRegistry[RectangleSelectorCallback]
	z        float64
}

// EllipseSelector stores a selected ellipse region by data bounding box.
type EllipseSelector struct {
	Min       geom.Pt
	Max       geom.Pt
	EdgeColor render.Color
	FillColor render.Color
	LineWidth float64
	Active    bool

	onSelect widgetCallbackRegistry[EllipseSelectorCallback]
	z        float64
}

// PolygonSelector stores a selected polygon path in data coordinates.
type PolygonSelector struct {
	Points    []geom.Pt
	EdgeColor render.Color
	FillColor render.Color
	LineWidth float64
	Closed    bool
	Active    bool

	onSelect widgetCallbackRegistry[PolygonSelectorCallback]
	z        float64
}

// LassoSelector stores a drawn freehand path in data coordinates.
type LassoSelector struct {
	Points    []geom.Pt
	LineColor render.Color
	LineWidth float64
	Active    bool
	Tracking  bool

	onSelect widgetCallbackRegistry[LassoSelectorCallback]
	z        float64
}

// Cursor draws cursor cross-lines and is updated by hover events.
type Cursor struct {
	X, Y       float64
	Color      render.Color
	LineWidth  float64
	Horizontal bool
	Vertical   bool

	axis *Axes
	show bool

	z float64
}

// MultiCursor draws shared cross-lines across multiple axes.
type MultiCursor struct {
	Axes       []*Axes
	Color      render.Color
	LineWidth  float64
	Horizontal bool
	Vertical   bool

	FigureX, FigureY float64
	show             bool

	z float64
}

// SpanSelector creates a span selector bound to the axes.
func (a *Axes) SpanSelector(orientation string, opts ...SpanSelectorOptions) *SpanSelector {
	if a == nil {
		return nil
	}
	config := SpanSelectorOptions{
		Orientation: "horizontal",
		Color:       render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
		FillColor:   render.Color{R: 0.16, G: 0.42, B: 0.76, A: 0.18},
		LineWidth:   1.2,
	}
	if len(opts) > 0 {
		config = mergeSpanSelectorOptions(config, opts[0])
	}
	sel := &SpanSelector{
		Orientation: normalizeSpanOrientation(config.Orientation),
		Color:       config.Color,
		FillColor:   config.FillColor,
		LineWidth:   config.LineWidth,
		z:           1200,
	}
	if config.Min != nil || config.Max != nil {
		minV := sel.Start
		maxV := sel.End
		if config.Min != nil {
			minV = *config.Min
		}
		if config.Max != nil {
			maxV = *config.Max
		}
		sel.SetSpan(minV, maxV)
	}
	a.AddWidget(sel)
	return sel
}

// RectangleSelector creates a rectangle selector bound to the axes.
func (a *Axes) RectangleSelector(opts ...RectangleSelectorOptions) *RectangleSelector {
	if a == nil {
		return nil
	}
	config := RectangleSelectorOptions{
		EdgeColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
		FillColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 0.18},
		LineWidth: 1.2,
	}
	if len(opts) > 0 {
		config = mergeRectangleSelectorOptions(config, opts[0])
	}
	sel := &RectangleSelector{
		EdgeColor: config.EdgeColor,
		FillColor: config.FillColor,
		LineWidth: config.LineWidth,
		z:         1200,
	}
	a.AddWidget(sel)
	return sel
}

// EllipseSelector creates an ellipse selector bound to the axes.
func (a *Axes) EllipseSelector(opts ...EllipseSelectorOptions) *EllipseSelector {
	if a == nil {
		return nil
	}
	config := EllipseSelectorOptions{
		EdgeColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
		FillColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 0.18},
		LineWidth: 1.2,
	}
	if len(opts) > 0 {
		config = mergeEllipseSelectorOptions(config, opts[0])
	}
	sel := &EllipseSelector{
		EdgeColor: config.EdgeColor,
		FillColor: config.FillColor,
		LineWidth: config.LineWidth,
		z:         1200,
	}
	a.AddWidget(sel)
	return sel
}

// PolygonSelector creates a polygon selector bound to the axes.
func (a *Axes) PolygonSelector(opts ...PolygonSelectorOptions) *PolygonSelector {
	if a == nil {
		return nil
	}
	config := PolygonSelectorOptions{
		EdgeColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
		FillColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 0.18},
		LineWidth: 1.2,
	}
	if len(opts) > 0 {
		config = mergePolygonSelectorOptions(config, opts[0])
	}
	sel := &PolygonSelector{
		EdgeColor: config.EdgeColor,
		FillColor: config.FillColor,
		LineWidth: config.LineWidth,
		z:         1200,
	}
	a.AddWidget(sel)
	return sel
}

// LassoSelector creates a lasso selector bound to the axes.
func (a *Axes) LassoSelector(opts ...LassoSelectorOptions) *LassoSelector {
	if a == nil {
		return nil
	}
	config := LassoSelectorOptions{
		LineColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
		LineWidth: 1.2,
	}
	if len(opts) > 0 {
		config = mergeLassoSelectorOptions(config, opts[0])
	}
	sel := &LassoSelector{
		LineColor: config.LineColor,
		LineWidth: config.LineWidth,
		z:         1200,
	}
	a.AddWidget(sel)
	return sel
}

// Cursor creates a hover cursor for this axis.
func (a *Axes) Cursor(opts ...CursorOptions) *Cursor {
	if a == nil {
		return nil
	}
	config := CursorOptions{
		Color:      render.Color{R: 0.24, G: 0.24, B: 0.24, A: 1},
		LineWidth:  0.9,
		HorizOn:    boolPtr(true),
		VerticalOn: boolPtr(true),
	}
	if len(opts) > 0 {
		config = mergeCursorOptions(config, opts[0])
	}
	c := &Cursor{
		axis:       a,
		Color:      config.Color,
		LineWidth:  config.LineWidth,
		Horizontal: boolValue(config.HorizOn, true),
		Vertical:   boolValue(config.VerticalOn, true),
		z:          1200,
	}
	a.AddWidget(c)
	return c
}

// MultiCursor creates a synchronized cross-axis cursor helper.
func (a *Axes) MultiCursor(axes ...*Axes) *MultiCursor {
	if a == nil {
		return nil
	}
	config := MultiCursorOptions{
		Color:      render.Color{R: 0.24, G: 0.24, B: 0.24, A: 1},
		LineWidth:  0.9,
		HorizOn:    boolPtr(true),
		VerticalOn: boolPtr(true),
	}
	selector := &MultiCursor{
		Axes:       dedupeAxes(append([]*Axes{a}, axes...)),
		Color:      config.Color,
		LineWidth:  config.LineWidth,
		Horizontal: boolValue(config.HorizOn, true),
		Vertical:   boolValue(config.VerticalOn, true),
		z:          1200,
	}
	a.AddWidget(selector)
	return selector
}

func (a *Axes) MultiCursorWithOptions(opts []MultiCursorOptions, axes ...*Axes) *MultiCursor {
	if a == nil {
		return nil
	}
	config := MultiCursorOptions{
		Color:      render.Color{R: 0.24, G: 0.24, B: 0.24, A: 1},
		LineWidth:  0.9,
		HorizOn:    boolPtr(true),
		VerticalOn: boolPtr(true),
	}
	if len(opts) > 0 {
		config = mergeMultiCursorOptions(config, opts[0])
	}
	mc := &MultiCursor{
		Axes:       dedupeAxes(append([]*Axes{a}, axes...)),
		Color:      config.Color,
		LineWidth:  config.LineWidth,
		Horizontal: boolValue(config.HorizOn, true),
		Vertical:   boolValue(config.VerticalOn, true),
		z:          1200,
	}
	a.AddWidget(mc)
	return mc
}

func (s *SpanSelector) OnSelect(cb SpanSelectorCallback) WidgetCallbackID {
	if s == nil || cb == nil {
		return 0
	}
	return s.onSelect.add(cb)
}

func (s *SpanSelector) RemoveOnSelect(id WidgetCallbackID) {
	if s == nil {
		return
	}
	s.onSelect.remove(id)
}

func (s *SpanSelector) SetSpan(minV, maxV float64) bool {
	if s == nil || !isFinite(minV) || !isFinite(maxV) {
		return false
	}
	if minV > maxV {
		minV, maxV = maxV, minV
	}
	changed := !s.Active || s.Start != minV || s.End != maxV
	s.Active = true
	s.Start = minV
	s.End = maxV
	return changed
}

func (s *SpanSelector) Move(delta float64) bool {
	if s == nil || !s.Active || !isFinite(delta) {
		return false
	}
	s.Start += delta
	s.End += delta
	return true
}

func (s *SpanSelector) Clear() bool {
	if s == nil || !s.Active {
		return false
	}
	s.Active = false
	s.Start = 0
	s.End = 0
	return true
}

func (s *SpanSelector) TriggerOnSelect() {
	if s == nil || !s.Active {
		return
	}
	minV, maxV := s.Start, s.End
	if minV > maxV {
		minV, maxV = maxV, minV
	}
	s.onSelect.each(func(cb SpanSelectorCallback) {
		cb(s, minV, maxV)
	})
}

func (s *SpanSelector) Draw(r render.Renderer, ctx *DrawContext) {
	if s == nil || r == nil || ctx == nil || !s.Active {
		return
	}
	minV, maxV := s.Start, s.End
	if minV > maxV {
		minV, maxV = maxV, minV
	}
	if s.Orientation == "vertical" {
		p1 := ctx.DataToPixel.Apply(geom.Pt{X: 0, Y: minV})
		p2 := ctx.DataToPixel.Apply(geom.Pt{X: 0, Y: maxV})
		rect := geom.Rect{
			Min: geom.Pt{X: ctx.Clip.Min.X, Y: math.Min(p1.Y, p2.Y)},
			Max: geom.Pt{X: ctx.Clip.Max.X, Y: math.Max(p1.Y, p2.Y)},
		}
		if rect.W() > 0 && rect.H() > 0 {
			r.Path(pixelRectPath(rect), &render.Paint{Fill: s.FillColor, Stroke: s.Color, LineWidth: s.LineWidth})
		}
		return
	}
	p1 := ctx.DataToPixel.Apply(geom.Pt{X: minV, Y: 0})
	p2 := ctx.DataToPixel.Apply(geom.Pt{X: maxV, Y: 0})
	rect := geom.Rect{
		Min: geom.Pt{X: math.Min(p1.X, p2.X), Y: ctx.Clip.Min.Y},
		Max: geom.Pt{X: math.Max(p1.X, p2.X), Y: ctx.Clip.Max.Y},
	}
	if rect.W() <= 0 || rect.H() <= 0 {
		return
	}
	r.Path(pixelRectPath(rect), &render.Paint{Fill: s.FillColor, Stroke: s.Color, LineWidth: s.LineWidth})
}

func (s *SpanSelector) Bounds(ctx *DrawContext) geom.Rect {
	if s == nil || !s.Active || ctx == nil {
		return geom.Rect{}
	}
	minV, maxV := s.Start, s.End
	if minV > maxV {
		minV, maxV = maxV, minV
	}
	if s.Orientation == "vertical" {
		p1 := ctx.DataToPixel.Apply(geom.Pt{X: 0, Y: minV})
		p2 := ctx.DataToPixel.Apply(geom.Pt{X: 0, Y: maxV})
		return geom.Rect{Min: geom.Pt{X: ctx.Clip.Min.X, Y: math.Min(p1.Y, p2.Y)}, Max: geom.Pt{X: ctx.Clip.Max.X, Y: math.Max(p1.Y, p2.Y)}}
	}
	p1 := ctx.DataToPixel.Apply(geom.Pt{X: minV, Y: 0})
	p2 := ctx.DataToPixel.Apply(geom.Pt{X: maxV, Y: 0})
	return geom.Rect{Min: geom.Pt{X: math.Min(p1.X, p2.X), Y: ctx.Clip.Min.Y}, Max: geom.Pt{X: math.Max(p1.X, p2.X), Y: ctx.Clip.Max.Y}}
}

func (s *SpanSelector) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if s == nil || !s.Active || ctx == nil {
		return false, PickInfo{}
	}
	value := p
	if s.Orientation == "vertical" {
		data, ok := (&ctx.DataToPixel).Invert(value)
		if !ok {
			return false, PickInfo{}
		}
		minV, maxV := s.Start, s.End
		if minV > maxV {
			minV, maxV = maxV, minV
		}
		return data.Y >= minV && data.Y <= maxV, PickInfo{}
	}
	data, ok := (&ctx.DataToPixel).Invert(value)
	if !ok {
		return false, PickInfo{}
	}
	minV, maxV := s.Start, s.End
	if minV > maxV {
		minV, maxV = maxV, minV
	}
	return data.X >= minV && data.X <= maxV, PickInfo{}
}

func (s *SpanSelector) Z() float64   { return s.z }
func (s *SpanSelector) WidgetLayer() {}

func (s *RectangleSelector) OnSelect(cb RectangleSelectorCallback) WidgetCallbackID {
	if s == nil || cb == nil {
		return 0
	}
	return s.onSelect.add(cb)
}

func (s *RectangleSelector) RemoveOnSelect(id WidgetCallbackID) {
	if s == nil {
		return
	}
	s.onSelect.remove(id)
}

func (s *RectangleSelector) SetBounds(minV, maxV geom.Pt) bool {
	if s == nil || !isFinite(minV.X) || !isFinite(minV.Y) || !isFinite(maxV.X) || !isFinite(maxV.Y) {
		return false
	}
	minV, maxV = normalizedRect(minV, maxV)
	changed := !s.Active || s.Min != minV || s.Max != maxV
	s.Active = true
	s.Min = minV
	s.Max = maxV
	return changed
}

func (s *RectangleSelector) MoveBy(delta geom.Pt) bool {
	if s == nil || !s.Active || !isFinite(delta.X) || !isFinite(delta.Y) {
		return false
	}
	s.Min.X += delta.X
	s.Min.Y += delta.Y
	s.Max.X += delta.X
	s.Max.Y += delta.Y
	return true
}

func (s *RectangleSelector) Clear() bool {
	if s == nil || !s.Active {
		return false
	}
	s.Active = false
	s.Min = geom.Pt{}
	s.Max = geom.Pt{}
	return true
}

func (s *RectangleSelector) TriggerOnSelect() {
	if s == nil || !s.Active {
		return
	}
	s.onSelect.each(func(cb RectangleSelectorCallback) {
		cb(s, s.BoundsRect())
	})
}

func (s *RectangleSelector) BoundsRect() geom.Rect {
	if s == nil || !s.Active {
		return geom.Rect{}
	}
	return geom.Rect{Min: s.Min, Max: s.Max}
}

func (s *RectangleSelector) Draw(r render.Renderer, ctx *DrawContext) {
	if s == nil || r == nil || ctx == nil || !s.Active {
		return
	}
	rect := s.displayRectFromData(ctx)
	if rect.W() <= 0 || rect.H() <= 0 {
		return
	}
	r.Path(pixelRectPath(rect), &render.Paint{Fill: s.FillColor, Stroke: s.EdgeColor, LineWidth: s.LineWidth})
}

func (s *RectangleSelector) Bounds(ctx *DrawContext) geom.Rect {
	if s == nil || !s.Active || ctx == nil {
		return geom.Rect{}
	}
	return s.displayRectFromData(ctx)
}

func (s *RectangleSelector) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if s == nil || !s.Active || ctx == nil {
		return false, PickInfo{}
	}
	b := s.Bounds(ctx)
	if b.W() <= 0 || b.H() <= 0 {
		return false, PickInfo{}
	}
	if p.X < b.Min.X || p.X > b.Max.X || p.Y < b.Min.Y || p.Y > b.Max.Y {
		return false, PickInfo{}
	}
	return true, PickInfo{}
}

func (s *RectangleSelector) displayRectFromData(ctx *DrawContext) geom.Rect {
	if s == nil || ctx == nil || !s.Active {
		return geom.Rect{}
	}
	p1 := ctx.DataToPixel.Apply(s.Min)
	p2 := ctx.DataToPixel.Apply(s.Max)
	return geom.Rect{Min: geom.Pt{X: math.Min(p1.X, p2.X), Y: math.Min(p1.Y, p2.Y)}, Max: geom.Pt{X: math.Max(p1.X, p2.X), Y: math.Max(p1.Y, p2.Y)}}
}

func (s *RectangleSelector) Z() float64   { return s.z }
func (s *RectangleSelector) WidgetLayer() {}

func (e *EllipseSelector) OnSelect(cb EllipseSelectorCallback) WidgetCallbackID {
	if e == nil || cb == nil {
		return 0
	}
	return e.onSelect.add(cb)
}

func (e *EllipseSelector) RemoveOnSelect(id WidgetCallbackID) {
	if e == nil {
		return
	}
	e.onSelect.remove(id)
}

func (e *EllipseSelector) SetBounds(minV, maxV geom.Pt) bool {
	if e == nil || !isFinite(minV.X) || !isFinite(minV.Y) || !isFinite(maxV.X) || !isFinite(maxV.Y) {
		return false
	}
	minV, maxV = normalizedRect(minV, maxV)
	changed := !e.Active || e.Min != minV || e.Max != maxV
	e.Active = true
	e.Min = minV
	e.Max = maxV
	return changed
}

func (e *EllipseSelector) MoveBy(delta geom.Pt) bool {
	if e == nil || !e.Active || !isFinite(delta.X) || !isFinite(delta.Y) {
		return false
	}
	e.Min.X += delta.X
	e.Min.Y += delta.Y
	e.Max.X += delta.X
	e.Max.Y += delta.Y
	return true
}

func (e *EllipseSelector) Clear() bool {
	if e == nil || !e.Active {
		return false
	}
	e.Active = false
	e.Min = geom.Pt{}
	e.Max = geom.Pt{}
	return true
}

func (e *EllipseSelector) TriggerOnSelect() {
	if e == nil || !e.Active {
		return
	}
	e.onSelect.each(func(cb EllipseSelectorCallback) {
		cb(e, geom.Rect{Min: e.Min, Max: e.Max})
	})
}

func (e *EllipseSelector) Draw(r render.Renderer, ctx *DrawContext) {
	if e == nil || r == nil || ctx == nil || !e.Active {
		return
	}
	if e.Min.X == e.Max.X || e.Min.Y == e.Max.Y {
		return
	}
	min := geom.Pt{X: e.Min.X, Y: e.Min.Y}
	max := geom.Pt{X: e.Max.X, Y: e.Max.Y}
	centerData := geom.Pt{X: (min.X + max.X) / 2, Y: (min.Y + max.Y) / 2}
	center := ctx.DataToPixel.Apply(centerData)
	p1 := ctx.DataToPixel.Apply(geom.Pt{X: min.X, Y: centerData.Y})
	p2 := ctx.DataToPixel.Apply(geom.Pt{X: max.X, Y: centerData.Y})
	q1 := ctx.DataToPixel.Apply(geom.Pt{X: centerData.X, Y: min.Y})
	q2 := ctx.DataToPixel.Apply(geom.Pt{X: centerData.X, Y: max.Y})
	width := math.Abs(p2.X - p1.X)
	height := math.Abs(q2.Y - q1.Y)
	if width <= 0 || height <= 0 {
		return
	}
	path := ellipsePath(width, height)
	path = applyAffinePath(path, translateAffine(center))
	r.Path(path, &render.Paint{Fill: e.FillColor, Stroke: e.EdgeColor, LineWidth: e.LineWidth})
}

func (e *EllipseSelector) Bounds(ctx *DrawContext) geom.Rect {
	if e == nil || !e.Active || ctx == nil {
		return geom.Rect{}
	}
	return e.displayRect(ctx)
}

func (e *EllipseSelector) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if e == nil || !e.Active || ctx == nil {
		return false, PickInfo{}
	}
	if e.Min.X == e.Max.X || e.Min.Y == e.Max.Y {
		return false, PickInfo{}
	}
	data, ok := (&ctx.DataToPixel).Invert(p)
	if !ok {
		return false, PickInfo{}
	}
	cx := (e.Min.X + e.Max.X) / 2
	cy := (e.Min.Y + e.Max.Y) / 2
	rx := (e.Max.X - e.Min.X) / 2
	ry := (e.Max.Y - e.Min.Y) / 2
	nx := (data.X - cx) / rx
	ny := (data.Y - cy) / ry
	return nx*nx+ny*ny <= 1, PickInfo{}
}

func (e *EllipseSelector) displayRect(ctx *DrawContext) geom.Rect {
	p1 := ctx.DataToPixel.Apply(e.Min)
	p2 := ctx.DataToPixel.Apply(e.Max)
	return geom.Rect{Min: geom.Pt{X: math.Min(p1.X, p2.X), Y: math.Min(p1.Y, p2.Y)}, Max: geom.Pt{X: math.Max(p1.X, p2.X), Y: math.Max(p1.Y, p2.Y)}}
}

func (e *EllipseSelector) Z() float64   { return e.z }
func (e *EllipseSelector) WidgetLayer() {}

func (p *PolygonSelector) OnSelect(cb PolygonSelectorCallback) WidgetCallbackID {
	if p == nil || cb == nil {
		return 0
	}
	return p.onSelect.add(cb)
}

func (p *PolygonSelector) RemoveOnSelect(id WidgetCallbackID) {
	if p == nil {
		return
	}
	p.onSelect.remove(id)
}

func (p *PolygonSelector) AppendPoint(point geom.Pt) bool {
	if p == nil || !isFinite(point.X) || !isFinite(point.Y) {
		return false
	}
	p.Points = append(p.Points, point)
	p.Active = true
	return true
}

func (p *PolygonSelector) SetPoint(index int, point geom.Pt) bool {
	if p == nil || !isFinite(point.X) || !isFinite(point.Y) || index < 0 || index >= len(p.Points) {
		return false
	}
	p.Points[index] = point
	p.Active = true
	return true
}

func (p *PolygonSelector) MovePointBy(index int, delta geom.Pt) bool {
	if p == nil || !isFinite(delta.X) || !isFinite(delta.Y) || index < 0 || index >= len(p.Points) {
		return false
	}
	p.Points[index].X += delta.X
	p.Points[index].Y += delta.Y
	p.Active = true
	return true
}

func (p *PolygonSelector) MoveBy(delta geom.Pt) bool {
	if p == nil || !p.Active || len(p.Points) == 0 || !isFinite(delta.X) || !isFinite(delta.Y) {
		return false
	}
	for i := range p.Points {
		p.Points[i].X += delta.X
		p.Points[i].Y += delta.Y
	}
	return true
}

func (p *PolygonSelector) Close() bool {
	if p == nil || len(p.Points) < 3 {
		return false
	}
	p.Closed = true
	p.Active = true
	return true
}

func (p *PolygonSelector) Clear() bool {
	if p == nil || !p.Active {
		return false
	}
	p.Active = false
	p.Closed = false
	p.Points = nil
	return true
}

func (p *PolygonSelector) TriggerOnSelect() {
	if p == nil || !p.Active || len(p.Points) < 3 {
		return
	}
	points := append([]geom.Pt(nil), p.Points...)
	p.onSelect.each(func(cb PolygonSelectorCallback) {
		cb(p, points)
	})
}

func (p *PolygonSelector) Draw(r render.Renderer, ctx *DrawContext) {
	if p == nil || r == nil || ctx == nil || len(p.Points) == 0 {
		return
	}
	path := polygonSelectorPathFromData(p.Points, ctx)
	if len(path.V) == 0 {
		return
	}
	paint := &render.Paint{Stroke: p.EdgeColor, LineWidth: p.LineWidth}
	if p.Closed {
		path.Close()
		paint.Fill = p.FillColor
	}
	r.Path(path, paint)
}

func (p *PolygonSelector) Bounds(ctx *DrawContext) geom.Rect {
	if p == nil || ctx == nil || len(p.Points) == 0 {
		return geom.Rect{}
	}
	path := polygonSelectorPathFromData(p.Points, ctx)
	bounds, ok := pathBounds(path)
	if !ok {
		return geom.Rect{}
	}
	return bounds
}

func (p *PolygonSelector) Contains(point geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if p == nil || ctx == nil || len(p.Points) == 0 {
		return false, PickInfo{}
	}
	pix := polygonSelectorPathFromData(p.Points, ctx)
	if len(pix.V) < 2 {
		return false, PickInfo{}
	}
	for i, vertex := range pix.V {
		if math.Hypot(point.X-vertex.X, point.Y-vertex.Y) <= DefaultPickRadius {
			return true, PickInfo{Index: i}
		}
	}
	for i := 1; i < len(pix.V); i++ {
		if distancePointToSegment(pix.V[i-1], pix.V[i], point) <= DefaultPickRadius {
			return true, PickInfo{Index: i - 1}
		}
	}
	if p.Closed && len(pix.V) >= 3 && pointInPolygon(point, pix.V) {
		return true, PickInfo{Index: 0}
	}
	return false, PickInfo{}
}

func (p *PolygonSelector) Z() float64   { return p.z }
func (p *PolygonSelector) WidgetLayer() {}

func (l *LassoSelector) OnSelect(cb LassoSelectorCallback) WidgetCallbackID {
	if l == nil || cb == nil {
		return 0
	}
	return l.onSelect.add(cb)
}

func (l *LassoSelector) RemoveOnSelect(id WidgetCallbackID) {
	if l == nil {
		return
	}
	l.onSelect.remove(id)
}

func (l *LassoSelector) Begin(point geom.Pt) bool {
	if l == nil || !isFinite(point.X) || !isFinite(point.Y) {
		return false
	}
	l.Active = true
	l.Tracking = true
	l.Points = []geom.Pt{point}
	return true
}

func (l *LassoSelector) AddPoint(point geom.Pt) bool {
	if l == nil || !l.Tracking || !isFinite(point.X) || !isFinite(point.Y) {
		return false
	}
	l.Points = append(l.Points, point)
	return true
}

func (l *LassoSelector) Finish() bool {
	if l == nil || !l.Tracking {
		return false
	}
	l.Tracking = false
	if len(l.Points) == 0 {
		l.Active = false
		return false
	}
	points := append([]geom.Pt(nil), l.Points...)
	l.Active = true
	l.onSelect.each(func(cb LassoSelectorCallback) {
		cb(l, points)
	})
	return true
}

func (l *LassoSelector) Clear() bool {
	if l == nil || (!l.Active && !l.Tracking) {
		return false
	}
	l.Active = false
	l.Tracking = false
	l.Points = nil
	return true
}

func (l *LassoSelector) Draw(r render.Renderer, ctx *DrawContext) {
	if l == nil || r == nil || ctx == nil || len(l.Points) < 2 {
		return
	}
	path := polygonSelectorPathFromData(l.Points, ctx)
	if len(path.V) == 0 {
		return
	}
	r.Path(path, &render.Paint{Stroke: l.LineColor, LineWidth: l.LineWidth})
}

func (l *LassoSelector) Bounds(ctx *DrawContext) geom.Rect {
	if l == nil || ctx == nil || len(l.Points) < 2 {
		return geom.Rect{}
	}
	path := polygonSelectorPathFromData(l.Points, ctx)
	bounds, ok := pathBounds(path)
	if !ok {
		return geom.Rect{}
	}
	return bounds
}

func (l *LassoSelector) Contains(point geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if l == nil || ctx == nil || len(l.Points) < 2 {
		return false, PickInfo{}
	}
	path := polygonSelectorPathFromData(l.Points, ctx)
	for i := 1; i < len(path.V); i++ {
		if distancePointToSegment(path.V[i-1], path.V[i], point) <= DefaultPickRadius {
			return true, PickInfo{Index: i - 1}
		}
	}
	return false, PickInfo{}
}

func (l *LassoSelector) Z() float64   { return l.z }
func (l *LassoSelector) WidgetLayer() {}

func (c *Cursor) SetData(x, y float64) bool {
	if c == nil || !isFinite(x) || !isFinite(y) {
		return false
	}
	changed := c.X != x || c.Y != y || !c.show
	c.X = x
	c.Y = y
	c.show = true
	return changed
}

func (c *Cursor) Hide() bool {
	if c == nil || !c.show {
		return false
	}
	c.show = false
	return true
}

func (c *Cursor) Position() (float64, float64, bool) {
	if c == nil || !c.show {
		return 0, 0, false
	}
	return c.X, c.Y, true
}

func (c *Cursor) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil || r == nil || ctx == nil || c.axis == nil || !c.show || ctx.Axes != c.axis {
		return
	}
	pt := ctx.DataToPixel.Apply(geom.Pt{X: c.X, Y: c.Y})
	paint := render.Paint{Stroke: c.Color, LineWidth: c.LineWidth}
	if c.Vertical {
		r.Path(pixelLinePath(geom.Pt{X: pt.X, Y: ctx.Clip.Min.Y}, geom.Pt{X: pt.X, Y: ctx.Clip.Max.Y}), &paint)
	}
	if c.Horizontal {
		r.Path(pixelLinePath(geom.Pt{X: ctx.Clip.Min.X, Y: pt.Y}, geom.Pt{X: ctx.Clip.Max.X, Y: pt.Y}), &paint)
	}
}

func (c *Cursor) Bounds(ctx *DrawContext) geom.Rect {
	if c == nil || !c.show {
		return geom.Rect{}
	}
	return ctx.Clip
}

func (c *Cursor) Contains(geom.Pt, *DrawContext) (bool, PickInfo) {
	return false, PickInfo{}
}

func (c *Cursor) Z() float64   { return c.z }
func (c *Cursor) WidgetLayer() {}

func (c *MultiCursor) SetFigurePoint(point geom.Pt) bool {
	if c == nil || !isFinite(point.X) || !isFinite(point.Y) {
		return false
	}
	changed := c.FigureX != point.X || c.FigureY != point.Y || !c.show
	c.FigureX = point.X
	c.FigureY = point.Y
	c.show = true
	return changed
}

func (c *MultiCursor) Hide() bool {
	if c == nil || !c.show {
		return false
	}
	c.show = false
	return true
}

func (c *MultiCursor) Position() (float64, float64, bool) {
	if c == nil || !c.show {
		return 0, 0, false
	}
	return c.FigureX, c.FigureY, true
}

func (c *MultiCursor) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil || r == nil || ctx == nil || !c.show || len(c.Axes) == 0 || !axesContains(c.Axes, ctx.Axes) {
		return
	}
	paint := render.Paint{Stroke: c.Color, LineWidth: c.LineWidth}
	if c.Vertical {
		r.Path(pixelLinePath(geom.Pt{X: c.FigureX, Y: ctx.Clip.Min.Y}, geom.Pt{X: c.FigureX, Y: ctx.Clip.Max.Y}), &paint)
	}
	if c.Horizontal {
		r.Path(pixelLinePath(geom.Pt{X: ctx.Clip.Min.X, Y: c.FigureY}, geom.Pt{X: ctx.Clip.Max.X, Y: c.FigureY}), &paint)
	}
}

func (c *MultiCursor) Bounds(ctx *DrawContext) geom.Rect {
	if c == nil || ctx == nil || !c.show || len(c.Axes) == 0 {
		return geom.Rect{}
	}
	fig := ctx.Axes.figure
	first := true
	var out geom.Rect
	for _, ax := range c.Axes {
		if ax == nil || ax.figure != fig {
			continue
		}
		clip := ax.adjustedLayout(fig)
		if first {
			out = clip
			first = false
			continue
		}
		out = unionRect(out, clip)
	}
	if first {
		return geom.Rect{}
	}
	return out
}

func (c *MultiCursor) Contains(geom.Pt, *DrawContext) (bool, PickInfo) {
	return false, PickInfo{}
}

func (c *MultiCursor) Z() float64   { return c.z }
func (c *MultiCursor) WidgetLayer() {}

func mergeSpanSelectorOptions(base, override SpanSelectorOptions) SpanSelectorOptions {
	if override.Orientation != "" {
		base.Orientation = override.Orientation
	}
	if override.Color != (render.Color{}) {
		base.Color = override.Color
	}
	if override.FillColor != (render.Color{}) {
		base.FillColor = override.FillColor
	}
	if override.LineWidth > 0 {
		base.LineWidth = override.LineWidth
	}
	if override.Min != nil {
		base.Min = override.Min
	}
	if override.Max != nil {
		base.Max = override.Max
	}
	return base
}

func mergeRectangleSelectorOptions(base, override RectangleSelectorOptions) RectangleSelectorOptions {
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.FillColor != (render.Color{}) {
		base.FillColor = override.FillColor
	}
	if override.LineWidth > 0 {
		base.LineWidth = override.LineWidth
	}
	return base
}

func mergeEllipseSelectorOptions(base, override EllipseSelectorOptions) EllipseSelectorOptions {
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.FillColor != (render.Color{}) {
		base.FillColor = override.FillColor
	}
	if override.LineWidth > 0 {
		base.LineWidth = override.LineWidth
	}
	return base
}

func mergePolygonSelectorOptions(base, override PolygonSelectorOptions) PolygonSelectorOptions {
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.FillColor != (render.Color{}) {
		base.FillColor = override.FillColor
	}
	if override.LineWidth > 0 {
		base.LineWidth = override.LineWidth
	}
	return base
}

func mergeLassoSelectorOptions(base, override LassoSelectorOptions) LassoSelectorOptions {
	if override.LineColor != (render.Color{}) {
		base.LineColor = override.LineColor
	}
	if override.LineWidth > 0 {
		base.LineWidth = override.LineWidth
	}
	return base
}

func mergeCursorOptions(base, override CursorOptions) CursorOptions {
	if override.Color != (render.Color{}) {
		base.Color = override.Color
	}
	if override.LineWidth > 0 {
		base.LineWidth = override.LineWidth
	}
	if override.HorizOn != nil {
		base.HorizOn = override.HorizOn
	}
	if override.VerticalOn != nil {
		base.VerticalOn = override.VerticalOn
	}
	return base
}

func mergeMultiCursorOptions(base, override MultiCursorOptions) MultiCursorOptions {
	if override.Color != (render.Color{}) {
		base.Color = override.Color
	}
	if override.LineWidth > 0 {
		base.LineWidth = override.LineWidth
	}
	if override.HorizOn != nil {
		base.HorizOn = override.HorizOn
	}
	if override.VerticalOn != nil {
		base.VerticalOn = override.VerticalOn
	}
	return base
}

func normalizeSpanOrientation(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	if s == "vertical" || s == "y" || s == "v" {
		return "vertical"
	}
	return "horizontal"
}

func polygonSelectorPathFromData(points []geom.Pt, ctx *DrawContext) geom.Path {
	path := geom.Path{}
	if len(points) == 0 || ctx == nil {
		return path
	}
	for i, pt := range points {
		d := ctx.DataToPixel.Apply(pt)
		if i == 0 {
			path.MoveTo(d)
			continue
		}
		path.LineTo(d)
	}
	return path
}

func normalizedRect(a, b geom.Pt) (geom.Pt, geom.Pt) {
	return geom.Pt{X: math.Min(a.X, b.X), Y: math.Min(a.Y, b.Y)}, geom.Pt{X: math.Max(a.X, b.X), Y: math.Max(a.Y, b.Y)}
}

func axesContains(axes []*Axes, target *Axes) bool {
	for _, ax := range axes {
		if ax == target {
			return true
		}
	}
	return false
}

func dedupeAxes(axes []*Axes) []*Axes {
	seen := map[*Axes]bool{}
	out := make([]*Axes, 0, len(axes))
	for _, ax := range axes {
		if ax == nil {
			continue
		}
		if seen[ax] {
			continue
		}
		seen[ax] = true
		out = append(out, ax)
	}
	return out
}
