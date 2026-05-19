package tex

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"testing"
)

type testDVIMetrics struct{}

func (testDVIMetrics) GlyphMetrics(_ DVIFont, glyph uint32) (DVIGlyphMetrics, bool) {
	if glyph != 'A' {
		return DVIGlyphMetrics{}, false
	}
	return DVIGlyphMetrics{Width: 30, Height: 10, Depth: 3}, true
}

func TestParseDVIGeometryUsesGlyphHeightDepthAndWidth(t *testing.T) {
	pages, err := ParseDVI(minimalDVI(func(b *bytes.Buffer) {
		writeFontDef1(b, 1, "cm")
		writeBOP(b)
		b.WriteByte(235) // fnt1
		b.WriteByte(1)
		b.WriteByte('A') // set_char_65
		b.WriteByte(140) // eop
	}), 72, testDVIMetrics{})
	if err != nil {
		t.Fatalf("ParseDVI: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("ParseDVI page count = %d, want 1", len(pages))
	}

	scale := 72.0 / (72.27 * 65536)
	page := pages[0]
	if !almostDVI(page.Width, 30*scale) {
		t.Fatalf("page width = %v, want %v", page.Width, 30*scale)
	}
	if !almostDVI(page.Height, 10*scale) {
		t.Fatalf("page height = %v, want %v", page.Height, 10*scale)
	}
	if !almostDVI(page.Descent, 3*scale) {
		t.Fatalf("page descent = %v, want %v", page.Descent, 3*scale)
	}
	if len(page.Text) != 1 || page.Text[0].Glyph != 'A' || !almostDVI(page.Text[0].Width, page.Width) {
		t.Fatalf("unexpected text entries: %+v", page.Text)
	}
}

func TestParseDVIGeometryUsesRules(t *testing.T) {
	pages, err := ParseDVI(minimalDVI(func(b *bytes.Buffer) {
		writeBOP(b)
		b.WriteByte(132) // set_rule
		writeS4(b, 10)
		writeS4(b, 20)
		b.WriteByte(140) // eop
	}), 72, nil)
	if err != nil {
		t.Fatalf("ParseDVI: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("ParseDVI page count = %d, want 1", len(pages))
	}

	scale := 72.0 / (72.27 * 65536)
	page := pages[0]
	if !almostDVI(page.Width, 20*scale) || !almostDVI(page.Height, 10*scale) || page.Descent != 0 {
		t.Fatalf("page metrics = %+v, want width=%v height=%v descent=0", page, 20*scale, 10*scale)
	}
	if len(page.Boxes) != 1 || !almostDVI(page.Boxes[0].Width, page.Width) || !almostDVI(page.Boxes[0].Height, page.Height) {
		t.Fatalf("unexpected rule boxes: %+v", page.Boxes)
	}
}

func TestManagerRenderPrefersDVIPageMetricsWhenAvailable(t *testing.T) {
	dir := t.TempDir()
	manager := NewManager(ManagerConfig{
		CacheDir:      dir,
		LaTeXCommand:  "unused-latex",
		DVIPNGCommand: "unused-dvipng",
	})
	text := `\rule{20sp}{10sp}`
	source := Source(text, 12, "DejaVu Sans")
	dviBase, err := manager.basePath(source, 0)
	if err != nil {
		t.Fatalf("basePath dvi: %v", err)
	}
	if err := os.WriteFile(dviBase+".dvi", minimalDVI(func(b *bytes.Buffer) {
		writeBOP(b)
		b.WriteByte(132) // set_rule
		writeS4(b, 10)
		writeS4(b, 20)
		b.WriteByte(140) // eop
	}), 0o644); err != nil {
		t.Fatalf("write cached dvi: %v", err)
	}
	pngBase, err := manager.basePath(source, 72)
	if err != nil {
		t.Fatalf("basePath png: %v", err)
	}
	writeDVIPNGFixture(t, pngBase+".png", 100, 50)

	result, err := manager.Render(text, 12, 72, "DejaVu Sans")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	scale := 72.0 / (72.27 * 65536)
	if !almostDVI(result.Metrics.W, 20*scale) || !almostDVI(result.Metrics.H, 10*scale) || result.Metrics.Descent != 0 {
		t.Fatalf("Render metrics = %+v, want DVI geometry width=%v height=%v descent=0", result.Metrics, 20*scale, 10*scale)
	}
}

func TestManagerRenderUsesTFMMetricsForDVIGlyphs(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/cm.tfm", minimalTFM('A', 30, 10, 3), 0o644); err != nil {
		t.Fatalf("write tfm: %v", err)
	}
	manager := NewManager(ManagerConfig{
		CacheDir:      dir,
		LaTeXCommand:  "unused-latex",
		DVIPNGCommand: "unused-dvipng",
		TFMDirs:       []string{dir},
	})
	text := `A`
	source := Source(text, 12, "DejaVu Sans")
	dviBase, err := manager.basePath(source, 0)
	if err != nil {
		t.Fatalf("basePath dvi: %v", err)
	}
	if err := os.WriteFile(dviBase+".dvi", minimalDVI(func(b *bytes.Buffer) {
		writeFontDef1Scale(b, 1, "cm", 1<<20)
		writeBOP(b)
		b.WriteByte(235) // fnt1
		b.WriteByte(1)
		b.WriteByte('A')
		b.WriteByte(140) // eop
	}), 0o644); err != nil {
		t.Fatalf("write cached dvi: %v", err)
	}
	pngBase, err := manager.basePath(source, 72)
	if err != nil {
		t.Fatalf("basePath png: %v", err)
	}
	writeDVIPNGFixture(t, pngBase+".png", 100, 50)

	result, err := manager.Render(text, 12, 72, "DejaVu Sans")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	scale := 72.0 / (72.27 * 65536)
	if !almostDVI(result.Metrics.W, 30*scale) || !almostDVI(result.Metrics.H, 13*scale) || !almostDVI(result.Metrics.Descent, 3*scale) {
		t.Fatalf("Render metrics = %+v, want DVI/TFM width=%v height=%v descent=%v", result.Metrics, 30*scale, 13*scale, 3*scale)
	}
}

func minimalDVI(body func(*bytes.Buffer)) []byte {
	var b bytes.Buffer
	b.WriteByte(247) // pre
	b.WriteByte(2)
	writeU4(&b, 25400000)
	writeU4(&b, 7227*65536)
	writeU4(&b, 1000)
	b.WriteByte(0)
	body(&b)
	b.WriteByte(248) // post
	return b.Bytes()
}

func writeBOP(b *bytes.Buffer) {
	b.WriteByte(139)
	for i := 0; i < 10; i++ {
		writeS4(b, 0)
	}
	writeS4(b, -1)
}

func writeFontDef1(b *bytes.Buffer, id uint8, name string) {
	writeFontDef1Scale(b, id, name, 65536)
}

func writeFontDef1Scale(b *bytes.Buffer, id uint8, name string, scale uint32) {
	b.WriteByte(243)
	b.WriteByte(id)
	writeU4(b, 0)
	writeU4(b, scale)
	writeU4(b, 65536)
	b.WriteByte(0)
	b.WriteByte(byte(len(name)))
	b.WriteString(name)
}

func minimalTFM(char byte, width, height, depth int32) []byte {
	var b bytes.Buffer
	for _, value := range []uint16{
		13, 0, uint16(char), uint16(char),
		2, 2, 2, 0,
		0, 0, 0, 0,
	} {
		_ = binary.Write(&b, binary.BigEndian, value)
	}
	b.WriteByte(1)    // width index
	b.WriteByte(0x11) // height index 1, depth index 1
	b.WriteByte(0)
	b.WriteByte(0)
	for _, value := range []int32{0, width, 0, height, 0, depth} {
		_ = binary.Write(&b, binary.BigEndian, value)
	}
	return b.Bytes()
}

func writeU4(b *bytes.Buffer, value uint32) {
	_ = binary.Write(b, binary.BigEndian, value)
}

func writeS4(b *bytes.Buffer, value int32) {
	_ = binary.Write(b, binary.BigEndian, value)
}

func almostDVI(got, want float64) bool {
	return math.Abs(got-want) < 1e-12
}

func writeDVIPNGFixture(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write png: %v", err)
	}
}
