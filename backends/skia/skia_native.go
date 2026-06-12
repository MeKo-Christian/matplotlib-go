//go:build skia

package skia

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

var (
	_ render.ImageTransformer      = (*Renderer)(nil)
	_ render.MarkerDrawer          = (*Renderer)(nil)
	_ render.PathCollectionDrawer  = (*Renderer)(nil)
	_ render.QuadMeshDrawer        = (*Renderer)(nil)
	_ render.GouraudTriangleDrawer = (*Renderer)(nil)
	_ render.NativeHatcher         = (*Renderer)(nil)
)

// nativeBatchBridge is implemented by a surfaceBridge backed by a real Skia
// library (the cgo bridge under the skiacgo build tag). When the active bridge
// satisfies it, the batch entry points route markers and Gouraud triangles
// through native Skia rasterization; otherwise they use the per-item CPU loops
// below. Each method returns false when it declines a batch, so the caller can
// fall back without losing output.
type nativeBatchBridge interface {
	drawMarkersNative(dst *image.RGBA, batch render.MarkerBatch, state bridgeDrawState, height float64) bool
	drawPathCollectionNative(dst *image.RGBA, batch render.PathCollectionBatch, state bridgeDrawState) bool
	drawQuadMeshNative(dst *image.RGBA, batch render.QuadMeshBatch, state bridgeDrawState) bool
	drawGouraudNative(dst *image.RGBA, batch render.GouraudTriangleBatch, state bridgeDrawState) bool
}

// bridgeClipState snapshots the current clip into the bridge-facing struct the
// native batch methods consume.
func (r *Renderer) bridgeClipState() bridgeDrawState {
	return bridgeDrawState{clipRect: r.clipRect, clipPaths: r.clipPaths}
}

// DrawMarkers renders one marker path at many display-space offsets. Native
// Skia builds batch the markers through one Skia surface; pure-Go Skia builds
// loop per item through Path.
func (r *Renderer) DrawMarkers(batch render.MarkerBatch) bool {
	if r == nil || len(batch.Marker.C) == 0 || len(batch.Items) == 0 {
		return false
	}
	if nb, ok := r.bridge.(nativeBatchBridge); ok {
		if nb.drawMarkersNative(r.GetImage(), batch, r.bridgeClipState(), float64(r.height)) {
			return true
		}
	}
	for i := range batch.Items {
		item := &batch.Items[i]
		path := transformMarkerPath(batch.Marker, item.Transform, item.Offset)
		if len(path.C) == 0 {
			continue
		}
		paint := item.Paint
		if !item.Antialiased {
			paint.Antialias = render.AntialiasOff
		}
		r.Path(path, &paint)
	}
	return true
}

// DrawPathCollection renders a display-space path collection. Native Skia
// builds draw supported solid fill/stroke items through one Skia surface;
// unsupported paint features and pure-Go Skia builds loop per item through Path.
func (r *Renderer) DrawPathCollection(batch render.PathCollectionBatch) bool {
	if r == nil || len(batch.Items) == 0 {
		return false
	}
	if nb, ok := r.bridge.(nativeBatchBridge); ok {
		if nb.drawPathCollectionNative(r.GetImage(), batch, r.bridgeClipState()) {
			return true
		}
	}
	for i := range batch.Items {
		item := &batch.Items[i]
		if len(item.Path.C) == 0 {
			continue
		}
		paint := item.Paint
		if item.Hatch != "" {
			paint.Hatch = item.Hatch
			paint.HatchColor = item.HatchColor
			paint.HatchLineWidth = item.HatchWidth
			paint.HatchSpacing = item.HatchSpacing
		}
		if !item.Antialiased {
			paint.Antialias = render.AntialiasOff
		}
		r.Path(item.Path, &paint)
	}
	return true
}

// DrawQuadMesh renders pcolor/pcolormesh-style quadrilateral cells. Native Skia
// builds draw face-only cells through SkVertices; styled cells and pure-Go Skia
// builds construct one Path per cell through the CPU bridge.
func (r *Renderer) DrawQuadMesh(batch render.QuadMeshBatch) bool {
	if r == nil || len(batch.Cells) == 0 {
		return false
	}
	if nb, ok := r.bridge.(nativeBatchBridge); ok {
		if nb.drawQuadMeshNative(r.GetImage(), batch, r.bridgeClipState()) {
			return true
		}
	}
	for i := range batch.Cells {
		cell := &batch.Cells[i]
		path := geom.Path{}
		path.MoveTo(cell.Quad[0])
		path.LineTo(cell.Quad[1])
		path.LineTo(cell.Quad[2])
		path.LineTo(cell.Quad[3])
		path.Close()
		paint := render.Paint{
			Fill:         cell.Face,
			Stroke:       cell.Edge,
			LineWidth:    cell.LineWidth,
			LineJoin:     render.JoinMiter,
			LineCap:      render.CapButt,
			Dashes:       append([]float64(nil), cell.Dashes...),
			Hatch:        cell.Hatch,
			HatchColor:   cell.HatchColor,
			HatchSpacing: cell.HatchSpacing,
			Antialias:    render.AntialiasDefault,
			Snap:         render.SnapOn,
		}
		if cell.HatchWidth > 0 {
			paint.HatchLineWidth = cell.HatchWidth
		}
		if !cell.Antialiased {
			paint.Antialias = render.AntialiasOff
		}
		if paint.LineWidth <= 0 || paint.Stroke.A <= 0 {
			paint.Stroke = render.Color{}
			paint.LineWidth = 0
		}
		if paint.Fill.A <= 0 {
			paint.Fill = render.Color{}
		}
		r.Path(path, &paint)
	}
	return true
}

// DrawGouraudTriangles renders interpolated-color triangles into the CPU Skia
// output surface. This mirrors the renderer contract that future SkVertices
// integration will satisfy through the external Skia bridge.
func (r *Renderer) DrawGouraudTriangles(batch render.GouraudTriangleBatch) bool {
	if r == nil || len(batch.Triangles) == 0 || r.GetImage() == nil {
		return false
	}
	if nb, ok := r.bridge.(nativeBatchBridge); ok {
		if nb.drawGouraudNative(r.GetImage(), batch, r.bridgeClipState()) {
			return true
		}
	}
	for i := range batch.Triangles {
		r.drawGouraudTriangle(&batch.Triangles[i])
	}
	return true
}

// SupportsNativeHatch reports that Skia consumes hatch metadata during Path.
// The actual hatch geometry is produced by the renderer-neutral
// render.DrawHatchFallback helper, so IsCapabilityBridged reports
// NativeHatcher as bridged until the external Skia C ABI provides tiled
// SkShader hatches.
func (r *Renderer) SupportsNativeHatch() bool { return r != nil }

func (r *Renderer) drawNativeHatchPath(path geom.Path, paint *render.Paint) bool {
	if r == nil || r.Renderer == nil || paint == nil || paint.Hatch == "" {
		return false
	}
	hatchPaint := *paint
	applyForcedAlpha(&hatchPaint)

	if hatchPaint.Fill.A > 0 {
		fillPaint := hatchPaint
		fillPaint.Stroke = render.Color{}
		fillPaint.LineWidth = 0
		fillPaint.Hatch = ""
		fillPaint.HatchColor = render.Color{}
		fillPaint.FillGradient = render.GradientFill{}
		fillPaint.FillPattern = render.PatternFill{}
		r.Renderer.Path(path, &fillPaint)
	}

	fallbackPaint := hatchPaint
	fallbackPaint.Fill = render.Color{}
	fallbackPaint.Stroke = render.Color{}
	fallbackPaint.FillGradient = render.GradientFill{}
	fallbackPaint.FillPattern = render.PatternFill{}
	if !render.DrawHatchFallback(r, path, fallbackPaint) {
		r.Renderer.Path(path, &fallbackPaint)
	}

	if hatchPaint.Stroke.A > 0 && hatchPaint.LineWidth > 0 {
		strokePaint := hatchPaint
		strokePaint.Fill = render.Color{}
		strokePaint.Hatch = ""
		strokePaint.FillGradient = render.GradientFill{}
		strokePaint.FillPattern = render.PatternFill{}
		r.Renderer.Path(path, &strokePaint)
	}
	return true
}

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
	img := r.GetImage()
	if img == nil {
		return
	}
	bounds := gouraudTrianglePixelBounds(img.Bounds(), tri, r.clipRect)
	if bounds.Empty() {
		return
	}
	area := edgeFunction(tri.P[0], tri.P[1], tri.P[2])
	if area == 0 || math.IsNaN(area) || math.IsInf(area, 0) {
		return
	}

	var clipMasks []*image.Alpha
	for _, clip := range r.clipPaths {
		mask := rasterizeMask(img.Bounds().Dx(), img.Bounds().Dy(), clip)
		if mask == nil {
			return
		}
		clipMasks = append(clipMasks, mask)
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			p := geom.Pt{X: float64(x) + 0.5, Y: float64(y) + 0.5}
			w0 := edgeFunction(tri.P[1], tri.P[2], p) / area
			w1 := edgeFunction(tri.P[2], tri.P[0], p) / area
			w2 := edgeFunction(tri.P[0], tri.P[1], p) / area
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			src := renderColorToRGBA(interpolateColor(tri.Color[0], tri.Color[1], tri.Color[2], w0, w1, w2))
			if src.A == 0 {
				continue
			}
			for _, mask := range clipMasks {
				src.A = uint8(uint32(src.A) * uint32(alphaAt(mask, x, y)) / 255)
				if src.A == 0 {
					break
				}
			}
			blendPixel(img, x, y, src)
		}
	}
}

func gouraudTrianglePixelBounds(img image.Rectangle, tri *render.GouraudTriangle, clipRect *geom.Rect) image.Rectangle {
	minX := int(math.Floor(math.Min(tri.P[0].X, math.Min(tri.P[1].X, tri.P[2].X))))
	maxX := int(math.Ceil(math.Max(tri.P[0].X, math.Max(tri.P[1].X, tri.P[2].X))))
	minY := int(math.Floor(math.Min(tri.P[0].Y, math.Min(tri.P[1].Y, tri.P[2].Y))))
	maxY := int(math.Ceil(math.Max(tri.P[0].Y, math.Max(tri.P[1].Y, tri.P[2].Y))))
	bounds := image.Rect(minX, minY, maxX+1, maxY+1).Intersect(img)
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
