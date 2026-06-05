package core

import (
	"fmt"
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Collection stores shared metadata for collection-style artists.
//
// As with Patch, this is an embedded base in Go rather than a directly
// instantiable artist.
type Collection struct {
	ArtistRasterization
	Coords         CoordinateSpec
	Label          string
	Alpha          float64
	Antialias      render.AntialiasMode
	Colormap       string
	Norm           ScalarNormalizer
	VMin           float64
	VMax           float64
	ScalarValues   []float64
	EdgeColorsFace bool
	PathEffects    []render.PathEffect
	z              float64
	scalarCLimSet  bool
}

// AddCollection mirrors Matplotlib's collection-oriented API.
func (a *Axes) AddCollection(art Artist) {
	if a != nil && art != nil {
		a.Add(art)
	}
}

// Z returns the collection z-order for sorting.
func (c *Collection) Z() float64 {
	if c == nil {
		return 0
	}
	return zOrDefault(c.z, defaultPatchZ)
}

func (c *Collection) alphaValue() float64 {
	alpha := 1.0
	if c != nil && c.Alpha > 0 && c.Alpha <= 1 {
		alpha = c.Alpha
	}
	return c.EffectiveAlpha(alpha)
}

func (c *Collection) alphaColor(color render.Color) render.Color {
	alpha := 1.0
	if c != nil && c.Alpha > 0 && c.Alpha <= 1 {
		alpha = c.Alpha
	}
	color = patchAlphaColor(color, alpha)
	if artistAlpha, ok := c.ArtistAlpha(); ok {
		color.A *= artistAlpha
	}
	color.A = clampOneToOne(color.A)
	return color
}

func (c *Collection) antialiased() bool {
	return c == nil || c.Antialias != render.AntialiasOff
}

func (c *Collection) label() string {
	if c == nil {
		return ""
	}
	return c.Label
}

// ScalarMap exposes the collection's scalar mapping when it is used as a
// scalar-mappable artist, such as QuadMesh or tripcolor-style PolyCollections.
func (c Collection) ScalarMap() ScalarMapInfo {
	return ScalarMapInfo{
		Colormap: c.Colormap,
		VMin:     c.VMin,
		VMax:     c.VMax,
		Norm:     c.Norm,
	}
}

// GetArray returns a copy of the scalar array mapped through the collection's
// colormap, matching Matplotlib's scalar-mappable collection concept.
func (c *Collection) GetArray() []float64 {
	if c == nil || len(c.ScalarValues) == 0 {
		return nil
	}
	return append([]float64(nil), c.ScalarValues...)
}

func (c *Collection) setArray(values []float64) error {
	if c == nil {
		return nil
	}
	c.ScalarValues = cloneFloat64s(values)
	if len(values) == 0 {
		c.SetStale(true)
		return nil
	}
	cfg := ScalarMapConfig{Colormap: c.Colormap}
	if c.Norm != nil {
		cfg.Norm = c.Norm
	} else if c.scalarCLimSet {
		vmin, vmax := c.VMin, c.VMax
		cfg.VMin = &vmin
		cfg.VMax = &vmax
	}
	mapping, err := ResolveScalarMapValues(c.ScalarValues, cfg)
	if err != nil {
		return err
	}
	c.applyScalarMap(mapping)
	return nil
}

func (c *Collection) setColormap(name string) {
	if c == nil {
		return
	}
	c.Colormap = resolvedColormapName(name)
	c.SetStale(true)
}

func (c *Collection) setNorm(norm ScalarNormalizer) error {
	if c == nil {
		return nil
	}
	if norm != nil {
		if len(c.ScalarValues) > 0 {
			norm = norm.Autoscale(c.ScalarValues)
		}
		if err := norm.Validate(); err != nil {
			return err
		}
	}
	c.Norm = norm
	if norm != nil {
		c.VMin, c.VMax = norm.Range()
	}
	c.SetStale(true)
	return nil
}

func (c *Collection) setCLim(vmin, vmax float64) error {
	if c == nil {
		return nil
	}
	if !isFinite(vmin) || !isFinite(vmax) {
		return fmt.Errorf("color limits must be finite")
	}
	norm := scalarNormWithRange(c.Norm, vmin, vmax)
	if err := norm.Validate(); err != nil {
		return err
	}
	c.Norm = norm
	c.VMin = vmin
	c.VMax = vmax
	c.scalarCLimSet = true
	c.SetStale(true)
	return nil
}

func (c *Collection) applyScalarMap(mapping ScalarMapInfo) {
	if c == nil {
		return
	}
	mapping = mapping.Resolved()
	c.Colormap = mapping.Colormap
	c.Norm = mapping.Norm
	c.VMin = mapping.VMin
	c.VMax = mapping.VMax
	c.SetStale(true)
}

func (c *Collection) mappedScalarColors() []render.Color {
	if c == nil || len(c.ScalarValues) == 0 {
		return nil
	}
	mapping := c.ScalarMap().Resolved()
	colors := make([]render.Color, len(c.ScalarValues))
	for i, value := range c.ScalarValues {
		colors[i] = mapping.Color(value, 1)
	}
	return colors
}

func (c *Collection) scalarMappedColorAt(i int) (render.Color, bool) {
	if c == nil || i < 0 || i >= len(c.ScalarValues) {
		return render.Color{}, false
	}
	return c.ScalarMap().Resolved().Color(c.ScalarValues[i], c.alphaValue()), true
}

func scalarNormWithRange(norm ScalarNormalizer, vmin, vmax float64) ScalarNormalizer {
	switch n := norm.(type) {
	case Normalize:
		n.VMin, n.VMax = vmin, vmax
		return n
	case LogNorm:
		n.VMin, n.VMax = vmin, vmax
		return n
	case SymLogNorm:
		n.VMin, n.VMax = vmin, vmax
		return n
	case PowerNorm:
		n.VMin, n.VMax = vmin, vmax
		return n
	case TwoSlopeNorm:
		n.VMin, n.VMax = vmin, vmax
		return n
	case CenteredNorm:
		n.VCenter = (vmin + vmax) * 0.5
		n.HalfRange = math.Abs(vmax-vmin) * 0.5
		return n
	case BoundaryNorm:
		return n
	default:
		return Normalize{VMin: vmin, VMax: vmax}
	}
}

func cloneRenderColors(colors []render.Color) []render.Color {
	if len(colors) == 0 {
		return nil
	}
	return append([]render.Color(nil), colors...)
}

func cloneFloat64s(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	return append([]float64(nil), values...)
}

func collectionPaint(fill, edge render.Color, width float64, join render.LineJoin, cap render.LineCap, dashes []float64) render.Paint {
	paint := render.Paint{
		Fill:      fill,
		Stroke:    edge,
		LineWidth: width,
		LineJoin:  join,
		LineCap:   cap,
		Dashes:    append([]float64(nil), dashes...),
		Snap:      render.SnapAuto,
	}
	if width <= 0 || edge.A <= 0 {
		paint.Stroke = render.Color{}
		paint.LineWidth = 0
	}
	if fill.A <= 0 {
		paint.Fill = render.Color{}
	}
	return paint
}

func colorAt(fallback render.Color, colors []render.Color, i int) render.Color {
	if len(colors) > 0 && i < len(colors) {
		return colors[i]
	}
	return fallback
}

func widthAt(fallback float64, widths []float64, i int) float64 {
	if len(widths) > 0 && i < len(widths) {
		return widths[i]
	}
	return fallback
}

func stringAt(fallback string, items []string, i int) string {
	if len(items) > 0 && i < len(items) {
		return items[i]
	}
	return fallback
}

func dashesAt(fallback []float64, items [][]float64, i int) []float64 {
	if len(items) > 0 && i < len(items) {
		return append([]float64(nil), items[i]...)
	}
	return append([]float64(nil), fallback...)
}

func polylinePath(points []geom.Pt) geom.Path {
	if len(points) == 0 {
		return geom.Path{}
	}
	path := geom.Path{}
	for i, pt := range points {
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	return path
}

func scaleAndTranslatePath(path geom.Path, scale float64, offset geom.Pt) geom.Path {
	affine := geom.Affine{A: scale, D: scale, E: offset.X, F: offset.Y}
	return applyAffinePath(path, affine)
}

func unionCollectionRect(a, b geom.Rect) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: math.Min(a.Min.X, b.Min.X), Y: math.Min(a.Min.Y, b.Min.Y)},
		Max: geom.Pt{X: math.Max(a.Max.X, b.Max.X), Y: math.Max(a.Max.Y, b.Max.Y)},
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
