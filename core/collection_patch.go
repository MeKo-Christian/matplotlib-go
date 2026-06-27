package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// PatchCollection draws many closed paths with shared or per-item patch style.
type PatchCollection struct {
	Collection
	Paths       []geom.Path
	FaceColors  []render.Color
	FaceColor   render.Color
	EdgeColors  []render.Color
	EdgeColor   render.Color
	EdgeWidths  []float64
	EdgeWidth   float64
	Hatches     []string
	Hatch       string
	HatchColors []render.Color
	HatchColor  render.Color
	HatchWidths []float64
	HatchWidth  float64
	LineJoin    render.LineJoin
	LineCap     render.LineCap
}

// Draw renders the patch collection.
func (c *PatchCollection) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil || r == nil || ctx == nil {
		return
	}
	if c.drawPathCollection(r, ctx) {
		return
	}
	for i, path := range c.Paths {
		if len(path.C) == 0 {
			continue
		}
		path = buildCachedDisplayPath(ctx, c.pathCacheSlot(i), c, c.Coords, path, geom.Identity())
		patch := Patch{
			FaceColor:   c.alphaColor(colorAt(c.FaceColor, c.FaceColors, i)),
			EdgeColor:   c.alphaColor(colorAt(c.EdgeColor, c.EdgeColors, i)),
			EdgeWidth:   widthAt(c.EdgeWidth, c.EdgeWidths, i),
			Hatch:       stringAt(c.Hatch, c.Hatches, i),
			HatchColor:  c.alphaColor(colorAt(c.HatchColor, c.HatchColors, i)),
			HatchWidth:  widthAt(c.HatchWidth, c.HatchWidths, i),
			PathEffects: cloneRenderPathEffects(c.PathEffects),
			LineJoin:    c.LineJoin,
			LineCap:     c.LineCap,
		}
		if patch.LineJoin == 0 {
			patch.LineJoin = render.JoinMiter
		}
		if patch.LineCap == 0 {
			patch.LineCap = render.CapButt
		}
		patch.drawStyledPath(r, path, geom.Path{})
	}
}

// Bounds returns the patch collection's data-space bounds when applicable.
func (c *PatchCollection) Bounds(*DrawContext) geom.Rect {
	if c == nil || !artistUsesDataCoords(c, c.Coords) || len(c.Paths) == 0 {
		return geom.Rect{}
	}
	var bounds geom.Rect
	haveBounds := false
	for _, path := range c.Paths {
		pathBounds, ok := pathBounds(path)
		if !ok {
			continue
		}
		if !haveBounds {
			bounds = pathBounds
			haveBounds = true
			continue
		}
		bounds = unionCollectionRect(bounds, pathBounds)
	}
	if !haveBounds {
		return geom.Rect{}
	}
	return bounds
}

func (c *PatchCollection) legendEntry() (legendEntry, bool) {
	if c == nil || c.label() == "" {
		return legendEntry{}, false
	}
	fill := c.alphaColor(colorAt(c.FaceColor, c.FaceColors, 0))
	if len(c.FaceColors) == 0 && c.FaceColor == (render.Color{}) {
		mapped, ok := c.scalarMappedColorAt(0)
		if !ok {
			mapped = fill
		}
		fill = mapped
	}
	return legendEntryFromPatchStyle(
		c.label(),
		fill,
		c.alphaColor(colorAt(c.EdgeColor, c.EdgeColors, 0)),
		widthAt(c.EdgeWidth, c.EdgeWidths, 0),
		stringAt(c.Hatch, c.Hatches, 0),
		c.alphaColor(colorAt(c.HatchColor, c.HatchColors, 0)),
		widthAt(c.HatchWidth, c.HatchWidths, 0),
	), true
}

// SetArray stores scalar values and refreshes mapped patch-collection face
// colors. When SetEdgeColorFace is active, edge colors follow the mapped faces.
func (c *PatchCollection) SetArray(values []float64) error {
	if c == nil {
		return nil
	}
	if err := c.setArray(values); err != nil {
		return err
	}
	c.refreshScalarMappedColors()
	return nil
}

// SetColormap updates the collection colormap and refreshes scalar-derived
// colors when a scalar array is active.
func (c *PatchCollection) SetColormap(name string) {
	if c == nil {
		return
	}
	c.setColormap(name)
	c.refreshScalarMappedColors()
}

// SetNorm updates the collection normalizer and refreshes scalar-derived
// colors when a scalar array is active.
func (c *PatchCollection) SetNorm(norm ScalarNormalizer) error {
	if c == nil {
		return nil
	}
	if err := c.setNorm(norm); err != nil {
		return err
	}
	c.refreshScalarMappedColors()
	return nil
}

// SetCLim updates color limits and refreshes scalar-derived colors.
func (c *PatchCollection) SetCLim(vmin, vmax float64) error {
	if c == nil {
		return nil
	}
	if err := c.setCLim(vmin, vmax); err != nil {
		return err
	}
	c.refreshScalarMappedColors()
	return nil
}

// SetEdgeColorFace makes collection edges track the resolved face colors.
func (c *PatchCollection) SetEdgeColorFace() {
	if c == nil {
		return
	}
	c.EdgeColorsFace = true
	if len(c.FaceColors) > 0 {
		c.EdgeColors = cloneRenderColors(c.FaceColors)
	}
	c.SetStale(true)
}

func (c *PatchCollection) drawPathCollection(r render.Renderer, ctx *DrawContext) bool {
	drawer, ok := r.(render.PathCollectionDrawer)
	if !ok || c == nil || ctx == nil || len(c.Paths) == 0 {
		return false
	}
	if len(c.PathEffects) > 0 {
		return false
	}
	nativeHatch := false
	if hatcher, ok := r.(render.NativeHatcher); ok {
		nativeHatch = hatcher.SupportsNativeHatch()
	}
	if c.hasHatches() && !nativeHatch {
		return false
	}

	batch := render.PathCollectionBatch{Items: make([]render.PathCollectionItem, 0, len(c.Paths))}
	for i, path := range c.Paths {
		if len(path.C) == 0 {
			continue
		}
		path = buildCachedDisplayPath(ctx, c.pathCacheSlot(i), c, c.Coords, path, geom.Identity())
		fill := c.alphaColor(colorAt(c.FaceColor, c.FaceColors, i))
		edge := c.alphaColor(colorAt(c.EdgeColor, c.EdgeColors, i))
		width := widthAt(c.EdgeWidth, c.EdgeWidths, i)
		lineJoin := c.LineJoin
		if lineJoin == 0 {
			lineJoin = render.JoinMiter
		}
		lineCap := c.LineCap
		if lineCap == 0 {
			lineCap = render.CapButt
		}
		hatch := stringAt(c.Hatch, c.Hatches, i)
		hatchColor := c.alphaColor(colorAt(c.HatchColor, c.HatchColors, i))
		hatchWidth := widthAt(c.HatchWidth, c.HatchWidths, i)
		if fill.A <= 0 && (width <= 0 || edge.A <= 0) && (hatch == "" || hatchColor.A <= 0) {
			continue
		}
		batch.Items = append(batch.Items, render.PathCollectionItem{
			Path:         path,
			Paint:        collectionPaint(fill, edge, width, lineJoin, lineCap, nil),
			Hatch:        hatch,
			HatchColor:   hatchColor,
			HatchWidth:   hatchWidth,
			HatchSpacing: render.DefaultHatchSpacing,
			Antialiased:  c.antialiased(),
		})
	}
	if len(batch.Items) == 0 {
		return false
	}
	return drawer.DrawPathCollection(batch)
}

func (c *PatchCollection) hasHatches() bool {
	if c == nil {
		return false
	}
	if c.Hatch != "" || c.HatchColor.A > 0 || c.HatchWidth > 0 {
		return true
	}
	return len(c.Hatches) > 0 || len(c.HatchColors) > 0 || len(c.HatchWidths) > 0
}

func (c *PatchCollection) refreshScalarMappedColors() {
	if c == nil || len(c.ScalarValues) == 0 {
		return
	}
	colors := c.mappedScalarColors()
	c.FaceColors = colors
	if c.EdgeColorsFace {
		c.EdgeColors = cloneRenderColors(colors)
	}
	c.SetStale(true)
}
