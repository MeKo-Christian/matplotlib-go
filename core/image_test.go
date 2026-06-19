package core

import (
	"image"
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

func TestAxesImage_DefaultOptions(t *testing.T) {
	data := [][]float64{
		{0, 1},
		{2, 3},
	}
	ax := &Axes{}
	img := ax.Image(data)

	if img == nil {
		t.Fatal("expected image artist")
	}
	if img.Colormap != "viridis" {
		t.Fatalf("expected default colormap viridis, got %q", img.Colormap)
	}
	if img.VMin != 0 || img.VMax != 3 {
		t.Fatalf("expected vmin/vmax 0..3, got %v..%v", img.VMin, img.VMax)
	}
	if img.Alpha != 1 {
		t.Fatalf("expected alpha 1, got %v", img.Alpha)
	}
	if img.AngleDeg != 0 {
		t.Fatalf("expected angle 0, got %v", img.AngleDeg)
	}
	if img.Origin != ImageOriginUpper {
		t.Fatalf("expected default origin upper, got %v", img.Origin)
	}
	if img.Z() >= NewGrid(AxisBottom).Z() {
		t.Fatalf("expected default image z-order below grid, got image=%v grid=%v", img.Z(), NewGrid(AxisBottom).Z())
	}
}

func TestAxesImage_CustomOptions(t *testing.T) {
	data := [][]float64{{1, 2}, {3, 4}}
	cmap := "gray"
	vmin := -5.0
	vmax := 10.0
	alpha := 2.0
	angle := 45.0
	xMin := -1.0
	xMax := 3.0
	yMin := -2.0
	yMax := 4.0
	rotateX := 1.5
	rotateY := 2.5
	ax := &Axes{}
	img := ax.Image(
		data,
		ImageOptions{
			Colormap:        &cmap,
			VMin:            &vmin,
			VMax:            &vmax,
			Alpha:           &alpha,
			Angle:           &angle,
			XMin:            &xMin,
			XMax:            &xMax,
			YMin:            &yMin,
			YMax:            &yMax,
			Origin:          ImageOriginUpper,
			RotationAnchor:  ImageAnchorCustom,
			RotationAnchorX: &rotateX,
			RotationAnchorY: &rotateY,
		},
	)

	if img.Colormap != "gray" {
		t.Fatalf("expected colormap gray, got %q", img.Colormap)
	}
	if img.VMin != -5 || img.VMax != 10 {
		t.Fatalf("expected vmin/vmax -5..10, got %v..%v", img.VMin, img.VMax)
	}
	if img.Alpha != 1 {
		t.Fatalf("expected alpha clamped to 1, got %v", img.Alpha)
	}
	if img.Origin != ImageOriginUpper {
		t.Fatalf("expected origin upper")
	}
	if img.AngleDeg != 45 {
		t.Fatalf("expected angle 45, got %v", img.AngleDeg)
	}
	if img.RotateAt != ImageAnchorCustom {
		t.Fatalf("expected custom anchor, got %v", img.RotateAt)
	}
	if img.RotateX != rotateX || img.RotateY != rotateY {
		t.Fatalf("unexpected rotation anchor: %v,%v", img.RotateX, img.RotateY)
	}
	if img.XMin != -1 || img.XMax != 3 || img.YMin != -2 || img.YMax != 4 {
		t.Fatalf("unexpected extents: %v..%v / %v..%v", img.XMin, img.XMax, img.YMin, img.YMax)
	}
}

func TestAxesImageRejectsNormWithVMinOrVMax(t *testing.T) {
	vmin := 0.0
	ax := &Axes{}
	img := ax.Image([][]float64{{1, 2}}, ImageOptions{
		Norm: Normalize{VMin: 1, VMax: 2},
		VMin: &vmin,
	})
	if img != nil {
		t.Fatal("expected image construction to reject explicit norm with vmin")
	}
}

func TestImage2D_InterpolationField(t *testing.T) {
	bilinear := "bilinear"
	ax := &Axes{}
	img := ax.Image([][]float64{{0, 1}, {1, 0}}, ImageOptions{Interpolation: &bilinear})
	if img == nil {
		t.Fatal("Image returned nil")
	}
	if img.Interpolation != "bilinear" {
		t.Fatalf("Interpolation = %q, want %q", img.Interpolation, "bilinear")
	}
}

func TestImage2D_InterpolationDefaultsEmpty(t *testing.T) {
	ax := &Axes{}
	img := ax.Image([][]float64{{0, 1}}, ImageOptions{})
	if img == nil {
		t.Fatal("Image returned nil")
	}
	if img.Interpolation != "" {
		t.Fatalf("default Interpolation = %q, want empty", img.Interpolation)
	}
}

func TestImageRasterizeUsesConfiguredLogNorm(t *testing.T) {
	cmap := "gray"
	ax := &Axes{}
	img := ax.Image([][]float64{{1, 10, 100}}, ImageOptions{
		Colormap: &cmap,
		Norm:     LogNorm{VMin: 1, VMax: 100},
	})
	if img == nil {
		t.Fatal("expected image artist")
	}
	if img.Norm == nil || img.Norm.NormName() != "log" {
		t.Fatalf("image norm = %#v, want log norm", img.Norm)
	}

	rendered, ok := img.rasterize()
	if !ok {
		t.Fatal("expected rasterization to succeed")
	}
	rgbaData, ok := rendered.(*render.ImageData)
	if !ok {
		t.Fatalf("expected render.ImageData, got %T", rendered)
	}
	pix := rgbaData.RGBA()
	left := pix.RGBAAt(0, 0)
	mid := pix.RGBAAt(1, 0)
	right := pix.RGBAAt(2, 0)
	if left.R != 0 || left.G != 0 || left.B != 0 {
		t.Fatalf("log-norm low pixel = %+v, want black", left)
	}
	if mid.R < 126 || mid.R > 128 || mid.G < 126 || mid.G > 128 || mid.B < 126 || mid.B > 128 {
		t.Fatalf("log-norm middle pixel = %+v, want mid gray", mid)
	}
	if right.R != 255 || right.G != 255 || right.B != 255 {
		t.Fatalf("log-norm high pixel = %+v, want white", right)
	}
}

func TestImageRasterizeRejectsEmptyInputs(t *testing.T) {
	if _, ok := (&Image2D{}).rasterize(); ok {
		t.Fatal("expected empty image data to fail rasterization")
	}
	if _, ok := (&Image2D{Data: [][]float64{}}).rasterize(); ok {
		t.Fatal("expected zero-row image data to fail rasterization")
	}
	if _, ok := (&Image2D{Data: [][]float64{{}}}).rasterize(); ok {
		t.Fatal("expected zero-column image data to fail rasterization")
	}
}

func TestImageRasterizeUsesOriginAndSkipsNonFinite(t *testing.T) {
	data := [][]float64{
		{0, 1},
		{math.NaN(), 2},
	}
	img := &Image2D{
		Data:     data,
		Colormap: "gray",
		VMin:     0,
		VMax:     2,
		Alpha:    1,
		Origin:   ImageOriginLower,
	}

	rendered, ok := img.rasterize()
	if !ok {
		t.Fatal("expected rasterization to succeed")
	}

	rgbaData, ok := rendered.(*render.ImageData)
	if !ok {
		t.Fatalf("expected render.ImageData, got %T", rendered)
	}
	pix := rgbaData.RGBA()
	if pix == nil {
		t.Fatal("expected RGBA image from rasterizer")
	}

	// Origin lower flips row order: second data row is written at the top.
	bottomLeft := pix.RGBAAt(0, 1)
	if bottomLeft.R != 0 || bottomLeft.G != 0 || bottomLeft.B != 0 || bottomLeft.A != 255 {
		t.Fatalf("expected bottom-left pixel to be black, got %+v", bottomLeft)
	}

	bottomMid := pix.RGBAAt(1, 1)
	if bottomMid.R != 128 || bottomMid.G != 128 || bottomMid.B != 128 || bottomMid.A != 255 {
		t.Fatalf("expected bottom-middle pixel to be mid-gray, got %+v", bottomMid)
	}

	topMid := pix.RGBAAt(0, 0)
	if topMid.A != 0 {
		t.Fatalf("expected non-finite top-left value to be transparent, got alpha %d", topMid.A)
	}

	topRight := pix.RGBAAt(1, 0)
	if topRight.R != 255 || topRight.G != 255 || topRight.B != 255 || topRight.A != 255 {
		t.Fatalf("expected top-right white pixel, got %+v", topRight)
	}
}

func TestImageRasterizeLeavesPixelAlphaUnscaledWhenImageAlphaApplies(t *testing.T) {
	data := [][]float64{
		{0, 1},
	}
	img := &Image2D{
		Data:     data,
		Colormap: "gray",
		VMin:     0,
		VMax:     1,
		Alpha:    0.4,
		Origin:   ImageOriginLower,
	}

	rendered, ok := img.rasterize()
	if !ok {
		t.Fatal("expected rasterization to succeed")
	}

	rgbaData, ok := rendered.(*render.ImageData)
	if !ok {
		t.Fatalf("expected render.ImageData, got %T", rendered)
	}
	if got := rgbaData.Alpha(); got != 0.4 {
		t.Fatalf("image alpha = %v, want 0.4", got)
	}

	rgba := rgbaData.RGBA()
	black := rgba.RGBAAt(0, 0)
	if black.A != 255 {
		t.Fatalf("expected rasterized black pixel alpha to remain 255, got %d", black.A)
	}
}

func TestImageRasterizeBilinearUpsamplingInterpolatesScalarData(t *testing.T) {
	img := &Image2D{
		Data:          [][]float64{{0, 1}},
		Colormap:      "viridis",
		VMin:          0,
		VMax:          1,
		Alpha:         1,
		Origin:        ImageOriginUpper,
		Interpolation: "bilinear",
	}

	rendered, ok := img.rasterizeToSize(3, 1)
	if !ok {
		t.Fatal("expected rasterization to succeed")
	}
	rgbaData, ok := rendered.(*render.ImageData)
	if !ok {
		t.Fatalf("expected render.ImageData, got %T", rendered)
	}
	center := rgbaData.RGBA().RGBAAt(1, 0)
	if center.R != 32 || center.G != 144 || center.B != 140 {
		t.Fatalf("center pixel = %+v, want viridis scalar midpoint", center)
	}
}

func TestImage_DrawAngleZeroCallsImage(t *testing.T) {
	i := &Image2D{
		Data: [][]float64{
			{0, 1},
		},
		Alpha: 1,
		XMax:  1,
		YMax:  1,
	}
	r := &imageSpyRenderer{}
	ctx := createTestDrawContext()

	err := r.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	i.Draw(r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	if r.imageCalls != 1 {
		t.Fatalf("expected Image to be called once, got %d", r.imageCalls)
	}
	if r.transformedCalls != 0 {
		t.Fatalf("expected ImageTransformed not to be called, got %d", r.transformedCalls)
	}
}

func TestImage_DrawRotatedScalarUsesPrefilteredImage(t *testing.T) {
	angle := 30.0
	i := &Image2D{
		Data:     [][]float64{{0, 1}},
		AngleDeg: angle,
		Alpha:    1,
		XMax:     1,
		YMax:     1,
	}
	r := &imageSpyRenderer{}
	ctx := createTestDrawContext()

	err := r.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	i.Draw(r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	if r.imageCalls != 1 {
		t.Fatalf("expected prefiltered Image to be called once, got %d", r.imageCalls)
	}
	if r.transformedCalls != 0 {
		t.Fatalf("expected scalar rotation to avoid backend ImageTransformed, got %d", r.transformedCalls)
	}
	if r.lastImageWidth <= 2 || r.lastImageHeight <= 1 {
		t.Fatalf("prefiltered image size = %dx%d, want transformed display-sized raster", r.lastImageWidth, r.lastImageHeight)
	}
}

func TestImageTransformPositiveAngleUsesDataSpaceOrientation(t *testing.T) {
	raster := render.NewImageData(image.NewRGBA(image.Rect(0, 0, 10, 10)))
	tr := imageTransform(
		geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 10, Y: 10}},
		raster,
		geom.Pt{X: 5, Y: 5},
		math.Pi/2,
	)

	// Display space is y-up: a +90 deg (CCW) rotation maps right-center to
	// top-center, which is y=10 (the top edge) rather than y=0.
	got := tr.Apply(geom.Pt{X: 10, Y: 5})
	if !almostEqualFloat(got.X, 5) || !almostEqualFloat(got.Y, 10) {
		t.Fatalf("positive rotation maps right-center to %+v, want top-center", got)
	}
}

func TestImage_DrawRotatedFallsBackToImage(t *testing.T) {
	angle := 30.0
	i := &Image2D{
		Data:     [][]float64{{0, 1}},
		AngleDeg: angle,
		Alpha:    1,
		XMax:     1,
		YMax:     1,
	}
	r := &imageSpyNoTransformRenderer{}
	ctx := createTestDrawContext()

	err := r.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	i.Draw(r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	if r.imageCalls != 1 {
		t.Fatalf("expected fallback to Image, got %d", r.imageCalls)
	}
}

func TestImage2D_DrawPrefilteredBilinearUsesNearestRendererInterpolation(t *testing.T) {
	bilinear := "bilinear"
	ax := &Axes{}
	img := ax.Image([][]float64{{0, 1}, {1, 0}}, ImageOptions{Interpolation: &bilinear})
	if img == nil {
		t.Fatal("Image returned nil")
	}
	img.XMax = 1
	img.YMax = 1
	img.Alpha = 1

	rec := &imageSpyRenderer{}
	ctx := createTestDrawContext()

	if err := rec.Begin(geom.Rect{}); err != nil {
		t.Fatalf("begin: %v", err)
	}
	img.Draw(rec, ctx)
	if err := rec.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	if rec.imageCalls != 1 {
		t.Fatalf("imageCalls = %d, want 1", rec.imageCalls)
	}
	if rec.lastImageWidth <= 2 || rec.lastImageHeight <= 2 {
		t.Fatalf("prefiltered image size = %dx%d, want destination-sized upsample", rec.lastImageWidth, rec.lastImageHeight)
	}
	if rec.lastInterpolation != "nearest" {
		t.Fatalf("lastInterpolation = %q, want nearest for prefiltered bilinear raster", rec.lastInterpolation)
	}
}

func TestImage2D_DrawHighUpsampleBicubicUsesScalarDataStage(t *testing.T) {
	bicubic := "bicubic"
	ax := &Axes{}
	img := ax.Image([][]float64{{0, 1}, {1, 0}}, ImageOptions{Interpolation: &bicubic})
	if img == nil {
		t.Fatal("Image returned nil")
	}
	img.XMax = 1
	img.YMax = 1
	img.Alpha = 1

	rec := &imageSpyRenderer{}
	ctx := createTestDrawContext()

	if err := rec.Begin(geom.Rect{}); err != nil {
		t.Fatalf("begin: %v", err)
	}
	img.Draw(rec, ctx)
	if err := rec.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	if rec.imageCalls != 1 {
		t.Fatalf("imageCalls = %d, want 1", rec.imageCalls)
	}
	if rec.lastImageWidth <= 2 || rec.lastImageHeight <= 2 {
		t.Fatalf("prefiltered image size = %dx%d, want destination-sized upsample", rec.lastImageWidth, rec.lastImageHeight)
	}
	if rec.lastInterpolation != "nearest" {
		t.Fatalf("lastInterpolation = %q, want nearest after scalar-stage bicubic resampling", rec.lastInterpolation)
	}
}

func TestImage2D_DrawCeilsRasterSizeAndKeepsMatplotlibAnchor(t *testing.T) {
	img := &Image2D{
		Data: [][]float64{
			{0, 1},
			{1, 0},
		},
		Colormap:      "gray",
		VMin:          0,
		VMax:          1,
		Alpha:         1,
		XMax:          1,
		YMax:          1,
		Interpolation: "nearest",
	}
	rec := &imageSpyRenderer{}
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale: transform.NewLinear(0, 1),
			YScale: transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Affine{
				A: 273.6,
				D: -273.6,
				E: 183.2,
				F: 309.6,
			}),
		},
	}

	if err := rec.Begin(geom.Rect{}); err != nil {
		t.Fatalf("begin: %v", err)
	}
	img.Draw(rec, ctx)
	if err := rec.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	if rec.lastImageWidth != 274 || rec.lastImageHeight != 274 {
		t.Fatalf("prefiltered image size = %dx%d, want 274x274", rec.lastImageWidth, rec.lastImageHeight)
	}
	// Display space is y-up: Matplotlib's _make_image rounds the bbox edges
	// (image.py:415-418) to y0=ceil(35.5-eps)=36, y1=ceil(309.1)=310, which the
	// bottom-anchored ceil in matplotlibImageDrawRect reproduces.
	want := geom.Rect{
		Min: geom.Pt{X: 183.2, Y: 36},
		Max: geom.Pt{X: 457.2, Y: 310},
	}
	if !rectsApprox(rec.lastDst, want, 1e-9) {
		t.Fatalf("image destination = %+v, want %+v", rec.lastDst, want)
	}
}

func TestImage2D_DrawRotatedBilinearUsesScalarDataStage(t *testing.T) {
	bilinear := "bilinear"
	angle := 15.0
	ax := &Axes{}
	img := ax.Image([][]float64{{0, 1}, {1, 0}}, ImageOptions{
		Interpolation: &bilinear,
		Angle:         &angle,
	})
	if img == nil {
		t.Fatal("Image returned nil")
	}
	img.XMax = 1
	img.YMax = 1
	img.Alpha = 1

	rec := &imageSpyRenderer{}
	ctx := createTestDrawContext()

	if err := rec.Begin(geom.Rect{}); err != nil {
		t.Fatalf("begin: %v", err)
	}
	img.Draw(rec, ctx)
	if err := rec.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	if rec.imageCalls != 1 {
		t.Fatalf("imageCalls = %d, want 1", rec.imageCalls)
	}
	if rec.transformedCalls != 0 {
		t.Fatalf("transformedCalls = %d, want 0", rec.transformedCalls)
	}
	if rec.lastInterpolation != "nearest" {
		t.Fatalf("lastInterpolation = %q, want nearest after scalar-stage transformed resampling", rec.lastInterpolation)
	}
}

func TestImage2D_DrawDefaultInterpolationEmpty(t *testing.T) {
	ax := &Axes{}
	img := ax.Image([][]float64{{0, 1}}, ImageOptions{})
	if img == nil {
		t.Fatal("Image returned nil")
	}
	img.XMax = 1
	img.YMax = 1
	img.Alpha = 1

	rec := &imageSpyRenderer{}
	ctx := createTestDrawContext()

	if err := rec.Begin(geom.Rect{}); err != nil {
		t.Fatalf("begin: %v", err)
	}
	img.Draw(rec, ctx)
	if err := rec.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	if rec.imageCalls != 1 {
		t.Fatalf("imageCalls = %d, want 1", rec.imageCalls)
	}
	if rec.lastInterpolation != "" {
		t.Fatalf("lastInterpolation = %q, want empty", rec.lastInterpolation)
	}
}

func TestImage_DrawNilRendererDoesNothing(t *testing.T) {
	i := &Image2D{
		Data: [][]float64{{1}},
	}
	ctx := createTestDrawContext()
	i.Draw(nil, ctx)
}

type imageSpyRenderer struct {
	imageCalls        int
	transformedCalls  int
	lastDst           geom.Rect
	lastTransform     geom.Affine
	lastInterpolation string
	lastImageWidth    int
	lastImageHeight   int
}

func (r *imageSpyRenderer) Begin(geom.Rect) error { return nil }
func (r *imageSpyRenderer) End() error            { return nil }
func (r *imageSpyRenderer) Save()                 {}
func (r *imageSpyRenderer) Restore()              {}
func (r *imageSpyRenderer) ClipRect(geom.Rect)    {}
func (r *imageSpyRenderer) ClipPath(geom.Path)    {}
func (r *imageSpyRenderer) Path(geom.Path, *render.Paint) {
}

func (r *imageSpyRenderer) Image(img render.Image, dst geom.Rect) {
	r.imageCalls++
	r.lastDst = dst
	if img != nil {
		r.lastInterpolation = img.Interpolation()
		r.lastImageWidth, r.lastImageHeight = img.Size()
	}
}

func (r *imageSpyRenderer) ImageTransformed(img render.Image, dst geom.Rect, t geom.Affine) {
	r.transformedCalls++
	r.lastDst = dst
	r.lastTransform = t
	if img != nil {
		r.lastInterpolation = img.Interpolation()
		r.lastImageWidth, r.lastImageHeight = img.Size()
	}
}
func (r *imageSpyRenderer) GlyphRun(render.GlyphRun, render.Color) {}
func (r *imageSpyRenderer) MeasureText(string, float64, string) render.TextMetrics {
	return render.TextMetrics{}
}

type imageSpyNoTransformRenderer struct {
	imageCalls int
}

func (r *imageSpyNoTransformRenderer) Begin(geom.Rect) error { return nil }
func (r *imageSpyNoTransformRenderer) End() error            { return nil }
func (r *imageSpyNoTransformRenderer) Save()                 {}
func (r *imageSpyNoTransformRenderer) Restore()              {}
func (r *imageSpyNoTransformRenderer) ClipRect(geom.Rect)    {}
func (r *imageSpyNoTransformRenderer) ClipPath(geom.Path)    {}
func (r *imageSpyNoTransformRenderer) Path(geom.Path, *render.Paint) {
}

func (r *imageSpyNoTransformRenderer) Image(_ render.Image, dst geom.Rect) { r.imageCalls++; _ = dst }

func (r *imageSpyNoTransformRenderer) GlyphRun(render.GlyphRun, render.Color) {
}

func (r *imageSpyNoTransformRenderer) MeasureText(string, float64, string) render.TextMetrics {
	return render.TextMetrics{}
}
