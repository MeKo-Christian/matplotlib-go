package core

import (
	"fmt"
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

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

// SetArray stores flattened mesh scalar values and refreshes colors according
// to the mesh shading mode. Flat and nearest meshes expect one value per cell;
// Gouraud meshes expect one value per grid vertex.
func (m *QuadMesh) SetArray(values []float64) error {
	if m == nil {
		return nil
	}
	rows, cols, ok := m.scalarGridShape()
	if !ok {
		return fmt.Errorf("mesh geometry is incomplete")
	}
	if len(values) == 0 {
		m.Collection.ScalarValues = nil
		m.SetStale(true)
		return nil
	}
	if len(values) != rows*cols {
		return fmt.Errorf("mesh scalar array has %d values, want %d for %dx%d %s mesh", len(values), rows*cols, rows, cols, m.resolvedShading())
	}
	if err := m.Collection.setArray(values); err != nil {
		return err
	}
	m.Values = reshapeMeshValues(m.Collection.ScalarValues, rows, cols)
	m.refreshScalarMappedColors()
	return nil
}

// SetColormap updates the mesh colormap and refreshes scalar-derived colors.
func (m *QuadMesh) SetColormap(name string) {
	if m == nil {
		return
	}
	m.Collection.setColormap(name)
	m.refreshScalarMappedColors()
}

// SetNorm updates the mesh normalizer and refreshes scalar-derived colors.
func (m *QuadMesh) SetNorm(norm ScalarNormalizer) error {
	if m == nil {
		return nil
	}
	if err := m.Collection.setNorm(norm); err != nil {
		return err
	}
	m.refreshScalarMappedColors()
	return nil
}

// SetCLim updates mesh color limits and refreshes scalar-derived colors.
func (m *QuadMesh) SetCLim(vmin, vmax float64) error {
	if m == nil {
		return nil
	}
	if err := m.Collection.setCLim(vmin, vmax); err != nil {
		return err
	}
	m.refreshScalarMappedColors()
	return nil
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

func (m *QuadMesh) refreshScalarMappedColors() {
	if m == nil || len(m.ScalarValues) == 0 {
		return
	}
	rows, cols, ok := m.scalarGridShape()
	if !ok || len(m.ScalarValues) != rows*cols {
		return
	}
	if !meshValuesShapeOK(m.Values, rows, cols) {
		m.Values = reshapeMeshValues(m.ScalarValues, rows, cols)
	}

	mapping := m.ScalarMap().Resolved()
	alpha := m.alphaValue()
	switch m.resolvedShading() {
	case MeshShadingGouraud:
		colors := make([]render.Color, 0, maxInt(0, rows-1)*maxInt(0, cols-1))
		for yi := 0; yi+1 < rows; yi++ {
			for xi := 0; xi+1 < cols; xi++ {
				value := meshValueAverage([]float64{
					m.Values[yi][xi],
					m.Values[yi][xi+1],
					m.Values[yi+1][xi],
					m.Values[yi+1][xi+1],
				})
				colors = append(colors, mapping.Color(value, alpha))
			}
		}
		m.FaceColors = colors
	default:
		colors := make([]render.Color, len(m.ScalarValues))
		for i, value := range m.ScalarValues {
			colors[i] = mapping.Color(value, alpha)
		}
		m.FaceColors = colors
	}
	if m.EdgeColorsFace {
		m.EdgeColors = cloneRenderColors(m.FaceColors)
	}
	m.SetStale(true)
}

func (m *QuadMesh) scalarGridShape() (rows, cols int, ok bool) {
	if m == nil {
		return 0, 0, false
	}
	switch m.resolvedShading() {
	case MeshShadingGouraud:
		rows = len(m.YEdges)
		cols = len(m.XEdges)
	default:
		rows = maxInt(0, len(m.YEdges)-1)
		cols = maxInt(0, len(m.XEdges)-1)
	}
	return rows, cols, rows > 0 && cols > 0
}

func (m *QuadMesh) resolvedShading() MeshShading {
	if m == nil {
		return MeshShadingFlat
	}
	if m.Shading == MeshShadingGouraud {
		return MeshShadingGouraud
	}
	return MeshShadingFlat
}

func flattenMeshValues(values [][]float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	count := 0
	for _, row := range values {
		count += len(row)
	}
	out := make([]float64, 0, count)
	for _, row := range values {
		out = append(out, row...)
	}
	return out
}

func reshapeMeshValues(values []float64, rows, cols int) [][]float64 {
	if rows <= 0 || cols <= 0 || len(values) != rows*cols {
		return nil
	}
	out := make([][]float64, rows)
	for yi := 0; yi < rows; yi++ {
		start := yi * cols
		out[yi] = append([]float64(nil), values[start:start+cols]...)
	}
	return out
}

func meshValuesShapeOK(values [][]float64, rows, cols int) bool {
	if len(values) != rows {
		return false
	}
	for _, row := range values {
		if len(row) != cols {
			return false
		}
	}
	return true
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
