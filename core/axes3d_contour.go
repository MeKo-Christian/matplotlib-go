package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Contour projects a structured z grid and emits a placeholder wireframe contour.
func (a *Axes3D) Contour(x, y []float64, z [][]float64, opts ...PlotOptions) *LineCollection {
	limitsChanged := a.observe3DGrid(x, y, z)
	opt := firstPlotOptions(opts)
	segments, segmentLevels, levels, values, zorder := a.projectedContourLineData(x, y, z, opt)
	if len(segments) == 0 {
		return nil
	}

	color := a.NextColor()
	lineWidth := 1.5 // points (matplotlib contour.linewidth/lines.linewidth default); converted at the collection Paint sink
	alpha := 1.0
	label := ""
	colorOverride := false
	if opt.Color != nil {
		color = *opt.Color
		colorOverride = true
	}
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}
	label = opt.Label

	mapping := ScalarMapInfo{}
	colors := []render.Color(nil)
	scalarValues := []float64(nil)
	collectionAlpha := alpha
	if !colorOverride {
		mapping = contourScalarMap(values, levels, opt)
		colors = make([]render.Color, len(segmentLevels))
		for i, level := range segmentLevels {
			colors[i] = mapping.Color(level, alpha)
		}
		scalarValues = append([]float64(nil), levels...)
		collectionAlpha = 1
	}

	collection := &LineCollection{
		Collection: Collection{
			Coords:       Coords(CoordData),
			Label:        label,
			Alpha:        collectionAlpha,
			z:            zorder,
			Colormap:     mapping.Colormap,
			Norm:         mapping.Norm,
			VMin:         mapping.VMin,
			VMax:         mapping.VMax,
			ScalarValues: scalarValues,
		},
		Segments:  segments,
		Color:     color,
		Colors:    colors,
		LineWidth: lineWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			segments, segmentLevels, levels, values, zorder := a.projectedContourLineData(x, y, z, opt)
			collection.Segments = segments
			if !colorOverride {
				mapping := contourScalarMap(values, levels, opt)
				colors := make([]render.Color, len(segmentLevels))
				for i, level := range segmentLevels {
					colors[i] = mapping.Color(level, alpha)
				}
				collection.Colors = colors
				collection.Colormap = mapping.Colormap
				collection.Norm = mapping.Norm
				collection.VMin = mapping.VMin
				collection.VMax = mapping.VMax
				collection.ScalarValues = append([]float64(nil), levels...)
			} else {
				collection.Colormap = ""
				collection.Norm = nil
				collection.VMin = 0
				collection.VMax = 0
				collection.ScalarValues = nil
				collection.Colors = nil
			}
			collection.z = zorder
		}
	}, limitsChanged)
	return collection
}

// TriContour projects contour lines over an explicit triangulated 3D mesh.
func (a *Axes3D) TriContour(tri Triangulation, z []float64, opts ...PlotOptions) *LineCollection {
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

	opt := firstPlotOptions(opts)
	limitsChanged := a.observe3DTriangulation(tri, z)
	segments, segmentLevels, levels, values, zorder := a.projectedTriContourLineData(tri, z, opt)
	if len(segments) == 0 {
		return nil
	}

	color := a.NextColor()
	lineWidth := 1.5 // points (matplotlib contour.linewidth/lines.linewidth default); converted at the collection Paint sink
	alpha := 1.0
	colorOverride := false
	if opt.Color != nil {
		color = *opt.Color
		colorOverride = true
	}
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}

	mapping := ScalarMapInfo{}
	colors := []render.Color(nil)
	scalarValues := []float64(nil)
	collectionAlpha := alpha
	if !colorOverride {
		mapping = contourScalarMap(values, levels, opt)
		colors = make([]render.Color, len(segmentLevels))
		for i, level := range segmentLevels {
			colors[i] = mapping.Color(level, alpha)
		}
		scalarValues = append([]float64(nil), levels...)
		collectionAlpha = 1
	}

	collection := &LineCollection{
		Collection: Collection{
			Coords:       Coords(CoordData),
			Label:        opt.Label,
			Alpha:        collectionAlpha,
			z:            zorder,
			Colormap:     mapping.Colormap,
			Norm:         mapping.Norm,
			VMin:         mapping.VMin,
			VMax:         mapping.VMax,
			ScalarValues: scalarValues,
		},
		Segments:  segments,
		Color:     color,
		Colors:    colors,
		LineWidth: lineWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			segments, segmentLevels, levels, values, zorder := a.projectedTriContourLineData(tri, z, opt)
			collection.Segments = segments
			if !colorOverride {
				mapping := contourScalarMap(values, levels, opt)
				colors := make([]render.Color, len(segmentLevels))
				for i, level := range segmentLevels {
					colors[i] = mapping.Color(level, alpha)
				}
				collection.Colors = colors
				collection.Colormap = mapping.Colormap
				collection.Norm = mapping.Norm
				collection.VMin = mapping.VMin
				collection.VMax = mapping.VMax
				collection.ScalarValues = append([]float64(nil), levels...)
			} else {
				collection.Colormap = ""
				collection.Norm = nil
				collection.VMin = 0
				collection.VMax = 0
				collection.ScalarValues = nil
				collection.Colors = nil
			}
			collection.z = zorder
		}
	}, limitsChanged)
	return collection
}

func (a *Axes3D) projectedContourSegments(x, y []float64, z [][]float64, levelCount int) [][]geom.Pt {
	segments, _, _, _, _ := a.projectedContourLineData(x, y, z, PlotOptions{LevelCount: levelCount})
	return segments
}

func (a *Axes3D) projectedContourLineData(x, y []float64, z [][]float64, opt PlotOptions) ([][]geom.Pt, []float64, []float64, []float64, float64) {
	if a == nil {
		return nil, nil, nil, nil, defaultPatchZ
	}
	zdir := normalized3DDir(opt.ZDir)
	rawLines, rawLevels, levels, values, ok := a.contourLines3D(x, y, z, opt, zdir)
	if !ok || len(rawLines) == 0 {
		return nil, nil, nil, nil, defaultPatchZ
	}
	segments := make([][]geom.Pt, 0, len(rawLines))
	segmentLevels := make([]float64, 0, len(rawLines))
	depth := math.Inf(1)
	for i, polyline := range rawLines {
		if len(polyline) < 2 {
			continue
		}
		level := rawLevels[i]
		planeLevel := contourPlaneLevel(level, opt.Offset)
		runs := [][]vec3{contourPolyline3D(polyline, planeLevel, zdir)}
		if opt.AxLimClip {
			runs = a.clip3DPolylineRuns(runs[0])
		}
		for _, run := range runs {
			if len(run) < 2 {
				continue
			}
			projected := make([]geom.Pt, len(run))
			for j, point3D := range run {
				var zDepth float64
				projected[j], zDepth = a.projectPointDepth(point3D[0], point3D[1], point3D[2])
				if zDepth < depth {
					depth = zDepth
				}
			}
			segments = append(segments, projected)
			segmentLevels = append(segmentLevels, level)
		}
	}
	return segments, segmentLevels, levels, values, computed3DCollectionZ(depth)
}

func (a *Axes3D) projectedTriContourLineData(tri Triangulation, z []float64, opt PlotOptions) ([][]geom.Pt, []float64, []float64, []float64, float64) {
	if a == nil {
		return nil, nil, nil, nil, defaultPatchZ
	}
	zdir := normalized3DDir(opt.ZDir)
	rotatedTri, rotatedValues, ok := rotatedTriangulation3D(tri, z, zdir)
	if !ok {
		return nil, nil, nil, nil, defaultPatchZ
	}
	levels := contourLevels(rotatedValues, opt.Levels, opt.LevelCount, false)
	if len(levels) == 0 {
		return nil, nil, nil, nil, defaultPatchZ
	}
	rawLines, rawLevels := contourPolylines(rotatedTri, rotatedValues, levels)
	if len(rawLines) == 0 {
		return nil, nil, nil, nil, defaultPatchZ
	}

	segments := make([][]geom.Pt, 0, len(rawLines))
	segmentLevels := make([]float64, 0, len(rawLines))
	depth := math.Inf(1)
	for i, polyline := range rawLines {
		if len(polyline) < 2 {
			continue
		}
		level := rawLevels[i]
		planeLevel := contourPlaneLevel(level, opt.Offset)
		runs := [][]vec3{contourPolyline3D(polyline, planeLevel, zdir)}
		if opt.AxLimClip {
			runs = a.clip3DPolylineRuns(runs[0])
		}
		for _, run := range runs {
			if len(run) < 2 {
				continue
			}
			projected := make([]geom.Pt, len(run))
			for j, point3D := range run {
				var zDepth float64
				projected[j], zDepth = a.projectPointDepth(point3D[0], point3D[1], point3D[2])
				if zDepth < depth {
					depth = zDepth
				}
			}
			segments = append(segments, projected)
			segmentLevels = append(segmentLevels, level)
		}
	}
	return segments, segmentLevels, levels, rotatedValues, computed3DCollectionZ(depth)
}

func (a *Axes3D) contourLines3D(x, y []float64, z [][]float64, opt PlotOptions, zdir string) ([][]geom.Pt, []float64, []float64, []float64, bool) {
	rows, cols, ok := validate3DGridContourInput(x, y, z)
	if !ok {
		return nil, nil, nil, nil, false
	}
	values := flattenGridValues(z)
	levels := contourLevels(values, opt.Levels, opt.LevelCount, false)
	if len(levels) == 0 {
		return nil, nil, nil, nil, false
	}
	if zdir == "z" {
		lines, lineLevels := contourGridPolylines(x[:cols], y[:rows], z, levels)
		return lines, lineLevels, levels, values, true
	}

	tri, rotatedValues, ok := rotatedContourTriangulation(x[:cols], y[:rows], z, zdir)
	if !ok {
		return nil, nil, nil, nil, false
	}
	lines, lineLevels := contourPolylines(tri, rotatedValues, levels)
	return lines, lineLevels, levels, rotatedValues, true
}

func contourPlaneLevel(level float64, offset *float64) float64 {
	if offset != nil && isFinite(*offset) {
		return *offset
	}
	return level
}

func contourPolyline3D(polyline []geom.Pt, planeLevel float64, zdir string) []vec3 {
	points := make([]vec3, len(polyline))
	for i, pt := range polyline {
		points[i] = juggle3DPointSigned(pt.X, pt.Y, planeLevel, "-"+zdir)
	}
	return points
}
