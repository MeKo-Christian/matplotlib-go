package pdf

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/pdfcompare"
	"github.com/cwbudde/matplotlib-go/render"
)

func newTestRenderer(t *testing.T) *Renderer {
	t.Helper()
	r, err := New(200, 100, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return r
}

func rewriteXRefOffsetsForTest(data []byte) []byte {
	out := append([]byte(nil), data...)
	xrefStart := bytes.Index(out, []byte("xref\n"))
	trailerStart := bytes.Index(out, []byte("trailer\n"))
	if xrefStart < 0 || trailerStart < 0 || trailerStart <= xrefStart {
		return out
	}
	for i := xrefStart; i+20 <= trailerStart; i++ {
		if (i == xrefStart || out[i-1] == '\n') && tenDigits(out[i:i+10]) && out[i+10] == ' ' {
			copy(out[i:i+10], []byte("9999999999"))
		}
	}
	startXRef := bytes.Index(out, []byte("startxref\n"))
	if startXRef >= 0 {
		valueStart := startXRef + len("startxref\n")
		valueEnd := valueStart
		for valueEnd < len(out) && out[valueEnd] >= '0' && out[valueEnd] <= '9' {
			out[valueEnd] = '1'
			valueEnd++
		}
	}
	return out
}

func tenDigits(b []byte) bool {
	if len(b) != 10 {
		return false
	}
	for _, c := range b {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func pdfTestRectPath(x0, y0, x1, y1 float64) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: x0, Y: y0})
	p.LineTo(geom.Pt{X: x1, Y: y0})
	p.LineTo(geom.Pt{X: x1, Y: y1})
	p.LineTo(geom.Pt{X: x0, Y: y1})
	p.Close()
	return p
}

func pdfTestTrianglePath() geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: -4})
	p.LineTo(geom.Pt{X: 4, Y: 4})
	p.LineTo(geom.Pt{X: -4, Y: 4})
	p.Close()
	return p
}

func mustParsePDF(t *testing.T, r *Renderer) *pdfcompare.Document {
	t.Helper()
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	doc, err := pdfcompare.Parse(data)
	if err != nil {
		t.Fatalf("pdfcompare.Parse: %v", err)
	}
	return doc
}

func pdfDocumentBodyContains(doc *pdfcompare.Document, needle string) bool {
	return pdfDocumentObjectBodyContaining(doc, needle) != ""
}

func pdfDocumentObjectCountContaining(doc *pdfcompare.Document, needle string) int {
	if doc == nil {
		return 0
	}
	count := 0
	for _, obj := range doc.Objects {
		if strings.Contains(obj.Body, needle) {
			count++
		}
	}
	return count
}

func pdfDocumentObjectBodyContaining(doc *pdfcompare.Document, needle string) string {
	if doc == nil {
		return ""
	}
	for _, obj := range doc.Objects {
		if strings.Contains(obj.Body, needle) {
			return obj.Body
		}
	}
	return ""
}

func writePDFFakeCommand(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write fake command: %v", err)
	}
	return path
}

func writePDFTestPNG(t *testing.T, path string, c color.RGBA) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode fixture PNG: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write fixture PNG: %v", err)
	}
}

type jpegTestImage struct {
	w, h int
	data []byte
}

func (j jpegTestImage) Size() (int, int)      { return j.w, j.h }
func (j jpegTestImage) Interpolation() string { return "" }
func (j jpegTestImage) JPEGData() []byte      { return append([]byte(nil), j.data...) }
