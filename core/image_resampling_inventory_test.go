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

func TestTransformedImageMatplotlibComparisonIsDocumented(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "third_party", "matplotlib", "lib", "matplotlib", "image.py"))
	if err != nil {
		t.Fatalf("read upstream image.py: %v", err)
	}
	upstream := string(data)
	upstreamRequirements := []string{
		"def _resample(",
		"def _make_image(self, A, in_bbox, out_bbox, clip_bbox, magnification=1.0,",
		"clipped_bbox = Bbox.intersection(out_bbox, clip_bbox)",
		"out_width_base = clipped_bbox.width * magnification",
		"round_to_pixel_border",
		"if self.origin == 'upper':",
		"interpolation_stage = self._interpolation_stage",
		"if interpolation_stage in ['antialiased', 'auto']:",
		"interpolation_stage = 'rgba'",
		"interpolation_stage = 'data'",
		"def make_image(self, renderer, magnification=1.0, unsampled=False):",
		"clip = ((self.get_clip_box() or self.axes.bbox) if self.get_clip_on()",
		"return self._make_image(self._A, bbox, transformed_bbox, clip,",
		"def set_extent(self, extent, **kwargs):",
		"def get_extent(self):",
	}
	for _, phrase := range upstreamRequirements {
		if !strings.Contains(upstream, phrase) {
			t.Fatalf("upstream image.py missing comparison marker %q", phrase)
		}
	}

	data, err = os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.75.5 Transformed Image Matplotlib Comparison",
		"upstream comparison anchor is `third_party/matplotlib/lib/matplotlib/image.py`",
		"`_ImageBase._make_image`",
		"`AxesImage.make_image`",
		"`AxesImage.set_extent`",
		"`AxesImage.get_extent`",
		"intersects `out_bbox` with `clip_bbox`",
		"`round_to_pixel_border=True`",
		"`origin='upper'`",
		"`origin='lower'`",
		"`interpolation_stage`",
		"`data` and `rgba`",
		"`auto` and `antialiased`",
		"covered Go examples are `image_heatmap`, `collection_mutable_scalarmap`, `colorbar_composition`, and matrix helpers",
		"current Go comparison points are `Image2D.Draw`, `matplotlibImageDrawRect`, `Image2D.rasterizeForRect`, and `Axes.ImShow`",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("transformed image Matplotlib comparison docs missing %q", phrase)
		}
	}
}
