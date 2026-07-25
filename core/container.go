package core

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/diag"
	"github.com/cwbudde/matplotlib-go/render"
)

// stemIsHorizontal normalizes Matplotlib's stem orientation kwarg. Empty and
// "vertical" keep the default vertical stems; "horizontal" flips to horizontal
// stems. Any other value warns and falls back to vertical.
func stemIsHorizontal(orientation string) bool {
	switch strings.ToLower(strings.TrimSpace(orientation)) {
	case "", "vertical":
		return false
	case "horizontal":
		return true
	default:
		diag.Warnf("Stem: invalid orientation %q; using \"vertical\"", orientation)
		return false
	}
}

// BarContainer groups the artist returned by a bar call with rectangle-like
// patch views over each bar.
type BarContainer struct {
	Bar      *Bar2D
	Patches  []*Rectangle
	Errorbar *ErrorbarContainer
	Label    string
}

// ErrorbarContainer groups an errorbar artist and any related artists.
type ErrorbarContainer struct {
	Errorbar *ErrorBar
	DataLine *Line2D
	Label    string
}

// StemContainer groups the artists that make up a stem plot.
type StemContainer struct {
	MarkerCollection *PathCollection
	StemLines        *LineCollection
	Baseline         *Line2D
	Label            string
}

// StemOptions configures Axes.Stem.
type StemOptions struct {
	Color           *render.Color
	LineWidth       *float64
	Marker          *MarkerType
	MarkerPath      *geom.Path
	MarkerSize      *float64
	Baseline        *float64
	BaselineColor   *render.Color
	BaselineWidth   *float64
	MarkerEdgeColor *render.Color
	MarkerEdgeWidth *float64
	Label           string
	Alpha           *float64
	// Orientation selects the stem direction, matching Matplotlib's stem
	// orientation kwarg. "" and "vertical" keep stems along y (the default);
	// "horizontal" runs stems along x with a vertical baseline. Any other value
	// warns and falls back to vertical.
	Orientation string
}

// Len reports the number of bars in the container.
func (c *BarContainer) Len() int {
	if c == nil {
		return 0
	}
	return len(c.Patches)
}

// Artists returns the primary artists referenced by the bar container.
func (c *BarContainer) Artists() []Artist {
	if c == nil || c.Bar == nil {
		return nil
	}
	out := []Artist{c.Bar}
	if c.Errorbar != nil {
		out = append(out, c.Errorbar.Artists()...)
	}
	return out
}

// Len reports the number of errorbar points in the container.
func (c *ErrorbarContainer) Len() int {
	if c == nil || c.Errorbar == nil {
		return 0
	}
	return len(c.Errorbar.XY)
}

// Artists returns the primary artists referenced by the errorbar container.
func (c *ErrorbarContainer) Artists() []Artist {
	if c == nil {
		return nil
	}
	out := []Artist{}
	if c.DataLine != nil {
		out = append(out, c.DataLine)
	}
	if c.Errorbar != nil {
		out = append(out, c.Errorbar)
	}
	return out
}

// Len reports the number of stems in the container.
func (c *StemContainer) Len() int {
	if c == nil || c.StemLines == nil {
		return 0
	}
	return len(c.StemLines.Segments)
}

// Artists returns the artists referenced by the stem container.
func (c *StemContainer) Artists() []Artist {
	if c == nil {
		return nil
	}
	out := []Artist{}
	if c.MarkerCollection != nil {
		out = append(out, c.MarkerCollection)
	}
	if c.StemLines != nil {
		out = append(out, c.StemLines)
	}
	if c.Baseline != nil {
		out = append(out, c.Baseline)
	}
	return out
}

// Container returns a Matplotlib-style result container for the bar artist.
func (b *Bar2D) Container() *BarContainer {
	if b == nil {
		return nil
	}
	return &BarContainer{
		Bar:      b,
		Patches:  b.Rectangles(),
		Errorbar: b.errorbar.Container(),
		Label:    b.Label,
	}
}

// Rectangles returns rectangle patch views for each bar.
func (b *Bar2D) Rectangles() []*Rectangle {
	if b == nil || len(b.X) == 0 || len(b.Heights) == 0 {
		return nil
	}
	n := len(b.X)
	if len(b.Heights) < n {
		n = len(b.Heights)
	}
	rectangles := make([]*Rectangle, 0, n)
	for i := 0; i < n; i++ {
		width := b.Width
		if len(b.Widths) > 0 && i < len(b.Widths) {
			width = b.Widths[i]
		}
		fill := b.Color
		if len(b.Colors) > 0 && i < len(b.Colors) {
			fill = b.Colors[i]
		}
		edge := b.EdgeColor
		if len(b.EdgeColors) > 0 && i < len(b.EdgeColors) {
			edge = b.EdgeColors[i]
		}
		alpha := b.Alpha
		if alpha <= 0 {
			alpha = 1
		}
		fill = fill.WithAlphaMultiplier(alpha)
		edge = edge.WithAlphaMultiplier(alpha)

		x := b.X[i]
		height := b.Heights[i]
		baseline := b.baselineAt(i)

		var xy geom.Pt
		var rectWidth, rectHeight float64
		if b.Orientation == BarHorizontal {
			left := baseline
			right := baseline + height
			if height < 0 {
				left = baseline + height
				right = baseline
			}
			half := width / 2
			xy = geom.Pt{X: left, Y: x - half}
			rectWidth = right - left
			rectHeight = width
		} else {
			bottom := baseline
			top := baseline + height
			if height < 0 {
				bottom = baseline + height
				top = baseline
			}
			half := width / 2
			xy = geom.Pt{X: x - half, Y: bottom}
			rectWidth = width
			rectHeight = top - bottom
		}

		rectangles = append(rectangles, &Rectangle{
			Patch: Patch{
				FaceColor: fill,
				EdgeColor: edge,
				EdgeWidth: b.EdgeWidth,
				Label:     b.Label,
				z:         b.z,
			},
			XY:     xy,
			Width:  rectWidth,
			Height: rectHeight,
		})
	}
	return rectangles
}

// Container returns a Matplotlib-style result container for the errorbar artist.
func (e *ErrorBar) Container() *ErrorbarContainer {
	if e == nil {
		return nil
	}
	return &ErrorbarContainer{
		Errorbar: e,
		Label:    e.Label,
	}
}

// BarContainer creates bars and returns the corresponding result container.
func (a *Axes) BarContainer(x, heights []float64, opts ...BarOptions) *BarContainer {
	bar := a.bar(x, heights, opts...)
	if bar == nil {
		return nil
	}
	return bar.Container()
}

// ErrorBarContainer creates error bars and returns the corresponding result
// container. It reports the same rejections as Axes.ErrorBar.
func (a *Axes) ErrorBarContainer(x, y, xErr, yErr []float64, opts ...ErrorBarOptions) (*ErrorbarContainer, error) {
	errBar, err := a.ErrorBar(x, y, xErr, yErr, opts...)
	if err != nil {
		return nil, err
	}
	return errBar.Container(), nil
}

// Stem renders a simple stem plot and returns a Matplotlib-style result container.
func (a *Axes) Stem(x, y []float64, opts ...StemOptions) *StemContainer {
	if a == nil || len(x) == 0 || len(y) == 0 {
		return nil
	}

	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	if n == 0 {
		return nil
	}

	var opt StemOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	alpha := 1.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}
	color = color.WithAlphaMultiplier(alpha)

	lineWidth := 1.5 // points; converted at the collection Paint sink
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}

	markerType := MarkerCircle
	if opt.Marker != nil {
		markerType = *opt.Marker
	}
	markerSize := 6.0
	if opt.MarkerSize != nil {
		markerSize = *opt.MarkerSize
	}
	markerEdgeColor := color
	if opt.MarkerEdgeColor != nil {
		markerEdgeColor = *opt.MarkerEdgeColor
		markerEdgeColor = markerEdgeColor.WithAlphaMultiplier(alpha)
	}
	markerEdgeWidth := 1.0 // points; converted at the collection Paint sink
	if opt.MarkerEdgeWidth != nil {
		markerEdgeWidth = *opt.MarkerEdgeWidth
	}

	baseline := 0.0
	if opt.Baseline != nil {
		baseline = *opt.Baseline
	}
	baselineColor := a.colorCycleAt(3)
	if opt.BaselineColor != nil {
		baselineColor = *opt.BaselineColor
		baselineColor = baselineColor.WithAlphaMultiplier(alpha)
	} else {
		baselineColor = baselineColor.WithAlphaMultiplier(alpha)
	}
	baselineWidth := lineWidth
	if opt.BaselineWidth != nil {
		baselineWidth = *opt.BaselineWidth
	}

	horizontal := stemIsHorizontal(opt.Orientation)

	segments := make([][]geom.Pt, 0, n)
	offsets := make([]geom.Pt, 0, n)
	for i := 0; i < n; i++ {
		var foot, head geom.Pt
		if horizontal {
			// locs run along y; heads extend along x from a vertical baseline.
			foot = geom.Pt{X: baseline, Y: x[i]}
			head = geom.Pt{X: y[i], Y: x[i]}
		} else {
			foot = geom.Pt{X: x[i], Y: baseline}
			head = geom.Pt{X: x[i], Y: y[i]}
		}
		segments = append(segments, []geom.Pt{foot, head})
		offsets = append(offsets, head)
	}

	baselineStart := geom.Pt{X: x[0], Y: baseline}
	baselineEnd := geom.Pt{X: x[n-1], Y: baseline}
	if horizontal {
		baselineStart = geom.Pt{X: baseline, Y: x[0]}
		baselineEnd = geom.Pt{X: baseline, Y: x[n-1]}
	}

	markerPath := geom.Path{}
	if opt.MarkerPath != nil {
		markerPath = *opt.MarkerPath
	}
	if len(markerPath.C) == 0 {
		scatter := Scatter2D{Marker: markerType}
		markerPath = scatter.markerPrototypePath()
	}

	stems := &LineCollection{
		Collection: Collection{Label: opt.Label, Alpha: alpha, z: 2},
		Segments:   segments,
		Color:      color,
		LineWidth:  lineWidth,
	}
	markers := &PathCollection{
		Collection:    Collection{Label: opt.Label, Alpha: alpha, z: 2.5},
		Path:          markerPath,
		Offsets:       offsets,
		Size:          pointsToPixels(a.resolvedRC(), markerSize),
		PathInDisplay: true,
		FaceColor:     color,
		EdgeColor:     markerEdgeColor,
		EdgeWidth:     markerEdgeWidth,
		LineOnly:      markerLineOnly(NewMarkerStyle(markerType)),
	}
	baselineArtist := &Line2D{
		XY:    []geom.Pt{baselineStart, baselineEnd},
		W:     baselineWidth,
		Col:   baselineColor,
		Label: "",
		z:     1.5,
	}

	a.AddCollection(stems)
	a.AddCollection(markers)
	a.Add(baselineArtist)

	return &StemContainer{
		MarkerCollection: markers,
		StemLines:        stems,
		Baseline:         baselineArtist,
		Label:            opt.Label,
	}
}
