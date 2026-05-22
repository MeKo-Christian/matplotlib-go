package test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/pdf"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/pdfcompare"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

var updatePDFGolden = flag.Bool("update-pdf-golden", false, "Update PDF golden fixtures instead of comparing")

var mathTextFixtureIDs = []string{
	"mathtext_basic",
	"mathtext_fractions",
	"mathtext_integrals",
	"mathtext_matrices",
	"mathtext_inline_labels",
}

func TestPDFGolden(t *testing.T) {
	for _, id := range pdfGoldenIDs() {
		id := id
		t.Run(id, func(t *testing.T) {
			runPDFGoldenTest(t, id)
		})
	}
}

func TestPDFGoldenIncludesMathTextFixtures(t *testing.T) {
	ids := map[string]bool{}
	for _, id := range pdfGoldenIDs() {
		ids[id] = true
	}
	for _, id := range mathTextFixtureIDs {
		if !ids[id] {
			t.Fatalf("PDF golden set is missing MathText fixture %q", id)
		}
	}
}

func TestMathTextFixturesHaveToleranceCoverage(t *testing.T) {
	pdfIDs := map[string]bool{}
	for _, id := range pdfGoldenIDs() {
		pdfIDs[id] = true
	}
	for _, id := range mathTextFixtureIDs {
		if !goldenExists(id) {
			t.Fatalf("MathText fixture %q is missing AGG PNG golden coverage", id)
		}
		if !matplotlibRefExists(id) {
			t.Fatalf("MathText fixture %q is missing Matplotlib reference coverage", id)
		}
		if !pdfIDs[id] {
			t.Fatalf("MathText fixture %q is missing PDF golden coverage", id)
		}
		if _, _, err := parity.Figure(id); err != nil {
			t.Fatalf("MathText fixture %q is not registered in parity registry: %v", id, err)
		}
	}
}

func pdfGoldenIDs() []string {
	return []string{
		"basic_line",
		"bar_basic",
		"scatter_basic",
		"hist_basic",
		"mesh_contour_tri",
		"image_heatmap",
		"polar_axes",
		"patch_showcase",
		"text_labels_strict",
		"mathtext_basic",
		"mathtext_fractions",
		"mathtext_integrals",
		"mathtext_matrices",
		"mathtext_inline_labels",
		"imshow_clipped",
		"imshow_transformed",
		"mixed_raster_vector",
	}
}

func runPDFGoldenTest(t *testing.T, id string) {
	t.Helper()

	actual := renderParityPDF(t, id)
	goldenPath := filepath.Join("..", "testdata", "pdf_golden", id+".pdf")

	if *updatePDFGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("create PDF golden dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, actual, 0o644); err != nil {
			t.Fatalf("update PDF golden %s: %v", goldenPath, err)
		}
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read PDF golden %s: %v\n(rerun with -update-pdf-golden to create it)", goldenPath, err)
	}
	diff, err := pdfcompare.ParseAndDiff(expected, actual)
	if err != nil {
		t.Fatalf("compare PDF golden %s: %v", goldenPath, err)
	}
	if diff != "" {
		artifactsDir := writableArtifactsDir(t, filepath.Join("..", "testdata", "_artifacts", "pdf_golden"))
		gotPath := filepath.Join(artifactsDir, id+"_got.pdf")
		if err := os.WriteFile(gotPath, actual, 0o644); err != nil {
			t.Fatalf("write PDF golden artifact %s: %v", gotPath, err)
		}
		t.Fatalf("PDF golden mismatch for %s:\n%s\nactual: %s", id, diff, gotPath)
	}
}

func renderParityPDF(t *testing.T, id string) []byte {
	t.Helper()

	fig, _, err := parity.Figure(id)
	if err != nil {
		t.Fatalf("parity figure %s: %v", id, err)
	}
	r, err := pdf.New(int(fig.SizePx.X), int(fig.SizePx.Y), render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("pdf.New: %v", err)
	}
	core.DrawFigure(fig, r)
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("PDF bytes for %s: %v", id, err)
	}
	return data
}
