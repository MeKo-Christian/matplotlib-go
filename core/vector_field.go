package core

import (
	"math"
	"sort"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
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

// Quiver adds a vector-arrow artist to the axes.
func (a *Axes) Quiver(x, y, u, v []float64, opts ...QuiverOptions) *Quiver {
	if a == nil {
		return nil
	}
	anchors, uu, vv, scalars, ok := flattenVectorSamples(x, y, u, v, vectorScalarOptions(opts))
	if !ok || len(anchors) == 0 {
		return nil
	}

	var opt QuiverOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	alpha := optionAlpha(opt.Alpha)
	edgeWidth := optionFloat(opt.EdgeWidth, 0)
	edgeColor := render.Color{}
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}
	mapping, err := ResolveScalarMapValues(scalars, ScalarMapConfig{
		Colormap: scalarColormap(opt.Colormap),
		Norm:     opt.Norm,
		VMin:     opt.VMin,
		VMax:     opt.VMax,
	})
	if err != nil {
		return nil
	}

	q := &Quiver{
		Anchors:        anchors,
		U:              uu,
		V:              vv,
		Colors:         append([]render.Color(nil), opt.Colors...),
		Color:          color,
		ScalarColors:   append([]float64(nil), scalars...),
		EdgeColor:      edgeColor,
		EdgeWidth:      edgeWidth,
		Alpha:          alpha,
		Pivot:          normalizeVectorPivot(opt.Pivot, vectorPivotTail),
		Angles:         normalizeQuiverAngles(opt.Angles),
		AngleValues:    append([]float64(nil), opt.AngleValues...),
		ScaleUnits:     normalizeVectorUnits(opt.ScaleUnits, "width"),
		Units:          normalizeVectorUnits(opt.Units, "width"),
		Width:          optionFloat(opt.Width, 0),
		HeadWidth:      optionFloat(opt.HeadWidth, 3),
		HeadLength:     optionFloat(opt.HeadLength, 5),
		HeadAxisLength: optionFloat(opt.HeadAxisLength, 4.5),
		MinShaft:       optionFloat(opt.MinShaft, 1),
		MinLength:      optionFloat(opt.MinLength, 1),
		Label:          opt.Label,
		Colormap:       mapping.Colormap,
		Norm:           mapping.Norm,
		VMin:           mapping.VMin,
		VMax:           mapping.VMax,
		z:              optionFloat(opt.ZOrder, 1),
	}
	if opt.Scale != nil && *opt.Scale > 0 {
		q.Scale = *opt.Scale
		q.ScaleSet = true
	}
	a.Add(q)
	return q
}

// QuiverGrid expands rectilinear x/y coordinates with u/v grids.
func (a *Axes) QuiverGrid(x, y []float64, u, v [][]float64, opts ...QuiverOptions) *Quiver {
	if a == nil {
		return nil
	}
	anchors, uu, vv, scalars, ok := flattenVectorGrid(x, y, u, v, vectorScalarOptions(opts))
	if !ok {
		return nil
	}
	var opt QuiverOptions
	if len(opts) > 0 {
		opt = opts[0]
		opt.C = scalars
		opt.CGrid = nil
	}
	return a.quiverFromFlattened(anchors, uu, vv, scalars, opt)
}

func (a *Axes) quiverFromFlattened(anchors []geom.Pt, u, v, scalars []float64, opt QuiverOptions) *Quiver {
	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	mapping, err := ResolveScalarMapValues(scalars, ScalarMapConfig{
		Colormap: scalarColormap(opt.Colormap),
		Norm:     opt.Norm,
		VMin:     opt.VMin,
		VMax:     opt.VMax,
	})
	if err != nil {
		return nil
	}
	q := &Quiver{
		Anchors:        anchors,
		U:              append([]float64(nil), u...),
		V:              append([]float64(nil), v...),
		Colors:         append([]render.Color(nil), opt.Colors...),
		Color:          color,
		ScalarColors:   append([]float64(nil), scalars...),
		EdgeColor:      derefColor(opt.EdgeColor),
		EdgeWidth:      optionFloat(opt.EdgeWidth, 0),
		Alpha:          optionAlpha(opt.Alpha),
		Pivot:          normalizeVectorPivot(opt.Pivot, vectorPivotTail),
		Angles:         normalizeQuiverAngles(opt.Angles),
		AngleValues:    append([]float64(nil), opt.AngleValues...),
		ScaleUnits:     normalizeVectorUnits(opt.ScaleUnits, "width"),
		Units:          normalizeVectorUnits(opt.Units, "width"),
		Width:          optionFloat(opt.Width, 0),
		HeadWidth:      optionFloat(opt.HeadWidth, 3),
		HeadLength:     optionFloat(opt.HeadLength, 5),
		HeadAxisLength: optionFloat(opt.HeadAxisLength, 4.5),
		MinShaft:       optionFloat(opt.MinShaft, 1),
		MinLength:      optionFloat(opt.MinLength, 1),
		Label:          opt.Label,
		Colormap:       mapping.Colormap,
		Norm:           mapping.Norm,
		VMin:           mapping.VMin,
		VMax:           mapping.VMax,
		z:              optionFloat(opt.ZOrder, 1),
	}
	if opt.Scale != nil && *opt.Scale > 0 {
		q.Scale = *opt.Scale
		q.ScaleSet = true
	}
	a.Add(q)
	return q
}

// QuiverKey adds a labeled reference arrow reusing the style of q.
func (a *Axes) QuiverKey(q *Quiver, x, y, u float64, label string, opts ...QuiverKeyOptions) *QuiverKey {
	if a == nil || q == nil {
		return nil
	}
	opt := QuiverKeyOptions{
		Coords:   Coords(CoordAxes),
		LabelPos: "N",
		LabelSep: 10,
	}
	if len(opts) > 0 {
		opt = opts[0]
		if opt.LabelPos == "" {
			opt.LabelPos = "N"
		}
		if opt.LabelSep <= 0 {
			opt.LabelSep = 10
		}
	}

	key := &QuiverKey{
		Quiver:     q,
		Position:   geom.Pt{X: x, Y: y},
		U:          u,
		Label:      label,
		Coords:     opt.Coords,
		Angle:      opt.Angle,
		LabelPos:   strings.ToUpper(opt.LabelPos),
		LabelSep:   opt.LabelSep,
		Color:      opt.Color,
		LabelColor: opt.LabelColor,
		FontSize:   opt.FontSize,
		z:          optionFloat(opt.ZOrder, q.Z()+0.1),
	}
	a.Add(key)
	return key
}

// Barbs adds a wind-barb artist to the axes.
func (a *Axes) Barbs(x, y, u, v []float64, opts ...BarbsOptions) *Barbs {
	if a == nil {
		return nil
	}
	anchors, uu, vv, scalars, ok := flattenVectorSamples(x, y, u, v, barbsScalarOptions(opts))
	if !ok || len(anchors) == 0 {
		return nil
	}

	var opt BarbsOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	barbColor := color
	if opt.BarbColor != nil {
		barbColor = *opt.BarbColor
	}
	flagColor := barbColor
	if opt.FlagColor != nil {
		flagColor = *opt.FlagColor
	}
	mapping, err := ResolveScalarMapValues(scalars, ScalarMapConfig{
		Colormap: scalarColormap(opt.Colormap),
		Norm:     opt.Norm,
		VMin:     opt.VMin,
		VMax:     opt.VMax,
	})
	if err != nil {
		return nil
	}

	b := &Barbs{
		Anchors:      anchors,
		U:            uu,
		V:            vv,
		Colors:       append([]render.Color(nil), opt.Colors...),
		Color:        color,
		ScalarColors: append([]float64(nil), scalars...),
		BarbColor:    barbColor,
		FlagColor:    flagColor,
		LineWidth:    optionFloat(opt.LineWidth, 1),
		Alpha:        optionAlpha(opt.Alpha),
		Pivot:        normalizeVectorPivot(opt.Pivot, vectorPivotTip),
		Length:       optionFloat(opt.Length, 7),
		Units:        normalizeVectorUnits(opt.Units, "points"),
		Sizes:        defaultBarbSizes(opt.Sizes),
		Increments:   defaultBarbIncrements(opt.Increments),
		FillEmpty:    optionBool(opt.FillEmpty, false),
		Rounding:     optionBool(opt.Rounding, true),
		Flip:         normalizeFlipSlice(opt.FlipBarb, opt.Flip, len(anchors)),
		Label:        opt.Label,
		Colormap:     mapping.Colormap,
		Norm:         mapping.Norm,
		VMin:         mapping.VMin,
		VMax:         mapping.VMax,
		z:            optionFloat(opt.ZOrder, 1),
	}
	a.Add(b)
	return b
}

// BarbsGrid expands rectilinear x/y coordinates with u/v barb grids.
func (a *Axes) BarbsGrid(x, y []float64, u, v [][]float64, opts ...BarbsOptions) *Barbs {
	if a == nil {
		return nil
	}
	anchors, uu, vv, scalars, ok := flattenVectorGrid(x, y, u, v, barbsScalarOptions(opts))
	if !ok {
		return nil
	}

	var opt BarbsOptions
	if len(opts) > 0 {
		opt = opts[0]
		opt.C = scalars
		opt.CGrid = nil
	}

	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	barbColor := color
	if opt.BarbColor != nil {
		barbColor = *opt.BarbColor
	}
	flagColor := barbColor
	if opt.FlagColor != nil {
		flagColor = *opt.FlagColor
	}
	mapping, err := ResolveScalarMapValues(scalars, ScalarMapConfig{
		Colormap: scalarColormap(opt.Colormap),
		Norm:     opt.Norm,
		VMin:     opt.VMin,
		VMax:     opt.VMax,
	})
	if err != nil {
		return nil
	}

	b := &Barbs{
		Anchors:      anchors,
		U:            append([]float64(nil), uu...),
		V:            append([]float64(nil), vv...),
		Colors:       append([]render.Color(nil), opt.Colors...),
		Color:        color,
		ScalarColors: append([]float64(nil), scalars...),
		BarbColor:    barbColor,
		FlagColor:    flagColor,
		LineWidth:    optionFloat(opt.LineWidth, 1),
		Alpha:        optionAlpha(opt.Alpha),
		Pivot:        normalizeVectorPivot(opt.Pivot, vectorPivotTip),
		Length:       optionFloat(opt.Length, 7),
		Units:        normalizeVectorUnits(opt.Units, "points"),
		Sizes:        defaultBarbSizes(opt.Sizes),
		Increments:   defaultBarbIncrements(opt.Increments),
		FillEmpty:    optionBool(opt.FillEmpty, false),
		Rounding:     optionBool(opt.Rounding, true),
		Flip:         normalizeFlipSlice(opt.FlipBarb, opt.Flip, len(anchors)),
		Label:        opt.Label,
		Colormap:     mapping.Colormap,
		Norm:         mapping.Norm,
		VMin:         mapping.VMin,
		VMax:         mapping.VMax,
		z:            optionFloat(opt.ZOrder, 1),
	}
	a.Add(b)
	return b
}

// Streamplot adds a streamline set over a rectilinear vector grid.
func (a *Axes) Streamplot(x, y []float64, u, v [][]float64, opts ...StreamplotOptions) *StreamplotSet {
	if a == nil {
		return nil
	}
	if !vectorStrictlyIncreasing(x) || !vectorStrictlyIncreasing(y) {
		return nil
	}
	if !sameGridShape(u, len(y), len(x)) || !sameGridShape(v, len(y), len(x)) {
		return nil
	}

	opt := StreamplotOptions{
		Density:              1,
		IntegrationDirection: streamDirectionBoth,
	}
	if len(opts) > 0 {
		opt = opts[0]
	}
	densityX, densityY := resolvedStreamDensity(opt)
	if densityX <= 0 || densityY <= 0 {
		return nil
	}

	grid := streamplotGrid{
		x: append([]float64(nil), x...),
		y: append([]float64(nil), y...),
		u: cloneMatrix(u),
		v: cloneMatrix(v),
	}
	var mapping ScalarMapInfo
	hasScalarColors := len(opt.CGrid) > 0
	if hasScalarColors {
		if !sameGridShape(opt.CGrid, len(y), len(x)) {
			return nil
		}
		var err error
		mapping, err = ResolveScalarMapGrid(opt.CGrid, ScalarMapConfig{
			Colormap: scalarColormap(opt.Colormap),
			Norm:     opt.Norm,
			VMin:     opt.VMin,
			VMax:     opt.VMax,
		})
		if err != nil {
			return nil
		}
	}
	trajectories := computeStreamTrajectories(grid, opt, densityX, densityY)
	if len(trajectories) == 0 {
		return nil
	}

	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	lineWidth := optionFloat(opt.LineWidth, 1.5)
	z := optionFloat(opt.ZOrder, 1)

	lines := &LineCollection{
		Collection: Collection{
			Coords:   Coords(CoordData),
			Colormap: mapping.Colormap,
			Norm:     mapping.Norm,
			VMin:     mapping.VMin,
			VMax:     mapping.VMax,
			z:        z,
		},
		Color:     color,
		LineWidth: lineWidth,
		Segments:  make([][]geom.Pt, 0, len(trajectories)),
	}
	for _, trajectory := range trajectories {
		lines.Segments = append(lines.Segments, trajectory.points)
		if hasScalarColors {
			value, ok := streamScalarAverage(grid, opt.CGrid, trajectory.points)
			if !ok {
				value = math.NaN()
			}
			lines.Colors = append(lines.Colors, mapping.Color(value, 1))
		}
	}
	if hasScalarColors {
		lines.Color = render.Color{}
	}

	arrowColor := color
	if opt.ArrowColor != nil {
		arrowColor = *opt.ArrowColor
	}
	arrowSize := optionFloat(opt.ArrowSize, 1)
	arrowCount := optionInt(opt.ArrowCount, 1)
	if arrowCount < 0 {
		return nil
	}

	arrowAnchors, arrowU, arrowV := sampleStreamArrows(trajectories, arrowCount)
	arrowScalars := []float64(nil)
	if hasScalarColors {
		arrowScalars = make([]float64, len(arrowAnchors))
		for i, anchor := range arrowAnchors {
			value, ok := interpolateStreamScalar(grid, opt.CGrid, anchor)
			if !ok {
				value = math.NaN()
			}
			arrowScalars[i] = value
		}
	}
	arrows := &Quiver{
		Anchors:        arrowAnchors,
		U:              arrowU,
		V:              arrowV,
		ScalarColors:   arrowScalars,
		Color:          arrowColor,
		EdgeColor:      render.Color{},
		EdgeWidth:      0,
		Alpha:          1,
		Pivot:          vectorPivotMiddle,
		Angles:         quiverAnglesXY,
		Units:          "dots",
		ScaleUnits:     "dots",
		HeadWidth:      3,
		HeadLength:     5,
		HeadAxisLength: 4.5,
		MinShaft:       1,
		MinLength:      1,
		Colormap:       mapping.Colormap,
		Norm:           mapping.Norm,
		VMin:           mapping.VMin,
		VMax:           mapping.VMax,
		z:              z,
		forceLengthPx:  12 * arrowSize,
	}
	if lineWidth > 0 {
		arrows.Width = math.Max(1, lineWidth*1.8)
	}

	set := &StreamplotSet{
		Lines:  lines,
		Arrows: arrows,
		Label:  opt.Label,
		z:      z,
	}
	a.Add(set)
	return set
}

// Draw renders quiver arrows.
func (q *Quiver) Draw(r render.Renderer, ctx *DrawContext) {
	if q == nil || r == nil || ctx == nil {
		return
	}
	q.asPathCollection(ctx).Draw(r, ctx)
}

// Bounds returns the anchor bounds for autoscale purposes.
func (q *Quiver) Bounds(*DrawContext) geom.Rect {
	return vectorAnchorBounds(q.Anchors, q.U, q.V)
}

// Z returns the draw order.
func (q *Quiver) Z() float64 {
	if q == nil {
		return 0
	}
	return q.z
}

// ScalarMap exposes the scalar color mapping for helpers such as colorbars.
func (q *Quiver) ScalarMap() ScalarMapInfo {
	if q == nil || len(q.ScalarColors) == 0 {
		return ScalarMapInfo{}
	}
	return ScalarMapInfo{Colormap: q.Colormap, Norm: q.Norm, VMin: q.VMin, VMax: q.VMax}
}

func (q *Quiver) legendEntry() (legendEntry, bool) {
	if q == nil || q.Label == "" {
		return legendEntry{}, false
	}
	fill := q.sampleFillColor()
	edge := q.sampleEdgeColor(fill)
	return legendEntry{
		Label:           q.Label,
		kind:            legendEntryMarker,
		markerPath:      quiverGlyphPath(2.0, 0.35, q.HeadWidth, q.HeadLength, q.HeadAxisLength, q.MinShaft, q.MinLength, vectorPivotMiddle),
		markerFill:      fill,
		markerEdge:      edge,
		markerEdgeWidth: q.EdgeWidth,
	}, true
}

func (q *Quiver) sampleFillColor() render.Color {
	if len(q.ScalarColors) > 0 {
		return q.ScalarMap().Resolved().Color(q.ScalarColors[0], q.Alpha)
	}
	if len(q.Colors) > 0 {
		return patchAlphaColor(q.Colors[0], q.Alpha)
	}
	return patchAlphaColor(q.Color, q.Alpha)
}

func (q *Quiver) sampleEdgeColor(fill render.Color) render.Color {
	edge := patchAlphaColor(q.EdgeColor, q.Alpha)
	if edge.A <= 0 && q.EdgeWidth > 0 {
		return fill
	}
	return edge
}

func (q *Quiver) asPathCollection(ctx *DrawContext) *PathCollection {
	paths, colors := q.pathsForContext(ctx)
	return &PathCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  q.Label,
			Alpha:  1,
			z:      q.z,
		},
		Paths:         paths,
		Offsets:       append([]geom.Pt(nil), q.Anchors...),
		PathInDisplay: true,
		FaceColors:    colors,
		EdgeColor:     q.sampleEdgeColor(q.sampleFillColor()),
		EdgeWidth:     q.EdgeWidth,
	}
}

func (q *Quiver) pathsForContext(ctx *DrawContext) ([]geom.Path, []render.Color) {
	if q == nil || ctx == nil {
		return nil, nil
	}
	paths := make([]geom.Path, len(q.Anchors))
	colors := make([]render.Color, len(q.Anchors))
	state := q.renderState(ctx)
	for i := range q.Anchors {
		vector, ok := q.displayVectorAt(ctx, i, state)
		if !ok {
			continue
		}
		length := math.Hypot(vector.X, vector.Y)
		if length == 0 {
			continue
		}
		angle := math.Atan2(vector.Y, vector.X)
		paths[i] = applyAffinePath(
			quiverGlyphPath(length, state.widthPx, q.HeadWidth, q.HeadLength, q.HeadAxisLength, q.MinShaft, q.MinLength, q.Pivot),
			geom.Affine{
				A: math.Cos(angle),
				B: math.Sin(angle),
				C: -math.Sin(angle),
				D: math.Cos(angle),
			},
		)
		colors[i] = q.fillColorAt(i)
	}
	return paths, colors
}

func (q *Quiver) fillColorAt(i int) render.Color {
	switch {
	case len(q.ScalarColors) > 0 && i < len(q.ScalarColors):
		return q.ScalarMap().Resolved().Color(q.ScalarColors[i], q.Alpha)
	case len(q.Colors) > 0 && i < len(q.Colors):
		return patchAlphaColor(q.Colors[i], q.Alpha)
	default:
		return patchAlphaColor(q.Color, q.Alpha)
	}
}

func (q *Quiver) renderState(ctx *DrawContext) vectorRenderState {
	widthPx := q.Width
	if widthPx > 0 {
		widthPx *= dotsPerUnit(ctx, q.Units)
	} else {
		n := math.Sqrt(float64(max(len(q.Anchors), 1)))
		if n < 8 {
			n = 8
		}
		if n > 25 {
			n = 25
		}
		widthPx = 0.06 * ctx.Clip.W() / n
	}
	if widthPx <= 0 {
		widthPx = 4
	}

	scale := q.Scale
	if !q.ScaleSet && q.forceLengthPx <= 0 {
		baseLengths := make([]float64, 0, len(q.Anchors))
		for i := range q.Anchors {
			base, ok := q.baseDisplayLengthAt(ctx, i)
			if ok && isFinite(base) && base > 0 {
				baseLengths = append(baseLengths, base)
			}
		}
		if len(baseLengths) == 0 {
			scale = 1
		} else {
			mean := 0.0
			for _, value := range baseLengths {
				mean += value
			}
			mean /= float64(len(baseLengths))
			target := 0.18 * math.Min(ctx.Clip.W(), ctx.Clip.H())
			if len(baseLengths) > 1 {
				target /= math.Max(1, math.Sqrt(float64(len(baseLengths))))
			}
			if target <= 0 {
				target = 1
			}
			scale = mean / target
		}
	}
	if scale <= 0 {
		scale = 1
	}
	return vectorRenderState{widthPx: widthPx, scale: scale}
}

func (q *Quiver) displayVectorAt(ctx *DrawContext, i int, state vectorRenderState) (geom.Pt, bool) {
	if q == nil || ctx == nil || i >= len(q.Anchors) || i >= len(q.U) || i >= len(q.V) {
		return geom.Pt{}, false
	}
	u, v := q.U[i], q.V[i]
	if !isFinite(u) || !isFinite(v) {
		return geom.Pt{}, false
	}
	if q.forceLengthPx > 0 {
		direction, ok := q.directionVectorAt(ctx, i)
		if !ok {
			return geom.Pt{}, false
		}
		return direction, true
	}

	length, ok := q.baseDisplayLengthAt(ctx, i)
	if !ok {
		return geom.Pt{}, false
	}
	length /= state.scale
	if length <= 0 {
		return geom.Pt{}, false
	}

	unit, ok := q.unitDirectionAt(ctx, i)
	if !ok {
		return geom.Pt{}, false
	}
	return geom.Pt{X: unit.X * length, Y: unit.Y * length}, true
}

func (q *Quiver) directionVectorAt(ctx *DrawContext, i int) (geom.Pt, bool) {
	unit, ok := q.unitDirectionAt(ctx, i)
	if !ok {
		return geom.Pt{}, false
	}
	return geom.Pt{X: unit.X * q.forceLengthPx, Y: unit.Y * q.forceLengthPx}, true
}

func (q *Quiver) unitDirectionAt(ctx *DrawContext, i int) (geom.Pt, bool) {
	anchor := q.Anchors[i]
	u, v := q.U[i], q.V[i]
	if len(q.AngleValues) > 0 {
		if i >= len(q.AngleValues) || !isFinite(q.AngleValues[i]) {
			return geom.Pt{}, false
		}
		angle := q.AngleValues[i] * math.Pi / 180
		return geom.Pt{X: math.Cos(angle), Y: math.Sin(angle)}, true
	}

	if q.Angles == quiverAnglesXY {
		p1 := ctx.TransData().Apply(anchor)
		p2 := ctx.TransData().Apply(geom.Pt{X: anchor.X + u, Y: anchor.Y + v})
		dx := p2.X - p1.X
		dy := p2.Y - p1.Y
		length := math.Hypot(dx, dy)
		if length == 0 {
			return geom.Pt{}, false
		}
		return geom.Pt{X: dx / length, Y: dy / length}, true
	}

	length := math.Hypot(u, v)
	if length == 0 {
		return geom.Pt{}, false
	}
	return geom.Pt{X: u / length, Y: -v / length}, true
}

func (q *Quiver) baseDisplayLengthAt(ctx *DrawContext, i int) (float64, bool) {
	if q == nil || ctx == nil || i >= len(q.U) || i >= len(q.V) {
		return 0, false
	}
	u, v := q.U[i], q.V[i]
	if !isFinite(u) || !isFinite(v) {
		return 0, false
	}

	scaleUnits := q.ScaleUnits
	if scaleUnits == "" {
		scaleUnits = "width"
	}

	if scaleUnits == "xy" {
		anchor := q.Anchors[i]
		p1 := ctx.TransData().Apply(anchor)
		p2 := ctx.TransData().Apply(geom.Pt{X: anchor.X + u, Y: anchor.Y + v})
		return math.Hypot(p2.X-p1.X, p2.Y-p1.Y), true
	}

	switch scaleUnits {
	case "x":
		return math.Abs(u) * dotsPerUnit(ctx, "x"), true
	case "y":
		return math.Abs(v) * dotsPerUnit(ctx, "y"), true
	default:
		return math.Hypot(u, v) * dotsPerUnit(ctx, scaleUnits), true
	}
}

func dotsPerUnit(ctx *DrawContext, units string) float64 {
	if ctx == nil {
		return 1
	}
	units = normalizeVectorUnits(units, "width")
	xmin, xmax := ctx.DataToPixel.XScale.Domain()
	ymin, ymax := ctx.DataToPixel.YScale.Domain()
	xspan := math.Abs(xmax - xmin)
	yspan := math.Abs(ymax - ymin)
	if xspan == 0 {
		xspan = 1
	}
	if yspan == 0 {
		yspan = 1
	}

	switch units {
	case "x":
		return ctx.Clip.W() / xspan
	case "y":
		return ctx.Clip.H() / yspan
	case "xy":
		return math.Hypot(ctx.Clip.W(), ctx.Clip.H()) / math.Hypot(xspan, yspan)
	case "height":
		return ctx.Clip.H()
	case "dots":
		return 1
	case "points":
		dpi := ctx.RC.DPI
		if dpi <= 0 {
			dpi = 100
		}
		return dpi / 72
	case "inches":
		dpi := ctx.RC.DPI
		if dpi <= 0 {
			dpi = 100
		}
		return dpi
	default:
		return ctx.Clip.W()
	}
}

func quiverGlyphPath(lengthPx, widthPx, headWidthMul, headLengthMul, headAxisLengthMul, minShaft, minLength float64, pivot string) geom.Path {
	if lengthPx <= 0 || widthPx <= 0 {
		return geom.Path{}
	}
	if minLength <= 0 {
		minLength = 1
	}
	if lengthPx < minLength*widthPx {
		return regularPolygonPath(6, math.Max(widthPx*minLength*0.5, widthPx))
	}

	headLength := math.Max(0, headLengthMul*widthPx)
	headAxis := math.Max(0, headAxisLengthMul*widthPx)
	headWidth := math.Max(widthPx, headWidthMul*widthPx)
	minShaftLength := math.Max(0, minShaft*headLength)
	if minShaftLength > 0 && lengthPx < minShaftLength {
		scale := lengthPx / minShaftLength
		headLength *= scale
		headAxis *= scale
		headWidth *= scale
	}

	shaftHalf := widthPx * 0.5
	shaftEnd := math.Max(0, lengthPx-headAxis)
	headBase := math.Max(0, lengthPx-headLength)

	points := []geom.Pt{
		{X: 0, Y: -shaftHalf},
		{X: shaftEnd, Y: -shaftHalf},
		{X: headBase, Y: -headWidth * 0.5},
		{X: lengthPx, Y: 0},
		{X: headBase, Y: headWidth * 0.5},
		{X: shaftEnd, Y: shaftHalf},
		{X: 0, Y: shaftHalf},
	}
	path := polygonPath(points, true)

	shift := 0.0
	switch normalizeVectorPivot(pivot, vectorPivotTail) {
	case vectorPivotMiddle:
		shift = -lengthPx * 0.5
	case vectorPivotTip:
		shift = -lengthPx
	}
	if shift == 0 {
		return path
	}
	return applyAffinePath(path, translateAffine(geom.Pt{X: shift}))
}

func regularPolygonPath(sides int, radius float64) geom.Path {
	if sides < 3 || radius <= 0 {
		return geom.Path{}
	}
	points := make([]geom.Pt, 0, sides)
	for i := 0; i < sides; i++ {
		angle := 2 * math.Pi * float64(i) / float64(sides)
		points = append(points, geom.Pt{
			X: math.Cos(angle) * radius,
			Y: math.Sin(angle) * radius,
		})
	}
	return polygonPath(points, true)
}

// Draw is a no-op because quiver keys render outside the axes clip.
func (k *QuiverKey) Draw(render.Renderer, *DrawContext) {}

// DrawOverlay renders the key arrow and label after clipping has been removed.
func (k *QuiverKey) DrawOverlay(r render.Renderer, ctx *DrawContext) {
	if k == nil || k.Quiver == nil || ctx == nil {
		return
	}
	q := k.Quiver
	fill := k.Color
	if fill == (render.Color{}) {
		fill = q.sampleFillColor()
	}
	edge := q.sampleEdgeColor(fill)

	state := q.renderState(ctx)
	vector := k.displayVector(ctx, state)
	if vector == (geom.Pt{}) {
		return
	}
	length := math.Hypot(vector.X, vector.Y)
	angle := math.Atan2(vector.Y, vector.X)
	path := applyAffinePath(
		quiverGlyphPath(length, state.widthPx, q.HeadWidth, q.HeadLength, q.HeadAxisLength, q.MinShaft, q.MinLength, quiverKeyPivot(k.LabelPos)),
		geom.Affine{A: math.Cos(angle), B: math.Sin(angle), C: -math.Sin(angle), D: math.Cos(angle)},
	)
	anchor := transformedPoint(ctx, k.Coords, k.Position, 0, 0)
	path = applyAffinePath(path, translateAffine(anchor))
	r.Path(path, &render.Paint{
		Fill:      fill,
		Stroke:    edge,
		LineWidth: q.EdgeWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	})

	textRen, ok := r.(render.TextDrawer)
	if !ok || displayTextIsEmpty(k.Label) {
		return
	}
	labelPos := geom.Pt{X: anchor.X, Y: anchor.Y}
	switch strings.ToUpper(k.LabelPos) {
	case "S":
		labelPos.Y += k.LabelSep
	case "E":
		labelPos.X += k.LabelSep
	case "W":
		labelPos.X -= k.LabelSep
	default:
		labelPos.Y -= k.LabelSep
	}

	fontSize := resolvedFontSize(k.FontSize, ctx)
	layout := measureSingleLineTextLayout(r, k.Label, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
	origin := alignedSingleLineOrigin(labelPos, layout, quiverKeyHAlign(k.LabelPos), quiverKeyVAlign(k.LabelPos))
	drawDisplayText(textRen, k.Label, origin, fontSize, resolvedTextColor(k.LabelColor, ctx), ctx.RC.FontKey, ctx.RC.UseTeX)
}

func (k *QuiverKey) displayVector(ctx *DrawContext, state vectorRenderState) geom.Pt {
	if k == nil || k.Quiver == nil || ctx == nil || !isFinite(k.U) {
		return geom.Pt{}
	}
	lengthUnit := dotsPerUnit(ctx, k.Quiver.ScaleUnits)
	length := k.U * lengthUnit
	if !k.Quiver.ScaleSet && state.scale > 0 {
		length /= state.scale
	} else if k.Quiver.ScaleSet && k.Quiver.Scale > 0 {
		length /= k.Quiver.Scale
	}
	if length <= 0 {
		return geom.Pt{}
	}
	angle := k.Angle * math.Pi / 180
	return geom.Pt{X: math.Cos(angle) * length, Y: math.Sin(angle) * length}
}

func quiverKeyPivot(labelPos string) string {
	switch strings.ToUpper(labelPos) {
	case "E":
		return vectorPivotTip
	case "W":
		return vectorPivotTail
	default:
		return vectorPivotMiddle
	}
}

func quiverKeyHAlign(labelPos string) TextAlign {
	switch strings.ToUpper(labelPos) {
	case "E":
		return TextAlignLeft
	case "W":
		return TextAlignRight
	default:
		return TextAlignCenter
	}
}

func quiverKeyVAlign(labelPos string) textLayoutVerticalAlign {
	switch strings.ToUpper(labelPos) {
	case "S":
		return textLayoutVAlignTop
	case "E", "W":
		return textLayoutVAlignCenter
	default:
		return textLayoutVAlignBottom
	}
}

// Bounds returns an empty rect so keys do not affect autoscaling.
func (k *QuiverKey) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the draw order.
func (k *QuiverKey) Z() float64 {
	if k == nil {
		return 0
	}
	return k.z
}

// Draw renders barbs.
func (b *Barbs) Draw(r render.Renderer, ctx *DrawContext) {
	if b == nil || r == nil || ctx == nil {
		return
	}
	b.asPathCollection(ctx).Draw(r, ctx)
}

// Bounds returns the anchor bounds for autoscaling.
func (b *Barbs) Bounds(*DrawContext) geom.Rect {
	return vectorAnchorBounds(b.Anchors, b.U, b.V)
}

// Z returns the draw order.
func (b *Barbs) Z() float64 {
	if b == nil {
		return 0
	}
	return b.z
}

// ScalarMap exposes the scalar color mapping for colorbar helpers.
func (b *Barbs) ScalarMap() ScalarMapInfo {
	if b == nil || len(b.ScalarColors) == 0 {
		return ScalarMapInfo{}
	}
	return ScalarMapInfo{Colormap: b.Colormap, Norm: b.Norm, VMin: b.VMin, VMax: b.VMax}
}

func (b *Barbs) legendEntry() (legendEntry, bool) {
	if b == nil || b.Label == "" {
		return legendEntry{}, false
	}
	fill, edge := b.sampleColors(0)
	return legendEntry{
		Label:           b.Label,
		kind:            legendEntryMarker,
		markerPath:      b.barbGlyphPath(20, 0, false),
		markerFill:      fill,
		markerEdge:      edge,
		markerEdgeWidth: b.LineWidth,
	}, true
}

func (b *Barbs) asPathCollection(ctx *DrawContext) *PathCollection {
	paths := make([]geom.Path, len(b.Anchors))
	faceColors := make([]render.Color, len(b.Anchors))
	edgeColors := make([]render.Color, len(b.Anchors))
	lengthPx := b.Length * dotsPerUnit(ctx, b.Units)
	if lengthPx <= 0 {
		lengthPx = 18
	}

	for i := range b.Anchors {
		if i >= len(b.U) || i >= len(b.V) {
			continue
		}
		u, v := b.U[i], b.V[i]
		if !isFinite(u) || !isFinite(v) {
			continue
		}
		magnitude := math.Hypot(u, v)
		nFlags, nBarbs, half, empty := b.findTails(magnitude)
		path := b.barbGlyphPath(lengthPx, i, empty)
		if len(path.C) == 0 {
			continue
		}

		dx := u
		dy := v
		p1 := ctx.TransData().Apply(b.Anchors[i])
		p2 := ctx.TransData().Apply(geom.Pt{X: b.Anchors[i].X + dx, Y: b.Anchors[i].Y + dy})
		angle := math.Atan2(p2.Y-p1.Y, p2.X-p1.X)
		paths[i] = applyAffinePath(path, geom.Affine{
			A: math.Cos(angle),
			B: math.Sin(angle),
			C: -math.Sin(angle),
			D: math.Cos(angle),
		})
		fill, edge := b.colorsForIndex(i, nFlags > 0 || (empty && b.FillEmpty))
		faceColors[i] = fill
		edgeColors[i] = edge

		_ = nBarbs
		_ = half
	}

	return &PathCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  b.Label,
			z:      b.z,
		},
		Paths:         paths,
		Offsets:       append([]geom.Pt(nil), b.Anchors...),
		PathInDisplay: true,
		FaceColors:    faceColors,
		EdgeColors:    edgeColors,
		EdgeWidth:     b.LineWidth,
	}
}

func (b *Barbs) colorsForIndex(i int, hasFill bool) (render.Color, render.Color) {
	fill, edge := b.sampleColors(i)
	if !hasFill {
		fill.A = 0
	}
	return fill, edge
}

func (b *Barbs) sampleColors(i int) (render.Color, render.Color) {
	switch {
	case len(b.ScalarColors) > 0 && i < len(b.ScalarColors):
		color := b.ScalarMap().Resolved().Color(b.ScalarColors[i], b.Alpha)
		return color, color
	case len(b.Colors) > 0 && i < len(b.Colors):
		color := patchAlphaColor(b.Colors[i], b.Alpha)
		return color, color
	default:
		fill := patchAlphaColor(b.FlagColor, b.Alpha)
		edge := patchAlphaColor(b.BarbColor, b.Alpha)
		return fill, edge
	}
}

func (b *Barbs) findTails(mag float64) (nFlags, nBarbs int, half, empty bool) {
	if !isFinite(mag) || mag < 0 {
		return 0, 0, false, true
	}
	halfInc := positiveOrDefault(b.Increments.Half, 5)
	fullInc := positiveOrDefault(b.Increments.Full, 10)
	flagInc := positiveOrDefault(b.Increments.Flag, 50)
	if b.Rounding {
		mag = halfInc * math.Round(mag/halfInc)
	}
	nFlags = int(mag / flagInc)
	mag -= float64(nFlags) * flagInc
	nBarbs = int(mag / fullInc)
	mag -= float64(nBarbs) * fullInc
	half = mag >= halfInc
	empty = !half && nFlags == 0 && nBarbs == 0
	return nFlags, nBarbs, half, empty
}

func (b *Barbs) barbGlyphPath(lengthPx float64, i int, empty bool) geom.Path {
	if lengthPx <= 0 {
		return geom.Path{}
	}
	nFlags, nBarbs, half, isEmpty := b.findTails(math.Hypot(b.U[i], b.V[i]))
	if empty {
		isEmpty = true
	}
	if isEmpty {
		radius := lengthPx * positiveOrDefault(b.Sizes.EmptyBarb, 0.15)
		return regularPolygonPath(14, radius)
	}

	spacing := lengthPx * positiveOrDefault(b.Sizes.Spacing, 0.125)
	height := lengthPx * positiveOrDefault(b.Sizes.Height, 0.4)
	width := lengthPx * positiveOrDefault(b.Sizes.Width, 0.25)
	flip := len(b.Flip) > i && b.Flip[i]
	dir := -1.0
	if flip {
		dir = 1.0
	}

	path := geom.Path{}
	path.MoveTo(geom.Pt{})
	path.LineTo(geom.Pt{X: -lengthPx, Y: 0})

	offset := lengthPx
	for j := 0; j < nFlags; j++ {
		if offset != lengthPx {
			offset += spacing / 2
			path.LineTo(geom.Pt{X: -offset, Y: 0})
		}
		path.LineTo(geom.Pt{X: -(offset - width/2), Y: dir * height})
		path.LineTo(geom.Pt{X: -(offset - width), Y: 0})
		offset -= width + spacing
	}
	for j := 0; j < nBarbs; j++ {
		if offset != lengthPx {
			path.LineTo(geom.Pt{X: -offset, Y: 0})
		}
		path.LineTo(geom.Pt{X: -(offset + width/2), Y: dir * height})
		path.LineTo(geom.Pt{X: -offset, Y: 0})
		offset -= spacing
	}
	if half {
		if offset == lengthPx {
			offset -= 1.5 * spacing
		}
		path.LineTo(geom.Pt{X: -offset, Y: 0})
		path.LineTo(geom.Pt{X: -(offset + width/4), Y: dir * height * 0.5})
		path.LineTo(geom.Pt{X: -offset, Y: 0})
	}

	shift := 0.0
	switch normalizeVectorPivot(b.Pivot, vectorPivotTip) {
	case vectorPivotMiddle:
		shift = lengthPx * 0.5
	case vectorPivotTail:
		shift = lengthPx
	}
	if shift != 0 {
		return applyAffinePath(path, translateAffine(geom.Pt{X: shift}))
	}
	return path
}

// Draw renders the streamline line collection and its arrows.
func (s *StreamplotSet) Draw(r render.Renderer, ctx *DrawContext) {
	if s == nil {
		return
	}
	if s.Lines != nil {
		s.Lines.Draw(r, ctx)
	}
	if s.Arrows != nil {
		s.Arrows.Draw(r, ctx)
	}
}

// Bounds returns the line extents for autoscaling.
func (s *StreamplotSet) Bounds(ctx *DrawContext) geom.Rect {
	if s == nil || s.Lines == nil {
		return geom.Rect{}
	}
	return s.Lines.Bounds(ctx)
}

// Z returns the draw order.
func (s *StreamplotSet) Z() float64 {
	if s == nil {
		return 0
	}
	return s.z
}

// ScalarMap exposes scalar color mapping when streamplot arrows or lines carry one.
func (s *StreamplotSet) ScalarMap() ScalarMapInfo {
	if s == nil {
		return ScalarMapInfo{}
	}
	if s.Arrows != nil {
		if mapping := s.Arrows.ScalarMap(); scalarMapConfigured(mapping) {
			return mapping
		}
	}
	if s.Lines != nil && len(s.Lines.Colors) > 0 {
		if mapping := s.Lines.ScalarMap(); scalarMapConfigured(mapping) {
			return mapping
		}
	}
	return ScalarMapInfo{}
}

func (s *StreamplotSet) legendEntry() (legendEntry, bool) {
	if s == nil || s.Label == "" || s.Lines == nil {
		return legendEntry{}, false
	}
	return legendEntryFromLine(s.Label, s.Lines.Color, s.Lines.LineWidth, nil), true
}

func computeStreamTrajectories(grid streamplotGrid, opt StreamplotOptions, densityX, densityY float64) []streamTrajectory {
	mask := newStreamplotMask(grid.x[0], grid.x[len(grid.x)-1], grid.y[0], grid.y[len(grid.y)-1], densityX, densityY)
	minLength := optionFloat(opt.MinLength, 0.08)
	maxLength := optionFloat(opt.MaxLength, 4.0)
	if minLength <= 0 || maxLength <= 0 {
		return nil
	}
	broken := optionBool(opt.BrokenStreamlines, true)
	direction := normalizeStreamDirection(opt.IntegrationDirection)
	step := 0.35 * math.Min(minSpacing(grid.x), minSpacing(grid.y))
	if step <= 0 || !isFinite(step) {
		return nil
	}

	starts := opt.StartPoints
	if len(starts) == 0 {
		starts = mask.seedPoints()
	}

	trajectories := make([]streamTrajectory, 0, len(starts))
	for _, start := range starts {
		if !pointInsideGrid(start, grid) {
			continue
		}
		traj, used := integrateStream(grid, mask, start, step, minLength, maxLength, direction, broken)
		if len(traj.points) < 2 {
			continue
		}
		for idx := range used {
			mask.used[idx] = true
		}
		trajectories = append(trajectories, traj)
	}
	return trajectories
}

func integrateStream(grid streamplotGrid, mask *streamplotMask, start geom.Pt, step, minLength, maxLength float64, direction string, broken bool) (streamTrajectory, map[int]struct{}) {
	switch direction {
	case streamDirectionBackward:
		points, used := streamDirection(grid, mask, start, step, maxLength, -1, broken)
		if normalizedPathLength(points, grid) < minLength {
			return streamTrajectory{}, nil
		}
		return streamTrajectory{points: points}, used
	case streamDirectionForward:
		points, used := streamDirection(grid, mask, start, step, maxLength, 1, broken)
		if normalizedPathLength(points, grid) < minLength {
			return streamTrajectory{}, nil
		}
		return streamTrajectory{points: points}, used
	default:
		backward, usedBack := streamDirection(grid, mask, start, step, maxLength*0.5, -1, broken)
		forward, usedForward := streamDirection(grid, mask, start, step, maxLength*0.5, 1, broken)
		if len(backward) == 0 && len(forward) == 0 {
			return streamTrajectory{}, nil
		}
		points := make([]geom.Pt, 0, len(backward)+len(forward))
		for i := len(backward) - 1; i >= 0; i-- {
			points = append(points, backward[i])
		}
		if len(forward) > 0 {
			points = append(points, forward[1:]...)
		}
		if normalizedPathLength(points, grid) < minLength {
			return streamTrajectory{}, nil
		}
		used := map[int]struct{}{}
		for idx := range usedBack {
			used[idx] = struct{}{}
		}
		for idx := range usedForward {
			used[idx] = struct{}{}
		}
		return streamTrajectory{points: points}, used
	}
}

func streamDirection(grid streamplotGrid, mask *streamplotMask, start geom.Pt, step, maxLength float64, sign float64, broken bool) ([]geom.Pt, map[int]struct{}) {
	if !pointInsideGrid(start, grid) {
		return nil, nil
	}
	points := []geom.Pt{start}
	used := map[int]struct{}{mask.index(start): {}}
	total := 0.0
	current := start

	for total < maxLength {
		next, ok := streamStep(grid, current, step*sign)
		if !ok || !pointInsideGrid(next, grid) {
			break
		}
		idx := mask.index(next)
		if broken {
			if mask.used[idx] {
				break
			}
			if _, ok := used[idx]; ok {
				break
			}
		}
		total += normalizedSegmentLength(current, next, grid)
		if total > maxLength {
			break
		}
		points = append(points, next)
		used[idx] = struct{}{}
		current = next
	}
	return points, used
}

func streamStep(grid streamplotGrid, point geom.Pt, step float64) (geom.Pt, bool) {
	u1, v1, ok := interpolateStreamVector(grid, point)
	if !ok {
		return geom.Pt{}, false
	}
	mag1 := math.Hypot(u1, v1)
	if mag1 == 0 {
		return geom.Pt{}, false
	}
	mid := geom.Pt{
		X: point.X + (u1/mag1)*(step*0.5),
		Y: point.Y + (v1/mag1)*(step*0.5),
	}
	u2, v2, ok := interpolateStreamVector(grid, mid)
	if !ok {
		return geom.Pt{}, false
	}
	mag2 := math.Hypot(u2, v2)
	if mag2 == 0 {
		return geom.Pt{}, false
	}
	return geom.Pt{
		X: point.X + (u2/mag2)*step,
		Y: point.Y + (v2/mag2)*step,
	}, true
}

func interpolateStreamVector(grid streamplotGrid, point geom.Pt) (float64, float64, bool) {
	if !pointInsideGrid(point, grid) {
		return 0, 0, false
	}
	ix := locateInterval(grid.x, point.X)
	iy := locateInterval(grid.y, point.Y)
	if ix < 0 || iy < 0 {
		return 0, 0, false
	}
	x0, x1 := grid.x[ix], grid.x[ix+1]
	y0, y1 := grid.y[iy], grid.y[iy+1]
	tx := 0.0
	if x1 != x0 {
		tx = (point.X - x0) / (x1 - x0)
	}
	ty := 0.0
	if y1 != y0 {
		ty = (point.Y - y0) / (y1 - y0)
	}
	u00, u10 := grid.u[iy][ix], grid.u[iy][ix+1]
	u01, u11 := grid.u[iy+1][ix], grid.u[iy+1][ix+1]
	v00, v10 := grid.v[iy][ix], grid.v[iy][ix+1]
	v01, v11 := grid.v[iy+1][ix], grid.v[iy+1][ix+1]
	if !isFinite(u00) || !isFinite(u10) || !isFinite(u01) || !isFinite(u11) ||
		!isFinite(v00) || !isFinite(v10) || !isFinite(v01) || !isFinite(v11) {
		return 0, 0, false
	}
	u0 := u00*(1-tx) + u10*tx
	u1 := u01*(1-tx) + u11*tx
	v0 := v00*(1-tx) + v10*tx
	v1 := v01*(1-tx) + v11*tx
	return u0*(1-ty) + u1*ty, v0*(1-ty) + v1*ty, true
}

func streamScalarAverage(grid streamplotGrid, scalars [][]float64, points []geom.Pt) (float64, bool) {
	sum := 0.0
	count := 0
	for _, point := range points {
		value, ok := interpolateStreamScalar(grid, scalars, point)
		if !ok {
			continue
		}
		sum += value
		count++
	}
	if count == 0 {
		return 0, false
	}
	return sum / float64(count), true
}

func interpolateStreamScalar(grid streamplotGrid, scalars [][]float64, point geom.Pt) (float64, bool) {
	if !pointInsideGrid(point, grid) || !sameGridShape(scalars, len(grid.y), len(grid.x)) {
		return 0, false
	}
	ix := locateInterval(grid.x, point.X)
	iy := locateInterval(grid.y, point.Y)
	if ix < 0 || iy < 0 {
		return 0, false
	}
	x0, x1 := grid.x[ix], grid.x[ix+1]
	y0, y1 := grid.y[iy], grid.y[iy+1]
	tx := 0.0
	if x1 != x0 {
		tx = (point.X - x0) / (x1 - x0)
	}
	ty := 0.0
	if y1 != y0 {
		ty = (point.Y - y0) / (y1 - y0)
	}
	c00, c10 := scalars[iy][ix], scalars[iy][ix+1]
	c01, c11 := scalars[iy+1][ix], scalars[iy+1][ix+1]
	if !isFinite(c00) || !isFinite(c10) || !isFinite(c01) || !isFinite(c11) {
		return 0, false
	}
	c0 := c00*(1-tx) + c10*tx
	c1 := c01*(1-tx) + c11*tx
	return c0*(1-ty) + c1*ty, true
}

func pointInsideGrid(point geom.Pt, grid streamplotGrid) bool {
	return point.X >= grid.x[0] && point.X <= grid.x[len(grid.x)-1] &&
		point.Y >= grid.y[0] && point.Y <= grid.y[len(grid.y)-1]
}

func locateInterval(values []float64, target float64) int {
	if len(values) < 2 || target < values[0] || target > values[len(values)-1] {
		return -1
	}
	if target == values[len(values)-1] {
		return len(values) - 2
	}
	idx := sort.Search(len(values)-1, func(i int) bool { return values[i+1] > target })
	if idx < 0 || idx >= len(values)-1 {
		return -1
	}
	return idx
}

func normalizedPathLength(points []geom.Pt, grid streamplotGrid) float64 {
	total := 0.0
	for i := 1; i < len(points); i++ {
		total += normalizedSegmentLength(points[i-1], points[i], grid)
	}
	return total
}

func normalizedSegmentLength(a, b geom.Pt, grid streamplotGrid) float64 {
	xspan := grid.x[len(grid.x)-1] - grid.x[0]
	yspan := grid.y[len(grid.y)-1] - grid.y[0]
	if xspan == 0 {
		xspan = 1
	}
	if yspan == 0 {
		yspan = 1
	}
	return math.Hypot((b.X-a.X)/xspan, (b.Y-a.Y)/yspan)
}

func sampleStreamArrows(trajectories []streamTrajectory, count int) ([]geom.Pt, []float64, []float64) {
	if count <= 0 {
		return nil, nil, nil
	}
	var anchors []geom.Pt
	var u []float64
	var v []float64
	for _, trajectory := range trajectories {
		if len(trajectory.points) < 2 {
			continue
		}
		cum := make([]float64, len(trajectory.points))
		for i := 1; i < len(trajectory.points); i++ {
			cum[i] = cum[i-1] + math.Hypot(
				trajectory.points[i].X-trajectory.points[i-1].X,
				trajectory.points[i].Y-trajectory.points[i-1].Y,
			)
		}
		total := cum[len(cum)-1]
		if total == 0 {
			continue
		}
		for arrow := 0; arrow < count; arrow++ {
			target := total * float64(arrow+1) / float64(count+1)
			idx := sort.Search(len(cum), func(i int) bool { return cum[i] >= target })
			if idx <= 0 || idx >= len(cum) {
				continue
			}
			prev := trajectory.points[idx-1]
			next := trajectory.points[idx]
			segment := cum[idx] - cum[idx-1]
			t := 0.0
			if segment > 0 {
				t = (target - cum[idx-1]) / segment
			}
			point := geom.Pt{
				X: prev.X + (next.X-prev.X)*t,
				Y: prev.Y + (next.Y-prev.Y)*t,
			}
			anchors = append(anchors, point)
			u = append(u, next.X-prev.X)
			v = append(v, next.Y-prev.Y)
		}
	}
	return anchors, u, v
}

func newStreamplotMask(xmin, xmax, ymin, ymax, densityX, densityY float64) *streamplotMask {
	nx := int(math.Round(30 * densityX))
	ny := int(math.Round(30 * densityY))
	if nx < 1 {
		nx = 1
	}
	if ny < 1 {
		ny = 1
	}
	return &streamplotMask{
		nx:    nx,
		ny:    ny,
		used:  make([]bool, nx*ny),
		xmin:  xmin,
		xspan: xmax - xmin,
		ymin:  ymin,
		yspan: ymax - ymin,
	}
}

func (m *streamplotMask) seedPoints() []geom.Pt {
	points := make([]geom.Pt, 0, m.nx*m.ny)
	for yi := 0; yi < m.ny; yi++ {
		for xi := 0; xi < m.nx; xi++ {
			x := m.xmin + (float64(xi)+0.5)/float64(m.nx)*m.xspan
			y := m.ymin + (float64(yi)+0.5)/float64(m.ny)*m.yspan
			points = append(points, geom.Pt{X: x, Y: y})
		}
	}
	return points
}

func (m *streamplotMask) index(point geom.Pt) int {
	xn := clamp01((point.X - m.xmin) / maxF(m.xspan, 1e-9))
	yn := clamp01((point.Y - m.ymin) / maxF(m.yspan, 1e-9))
	xi := int(math.Min(float64(m.nx-1), math.Floor(xn*float64(m.nx))))
	yi := int(math.Min(float64(m.ny-1), math.Floor(yn*float64(m.ny))))
	return yi*m.nx + xi
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
