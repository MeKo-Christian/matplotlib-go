package core

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// LineCollection draws many line segments or polylines with shared or per-item
// stroke styling.
type LineCollection struct {
	Collection
	Segments     [][]geom.Pt
	Colors       []render.Color
	Color        render.Color
	LineWidths   []float64
	LineWidth    float64
	DashPatterns [][]float64
	Dashes       []float64
	// LineStyle / LineStyles accept Matplotlib linestyle strings ("-", "--",
	// "-.", ":", or the named "solid"/"dashed"/"dashdot"/"dotted") and are
	// converted to width-scaled dash patterns at draw time. Explicit numeric
	// Dashes/DashPatterns take precedence; a per-item LineStyles entry overrides
	// the scalar LineStyle.
	LineStyle  string
	LineStyles []string
	LineJoin   render.LineJoin
	LineCap    render.LineCap
}

// Draw renders the line collection.
func (c *LineCollection) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil || r == nil || ctx == nil {
		return
	}
	for i, segment := range c.Segments {
		if len(segment) < 2 {
			continue
		}
		path := polylinePath(segment)
		path = buildArtistDisplayPath(ctx, c, c.Coords, path, geom.Identity())
		color := c.alphaColor(colorAt(c.Color, c.Colors, i))
		width := widthAt(c.LineWidth, c.LineWidths, i)
		if width <= 0 || color.A <= 0 {
			continue
		}
		dashes := dashesAt(c.Dashes, c.DashPatterns, i)
		if len(dashes) == 0 {
			if style := stringAt(c.LineStyle, c.LineStyles, i); style != "" {
				dashes = lineStyleToDashes(style, width)
			}
		}
		lineJoin := c.LineJoin
		if lineJoin == 0 {
			lineJoin = render.JoinRound
		}
		lineCap := c.LineCap
		if lineCap == 0 {
			lineCap = render.CapButt
		}
		r.Path(path, &render.Paint{
			Stroke:      color,
			LineWidth:   width,
			LineJoin:    lineJoin,
			LineCap:     lineCap,
			Dashes:      dashes,
			PathEffects: cloneRenderPathEffects(c.PathEffects),
			Snap:        render.SnapAuto,
		})
	}
}

// Bounds returns the line collection's data-space bounds when applicable.
func (c *LineCollection) Bounds(*DrawContext) geom.Rect {
	if c == nil || !artistUsesDataCoords(c, c.Coords) || len(c.Segments) == 0 {
		return geom.Rect{}
	}

	var bounds geom.Rect
	haveBounds := false
	for _, segment := range c.Segments {
		if len(segment) == 0 {
			continue
		}
		segmentBounds := geom.Rect{Min: segment[0], Max: segment[0]}
		for _, pt := range segment[1:] {
			segmentBounds = expandRect(segmentBounds, pt)
		}
		if !haveBounds {
			bounds = segmentBounds
			haveBounds = true
			continue
		}
		bounds = unionCollectionRect(bounds, segmentBounds)
	}
	if !haveBounds {
		return geom.Rect{}
	}
	return bounds
}

func (c *LineCollection) legendEntry() (legendEntry, bool) {
	if c == nil || c.label() == "" {
		return legendEntry{}, false
	}
	color := c.alphaColor(colorAt(c.Color, c.Colors, 0))
	if len(c.Colors) == 0 && c.Color == (render.Color{}) {
		mapped, ok := c.scalarMappedColorAt(0)
		if !ok {
			mapped = color
		}
		color = mapped
	}
	return legendEntryFromLine(c.label(), color, widthAt(c.LineWidth, c.LineWidths, 0), dashesAt(c.Dashes, c.DashPatterns, 0)), true
}

// SetArray stores scalar values and refreshes mapped line-collection stroke colors.
func (c *LineCollection) SetArray(values []float64) error {
	if c == nil {
		return nil
	}
	if len(values) == 0 {
		c.ScalarValues = nil
		c.SetStale(true)
		return nil
	}
	if len(c.Segments) > 0 && len(values) != len(c.Segments) {
		return fmt.Errorf("line collection scalar array has %d values, want %d", len(values), len(c.Segments))
	}
	if err := c.setArray(values); err != nil {
		return err
	}
	c.refreshScalarMappedColors()
	return nil
}

// SetColormap updates the line-collection colormap and refreshes scalar-derived strokes.
func (c *LineCollection) SetColormap(name string) {
	if c == nil {
		return
	}
	c.setColormap(name)
	c.refreshScalarMappedColors()
}

// SetNorm updates the line-collection normalizer and refreshes scalar-derived strokes.
func (c *LineCollection) SetNorm(norm ScalarNormalizer) error {
	if c == nil {
		return nil
	}
	if err := c.setNorm(norm); err != nil {
		return err
	}
	c.refreshScalarMappedColors()
	return nil
}

// SetCLim updates line-collection color limits and refreshes scalar-derived strokes.
func (c *LineCollection) SetCLim(vmin, vmax float64) error {
	if c == nil {
		return nil
	}
	if err := c.setCLim(vmin, vmax); err != nil {
		return err
	}
	c.refreshScalarMappedColors()
	return nil
}

func (c *LineCollection) refreshScalarMappedColors() {
	if c == nil || len(c.ScalarValues) == 0 {
		return
	}
	c.Colors = c.mappedScalarColors()
	c.SetStale(true)
}
