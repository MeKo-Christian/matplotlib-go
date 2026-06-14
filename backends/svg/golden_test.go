package svg

import (
	"flag"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/svgcompare"
	"github.com/cwbudde/matplotlib-go/render"
)

// updateGoldens controls whether failing golden tests rewrite the on-disk
// fixture instead of asserting. Pass -update on the test binary to re-bake
// fixtures after an intentional output change:
//
//	go test ./backends/svg/... -run Golden -update
var updateGoldens = flag.Bool("update", false, "rewrite SVG goldens instead of comparing")

// goldenFixture is one golden-test case. The render closure receives a
// renderer that has already had Begin called with the given viewport; the
// harness calls End and renderSVG after the closure returns.
type goldenFixture struct {
	name     string
	width    int
	height   int
	viewport geom.Rect
	render   func(r *Renderer)
}

func goldenFixtures() []goldenFixture {
	square := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}

	return []goldenFixture{
		{
			name:     "line_stroked",
			width:    100,
			height:   100,
			viewport: square,
			render: func(r *Renderer) {
				var path geom.Path
				path.MoveTo(geom.Pt{X: 10, Y: 10})
				path.LineTo(geom.Pt{X: 90, Y: 90})
				r.Path(path, &render.Paint{
					Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
					LineWidth: 1.5,
				})
			},
		},
		{
			name:     "scatter_markers",
			width:    100,
			height:   100,
			viewport: square,
			render: func(r *Renderer) {
				var marker geom.Path
				marker.MoveTo(geom.Pt{X: -2, Y: 0})
				marker.LineTo(geom.Pt{X: 2, Y: 0})
				marker.LineTo(geom.Pt{X: 0, Y: 2})
				marker.Close()
				r.DrawMarkers(render.MarkerBatch{
					Marker: marker,
					Items: []render.MarkerItem{
						{
							Offset: geom.Pt{X: 20, Y: 30}, Transform: geom.Identity(),
							Paint: render.Paint{Fill: render.Color{R: 1, A: 1}, Stroke: render.Color{A: 1}, LineWidth: 1},
						},
						{
							Offset: geom.Pt{X: 50, Y: 50}, Transform: geom.Identity(),
							Paint: render.Paint{Fill: render.Color{G: 1, A: 1}, Stroke: render.Color{A: 1}, LineWidth: 1},
						},
						{
							Offset: geom.Pt{X: 80, Y: 70}, Transform: geom.Identity(),
							Paint: render.Paint{Fill: render.Color{B: 1, A: 1}, Stroke: render.Color{A: 1}, LineWidth: 1},
						},
					},
				})
			},
		},
		{
			name:     "clipped_text",
			width:    100,
			height:   100,
			viewport: square,
			render: func(r *Renderer) {
				r.ClipRect(geom.Rect{Min: geom.Pt{X: 10, Y: 10}, Max: geom.Pt{X: 90, Y: 50}})
				r.DrawText("clipped", geom.Pt{X: 20, Y: 30}, 12, render.Color{A: 1})
			},
		},
		{
			name:     "image_transformed",
			width:    100,
			height:   100,
			viewport: square,
			render: func(r *Renderer) {
				img := image.NewRGBA(image.Rect(0, 0, 2, 2))
				img.SetRGBA(0, 0, color.RGBA{R: 200, G: 100, B: 50, A: 255})
				img.SetRGBA(1, 1, color.RGBA{R: 50, G: 100, B: 200, A: 255})
				r.ImageTransformed(render.NewImageData(img), geom.Rect{
					Min: geom.Pt{X: 0, Y: 0},
					Max: geom.Pt{X: 2, Y: 2},
				}, geom.Affine{A: 10, B: 0, C: 0, D: 10, E: 30, F: 40})
			},
		},
	}
}

func TestSVGGoldens(t *testing.T) {
	for _, fx := range goldenFixtures() {
		t.Run(fx.name, func(t *testing.T) {
			actual := renderFixture(t, fx)
			goldenPath := filepath.Join("testdata", "golden", fx.name+".svg")

			if *updateGoldens {
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
					t.Fatalf("mkdir golden dir: %v", err)
				}
				if err := os.WriteFile(goldenPath, []byte(actual), 0o644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				t.Logf("updated %s", goldenPath)
				return
			}

			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden %s: %v\n(rerun with -update to create it)", goldenPath, err)
			}
			diff, err := svgcompare.ParseAndDiff(normalizeLegacyGoldenMetadata(expected), []byte(actual))
			if err != nil {
				t.Fatalf("parse: %v\nactual:\n%s", err, actual)
			}
			if diff != "" {
				t.Fatalf("golden mismatch for %s:\n  %s\n\nactual SVG:\n%s", fx.name, diff, actual)
			}
		})
	}
}

func normalizeLegacyGoldenMetadata(expected []byte) []byte {
	content := string(expected)
	if strings.Contains(content, "<metadata") {
		return expected
	}
	const marker = "preserveAspectRatio=\"xMidYMid meet\">\n"
	if !strings.Contains(content, marker) {
		return expected
	}
	content = strings.Replace(content, marker, marker+"  <metadata></metadata>\n", 1)
	return []byte(content)
}

func renderFixture(t *testing.T, fx goldenFixture) string {
	t.Helper()
	r, err := New(fx.width, fx.height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := r.Begin(fx.viewport); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	fx.render(r)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	return r.renderSVG()
}
