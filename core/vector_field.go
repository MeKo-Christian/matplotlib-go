package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	quiverAnglesUV = "uv"
	quiverAnglesXY = "xy"

	vectorPivotTail   = "tail"
	vectorPivotMiddle = "middle"
	vectorPivotTip    = "tip"

	streamDirectionForward  = "forward"
	streamDirectionBackward = "backward"
	streamDirectionBoth     = "both"
)

// QuiverOptions configures vector-arrow plots.
type QuiverOptions struct {
	Color          *render.Color
	Colors         []render.Color
	C              []float64
	CGrid          [][]float64
	Colormap       *string
	Norm           ScalarNormalizer
	VMin           *float64
	VMax           *float64
	Alpha          *float64
	EdgeColor      *render.Color
	EdgeWidth      *float64
	Pivot          string
	Angles         string
	AngleValues    []float64
	Scale          *float64
	ScaleUnits     string
	Units          string
	Width          *float64
	HeadWidth      *float64
	HeadLength     *float64
	HeadAxisLength *float64
	MinShaft       *float64
	MinLength      *float64
	Label          string
	ZOrder         *float64
}

// Quiver renders repeated vector arrows anchored at data points.
type Quiver struct {
	Anchors []geom.Pt
	U       []float64
	V       []float64

	Colors       []render.Color
	Color        render.Color
	ScalarColors []float64
	EdgeColor    render.Color
	EdgeWidth    float64
	Alpha        float64

	Pivot          string
	Angles         string
	AngleValues    []float64
	Scale          float64
	ScaleSet       bool
	ScaleUnits     string
	Units          string
	Width          float64
	HeadWidth      float64
	HeadLength     float64
	HeadAxisLength float64
	MinShaft       float64
	MinLength      float64

	Label    string
	Colormap string
	Norm     ScalarNormalizer
	VMin     float64
	VMax     float64
	z        float64

	// Internal-only draw overrides used by composite artists such as streamplot.
	forceLengthPx float64
}

// QuiverKeyOptions configures a labeled quiver scale key.
type QuiverKeyOptions struct {
	Coords     CoordinateSpec
	Angle      float64
	LabelPos   string
	LabelSep   float64
	Color      render.Color
	LabelColor render.Color
	FontSize   float64
	ZOrder     *float64
}

// QuiverKey renders a labeled reference arrow using an existing quiver style.
type QuiverKey struct {
	Quiver     *Quiver
	Position   geom.Pt
	U          float64
	Label      string
	Coords     CoordinateSpec
	Angle      float64
	LabelPos   string
	LabelSep   float64
	Color      render.Color
	LabelColor render.Color
	FontSize   float64
	z          float64
}

// BarbIncrements configures the value represented by each barb segment.
type BarbIncrements struct {
	Half float64
	Full float64
	Flag float64
}

// BarbSizes configures the geometry ratios of a barb glyph.
type BarbSizes struct {
	Spacing   float64
	Height    float64
	Width     float64
	EmptyBarb float64
}

// BarbsOptions configures wind-barb style vector plots.
type BarbsOptions struct {
	Color      *render.Color
	Colors     []render.Color
	C          []float64
	CGrid      [][]float64
	Colormap   *string
	Norm       ScalarNormalizer
	VMin       *float64
	VMax       *float64
	Alpha      *float64
	BarbColor  *render.Color
	FlagColor  *render.Color
	LineWidth  *float64
	Pivot      string
	Length     *float64
	Units      string
	Sizes      *BarbSizes
	Increments *BarbIncrements
	FillEmpty  *bool
	Rounding   *bool
	FlipBarb   *bool
	Flip       []bool
	Label      string
	ZOrder     *float64
}

// Barbs renders meteorological barb glyphs anchored at data points.
type Barbs struct {
	Anchors []geom.Pt
	U       []float64
	V       []float64

	Colors       []render.Color
	Color        render.Color
	ScalarColors []float64
	BarbColor    render.Color
	FlagColor    render.Color
	LineWidth    float64
	Alpha        float64

	Pivot      string
	Length     float64
	Units      string
	Sizes      BarbSizes
	Increments BarbIncrements
	FillEmpty  bool
	Rounding   bool
	Flip       []bool

	Label    string
	Colormap string
	Norm     ScalarNormalizer
	VMin     float64
	VMax     float64
	z        float64
}

// StreamplotOptions configures streamline generation and styling.
type StreamplotOptions struct {
	Density              float64
	DensityX             float64
	DensityY             float64
	CGrid                [][]float64
	Colormap             *string
	Norm                 ScalarNormalizer
	VMin                 *float64
	VMax                 *float64
	StartPoints          []geom.Pt
	MinLength            *float64
	MaxLength            *float64
	IntegrationDirection string
	BrokenStreamlines    *bool
	ArrowSize            *float64
	ArrowCount           *int
	LineWidth            *float64
	Color                *render.Color
	ArrowColor           *render.Color
	Label                string
	ZOrder               *float64
}

// StreamplotSet owns the line and arrow artists produced by Axes.Streamplot.
type StreamplotSet struct {
	Lines  *LineCollection
	Arrows *Quiver
	Label  string
	z      float64
}

type vectorRenderState struct {
	widthPx float64
	scale   float64
}

type streamplotGrid struct {
	x []float64
	y []float64
	u [][]float64
	v [][]float64
}

type streamplotMask struct {
	nx    int
	ny    int
	used  []bool
	xmin  float64
	xspan float64
	ymin  float64
	yspan float64
}

type streamTrajectory struct {
	points []geom.Pt
}

func vectorAnchorBounds(anchors []geom.Pt, u, v []float64) geom.Rect {
	have := false
	var bounds geom.Rect
	for i, anchor := range anchors {
		if i >= len(u) || i >= len(v) || !isFinite(u[i]) || !isFinite(v[i]) {
			continue
		}
		if !have {
			bounds = geom.Rect{Min: anchor, Max: anchor}
			have = true
			continue
		}
		bounds = expandRect(bounds, anchor)
	}
	if !have {
		return geom.Rect{}
	}
	return bounds
}

func flattenVectorSamples(x, y, u, v []float64, scalars []float64) ([]geom.Pt, []float64, []float64, []float64, bool) {
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	if len(u) < n {
		n = len(u)
	}
	if len(v) < n {
		n = len(v)
	}
	if n == 0 {
		return nil, nil, nil, nil, false
	}
	if len(scalars) > 0 && len(scalars) < n {
		return nil, nil, nil, nil, false
	}
	anchors := make([]geom.Pt, 0, n)
	uu := make([]float64, 0, n)
	vv := make([]float64, 0, n)
	outScalars := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		if !isFinite(x[i]) || !isFinite(y[i]) || !isFinite(u[i]) || !isFinite(v[i]) {
			continue
		}
		if len(scalars) > 0 && !isFinite(scalars[i]) {
			continue
		}
		anchors = append(anchors, geom.Pt{X: x[i], Y: y[i]})
		uu = append(uu, u[i])
		vv = append(vv, v[i])
		if len(scalars) > 0 {
			outScalars = append(outScalars, scalars[i])
		}
	}
	if len(anchors) == 0 {
		return nil, nil, nil, nil, false
	}
	return anchors, uu, vv, outScalars, true
}

func flattenVectorGrid(x, y []float64, u, v [][]float64, scalars []float64) ([]geom.Pt, []float64, []float64, []float64, bool) {
	rows := len(y)
	cols := len(x)
	if rows == 0 || cols == 0 || !sameGridShape(u, rows, cols) || !sameGridShape(v, rows, cols) {
		return nil, nil, nil, nil, false
	}
	if len(scalars) > 0 && len(scalars) != rows*cols {
		return nil, nil, nil, nil, false
	}
	anchors := make([]geom.Pt, 0, rows*cols)
	uu := make([]float64, 0, rows*cols)
	vv := make([]float64, 0, rows*cols)
	outScalars := make([]float64, 0, rows*cols)
	for yi := 0; yi < rows; yi++ {
		for xi := 0; xi < cols; xi++ {
			idx := yi*cols + xi
			if !isFinite(x[xi]) || !isFinite(y[yi]) || !isFinite(u[yi][xi]) || !isFinite(v[yi][xi]) {
				continue
			}
			if len(scalars) > 0 && !isFinite(scalars[idx]) {
				continue
			}
			anchors = append(anchors, geom.Pt{X: x[xi], Y: y[yi]})
			uu = append(uu, u[yi][xi])
			vv = append(vv, v[yi][xi])
			if len(scalars) > 0 {
				outScalars = append(outScalars, scalars[idx])
			}
		}
	}
	if len(anchors) == 0 {
		return nil, nil, nil, nil, false
	}
	return anchors, uu, vv, outScalars, true
}

func vectorScalarOptions(opts []QuiverOptions) []float64 {
	if len(opts) == 0 {
		return nil
	}
	if len(opts[0].CGrid) > 0 {
		return flattenScalarGrid(opts[0].CGrid)
	}
	return append([]float64(nil), opts[0].C...)
}

func barbsScalarOptions(opts []BarbsOptions) []float64 {
	if len(opts) == 0 {
		return nil
	}
	if len(opts[0].CGrid) > 0 {
		return flattenScalarGrid(opts[0].CGrid)
	}
	return append([]float64(nil), opts[0].C...)
}

func flattenScalarGrid(grid [][]float64) []float64 {
	if len(grid) == 0 {
		return nil
	}
	cols := len(grid[0])
	if cols == 0 {
		return nil
	}
	out := make([]float64, 0, len(grid)*cols)
	for _, row := range grid {
		if len(row) != cols {
			return nil
		}
		out = append(out, row...)
	}
	return out
}

func sameGridShape(values [][]float64, rows, cols int) bool {
	if len(values) != rows {
		return false
	}
	for _, row := range values {
		if len(row) != cols {
			return false
		}
	}
	return true
}

func vectorStrictlyIncreasing(values []float64) bool {
	if len(values) < 2 {
		return false
	}
	for i := 1; i < len(values); i++ {
		if !isFinite(values[i-1]) || !isFinite(values[i]) || values[i] <= values[i-1] {
			return false
		}
	}
	return true
}

func cloneMatrix(values [][]float64) [][]float64 {
	if len(values) == 0 {
		return nil
	}
	out := make([][]float64, len(values))
	for i := range values {
		out[i] = append([]float64(nil), values[i]...)
	}
	return out
}

func minSpacing(values []float64) float64 {
	best := math.Inf(1)
	for i := 1; i < len(values); i++ {
		delta := values[i] - values[i-1]
		if delta > 0 && delta < best {
			best = delta
		}
	}
	if math.IsInf(best, 1) {
		return 0
	}
	return best
}

func resolvedStreamDensity(opt StreamplotOptions) (float64, float64) {
	base := opt.Density
	if base <= 0 {
		base = 1
	}
	x := opt.DensityX
	if x <= 0 {
		x = base
	}
	y := opt.DensityY
	if y <= 0 {
		y = base
	}
	return x, y
}

func scalarColormap(name *string) string {
	if name == nil {
		return "viridis"
	}
	return resolvedColormapName(*name)
}

func scalarVMin(original, scalars []float64, explicit *float64) float64 {
	minValue, _ := finiteRange(scalars)
	if len(scalars) == 0 {
		minValue = 0
	}
	if explicit != nil && isFinite(*explicit) {
		return *explicit
	}
	return minValue
}

func scalarVMax(original, scalars []float64, explicit *float64) float64 {
	_, maxValue := finiteRange(scalars)
	if len(scalars) == 0 {
		maxValue = 1
	}
	if explicit != nil && isFinite(*explicit) {
		return *explicit
	}
	return maxValue
}

func optionFloat(value *float64, fallback float64) float64 {
	if value == nil {
		return fallback
	}
	return *value
}

func optionInt(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func optionBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func optionAlpha(value *float64) float64 {
	if value == nil {
		return 1
	}
	return clampOneToOne(*value)
}

func derefColor(value *render.Color) render.Color {
	if value == nil {
		return render.Color{}
	}
	return *value
}

func normalizeVectorPivot(value, fallback string) string {
	switch strings.ToLower(value) {
	case "mid":
		return vectorPivotMiddle
	case vectorPivotMiddle, vectorPivotTip, vectorPivotTail:
		return strings.ToLower(value)
	default:
		return fallback
	}
}

func normalizeQuiverAngles(value string) string {
	switch strings.ToLower(value) {
	case quiverAnglesXY:
		return quiverAnglesXY
	default:
		return quiverAnglesUV
	}
}

func normalizeVectorUnits(value, fallback string) string {
	switch strings.ToLower(value) {
	case "height", "dots", "inches", "points", "x", "y", "xy", "width":
		return strings.ToLower(value)
	default:
		return fallback
	}
}

func normalizeStreamDirection(value string) string {
	switch strings.ToLower(value) {
	case streamDirectionForward, streamDirectionBackward:
		return strings.ToLower(value)
	default:
		return streamDirectionBoth
	}
}

func defaultBarbSizes(value *BarbSizes) BarbSizes {
	if value == nil {
		return BarbSizes{Spacing: 0.125, Height: 0.4, Width: 0.25, EmptyBarb: 0.15}
	}
	return BarbSizes{
		Spacing:   positiveOrDefault(value.Spacing, 0.125),
		Height:    positiveOrDefault(value.Height, 0.4),
		Width:     positiveOrDefault(value.Width, 0.25),
		EmptyBarb: positiveOrDefault(value.EmptyBarb, 0.15),
	}
}

func defaultBarbIncrements(value *BarbIncrements) BarbIncrements {
	if value == nil {
		return BarbIncrements{Half: 5, Full: 10, Flag: 50}
	}
	return BarbIncrements{
		Half: positiveOrDefault(value.Half, 5),
		Full: positiveOrDefault(value.Full, 10),
		Flag: positiveOrDefault(value.Flag, 50),
	}
}

func positiveOrDefault(value, fallback float64) float64 {
	if value <= 0 {
		return fallback
	}
	return value
}

func normalizeFlipSlice(flipBarb *bool, flip []bool, count int) []bool {
	if len(flip) > 0 {
		out := make([]bool, count)
		for i := range out {
			if i < len(flip) {
				out[i] = flip[i]
			}
		}
		return out
	}
	if flipBarb != nil && *flipBarb {
		out := make([]bool, count)
		for i := range out {
			out[i] = true
		}
		return out
	}
	return nil
}
