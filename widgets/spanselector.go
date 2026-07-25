package widgets

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/optarg"
	"github.com/cwbudde/matplotlib-go/render"
)

// SpanSelectorCallback receives the committed span in data coordinates.
type SpanSelectorCallback func(*SpanSelector, float64, float64)

// SpanSelectorOptions configures a span selector.
type SpanSelectorOptions struct {
	Orientation string
	Color       render.Color
	FillColor   render.Color
	LineWidth   float64
	Min         *float64
	Max         *float64
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

// NewSpanSelector creates a span selector bound to the axes.
func NewSpanSelector(a *core.Axes, orientation string, opts ...SpanSelectorOptions) *SpanSelector {
	if a == nil {
		return nil
	}
	config := SpanSelectorOptions{
		Orientation: "horizontal",
		Color:       render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
		FillColor:   render.Color{R: 0.16, G: 0.42, B: 0.76, A: 0.18},
		LineWidth:   1.2,
	}
	if opt, ok := optarg.Optional("spanselector", opts); ok {
		mergeSpanSelectorOptions(&config, &opt)
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
	a.Add(sel)
	return sel
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

func (s *SpanSelector) Draw(r render.Renderer, ctx *core.DrawContext) {
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
			r.Path(pixelRectPath(rect), &render.Paint{Fill: s.FillColor, Stroke: s.Color, LineWidth: pointsToPixels(&ctx.RC, s.LineWidth)})
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
	r.Path(pixelRectPath(rect), &render.Paint{Fill: s.FillColor, Stroke: s.Color, LineWidth: pointsToPixels(&ctx.RC, s.LineWidth)})
}

func (s *SpanSelector) Bounds(ctx *core.DrawContext) geom.Rect {
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

func (s *SpanSelector) Contains(p geom.Pt, ctx *core.DrawContext) (bool, core.PickInfo) {
	if s == nil || !s.Active || ctx == nil {
		return false, core.PickInfo{}
	}
	value := p
	if s.Orientation == "vertical" {
		data, ok := (&ctx.DataToPixel).Invert(value)
		if !ok {
			return false, core.PickInfo{}
		}
		minV, maxV := s.Start, s.End
		if minV > maxV {
			minV, maxV = maxV, minV
		}
		return data.Y >= minV && data.Y <= maxV, core.PickInfo{}
	}
	data, ok := (&ctx.DataToPixel).Invert(value)
	if !ok {
		return false, core.PickInfo{}
	}
	minV, maxV := s.Start, s.End
	if minV > maxV {
		minV, maxV = maxV, minV
	}
	return data.X >= minV && data.X <= maxV, core.PickInfo{}
}

func (s *SpanSelector) Z() float64   { return s.z }
func (s *SpanSelector) WidgetLayer() {}

func mergeSpanSelectorOptions(base, override *SpanSelectorOptions) {
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
}

func normalizeSpanOrientation(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	if s == "vertical" || s == "y" || s == "v" {
		return "vertical"
	}
	return "horizontal"
}
