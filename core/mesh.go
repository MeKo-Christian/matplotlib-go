package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/render"
)

// MeshShading selects how scalar values are assigned to mesh geometry.
type MeshShading string

const (
	MeshShadingAuto    MeshShading = "auto"
	MeshShadingFlat    MeshShading = "flat"
	MeshShadingNearest MeshShading = "nearest"
	MeshShadingGouraud MeshShading = "gouraud"
)

// MeshOptions configures rectilinear mesh plots such as pcolor/pcolormesh.
type MeshOptions struct {
	XEdges    []float64
	YEdges    []float64
	Mask      [][]bool
	Shading   MeshShading
	Colormap  *string
	Norm      ScalarNormalizer
	VMin      *float64
	VMax      *float64
	Alpha     *float64
	EdgeColor *render.Color
	EdgeWidth *float64
	Label     string
}

// Hist2DOptions configures 2D histogram binning and rendering.
type Hist2DOptions struct {
	XBins     int
	YBins     int
	XBinEdges []float64
	YBinEdges []float64
	Weights   []float64
	Norm      HistNorm
	ColorNorm ScalarNormalizer
	Colormap  *string
	VMin      *float64
	VMax      *float64
	Alpha     *float64
	EdgeColor *render.Color
	EdgeWidth *float64
	Label     string
}

// Hist2DResult stores the rendered mesh and the computed counts/edges.
type Hist2DResult struct {
	Mesh   *QuadMesh
	Counts [][]float64
	XEdges []float64
	YEdges []float64
}

// PColor renders a scalar matrix as a rectilinear quad mesh.
func (a *Axes) PColor(data [][]float64, opts ...MeshOptions) *QuadMesh {
	return a.pcolorMesh(data, render.SnapOff, opts...)
}

// PColorFast renders a scalar matrix through the rectilinear quad mesh path.
func (a *Axes) PColorFast(data [][]float64, opts ...MeshOptions) *QuadMesh {
	return a.PColorMesh(data, opts...)
}

// PColorMesh renders a scalar matrix as a rectilinear quad mesh.
func (a *Axes) PColorMesh(data [][]float64, opts ...MeshOptions) *QuadMesh {
	return a.pcolorMesh(data, render.SnapOn, opts...)
}

func (a *Axes) pcolorMesh(data [][]float64, snap render.SnapMode, opts ...MeshOptions) *QuadMesh {
	rows, cols, ok := finiteMatrixSize(data)
	if !ok {
		return nil
	}

	var opt MeshOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	xEdges, yEdges, shading, ok := resolvedMeshGeometry(rows, cols, opt)
	if !ok {
		return nil
	}

	cmap := ""
	if opt.Colormap != nil {
		cmap = *opt.Colormap
	}
	scalarData := meshScalarData(data, opt.Mask)
	mapping, err := ResolveScalarMapGrid(scalarData, ScalarMapConfig{
		Colormap: cmap,
		Norm:     opt.Norm,
		VMin:     opt.VMin,
		VMax:     opt.VMax,
	})
	if err != nil {
		return nil
	}
	alpha := meshAlpha(opt.Alpha)
	edgeWidth := 0.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	edgeColor := render.Color{}
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}

	cellCount := meshFaceColorCount(rows, cols, shading)
	faceColors := make([]render.Color, 0, cellCount)
	edgeColors := make([]render.Color, 0, cellCount)
	if shading == MeshShadingGouraud {
		for yi := 0; yi+1 < rows; yi++ {
			for xi := 0; xi+1 < cols; xi++ {
				value := meshValueAverage([]float64{
					scalarData[yi][xi],
					scalarData[yi][xi+1],
					scalarData[yi+1][xi],
					scalarData[yi+1][xi+1],
				})
				faceColors = append(faceColors, mapping.Color(value, alpha))
				edgeColors = append(edgeColors, edgeColor)
			}
		}
		if opt.EdgeColor != nil || opt.EdgeWidth != nil {
			edgeWidth = 0
		}
	} else {
		for _, row := range scalarData {
			for _, value := range row {
				faceColors = append(faceColors, mapping.Color(value, alpha))
				edgeColors = append(edgeColors, edgeColor)
			}
		}
	}

	values := make([][]float64, len(scalarData))
	for i := range scalarData {
		values[i] = append([]float64(nil), scalarData[i]...)
	}

	mesh := &QuadMesh{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Coords:       Coords(CoordData),
				Label:        opt.Label,
				Alpha:        1,
				Colormap:     mapping.Colormap,
				Norm:         mapping.Norm,
				VMin:         mapping.VMin,
				VMax:         mapping.VMax,
				ScalarValues: flattenMeshValues(values),
			},
			FaceColors: faceColors,
			EdgeColors: edgeColors,
			EdgeWidth:  edgeWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
		XEdges:  append([]float64(nil), xEdges...),
		YEdges:  append([]float64(nil), yEdges...),
		Shading: shading,
		Values:  values,
		Snap:    snap,
	}
	a.Add(mesh)
	return mesh
}

func meshFaceColorCount(rows, cols int, shading MeshShading) int {
	if shading == MeshShadingGouraud {
		return maxInt(0, rows-1) * maxInt(0, cols-1)
	}
	return rows * cols
}

func meshScalarData(data [][]float64, mask [][]bool) [][]float64 {
	out := make([][]float64, len(data))
	for yi, row := range data {
		out[yi] = make([]float64, len(row))
		copy(out[yi], row)
		for xi := range row {
			if meshMasked(mask, yi, xi) {
				out[yi][xi] = math.NaN()
			}
		}
	}
	return out
}

func meshMasked(mask [][]bool, yi, xi int) bool {
	return yi < len(mask) && xi < len(mask[yi]) && mask[yi][xi]
}

func resolvedMeshGeometry(rows, cols int, opt MeshOptions) (xEdges, yEdges []float64, shading MeshShading, ok bool) {
	shading = normalizeMeshShading(opt.Shading)
	x := resolvedMeshCoords(opt.XEdges, cols, shading)
	y := resolvedMeshCoords(opt.YEdges, rows, shading)
	if !allFiniteValues(x) || !allFiniteValues(y) {
		return nil, nil, "", false
	}

	switch shading {
	case MeshShadingAuto:
		switch {
		case len(x) == cols && len(y) == rows:
			xEdges = centersToEdges(x)
			yEdges = centersToEdges(y)
			return xEdges, yEdges, MeshShadingFlat, true
		case len(x) == cols+1 && len(y) == rows+1:
			return x, y, MeshShadingFlat, true
		default:
			return nil, nil, "", false
		}
	case MeshShadingFlat:
		if len(x) != cols+1 || len(y) != rows+1 {
			return nil, nil, "", false
		}
		return x, y, MeshShadingFlat, true
	case MeshShadingNearest:
		if len(x) != cols || len(y) != rows {
			return nil, nil, "", false
		}
		return centersToEdges(x), centersToEdges(y), MeshShadingFlat, true
	case MeshShadingGouraud:
		if len(x) != cols || len(y) != rows {
			return nil, nil, "", false
		}
		return x, y, MeshShadingGouraud, true
	default:
		return nil, nil, "", false
	}
}

func normalizeMeshShading(shading MeshShading) MeshShading {
	switch shading {
	case "", MeshShadingAuto:
		return MeshShadingAuto
	case MeshShadingFlat, MeshShadingNearest, MeshShadingGouraud:
		return shading
	default:
		return MeshShadingAuto
	}
}

func resolvedMeshCoords(coords []float64, cellCount int, shading MeshShading) []float64 {
	if len(coords) > 0 {
		return append([]float64(nil), coords...)
	}
	n := cellCount + 1
	if shading == MeshShadingNearest || shading == MeshShadingGouraud {
		n = cellCount
	}
	out := make([]float64, n)
	for i := range out {
		out[i] = float64(i)
	}
	return out
}

func centersToEdges(centers []float64) []float64 {
	if len(centers) == 0 {
		return nil
	}
	if len(centers) == 1 {
		return []float64{centers[0], centers[0]}
	}
	edges := make([]float64, len(centers)+1)
	edges[0] = centers[0] - (centers[1]-centers[0])*0.5
	for i := 0; i+1 < len(centers); i++ {
		edges[i+1] = (centers[i] + centers[i+1]) * 0.5
	}
	last := len(centers) - 1
	edges[len(edges)-1] = centers[last] + (centers[last]-centers[last-1])*0.5
	return edges
}

func allFiniteValues(values []float64) bool {
	for _, value := range values {
		if !isFinite(value) {
			return false
		}
	}
	return true
}

func meshValueColors(data [][]float64, mapping ScalarMapInfo, alpha float64) [][]render.Color {
	colors := make([][]render.Color, len(data))
	for yi, row := range data {
		colors[yi] = make([]render.Color, len(row))
		for xi, value := range row {
			if !isFinite(value) {
				continue
			}
			colors[yi][xi] = mapping.Color(value, alpha)
		}
	}
	return colors
}

// Hist2D bins paired samples into a 2D count matrix and renders the result as
// a QuadMesh.
func (a *Axes) Hist2D(x, y []float64, opts ...Hist2DOptions) *Hist2DResult {
	if len(x) == 0 || len(y) == 0 {
		return nil
	}

	var opt Hist2DOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	if n == 0 {
		return nil
	}

	xData := make([]float64, 0, n)
	yData := make([]float64, 0, n)
	weights := make([]float64, 0, n)
	if len(opt.Weights) > 0 && len(opt.Weights) < n {
		return nil
	}
	for i := 0; i < n; i++ {
		if !isFinite(x[i]) || !isFinite(y[i]) {
			continue
		}
		weight := 1.0
		if len(opt.Weights) > 0 {
			if !isFinite(opt.Weights[i]) {
				continue
			}
			weight = opt.Weights[i]
		}
		xData = append(xData, x[i])
		yData = append(yData, y[i])
		weights = append(weights, weight)
	}
	if len(xData) == 0 {
		return nil
	}

	xBins := opt.XBins
	if xBins <= 0 {
		xBins = 10
	}
	yBins := opt.YBins
	if yBins <= 0 {
		yBins = 10
	}

	xEdges := resolvedHistogramEdges(xData, xBins, opt.XBinEdges)
	yEdges := resolvedHistogramEdges(yData, yBins, opt.YBinEdges)
	if len(xEdges) < 2 || len(yEdges) < 2 {
		return nil
	}

	counts := make([][]float64, len(yEdges)-1)
	for row := range counts {
		counts[row] = make([]float64, len(xEdges)-1)
	}
	rawTotal := 0.0
	for i := range xData {
		xBin := findBin(xData[i], xEdges)
		yBin := findBin(yData[i], yEdges)
		if xBin < 0 || yBin < 0 {
			continue
		}
		counts[yBin][xBin] += weights[i]
		rawTotal += weights[i]
	}
	if rawTotal > 0 {
		for yi := range counts {
			for xi := range counts[yi] {
				switch opt.Norm {
				case HistNormProbability:
					counts[yi][xi] /= rawTotal
				case HistNormDensity:
					xWidth := xEdges[xi+1] - xEdges[xi]
					yWidth := yEdges[yi+1] - yEdges[yi]
					area := xWidth * yWidth
					if area > 0 {
						counts[yi][xi] /= rawTotal * area
					} else {
						counts[yi][xi] = 0
					}
				}
			}
		}
	}

	meshOpt := MeshOptions{
		XEdges:    xEdges,
		YEdges:    yEdges,
		Colormap:  opt.Colormap,
		Norm:      opt.ColorNorm,
		VMin:      opt.VMin,
		VMax:      opt.VMax,
		Alpha:     opt.Alpha,
		EdgeColor: opt.EdgeColor,
		EdgeWidth: opt.EdgeWidth,
		Label:     opt.Label,
	}
	mesh := a.PColorMesh(counts, meshOpt)
	if mesh == nil {
		return nil
	}
	return &Hist2DResult{
		Mesh:   mesh,
		Counts: counts,
		XEdges: append([]float64(nil), xEdges...),
		YEdges: append([]float64(nil), yEdges...),
	}
}

func resolvedMeshEdges(edges []float64, cellCount int) []float64 {
	if len(edges) > 0 {
		return append([]float64(nil), edges...)
	}
	out := make([]float64, cellCount+1)
	for i := range out {
		out[i] = float64(i)
	}
	return out
}

func resolvedHistogramEdges(data []float64, bins int, explicit []float64) []float64 {
	if len(explicit) > 1 {
		return append([]float64(nil), explicit...)
	}
	return computeBinEdges(data, bins, BinStrategyAuto)
}

func meshAlpha(alpha *float64) float64 {
	if alpha == nil {
		return 1
	}
	return clampOneToOne(*alpha)
}

func meshValueAverage(values []float64) float64 {
	sum := 0.0
	count := 0
	for _, value := range values {
		if !isFinite(value) {
			continue
		}
		sum += value
		count++
	}
	if count == 0 {
		return math.NaN()
	}
	return sum / float64(count)
}
