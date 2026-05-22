package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

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
