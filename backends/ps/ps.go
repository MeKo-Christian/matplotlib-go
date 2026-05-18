package ps

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const defaultFontHeight = 13.0

type state struct {
	inContent bool
}

// Renderer implements render.Renderer by emitting deterministic Level-2
// PostScript content.
type Renderer struct {
	width      int
	height     int
	background render.Color
	resolution uint

	began    bool
	viewport geom.Rect
	content  strings.Builder
	document []byte
	stack    []state
}

var (
	_ render.Renderer               = (*Renderer)(nil)
	_ render.PSExporter             = (*Renderer)(nil)
	_ render.DPIAware               = (*Renderer)(nil)
	_ render.ImageTransformer       = (*Renderer)(nil)
	_ render.NativeHatcher          = (*Renderer)(nil)
	_ render.FontTextDrawer         = (*Renderer)(nil)
	_ render.FontRotatedTextDrawer  = (*Renderer)(nil)
	_ render.FontVerticalTextDrawer = (*Renderer)(nil)
)

// New creates a PostScript renderer with a point-sized page.
func New(width, height int, background render.Color) (*Renderer, error) {
	if width <= 0 {
		width = 640
	}
	if height <= 0 {
		height = 480
	}
	if background == (render.Color{}) {
		background = render.Color{R: 1, G: 1, B: 1, A: 1}
	}
	return &Renderer{
		width:      width,
		height:     height,
		background: background,
		resolution: 72,
	}, nil
}

// SetResolution implements render.DPIAware.
func (r *Renderer) SetResolution(dpi uint) {
	if dpi == 0 {
		dpi = 72
	}
	r.resolution = dpi
}

// Begin starts a drawing session for the given viewport.
func (r *Renderer) Begin(viewport geom.Rect) error {
	if r.began {
		return errors.New("ps: Begin called twice")
	}
	r.began = true
	r.viewport = viewport
	r.content.Reset()
	r.document = nil
	r.stack = r.stack[:0]

	r.content.WriteString("gsave\n")
	fmt.Fprintf(&r.content, "0 %s translate\n", shortFloat(float64(r.height)))
	r.content.WriteString("1 -1 scale\n")
	if r.background.A > 0 {
		writeFillColor(&r.content, r.background)
		fmt.Fprintf(&r.content, "newpath 0 0 moveto %s 0 lineto %s %s lineto 0 %s lineto closepath fill\n",
			shortFloat(float64(r.width)),
			shortFloat(float64(r.width)),
			shortFloat(float64(r.height)),
			shortFloat(float64(r.height)),
		)
	}
	return nil
}

// End finalizes the current drawing session.
func (r *Renderer) End() error {
	if !r.began {
		return errors.New("ps: End called before Begin")
	}
	r.began = false
	r.content.WriteString("grestore\n")
	r.document = buildDocument(r.width, r.height, r.content.String(), false)
	return nil
}

// Save pushes graphics state.
func (r *Renderer) Save() {
	r.stack = append(r.stack, state{inContent: r.began})
	if r.began {
		r.content.WriteString("gsave\n")
	}
}

// Restore pops graphics state.
func (r *Renderer) Restore() {
	if len(r.stack) == 0 {
		return
	}
	top := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	if top.inContent && r.began {
		r.content.WriteString("grestore\n")
	}
}

// ClipRect installs a rectangular clip.
func (r *Renderer) ClipRect(rect geom.Rect) {
	if !r.began {
		return
	}
	fmt.Fprintf(&r.content, "newpath %s %s moveto %s %s lineto %s %s lineto %s %s lineto closepath clip newpath\n",
		shortFloat(rect.Min.X), shortFloat(rect.Min.Y),
		shortFloat(rect.Max.X), shortFloat(rect.Min.Y),
		shortFloat(rect.Max.X), shortFloat(rect.Max.Y),
		shortFloat(rect.Min.X), shortFloat(rect.Max.Y),
	)
}

// ClipPath installs an arbitrary path clip.
func (r *Renderer) ClipPath(p geom.Path) {
	if !r.began {
		return
	}
	if !writePathOps(&r.content, p) {
		return
	}
	r.content.WriteString("clip newpath\n")
}

// Path draws a path using the provided paint.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if !r.began || paint == nil {
		return
	}
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	hasFill := paint.Fill.A > 0 || hasHatch
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	if !hasFill && !hasStroke {
		return
	}
	if !writePathOps(&r.content, p) {
		return
	}

	switch {
	case hasHatch:
		if paint.Fill.A > 0 {
			writeFillColor(&r.content, paint.Fill)
			r.content.WriteString("gsave fill grestore\n")
		}
		r.writeHatchFill(p, paint)
		if hasStroke {
			if !writePathOps(&r.content, p) {
				return
			}
			writeStrokeColor(&r.content, paint.Stroke)
			writeLineState(&r.content, paint)
			r.content.WriteString("stroke\n")
		}
	case hasFill && hasStroke:
		writeFillColor(&r.content, paint.Fill)
		r.content.WriteString("gsave fill grestore\n")
		writeStrokeColor(&r.content, paint.Stroke)
		writeLineState(&r.content, paint)
		r.content.WriteString("stroke\n")
	case hasFill:
		writeFillColor(&r.content, paint.Fill)
		r.content.WriteString("fill\n")
	case hasStroke:
		writeStrokeColor(&r.content, paint.Stroke)
		writeLineState(&r.content, paint)
		r.content.WriteString("stroke\n")
	}
}

// SupportsNativeHatch reports that the PS backend consumes hatch metadata
// directly in Path.
func (r *Renderer) SupportsNativeHatch() bool { return true }

func (r *Renderer) writeHatchFill(p geom.Path, paint *render.Paint) {
	if paint == nil || paint.Hatch == "" || paint.HatchColor.A <= 0 {
		return
	}
	if !writePathOps(&r.content, p) {
		return
	}
	r.content.WriteString("gsave clip newpath\n")
	writeStrokeColor(&r.content, paint.HatchColor)
	lineWidth := paint.HatchLineWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	fmt.Fprintf(&r.content, "%s setlinewidth\n", shortFloat(lineWidth))
	r.content.WriteString("0 setlinecap\n")
	for _, line := range hatchPatternLines(paint.Hatch, paint.HatchSpacing) {
		fmt.Fprintf(&r.content, "newpath %s %s moveto %s %s lineto\nstroke\n",
			shortFloat(line[0].X), shortFloat(line[0].Y),
			shortFloat(line[1].X), shortFloat(line[1].Y),
		)
	}
	r.content.WriteString("grestore\n")
}

// Image draws an RGBA raster image into the destination rectangle using a
// Level-2 colorimage operator. PostScript has no native alpha channel, so
// translucent image pixels are pre-composited over white for this first slice.
func (r *Renderer) Image(img render.Image, dst geom.Rect) {
	if !r.began || img == nil || dst.W() <= 0 || dst.H() <= 0 {
		return
	}
	rgb, width, height, ok := encodePSImageRGB(img)
	if !ok {
		return
	}
	r.writeImageWithMatrix(rgb, width, height, geom.Affine{
		A: dst.W(),
		D: dst.H(),
		E: dst.Min.X,
		F: dst.Min.Y,
	})
}

// ImageTransformed draws a raster image through an arbitrary affine transform.
// The affine maps source image pixels into display coordinates.
func (r *Renderer) ImageTransformed(img render.Image, _ geom.Rect, transform geom.Affine) {
	if !r.began || img == nil {
		return
	}
	width, height := img.Size()
	if width <= 0 || height <= 0 {
		return
	}
	rgb, _, _, ok := encodePSImageRGB(img)
	if !ok {
		return
	}
	r.writeImageWithMatrix(rgb, width, height, geom.Affine{
		A: transform.A * float64(width),
		B: transform.B * float64(width),
		C: transform.C * float64(height),
		D: transform.D * float64(height),
		E: transform.E,
		F: transform.F,
	})
}

func (r *Renderer) writeImageWithMatrix(rgb string, width, height int, matrix geom.Affine) {
	fmt.Fprintf(&r.content, "gsave\n[%s %s %s %s %s %s] concat\n/DeviceRGB setcolorspace\n",
		shortFloat(matrix.A),
		shortFloat(matrix.B),
		shortFloat(matrix.C),
		shortFloat(matrix.D),
		shortFloat(matrix.E),
		shortFloat(matrix.F),
	)
	fmt.Fprintf(&r.content, "%d %d 8 [%d 0 0 -%d 0 %d]\n", width, height, width, height, height)
	fmt.Fprintf(&r.content, "{<%s>} false 3 colorimage\ngrestore\n", rgb)
}

// GlyphRun draws a shaped run using glyph IDs as Unicode scalar values when
// possible. Core text paths normally call DrawText directly.
func (r *Renderer) GlyphRun(run render.GlyphRun, textColor render.Color) {
	if !r.began || len(run.Glyphs) == 0 || run.Size <= 0 || textColor.A <= 0 {
		return
	}
	penX := run.Origin.X
	penY := run.Origin.Y
	for _, glyph := range run.Glyphs {
		if glyph.ID != 0 {
			r.DrawTextWithFont(string(rune(glyph.ID)), geom.Pt{
				X: penX + glyph.Offset.X,
				Y: penY + glyph.Offset.Y,
			}, run.Size, textColor, run.FontKey)
		}
		advance := glyph.Advance
		if advance == 0 && glyph.ID != 0 {
			advance = r.MeasureText(string(rune(glyph.ID)), run.Size, run.FontKey).W
		}
		penX += advance
	}
}

// MeasureText returns deterministic approximate text metrics for layout.
func (r *Renderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	if size <= 0 {
		size = defaultFontHeight
	}
	width := 0.6 * size * float64(len([]rune(text)))
	ascent := 0.8 * size
	descent := 0.2 * size
	return render.TextMetrics{W: width, H: ascent + descent, Ascent: ascent, Descent: descent}
}

// DrawText draws text using a standard PostScript base font.
func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	r.DrawTextWithFont(text, origin, size, textColor, "")
}

// DrawTextWithFont draws text with an explicit font key. The first slice maps
// every key to Helvetica so text remains searchable in PS without embedding.
func (r *Renderer) DrawTextWithFont(text string, origin geom.Pt, size float64, textColor render.Color, _ string) {
	if !r.began || text == "" || size <= 0 || textColor.A <= 0 {
		return
	}
	r.writeTextAt(text, origin, size, 0, textColor)
}

// DrawTextRotated draws text around the supplied anchor point.
func (r *Renderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	r.DrawTextRotatedWithFont(text, anchor, size, angle, textColor, "")
}

// DrawTextRotatedWithFont draws rotated text with an explicit font key.
func (r *Renderer) DrawTextRotatedWithFont(text string, anchor geom.Pt, size, angle float64, textColor render.Color, _ string) {
	if !r.began || text == "" || size <= 0 || textColor.A <= 0 {
		return
	}
	r.writeTextAt(text, anchor, size, angle, textColor)
}

// DrawTextVertical draws vertical text centered on the supplied point.
func (r *Renderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	r.DrawTextVerticalWithFont(text, center, size, textColor, "")
}

// DrawTextVerticalWithFont draws vertical text with an explicit font key.
func (r *Renderer) DrawTextVerticalWithFont(text string, center geom.Pt, size float64, textColor render.Color, _ string) {
	if !r.began || text == "" || size <= 0 || textColor.A <= 0 {
		return
	}
	runes := []rune(text)
	lineHeight := r.MeasureText("M", size, "").H
	startY := center.Y - lineHeight*float64(len(runes)-1)/2
	for i, ch := range runes {
		r.writeTextAt(string(ch), geom.Pt{X: center.X, Y: startY + float64(i)*lineHeight}, size, 0, textColor)
	}
}

func (r *Renderer) writeTextAt(text string, origin geom.Pt, size, angle float64, textColor render.Color) {
	r.content.WriteString("gsave\n")
	writeFillColor(&r.content, textColor)
	fmt.Fprintf(&r.content, "%s %s translate\n", shortFloat(origin.X), shortFloat(origin.Y))
	if angle != 0 && !math.IsNaN(angle) && !math.IsInf(angle, 0) {
		fmt.Fprintf(&r.content, "%s rotate\n", shortFloat(angle))
	}
	r.content.WriteString("1 -1 scale\n")
	fmt.Fprintf(&r.content, "/Helvetica findfont %s scalefont setfont\n", shortFloat(size))
	fmt.Fprintf(&r.content, "0 0 moveto (%s) show\n", escapePSString(text))
	r.content.WriteString("grestore\n")
}

// SavePS writes the finalized document to path.
func (r *Renderer) SavePS(path string) error {
	if len(r.document) == 0 {
		return errors.New("ps: SavePS called before End")
	}
	ext := strings.ToLower(filepath.Ext(path))
	data := r.document
	if ext == ".eps" {
		data = buildDocument(r.width, r.height, r.content.String(), true)
	}
	return os.WriteFile(path, data, 0o644)
}

func buildDocument(width, height int, content string, eps bool) []byte {
	var b strings.Builder
	if eps {
		b.WriteString("%!PS-Adobe-3.0 EPSF-3.0\n")
	} else {
		b.WriteString("%!PS-Adobe-3.0\n")
	}
	b.WriteString("%%Creator: matplotlib-go\n")
	fmt.Fprintf(&b, "%%%%BoundingBox: 0 0 %d %d\n", width, height)
	b.WriteString("%%LanguageLevel: 2\n")
	if !eps {
		b.WriteString("%%Pages: 1\n")
	}
	b.WriteString("%%EndComments\n")
	if !eps {
		b.WriteString("%%Page: 1 1\n")
	}
	b.WriteString(content)
	if !eps {
		b.WriteString("showpage\n")
	}
	b.WriteString("%%EOF\n")
	return []byte(b.String())
}

func writePathOps(w *strings.Builder, p geom.Path) bool {
	if !p.Validate() || len(p.C) == 0 {
		return false
	}
	w.WriteString("newpath\n")
	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			pt := p.V[vi]
			vi++
			fmt.Fprintf(w, "%s %s moveto\n", shortFloat(pt.X), shortFloat(pt.Y))
		case geom.LineTo:
			pt := p.V[vi]
			vi++
			fmt.Fprintf(w, "%s %s lineto\n", shortFloat(pt.X), shortFloat(pt.Y))
		case geom.QuadTo:
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
			fmt.Fprintf(w, "%s %s %s %s %s %s curveto\n",
				shortFloat(c1.X), shortFloat(c1.Y),
				shortFloat(c2.X), shortFloat(c2.Y),
				shortFloat(end.X), shortFloat(end.Y),
			)
		case geom.CubicTo:
			c1 := p.V[vi]
			c2 := p.V[vi+1]
			end := p.V[vi+2]
			vi += 3
			fmt.Fprintf(w, "%s %s %s %s %s %s curveto\n",
				shortFloat(c1.X), shortFloat(c1.Y),
				shortFloat(c2.X), shortFloat(c2.Y),
				shortFloat(end.X), shortFloat(end.Y),
			)
		case geom.ClosePath:
			w.WriteString("closepath\n")
		}
	}
	return true
}

func lastEndpoint(p geom.Path, vi int) geom.Pt {
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
		}
	}
	return geom.Pt{}
}

func writeLineState(w *strings.Builder, paint *render.Paint) {
	lineWidth := paint.LineWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	fmt.Fprintf(w, "%s setlinewidth\n", shortFloat(lineWidth))
	fmt.Fprintf(w, "%d setlinejoin\n", lineJoin(paint.LineJoin))
	fmt.Fprintf(w, "%d setlinecap\n", lineCap(paint.LineCap))
	if paint.MiterLimit > 0 {
		fmt.Fprintf(w, "%s setmiterlimit\n", shortFloat(paint.MiterLimit))
	}
	if len(paint.Dashes) > 0 {
		w.WriteByte('[')
		for i, d := range paint.Dashes {
			if i > 0 {
				w.WriteByte(' ')
			}
			w.WriteString(shortFloat(d))
		}
		w.WriteString("] 0 setdash\n")
	} else {
		w.WriteString("[] 0 setdash\n")
	}
}

func writeStrokeColor(w *strings.Builder, c render.Color) {
	writeColor(w, c)
}

func writeFillColor(w *strings.Builder, c render.Color) {
	writeColor(w, c)
}

func writeColor(w *strings.Builder, c render.Color) {
	fmt.Fprintf(w, "%s %s %s setrgbcolor\n", shortFloat(clamp01(c.R)), shortFloat(clamp01(c.G)), shortFloat(clamp01(c.B)))
}

func lineJoin(join render.LineJoin) int {
	switch join {
	case render.JoinRound:
		return 1
	case render.JoinBevel:
		return 2
	default:
		return 0
	}
}

func lineCap(lineCap render.LineCap) int {
	switch lineCap {
	case render.CapRound:
		return 1
	case render.CapSquare:
		return 2
	default:
		return 0
	}
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

func escapePSString(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '(', ')':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 || r > 0x7e {
				b.WriteByte('?')
				continue
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}

func imageAlphaMultiplier(img render.Image) float64 {
	if alphaImage, ok := img.(render.ImageAlpha); ok {
		return clamp01(alphaImage.Alpha())
	}
	return 1
}

func encodePSImageRGB(img render.Image) (string, int, int, bool) {
	rgbaSource, ok := img.(render.RGBAImage)
	if !ok || rgbaSource.RGBA() == nil {
		return "", 0, 0, false
	}
	src := rgbaSource.RGBA()
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return "", 0, 0, false
	}
	alphaMul := imageAlphaMultiplier(img)
	var b strings.Builder
	b.Grow(width * height * 6)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.RGBAAt(x, y)
			r, g, blue := compositeImagePixelOverWhite(c.R, c.G, c.B, c.A, alphaMul)
			fmt.Fprintf(&b, "%02x%02x%02x", r, g, blue)
		}
	}
	return b.String(), width, height, true
}

func compositeImagePixelOverWhite(red, green, blue, alpha uint8, alphaMul float64) (uint8, uint8, uint8) {
	a := clamp01(float64(alpha) / 255.0 * alphaMul)
	return uint8(float64(red)*a + 255*(1-a) + 0.5),
		uint8(float64(green)*a + 255*(1-a) + 0.5),
		uint8(float64(blue)*a + 255*(1-a) + 0.5)
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
