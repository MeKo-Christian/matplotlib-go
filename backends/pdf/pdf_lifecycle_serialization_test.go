package pdf

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/pdfcompare"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestShortFloat(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{1.5, "1.5"},
		{1.5000001, "1.5"},
		{math.Copysign(0, -1), "0"},
		{-1.25, "-1.25"},
		{123.456, "123.456"},
	}
	for _, c := range cases {
		got := shortFloat(c.in)
		if got != c.want {
			t.Errorf("shortFloat(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRendererProducesPDFHeaderAndEOF(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-1.7\n")) {
		t.Errorf("missing PDF-1.7 header; got prefix %q", data[:min(len(data), 16)])
	}
	if !bytes.HasSuffix(data, []byte("%%EOF\n")) {
		tail := data
		if len(tail) > 64 {
			tail = tail[len(tail)-64:]
		}
		t.Errorf("missing %%%%EOF trailer; got tail %q", tail)
	}
}

func TestRendererBeginTwiceFails(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("first Begin: %v", err)
	}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err == nil {
		t.Errorf("second Begin should fail")
	}
}

func TestRendererEndBeforeBeginFails(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.End(); err == nil {
		t.Errorf("End before Begin should fail")
	}
}

func TestRendererSaveRestoreEmitsBracketedQ(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	r.Save()
	r.Save()
	r.Restore()
	r.Restore()
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	// The content stream is flate-compressed, but the buffer is still around
	// while End ran. Re-build a probe renderer to inspect the raw content.
	probe := newTestRenderer(t)
	_ = probe.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	probe.Save()
	probe.Save()
	probe.Restore()
	probe.Restore()
	raw := probe.content.String()
	if strings.Count(raw, "q\n") != 2 {
		t.Errorf("expected 2 q lines, got %d in %q", strings.Count(raw, "q\n"), raw)
	}
	if strings.Count(raw, "Q\n") != 2 {
		t.Errorf("expected 2 Q lines, got %d in %q", strings.Count(raw, "Q\n"), raw)
	}
}

func TestSavePDFWritesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.pdf")

	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 50, Y: 50})
	r.Path(p, &render.Paint{
		Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth: 2,
	})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if err := r.SavePDF(path); err != nil {
		t.Fatalf("SavePDF: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-1.7\n")) {
		t.Errorf("missing PDF header in %q", data[:min(len(data), 16)])
	}
	if !bytes.Contains(data, []byte("startxref")) {
		t.Errorf("missing startxref")
	}
}

func TestSavePDFBeforeEndFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.pdf")

	r := newTestRenderer(t)
	if err := r.SavePDF(path); err == nil {
		t.Errorf("SavePDF before End should fail")
	}
}

func TestSerializationDeterministic(t *testing.T) {
	build := func() []byte {
		r := newTestRenderer(t)
		_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
		var p geom.Path
		p.MoveTo(geom.Pt{X: 10, Y: 10})
		p.LineTo(geom.Pt{X: 90, Y: 10})
		p.LineTo(geom.Pt{X: 90, Y: 50})
		p.Close()
		r.Path(p, &render.Paint{
			Fill:      render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1},
			Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
			LineWidth: 1,
		})
		_ = r.End()
		out, err := r.Bytes()
		if err != nil {
			t.Fatalf("Bytes: %v", err)
		}
		cp := make([]byte, len(out))
		copy(cp, out)
		return cp
	}
	a := build()
	b := build()
	if !bytes.Equal(a, b) {
		t.Errorf("PDF output is not deterministic; len(a)=%d len(b)=%d", len(a), len(b))
	}
}

func TestGeneratedPDFStructuralCompareIgnoresXRefOffsetNoise(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 90, Y: 50})
	r.Path(p, &render.Paint{Stroke: render.Color{A: 1}, LineWidth: 1})
	r.DrawTextWithFont("A", geom.Pt{X: 20, Y: 70}, 12, render.Color{A: 1}, "DejaVu Sans")
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	noisy := rewriteXRefOffsetsForTest(data)
	if bytes.Equal(data, noisy) {
		t.Fatal("test setup failed: xref rewrite did not change the PDF bytes")
	}

	diff, err := pdfcompare.ParseAndDiff(data, noisy)
	if err != nil {
		t.Fatalf("ParseAndDiff: %v", err)
	}
	if diff != "" {
		t.Fatalf("xref offset noise should not produce a structural diff, got: %s", diff)
	}
}

func TestQuadraticPromotedToCubic(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: 0})
	p.QuadTo(geom.Pt{X: 10, Y: 20}, geom.Pt{X: 20, Y: 0})
	r.Path(p, &render.Paint{
		Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth: 1,
	})
	raw := r.content.String()
	// We promoted Quad to Cubic, so we should see a `c` curve operator but no
	// stray Quad-style operator.
	if !strings.Contains(raw, " c\n") {
		t.Errorf("expected cubic-curve operator c in %q", raw)
	}
}
