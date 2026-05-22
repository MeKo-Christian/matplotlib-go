package agg

import (
	"math"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func transformMarkerPath(path geom.Path, affine geom.Affine, offset geom.Pt) geom.Path {
	if len(path.C) == 0 {
		return geom.Path{}
	}
	out := geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: make([]geom.Pt, len(path.V)),
	}
	for i, pt := range path.V {
		pt = affine.Apply(pt)
		out.V[i] = geom.Pt{X: pt.X + offset.X, Y: pt.Y + offset.Y}
	}
	return out
}

func (r *Renderer) drawGouraudTriangle(tri *render.GouraudTriangle) {
	if tri == nil {
		return
	}
	img := r.ctx.image
	if img == nil || img.Width() <= 0 || img.Height() <= 0 {
		return
	}

	minX := int(math.Floor(math.Min(tri.P[0].X, math.Min(tri.P[1].X, tri.P[2].X))))
	maxX := int(math.Ceil(math.Max(tri.P[0].X, math.Max(tri.P[1].X, tri.P[2].X))))
	minY := int(math.Floor(math.Min(tri.P[0].Y, math.Min(tri.P[1].Y, tri.P[2].Y))))
	maxY := int(math.Ceil(math.Max(tri.P[0].Y, math.Max(tri.P[1].Y, tri.P[2].Y))))

	clipMinX, clipMinY := 0, 0
	clipMaxX, clipMaxY := img.Width()-1, img.Height()-1
	if r.clipRect != nil {
		clipMinX = maxInt(clipMinX, int(math.Floor(r.clipRect.Min.X)))
		clipMinY = maxInt(clipMinY, int(math.Floor(r.clipRect.Min.Y)))
		clipMaxX = minInt(clipMaxX, int(math.Ceil(r.clipRect.Max.X))-1)
		clipMaxY = minInt(clipMaxY, int(math.Ceil(r.clipRect.Max.Y))-1)
	}
	minX = maxInt(minX, clipMinX)
	minY = maxInt(minY, clipMinY)
	maxX = minInt(maxX, clipMaxX)
	maxY = minInt(maxY, clipMaxY)
	if minX > maxX || minY > maxY {
		return
	}

	area := edgeFunction(tri.P[0], tri.P[1], tri.P[2])
	if area == 0 || math.IsNaN(area) || math.IsInf(area, 0) {
		return
	}

	stride := img.Stride()
	if stride <= 0 {
		return
	}
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			p := geom.Pt{X: float64(x) + 0.5, Y: float64(y) + 0.5}
			w0 := edgeFunction(tri.P[1], tri.P[2], p) / area
			w1 := edgeFunction(tri.P[2], tri.P[0], p) / area
			w2 := edgeFunction(tri.P[0], tri.P[1], p) / area
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			src := interpolateColor(tri.Color[0], tri.Color[1], tri.Color[2], w0, w1, w2)
			if src.A <= 0 {
				continue
			}
			off := y*stride + x*4
			if off < 0 || off+3 >= len(img.Data) {
				continue
			}
			blendPixelRGBA(img.Data[off:off+4], src)
		}
	}
}

func edgeFunction(a, b, c geom.Pt) float64 {
	return (c.X-a.X)*(b.Y-a.Y) - (c.Y-a.Y)*(b.X-a.X)
}

func interpolateColor(c0, c1, c2 render.Color, w0, w1, w2 float64) render.Color {
	return render.Color{
		R: c0.R*w0 + c1.R*w1 + c2.R*w2,
		G: c0.G*w0 + c1.G*w1 + c2.G*w2,
		B: c0.B*w0 + c1.B*w1 + c2.B*w2,
		A: c0.A*w0 + c1.A*w1 + c2.A*w2,
	}
}

func blendPixelRGBA(dst []uint8, src render.Color) {
	sa := uint32(math.Round(clamp01(src.A) * 255))
	if sa == 0 {
		return
	}
	sr := uint8(math.Round(clamp01(src.R) * 255))
	sg := uint8(math.Round(clamp01(src.G) * 255))
	sb := uint8(math.Round(clamp01(src.B) * 255))
	if sa >= 255 {
		dst[0] = sr
		dst[1] = sg
		dst[2] = sb
		dst[3] = 255
		return
	}

	da := uint32(dst[3])
	combinedA := ((sa + da) << 8) - sa*da
	if combinedA == 0 {
		dst[0], dst[1], dst[2], dst[3] = 0, 0, 0, 0
		return
	}
	dst[3] = uint8(combinedA >> 8)
	dst[0] = fixedBlendChannel(dst[0], sr, uint8(da), uint8(sa), combinedA)
	dst[1] = fixedBlendChannel(dst[1], sg, uint8(da), uint8(sa), combinedA)
	dst[2] = fixedBlendChannel(dst[2], sb, uint8(da), uint8(sa), combinedA)
}

func fixedBlendChannel(dst, src, dstA, srcA uint8, combinedA uint32) uint8 {
	dstPremul := int64(uint32(dst) * uint32(dstA))
	numerator := ((int64(src) << 8) - dstPremul) * int64(srcA)
	numerator += dstPremul << 8
	return uint8(numerator / int64(combinedA))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// renderColorToAGG converts a normalized render.Color to AGG's 8-bit SRGBA
// color type without applying any transfer-curve conversion.
func renderColorToAGG(c render.Color) agglib.Color {
	return agglib.NewColor(
		uint8(math.Round(clamp01(c.R)*255)),
		uint8(math.Round(clamp01(c.G)*255)),
		uint8(math.Round(clamp01(c.B)*255)),
		uint8(math.Round(clamp01(c.A)*255)),
	)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// quantize snaps values for cache keys and text metrics. Path rasterization
// itself uses explicit snapping/simplification policy instead of this grid.
const quantizationGrid = 1e-6

func quantize(v float64) float64 {
	return math.Round(v/quantizationGrid) * quantizationGrid
}

func quantizePt(p geom.Pt) geom.Pt {
	return geom.Pt{X: quantize(p.X), Y: quantize(p.Y)}
}

// SupportsNativeHatch reports that AGG consumes render.Paint hatch metadata
// directly while rasterizing a path.
func (r *Renderer) SupportsNativeHatch() bool { return true }

func applyForcedAlpha(paint *render.Paint) {
	if paint == nil || !paint.ForceAlpha {
		return
	}
	alpha := clamp01(paint.Alpha)
	if paint.Stroke.A > 0 {
		paint.Stroke.A = alpha
	}
	if paint.Fill.A > 0 {
		paint.Fill.A = alpha
	}
	if paint.HatchColor.A > 0 {
		paint.HatchColor.A = alpha
	}
}

func colorWithForcedAlpha(c render.Color, paint *render.Paint) render.Color {
	if paint != nil && paint.ForceAlpha && c.A > 0 {
		c.A = clamp01(paint.Alpha)
	}
	return c
}

func (r *Renderer) applyAntialiasMode(mode render.AntialiasMode) func() {
	if r.ctx == nil {
		return func() {}
	}
	prev := r.ctx.GetAntiAliasGamma()
	switch mode {
	case render.AntialiasOn:
		r.ctx.SetAntiAliasGamma(1.0)
	case render.AntialiasOff:
		// AGG exposes antialiasing through the rasterizer gamma curve rather
		// than a boolean switch. A low gamma sharply suppresses partial
		// coverage and gives callers an aliased-style path when requested.
		r.ctx.SetAntiAliasGamma(0.1)
	default:
		return func() {}
	}
	return func() {
		r.ctx.SetAntiAliasGamma(prev)
	}
}

const defaultPathChunkVertices = 32768

func (r *Renderer) preparePathForPaint(path geom.Path, paint *render.Paint) (geom.Path, bool) {
	path = removeNonFinitePathVertices(path)
	if len(path.C) == 0 || !path.Validate() {
		return geom.Path{}, false
	}
	if r.pathOutsideVisibleArea(path, paint) {
		return geom.Path{}, false
	}
	if shouldSnapPath(path, paint) {
		path = snapPath(path, paint)
	}
	if paint.Simplify && paint.SimplifyThreshold > 0 {
		path = simplifyLinePath(path, paint.SimplifyThreshold)
	}
	return path, len(path.C) > 0
}

func removeNonFinitePathVertices(path geom.Path) geom.Path {
	out := geom.Path{}
	vi := 0
	haveCurrent := false
	needMove := true

	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				return out
			}
			to := path.V[vi]
			vi++
			if !finitePt(to) {
				haveCurrent = false
				needMove = true
				continue
			}
			out.MoveTo(to)
			haveCurrent = true
			needMove = false
		case geom.LineTo:
			if vi >= len(path.V) {
				return out
			}
			to := path.V[vi]
			vi++
			if !finitePt(to) {
				haveCurrent = false
				needMove = true
				continue
			}
			if !haveCurrent || needMove {
				out.MoveTo(to)
			} else {
				out.LineTo(to)
			}
			haveCurrent = true
			needMove = false
		case geom.QuadTo:
			if vi+1 >= len(path.V) {
				return out
			}
			ctrl, to := path.V[vi], path.V[vi+1]
			vi += 2
			if !finitePt(ctrl) || !finitePt(to) {
				haveCurrent = false
				needMove = true
				continue
			}
			if !haveCurrent || needMove {
				out.MoveTo(to)
			} else {
				out.QuadTo(ctrl, to)
			}
			haveCurrent = true
			needMove = false
		case geom.CubicTo:
			if vi+2 >= len(path.V) {
				return out
			}
			c1, c2, to := path.V[vi], path.V[vi+1], path.V[vi+2]
			vi += 3
			if !finitePt(c1) || !finitePt(c2) || !finitePt(to) {
				haveCurrent = false
				needMove = true
				continue
			}
			if !haveCurrent || needMove {
				out.MoveTo(to)
			} else {
				out.CubicTo(c1, c2, to)
			}
			haveCurrent = true
			needMove = false
		case geom.ClosePath:
			if haveCurrent && !needMove {
				out.Close()
			}
			haveCurrent = false
			needMove = true
		}
	}
	return out
}

func finitePt(pt geom.Pt) bool {
	return !math.IsNaN(pt.X) && !math.IsInf(pt.X, 0) && !math.IsNaN(pt.Y) && !math.IsInf(pt.Y, 0)
}

func (r *Renderer) pathOutsideVisibleArea(path geom.Path, paint *render.Paint) bool {
	bounds, ok := pathBounds(path)
	if !ok {
		return true
	}
	visible := geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: float64(r.width), Y: float64(r.height)}}
	if r.viewport != (geom.Rect{}) {
		visible = visible.Intersect(r.viewport)
	}
	if r.clipRect != nil {
		visible = visible.Intersect(*r.clipRect)
	}
	if visible.W() <= 0 || visible.H() <= 0 {
		return true
	}

	pad := 1.0
	if paint != nil && paint.Stroke.A > 0 && paint.LineWidth > 0 {
		pad += paint.LineWidth / 2
	}
	return !rectsOverlap(bounds.Inflate(pad, pad), visible)
}

func pathBounds(path geom.Path) (geom.Rect, bool) {
	if len(path.V) == 0 {
		return geom.Rect{}, false
	}
	var bounds geom.Rect
	ok := false
	for _, pt := range path.V {
		if !finitePt(pt) {
			continue
		}
		if !ok {
			bounds = geom.Rect{Min: pt, Max: pt}
			ok = true
			continue
		}
		if pt.X < bounds.Min.X {
			bounds.Min.X = pt.X
		}
		if pt.Y < bounds.Min.Y {
			bounds.Min.Y = pt.Y
		}
		if pt.X > bounds.Max.X {
			bounds.Max.X = pt.X
		}
		if pt.Y > bounds.Max.Y {
			bounds.Max.Y = pt.Y
		}
	}
	return bounds, ok
}

func pathDrawBounds(path geom.Path, paint *render.Paint) (geom.Rect, bool) {
	if paint == nil {
		return geom.Rect{}, false
	}
	bounds, ok := pathBounds(path)
	if !ok {
		return geom.Rect{}, false
	}
	pad := 2.0
	if paint.Stroke.A > 0 && paint.LineWidth > 0 {
		pad += paint.LineWidth / 2
	}
	if paint.Hatch != "" && paint.HatchLineWidth > 0 {
		pad = math.Max(pad, paint.HatchLineWidth/2+2)
	}
	return bounds.Inflate(pad, pad), true
}

func imageDrawBounds(dst geom.Rect) (geom.Rect, bool) {
	minX := math.Min(dst.Min.X, dst.Max.X)
	minY := math.Min(dst.Min.Y, dst.Max.Y)
	maxX := math.Max(dst.Min.X, dst.Max.X)
	maxY := math.Max(dst.Min.Y, dst.Max.Y)
	if minX == maxX || minY == maxY {
		return geom.Rect{}, false
	}
	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}.Inflate(2, 2), true
}

func transformedImageDrawBounds(img render.Image, affine geom.Affine) (geom.Rect, bool) {
	if img == nil {
		return geom.Rect{}, false
	}
	w, h := img.Size()
	if w <= 0 || h <= 0 {
		return geom.Rect{}, false
	}
	return pointsBounds([]geom.Pt{
		affine.Apply(geom.Pt{}),
		affine.Apply(geom.Pt{X: float64(w)}),
		affine.Apply(geom.Pt{Y: float64(h)}),
		affine.Apply(geom.Pt{X: float64(w), Y: float64(h)}),
	}, 2)
}

func imageTransformDisplaySpan(img render.Image, affine geom.Affine) (float64, float64) {
	w, h := img.Size()
	if w <= 0 || h <= 0 {
		return 0, 0
	}

	bounds, ok := pointsBounds([]geom.Pt{
		affine.Apply(geom.Pt{}),
		affine.Apply(geom.Pt{X: float64(w)}),
		affine.Apply(geom.Pt{Y: float64(h)}),
		affine.Apply(geom.Pt{X: float64(w), Y: float64(h)}),
	}, 0)
	if !ok {
		return 0, 0
	}
	return bounds.W(), bounds.H()
}

func gouraudTriangleBatchBounds(batch render.GouraudTriangleBatch) (geom.Rect, bool) {
	points := make([]geom.Pt, 0, len(batch.Triangles)*3)
	for i := range batch.Triangles {
		points = append(points, batch.Triangles[i].P[:]...)
	}
	return pointsBounds(points, 1)
}

func pointsBounds(points []geom.Pt, pad float64) (geom.Rect, bool) {
	var bounds geom.Rect
	ok := false
	for _, pt := range points {
		if !finitePt(pt) {
			continue
		}
		if !ok {
			bounds = geom.Rect{Min: pt, Max: pt}
			ok = true
			continue
		}
		if pt.X < bounds.Min.X {
			bounds.Min.X = pt.X
		}
		if pt.Y < bounds.Min.Y {
			bounds.Min.Y = pt.Y
		}
		if pt.X > bounds.Max.X {
			bounds.Max.X = pt.X
		}
		if pt.Y > bounds.Max.Y {
			bounds.Max.Y = pt.Y
		}
	}
	if !ok {
		return geom.Rect{}, false
	}
	return bounds.Inflate(pad, pad), true
}

func rectsOverlap(a, b geom.Rect) bool {
	return a.Max.X >= b.Min.X && b.Max.X >= a.Min.X && a.Max.Y >= b.Min.Y && b.Max.Y >= a.Min.Y
}

func shouldSnapPath(path geom.Path, paint *render.Paint) bool {
	switch paint.Snap {
	case render.SnapOn:
		return true
	case render.SnapOff:
		return false
	case render.SnapAuto:
	default:
		return false
	}
	if len(path.V) > 1024 {
		return false
	}
	vi := 0
	var last geom.Pt
	haveLast := false
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				return false
			}
			last = path.V[vi]
			vi++
			haveLast = true
		case geom.LineTo:
			if vi >= len(path.V) {
				return false
			}
			to := path.V[vi]
			vi++
			if haveLast && math.Abs(last.X-to.X) >= 1e-4 && math.Abs(last.Y-to.Y) >= 1e-4 {
				return false
			}
			last = to
			haveLast = true
		case geom.QuadTo, geom.CubicTo:
			return false
		case geom.ClosePath:
			haveLast = false
		}
	}
	return true
}

func snapPath(path geom.Path, paint *render.Paint) geom.Path {
	out := clonePath(path)
	snapValue := 0.0
	strokeWidth := 0.0
	if paint.Stroke.A > 0 && paint.LineWidth > 0 {
		strokeWidth = paint.LineWidth
	}
	if int(math.Round(strokeWidth))%2 != 0 {
		snapValue = 0.5
	}
	for i, pt := range out.V {
		out.V[i] = geom.Pt{
			X: math.Floor(pt.X+0.5) + snapValue,
			Y: math.Floor(pt.Y+0.5) + snapValue,
		}
	}
	return out
}

func simplifyLinePath(path geom.Path, threshold float64) geom.Path {
	if threshold <= 0 || pathHasCurvesOrClose(path) {
		return path
	}
	out := geom.Path{}
	var current []geom.Pt
	flush := func() {
		if len(current) == 0 {
			return
		}
		points := simplifyPolyline(current, threshold)
		if len(points) > 0 {
			out.MoveTo(points[0])
			for _, pt := range points[1:] {
				out.LineTo(pt)
			}
		}
		current = current[:0]
	}

	vi := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			flush()
			current = append(current, path.V[vi])
			vi++
		case geom.LineTo:
			current = append(current, path.V[vi])
			vi++
		}
	}
	flush()
	return out
}

func pathHasCurvesOrClose(path geom.Path) bool {
	for _, cmd := range path.C {
		if cmd == geom.QuadTo || cmd == geom.CubicTo || cmd == geom.ClosePath {
			return true
		}
	}
	return false
}

func simplifyPolyline(points []geom.Pt, threshold float64) []geom.Pt {
	if len(points) <= 2 {
		return append([]geom.Pt(nil), points...)
	}
	keep := make([]bool, len(points))
	keep[0] = true
	keep[len(points)-1] = true
	simplifyPolylineRange(points, threshold*threshold, 0, len(points)-1, keep)
	out := make([]geom.Pt, 0, len(points))
	for i, pt := range points {
		if keep[i] {
			out = append(out, pt)
		}
	}
	return out
}

func simplifyPolylineRange(points []geom.Pt, threshold2 float64, first, last int, keep []bool) {
	if last <= first+1 {
		return
	}
	maxDist2 := -1.0
	maxIndex := -1
	for i := first + 1; i < last; i++ {
		dist2 := pointSegmentDistanceSquared(points[i], points[first], points[last])
		if dist2 > maxDist2 {
			maxDist2 = dist2
			maxIndex = i
		}
	}
	if maxDist2 > threshold2 && maxIndex >= 0 {
		keep[maxIndex] = true
		simplifyPolylineRange(points, threshold2, first, maxIndex, keep)
		simplifyPolylineRange(points, threshold2, maxIndex, last, keep)
	}
}

func pointSegmentDistanceSquared(p, a, b geom.Pt) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if dx == 0 && dy == 0 {
		return squaredDistance(p, a)
	}
	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	proj := geom.Pt{X: a.X + t*dx, Y: a.Y + t*dy}
	return squaredDistance(p, proj)
}

func squaredDistance(a, b geom.Pt) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return dx*dx + dy*dy
}

func chunkStrokePath(path geom.Path, maxVertices int) []geom.Path {
	if maxVertices <= 0 {
		maxVertices = defaultPathChunkVertices
	}
	if len(path.V) <= maxVertices || pathHasCurvesOrClose(path) {
		return []geom.Path{path}
	}

	chunks := make([]geom.Path, 0, len(path.V)/maxVertices+1)
	vi := 0
	var current geom.Path
	currentVertices := 0
	haveCurrent := false
	var last geom.Pt

	flush := func() {
		if len(current.C) > 1 {
			chunks = append(chunks, current)
		}
		current = geom.Path{}
		currentVertices = 0
		haveCurrent = false
	}

	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				flush()
				return chunks
			}
			if currentVertices >= maxVertices {
				flush()
			}
			last = path.V[vi]
			vi++
			current.MoveTo(last)
			currentVertices++
			haveCurrent = true
		case geom.LineTo:
			if vi >= len(path.V) {
				flush()
				return chunks
			}
			to := path.V[vi]
			vi++
			if !haveCurrent {
				current.MoveTo(to)
				currentVertices++
			} else if currentVertices >= maxVertices {
				flush()
				current.MoveTo(last)
				currentVertices++
			}
			current.LineTo(to)
			currentVertices++
			last = to
			haveCurrent = true
		}
	}
	flush()
	if len(chunks) == 0 {
		return []geom.Path{path}
	}
	return chunks
}

func (r *Renderer) drawNativeHatch(clipPath geom.Path, paint *render.Paint) {
	if paint == nil || paint.Hatch == "" {
		return
	}
	color := colorWithForcedAlpha(paint.HatchColor, paint)
	if color.A <= 0 {
		return
	}
	bounds, ok := pathBounds(clipPath)
	if !ok {
		return
	}
	counts := hatchCounts(paint.Hatch)
	if len(counts) == 0 {
		return
	}

	oldPaths := r.clipPaths
	r.clipPaths = append(clonePaths(oldPaths), clonePath(clipPath))
	defer func() {
		r.clipPaths = oldPaths
	}()

	for pattern, count := range counts {
		spacing := math.Max(2, 32/float64(count))
		if paint.HatchSpacing > 0 {
			spacing = math.Max(2, paint.HatchSpacing/float64(count))
		}
		hatchPaint := render.Paint{
			Stroke:    color,
			LineWidth: paint.HatchLineWidth,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
			Antialias: paint.Antialias,
			Snap:      render.SnapOff,
		}
		if hatchPaint.LineWidth <= 0 {
			hatchPaint.LineWidth = 1
		}
		for _, hatchPath := range hatchPatternPaths(pattern, bounds, spacing) {
			if len(hatchPath.C) == 0 {
				continue
			}
			r.Path(hatchPath, &hatchPaint)
		}
	}
}

func hatchCounts(pattern string) map[rune]int {
	counts := make(map[rune]int)
	for _, ch := range pattern {
		switch ch {
		case '|', '-', '/', '\\', '+', 'x', 'X':
			counts[ch]++
		}
	}
	return counts
}

func hatchPatternPaths(pattern rune, bounds geom.Rect, spacing float64) []geom.Path {
	switch pattern {
	case '|':
		return []geom.Path{verticalHatchPath(bounds, spacing)}
	case '-':
		return []geom.Path{horizontalHatchPath(bounds, spacing)}
	case '/':
		return []geom.Path{slashHatchPath(bounds, spacing)}
	case '\\':
		return []geom.Path{backslashHatchPath(bounds, spacing)}
	case '+':
		return []geom.Path{
			verticalHatchPath(bounds, spacing),
			horizontalHatchPath(bounds, spacing),
		}
	case 'x', 'X':
		return []geom.Path{
			slashHatchPath(bounds, spacing),
			backslashHatchPath(bounds, spacing),
		}
	default:
		return nil
	}
}

func verticalHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	var path geom.Path
	minX := math.Floor(bounds.Min.X/spacing)*spacing - spacing
	maxX := bounds.Max.X + spacing
	for x := minX; x <= maxX; x += spacing {
		path.MoveTo(geom.Pt{X: x, Y: bounds.Min.Y - spacing})
		path.LineTo(geom.Pt{X: x, Y: bounds.Max.Y + spacing})
	}
	return path
}

func horizontalHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	var path geom.Path
	minY := math.Floor(bounds.Min.Y/spacing)*spacing - spacing
	maxY := bounds.Max.Y + spacing
	for y := minY; y <= maxY; y += spacing {
		path.MoveTo(geom.Pt{X: bounds.Min.X - spacing, Y: y})
		path.LineTo(geom.Pt{X: bounds.Max.X + spacing, Y: y})
	}
	return path
}

func slashHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	var path geom.Path
	width := bounds.W()
	height := bounds.H()
	extent := width + height + 2*spacing
	start := bounds.Min.X - height - spacing
	end := bounds.Max.X + spacing
	for x := start; x <= end; x += spacing {
		path.MoveTo(geom.Pt{X: x, Y: bounds.Max.Y + spacing})
		path.LineTo(geom.Pt{X: x + extent, Y: bounds.Min.Y - spacing})
	}
	return path
}

func backslashHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	var path geom.Path
	width := bounds.W()
	height := bounds.H()
	extent := width + height + 2*spacing
	start := bounds.Min.X - height - spacing
	end := bounds.Max.X + spacing
	for x := start; x <= end; x += spacing {
		path.MoveTo(geom.Pt{X: x, Y: bounds.Min.Y - spacing})
		path.LineTo(geom.Pt{X: x + extent, Y: bounds.Max.Y + spacing})
	}
	return path
}
