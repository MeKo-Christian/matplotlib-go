//go:build skia

package skia

import (
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/vector"
)

type trackedState struct {
	clipRect  *geom.Rect
	clipPaths []geom.Path
}

type bridgeDrawState struct {
	clipRect  *geom.Rect
	clipPaths []geom.Path
}

type surfaceBridge interface {
	Info() BridgeInfo
	SupportsShaders() bool
	DrawPathFill(dst *image.RGBA, path geom.Path, paint render.Paint, state bridgeDrawState) bool
}

type cpuSurfaceBridge struct {
	width  int
	height int
}

func newCPUSurfaceBridge(width, height int) surfaceBridge {
	return &cpuSurfaceBridge{width: width, height: height}
}

func (b *cpuSurfaceBridge) Info() BridgeInfo {
	return BridgeInfo{
		Binding:         BindingExternalCAPI,
		Mode:            ModeCPU,
		NativeSurface:   false,
		SupportsShaders: true,
		Description:     "CPU Skia bridge boundary with local deterministic shader fallback",
	}
}

func (b *cpuSurfaceBridge) SupportsShaders() bool { return true }

func (b *cpuSurfaceBridge) DrawPathFill(dst *image.RGBA, path geom.Path, paint render.Paint, state bridgeDrawState) bool {
	if dst == nil || !path.Validate() {
		return false
	}
	shader, ok := newFillShader(paint)
	if !ok {
		return false
	}
	bounds := shaderDrawBounds(dst.Bounds(), path, state.clipRect)
	if bounds.Empty() {
		return true
	}

	pathMask := rasterizeMask(dst.Bounds().Dx(), dst.Bounds().Dy(), path)
	if pathMask == nil {
		return false
	}
	var clipMasks []*image.Alpha
	for _, clip := range state.clipPaths {
		mask := rasterizeMask(dst.Bounds().Dx(), dst.Bounds().Dy(), clip)
		if mask == nil {
			return true
		}
		clipMasks = append(clipMasks, mask)
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			coverage := alphaAt(pathMask, x, y)
			if coverage == 0 {
				continue
			}
			for _, mask := range clipMasks {
				coverage = uint8(uint32(coverage) * uint32(alphaAt(mask, x, y)) / 255)
				if coverage == 0 {
					break
				}
			}
			if coverage == 0 {
				continue
			}
			// The path mask and clip masks above were rasterized in y-down
			// device space, but gradient/pattern shaders are defined in y-up
			// display space (the same space their endpoints/cells arrive in).
			// Map this device pixel back to display space before sampling so the
			// shader lines up with the rest of the (gobasic-flipped) scene.
			src := shader.ColorAt(geom.Pt{X: float64(x) + 0.5, Y: float64(b.height) - (float64(y) + 0.5)})
			src.A = uint8(uint32(src.A) * uint32(coverage) / 255)
			blendPixel(dst, x, y, src)
		}
	}
	return true
}

type fillShader interface {
	ColorAt(p geom.Pt) color.RGBA
}

func newFillShader(paint render.Paint) (fillShader, bool) {
	if paint.FillGradient.Kind != render.GradientNone && len(paint.FillGradient.Stops) > 0 {
		return newGradientShader(paint.FillGradient, &paint), true
	}
	if paint.Hatch == "" && (paint.FillPattern.ID != "" || len(paint.FillPattern.Path.V) > 0) {
		return newPatternShader(paint.FillPattern, &paint), true
	}
	return nil, false
}

type gradientShader struct {
	gradient render.GradientFill
	stops    []render.GradientStop
	inverse  geom.Affine
	hasInv   bool
	paint    *render.Paint
}

func newGradientShader(gradient render.GradientFill, paint *render.Paint) *gradientShader {
	stops := append([]render.GradientStop(nil), gradient.Stops...)
	sort.SliceStable(stops, func(i, j int) bool {
		return stops[i].Offset < stops[j].Offset
	})
	for i := range stops {
		stops[i].Offset = clamp01(stops[i].Offset)
	}
	shader := &gradientShader{gradient: gradient, stops: stops, paint: paint}
	if gradient.HasTransform {
		if inv, ok := gradient.Transform.Invert(); ok {
			shader.inverse = inv
			shader.hasInv = true
		}
	}
	return shader
}

func (s *gradientShader) ColorAt(p geom.Pt) color.RGBA {
	if s.hasInv {
		p = s.inverse.Apply(p)
	}
	var t float64
	switch s.gradient.Kind {
	case render.LinearGradient:
		dx := s.gradient.End.X - s.gradient.Start.X
		dy := s.gradient.End.Y - s.gradient.Start.Y
		denom := dx*dx + dy*dy
		if denom == 0 {
			t = 0
		} else {
			t = ((p.X-s.gradient.Start.X)*dx + (p.Y-s.gradient.Start.Y)*dy) / denom
		}
	case render.RadialGradient:
		if s.gradient.Radius <= 0 {
			t = 0
		} else {
			t = math.Hypot(p.X-s.gradient.Center.X, p.Y-s.gradient.Center.Y) / s.gradient.Radius
		}
	default:
		t = 0
	}
	return renderColorToRGBA(s.interpolate(clamp01(t), s.paint))
}

func (s *gradientShader) interpolate(t float64, paint *render.Paint) render.Color {
	if len(s.stops) == 0 {
		return render.Color{}
	}
	if t <= s.stops[0].Offset {
		return colorWithForcedAlpha(s.stops[0].Color, paint)
	}
	last := s.stops[len(s.stops)-1]
	if t >= last.Offset {
		return colorWithForcedAlpha(last.Color, paint)
	}
	for i := 1; i < len(s.stops); i++ {
		left := s.stops[i-1]
		right := s.stops[i]
		if t > right.Offset {
			continue
		}
		span := right.Offset - left.Offset
		local := 0.0
		if span > 0 {
			local = (t - left.Offset) / span
		}
		lc := colorWithForcedAlpha(left.Color, paint)
		rc := colorWithForcedAlpha(right.Color, paint)
		return render.Color{
			R: lc.R + (rc.R-lc.R)*local,
			G: lc.G + (rc.G-lc.G)*local,
			B: lc.B + (rc.B-lc.B)*local,
			A: lc.A + (rc.A-lc.A)*local,
		}
	}
	return colorWithForcedAlpha(last.Color, paint)
}

type patternShader struct {
	pattern render.PatternFill
	cell    geom.Rect
	paint   *render.Paint
}

func newPatternShader(pattern render.PatternFill, paint *render.Paint) *patternShader {
	cell := pattern.Cell
	if cell.W() <= 0 || cell.H() <= 0 {
		if bounds, ok := pathBounds(pattern.Path); ok {
			cell = bounds
		}
	}
	if cell.W() <= 0 {
		cell.Max.X = cell.Min.X + 16
	}
	if cell.H() <= 0 {
		cell.Max.Y = cell.Min.Y + 16
	}
	return &patternShader{pattern: pattern, cell: cell, paint: paint}
}

func (s *patternShader) ColorAt(p geom.Pt) color.RGBA {
	bg := colorWithForcedAlpha(s.pattern.Background, s.paint)
	fg := colorWithForcedAlpha(s.pattern.Foreground, s.paint)
	if len(s.pattern.Path.C) == 0 || fg.A <= 0 {
		return renderColorToRGBA(bg)
	}
	if s.pointInPatternPath(p) {
		return renderColorToRGBA(fg)
	}
	return renderColorToRGBA(bg)
}

func (s *patternShader) pointInPatternPath(p geom.Pt) bool {
	cellW := s.cell.W()
	cellH := s.cell.H()
	if cellW <= 0 || cellH <= 0 {
		return false
	}
	tileX := s.cell.Min.X + math.Floor((p.X-s.cell.Min.X)/cellW)*cellW
	tileY := s.cell.Min.Y + math.Floor((p.Y-s.cell.Min.Y)/cellH)*cellH
	q := geom.Pt{X: p.X - tileX, Y: p.Y - tileY}
	if s.pattern.HasTransform {
		if inv, ok := s.pattern.Transform.Invert(); ok {
			q = inv.Apply(q)
		}
	}
	return pointInPath(q, s.pattern.Path)
}

func (r *Renderer) drawShaderFill(path geom.Path, paint *render.Paint) bool {
	if r.bridge == nil || paint == nil {
		return false
	}
	if _, ok := newFillShader(*paint); !ok {
		return false
	}
	shaderPaint := *paint
	applyForcedAlpha(&shaderPaint)
	// The bridge rasterizes into the y-down image buffer; flip the fill path and
	// clips to device space here so the mask aligns with the buffer (the bridge
	// maps each device pixel back to display space when sampling the shader). The
	// hatch/stroke fallbacks below keep the original display-space path because
	// r.Renderer.Path (gobasic) owns its own device flip.
	height := float64(r.height)
	if !r.bridge.DrawPathFill(r.GetImage(), flipPathY(path, height), shaderPaint, bridgeDrawState{
		clipRect:  flipRectPtrY(r.clipRect, height),
		clipPaths: flipPathsY(r.clipPaths, height),
	}) {
		return false
	}

	if shaderPaint.Hatch != "" {
		hatchPaint := shaderPaint
		hatchPaint.Fill = render.Color{}
		hatchPaint.FillGradient = render.GradientFill{}
		hatchPaint.FillPattern = render.PatternFill{}
		if !render.DrawHatchFallback(r, path, hatchPaint) {
			r.Renderer.Path(path, &hatchPaint)
		}
	}
	if shaderPaint.Stroke.A > 0 && shaderPaint.LineWidth > 0 {
		strokePaint := shaderPaint
		strokePaint.Fill = render.Color{}
		strokePaint.FillGradient = render.GradientFill{}
		strokePaint.FillPattern = render.PatternFill{}
		strokePaint.Hatch = ""
		r.Renderer.Path(path, &strokePaint)
	}
	return true
}

func (r *Renderer) drawHatchPrecedencePath(path geom.Path, paint *render.Paint) bool {
	if paint == nil || paint.Hatch == "" || paint.FillGradient.Kind != render.GradientNone || !hasPatternFill(paint.FillPattern) {
		return false
	}
	hatchPaint := *paint
	applyForcedAlpha(&hatchPaint)
	if hatchPaint.Fill.A > 0 {
		fillPaint := hatchPaint
		fillPaint.FillPattern = render.PatternFill{}
		fillPaint.Hatch = ""
		fillPaint.HatchColor = render.Color{}
		fillPaint.Stroke = render.Color{}
		r.Renderer.Path(path, &fillPaint)
	}
	fallbackPaint := hatchPaint
	fallbackPaint.Fill = render.Color{}
	fallbackPaint.FillPattern = render.PatternFill{}
	fallbackPaint.Stroke = render.Color{}
	if !render.DrawHatchFallback(r, path, fallbackPaint) {
		r.Renderer.Path(path, &fallbackPaint)
	}
	if hatchPaint.Stroke.A > 0 && hatchPaint.LineWidth > 0 {
		strokePaint := hatchPaint
		strokePaint.Fill = render.Color{}
		strokePaint.FillPattern = render.PatternFill{}
		strokePaint.Hatch = ""
		r.Renderer.Path(path, &strokePaint)
	}
	return true
}

func hasPatternFill(pattern render.PatternFill) bool {
	return pattern.ID != "" || len(pattern.Path.V) > 0
}

func rasterizeMask(width, height int, path geom.Path) *image.Alpha {
	if width <= 0 || height <= 0 || len(path.C) == 0 || !path.Validate() {
		return nil
	}
	mask := image.NewAlpha(image.Rect(0, 0, width, height))
	ras := vector.NewRasterizer(width, height)
	appendPathToRasterizer(ras, path, 0, 0)
	ras.Draw(mask, mask.Bounds(), image.NewUniform(color.Alpha{A: 255}), image.Point{})
	return mask
}

func appendPathToRasterizer(ras *vector.Rasterizer, p geom.Path, offsetX, offsetY float64) {
	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			pt := p.V[vi]
			ras.MoveTo(rasterCoord(pt.X-offsetX), rasterCoord(pt.Y-offsetY))
			vi++
		case geom.LineTo:
			pt := p.V[vi]
			ras.LineTo(rasterCoord(pt.X-offsetX), rasterCoord(pt.Y-offsetY))
			vi++
		case geom.QuadTo:
			ctrl := p.V[vi]
			to := p.V[vi+1]
			ras.QuadTo(
				rasterCoord(ctrl.X-offsetX), rasterCoord(ctrl.Y-offsetY),
				rasterCoord(to.X-offsetX), rasterCoord(to.Y-offsetY),
			)
			vi += 2
		case geom.CubicTo:
			c1 := p.V[vi]
			c2 := p.V[vi+1]
			to := p.V[vi+2]
			ras.CubeTo(
				rasterCoord(c1.X-offsetX), rasterCoord(c1.Y-offsetY),
				rasterCoord(c2.X-offsetX), rasterCoord(c2.Y-offsetY),
				rasterCoord(to.X-offsetX), rasterCoord(to.Y-offsetY),
			)
			vi += 3
		case geom.ClosePath:
			ras.ClosePath()
		}
	}
}

func shaderDrawBounds(img image.Rectangle, path geom.Path, clipRect *geom.Rect) image.Rectangle {
	bounds, ok := pathPixelBounds(path)
	if !ok {
		return image.Rectangle{}
	}
	bounds = bounds.Intersect(img)
	if clipRect != nil {
		bounds = bounds.Intersect(image.Rect(
			int(math.Floor(clipRect.Min.X)),
			int(math.Floor(clipRect.Min.Y)),
			int(math.Ceil(clipRect.Max.X)),
			int(math.Ceil(clipRect.Max.Y)),
		))
	}
	return bounds
}

func pathPixelBounds(p geom.Path) (image.Rectangle, bool) {
	if len(p.V) == 0 {
		return image.Rectangle{}, false
	}
	minX, maxX := p.V[0].X, p.V[0].X
	minY, maxY := p.V[0].Y, p.V[0].Y
	for _, pt := range p.V[1:] {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}
	return image.Rect(
		int(math.Floor(minX))-1,
		int(math.Floor(minY))-1,
		int(math.Ceil(maxX))+1,
		int(math.Ceil(maxY))+1,
	), true
}

func pathBounds(p geom.Path) (geom.Rect, bool) {
	if len(p.V) == 0 {
		return geom.Rect{}, false
	}
	minX, maxX := p.V[0].X, p.V[0].X
	minY, maxY := p.V[0].Y, p.V[0].Y
	for _, pt := range p.V[1:] {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}
	return geom.Rect{Min: geom.Pt{X: minX, Y: minY}, Max: geom.Pt{X: maxX, Y: maxY}}, true
}

func pointInPath(p geom.Pt, path geom.Path) bool {
	poly := make([]geom.Pt, 0, len(path.V))
	vi := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo, geom.LineTo:
			poly = append(poly, path.V[vi])
			vi++
		case geom.QuadTo:
			vi += 2
		case geom.CubicTo:
			vi += 3
		case geom.ClosePath:
		}
	}
	if len(poly) < 3 {
		return false
	}
	inside := false
	j := len(poly) - 1
	for i := range poly {
		pi := poly[i]
		pj := poly[j]
		if (pi.Y > p.Y) != (pj.Y > p.Y) {
			x := (pj.X-pi.X)*(p.Y-pi.Y)/(pj.Y-pi.Y) + pi.X
			if p.X < x {
				inside = !inside
			}
		}
		j = i
	}
	return inside
}

func alphaAt(mask *image.Alpha, x, y int) uint8 {
	if mask == nil || !image.Pt(x, y).In(mask.Bounds()) {
		return 0
	}
	return mask.AlphaAt(x, y).A
}

func renderColorToRGBA(c render.Color) color.RGBA {
	return color.RGBA{
		R: uint8(clamp01(c.R)*255 + 0.5),
		G: uint8(clamp01(c.G)*255 + 0.5),
		B: uint8(clamp01(c.B)*255 + 0.5),
		A: uint8(clamp01(c.A)*255 + 0.5),
	}
}

func blendPixel(dst *image.RGBA, x, y int, src color.RGBA) {
	if dst == nil || src.A == 0 || !image.Pt(x, y).In(dst.Bounds()) {
		return
	}
	i := dst.PixOffset(x, y)
	dr := uint32(dst.Pix[i])
	dg := uint32(dst.Pix[i+1])
	db := uint32(dst.Pix[i+2])
	da := uint32(dst.Pix[i+3])
	sr := uint32(src.R)
	sg := uint32(src.G)
	sb := uint32(src.B)
	sa := uint32(src.A)
	if sa == 255 {
		dst.Pix[i] = src.R
		dst.Pix[i+1] = src.G
		dst.Pix[i+2] = src.B
		dst.Pix[i+3] = src.A
		return
	}
	outA := sa + ((255 - sa) * da / 255)
	if outA == 0 {
		dst.Pix[i] = 0
		dst.Pix[i+1] = 0
		dst.Pix[i+2] = 0
		dst.Pix[i+3] = 0
		return
	}
	dst.Pix[i] = uint8((sr*sa + dr*(255-sa)*da/255) / outA)
	dst.Pix[i+1] = uint8((sg*sa + dg*(255-sa)*da/255) / outA)
	dst.Pix[i+2] = uint8((sb*sa + db*(255-sa)*da/255) / outA)
	dst.Pix[i+3] = uint8(outA)
}

func rasterCoord(v float64) float32 {
	return float32(math.Round(v*1e6) / 1e6)
}

func clonePath(p geom.Path) geom.Path {
	out := geom.Path{
		C: make([]geom.Cmd, len(p.C)),
		V: make([]geom.Pt, len(p.V)),
	}
	copy(out.C, p.C)
	copy(out.V, p.V)
	return out
}

func clonePaths(paths []geom.Path) []geom.Path {
	if len(paths) == 0 {
		return nil
	}
	out := make([]geom.Path, len(paths))
	for i, path := range paths {
		out[i] = clonePath(path)
	}
	return out
}

func cloneRectPtr(rect *geom.Rect) *geom.Rect {
	if rect == nil {
		return nil
	}
	copied := *rect
	return &copied
}

// flipPathY returns a copy of p with every vertex flipped from y-up display
// space to the y-down device buffer (y -> height - y), mirroring
// gobasic.Renderer.devPath. The command slice is shared (read-only).
func flipPathY(p geom.Path, height float64) geom.Path {
	out := p
	out.V = make([]geom.Pt, len(p.V))
	for i, v := range p.V {
		out.V[i] = geom.Pt{X: v.X, Y: height - v.Y}
	}
	return out
}

// flipPathsY device-flips each path, returning nil for an empty input.
func flipPathsY(paths []geom.Path, height float64) []geom.Path {
	if len(paths) == 0 {
		return nil
	}
	out := make([]geom.Path, len(paths))
	for i, p := range paths {
		out[i] = flipPathY(p, height)
	}
	return out
}

// flipRectPtrY device-flips a display-space rect, swapping Min.Y/Max.Y so the
// result keeps Min.Y <= Max.Y (mirrors gobasic.Renderer.devRect).
func flipRectPtrY(rect *geom.Rect, height float64) *geom.Rect {
	if rect == nil {
		return nil
	}
	flipped := geom.Rect{
		Min: geom.Pt{X: rect.Min.X, Y: height - rect.Max.Y},
		Max: geom.Pt{X: rect.Max.X, Y: height - rect.Min.Y},
	}
	return &flipped
}

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
	for i := range paint.FillGradient.Stops {
		if paint.FillGradient.Stops[i].Color.A > 0 {
			paint.FillGradient.Stops[i].Color.A = alpha
		}
	}
	if paint.FillPattern.Foreground.A > 0 {
		paint.FillPattern.Foreground.A = alpha
	}
	if paint.FillPattern.Background.A > 0 {
		paint.FillPattern.Background.A = alpha
	}
}

func colorWithForcedAlpha(c render.Color, paint *render.Paint) render.Color {
	if paint != nil && paint.ForceAlpha && c.A > 0 {
		c.A = clamp01(paint.Alpha)
	}
	return c
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
