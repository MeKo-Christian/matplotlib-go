package pdf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"image"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/cwbudde/matplotlib-go/backends/internal/vectorhatch"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/diag"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/font/sfnt"
)

// gradientCollectionDropWarnOnce guards a single diagnostic for gradient- or
// pattern-filled path-collection / marker items that the Form XObject path
// cannot yet render and therefore skips.
var gradientCollectionDropWarnOnce sync.Once

// warnGradientCollectionDrop emits a one-shot diagnostic when a collection or
// marker item is skipped solely because its only fill is a gradient/pattern.
// This replaces a silent drop (the item simply vanished from the page).
func warnGradientCollectionDrop(paint *render.Paint) {
	if paint == nil {
		return
	}
	grad := paint.FillGradient.Kind != render.GradientNone && len(paint.FillGradient.Stops) > 0
	pat := paint.FillPattern.ID != "" || len(paint.FillPattern.Path.V) > 0
	if !grad && !pat {
		return
	}
	gradientCollectionDropWarnOnce.Do(func() {
		diag.Warnf("pdf: gradient/pattern-filled path-collection or marker items are not yet supported and were skipped")
	})
}

// writePathOps emits PDF path-construction operators for path p. Returns
// false if the path is empty or invalid.
func writePathOps(w *bytes.Buffer, p geom.Path) bool {
	if !p.Validate() || len(p.C) == 0 {
		return false
	}
	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			pt := p.V[vi]
			vi++
			fmt.Fprintf(w, "%s %s m\n", shortFloat(pt.X), shortFloat(pt.Y))
		case geom.LineTo:
			pt := p.V[vi]
			vi++
			fmt.Fprintf(w, "%s %s l\n", shortFloat(pt.X), shortFloat(pt.Y))
		case geom.QuadTo:
			// PDF has no quadratic curve operator; promote to cubic.
			// The previous endpoint must exist; if it does not, skip.
			if vi == 0 {
				vi += 2
				continue
			}
			prev := lastEndpoint(p, vi)
			ctrl := p.V[vi]
			end := p.V[vi+1]
			vi += 2
			c1 := geom.Pt{
				X: prev.X + (2.0/3.0)*(ctrl.X-prev.X),
				Y: prev.Y + (2.0/3.0)*(ctrl.Y-prev.Y),
			}
			c2 := geom.Pt{
				X: end.X + (2.0/3.0)*(ctrl.X-end.X),
				Y: end.Y + (2.0/3.0)*(ctrl.Y-end.Y),
			}
			fmt.Fprintf(
				w, "%s %s %s %s %s %s c\n",
				shortFloat(c1.X), shortFloat(c1.Y),
				shortFloat(c2.X), shortFloat(c2.Y),
				shortFloat(end.X), shortFloat(end.Y),
			)
		case geom.CubicTo:
			c1 := p.V[vi]
			c2 := p.V[vi+1]
			end := p.V[vi+2]
			vi += 3
			fmt.Fprintf(
				w, "%s %s %s %s %s %s c\n",
				shortFloat(c1.X), shortFloat(c1.Y),
				shortFloat(c2.X), shortFloat(c2.Y),
				shortFloat(end.X), shortFloat(end.Y),
			)
		case geom.ClosePath:
			w.WriteString("h\n")
		}
	}
	return true
}

// lastEndpoint returns the endpoint emitted by the command immediately before
// vi. Used to promote quadratic curves to cubics.
func lastEndpoint(p geom.Path, vi int) geom.Pt {
	// Walk commands counting vertices to find the verb that ends at vi-1.
	consumed := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo, geom.LineTo:
			consumed++
			if consumed == vi {
				return p.V[consumed-1]
			}
		case geom.QuadTo:
			consumed += 2
			if consumed == vi {
				return p.V[consumed-1]
			}
		case geom.CubicTo:
			consumed += 3
			if consumed == vi {
				return p.V[consumed-1]
			}
		case geom.ClosePath:
			// No vertices.
		}
	}
	if vi > 0 && vi-1 < len(p.V) {
		return p.V[vi-1]
	}
	return geom.Pt{}
}

func writeFillColor(w *bytes.Buffer, c render.Color) {
	fmt.Fprintf(
		w, "%s %s %s rg\n",
		shortFloat(clamp01(c.R)),
		shortFloat(clamp01(c.G)),
		shortFloat(clamp01(c.B)),
	)
}

func writeStrokeColor(w *bytes.Buffer, c render.Color) {
	fmt.Fprintf(
		w, "%s %s %s RG\n",
		shortFloat(clamp01(c.R)),
		shortFloat(clamp01(c.G)),
		shortFloat(clamp01(c.B)),
	)
}

func writePatternFill(w *bytes.Buffer, name string) {
	fmt.Fprintf(w, "/Pattern cs\n/%s scn\n", escapeName(name))
}

func (r *Renderer) writePaintState(paint *render.Paint) {
	if paint == nil {
		return
	}
	r.writeAlphaState(paint)
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	hasGradient := !hasHatch && paint.FillGradient.Kind != render.GradientNone && len(paint.FillGradient.Stops) > 0
	hasPattern := !hasHatch && !hasGradient && (paint.FillPattern.ID != "" || len(paint.FillPattern.Path.V) > 0)
	if hasHatch {
		writePatternFill(&r.content, r.registerHatchPattern(*paint))
	} else if hasPattern {
		writePatternFill(&r.content, r.registerFillPattern(paint.FillPattern))
	} else if hasGradient {
		// Form XObjects cannot use the page path as a shading clip, so
		// gradient-painted batches fall back to the main Path path before
		// reaching writePaintState.
	} else if paint.Fill.A > 0 {
		writeFillColor(&r.content, paint.Fill)
	}
	if paint.Stroke.A > 0 && paint.LineWidth > 0 {
		writeStrokeColor(&r.content, paint.Stroke)
		writeLineState(&r.content, paint)
	}
}

func writeLineState(w *bytes.Buffer, paint *render.Paint) {
	if paint.LineWidth > 0 {
		fmt.Fprintf(w, "%s w\n", shortFloat(paint.LineWidth))
	}
	switch paint.LineCap {
	case render.CapButt:
		w.WriteString("0 J\n")
	case render.CapRound:
		w.WriteString("1 J\n")
	case render.CapSquare:
		w.WriteString("2 J\n")
	}
	switch paint.LineJoin {
	case render.JoinMiter:
		w.WriteString("0 j\n")
	case render.JoinRound:
		w.WriteString("1 j\n")
	case render.JoinBevel:
		w.WriteString("2 j\n")
	}
	if paint.MiterLimit > 0 {
		fmt.Fprintf(w, "%s M\n", shortFloat(paint.MiterLimit))
	}
	if len(paint.Dashes) > 0 {
		w.WriteString("[")
		for i, d := range paint.Dashes {
			if i > 0 {
				w.WriteString(" ")
			}
			w.WriteString(shortFloat(d))
		}
		fmt.Fprintf(w, "] %s d\n", shortFloat(paint.DashOffset))
	} else {
		w.WriteString("[] 0 d\n")
	}
}

// shortFloat mirrors the SVG backend's compact float formatter: up to six
// decimals, trailing zeros stripped, -0 normalized, and NaN/Inf clamped to 0.
func shortFloat(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "0"
	}
	if v == 0 {
		return "0"
	}
	s := strconv.FormatFloat(v, 'f', 6, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	if s == "" || s == "-0" {
		return "0"
	}
	return s
}

func pdfCIDHexString(cids []uint16) string {
	var b strings.Builder
	b.WriteByte('<')
	for _, cid := range cids {
		fmt.Fprintf(&b, "%04X", cid)
	}
	b.WriteByte('>')
	return b.String()
}

func pdfFontFaceKey(face render.FontFace) string {
	if face.Path != "" {
		return "path:" + face.Path
	}
	if face.Family != "" {
		return "embedded:" + face.Family
	}
	return ""
}

func pdfFontFaceData(face render.FontFace) ([]byte, bool) {
	if len(face.Data) > 0 {
		return append([]byte(nil), face.Data...), true
	}
	if face.Path == "" {
		return nil, false
	}
	data, err := os.ReadFile(face.Path)
	if err != nil {
		return nil, false
	}
	return data, true
}

func pdfFontPostScriptName(fontData *sfnt.Font, face render.FontFace) string {
	if fontData != nil {
		if name, err := fontData.Name(nil, sfnt.NameIDPostScript); err == nil {
			if cleaned := cleanPDFFontName(name); cleaned != "" {
				return cleaned
			}
		}
	}
	if cleaned := cleanPDFFontName(face.Family); cleaned != "" {
		return cleaned
	}
	return "MatplotlibGoFont"
}

func cleanPDFFontName(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '+':
			b.WriteRune(r)
		}
	}
	return b.String()
}

func subsetFontName(font pdfEmbeddedFont) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(font.baseName))
	cids := sortedFontCIDs(font.gidByCID)
	for _, cid := range cids {
		var buf [6]byte
		binary.BigEndian.PutUint16(buf[0:2], cid)
		binary.BigEndian.PutUint32(buf[2:6], uint32(font.gidByCID[cid]))
		_, _ = h.Write(buf[:])
	}
	value := h.Sum64()
	prefix := make([]byte, 6)
	for i := range prefix {
		prefix[i] = byte('A' + value%26)
		value /= 26
	}
	return string(prefix) + "+" + font.baseName
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

func rotationAffine(angle float64, pivot geom.Pt) geom.Affine {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return translateAffine(pivot).
		Mul(geom.Affine{A: cos, B: sin, C: -sin, D: cos}).
		Mul(translateAffine(geom.Pt{X: -pivot.X, Y: -pivot.Y}))
}

func translateAffine(p geom.Pt) geom.Affine {
	return geom.Affine{A: 1, D: 1, E: p.X, F: p.Y}
}

func affinePath(path geom.Path, affine geom.Affine) geom.Path {
	if len(path.V) == 0 {
		return path
	}
	out := geom.Path{
		V: make([]geom.Pt, len(path.V)),
		C: append([]geom.Cmd(nil), path.C...),
	}
	for i, pt := range path.V {
		out.V[i] = affine.Apply(pt)
	}
	return out
}

func clonePath(path geom.Path) geom.Path {
	return geom.Path{
		V: append([]geom.Pt(nil), path.V...),
		C: append([]geom.Cmd(nil), path.C...),
	}
}

func cloneRectPtr(rect *geom.Rect) *geom.Rect {
	if rect == nil {
		return nil
	}
	cloned := *rect
	return &cloned
}

func normalizeRect(rect geom.Rect) geom.Rect {
	minX, maxX := rect.Min.X, rect.Max.X
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	minY, maxY := rect.Min.Y, rect.Max.Y
	if minY > maxY {
		minY, maxY = maxY, minY
	}
	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}
}

func pathBounds(path geom.Path) (geom.Rect, bool) {
	if len(path.V) == 0 {
		return geom.Rect{}, false
	}
	minX, maxX := path.V[0].X, path.V[0].X
	minY, maxY := path.V[0].Y, path.V[0].Y
	for _, pt := range path.V[1:] {
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

func paintOperator(paint *render.Paint) string {
	if paint == nil {
		return ""
	}
	hasFill := paint.Fill.A > 0 ||
		(paint.Hatch != "" && paint.HatchColor.A > 0) ||
		(paint.FillPattern.ID != "" || len(paint.FillPattern.Path.V) > 0)
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	switch {
	case hasFill && hasStroke:
		return "B"
	case hasFill:
		return "f"
	case hasStroke:
		return "S"
	default:
		return ""
	}
}

func formPadding(paint *render.Paint) float64 {
	if paint == nil || paint.Stroke.A <= 0 || paint.LineWidth <= 0 {
		return 0
	}
	padding := paint.LineWidth / 2
	if paint.LineJoin == render.JoinMiter {
		miter := paint.MiterLimit
		if miter <= 0 {
			miter = 10
		}
		padding = math.Max(padding, paint.LineWidth*miter/2)
	}
	return padding
}

func formXObjectKey(prefix string, path geom.Path, paintOp string, paint *render.Paint) string {
	var b strings.Builder
	b.WriteString(prefix)
	b.WriteByte('\x00')
	b.WriteString(paintOp)
	b.WriteByte('\x00')
	if paint != nil {
		b.WriteString(strconv.Itoa(int(paint.LineJoin)))
		b.WriteByte('\x00')
		b.WriteString(strconv.Itoa(int(paint.LineCap)))
	}
	b.WriteString(pathKey(path))
	return b.String()
}

func pathEffectFormKey(content []byte, bbox geom.Rect) string {
	return strings.Join([]string{
		"E",
		string(content),
		shortFloat(bbox.Min.X),
		shortFloat(bbox.Min.Y),
		shortFloat(bbox.Max.X),
		shortFloat(bbox.Max.Y),
	}, "\x00")
}

func isPDFIdentityPathEffectFilter(effect render.PathEffect) bool {
	name := strings.ToLower(strings.TrimSpace(effect.Filter))
	return name == "" || name == "none" || name == "identity"
}

func isPDFBlurPathEffectFilter(effect render.PathEffect) bool {
	name := strings.ToLower(strings.TrimSpace(effect.Filter))
	switch name {
	case "blur", "gaussian", "gaussian-blur", "shadow":
		return true
	default:
		return false
	}
}

func pathKey(path geom.Path) string {
	var b strings.Builder
	for _, cmd := range path.C {
		b.WriteByte(byte(cmd))
	}
	b.WriteByte('\x00')
	for _, pt := range path.V {
		b.WriteString(shortFloat(pt.X))
		b.WriteByte(',')
		b.WriteString(shortFloat(pt.Y))
		b.WriteByte(';')
	}
	return b.String()
}

func imageAlphaMultiplier(img render.Image) float64 {
	if alphaImage, ok := img.(render.ImageAlpha); ok {
		return clamp01(alphaImage.Alpha())
	}
	return 1
}

func encodePDFImage(name string, src *image.RGBA, alphaMul float64) (pdfImage, bool) {
	if src == nil {
		return pdfImage{}, false
	}
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return pdfImage{}, false
	}
	rgb := make([]byte, 0, width*height*3)
	alpha := make([]byte, 0, width*height)
	hasAlpha := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.RGBAAt(x, y)
			a := uint8(float64(c.A)*alphaMul + 0.5)
			rgb = append(rgb, c.R, c.G, c.B)
			alpha = append(alpha, a)
			if a != 0xff {
				hasAlpha = true
			}
		}
	}
	return pdfImage{
		name:     name,
		width:    width,
		height:   height,
		colors:   3,
		rgb:      rgb,
		alpha:    alpha,
		hasAlpha: hasAlpha,
		filter:   "FlateDecode",
	}, true
}

func encodePDFJPEGImage(name string, src render.JPEGImage) (pdfImage, bool) {
	if src == nil {
		return pdfImage{}, false
	}
	width, height := src.Size()
	data := src.JPEGData()
	if width <= 0 || height <= 0 || len(data) == 0 {
		return pdfImage{}, false
	}
	return pdfImage{
		name:   name,
		width:  width,
		height: height,
		colors: 3,
		rgb:    append([]byte(nil), data...),
		filter: "DCTDecode",
	}, true
}

func imageKey(img pdfImage) string {
	var b strings.Builder
	b.WriteString(strconv.Itoa(img.width))
	b.WriteByte('x')
	b.WriteString(strconv.Itoa(img.height))
	b.WriteByte('\x00')
	if img.hasAlpha {
		b.WriteByte('a')
	} else {
		b.WriteByte('o')
	}
	b.WriteByte('\x00')
	b.WriteString(img.filter)
	b.WriteByte('\x00')
	b.WriteString(string(img.rgb))
	b.WriteByte('\x00')
	if img.hasAlpha {
		b.WriteString(string(img.alpha))
	}
	return b.String()
}

func imageColorCount(img pdfImage) int {
	if img.colors > 0 {
		return img.colors
	}
	return 3
}

func pngPredictorRows(data []byte, width, height, colors int) []byte {
	if width <= 0 || height <= 0 || colors <= 0 {
		return data
	}
	rowLen := width * colors
	if rowLen <= 0 || len(data) != rowLen*height {
		return data
	}
	out := make([]byte, 0, len(data)+height)
	for y := 0; y < height; y++ {
		out = append(out, 0) // PNG filter type 0 (None).
		start := y * rowLen
		out = append(out, data[start:start+rowLen]...)
	}
	return out
}

func hatchPatternKey(hatch string, face, line render.Color, lineWidth, spacing float64) string {
	return strings.Join([]string{
		hatch,
		shortFloat(face.R),
		shortFloat(face.G),
		shortFloat(face.B),
		shortFloat(face.A),
		shortFloat(line.R),
		shortFloat(line.G),
		shortFloat(line.B),
		shortFloat(line.A),
		shortFloat(lineWidth),
		shortFloat(spacing),
	}, "\x00")
}

func hatchPatternStream(pattern pdfHatchPattern) []byte {
	const side = 72.0
	var buf bytes.Buffer
	if pattern.faceColor.A > 0 {
		writeFillColor(&buf, pattern.faceColor)
		fmt.Fprintf(&buf, "0 0 %s %s re f\n", shortFloat(side), shortFloat(side))
	}
	writeStrokeColor(&buf, pattern.lineColor)
	fmt.Fprintf(&buf, "%s w\n", shortFloat(pattern.lineWidth))
	buf.WriteString("0 J\n")
	for _, line := range hatchPatternLines(pattern.hatch, pattern.spacing) {
		fmt.Fprintf(
			&buf, "%s %s m\n%s %s l\nS\n",
			shortFloat(line[0].X), shortFloat(line[0].Y),
			shortFloat(line[1].X), shortFloat(line[1].Y),
		)
	}
	for _, shape := range vectorhatch.ShapePaths(pattern.hatch, pattern.spacing) {
		if !writePathOps(&buf, shape.Path) {
			continue
		}
		if shape.Filled {
			writeFillColor(&buf, pattern.lineColor)
			buf.WriteString("f\n")
			writeStrokeColor(&buf, pattern.lineColor)
		} else {
			buf.WriteString("S\n")
		}
	}
	return buf.Bytes()
}

func fillPatternKey(pattern render.PatternFill) string {
	parts := []string{
		pattern.ID,
		shortFloat(pattern.Cell.Min.X),
		shortFloat(pattern.Cell.Min.Y),
		shortFloat(pattern.Cell.Max.X),
		shortFloat(pattern.Cell.Max.Y),
		shortFloat(pattern.Foreground.R),
		shortFloat(pattern.Foreground.G),
		shortFloat(pattern.Foreground.B),
		shortFloat(pattern.Foreground.A),
		shortFloat(pattern.Background.R),
		shortFloat(pattern.Background.G),
		shortFloat(pattern.Background.B),
		shortFloat(pattern.Background.A),
		shortFloat(pattern.LineWidth),
		strconv.FormatBool(pattern.HasTransform),
		pathKey(pattern.Path),
	}
	if pattern.HasTransform {
		parts = append(
			parts,
			shortFloat(pattern.Transform.A),
			shortFloat(pattern.Transform.B),
			shortFloat(pattern.Transform.C),
			shortFloat(pattern.Transform.D),
			shortFloat(pattern.Transform.E),
			shortFloat(pattern.Transform.F),
		)
	}
	return strings.Join(parts, "\x00")
}

func fillPatternStream(pattern render.PatternFill) []byte {
	cell := normalizedPatternCell(pattern.Cell)
	var buf bytes.Buffer
	if pattern.Background.A > 0 {
		writeFillColor(&buf, pattern.Background)
		fmt.Fprintf(
			&buf, "%s %s %s %s re f\n",
			shortFloat(cell.Min.X),
			shortFloat(cell.Min.Y),
			shortFloat(cell.W()),
			shortFloat(cell.H()),
		)
	}
	if len(pattern.Path.C) > 0 && pattern.Foreground.A > 0 && writePathOps(&buf, pattern.Path) {
		writeFillColor(&buf, pattern.Foreground)
		if pattern.LineWidth > 0 {
			writeStrokeColor(&buf, pattern.Foreground)
			fmt.Fprintf(&buf, "%s w\n", shortFloat(pattern.LineWidth))
			buf.WriteString("B\n")
		} else {
			buf.WriteString("f\n")
		}
	}
	return buf.Bytes()
}

func normalizedPatternCell(cell geom.Rect) geom.Rect {
	if cell.W() > 0 && cell.H() > 0 {
		return cell
	}
	return geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 16, Y: 16}}
}

func hatchPatternLines(hatch string, spacing float64) [][2]geom.Pt {
	if spacing <= 0 {
		spacing = 8
	}
	lines := make([][2]geom.Pt, 0)
	writeHatchLines := func(count int, draw func(float64)) {
		if count <= 0 {
			return
		}
		step := math.Max(2, spacing/float64(count))
		for v := -72.0; v <= 144; v += step {
			draw(v)
		}
	}
	add := func(x1, y1, x2, y2 float64) {
		lines = append(lines, [2]geom.Pt{{X: x1, Y: y1}, {X: x2, Y: y2}})
	}
	verticalCount := strings.Count(hatch, "|") + strings.Count(hatch, "+")
	horizontalCount := strings.Count(hatch, "-") + strings.Count(hatch, "+")
	slashCount := strings.Count(hatch, "/") + strings.Count(hatch, "x") + strings.Count(hatch, "X")
	backslashCount := strings.Count(hatch, `\`) + strings.Count(hatch, "x") + strings.Count(hatch, "X")

	writeHatchLines(verticalCount, func(x float64) { add(x, 0, x, 72) })
	writeHatchLines(horizontalCount, func(y float64) { add(0, y, 72, y) })
	writeHatchLines(slashCount, func(x float64) { add(x, 72, x+72, 0) })
	writeHatchLines(backslashCount, func(x float64) { add(x, 0, x+72, 72) })
	return lines
}

func shadingDictionary(gradient render.GradientFill) string {
	gradient.Stops = normalizeGradientStops(gradient.Stops)
	switch gradient.Kind {
	case render.LinearGradient:
		start := transformedGradientPoint(gradient.Start, gradient)
		end := transformedGradientPoint(gradient.End, gradient)
		return fmt.Sprintf(
			"<< /ShadingType 2 /ColorSpace /DeviceRGB /Coords [%s %s %s %s] /Domain [0 1] /Function %s /Extend [true true] >>",
			shortFloat(start.X), shortFloat(start.Y),
			shortFloat(end.X), shortFloat(end.Y),
			gradientFunctionDictionary(gradient.Stops),
		)
	case render.RadialGradient:
		center := transformedGradientPoint(gradient.Center, gradient)
		focal := center
		if gradient.Focal != (geom.Pt{}) {
			focal = transformedGradientPoint(gradient.Focal, gradient)
		}
		radius := transformedGradientRadius(gradient.Radius, gradient)
		return fmt.Sprintf(
			"<< /ShadingType 3 /ColorSpace /DeviceRGB /Coords [%s %s 0 %s %s %s] /Domain [0 1] /Function %s /Extend [true true] >>",
			shortFloat(focal.X), shortFloat(focal.Y),
			shortFloat(center.X), shortFloat(center.Y), shortFloat(radius),
			gradientFunctionDictionary(gradient.Stops),
		)
	default:
		stops := normalizeGradientStops(gradient.Stops)
		return fmt.Sprintf("<< /ShadingType 2 /ColorSpace /DeviceRGB /Coords [0 0 1 0] /Domain [0 1] /Function %s /Extend [true true] >>",
			gradientFunctionDictionary(stops))
	}
}

func transformedGradientPoint(p geom.Pt, gradient render.GradientFill) geom.Pt {
	if !gradient.HasTransform {
		return p
	}
	return gradient.Transform.Apply(p)
}

func transformedGradientRadius(radius float64, gradient render.GradientFill) float64 {
	if radius <= 0 {
		return 1
	}
	if !gradient.HasTransform {
		return radius
	}
	xScale := math.Hypot(gradient.Transform.A, gradient.Transform.B)
	yScale := math.Hypot(gradient.Transform.C, gradient.Transform.D)
	scale := (xScale + yScale) / 2
	if scale <= 0 {
		return radius
	}
	return radius * scale
}

func gradientFunctionDictionary(stops []render.GradientStop) string {
	stops = normalizeGradientStops(stops)
	if len(stops) == 0 {
		stops = []render.GradientStop{
			{Offset: 0, Color: render.Color{A: 1}},
			{Offset: 1, Color: render.Color{A: 1}},
		}
	}
	if len(stops) == 1 {
		stops = []render.GradientStop{
			{Offset: 0, Color: stops[0].Color},
			{Offset: 1, Color: stops[0].Color},
		}
	}
	if len(stops) == 2 {
		return type2FunctionDictionary(stops[0].Color, stops[1].Color)
	}

	var functions strings.Builder
	var bounds strings.Builder
	var encode strings.Builder
	for i := 0; i < len(stops)-1; i++ {
		if i > 0 {
			functions.WriteByte(' ')
			encode.WriteByte(' ')
		}
		functions.WriteString(type2FunctionDictionary(stops[i].Color, stops[i+1].Color))
		encode.WriteString("0 1")
	}
	for i := 1; i < len(stops)-1; i++ {
		if i > 1 {
			bounds.WriteByte(' ')
		}
		bounds.WriteString(shortFloat(stops[i].Offset))
	}
	return fmt.Sprintf("<< /FunctionType 3 /Domain [0 1] /Functions [%s] /Bounds [%s] /Encode [%s] >>",
		functions.String(), bounds.String(), encode.String())
}

func type2FunctionDictionary(c0, c1 render.Color) string {
	return fmt.Sprintf("<< /FunctionType 2 /Domain [0 1] /C0 %s /C1 %s /N 1 >>",
		pdfColorArray(c0), pdfColorArray(c1))
}

func pdfColorArray(c render.Color) string {
	return fmt.Sprintf(
		"[%s %s %s]",
		shortFloat(clamp01(c.R)),
		shortFloat(clamp01(c.G)),
		shortFloat(clamp01(c.B)),
	)
}

func normalizeGradientStops(in []render.GradientStop) []render.GradientStop {
	if len(in) == 0 {
		return nil
	}
	out := append([]render.GradientStop(nil), in...)
	for i := range out {
		out[i].Offset = clamp01(out[i].Offset)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Offset < out[j].Offset
	})
	if len(out) == 1 {
		return out
	}
	for i := 1; i < len(out); i++ {
		if out[i].Offset <= out[i-1].Offset {
			out[i].Offset = math.Nextafter(out[i-1].Offset, 1)
		}
	}
	return out
}

func gradientAlpha(stops []render.GradientStop) float64 {
	if len(stops) == 0 {
		return 1
	}
	alpha := clamp01(stops[0].Color.A)
	for _, stop := range stops[1:] {
		if a := clamp01(stop.Color.A); a > alpha {
			alpha = a
		}
	}
	return alpha
}

func patternAlpha(pattern render.PatternFill) float64 {
	alpha := clamp01(pattern.Foreground.A)
	if a := clamp01(pattern.Background.A); a > alpha {
		alpha = a
	}
	if alpha <= 0 {
		return 1
	}
	return alpha
}

func shadingKey(gradient render.GradientFill) string {
	stops := normalizeGradientStops(gradient.Stops)
	parts := []string{
		strconv.Itoa(int(gradient.Kind)),
		shortFloat(gradient.Start.X), shortFloat(gradient.Start.Y),
		shortFloat(gradient.End.X), shortFloat(gradient.End.Y),
		shortFloat(gradient.Center.X), shortFloat(gradient.Center.Y),
		shortFloat(gradient.Focal.X), shortFloat(gradient.Focal.Y),
		shortFloat(gradient.Radius),
		strconv.FormatBool(gradient.HasTransform),
	}
	if gradient.HasTransform {
		parts = append(
			parts,
			shortFloat(gradient.Transform.A), shortFloat(gradient.Transform.B),
			shortFloat(gradient.Transform.C), shortFloat(gradient.Transform.D),
			shortFloat(gradient.Transform.E), shortFloat(gradient.Transform.F),
		)
	}
	for _, stop := range stops {
		parts = append(
			parts,
			shortFloat(stop.Offset),
			shortFloat(stop.Color.R), shortFloat(stop.Color.G),
			shortFloat(stop.Color.B), shortFloat(stop.Color.A),
		)
	}
	return strings.Join(parts, "\x00")
}

// --- PDF document assembly ---------------------------------------------------
