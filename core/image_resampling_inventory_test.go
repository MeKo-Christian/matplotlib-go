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

func TestTransformedImageResamplingGapInventoryIsDocumented(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.75.5 Resampling Gap Inventory",
		"Backend matrix coverage and Matplotlib pipeline comparison are the inventory inputs",
		"`Image2D.Draw`",
		"`Image2D.rasterizeForRect`",
		"`matplotlibImageDrawRect`",
		"`Axes.ImShow`",
		"`_ImageBase._make_image`",
		"`AxesImage.make_image`",
		"`AGG` is the raster backend to align first",
		"`GoBasic` remains the deterministic nearest-style fallback",
		"`SVG` and `PDF` preserve affine image placement",
		"remaining raster gaps are interpolation kernels, antialiasing stage selection, clipping, extent/origin placement, affine transforms, and pixel-center placement",
		"remaining vector gaps are interpolation hints, clipping structure, and documented viewer-side resampling divergence",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("transformed image resampling gap inventory docs missing %q", phrase)
		}
	}
}

func TestTransformedImageInterpolationKernelAlignmentIsDocumented(t *testing.T) {
	sourceRequirements := map[string][]string{
		filepath.Join("..", "backends", "agg", "interpolation.go"): {
			"case \"\", \"none\", \"nearest\":",
			"case \"bilinear\":",
			"case \"bicubic\":",
			"case \"auto\", \"antialiased\":",
			"shouldUseNearestForAutoResample",
		},
		filepath.Join("..", "backends", "agg", "image_interpolation_test.go"): {
			"TestAggImage_AutoInterpolationMatchesNearestForIntegerScale",
			"TestAggImage_AutoInterpolationUsesHanningForNonIntegerScale",
			"TestAggImage_AllMatplotlibInterpolationNamesRender",
		},
		filepath.Join("..", "backends", "gobasic", "gobasic.go"): {
			"nearestScaledSourceIndex",
			"math.Round((float64(rel)+0.5)*float64(srcSize)/float64(dstSize) - 0.5)",
		},
		filepath.Join("..", "backends", "gobasic", "gobasic_test.go"): {
			"TestImageScalingNearestUsesPixelCentersForNonIntegerUpscale",
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
				t.Fatalf("%s missing interpolation alignment source marker %q", path, phrase)
			}
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.75.5 Interpolation Kernel Alignment",
		"`nearest` and `none`",
		"`bilinear` and `bicubic`",
		"`auto` and `antialiased`",
		"`AGG` keeps the Matplotlib interpolation-name registry",
		"`GoBasic` remains nearest-only",
		"`nearestScaledSourceIndex`",
		"GoBasic direct image scaling now samples source pixels from destination pixel centers",
		"remaining kernel limits are AGG's Kaiser fallback and viewer-dependent vector resampling",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("transformed image interpolation kernel alignment docs missing %q", phrase)
		}
	}
}

func TestTransformedImageTransformAndExtentAlignmentIsDocumented(t *testing.T) {
	sourceRequirements := map[string][]string{
		filepath.Join("..", "core", "matrix_helpers.go"): {
			"if cfg.Extent == nil {",
			"cfg.Origin == ImageOriginUpper",
			"a.InvertY()",
		},
		filepath.Join("..", "core", "matrix_helpers_test.go"): {
			"TestImShow_ExplicitExtentOriginUpperDoesNotInvertYLimits",
			"want Matplotlib [30,40]",
		},
		filepath.Join("..", "core", "image.go"): {
			"matplotlibImageDrawRect",
			"imageTransform(dst, raster, anchor, angleRad)",
			"rotationAnchor(ctx, dst)",
		},
		filepath.Join("..", "core", "image_test.go"): {
			"TestImage2D_DrawCeilsRasterSizeAndKeepsMatplotlibAnchor",
			"TestImageTransformPositiveAngleUsesDataSpaceOrientation",
		},
		filepath.Join("..", "backends", "agg", "image_interpolation_test.go"): {
			"TestAggTransformedImagePreservesSourceOrientation",
			"TestAggTransformedImageRespectsClipPathAndAlpha",
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
				t.Fatalf("%s missing transform/extent source marker %q", path, phrase)
			}
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.75.5 Transform and Extent Alignment",
		"`Axes.ImShow` now matches Matplotlib explicit extent handling",
		"`origin='upper'` does not invert explicit `extent=(left, right, bottom, top)` limits",
		"default centered-pixel extents still use origin-driven Y presentation",
		"`matplotlibImageDrawRect`",
		"`imageTransform`",
		"`rotationAnchor`",
		"AGG transformed-image tests pin source orientation, clip-path masking, and alpha",
		"remaining clipping limitation is that Go still clips image pixels at the renderer layer rather than resampling from Matplotlib's `clipped_bbox` output shape",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("transformed image transform/extent alignment docs missing %q", phrase)
		}
	}
}

func TestTransformedImageAggRasterAlignmentIsDocumented(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.75.5 AGG and Raster Backend Alignment",
		"Interpolation Kernel Alignment and Transform and Extent Alignment are the raster alignment inputs",
		"`AGG` is the parity raster backend",
		"`GoBasic` is the deterministic nearest-only raster fallback",
		"`auto` and `antialiased` follow Matplotlib's nearest/Hanning scale split",
		"`nearest` and `none` preserve nearest-neighbor behavior",
		"`Axes.ImShow` explicit extents preserve user-provided limits",
		"`matplotlibImageDrawRect`, `imageTransform`, and `rotationAnchor` pin rounded placement and affine orientation",
		"the remaining raster limitation is clipped scalar-stage resampling from Matplotlib's `clipped_bbox` output shape",
		"fixture refresh can proceed with AGG as the raster reference and GoBasic documented as a fallback",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("transformed image AGG/raster alignment docs missing %q", phrase)
		}
	}
}
