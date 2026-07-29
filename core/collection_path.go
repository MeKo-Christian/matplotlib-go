package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

// PathCollection draws repeated or per-item paths with per-item offsets and
// styling, forming the basis for scatter-like artists.
type PathCollection struct {
	Collection
	Path    geom.Path
	Paths   []geom.Path
	Offsets []geom.Pt
	// OffsetCoords selects the coordinate system for Offsets. Absent means the
	// collection's own transform applies; the tri-state lives on the field, so a
	// direct write cannot leave a companion flag behind.
	OffsetCoords  optional.Value[CoordinateSpec]
	Sizes         []float64
	Size          float64
	PathInDisplay bool
	FaceColors    []render.Color
	FaceColor     render.Color
	EdgeColors    []render.Color
	EdgeColor     render.Color
	EdgeWidths    []float64
	EdgeWidth     float64
	Hatches       []string
	Hatch         string
	HatchColors   []render.Color
	HatchColor    render.Color
	HatchWidths   []float64
	HatchWidth    float64
	LineJoin      render.LineJoin
	LineJoinSet   bool
	LineCap       render.LineCap
	LineCapSet    bool
	LineOnly      bool
	Snap          render.SnapMode
	SnapSet       bool
}

// Draw renders the path collection.
func (c *PathCollection) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil || r == nil || ctx == nil {
		return
	}
	if c.drawMarkers(r, ctx) {
		return
	}
	if c.drawPathCollection(r, ctx) {
		return
	}
	for i := 0; i < c.itemCount(); i++ {
		base := c.pathAt(i)
		if len(base.C) == 0 {
			continue
		}
		path := c.displayPathAt(ctx, i, base)
		if len(path.C) == 0 {
			continue
		}

		fill := c.faceColorAt(i)
		edge := c.edgeColorAt(i)
		width := c.edgeWidthAt(i)
		hatch := stringAt(c.Hatch, c.Hatches, i)
		hatchColor := c.alphaColor(colorAt(c.HatchColor, c.HatchColors, i))
		hatchWidth := widthAt(c.HatchWidth, c.HatchWidths, i)
		if c.LineOnly {
			if edge.A <= 0 {
				edge = fill
			}
			fill.A = 0
		}
		if fill.A <= 0 && (width <= 0 || edge.A <= 0) && (hatch == "" || hatchColor.A <= 0) {
			continue
		}

		paint := c.collectionPaint(fill, edge, pointsToPixels(ctx.RC, width), nil)
		paint.PathEffects = devicePathEffects(ctx.RC, c.PathEffects)
		paint.Hatch = hatch
		paint.HatchColor = hatchColor
		paint.HatchLineWidth = pointsToPixels(ctx.RC, hatchWidth)
		paint.HatchSpacing = render.DefaultHatchSpacing
		if !c.antialiased() {
			paint.Antialias = render.AntialiasOff
		}
		r.Path(path, &paint)
	}
}

// Bounds returns the path collection's data-space bounds when applicable.
func (c *PathCollection) Bounds(*DrawContext) geom.Rect {
	if c == nil || !c.usesDataOffsets() {
		return geom.Rect{}
	}

	var bounds geom.Rect
	haveBounds := false
	for i := 0; i < c.itemCount(); i++ {
		base := c.pathAt(i)
		if len(base.C) == 0 {
			continue
		}
		path := scaleAndTranslatePath(base, c.sizeAt(i), c.offsetAt(i))
		if c.PathInDisplay {
			path = polygonPath([]geom.Pt{c.offsetAt(i)}, false)
		}
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

// SetOffsetCoords stores the coordinate system used to transform collection
// offsets, separately from the collection path transform.
func (c *PathCollection) SetOffsetCoords(spec CoordinateSpec) {
	if c == nil {
		return
	}
	c.OffsetCoords = optional.Of(spec)
	c.SetStale(true)
}

// ClearOffsetCoords removes the offset coordinate override.
func (c *PathCollection) ClearOffsetCoords() {
	if c == nil {
		return
	}
	c.OffsetCoords = optional.Value[CoordinateSpec]{}
	c.SetStale(true)
}

// SetOffsets replaces per-item offsets, cloning the input slice.
func (c *PathCollection) SetOffsets(offsets []geom.Pt) {
	if c == nil {
		return
	}
	c.Offsets = append([]geom.Pt(nil), offsets...)
	c.SetStale(true)
}

// SetSizes replaces per-item marker sizes, cloning the input slice.
func (c *PathCollection) SetSizes(sizes []float64) {
	if c == nil {
		return
	}
	c.Sizes = cloneFloat64s(sizes)
	c.SetStale(true)
}

// SetFaceColors replaces per-item face colors, cloning the input slice.
func (c *PathCollection) SetFaceColors(colors []render.Color) {
	if c == nil {
		return
	}
	c.FaceColors = cloneRenderColors(colors)
	c.SetStale(true)
}

// SetEdgeColors replaces per-item edge colors, cloning the input slice. Edges
// still follow the face colors while EdgeColorsFace is set; clear that field to
// make these colors take effect.
func (c *PathCollection) SetEdgeColors(colors []render.Color) {
	if c == nil {
		return
	}
	c.EdgeColors = cloneRenderColors(colors)
	c.SetStale(true)
}

func (c *PathCollection) legendEntry() (legendEntry, bool) {
	if c == nil || c.label() == "" {
		return legendEntry{}, false
	}
	fill := c.faceColorAt(0)
	edge := c.edgeColorAt(0)
	if len(c.FaceColors) == 0 && c.FaceColor == (render.Color{}) {
		mapped, ok := c.scalarMappedColorAt(0)
		if !ok {
			mapped = fill
		}
		fill = mapped
		if c.EdgeColorsFace {
			edge = mapped
		}
	}
	// Collection sizes are stored as linear path scales, so the size_min and
	// size_max HandlerRegularPolyCollection works in are their squares; the
	// mean of the two areas is the scale of a lone sample point.
	minScale, maxScale := c.scaleRange()
	entry := legendEntry{
		Label:           c.label(),
		kind:            legendEntryMarker,
		markerPath:      c.pathAt(0),
		markerFill:      fill,
		markerEdge:      edge,
		markerEdgeWidth: c.edgeWidthAt(0),
		markerScaleMin:  minScale,
		markerScaleMax:  maxScale,
	}
	entry.markerSize = math.Sqrt(0.5 * (minScale*minScale + maxScale*maxScale))
	return entry, true
}

// scaleRange returns the smallest and largest per-item path scale.
func (c *PathCollection) scaleRange() (float64, float64) {
	count := c.itemCount()
	if count <= 0 {
		count = 1
	}
	minScale, maxScale := math.Inf(1), 0.0
	for i := 0; i < count; i++ {
		scale := c.sizeAt(i)
		if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
			continue
		}
		minScale = math.Min(minScale, scale)
		maxScale = math.Max(maxScale, scale)
	}
	if maxScale <= 0 || math.IsInf(minScale, 1) {
		return 0, 0
	}
	return minScale, maxScale
}

// SetArray stores scalar values and refreshes mapped path-collection face
// colors. When SetEdgeColorFace is active, edge colors follow the mapped faces.
func (c *PathCollection) SetArray(values []float64) error {
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
func (c *PathCollection) SetColormap(name string) {
	if c == nil {
		return
	}
	c.setColormap(name)
	c.refreshScalarMappedColors()
}

// SetNorm updates the collection normalizer and refreshes scalar-derived
// colors when a scalar array is active.
func (c *PathCollection) SetNorm(norm ScalarNormalizer) error {
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
func (c *PathCollection) SetCLim(vmin, vmax float64) error {
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
func (c *PathCollection) SetEdgeColorFace() {
	if c == nil {
		return
	}
	c.EdgeColorsFace = true
	c.SetStale(true)
}

func (c *PathCollection) drawMarkers(r render.Renderer, ctx *DrawContext) bool {
	drawer, ok := r.(render.MarkerDrawer)
	if !ok || c == nil || ctx == nil || !c.singlePathMarkerOptimization() {
		return false
	}
	if len(c.PathEffects) > 0 {
		return false
	}

	count := c.itemCount()
	if count == 0 {
		return false
	}
	batch := render.MarkerBatch{
		Marker: c.Path,
		Items:  make([]render.MarkerItem, 0, count),
	}
	tr := artistTransformFor(ctx, c, c.Coords)
	for i := 0; i < count; i++ {
		fill := c.faceColorAt(i)
		edge := c.edgeColorAt(i)
		width := c.edgeWidthAt(i)
		if c.LineOnly {
			if edge.A <= 0 {
				edge = fill
			}
			fill.A = 0
		}
		if fill.A <= 0 && (width <= 0 || edge.A <= 0) {
			continue
		}
		offset := c.offsetAt(i)
		if tr != nil {
			offset = tr.Apply(offset)
		}
		scale := c.sizeAt(i)
		batch.Items = append(batch.Items, render.MarkerItem{
			Offset:      offset,
			Transform:   geom.Affine{A: scale, D: scale},
			Paint:       c.collectionPaint(fill, edge, pointsToPixels(ctx.RC, width), nil),
			Snap:        c.Snap,
			SnapSet:     c.SnapSet,
			Antialiased: c.antialiased(),
		})
	}
	if len(batch.Items) == 0 {
		return false
	}
	return drawer.DrawMarkers(batch)
}

func (c *PathCollection) singlePathMarkerOptimization() bool {
	if c == nil || !c.PathInDisplay || len(c.Path.C) == 0 || len(c.Paths) > 0 {
		return false
	}
	if c.hasHatches() {
		return false
	}
	if len(c.Sizes) > 1 || len(c.FaceColors) > 1 || len(c.EdgeColors) > 1 || len(c.EdgeWidths) > 1 {
		return false
	}
	return true
}

func (c *PathCollection) drawPathCollection(r render.Renderer, ctx *DrawContext) bool {
	drawer, ok := r.(render.PathCollectionDrawer)
	if !ok || c == nil || ctx == nil {
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

	count := c.itemCount()
	if count == 0 {
		return false
	}
	batch := render.PathCollectionBatch{Items: make([]render.PathCollectionItem, 0, count)}
	for i := 0; i < count; i++ {
		base := c.pathAt(i)
		if len(base.C) == 0 {
			continue
		}
		path := c.displayPathAt(ctx, i, base)
		if len(path.C) == 0 {
			continue
		}

		fill := c.faceColorAt(i)
		edge := c.edgeColorAt(i)
		width := c.edgeWidthAt(i)
		hatch := stringAt(c.Hatch, c.Hatches, i)
		hatchColor := c.alphaColor(colorAt(c.HatchColor, c.HatchColors, i))
		hatchWidth := widthAt(c.HatchWidth, c.HatchWidths, i)
		if c.LineOnly {
			if edge.A <= 0 {
				edge = fill
			}
			fill.A = 0
		}
		if fill.A <= 0 && (width <= 0 || edge.A <= 0) && (hatch == "" || hatchColor.A <= 0) {
			continue
		}
		batch.Items = append(batch.Items, render.PathCollectionItem{
			Path:         path,
			Paint:        c.collectionPaint(fill, edge, pointsToPixels(ctx.RC, width), nil),
			Hatch:        hatch,
			HatchColor:   hatchColor,
			HatchWidth:   pointsToPixels(ctx.RC, hatchWidth),
			HatchSpacing: render.DefaultHatchSpacing,
			Antialiased:  c.antialiased(),
		})
	}
	if len(batch.Items) == 0 {
		return false
	}
	return drawer.DrawPathCollection(batch)
}

func (c *PathCollection) collectionPaint(fill, edge render.Color, width float64, dashes []float64) render.Paint {
	paint := collectionPaint(fill, edge, width, c.lineJoin(), c.lineCap(), dashes)
	if c != nil && c.SnapSet {
		paint.Snap = c.Snap
	}
	return paint
}

func (c *PathCollection) lineJoin() render.LineJoin {
	if c == nil || !c.LineJoinSet {
		return render.JoinRound
	}
	return c.LineJoin
}

func (c *PathCollection) lineCap() render.LineCap {
	if c == nil || !c.LineCapSet {
		return render.CapRound
	}
	return c.LineCap
}

func (c *PathCollection) hasHatches() bool {
	if c == nil {
		return false
	}
	if c.Hatch != "" || c.HatchColor.A > 0 || c.HatchWidth > 0 {
		return true
	}
	return len(c.Hatches) > 0 || len(c.HatchColors) > 0 || len(c.HatchWidths) > 0
}

func (c *PathCollection) itemCount() int {
	count := maxInt(len(c.Paths), len(c.Offsets))
	if count == 0 && len(c.Path.C) > 0 {
		count = 1
	}
	if count == 0 && len(c.ScalarValues) > 0 {
		count = len(c.ScalarValues)
	}
	if count == 0 && len(c.FaceColors) > 0 {
		count = len(c.FaceColors)
	}
	if count == 0 && len(c.EdgeColors) > 0 {
		count = len(c.EdgeColors)
	}
	if count == 0 && len(c.Sizes) > 0 {
		count = len(c.Sizes)
	}
	return count
}

func (c *PathCollection) pathAt(i int) geom.Path {
	if len(c.Paths) > 0 {
		if i < len(c.Paths) && len(c.Paths[i].C) > 0 {
			return c.Paths[i]
		}
		if len(c.Path.C) == 0 {
			return geom.Path{}
		}
	}
	return c.Path
}

func (c *PathCollection) offsetAt(i int) geom.Pt {
	if len(c.Offsets) == 0 || i >= len(c.Offsets) {
		return geom.Pt{}
	}
	return c.Offsets[i]
}

func (c *PathCollection) usesDataOffsets() bool {
	if c == nil {
		return false
	}
	if spec, ok := c.OffsetCoords.Get(); ok {
		return isDataCoords(spec)
	}
	return artistUsesDataCoords(c, c.Coords)
}

func (c *PathCollection) offsetTransformFor(ctx *DrawContext) transform.T {
	if c == nil || ctx == nil {
		return nil
	}
	if spec, ok := c.OffsetCoords.Get(); ok {
		return ctx.TransformFor(spec)
	}
	return artistTransformFor(ctx, c, c.Coords)
}

func (c *PathCollection) sizeAt(i int) float64 {
	size := c.Size
	if size == 0 {
		size = 1
	}
	if len(c.Sizes) > 0 && i < len(c.Sizes) {
		size = c.Sizes[i]
	}
	if size == 0 {
		return 1
	}
	return size
}

func (c *PathCollection) faceColorAt(i int) render.Color {
	return c.alphaColor(colorAt(c.FaceColor, c.FaceColors, i))
}

func (c *PathCollection) edgeColorAt(i int) render.Color {
	return c.alphaColor(colorAt(c.EdgeColor, c.resolvedEdgeColors(), i))
}

// resolvedEdgeColors returns the per-item stroke colors. Edges follow the face
// colors while EdgeColorsFace is set and face colors exist; with no face colors
// there is nothing to follow and the stored edge colors apply.
func (c *PathCollection) resolvedEdgeColors() []render.Color {
	if c != nil && c.EdgeColorsFace && len(c.FaceColors) > 0 {
		return c.FaceColors
	}
	return c.EdgeColors
}

func (c *PathCollection) edgeWidthAt(i int) float64 {
	return widthAt(c.EdgeWidth, c.EdgeWidths, i)
}

func (c *PathCollection) displayPathAt(ctx *DrawContext, i int, base geom.Path) geom.Path {
	scale := c.sizeAt(i)
	offset := c.offsetAt(i)
	if c.PathInDisplay {
		tr := c.offsetTransformFor(ctx)
		if tr != nil {
			offset = tr.Apply(offset)
		}
		return scaleAndTranslatePath(base, scale, offset)
	}
	path := scaleAndTranslatePath(base, scale, offset)
	return buildCachedDisplayPath(ctx, c.pathCacheSlot(i), c, c.Coords, path, geom.Identity())
}

func (c *PathCollection) refreshScalarMappedColors() {
	if c == nil || len(c.ScalarValues) == 0 {
		return
	}
	c.FaceColors = c.mappedScalarColors()
	c.SetStale(true)
}
