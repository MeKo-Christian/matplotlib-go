package core

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// testPNGSVGRenderer is a minimal renderer that satisfies render.Renderer (via
// the embedded NullRenderer) plus the export interfaces covered by SaveFig. It
// records which export path was exercised so tests can assert dispatch.
type testPNGSVGRenderer struct {
	render.NullRenderer
	savedPNG   bool
	savedSVG   bool
	savedPDF   bool
	savedPS    bool
	savedPGF   bool
	pngPath    string
	svgPath    string
	pdfPath    string
	psPath     string
	pgfPath    string
	svgOptions render.SVGOptions
	pdfOptions render.PDFOptions
	psOptions  render.PSOptions
	pgfOptions render.PGFOptions
}

func newTestPNGSVGRenderer() *testPNGSVGRenderer { return &testPNGSVGRenderer{} }

func (r *testPNGSVGRenderer) SavePNG(path string) error {
	r.savedPNG = true
	r.pngPath = path
	return nil
}

func (r *testPNGSVGRenderer) SaveSVG(path string) error {
	r.savedSVG = true
	r.svgPath = path
	return nil
}

func (r *testPNGSVGRenderer) SaveSVGWithOptions(path string, opts render.SVGOptions) error {
	r.svgOptions = opts
	return r.SaveSVG(path)
}

func (r *testPNGSVGRenderer) SavePDF(path string) error {
	r.savedPDF = true
	r.pdfPath = path
	return nil
}

func (r *testPNGSVGRenderer) SavePDFWithOptions(path string, opts render.PDFOptions) error {
	r.pdfOptions = opts
	return r.SavePDF(path)
}

func (r *testPNGSVGRenderer) SavePS(path string) error {
	r.savedPS = true
	r.psPath = path
	return nil
}

func (r *testPNGSVGRenderer) SetPSOptions(opts render.PSOptions) {
	r.psOptions = opts
}

func (r *testPNGSVGRenderer) SavePGF(path string) error {
	r.savedPGF = true
	r.pgfPath = path
	return nil
}

func (r *testPNGSVGRenderer) SavePGFWithOptions(path string, opts render.PGFOptions) error {
	r.pgfOptions = opts
	return r.SavePGF(path)
}

// defaultAxesRect is the unit-square rect used by SaveFig tests when adding
// axes; the dispatch logic does not depend on its precise value.
var defaultAxesRect = geom.Rect{
	Min: geom.Pt{X: 0.1, Y: 0.1},
	Max: geom.Pt{X: 0.9, Y: 0.9},
}

func TestSaveFig_DispatchesByExtension_PNG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.png")

	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)

	r := newTestPNGSVGRenderer()
	if err := SaveFig(fig, r, path); err != nil {
		t.Fatalf("SaveFig: %v", err)
	}
	if !r.savedPNG {
		t.Fatal("expected PNG path to be exercised")
	}
	if r.savedSVG {
		t.Fatal("did not expect SVG path")
	}
	if r.pngPath != path {
		t.Fatalf("SavePNG received path %q, want %q", r.pngPath, path)
	}
}

func TestSaveFig_DispatchesByExtension_SVG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.svg")

	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)

	r := newTestPNGSVGRenderer()
	if err := SaveFig(fig, r, path); err != nil {
		t.Fatalf("SaveFig: %v", err)
	}
	if !r.savedSVG {
		t.Fatal("expected SVG path to be exercised")
	}
	if r.savedPNG {
		t.Fatal("did not expect PNG path")
	}
	if r.svgPath != path {
		t.Fatalf("SaveSVG received path %q, want %q", r.svgPath, path)
	}
}

func TestSaveFig_DispatchesByExtension_PDF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.pdf")

	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)

	r := newTestPNGSVGRenderer()
	if err := SaveFig(fig, r, path); err != nil {
		t.Fatalf("SaveFig: %v", err)
	}
	if !r.savedPDF {
		t.Fatal("expected PDF path to be exercised")
	}
	if r.savedPNG || r.savedSVG || r.savedPS {
		t.Fatal("did not expect PNG, SVG, or PostScript path")
	}
	if r.pdfPath != path {
		t.Fatalf("SavePDF received path %q, want %q", r.pdfPath, path)
	}
}

func TestSaveFig_DispatchesByExtension_PSAndEPS(t *testing.T) {
	for _, ext := range []string{".ps", ".eps"} {
		t.Run(ext, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "out"+ext)

			fig := NewFigure(100, 80)
			fig.AddAxes(defaultAxesRect)

			r := newTestPNGSVGRenderer()
			if err := SaveFig(fig, r, path); err != nil {
				t.Fatalf("SaveFig: %v", err)
			}
			if !r.savedPS {
				t.Fatal("expected PostScript path to be exercised")
			}
			if r.savedPNG || r.savedSVG {
				t.Fatal("did not expect PNG or SVG path")
			}
			if r.psPath != path {
				t.Fatalf("SavePS received path %q, want %q", r.psPath, path)
			}
		})
	}
}

func TestSaveFig_DispatchesByExtension_PGF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.pgf")

	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)

	r := newTestPNGSVGRenderer()
	if err := SaveFig(fig, r, path); err != nil {
		t.Fatalf("SaveFig: %v", err)
	}
	if !r.savedPGF {
		t.Fatal("expected PGF path to be exercised")
	}
	if r.savedPNG || r.savedSVG || r.savedPDF || r.savedPS {
		t.Fatal("did not expect PNG, SVG, PDF, or PostScript path")
	}
	if r.pgfPath != path {
		t.Fatalf("SavePGF received path %q, want %q", r.pgfPath, path)
	}
}

func TestSaveFig_ForwardsSVGOptions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.svg")

	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)

	r := newTestPNGSVGRenderer()
	if err := SaveFig(fig, r, path, render.WithSVGMetadata(map[string]string{"Title": "From SaveFig"})); err != nil {
		t.Fatalf("SaveFig: %v", err)
	}
	if !r.savedSVG {
		t.Fatal("expected SVG path to be exercised")
	}
	if got := r.svgOptions.Metadata["Title"]; got != "From SaveFig" {
		t.Fatalf("SaveFig did not forward SVG metadata option, got %q", got)
	}
}

func TestSaveFig_ForwardsPDFOptions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.pdf")

	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)

	r := newTestPNGSVGRenderer()
	creationDate := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	if err := SaveFig(
		fig,
		r,
		path,
		render.WithPDFMetadata(map[string]string{"Title": "From SaveFig"}),
		render.WithPDFCreationDate(creationDate),
		render.WithPDFFontPolicy(render.PDFFontPolicyEmbed),
	); err != nil {
		t.Fatalf("SaveFig: %v", err)
	}
	if !r.savedPDF {
		t.Fatal("expected PDF path to be exercised")
	}
	if got := r.pdfOptions.Metadata["Title"]; got != "From SaveFig" {
		t.Fatalf("SaveFig did not forward PDF metadata option, got %q", got)
	}
	if !r.pdfOptions.CreationDate.Equal(creationDate) {
		t.Fatalf("SaveFig did not forward PDF creation date, got %v", r.pdfOptions.CreationDate)
	}
	if r.pdfOptions.FontPolicy != render.PDFFontPolicyEmbed {
		t.Fatalf("SaveFig did not forward PDF font policy, got %q", r.pdfOptions.FontPolicy)
	}
}

func TestSaveFig_ForwardsPSOptions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.ps")

	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)

	r := newTestPNGSVGRenderer()
	if err := SaveFig(fig, r, path, render.WithPSFontPolicy(render.PSFontPolicyBase14)); err != nil {
		t.Fatalf("SaveFig: %v", err)
	}
	if !r.savedPS {
		t.Fatal("expected PostScript path to be exercised")
	}
	if r.psOptions.FontPolicy != render.PSFontPolicyBase14 {
		t.Fatalf("SaveFig did not forward PS font policy, got %q", r.psOptions.FontPolicy)
	}
}

func TestSaveFig_ForwardsPGFOptions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.pgf")

	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)

	r := newTestPNGSVGRenderer()
	if err := SaveFig(
		fig,
		r,
		path,
		render.WithPGFMetadata(map[string]string{"Title": "From SaveFig"}),
		render.WithPGFPreamble("\\usepackage{amsmath}"),
		render.WithPGFVerificationMode(render.PGFVerificationModeStrict),
	); err != nil {
		t.Fatalf("SaveFig: %v", err)
	}
	if !r.savedPGF {
		t.Fatal("expected PGF path to be exercised")
	}
	if got := r.pgfOptions.Metadata["Title"]; got != "From SaveFig" {
		t.Fatalf("SaveFig did not forward PGF metadata option, got %q", got)
	}
	if len(r.pgfOptions.Preamble) != 1 || r.pgfOptions.Preamble[0] != "\\usepackage{amsmath}" {
		t.Fatalf("SaveFig did not forward PGF preamble, got %#v", r.pgfOptions.Preamble)
	}
	if r.pgfOptions.VerificationMode != render.PGFVerificationModeStrict {
		t.Fatalf("SaveFig did not forward PGF verification mode, got %q", r.pgfOptions.VerificationMode)
	}
}

func TestSaveFig_RejectsOptionsForDifferentFormat(t *testing.T) {
	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)

	r := newTestPNGSVGRenderer()
	err := SaveFig(fig, r, filepath.Join(t.TempDir(), "out.svg"), render.WithPDFMetadata(map[string]string{"Title": "Wrong"}))
	if err == nil {
		t.Fatal("expected mismatched option error")
	}
	if !strings.Contains(err.Error(), "PDF options") || !strings.Contains(err.Error(), ".svg") {
		t.Fatalf("error should mention rejected PDF options and .svg, got: %v", err)
	}
	if r.savedSVG || r.savedPDF {
		t.Fatal("no exporter should have been invoked for mismatched options")
	}
}

func TestSaveFig_RejectsUnknownExtension(t *testing.T) {
	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)
	r := newTestPNGSVGRenderer()

	err := SaveFig(fig, r, "/tmp/out.tiff")
	if err == nil {
		t.Fatal("expected error for unknown extension")
	}
	if !strings.Contains(err.Error(), ".tiff") {
		t.Fatalf("error should mention extension .tiff, got: %v", err)
	}
	if !strings.Contains(err.Error(), ".png") || !strings.Contains(err.Error(), ".svg") ||
		!strings.Contains(err.Error(), ".pdf") ||
		!strings.Contains(err.Error(), ".ps") || !strings.Contains(err.Error(), ".eps") ||
		!strings.Contains(err.Error(), ".pgf") {
		t.Fatalf("error should list supported extensions, got: %v", err)
	}
	if r.savedPNG || r.savedSVG || r.savedPDF || r.savedPS || r.savedPGF {
		t.Fatal("no exporter should have been invoked for unknown extension")
	}
}

func TestSaveFig_NoExtensionRejected(t *testing.T) {
	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)
	r := newTestPNGSVGRenderer()

	err := SaveFig(fig, r, "/tmp/out")
	if err == nil {
		t.Fatal("expected error for missing extension")
	}
	if !strings.Contains(err.Error(), ".png") || !strings.Contains(err.Error(), ".svg") ||
		!strings.Contains(err.Error(), ".pdf") || !strings.Contains(err.Error(), ".ps") ||
		!strings.Contains(err.Error(), ".eps") || !strings.Contains(err.Error(), ".pgf") {
		t.Fatalf("error should list supported extensions, got: %v", err)
	}
	if r.savedPNG || r.savedSVG || r.savedPDF || r.savedPS || r.savedPGF {
		t.Fatal("no exporter should have been invoked for missing extension")
	}
}

func TestSaveFig_UppercaseExtensionWorks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.PNG")

	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)

	r := newTestPNGSVGRenderer()
	if err := SaveFig(fig, r, path); err != nil {
		t.Fatalf("SaveFig with .PNG: %v", err)
	}
	if !r.savedPNG {
		t.Fatal("uppercase extension should still hit PNG path")
	}
}
