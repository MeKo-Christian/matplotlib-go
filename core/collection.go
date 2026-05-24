package core

import (
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
	Coords      CoordinateSpec
	Label       string
	Alpha       float64
	Antialias   render.AntialiasMode
	Colormap    string
	Norm        ScalarNormalizer
	VMin        float64
	VMax        float64
	PathEffects []render.PathEffect
	z           float64
}

// PathCollection draws repeated or per-item paths with per-item offsets and
// styling, forming the basis for scatter-like artists.
type PathCollection struct {
	Collection
	Path          geom.Path
	Paths         []geom.Path
	Offsets       []geom.Pt
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
}

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
	LineJoin     render.LineJoin
	LineCap      render.LineCap
}

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

// PolyCollection draws many polygons with patch semantics.
type PolyCollection struct {
	PatchCollection
	Polygons [][]geom.Pt
}

// QuadMesh draws a rectilinear grid of quadrilateral cells, primarily for
// pcolor/pcolormesh-style primitives.
type QuadMesh struct {
	PatchCollection
	XEdges []float64
	YEdges []float64
	// Shading is flat for cell-colored meshes and gouraud for vertex-colored
	// meshes routed to native Gouraud triangle drawing when available.
	Shading MeshShading
	Values  [][]float64
}

// FillBetweenPolyCollection is a polygon-collection primitive specialized for
// fill_between-style regions.
type FillBetweenPolyCollection struct {
	PatchCollection
	X           []float64
	Y1          []float64
	Y2          []float64
	Baseline    float64
	Orientation FillOrientation
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

		paint := collectionPaint(fill, edge, width, c.lineJoin(), c.lineCap(), nil)
		paint.PathEffects = cloneRenderPathEffects(c.PathEffects)
		paint.Hatch = hatch
		paint.HatchColor = hatchColor
		paint.HatchLineWidth = hatchWidth
		paint.HatchSpacing = 32
		if !c.antialiased() {
			paint.Antialias = render.AntialiasOff
		}
		r.Path(path, &paint)
	}
}

// Bounds returns the path collection's data-space bounds when applicable.
func (c *PathCollection) Bounds(*DrawContext) geom.Rect {
	if c == nil || !artistUsesDataCoords(c, c.Coords) {
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

func (c *PathCollection) legendEntry() (legendEntry, bool) {
	if c == nil || c.label() == "" {
		return legendEntry{}, false
	}
	return legendEntry{
		Label:           c.label(),
		kind:            legendEntryMarker,
		markerPath:      c.pathAt(0),
		markerFill:      c.faceColorAt(0),
		markerEdge:      c.edgeColorAt(0),
		markerEdgeWidth: c.edgeWidthAt(0),
	}, true
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
	return legendEntryFromLine(c.label(), c.alphaColor(colorAt(c.Color, c.Colors, 0)), widthAt(c.LineWidth, c.LineWidths, 0), dashesAt(c.Dashes, c.DashPatterns, 0)), true
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
		path = buildArtistDisplayPath(ctx, c, c.Coords, path, geom.Identity())
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
	return legendEntryFromPatchStyle(
		c.label(),
		c.alphaColor(colorAt(c.FaceColor, c.FaceColors, 0)),
		c.alphaColor(colorAt(c.EdgeColor, c.EdgeColors, 0)),
		widthAt(c.EdgeWidth, c.EdgeWidths, 0),
		stringAt(c.Hatch, c.Hatches, 0),
		c.alphaColor(colorAt(c.HatchColor, c.HatchColors, 0)),
		widthAt(c.HatchWidth, c.HatchWidths, 0),
	), true
}

// Draw renders the polygon collection.
func (c *PolyCollection) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil {
		return
	}
	c.asPatchCollection().Draw(r, ctx)
}

// Bounds returns the polygon collection's data-space bounds when applicable.
func (c *PolyCollection) Bounds(ctx *DrawContext) geom.Rect {
	if c == nil {
		return geom.Rect{}
	}
	return c.asPatchCollection().Bounds(ctx)
}

func (c *PolyCollection) legendEntry() (legendEntry, bool) {
	if c == nil {
		return legendEntry{}, false
	}
	return c.asPatchCollection().legendEntry()
}

// Draw renders the quad mesh.
func (m *QuadMesh) Draw(r render.Renderer, ctx *DrawContext) {
	if m == nil {
		return
	}
	if m.drawGouraudMesh(r, ctx) {
		return
	}
	if m.drawQuadMesh(r, ctx) {
		return
	}
	m.asPatchCollection().Draw(r, ctx)
}

// Bounds returns the quad mesh's data-space bounds when applicable.
func (m *QuadMesh) Bounds(ctx *DrawContext) geom.Rect {
	if m == nil || !artistUsesDataCoords(m, m.Coords) || len(m.XEdges) < 2 || len(m.YEdges) < 2 {
		return geom.Rect{}
	}
	return geom.Rect{
		Min: geom.Pt{X: math.Min(m.XEdges[0], m.XEdges[len(m.XEdges)-1]), Y: math.Min(m.YEdges[0], m.YEdges[len(m.YEdges)-1])},
		Max: geom.Pt{X: math.Max(m.XEdges[0], m.XEdges[len(m.XEdges)-1]), Y: math.Max(m.YEdges[0], m.YEdges[len(m.YEdges)-1])},
	}
}

func (m *QuadMesh) legendEntry() (legendEntry, bool) {
	if m == nil {
		return legendEntry{}, false
	}
	return m.asPatchCollection().legendEntry()
}

// Draw renders the fill-between poly collection.
func (c *FillBetweenPolyCollection) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil {
		return
	}
	c.asPatchCollection().Draw(r, ctx)
}

// Bounds returns the fill-between collection's data-space bounds when applicable.
func (c *FillBetweenPolyCollection) Bounds(ctx *DrawContext) geom.Rect {
	if c == nil {
		return geom.Rect{}
	}
	return c.asPatchCollection().Bounds(ctx)
}

func (c *FillBetweenPolyCollection) legendEntry() (legendEntry, bool) {
	if c == nil {
		return legendEntry{}, false
	}
	return c.asPatchCollection().legendEntry()
}

func (c *PolyCollection) asPatchCollection() *PatchCollection {
	if c == nil {
		return nil
	}
	paths := make([]geom.Path, 0, len(c.Polygons)+len(c.Paths))
	for _, poly := range c.Polygons {
		if len(poly) == 0 {
			continue
		}
		paths = append(paths, polygonPath(poly, true))
	}
	paths = append(paths, c.Paths...)
	patches := c.PatchCollection
	patches.Paths = paths
	return &patches
}

func (m *QuadMesh) asPatchCollection() *PatchCollection {
	if m == nil {
		return nil
	}
	paths := make([]geom.Path, 0, maxInt(0, (len(m.XEdges)-1)*(len(m.YEdges)-1)))
	for yi := 0; yi+1 < len(m.YEdges); yi++ {
		for xi := 0; xi+1 < len(m.XEdges); xi++ {
			paths = append(paths, patchRectPath(geom.Rect{
				Min: geom.Pt{X: m.XEdges[xi], Y: m.YEdges[yi]},
				Max: geom.Pt{X: m.XEdges[xi+1], Y: m.YEdges[yi+1]},
			}))
		}
	}
	patches := m.PatchCollection
	patches.Paths = paths
	return &patches
}

func (c *FillBetweenPolyCollection) asPatchCollection() *PatchCollection {
	if c == nil {
		return nil
	}
	paths := []geom.Path{}
	if poly := fillBetweenPolygon(c.X, c.Y1, c.Y2, c.Baseline, c.Orientation); len(poly) > 0 {
		paths = append(paths, polygonPath(poly, true))
	}
	patches := c.PatchCollection
	patches.Paths = paths
	return &patches
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
			Paint:       collectionPaint(fill, edge, width, c.lineJoin(), c.lineCap(), nil),
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
			Paint:        collectionPaint(fill, edge, width, c.lineJoin(), c.lineCap(), nil),
			Hatch:        hatch,
			HatchColor:   hatchColor,
			HatchWidth:   hatchWidth,
			HatchSpacing: 32,
			Antialiased:  c.antialiased(),
		})
	}
	if len(batch.Items) == 0 {
		return false
	}
	return drawer.DrawPathCollection(batch)
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
		path = buildArtistDisplayPath(ctx, c, c.Coords, path, geom.Identity())
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
			HatchSpacing: 32,
			Antialiased:  c.antialiased(),
		})
	}
	if len(batch.Items) == 0 {
		return false
	}
	return drawer.DrawPathCollection(batch)
}

func (m *QuadMesh) drawQuadMesh(r render.Renderer, ctx *DrawContext) bool {
	drawer, ok := r.(render.QuadMeshDrawer)
	if !ok || m == nil || ctx == nil || m.Shading == MeshShadingGouraud || len(m.XEdges) < 2 || len(m.YEdges) < 2 {
		return false
	}
	nativeHatch := false
	if hatcher, ok := r.(render.NativeHatcher); ok {
		nativeHatch = hatcher.SupportsNativeHatch()
	}
	if m.hasHatches() && !nativeHatch {
		return false
	}

	cellCount := (len(m.XEdges) - 1) * (len(m.YEdges) - 1)
	batch := render.QuadMeshBatch{Cells: make([]render.QuadMeshCell, 0, cellCount)}
	idx := 0
	for yi := 0; yi+1 < len(m.YEdges); yi++ {
		for xi := 0; xi+1 < len(m.XEdges); xi++ {
			local := [4]geom.Pt{
				{X: m.XEdges[xi], Y: m.YEdges[yi]},
				{X: m.XEdges[xi+1], Y: m.YEdges[yi]},
				{X: m.XEdges[xi+1], Y: m.YEdges[yi+1]},
				{X: m.XEdges[xi], Y: m.YEdges[yi+1]},
			}
			var quad [4]geom.Pt
			tr := artistTransformFor(ctx, m, m.Coords)
			for i, pt := range local {
				if tr != nil {
					pt = tr.Apply(pt)
				}
				quad[i] = pt
			}
			face := m.alphaColor(colorAt(m.FaceColor, m.FaceColors, idx))
			edge := m.alphaColor(colorAt(m.EdgeColor, m.EdgeColors, idx))
			width := widthAt(m.EdgeWidth, m.EdgeWidths, idx)
			hatch := stringAt(m.Hatch, m.Hatches, idx)
			hatchColor := m.alphaColor(colorAt(m.HatchColor, m.HatchColors, idx))
			hatchWidth := widthAt(m.HatchWidth, m.HatchWidths, idx)
			if face.A > 0 || (width > 0 && edge.A > 0) || (hatch != "" && hatchColor.A > 0) {
				batch.Cells = append(batch.Cells, render.QuadMeshCell{
					Quad:         quad,
					Face:         face,
					Edge:         edge,
					LineWidth:    width,
					Hatch:        hatch,
					HatchColor:   hatchColor,
					HatchWidth:   hatchWidth,
					HatchSpacing: 32,
					Antialiased:  m.antialiased(),
				})
			}
			idx++
		}
	}
	if len(batch.Cells) == 0 {
		return false
	}
	return drawer.DrawQuadMesh(batch)
}

func (m *QuadMesh) drawGouraudMesh(r render.Renderer, ctx *DrawContext) bool {
	drawer, ok := r.(render.GouraudTriangleDrawer)
	if !ok || m == nil || ctx == nil || m.Shading != MeshShadingGouraud || len(m.XEdges) < 2 || len(m.YEdges) < 2 {
		return false
	}
	rows := len(m.YEdges)
	cols := len(m.XEdges)
	if len(m.Values) != rows {
		return false
	}
	for _, row := range m.Values {
		if len(row) != cols {
			return false
		}
	}

	mapping := m.ScalarMap().Resolved()
	alpha := m.alphaValue()
	colors := meshValueColors(m.Values, mapping, alpha)
	tr := artistTransformFor(ctx, m, m.Coords)
	pointAt := func(xi, yi int) geom.Pt {
		pt := geom.Pt{X: m.XEdges[xi], Y: m.YEdges[yi]}
		if tr != nil {
			pt = tr.Apply(pt)
		}
		return pt
	}

	batch := render.GouraudTriangleBatch{
		Triangles:   make([]render.GouraudTriangle, 0, (rows-1)*(cols-1)*4),
		Antialiased: m.antialiased(),
	}
	for yi := 0; yi+1 < rows; yi++ {
		for xi := 0; xi+1 < cols; xi++ {
			p00 := pointAt(xi, yi)
			p10 := pointAt(xi+1, yi)
			p11 := pointAt(xi+1, yi+1)
			p01 := pointAt(xi, yi+1)
			c00 := colors[yi][xi]
			c10 := colors[yi][xi+1]
			c11 := colors[yi+1][xi+1]
			c01 := colors[yi+1][xi]
			if c00.A <= 0 || c10.A <= 0 || c11.A <= 0 || c01.A <= 0 {
				continue
			}
			center := averagePoint4(p00, p10, p11, p01)
			centerColor := averageColor4(c00, c10, c11, c01)
			batch.Triangles = append(batch.Triangles,
				render.GouraudTriangle{P: [3]geom.Pt{p00, p10, center}, Color: [3]render.Color{c00, c10, centerColor}},
				render.GouraudTriangle{P: [3]geom.Pt{p10, p11, center}, Color: [3]render.Color{c10, c11, centerColor}},
				render.GouraudTriangle{P: [3]geom.Pt{p11, p01, center}, Color: [3]render.Color{c11, c01, centerColor}},
				render.GouraudTriangle{P: [3]geom.Pt{p01, p00, center}, Color: [3]render.Color{c01, c00, centerColor}},
			)
		}
	}
	if len(batch.Triangles) == 0 {
		return false
	}
	return drawer.DrawGouraudTriangles(batch)
}

func averagePoint4(a, b, c, d geom.Pt) geom.Pt {
	return geom.Pt{
		X: (a.X + b.X + c.X + d.X) / 4,
		Y: (a.Y + b.Y + c.Y + d.Y) / 4,
	}
}

func averageColor4(a, b, c, d render.Color) render.Color {
	return render.Color{
		R: (a.R + b.R + c.R + d.R) / 4,
		G: (a.G + b.G + c.G + d.G) / 4,
		B: (a.B + b.B + c.B + d.B) / 4,
		A: (a.A + b.A + c.A + d.A) / 4,
	}
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

func (c *PathCollection) hasHatches() bool {
	if c == nil {
		return false
	}
	if c.Hatch != "" || c.HatchColor.A > 0 || c.HatchWidth > 0 {
		return true
	}
	return len(c.Hatches) > 0 || len(c.HatchColors) > 0 || len(c.HatchWidths) > 0
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

func fillBetweenPolygon(x, y1, y2 []float64, baseline float64, orientation FillOrientation) []geom.Pt {
	if len(x) == 0 || len(y1) == 0 {
		return nil
	}
	n := len(x)
	if len(y1) < n {
		n = len(y1)
	}
	if len(y2) > 0 && len(y2) < n {
		n = len(y2)
	}
	if n < 2 {
		return nil
	}

	poly := make([]geom.Pt, 0, 2*n)
	for i := 0; i < n; i++ {
		poly = append(poly, fillBetweenPoint(orientation, x[i], y1[i]))
	}
	for i := n - 1; i >= 0; i-- {
		dep := baseline
		if len(y2) > 0 {
			dep = y2[i]
		}
		poly = append(poly, fillBetweenPoint(orientation, x[i], dep))
	}
	return poly
}

func fillBetweenPoint(orientation FillOrientation, primary, dependent float64) geom.Pt {
	if orientation == FillHorizontal {
		return geom.Pt{X: dependent, Y: primary}
	}
	return geom.Pt{X: primary, Y: dependent}
}

func (c *PathCollection) itemCount() int {
	count := maxInt(len(c.Paths), len(c.Offsets))
	if count == 0 && len(c.Path.C) > 0 {
		count = 1
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
	return c.alphaColor(colorAt(c.EdgeColor, c.EdgeColors, i))
}

func (c *PathCollection) edgeWidthAt(i int) float64 {
	return widthAt(c.EdgeWidth, c.EdgeWidths, i)
}

func (c *PathCollection) displayPathAt(ctx *DrawContext, i int, base geom.Path) geom.Path {
	scale := c.sizeAt(i)
	offset := c.offsetAt(i)
	if c.PathInDisplay {
		path := scaleAndTranslatePath(base, scale, geom.Pt{})
		tr := artistTransformFor(ctx, c, c.Coords)
		if tr != nil {
			offset = tr.Apply(offset)
		}
		return applyAffinePath(path, translateAffine(offset))
	}
	path := scaleAndTranslatePath(base, scale, offset)
	return buildArtistDisplayPath(ctx, c, c.Coords, path, geom.Identity())
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
