package plot3d

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/optarg"
	"github.com/cwbudde/matplotlib-go/render"
)

// Surface draws a structured surface as projected, z-sorted quadrilateral faces.
func (a *Axes3D) Surface(x, y []float64, z [][]float64, opts ...core.PlotOptions) *core.PolyCollection {
	limitsChanged := a.observe3DGrid(x, y, z)
	polygons, faceColors, scalarValues, zorder, mapping := a.projectSurfacePolygons(x, y, z, opts...)
	if len(polygons) == 0 {
		return nil
	}

	alpha := 1.0
	label := ""
	edgeWidth := 1.0
	edgeColor := render.Color{A: 0}
	antialias := render.AntialiasDefault
	if opt, ok := optarg.Optional("surface", opts); ok {
		if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
			alpha = *opt.Alpha
		}
		if opt.EdgeWidth != nil && *opt.EdgeWidth >= 0 {
			edgeWidth = *opt.EdgeWidth
		} else if opt.LineWidth != nil && *opt.LineWidth >= 0 {
			edgeWidth = *opt.LineWidth
		}
		if opt.EdgeColor != nil {
			edgeColor = *opt.EdgeColor
		}
		if opt.Antialiased != nil && !*opt.Antialiased {
			antialias = render.AntialiasOff
		}
		label = opt.Label
	}
	for i := range faceColors {
		faceColors[i] = faceColors[i].WithAlphaMultiplier(alpha)
	}
	edgeColor = edgeColor.WithAlphaMultiplier(alpha)
	edgeColors := surfaceEdgeColors(faceColors, optarg.One("surface", opts))

	collection := &core.PolyCollection{
		Polygons: polygons,
		PatchCollection: core.PatchCollection{
			Collection: core.Collection{
				Coords:       core.Coords(core.CoordData),
				Label:        label,
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
			polygons, faceColors, scalarValues, zorder, mapping := a.projectSurfacePolygons(x, y, z, opts...)
			for i := range faceColors {
				faceColors[i] = faceColors[i].WithAlphaMultiplier(alpha)
			}
			collection.Polygons = polygons
			collection.FaceColors = faceColors
			collection.EdgeColors = surfaceEdgeColors(faceColors, optarg.One("surface", opts))
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

func (a *Axes3D) projectSurfacePolygons(x, y []float64, z [][]float64, opts ...core.PlotOptions) ([][]geom.Pt, []render.Color, []float64, float64, core.ScalarMapInfo) {
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

	opt := optarg.One("surface", opts)
	faces := make([]surfaceFace, 0, (rows-1)*(cols-1))
	values := make([]float64, 0, (rows-1)*(cols-1))
	collectionDepth := math.Inf(1)
	rowIndices, colIndices := surfaceSampleIndices(rows, cols, opt)
	defaultColor := a.NextColor()
	useMapping := opt.Colormap != nil && *opt.Colormap != ""
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
			case opt.Color != nil:
				baseColor = *opt.Color
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
	hasStride := opt.RStride != nil || opt.CStride != nil
	if hasStride {
		rstride, cstride := 10, 10
		if opt.RStride != nil {
			rstride = *opt.RStride
		}
		if opt.CStride != nil {
			cstride = *opt.CStride
		}
		return steppedSampleIndices(rows, rstride), steppedSampleIndices(cols, cstride)
	}
	rcount, ccount := default3DSurfaceCount, default3DSurfaceCount
	if opt.RCount != nil {
		rcount = *opt.RCount
	}
	if opt.CCount != nil {
		ccount = *opt.CCount
	}
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
func (a *Axes3D) Trisurf(tri core.Triangulation, z []float64, opts ...core.PlotOptions) *core.PolyCollection {
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

	color := a.NextColor()
	lineWidth := 1.0
	alpha := 1.0
	label := ""
	edgeColor := render.Color{A: 0}
	antialias := render.AntialiasDefault
	if opt, ok := optarg.Optional("trisurf", opts); ok {
		if opt.Color != nil {
			color = *opt.Color
		}
		if opt.EdgeWidth != nil {
			lineWidth = *opt.EdgeWidth
		} else if opt.LineWidth != nil {
			lineWidth = *opt.LineWidth
		}
		if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
			alpha = *opt.Alpha
		}
		if opt.EdgeColor != nil {
			edgeColor = *opt.EdgeColor
			edgeColor = edgeColor.WithAlphaMultiplier(alpha)
		}
		if opt.Antialiased != nil && !*opt.Antialiased {
			antialias = render.AntialiasOff
		}
		label = opt.Label
	}

	faceColor := color
	faceColor = faceColor.WithAlphaMultiplier(alpha)
	faces, faceColors, scalarValues, faceZ, mapping := a.projectTriangulationFaces(tri, z, faceColor, optarg.One("trisurf", opts))
	if len(faces) == 0 {
		return nil
	}
	collection := &core.PolyCollection{
		Polygons: faces,
		PatchCollection: core.PatchCollection{
			Collection: core.Collection{
				Coords:       core.Coords(core.CoordData),
				Label:        label,
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
			faces, faceColors, scalarValues, faceZ, mapping := a.projectTriangulationFaces(tri, z, faceColor, optarg.One("trisurf", opts))
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
