package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// LegendLocation controls where the legend box is anchored inside the axes.
type LegendLocation uint8

const (
	LegendUpperRight LegendLocation = iota
	LegendUpperLeft
	LegendLowerRight
	LegendLowerLeft
	LegendRight
	LegendCenterLeft
	LegendCenterRight
	LegendLowerCenter
	LegendUpperCenter
	LegendCenter
	LegendBest
)

// Legend renders a styled legend box inside an axes.
// If no explicit internal entries are present, labeled artists on the owning axes are collected automatically.
type Legend struct {
	Axes   *Axes
	Figure *Figure

	entries  []legendEntry
	handlers []legendHandlerOverride

	Location        LegendLocation
	Locator         AnchoredBoxLocator
	Padding         float64
	Inset           float64
	RowGap          float64
	SampleWidth     float64
	SampleTextGap   float64
	ColumnSpacing   float64
	NumColumns      int
	MarkerScale     float64
	ScatterPoints   int
	Title           string
	TitleFontSize   float64
	FrameOn         bool
	CornerRadius    float64
	BackgroundColor render.Color
	BorderColor     render.Color
	TextColor       render.Color
	BorderWidth     float64
	FontSize        float64
	z               float64
}

// NewLegend creates a legend bound to the provided axes.
func NewLegend(ax *Axes) *Legend {
	rc := style.CurrentDefaults()
	if ax != nil {
		rc = ax.resolvedRC()
	}
	fontSize := rc.LegendSize()
	fontPx := pointsToPixels(rc, fontSize)
	return &Legend{
		Axes:            ax,
		Location:        LegendBest,
		Locator:         nil,
		Padding:         0.4 * fontPx,
		Inset:           0.5 * fontPx,
		RowGap:          0.5 * fontPx,
		SampleWidth:     2.0 * fontPx,
		SampleTextGap:   0.8 * fontPx,
		ColumnSpacing:   2.0 * fontPx,
		NumColumns:      1,
		MarkerScale:     1,
		ScatterPoints:   1,
		FrameOn:         true,
		CornerRadius:    0.2 * fontPx,
		BackgroundColor: rc.LegendBackground,
		BorderColor:     rc.LegendBorderColor,
		TextColor:       rc.LegendTextColor,
		BorderWidth:     pointsToPixels(rc, 1),
		FontSize:        fontSize,
		z:               1_000,
	}
}

// NewFigureLegend creates a legend bound to the provided figure.
func NewFigureLegend(fig *Figure) *Legend {
	rc := style.CurrentDefaults()
	if fig != nil {
		rc = fig.RC
	}
	fontSize := rc.LegendSize()
	fontPx := pointsToPixels(rc, fontSize)
	return &Legend{
		Figure:          fig,
		Location:        LegendUpperRight,
		Locator:         nil,
		Padding:         0.4 * fontPx,
		Inset:           0.5 * fontPx,
		RowGap:          0.5 * fontPx,
		SampleWidth:     2.0 * fontPx,
		SampleTextGap:   0.8 * fontPx,
		ColumnSpacing:   2.0 * fontPx,
		NumColumns:      1,
		MarkerScale:     1,
		ScatterPoints:   1,
		FrameOn:         true,
		CornerRadius:    0.2 * fontPx,
		BackgroundColor: rc.LegendBackground,
		BorderColor:     rc.LegendBorderColor,
		TextColor:       rc.LegendTextColor,
		BorderWidth:     pointsToPixels(rc, 1),
		FontSize:        fontSize,
		z:               1_000,
	}
}

// AddLegend appends a legend to the axes.
func (a *Axes) AddLegend() *Legend {
	legend := NewLegend(a)
	a.Add(legend)
	return legend
}

// AddLegend appends a figure-level legend that collects labeled artists from all axes.
func (f *Figure) AddLegend() *Legend {
	legend := NewFigureLegend(f)
	f.Add(legend)
	return legend
}

// Draw renders the legend box and entries.
func (l *Legend) Draw(r render.Renderer, ctx *DrawContext) {
	l.draw(r, ctx)
}

// DrawOverlay renders axes legends after the axes clip has been removed,
// matching Matplotlib's unclipped legend artists.
func (l *Legend) DrawOverlay(r render.Renderer, ctx *DrawContext) {
	l.draw(r, ctx)
}

func (l *Legend) draw(r render.Renderer, ctx *DrawContext) {
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	entries := l.entries
	if len(entries) == 0 {
		entries = l.collectEntries()
	}
	if len(entries) == 0 {
		return
	}

	fontSize := l.FontSize
	if fontSize <= 0 {
		fontSize = ctx.RC.LegendSize()
	}
	if fontSize < 8 {
		fontSize = 8
	}

	rowHeights := make([]float64, len(entries))
	labelLayouts := make([]singleLineTextLayout, len(entries))
	for i, entry := range entries {
		layout := measureSingleLineTextLayout(r, entry.Label, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
		labelLayouts[i] = layout
		rowHeights[i] = legendRowHeight(layout, fontSize, ctx)
	}

	layout := l.layoutEntries(labelLayouts, rowHeights, ctx, fontSize)
	titleLayout, titleHeight, titleFontSize := l.titleLayout(r, ctx, fontSize)
	contentWidth := layout.Width
	if titleLayout.Width > contentWidth {
		contentWidth = titleLayout.Width
	}
	contentHeight := layout.Height
	if l.Title != "" {
		contentHeight += titleHeight + l.RowGap
	}
	boxWidth := l.Padding*2 + contentWidth
	boxHeight := l.Padding*2 + contentHeight
	box := l.legendBoxRect(ctx, boxWidth, boxHeight)

	if l.FrameOn {
		boxPath := pixelRectPath(box)
		if l.CornerRadius > 0 {
			boxPath = matplotlibRoundBoxPath(box, l.CornerRadius)
		}
		r.Path(boxPath, &render.Paint{
			Fill:      l.BackgroundColor,
			Stroke:    l.BorderColor,
			LineWidth: l.BorderWidth,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
			Snap:      render.SnapAuto,
		})
	}

	y := box.Max.Y - l.Padding
	if l.Title != "" {
		drawDisplayText(
			textRen,
			l.Title,
			alignedSingleLineOrigin(
				geom.Pt{X: box.Min.X + l.Padding + contentWidth/2, Y: y - titleHeight/2},
				titleLayout,
				TextAlignCenter,
				textLayoutVAlignCenter,
			),
			titleFontSize,
			l.TextColor,
			ctx.RC.FontKey,
			ctx.RC.UseTeX,
		)
		y -= titleHeight + l.RowGap
	}

	x := box.Min.X + l.Padding
	for _, col := range layout.Columns {
		columnY := y
		for i := col.Start; i < col.End; i++ {
			entry := entries[i]
			rowHeight := rowHeights[i]
			centerY := columnY - rowHeight/2
			labelLayout := labelLayouts[i]
			labelOrigin := alignedSingleLineOrigin(
				geom.Pt{X: x + l.SampleWidth + l.SampleTextGap, Y: centerY},
				labelLayout,
				TextAlignLeft,
				textLayoutVAlignCenter,
			)
			fontPx := pointsToPixels(ctx.RC, fontSize)
			sampleCenterY := labelOrigin.Y + 0.35*fontPx

			l.drawSampleWithFontPixels(r, entry, geom.Rect{
				Min: geom.Pt{X: x, Y: sampleCenterY - fontPx/2},
				Max: geom.Pt{X: x + l.SampleWidth, Y: sampleCenterY + fontPx/2},
			}, fontPx)

			drawDisplayText(
				textRen,
				entry.Label,
				labelOrigin,
				fontSize,
				l.TextColor,
				ctx.RC.FontKey,
				ctx.RC.UseTeX,
			)

			columnY -= rowHeight + l.RowGap
		}
		x += col.Width + layout.ColumnSpacing
	}
}

// Z returns the legend z-order.
func (l *Legend) Z() float64 {
	return l.z
}

// Bounds returns an empty rect because legends do not contribute to data bounds.
func (l *Legend) Bounds(*DrawContext) geom.Rect {
	return geom.Rect{}
}

// SetLocator overrides the anchored-box placement strategy for this legend.
func (l *Legend) SetLocator(locator AnchoredBoxLocator) {
	if l == nil {
		return
	}
	l.Locator = locator
}

func (l *Legend) boxRect(r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	if l == nil || r == nil || ctx == nil {
		return geom.Rect{}, false
	}

	entries := l.entries
	if len(entries) == 0 {
		entries = l.collectEntries()
	}
	if len(entries) == 0 {
		return geom.Rect{}, false
	}

	fontSize := l.FontSize
	if fontSize <= 0 {
		fontSize = ctx.RC.LegendSize()
	}
	if fontSize < 8 {
		fontSize = 8
	}

	rowHeights := make([]float64, len(entries))
	labelLayouts := make([]singleLineTextLayout, len(entries))
	for i, entry := range entries {
		layout := measureSingleLineTextLayout(r, entry.Label, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
		labelLayouts[i] = layout
		rowHeights[i] = legendRowHeight(layout, fontSize, ctx)
	}

	layout := l.layoutEntries(labelLayouts, rowHeights, ctx, fontSize)
	titleLayout, titleHeight, _ := l.titleLayout(r, ctx, fontSize)
	contentWidth := layout.Width
	if titleLayout.Width > contentWidth {
		contentWidth = titleLayout.Width
	}
	contentHeight := layout.Height
	if l.Title != "" {
		contentHeight += titleHeight + l.RowGap
	}
	boxWidth := l.Padding*2 + contentWidth
	boxHeight := l.Padding*2 + contentHeight
	return l.legendBoxRect(ctx, boxWidth, boxHeight), true
}

func pixelRectPath(r geom.Rect) geom.Path {
	path := geom.Path{}
	corners := []geom.Pt{
		r.Min,
		{X: r.Max.X, Y: r.Min.Y},
		r.Max,
		{X: r.Min.X, Y: r.Max.Y},
	}
	for i, corner := range corners {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, corner)
	}
	path.C = append(path.C, geom.ClosePath)
	return path
}

func snappedPixelRectPath(r geom.Rect) geom.Path {
	return pixelRectPath(geom.Rect{
		Min: geom.Pt{X: math.Round(r.Min.X), Y: math.Round(r.Min.Y)},
		Max: geom.Pt{X: math.Round(r.Max.X), Y: math.Round(r.Max.Y)},
	})
}
