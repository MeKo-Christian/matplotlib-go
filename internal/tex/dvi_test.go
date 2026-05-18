package tex

import (
	"bytes"
	"encoding/binary"
	"math"
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
	b.WriteByte(243)
	b.WriteByte(id)
	writeU4(b, 0)
	writeU4(b, 65536)
	writeU4(b, 65536)
	b.WriteByte(0)
	b.WriteByte(byte(len(name)))
	b.WriteString(name)
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
