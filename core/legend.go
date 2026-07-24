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
	NumPoints       int
	MarkerScale     float64
	ScatterPoints   int
	HandleHeight    float64
	Title           string
	TitleFontSize   float64
	FrameOn         bool
	Shadow          bool
	CornerRadius    float64
	BackgroundColor render.Color
	BorderColor     render.Color
	TextColor       render.Color
	BorderWidth     float64
	FontSize        float64
	z               float64
	defaultsSet     bool
}

// NewLegend creates a legend bound to the provided axes.
func NewLegend(ax *Axes) *Legend {
	rc := style.CurrentDefaults()
	if ax != nil {
		rc = ax.resolvedRC()
	}
	fontSize := rc.LegendSize()
	fontPx := pointsToPixels(rc, fontSize)
	titleFontSize := rc.Legend.TitleFontSize
	if titleFontSize <= 0 {
		titleFontSize = rc.FontSize
	}
	cornerRadius := 0.0
	if rc.Legend.FancyBox {
		cornerRadius = 0.2 * fontPx
	}
	return &Legend{
		Axes:            ax,
		Location:        legendLocationFromRC(rc.Legend.Location),
		Locator:         nil,
		Padding:         rc.Legend.BorderPad * fontPx,
		Inset:           rc.Legend.BorderAxesPad * fontPx,
		RowGap:          rc.Legend.LabelSpacing * fontPx,
		SampleWidth:     rc.Legend.HandleLength * fontPx,
		SampleTextGap:   rc.Legend.HandleTextPad * fontPx,
		ColumnSpacing:   rc.Legend.ColumnSpacing * fontPx,
		NumColumns:      1,
		NumPoints:       rc.Legend.NumPoints,
		MarkerScale:     rc.Legend.MarkerScale,
		ScatterPoints:   rc.Legend.ScatterPoints,
		HandleHeight:    rc.Legend.HandleHeight,
		TitleFontSize:   titleFontSize,
		FrameOn:         rc.LegendFrameOn,
		Shadow:          rc.Legend.Shadow,
		CornerRadius:    cornerRadius,
		BackgroundColor: rc.LegendBackground,
		BorderColor:     rc.LegendBorderColor,
		TextColor:       rc.LegendTextColor,
		BorderWidth:     pointsToPixels(rc, 1),
		FontSize:        fontSize,
		z:               1_000,
		defaultsSet:     true,
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
	titleFontSize := rc.Legend.TitleFontSize
	if titleFontSize <= 0 {
		titleFontSize = rc.FontSize
	}
	location := legendLocationFromRC(rc.Legend.Location)
	if location == LegendBest {
		location = LegendUpperRight
	}
	cornerRadius := 0.0
	if rc.Legend.FancyBox {
		cornerRadius = 0.2 * fontPx
	}
	return &Legend{
		Figure:          fig,
		Location:        location,
		Locator:         nil,
		Padding:         rc.Legend.BorderPad * fontPx,
		Inset:           rc.Legend.BorderAxesPad * fontPx,
		RowGap:          rc.Legend.LabelSpacing * fontPx,
		SampleWidth:     rc.Legend.HandleLength * fontPx,
		SampleTextGap:   rc.Legend.HandleTextPad * fontPx,
		ColumnSpacing:   rc.Legend.ColumnSpacing * fontPx,
		NumColumns:      1,
		NumPoints:       rc.Legend.NumPoints,
		MarkerScale:     rc.Legend.MarkerScale,
		ScatterPoints:   rc.Legend.ScatterPoints,
		HandleHeight:    rc.Legend.HandleHeight,
		TitleFontSize:   titleFontSize,
		FrameOn:         rc.LegendFrameOn,
		Shadow:          rc.Legend.Shadow,
		CornerRadius:    cornerRadius,
		BackgroundColor: rc.LegendBackground,
		BorderColor:     rc.LegendBorderColor,
		TextColor:       rc.LegendTextColor,
		BorderWidth:     pointsToPixels(rc, 1),
		FontSize:        fontSize,
		z:               1_000,
		defaultsSet:     true,
	}
}

func legendLocationFromRC(location string) LegendLocation {
	switch location {
	case "upper right":
		return LegendUpperRight
	case "upper left":
		return LegendUpperLeft
	case "lower right":
		return LegendLowerRight
	case "lower left":
		return LegendLowerLeft
	case "right":
		return LegendRight
	case "center left":
		return LegendCenterLeft
	case "center right":
		return LegendCenterRight
	case "lower center":
		return LegendLowerCenter
	case "upper center":
		return LegendUpperCenter
	case "center":
		return LegendCenter
	default:
		return LegendBest
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
		handleHeight, _ := l.handleMetrics(pointsToPixels(ctx.RC, fontSize))
		if rowHeights[i] < handleHeight {
			rowHeights[i] = handleHeight
		}
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
		if l.Shadow {
			shadowBox := box
			shadowOffset := pointsToPixels(ctx.RC, 2)
			shadowBox.Min.X += shadowOffset
			shadowBox.Max.X += shadowOffset
			shadowBox.Min.Y -= shadowOffset
			shadowBox.Max.Y -= shadowOffset
			shadowPath := pixelRectPath(shadowBox)
			if l.CornerRadius > 0 {
				shadowPath = matplotlibRoundBoxPath(shadowBox, l.CornerRadius)
			}
			shadowColor := render.Color{
				R: l.BackgroundColor.R * 0.3,
				G: l.BackgroundColor.G * 0.3,
				B: l.BackgroundColor.B * 0.3,
				A: 0.5,
			}
			r.Path(shadowPath, &render.Paint{
				Fill:      shadowColor,
				Stroke:    shadowColor,
				LineWidth: l.BorderWidth,
				LineJoin:  render.JoinMiter,
				LineCap:   render.CapButt,
				Snap:      render.SnapAuto,
			})
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
			handleHeight, handleDescent := l.handleMetrics(fontPx)

			l.drawSampleWithFontPixels(r, &ctx.RC, &entry, geom.Rect{
				Min: geom.Pt{X: x, Y: labelOrigin.Y - handleDescent},
				Max: geom.Pt{X: x + l.SampleWidth, Y: labelOrigin.Y - handleDescent + handleHeight},
			}, fontPx, true)

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
		handleHeight, _ := l.handleMetrics(pointsToPixels(ctx.RC, fontSize))
		if rowHeights[i] < handleHeight {
			rowHeights[i] = handleHeight
		}
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
