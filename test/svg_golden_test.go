package test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/cwbudde/matplotlib-go/backends/svg"
	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
	"github.com/cwbudde/matplotlib-go/internal/svgcompare"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

var updateSVGGolden = flag.Bool("update-svg-golden", false, "Update SVG structural golden fixtures instead of comparing")

func TestSVGGolden(t *testing.T) {
	for _, c := range svgGoldenCases() {
		c := c
		t.Run(c.ID, func(t *testing.T) {
			runSVGGoldenTest(t, c.ID)
		})
	}
}

func TestSVGGoldenCasesComeFromCatalogAndParityFigures(t *testing.T) {
	cases := svgGoldenCases()
	if len(cases) == 0 {
		t.Fatal("no SVG structural golden cases cataloged")
	}
	for _, c := range cases {
		if c.SVGGoldenFamily == "" {
			t.Fatalf("%s has empty SVGGoldenFamily", c.ID)
		}
		if _, _, err := parity.Figure(c.ID); err != nil {
			t.Fatalf("%s is cataloged for SVG structural goldens but has no parity figure: %v", c.ID, err)
		}
	}
}

func runSVGGoldenTest(t *testing.T, id string) {
	t.Helper()

	actual := renderParitySVG(t, id)
	goldenPath := filepath.Join("..", "testdata", "svg_golden", id+".svg")

	if *updateSVGGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("create SVG golden dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, actual, 0o644); err != nil {
			t.Fatalf("update SVG golden %s: %v", goldenPath, err)
		}
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read SVG golden %s: %v\n(rerun with -update-svg-golden to create it)", goldenPath, err)
	}
	diff, err := svgcompare.ParseAndDiff(expected, actual)
	if err != nil {
		t.Fatalf("compare SVG golden %s: %v", goldenPath, err)
	}
	if diff != "" {
		artifactsDir := writableArtifactsDir(t, filepath.Join("..", "testdata", "_artifacts", "svg_golden"))
		gotPath := filepath.Join(artifactsDir, id+"_got.svg")
		if err := os.WriteFile(gotPath, actual, 0o644); err != nil {
			t.Fatalf("write SVG golden artifact %s: %v", gotPath, err)
		}
		t.Fatalf("SVG golden mismatch for %s:\n%s\nactual: %s", id, diff, gotPath)
	}
}

func renderParitySVG(t *testing.T, id string) []byte {
	t.Helper()

	fig, _, err := parity.Figure(id)
	if err != nil {
		t.Fatalf("parity figure %s: %v", id, err)
	}
	path := filepath.Join(t.TempDir(), id+".svg")
	if err := fig.Save(path); err != nil {
		t.Fatalf("render SVG %s: %v", path, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read rendered SVG %s: %v", path, err)
	}
	return data
}

func svgGoldenCases() []examplecatalog.Case {
	out := []examplecatalog.Case{}
	for _, c := range allCases() {
		if c.SVGGoldenFamily != "" {
			out = append(out, c)
		}
	}
	return out
}
