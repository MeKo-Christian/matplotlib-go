package plot3d

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/diag"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

// Bar3DPlaneOptions configures Axes3D.Bar, the projected 2D bar variant.
type Bar3DPlaneOptions struct {
	Color     optional.Value[render.Color]
	Width     optional.Value[float64]
	EdgeColor optional.Value[render.Color]
	EdgeWidth optional.Value[float64]
	Alpha     optional.Value[float64]
	Baseline  optional.Value[float64]
	Baselines []float64
	Z         optional.Value[float64]
	Zs        []float64
	ZDir      string
	Label     string
}

// Bar projects 2D bars into the plane orthogonal to ZDir, matching mplot3d's
// 2D bar compatibility path rather than the cuboid Bar3D helper.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) Bar(x, heights []float64, opt Bar3DPlaneOptions) *core.PolyCollection {
	if a == nil || a.Axes == nil {
		return nil
	}
	n := minLen(x, heights)
	if n <= 0 {
		return nil
	}

	width := 0.8
	if v, ok := opt.Width.Get(); ok {
		width = v
	}
	baseline := 0.0
	if v, ok := opt.Baseline.Get(); ok {
		baseline = v
	}
	z := 0.0
	if v, ok := opt.Z.Get(); ok {
		z = v
	}
	zdir := normalized3DDir(opt.ZDir)

	limitsChanged := a.observe3DPlaneBars(x[:n], heights[:n], width, baseline, opt.Baselines, z, opt.Zs, zdir)
	color := a.NextPatchColor()
	if v, ok := opt.Color.Get(); ok {
		color = v
	}
	alpha := 1.0
	if v, ok := opt.Alpha.Get(); ok && v >= 0 && opt.Alpha.OrZero() <= 1 {
		alpha = opt.Alpha.OrZero()
	}
	color = color.WithAlphaMultiplier(alpha)
	edgeColor := render.Color{}
	if opt.EdgeColor.IsSet() {
		edgeColor = opt.EdgeColor.OrZero()
		edgeColor = edgeColor.WithAlphaMultiplier(alpha)
	}
	edgeWidth := 0.0
	if v, ok := opt.EdgeWidth.Get(); ok {
		edgeWidth = v
	}

	polygons, zorder := a.project3DPlaneBars(x[:n], heights[:n], width, baseline, opt.Baselines, z, opt.Zs, zdir)
	if len(polygons) == 0 {
		return nil
	}
	collection := &core.PolyCollection{
		Polygons: polygons,
		PatchCollection: core.PatchCollection{
			Collection: core.Collection{Coords: core.Coords(core.CoordData), Label: opt.Label, Alpha: 1},
			FaceColors: repeatColor(color, len(polygons)),
			EdgeColor:  edgeColor,
			EdgeWidth:  edgeWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	collection.SetZ(zorder)
	a.Add(collection)
	a.add3DReprojector(func() {
		polygons, zorder := a.project3DPlaneBars(x[:n], heights[:n], width, baseline, opt.Baselines, z, opt.Zs, zdir)
		collection.Polygons = polygons
		collection.FaceColors = repeatColor(color, len(polygons))
		collection.SetZ(zorder)
	}, limitsChanged)
	return collection
}

// Bar3DOptions configures projected wireframe bars.
type Bar3DOptions struct {
	Color     optional.Value[render.Color]
	Colors    []render.Color
	LineWidth optional.Value[float64]
	Alpha     optional.Value[float64]
	Label     string
	AxLimClip bool
}

// VoxelOptions configures boolean-grid voxel rendering.
type VoxelOptions struct {
	FaceColor  optional.Value[render.Color]
	FaceColors map[[3]int]render.Color
	EdgeColor  optional.Value[render.Color]
	EdgeColors map[[3]int]render.Color
	Alpha      optional.Value[float64]
	Shade      optional.Value[bool]
	Label      string
	AxLimClip  bool
}

// Bar3D draws a simple projected wireframe column for each x/y/z sample.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) Bar3D(x, y, z, dx, dy, dz []float64, opt Bar3DOptions) *core.LineCollection {
	n := minLen(x, y, z, dx, dy, dz)
	if n <= 0 || a == nil {
		return nil
	}
	limitsChanged := a.observe3DBarData(x, y, z, dx, dy, dz)

	color := a.NextColor()
	lineWidth := 1.0
	alpha := 1.0
	edgeAlpha := 0.0
	label := ""
	if v, ok := opt.Color.Get(); ok {
		color = v
	}
	if v, ok := opt.LineWidth.Get(); ok {
		lineWidth = v
		edgeAlpha = alpha
	}
	alphaSet := false
	if v, ok := opt.Alpha.Get(); ok && v >= 0 && v <= 1 {
		alpha = v
		alphaSet = true
		if opt.LineWidth.IsSet() {
			edgeAlpha = alpha
		}
	}
	label = opt.Label

	faceColor := color
	// Only an explicitly requested alpha is baked into the face color.
	if alphaSet {
		faceColor = faceColor.WithAlphaMultiplier(alpha)
	}
	faceBaseColors := bar3DFaceBaseColors(faceColor, opt.Colors, alpha, n)
	faces, faceColors := a.projectBar3DShadedFaces(x, y, z, dx, dy, dz, faceBaseColors, opt.AxLimClip)
	barZ := a.bar3DCollectionZ(x, y, z, dx, dy, dz)
	if len(faces) > 0 {
		faceCollection := &core.PolyCollection{
			Polygons: faces,
			PatchCollection: core.PatchCollection{
				Collection: core.Collection{Coords: core.Coords(core.CoordData), Alpha: 1},
				FaceColors: faceColors,
				EdgeColor:  render.Color{A: 0},
				LineJoin:   render.JoinMiter,
				LineCap:    render.CapButt,
			},
		}
		faceCollection.SetZ(barZ)
		a.Add(faceCollection)
		a.add3DReprojector(func() {
			if faceCollection != nil {
				faces, faceColors := a.projectBar3DShadedFaces(x, y, z, dx, dy, dz, faceBaseColors, opt.AxLimClip)
				faceCollection.Polygons = faces
				faceCollection.FaceColors = faceColors
				faceCollection.SetZ(a.bar3DCollectionZ(x, y, z, dx, dy, dz))
			}
		}, limitsChanged)
	}

	segments := a.projectBar3DSegments(x, y, z, dx, dy, dz, opt.AxLimClip)

	collection := &core.LineCollection{
		Collection: core.Collection{
			Coords: core.Coords(core.CoordData),
			Label:  label,
			Alpha:  edgeAlpha,
		},
		Segments:  segments,
		Color:     bar3DEdgeColor(color, edgeAlpha),
		LineWidth: lineWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	collection.SetZ(barZ)
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			collection.Segments = a.projectBar3DSegments(x, y, z, dx, dy, dz, opt.AxLimClip)
			collection.SetZ(a.bar3DCollectionZ(x, y, z, dx, dy, dz))
		}
	}, limitsChanged)
	return collection
}

func bar3DEdgeColor(color render.Color, alpha float64) render.Color {
	if alpha > 0 {
		return color
	}
	color.A = 0
	return color
}

func bar3DFaceBaseColors(defaultColor render.Color, colors []render.Color, alpha float64, bars int) []render.Color {
	totalFaces := bars * 6
	if totalFaces <= 0 {
		return nil
	}
	applyAlpha := func(color render.Color) render.Color {
		return color.WithAlphaMultiplier(alpha)
	}
	resolved := make([]render.Color, totalFaces)
	if len(colors) == 0 {
		for i := range resolved {
			resolved[i] = defaultColor
		}
		return resolved
	}
	if len(colors) == bars {
		for bar := range bars {
			color := applyAlpha(colors[bar])
			for face := 0; face < 6; face++ {
				resolved[bar*6+face] = color
			}
		}
		return resolved
	}
	if len(colors) == 6 {
		for bar := range bars {
			for face := 0; face < 6; face++ {
				resolved[bar*6+face] = applyAlpha(colors[face])
			}
		}
		return resolved
	}
	if len(colors) == totalFaces {
		for i := range resolved {
			resolved[i] = applyAlpha(colors[i])
		}
		return resolved
	}
	for i := range resolved {
		resolved[i] = applyAlpha(colors[i%len(colors)])
	}
	return resolved
}

// Voxels renders a boolean occupancy grid as per-voxel face collections with
// internal-face culling, matching Matplotlib's voxel artist model.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) Voxels(filled [][][]bool, opt VoxelOptions) map[[3]int]*core.PolyCollection {
	if a == nil || a.Axes == nil {
		return nil
	}
	_, _, _, ok := voxelGridShape(filled)
	if !ok {
		return nil
	}

	alpha := 1.0
	if v, ok := opt.Alpha.Get(); ok && v >= 0 && opt.Alpha.OrZero() <= 1 {
		alpha = opt.Alpha.OrZero()
	}
	// Resolve the default face color once here so the reprojector closure
	// does not re-call NextPatchColor() and advance the color cycle.
	if !opt.FaceColor.IsSet() {
		faceColor := a.NextPatchColor()
		opt.FaceColor = optional.Of(faceColor)
	}
	// Default edge width matches matplotlib's patch.linewidth (1.0) when an
	// edge color is configured; without a positive width the edges are invisible.
	edgeWidth := 0.0
	if edge, ok := opt.EdgeColor.Get(); ok && edge.A > 0 {
		edgeWidth = 1.0
	}
	if edgeWidth == 0 && len(opt.EdgeColors) > 0 {
		edgeWidth = 1.0
	}
	limitsChanged := a.observe3DVoxels(filled)
	projected := a.projectVoxelCollections(filled, opt, alpha)
	if len(projected) == 0 {
		return map[[3]int]*core.PolyCollection{}
	}

	collections := make(map[[3]int]*core.PolyCollection, len(projected))
	for coord, voxel := range projected {
		collection := &core.PolyCollection{
			Polygons: voxel.polygons,
			PatchCollection: core.PatchCollection{
				Collection: core.Collection{Coords: core.Coords(core.CoordData), Label: opt.Label, Alpha: 1},
				FaceColors: voxel.faceColors,
				EdgeColor:  voxel.edgeColor,
				EdgeWidth:  edgeWidth,
				LineJoin:   render.JoinMiter,
				LineCap:    render.CapButt,
			},
		}
		collection.SetZ(voxel.zorder)
		a.Add(collection)
		collections[coord] = collection
	}
	a.add3DReprojector(func() {
		refreshed := a.projectVoxelCollections(filled, opt, alpha)
		for coord, collection := range collections {
			voxel, ok := refreshed[coord]
			if !ok {
				collection.Polygons = nil
				collection.FaceColors = nil
				collection.EdgeColor = render.Color{}
				collection.SetZ(1)
				continue
			}
			collection.Polygons = voxel.polygons
			collection.FaceColors = voxel.faceColors
			collection.EdgeColor = voxel.edgeColor
			collection.EdgeWidth = edgeWidth
			collection.SetZ(voxel.zorder)
		}
	}, limitsChanged)
	return collections
}

// Voxel projects unstructured rectangular prisms as wireframe voxels by
// delegating to [Axes3D.Bar3D]; it draws edges only, not filled cubes. For
// Matplotlib's filled voxels() — a boolean occupancy grid rendered as shaded
// solid cubes — use [Axes3D.Voxels] instead. The name is retained for
// backwards compatibility.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) Voxel(x, y, z, dx, dy, dz []float64, opt core.PlotOptions) *core.LineCollection {
	if a != nil && !a.voxelWarned {
		a.voxelWarned = true
		diag.Warnf("Axes3D.Voxel draws wireframe prisms, not filled cubes; use Axes3D.Voxels(grid) for filled voxels")
	}
	return a.Bar3D(x, y, z, dx, dy, dz, Bar3DOptions{
		Color:     opt.Color,
		LineWidth: opt.LineWidth,
		Alpha:     opt.Alpha,
		AxLimClip: opt.AxLimClip,
		Label:     opt.Label,
	})
}

type projectedVoxelCollection struct {
	polygons   [][]geom.Pt
	faceColors []render.Color
	edgeColor  render.Color
	zorder     float64
}

func voxelGridShape(filled [][][]bool) (int, int, int, bool) {
	if len(filled) == 0 || len(filled[0]) == 0 || len(filled[0][0]) == 0 {
		return 0, 0, 0, false
	}
	nx, ny, nz := len(filled), len(filled[0]), len(filled[0][0])
	for i := 0; i < nx; i++ {
		if len(filled[i]) != ny {
			return 0, 0, 0, false
		}
		for j := 0; j < ny; j++ {
			if len(filled[i][j]) != nz {
				return 0, 0, 0, false
			}
		}
	}
	return nx, ny, nz, true
}

func (a *Axes3D) observe3DVoxels(filled [][][]bool) bool {
	nx, ny, nz, ok := voxelGridShape(filled)
	if !ok {
		return false
	}
	changed := false
	for i := 0; i < nx; i++ {
		for j := 0; j < ny; j++ {
			for k := 0; k < nz; k++ {
				if !filled[i][j][k] {
					continue
				}
				if a.observe3DPoint(float64(i), float64(j), float64(k)) {
					changed = true
				}
				if a.observe3DPoint(float64(i+1), float64(j+1), float64(k+1)) {
					changed = true
				}
			}
		}
	}
	return changed
}

//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) projectVoxelCollections(filled [][][]bool, opt VoxelOptions, alpha float64) map[[3]int]projectedVoxelCollection {
	nx, ny, nz, ok := voxelGridShape(filled)
	if !ok {
		return nil
	}
	defaultFaceColor := a.NextPatchColor()
	if v, ok := opt.FaceColor.Get(); ok {
		defaultFaceColor = v
	}
	defaultFaceColor = defaultFaceColor.WithAlphaMultiplier(alpha)
	defaultEdgeColor := render.Color{}
	if opt.EdgeColor.IsSet() {
		defaultEdgeColor = opt.EdgeColor.OrZero()
		defaultEdgeColor = defaultEdgeColor.WithAlphaMultiplier(alpha)
	}
	shade := true
	if v, ok := opt.Shade.Get(); ok {
		shade = v
	}

	type voxelFace struct {
		polygon []geom.Pt
		color   render.Color
		depth   float64
	}
	projected := map[[3]int]projectedVoxelCollection{}
	for i := 0; i < nx; i++ {
		for j := 0; j < ny; j++ {
			for k := 0; k < nz; k++ {
				if !filled[i][j][k] {
					continue
				}
				coord := [3]int{i, j, k}
				faceColor := defaultFaceColor
				if color, ok := opt.FaceColors[coord]; ok {
					faceColor = color
					faceColor = faceColor.WithAlphaMultiplier(alpha)
				}
				edgeColor := defaultEdgeColor
				if color, ok := opt.EdgeColors[coord]; ok {
					edgeColor = color
					edgeColor = edgeColor.WithAlphaMultiplier(alpha)
				}

				faces := make([]voxelFace, 0, 6)
				for _, raw := range voxelVisibleFaces(filled, i, j, k) {
					if opt.AxLimClip && !a.polygonWithin3DViewLimits(raw.polygon) {
						continue
					}
					polygon := make([]geom.Pt, len(raw.polygon))
					depth := 0.0
					for idx, point := range raw.polygon {
						projectedPt, zDepth := a.projectPointDepth(point[0], point[1], point[2])
						polygon[idx] = projectedPt
						depth += zDepth
					}
					color := faceColor
					if shade {
						color = shade3DFaceColor(color, raw.normal)
					}
					faces = append(faces, voxelFace{
						polygon: polygon,
						color:   color,
						depth:   depth / float64(len(raw.polygon)),
					})
				}
				if len(faces) == 0 {
					continue
				}
				sort.SliceStable(faces, func(aIdx, bIdx int) bool {
					return faces[aIdx].depth > faces[bIdx].depth
				})
				polygons := make([][]geom.Pt, len(faces))
				colors := make([]render.Color, len(faces))
				minDepth := math.Inf(1)
				for idx, face := range faces {
					polygons[idx] = face.polygon
					colors[idx] = face.color
					if face.depth < minDepth {
						minDepth = face.depth
					}
				}
				projected[coord] = projectedVoxelCollection{
					polygons:   polygons,
					faceColors: colors,
					edgeColor:  edgeColor,
					zorder:     computed3DCollectionZ(minDepth),
				}
			}
		}
	}
	return projected
}

type voxelRawFace struct {
	polygon []vec3
	normal  vec3
}

func voxelVisibleFaces(filled [][][]bool, i, j, k int) []voxelRawFace {
	nx, ny, nz, _ := voxelGridShape(filled)
	visible := make([]voxelRawFace, 0, 6)
	neighbors := []struct {
		delta  [3]int
		normal vec3
		face   []vec3
	}{
		{
			delta:  [3]int{-1, 0, 0},
			normal: vec3{-1, 0, 0},
			face: []vec3{
				{float64(i), float64(j), float64(k)},
				{float64(i), float64(j + 1), float64(k)},
				{float64(i), float64(j + 1), float64(k + 1)},
				{float64(i), float64(j), float64(k + 1)},
			},
		},
		{
			delta:  [3]int{1, 0, 0},
			normal: vec3{1, 0, 0},
			face: []vec3{
				{float64(i + 1), float64(j), float64(k)},
				{float64(i + 1), float64(j), float64(k + 1)},
				{float64(i + 1), float64(j + 1), float64(k + 1)},
				{float64(i + 1), float64(j + 1), float64(k)},
			},
		},
		{
			delta:  [3]int{0, -1, 0},
			normal: vec3{0, -1, 0},
			face: []vec3{
				{float64(i), float64(j), float64(k)},
				{float64(i), float64(j), float64(k + 1)},
				{float64(i + 1), float64(j), float64(k + 1)},
				{float64(i + 1), float64(j), float64(k)},
			},
		},
		{
			delta:  [3]int{0, 1, 0},
			normal: vec3{0, 1, 0},
			face: []vec3{
				{float64(i), float64(j + 1), float64(k)},
				{float64(i + 1), float64(j + 1), float64(k)},
				{float64(i + 1), float64(j + 1), float64(k + 1)},
				{float64(i), float64(j + 1), float64(k + 1)},
			},
		},
		{
			delta:  [3]int{0, 0, -1},
			normal: vec3{0, 0, -1},
			face: []vec3{
				{float64(i), float64(j), float64(k)},
				{float64(i + 1), float64(j), float64(k)},
				{float64(i + 1), float64(j + 1), float64(k)},
				{float64(i), float64(j + 1), float64(k)},
			},
		},
		{
			delta:  [3]int{0, 0, 1},
			normal: vec3{0, 0, 1},
			face: []vec3{
				{float64(i), float64(j), float64(k + 1)},
				{float64(i), float64(j + 1), float64(k + 1)},
				{float64(i + 1), float64(j + 1), float64(k + 1)},
				{float64(i + 1), float64(j), float64(k + 1)},
			},
		},
	}
	for _, neighbor := range neighbors {
		ni := i + neighbor.delta[0]
		nj := j + neighbor.delta[1]
		nk := k + neighbor.delta[2]
		if ni >= 0 && ni < nx && nj >= 0 && nj < ny && nk >= 0 && nk < nz && filled[ni][nj][nk] {
			continue
		}
		visible = append(visible, voxelRawFace{polygon: neighbor.face, normal: neighbor.normal})
	}
	return visible
}
