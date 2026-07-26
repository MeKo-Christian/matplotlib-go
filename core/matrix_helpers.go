package core

import (
	"fmt"
	"image"
	"math"
	"strconv"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

// MatShowOptions configures Axes.MatShow.
type MatShowOptions struct {
	Colormap     optional.Value[string]
	Norm         ScalarNormalizer
	VMin         optional.Value[float64]
	VMax         optional.Value[float64]
	Alpha        optional.Value[float64]
	Aspect       ImageAspect
	IntegerTicks optional.Value[bool]
	Label        string
	// Interpolation selects the image resampling filter. Empty defers to the
	// renderer default.
	Interpolation optional.Value[string]
}

// ImShowOptions configures Axes.ImShow.
//
// Mirrors matplotlib.axes.Axes.imshow keyword arguments
// (third_party/matplotlib/lib/matplotlib/axes/_axes.py:6149).
type ImShowOptions struct {
	// Colormap names the colormap applied to the scalar data. Unset uses the
	// image.cmap rc default.
	Colormap optional.Value[string]
	// Norm maps data values into [0,1] before the colormap is applied. A nil
	// normalizer scales linearly between the data limits.
	Norm ScalarNormalizer
	// VMin and VMax pin the color limits. Unset derives them from the data.
	VMin optional.Value[float64]
	VMax optional.Value[float64]
	// Alpha is the image opacity in [0,1]. Unset is fully opaque.
	Alpha optional.Value[float64]
	// Aspect sets the axes aspect ("equal", "auto", or a numeric ratio).
	// Unset uses the image.aspect rc default; an explicit "" leaves the
	// current axes aspect alone.
	Aspect optional.Value[ImageAspect]
	// Origin places the [0,0] index at the upper or lower corner. Unset uses
	// the image.origin rc default.
	Origin optional.Value[ImageOrigin]
	// Label is the legend entry. Empty adds none.
	Label string
	// Extent overrides the centered-pixel default with explicit
	// (left, right, bottom, top) data coordinates.
	Extent optional.Value[[4]float64]
	// Interpolation selects the resampling filter (e.g. "nearest",
	// "bilinear", "bicubic"). Unset uses Matplotlib's rc default
	// "antialiased"; an explicit "" defers to the renderer default.
	Interpolation optional.Value[string]
}

// SpyOptions configures Axes.Spy.
type SpyOptions struct {
	Precision  float64
	UseImage   optional.Value[bool]
	Marker     optional.Value[MarkerType]
	MarkerSize float64
	Color      optional.Value[render.Color]
	Alpha      optional.Value[float64]
	Aspect     ImageAspect
	Label      string
}

// AnnotatedHeatmapOptions configures Axes.AnnotatedHeatmap.
type AnnotatedHeatmapOptions struct {
	MatShowOptions
	Format        string
	FontSize      float64
	TextColor     render.Color
	TextColorHigh render.Color
	Threshold     optional.Value[float64]
	SkipNaN       bool
	NaNText       string
}

// SpyResult groups the artists and coordinates produced by Axes.Spy.
type SpyResult struct {
	Image   *Image2D
	Markers *PathCollection
	Indices []geom.Pt
}

// AnnotatedHeatmapResult stores the image plus per-cell annotations.
type AnnotatedHeatmapResult struct {
	Image  *Image2D
	Labels []*Text
}

// MatShow renders a matrix with centered integer ticks and an equal aspect.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) MatShow(data [][]float64, opt MatShowOptions) *Image2D {
	rows, cols, ok := finiteMatrixSize(data)
	if !ok {
		return nil
	}

	cfg := MatShowOptions{
		Aspect: "equal",
	}
	if opt.Colormap.IsSet() {
		cfg.Colormap = opt.Colormap
	}
	if opt.Norm != nil {
		cfg.Norm = opt.Norm
	}
	if opt.VMin.IsSet() {
		cfg.VMin = opt.VMin
	}
	if opt.VMax.IsSet() {
		cfg.VMax = opt.VMax
	}
	if opt.Alpha.IsSet() {
		cfg.Alpha = opt.Alpha
	}
	if opt.Aspect != "" {
		cfg.Aspect = opt.Aspect
	}
	if opt.IntegerTicks.IsSet() {
		cfg.IntegerTicks = opt.IntegerTicks
	}
	if opt.Label != "" {
		cfg.Label = opt.Label
	}
	cfg.Interpolation = opt.Interpolation

	xMin := -0.5
	xMax := float64(cols) - 0.5
	yMin := -0.5
	yMax := float64(rows) - 0.5
	img := a.Image(data, ImageOptions{
		Colormap:      cfg.Colormap,
		Norm:          cfg.Norm,
		VMin:          cfg.VMin,
		VMax:          cfg.VMax,
		Alpha:         cfg.Alpha,
		XMin:          optional.Of(xMin),
		XMax:          optional.Of(xMax),
		YMin:          optional.Of(yMin),
		YMax:          optional.Of(yMax),
		Origin:        ImageOriginUpper,
		Label:         cfg.Label,
		Interpolation: cfg.Interpolation,
	})
	if img == nil {
		return nil
	}

	if cfg.Aspect != "" {
		_ = a.SetAspect(string(cfg.Aspect))
	}
	a.SetXLim(xMin, xMax)
	a.SetYLim(yMin, yMax)
	if !a.YInverted() {
		a.InvertY()
	}
	applyMatrixAxisPresentation(a, true)
	if boolValue(cfg.IntegerTicks.Ptr(), true) {
		applyMatrixTicks(a, rows, cols)
	}
	return img
}

// ImShow renders a matrix with Matplotlib imshow-style image extents,
// centered pixel coordinates, equal aspect, and the primary x-axis at bottom.
//
//nolint:gocritic // ImShowOptions is an immutable snapshot of the caller's options.
func (a *Axes) ImShow(data [][]float64, opt ImShowOptions) *Image2D {
	rows, cols, ok := finiteMatrixSize(data)
	if !ok {
		return nil
	}

	rc := a.resolvedRC()
	aspect := opt.Aspect.Or(imshowAspectDefault(&rc))
	origin := opt.Origin.Or(imageOriginFromRC(&rc))
	interpolation := opt.Interpolation.Or("antialiased")

	xMin := -0.5
	xMax := float64(cols) - 0.5
	yMin := -0.5
	yMax := float64(rows) - 0.5
	extent, extentGiven := opt.Extent.Get()
	if extentGiven {
		xMin, xMax, yMin, yMax = extent[0], extent[1], extent[2], extent[3]
	}
	img := a.Image(data, ImageOptions{
		Colormap:      opt.Colormap,
		Norm:          opt.Norm,
		VMin:          opt.VMin,
		VMax:          opt.VMax,
		Alpha:         opt.Alpha,
		XMin:          optional.Of(xMin),
		XMax:          optional.Of(xMax),
		YMin:          optional.Of(yMin),
		YMax:          optional.Of(yMax),
		Origin:        origin,
		Label:         opt.Label,
		Interpolation: optional.Of(interpolation),
	})
	if img == nil {
		return nil
	}

	a.finishImshow(xMin, xMax, yMin, yMax, aspect, origin, extentGiven)
	return img
}

// finishImshow applies the aspect, axis limits, and origin-driven y-inversion
// shared by ImShow, ImShowRGB, and ImShowImage. When extentGiven is true the
// caller supplied an explicit extent, so the automatic origin y-flip is skipped.
func (a *Axes) finishImshow(xMin, xMax, yMin, yMax float64, aspect ImageAspect, origin ImageOrigin, extentGiven bool) {
	if aspect != "" {
		// image.aspect (and the Aspect option) also admit a numeric ratio.
		if ratio, err := strconv.ParseFloat(string(aspect), 64); err == nil {
			_ = a.SetAspect("ratio", ratio)
		} else {
			_ = a.SetAspect(string(aspect))
		}
	}
	a.SetXLim(xMin, xMax)
	a.SetYLim(yMin, yMax)
	if extentGiven {
		return
	}
	if origin == ImageOriginUpper && !a.YInverted() {
		a.InvertY()
	}
	if origin == ImageOriginLower && a.YInverted() {
		a.InvertY()
	}
}

// ImShowRGBOptions configures Axes.ImShowRGB and Axes.ImShowImage.
//
// Unlike ImShowOptions it has no Colormap/Norm/VMin/VMax: the input is already
// colored, so matplotlib bypasses the scalar-mapping pipeline for RGB/RGBA
// images (third_party/matplotlib/lib/matplotlib/image.py).
type ImShowRGBOptions struct {
	// Alpha multiplies the image's per-pixel alpha in [0,1].
	Alpha optional.Value[float64]
	// Aspect sets the axes aspect ("equal", "auto", or a numeric ratio).
	// Empty uses the image.aspect rc default.
	Aspect ImageAspect
	// Origin places the [0,0] index at the upper or lower corner.
	// ImageOriginUpper is the zero value and doubles as "unset": the
	// image.origin rc default applies (see ImShowOptions.Origin).
	Origin ImageOrigin
	// Extent overrides the centered-pixel default with explicit
	// (left, right, bottom, top) data coordinates.
	Extent optional.Value[[4]float64]
	// Interpolation selects the resampling filter. Nil uses Matplotlib's rc
	// default "antialiased"; a pointer to "" defers to the renderer default.
	Interpolation optional.Value[string]
	Label         string
}

// resolveImShowRGBOptions merges a caller's options over the rc-derived
// defaults. supplied distinguishes an omitted option set from a zero one.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) resolveImShowRGBOptions(opt ImShowRGBOptions) ImShowRGBOptions {
	rc := a.resolvedRC()
	defaultInterpolation := "antialiased"
	cfg := ImShowRGBOptions{
		Aspect:        imshowAspectDefault(&rc),
		Origin:        imageOriginFromRC(&rc),
		Interpolation: optional.Of(defaultInterpolation),
	}
	if opt.Alpha.IsSet() {
		cfg.Alpha = opt.Alpha
	}
	if opt.Aspect != "" {
		cfg.Aspect = opt.Aspect
	}
	if opt.Origin != ImageOriginUpper {
		cfg.Origin = opt.Origin
	}
	cfg.Extent = opt.Extent
	if opt.Interpolation.IsSet() {
		cfg.Interpolation = opt.Interpolation
	}
	if opt.Label != "" {
		cfg.Label = opt.Label
	}
	return cfg
}

// ImShowRGB renders a pre-colored (M,N,3) or (M,N,4) float array (channel
// values in [0,1]) as a true-color image, bypassing the colormap+norm path.
// Out-of-range channel values are clipped to [0,1] (with a warning), mirroring
// matplotlib's imshow. A (M,N,1) array is squeezed and routed to the scalar
// ImShow colormap path.
//
// Rejected input leaves the axes unchanged.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) ImShowRGB(data [][][]float64, opt ImShowRGBOptions) (*Image2D, error) {
	if a == nil {
		return nil, fmt.Errorf("imshow rgb axes cannot be nil")
	}
	cfg := a.resolveImShowRGBOptions(opt)

	rgba, kind, err := normalizeRGBArray(data)
	if err != nil {
		return nil, fmt.Errorf("imshow rgb: %w", err)
	}
	if kind == rgbArrayScalar {
		// cfg already carries the rc-resolved values, so each one is passed as
		// an explicit request rather than left for ImShow to resolve again.
		img := a.ImShow(squeezeScalarArray(data), ImShowOptions{
			Alpha:         cfg.Alpha,
			Aspect:        optional.Of(cfg.Aspect),
			Origin:        optional.Of(cfg.Origin),
			Extent:        cfg.Extent,
			Interpolation: cfg.Interpolation,
			Label:         cfg.Label,
		})
		if img == nil {
			return nil, fmt.Errorf("imshow rgb: single-channel data is not a finite, rectangular matrix")
		}
		return img, nil
	}
	return a.imshowRGBA(rgba, cfg), nil
}

// ImShowImage renders a native Go image (e.g. the output of core.ImRead) as a
// true-color image, bypassing the colormap+norm path. The image's row 0 is
// placed at the top for the default ImageOriginUpper.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) ImShowImage(img image.Image, opt ImShowRGBOptions) *Image2D {
	if a == nil || img == nil {
		return nil
	}
	rgba := toRGBA(img)
	if rgba == nil || rgba.Bounds().Dx() == 0 || rgba.Bounds().Dy() == 0 {
		return nil
	}
	return a.imshowRGBA(rgba, a.resolveImShowRGBOptions(opt))
}

// imshowRGBA is the shared tail for ImShowRGB/ImShowImage: it sets the extent,
// builds the true-color Image2D, and applies aspect/limits/origin.
func (a *Axes) imshowRGBA(rgba *image.RGBA, cfg ImShowRGBOptions) *Image2D {
	b := rgba.Bounds()
	cols := b.Dx()
	rows := b.Dy()
	if cols == 0 || rows == 0 {
		return nil
	}

	xMin := -0.5
	xMax := float64(cols) - 0.5
	yMin := -0.5
	yMax := float64(rows) - 0.5
	if extent, ok := cfg.Extent.Get(); ok {
		xMin = extent[0]
		xMax = extent[1]
		yMin = extent[2]
		yMax = extent[3]
	}

	img := a.imageRGBA(rgba, ImageOptions{
		Alpha:         cfg.Alpha,
		XMin:          optional.Of(xMin),
		XMax:          optional.Of(xMax),
		YMin:          optional.Of(yMin),
		YMax:          optional.Of(yMax),
		Origin:        cfg.Origin,
		Label:         cfg.Label,
		Interpolation: cfg.Interpolation,
	})
	if img == nil {
		return nil
	}
	a.finishImshow(xMin, xMax, yMin, yMax, cfg.Aspect, cfg.Origin, cfg.Extent.IsSet())
	return img
}

// squeezeScalarArray collapses an (M,N,1) array to (M,N) for the scalar path.
func squeezeScalarArray(data [][][]float64) [][]float64 {
	out := make([][]float64, len(data))
	for y, row := range data {
		flat := make([]float64, len(row))
		for x, px := range row {
			if len(px) > 0 {
				flat[x] = px[0]
			}
		}
		out[y] = flat
	}
	return out
}

// Spy visualizes the sparsity pattern of a matrix.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) Spy(data [][]float64, opt SpyOptions) *SpyResult {
	rows, cols, ok := finiteMatrixSize(data)
	if !ok {
		return nil
	}

	cfg := SpyOptions{
		Aspect: "equal",
	}
	cfg.Precision = opt.Precision
	cfg.UseImage = opt.UseImage
	cfg.Marker = opt.Marker
	if opt.MarkerSize > 0 {
		cfg.MarkerSize = opt.MarkerSize
	}
	if opt.Color.IsSet() {
		cfg.Color = opt.Color
	}
	if opt.Alpha.IsSet() {
		cfg.Alpha = opt.Alpha
	}
	if opt.Aspect != "" {
		cfg.Aspect = opt.Aspect
	}
	if opt.Label != "" {
		cfg.Label = opt.Label
	}
	indices := make([]geom.Pt, 0)
	mask := make([][]float64, rows)
	for row := 0; row < rows; row++ {
		mask[row] = make([]float64, cols)
		for col := 0; col < cols; col++ {
			value := data[row][col]
			if !isFinite(value) || math.Abs(value) <= cfg.Precision {
				continue
			}
			mask[row][col] = 1
			indices = append(indices, geom.Pt{X: float64(col), Y: float64(row)})
		}
	}

	result := &SpyResult{Indices: indices}
	useImage := cfg.UseImage.Or(true)
	if cfg.Marker.IsSet() || cfg.MarkerSize > 0 {
		useImage = false
	}
	if useImage {
		cmap := "binary"
		nearest := "nearest"
		vMin := 0.0
		vMax := 1.0
		result.Image = a.MatShow(mask, MatShowOptions{
			Colormap:     optional.Of(cmap),
			VMin:         optional.Of(vMin),
			VMax:         optional.Of(vMax),
			Alpha:        cfg.Alpha,
			Aspect:       cfg.Aspect,
			IntegerTicks: optional.Of(true),
			Label:        cfg.Label,
		})
		result.Image.Interpolation = nearest
		return result
	}

	if cfg.MarkerSize <= 0 {
		cfg.MarkerSize = 10
	}
	marker := markerValue(cfg.Marker.Ptr(), MarkerSquare)
	color := render.Color{A: 1}
	if v, ok := cfg.Color.Get(); ok {
		color = v
	}
	alpha := 1.0
	if cfg.Alpha.IsSet() {
		alpha = clampOneToOne(cfg.Alpha.OrZero())
	}
	path := (&Scatter2D{Marker: marker}).markerPrototypePath()
	lineOnly := markerLineOnly(NewMarkerStyle(marker))
	markerSizePx := math.Ceil(cfg.MarkerSize * matrixMarkerDPI(a) / 72.0)
	markerEdgeWidth := 1.0 // points; converted at the collection Paint sink
	pc := &PathCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  cfg.Label,
			Alpha:  alpha,
		},
		Path:          path,
		Offsets:       append([]geom.Pt(nil), indices...),
		Size:          markerSizePx,
		PathInDisplay: true,
		FaceColor:     color,
		EdgeColor:     color,
		EdgeWidth:     markerEdgeWidth,
		LineOnly:      lineOnly,
	}
	a.AddCollection(pc)
	result.Markers = pc

	xMin := -0.5
	xMax := float64(cols) - 0.5
	yMin := -0.5
	yMax := float64(rows) - 0.5
	if cfg.Aspect != "" {
		_ = a.SetAspect(string(cfg.Aspect))
	}
	a.SetXLim(xMin, xMax)
	a.SetYLim(yMin, yMax)
	if !a.YInverted() {
		a.InvertY()
	}
	applyMatrixAxisPresentation(a, true)
	applyMatrixTicks(a, rows, cols)
	return result
}

func matrixMarkerDPI(a *Axes) float64 {
	if a != nil {
		if dpi := a.resolvedRC().DPI; dpi > 0 {
			return dpi
		}
	}
	return 72
}

// AnnotatedHeatmap renders a matrix display plus a centered value label in each cell.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) AnnotatedHeatmap(data [][]float64, opt AnnotatedHeatmapOptions) *AnnotatedHeatmapResult {
	rows, cols, ok := finiteMatrixSize(data)
	if !ok {
		return nil
	}

	cfg := opt
	if cfg.Aspect == "" {
		cfg.Aspect = "equal"
	}
	if cfg.Format == "" {
		cfg.Format = "%.3g"
	}
	if cfg.FontSize <= 0 {
		cfg.FontSize = 11
	}
	if cfg.TextColor.A == 0 {
		cfg.TextColor = render.Color{R: 0.1, G: 0.1, B: 0.1, A: 1}
	}
	if cfg.TextColorHigh.A == 0 {
		cfg.TextColorHigh = render.Color{R: 1, G: 1, B: 1, A: 1}
	}

	img := a.MatShow(data, cfg.MatShowOptions)
	if img == nil {
		return nil
	}

	mapping := img.ScalarMap().Resolved()
	threshold := mapping.VMin + 0.5*(mapping.VMax-mapping.VMin)
	if v, ok := cfg.Threshold.Get(); ok {
		threshold = v
	}

	labels := make([]*Text, 0, rows*cols)
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			value := data[row][col]
			if !isFinite(value) && cfg.SkipNaN {
				continue
			}
			text := formatAnnotatedHeatmapValue(value, cfg.Format, cfg.NaNText)
			color := cfg.TextColor
			if isFinite(value) && value >= threshold {
				color = cfg.TextColorHigh
			}
			label := a.Text(float64(col), float64(row), text, TextOptions{
				FontSize: cfg.FontSize,
				Color:    color,
				HAlign:   TextAlignCenter,
				VAlign:   TextVAlignMiddle,
			})
			labels = append(labels, label)
		}
	}

	return &AnnotatedHeatmapResult{
		Image:  img,
		Labels: labels,
	}
}

func applyMatrixAxisPresentation(a *Axes, showBottomTicks bool) {
	if a == nil {
		return
	}
	if a.XAxis != nil {
		a.XAxis.ShowTicks = showBottomTicks
		a.XAxis.ShowLabels = false
	}
	if top := a.TopAxis(); top != nil {
		top.ShowSpine = true
		top.ShowTicks = true
		top.ShowLabels = true
	}
}

func applyMatrixTicks(a *Axes, rows, cols int) {
	if a == nil {
		return
	}
	xLocator := integerMatrixLocator(cols)
	yLocator := integerMatrixLocator(rows)
	for _, axis := range []*Axis{a.XAxis, a.XAxisTop} {
		if axis == nil {
			continue
		}
		axis.Locator = xLocator
		axis.Formatter = ticker.ScalarFormatter{Prec: 0}
	}
	for _, axis := range []*Axis{a.YAxis, a.YAxisRight} {
		if axis == nil {
			continue
		}
		axis.Locator = yLocator
		axis.Formatter = ticker.ScalarFormatter{Prec: 0}
	}
}

func integerMatrixLocator(count int) ticker.Locator {
	if count <= 0 {
		return ticker.NullLocator{}
	}
	return ticker.MaxNLocator{N: 9, Steps: []float64{1, 2, 5, 10}, Integer: true}
}

func formatAnnotatedHeatmapValue(value float64, pattern, nanText string) string {
	if !isFinite(value) {
		if nanText != "" {
			return nanText
		}
		return "NaN"
	}
	if pattern == "" {
		return ticker.ScalarFormatter{Prec: 3}.Format(value)
	}
	return ticker.FormatStrFormatter{Pattern: pattern}.Format(value)
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func boolPtr(value bool) *bool {
	return &value
}

func markerValue(value *MarkerType, fallback MarkerType) MarkerType {
	if value == nil {
		return fallback
	}
	return *value
}
