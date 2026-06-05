package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTransformedImageBackendMatrixIsDocumented(t *testing.T) {
	sourceRequirements := map[string][]string{
		filepath.Join("..", "core", "image.go"): {
			"if tr, ok := r.(render.ImageTransformer); ok",
			"Fallback: ignore rotation and render axis-aligned image.",
			"data.SetInterpolation(i.Interpolation)",
		},
		filepath.Join("..", "backends", "agg", "init.go"): {
			"backends.ImageTransform",
		},
		filepath.Join("..", "backends", "agg", "agg_draw.go"): {
			"applyInterpolation(agg, img, w, h)",
			"nearestScaledImageForDirectDraw",
			"func (r *Renderer) ImageTransformed",
			"applyInterpolation(agg, img, affineDispX, affineDispY)",
		},
		filepath.Join("..", "backends", "gobasic", "init.go"): {
			"backends.ImageTransform",
		},
		filepath.Join("..", "backends", "gobasic", "gobasic.go"): {
			"func (r *Renderer) ImageTransformed",
			"drawBitmapScaledWithAlpha",
			"drawBitmapTransformed",
		},
		filepath.Join("..", "backends", "svg", "init.go"): {
			"backends.ImageTransform",
		},
		filepath.Join("..", "backends", "svg", "image.go"): {
			"func (r *Renderer) ImageTransformed",
			"r.renderImageNode(rgba, dst, matrixTransform",
		},
		filepath.Join("..", "backends", "pdf", "init.go"): {
			"backends.ImageTransform",
		},
		filepath.Join("..", "backends", "pdf", "pdf.go"): {
			"func (r *Renderer) ImageTransformed",
			"r.drawImageWithMatrix(img, matrix)",
		},
		filepath.Join("..", "backends", "skia", "init.go"): {
			"if available",
			"backends.ImageTransform",
		},
		filepath.Join("..", "backends", "skia", "skia_stub.go"): {
			"skia backend not available: build with -tags skia",
		},
		filepath.Join("..", "backends", "skia", "skia_native.go"): {
			"_ render.ImageTransformer",
		},
	}
	for path, phrases := range sourceRequirements {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		src := string(data)
		for _, phrase := range phrases {
			if !strings.Contains(src, phrase) {
				t.Fatalf("%s missing backend matrix source marker %q", path, phrase)
			}
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.75.5 Transformed Image Backend Matrix",
		"`AGG`",
		"native `render.ImageTransformer`",
		"consumes `Interpolation()`",
		"`GoBasic`",
		"nearest-style bitmap scaling",
		"does not consume interpolation names",
		"`SVG`",
		"`PDF`",
		"embed transformed raster image objects",
		"`Skia`",
		"optional `-tags skia`",
		"core fallback ignores rotation",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("transformed image backend matrix docs missing %q", phrase)
		}
	}
}
