package plot3d

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Surface draws a structured surface as projected, z-sorted quadrilateral faces.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) Surface(x, y []float64, z [][]float64, opt core.PlotOptions) *core.PolyCollection {
	limitsChanged := a.observe3DGrid(x, y, z)
	polygons, faceColors, scalarValues, zorder, mapping := a.projectSurfacePolygons(x, y, z, opt)
	if len(polygons) == 0 {
		return nil
	}

	alpha := 1.0
	if v, ok := opt.Alpha.Get(); ok && v >= 0 && v <= 1 {
		alpha = v
	}
	// EdgeWidth wins over LineWidth, and a negative value falls through to the
	// 1 pt default rather than being clamped.
	edgeWidth := 1.0
	if v, ok := opt.EdgeWidth.Get(); ok && v >= 0 {
		edgeWidth = v
	} else if v, ok := opt.LineWidth.Get(); ok && v >= 0 {
		edgeWidth = v
	}
	antialias := render.AntialiasDefault
	if v, ok := opt.Antialiased.Get(); ok && !v {
		antialias = render.AntialiasOff
	}
	for i := range faceColors {
		faceColors[i] = faceColors[i].WithAlphaMultiplier(alpha)
	}
	edgeColor := opt.EdgeColor.Or(render.Color{A: 0}).WithAlphaMultiplier(alpha)
	edgeColors := surfaceEdgeColors(faceColors, opt)

	collection := &core.PolyCollection{
		Polygons: polygons,
		PatchCollection: core.PatchCollection{
			Collection: core.Collection{
				Coords:       core.Coords(core.CoordData),
				Label:        opt.Label,
				Alpha:        1,
				Antialias:    antialias,
				Colormap:     mapping.Colormap,
				Norm:         mapping.Norm,
				VMin:         mapping.VMin,
				VMax:         mapping.VMax,
				ScalarValues: scalarValues,
			},
			FaceColors: faceColors,
			EdgeColor:  edgeColor,
			EdgeColors: edgeColors,
			EdgeWidth:  edgeWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	collection.SetZ(zorder)
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			polygons, faceColors, scalarValues, zorder, mapping := a.projectSurfacePolygons(x, y, z, opt)
			for i := range faceColors {
				faceColors[i] = faceColors[i].WithAlphaMultiplier(alpha)
			}
			collection.Polygons = polygons
			collection.FaceColors = faceColors
			collection.EdgeColors = surfaceEdgeColors(faceColors, opt)
			collection.Colormap = mapping.Colormap
			collection.Norm = mapping.Norm
			collection.VMin = mapping.VMin
			collection.VMax = mapping.VMax
			collection.ScalarValues = scalarValues
			collection.SetZ(zorder)
		}
	}, limitsChanged)
	return collection
}

type surfaceFace struct {
	polygon []geom.Pt
	value   float64
	color   render.Color
	normal  vec3
	depth   float64
}

//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) projectSurfacePolygons(x, y []float64, z [][]float64, opt core.PlotOptions) ([][]geom.Pt, []render.Color, []float64, float64, core.ScalarMapInfo) {
	if a == nil || len(z) == 0 {
		return nil, nil, nil, 0, core.ScalarMapInfo{}
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return nil, nil, nil, 0, core.ScalarMapInfo{}
	}
	for row := 1; row < rows; row++ {
		if len(z[row]) != cols {
			return nil, nil, nil, 0, core.ScalarMapInfo{}
		}
	}

	faces := make([]surfaceFace, 0, (rows-1)*(cols-1))
	values := make([]float64, 0, (rows-1)*(cols-1))
	collectionDepth := math.Inf(1)
	rowIndices, colIndices := surfaceSampleIndices(rows, cols, opt)
	defaultColor := a.NextColor()
	useMapping := opt.Colormap.OrZero() != ""
	useExplicitFaceColors := len(opt.FaceColors) > 0
	shade := surfaceShadeEnabled(opt, useMapping)
	for rowIdx := 0; rowIdx+1 < len(rowIndices); rowIdx++ {
		row0 := rowIndices[rowIdx]
		row1 := rowIndices[rowIdx+1]
		for colIdx := 0; colIdx+1 < len(colIndices); colIdx++ {
			col0 := colIndices[colIdx]
			col1 := colIndices[colIdx+1]
			corners := [4]vec3{
				{x[col0], y[row0], z[row0][col0]},
				{x[col1], y[row0], z[row0][col1]},
				{x[col1], y[row1], z[row1][col1]},
				{x[col0], y[row1], z[row1][col0]},
			}
			normal := corners[0].sub(corners[1]).cross(corners[1].sub(corners[2]))
			polygon := make([]geom.Pt, 0, 2*(row1-row0)+2*(col1-col0))
			rawPolygon := make([]vec3, 0, 2*(row1-row0)+2*(col1-col0))
			value := 0.0
			depth := 0.0
			valid := true
			count := 0
			surfacePatchPerimeter(row0, row1, col0, col1, func(row, col int) {
				if !valid {
					return
				}
				zVal := z[row][col]
				if !isFinite3D(x[col], y[row], zVal) {
					valid = false
					return
				}
				rawPolygon = append(rawPolygon, vec3{x[col], y[row], zVal})
				pt, zDepth := a.projectPointDepth(x[col], y[row], zVal)
				polygon = append(polygon, pt)
				value += zVal
				depth += zDepth
				if zDepth < collectionDepth {
					collectionDepth = zDepth
				}
				count++
			})
			if !valid {
				continue
			}
			if opt.AxLimClip && !a.polygonWithin3DViewLimits(rawPolygon) {
				continue
			}
			value /= float64(count)
			depth /= float64(count)
			baseColor := defaultColor
			switch {
			case useExplicitFaceColors:
				baseColor = faceColorAtIndex(opt.FaceColors, len(faces))
			case opt.Color.IsSet():
				baseColor = opt.Color.OrZero()
			}
			if shade && !useMapping {
				baseColor = shade3DFaceColor(baseColor, normal)
			}
			faces = append(faces, surfaceFace{
				polygon: polygon,
				value:   value,
				color:   baseColor,
				normal:  normal,
				depth:   depth,
			})
			values = append(values, value)
		}
	}
	if len(faces) == 0 {
		return nil, nil, nil, 0, core.ScalarMapInfo{}
	}

	sort.SliceStable(faces, func(i, j int) bool {
		return faces[i].depth > faces[j].depth
	})

	mapping := core.ScalarMapInfo{}
	if useMapping {
		mapping = resolvePlotScalarMap(values, opt)
	}
	polygons := make([][]geom.Pt, len(faces))
	colors := make([]render.Color, len(faces))
	scalarValues := []float64(nil)
	if useMapping {
		scalarValues = make([]float64, len(faces))
	}
	for i, face := range faces {
		polygons[i] = face.polygon
		if useMapping {
			colors[i] = mapping.Color(face.value, 1)
			scalarValues[i] = face.value
		} else {
			colors[i] = face.color
		}
	}
	return polygons, colors, scalarValues, computed3DCollectionZ(collectionDepth), mapping
}

func surfaceGridSampleIndices(length, count int) []int {
	if length <= 0 {
		return nil
	}
	if count <= 0 {
		count = default3DSurfaceCount
	}
	stride := int(math.Ceil(float64(length) / float64(count)))
	if stride < 1 {
		stride = 1
	}

	indices := make([]int, 0, (length+stride-1)/stride+1)
	if (length-1)%stride == 0 {
		for i := 0; i < length; i += stride {
			indices = append(indices, i)
		}
		return indices
	}

	for i := 0; i < length-1; i += stride {
		indices = append(indices, i)
	}
	return append(indices, length-1)
}

//nolint:gocritic // Sampling reads an immutable PlotOptions snapshot.
func surfaceSampleIndices(rows, cols int, opt core.PlotOptions) ([]int, []int) {
	if opt.RStride.IsSet() || opt.CStride.IsSet() {
		rstride := opt.RStride.Or(10)
		cstride := opt.CStride.Or(10)
		return steppedSampleIndices(rows, rstride), steppedSampleIndices(cols, cstride)
	}
	rcount := opt.RCount.Or(default3DSurfaceCount)
	ccount := opt.CCount.Or(default3DSurfaceCount)
	return surfaceGridSampleIndices(rows, rcount), surfaceGridSampleIndices(cols, ccount)
}

func surfacePatchPerimeter(row0, row1, col0, col1 int, emit func(row, col int)) {
	for col := col0; col < col1; col++ {
		emit(row0, col)
	}
	for row := row0; row < row1; row++ {
		emit(row, col1)
	}
	for col := col1; col > col0; col-- {
		emit(row1, col)
	}
	for row := row1; row > row0; row-- {
		emit(row, col0)
	}
}

// Trisurf projects a triangulated unstructured surface mesh as filled polygons.
//
//nolint:gocritic // Triangulation is a public value type; keep the pre-v1 method signature value-semantic.
func (a *Axes3D) Trisurf(tri core.Triangulation, z []float64, opt core.PlotOptions) *core.PolyCollection {
	if a == nil || len(tri.X) == 0 {
		return nil
	}
	if err := tri.Validate(); err != nil || len(z) != len(tri.X) {
		return nil
	}
	var ok bool
	tri, ok = tri.EnsureTriangles()
	if !ok {
		return nil
	}
	limitsChanged := a.observe3DTriangulation(tri, z)

	// NextColor runs even when the caller supplied a color, so the property
	// cycle advances once per triangulated surface either way.
	color := opt.Color.Or(a.NextColor())
	lineWidth := 1.0
	if v, ok := opt.EdgeWidth.Get(); ok {
		lineWidth = v
	} else if v, ok := opt.LineWidth.Get(); ok {
		lineWidth = v
	}
	alpha := 1.0
	if v, ok := opt.Alpha.Get(); ok && v >= 0 && v <= 1 {
		alpha = v
	}
	edgeColor := render.Color{A: 0}
	if v, ok := opt.EdgeColor.Get(); ok {
		edgeColor = v.WithAlphaMultiplier(alpha)
	}
	antialias := render.AntialiasDefault
	if v, ok := opt.Antialiased.Get(); ok && !v {
		antialias = render.AntialiasOff
	}

	faceColor := color.WithAlphaMultiplier(alpha)
	faces, faceColors, scalarValues, faceZ, mapping := a.projectTriangulationFaces(tri, z, faceColor, opt)
	if len(faces) == 0 {
		return nil
	}
	collection := &core.PolyCollection{
		Polygons: faces,
		PatchCollection: core.PatchCollection{
			Collection: core.Collection{
				Coords:       core.Coords(core.CoordData),
				Label:        opt.Label,
				Alpha:        1,
				Antialias:    antialias,
				Colormap:     mapping.Colormap,
				Norm:         mapping.Norm,
				VMin:         mapping.VMin,
				VMax:         mapping.VMax,
				ScalarValues: scalarValues,
			},
			FaceColors: faceColors,
			EdgeColor:  edgeColor,
			EdgeWidth:  lineWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	collection.SetZ(faceZ)
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			faces, faceColors, scalarValues, faceZ, mapping := a.projectTriangulationFaces(tri, z, faceColor, opt)
			collection.Polygons = faces
			collection.FaceColors = faceColors
			collection.Colormap = mapping.Colormap
			collection.Norm = mapping.Norm
			collection.VMin = mapping.VMin
			collection.VMax = mapping.VMax
			collection.ScalarValues = scalarValues
			collection.SetZ(faceZ)
		}
	}, limitsChanged)
	return collection
}
