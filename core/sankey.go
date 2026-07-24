package core

import (
	"fmt"
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// SankeyOptions configures a lightweight builder for Sankey-like ribbon diagrams.
type SankeyOptions struct {
	Center      geom.Pt
	Coords      CoordinateSpec
	Scale       float64
	TrunkWidth  float64
	TrunkLength float64
	PathLength  float64
	Gap         float64
	Radius      float64
	Shoulder    float64
	Offset      float64
	HeadAngle   float64
	Margin      float64
	Tolerance   float64
	Alpha       float64
	TextColor   render.Color
	FontSize    float64
	EdgeColor   *render.Color
	FaceColor   *render.Color
}

// SankeyAddOptions configures one diagram added through the builder.
type SankeyAddOptions struct {
	Labels       []string
	Orientations []int
	Colors       []render.Color
	OffsetX      float64
	OffsetY      float64
	TrunkLength  float64
	PathLength   float64
	PatchLabel   string
	Rotation     float64
}

// Sankey accumulates one or more simple ribbon diagrams on a single axes.
type Sankey struct {
	ax         *Axes
	opts       SankeyOptions
	diagrams   []*SankeyDiagram
	cursor     geom.Pt
	pitch      float64
	extent     geom.Rect
	haveExtent bool
}

// SankeyDiagram groups the artists emitted for one builder Add call.
type SankeyDiagram struct {
	Trunk   *Rectangle
	Patch   *PathPatch
	Ribbons []*PathPatch
	Text    *Text
	Labels  []*Text
	Values  []*Text
	Flows   []float64
	Angles  []int
	Tips    []geom.Pt
	Center  geom.Pt
	Extent  geom.Rect
}

const (
	sankeyRight = 0
	sankeyUp    = 1
	sankeyDown  = 3
	sankeyNone  = -1
)

type sankeyPathCmd uint8

const (
	sankeyMoveTo sankeyPathCmd = iota
	sankeyLineTo
	sankeyCurveTo
	sankeyClosePath
)

type sankeyPathStep struct {
	cmd sankeyPathCmd
	pt  geom.Pt
}

// Add supports the Matplotlib-style builder workflow: create one diagram and
// append its artists to the axes immediately.
func (s *Sankey) Add(flows []float64, opts ...SankeyAddOptions) *SankeyDiagram {
	if s == nil || s.ax == nil || len(flows) == 0 {
		return nil
	}

	cfg := SankeyAddOptions{}
	if len(opts) > 0 {
		cfg = opts[0]
	}

	scale := s.opts.Scale
	if scale <= 0 {
		scale = 1
	}
	alpha := clampOneToOne(s.opts.Alpha)
	if alpha == 0 {
		alpha = 1
	}
	edgeColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	if s.opts.EdgeColor != nil {
		edgeColor = *s.opts.EdgeColor
	}
	defaultFace := s.ax.NextColor()
	if s.opts.FaceColor != nil {
		defaultFace = *s.opts.FaceColor
	}

	center := geom.Pt{X: s.cursor.X + cfg.OffsetX, Y: s.cursor.Y + cfg.OffsetY}
	trunkLength := s.opts.TrunkLength
	if cfg.TrunkLength > 0 {
		trunkLength = cfg.TrunkLength
	}
	if trunkLength < 0 {
		trunkLength = 0
	}
	pathLength := s.opts.PathLength
	if cfg.PathLength > 0 {
		pathLength = cfg.PathLength
	}
	if pathLength < 0 {
		pathLength = 0
	}
	tolerance := s.opts.Tolerance
	if tolerance <= 0 {
		tolerance = 1e-6
	}

	scaledFlows := make([]float64, len(flows))
	gain := 0.0
	loss := 0.0
	for i, flow := range flows {
		scaledFlows[i] = scale * flow
		if scaledFlows[i] > 0 {
			gain += scaledFlows[i]
		} else if scaledFlows[i] < 0 {
			loss += scaledFlows[i]
		}
	}

	areInputs := make([]bool, len(flows))
	showFlow := make([]bool, len(flows))
	for i, flow := range flows {
		switch {
		case flow >= tolerance:
			areInputs[i] = true
			showFlow[i] = true
		case flow <= -tolerance:
			showFlow[i] = true
		}
	}
	if !sankeyHasVisibleFlow(showFlow) {
		return nil
	}

	orientations := sankeyBroadcastInts(cfg.Orientations, len(flows), 0)
	angles := make([]int, len(flows))
	for i := range angles {
		angles[i] = sankeyNone
		if !showFlow[i] {
			continue
		}
		switch orientations[i] {
		case 1:
			if areInputs[i] {
				angles[i] = sankeyDown
			} else {
				angles[i] = sankeyUp
			}
		case 0:
			angles[i] = sankeyRight
		case -1:
			if areInputs[i] {
				angles[i] = sankeyUp
			} else {
				angles[i] = sankeyDown
			}
		default:
			angles[i] = sankeyRight
		}
	}

	pathLengths := sankeyJustifiedPathLengths(pathLength, scaledFlows, areInputs, showFlow, angles)
	urpath, llpath, lrpath, ulpath := sankeyInitialPaths(s.opts.Gap, trunkLength, gain, loss)

	tips := make([]geom.Pt, len(flows))
	labelLocations := make([]geom.Pt, len(flows))
	for i := range flows {
		if angles[i] == sankeyDown && areInputs[i] {
			tips[i], labelLocations[i] = s.sankeyAddInput(&ulpath, angles[i], scaledFlows[i], pathLengths[i])
		} else if angles[i] == sankeyUp && !areInputs[i] && showFlow[i] {
			tips[i], labelLocations[i] = s.sankeyAddOutput(&urpath, angles[i], scaledFlows[i], pathLengths[i])
		}
	}
	for i := len(flows) - 1; i >= 0; i-- {
		if angles[i] == sankeyUp && areInputs[i] {
			tips[i], labelLocations[i] = s.sankeyAddInput(&llpath, angles[i], scaledFlows[i], pathLengths[i])
		} else if angles[i] == sankeyDown && !areInputs[i] && showFlow[i] {
			tips[i], labelLocations[i] = s.sankeyAddOutput(&lrpath, angles[i], scaledFlows[i], pathLengths[i])
		}
	}
	hasLeftInput := false
	for i := len(flows) - 1; i >= 0; i-- {
		if angles[i] != sankeyRight || !areInputs[i] {
			continue
		}
		if !hasLeftInput {
			if sankeyLastPoint(llpath).X > sankeyLastPoint(ulpath).X {
				llpath = append(llpath, sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: sankeyLastPoint(ulpath).X, Y: sankeyLastPoint(llpath).Y}})
			}
			hasLeftInput = true
		}
		tips[i], labelLocations[i] = s.sankeyAddInput(&llpath, angles[i], scaledFlows[i], pathLengths[i])
	}
	hasRightOutput := false
	for i := range flows {
		if angles[i] != sankeyRight || areInputs[i] || !showFlow[i] {
			continue
		}
		if !hasRightOutput {
			if sankeyLastPoint(urpath).X < sankeyLastPoint(lrpath).X {
				urpath = append(urpath, sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: sankeyLastPoint(lrpath).X, Y: sankeyLastPoint(urpath).Y}})
			}
			hasRightOutput = true
		}
		tips[i], labelLocations[i] = s.sankeyAddOutput(&urpath, angles[i], scaledFlows[i], pathLengths[i])
	}
	if !hasLeftInput {
		ulpath = sankeyTrimLast(ulpath)
		llpath = sankeyTrimLast(llpath)
	}
	if !hasRightOutput {
		lrpath = sankeyTrimLast(lrpath)
		urpath = sankeyTrimLast(urpath)
	}

	steps := make([]sankeyPathStep, 0, len(urpath)+len(lrpath)+len(llpath)+len(ulpath)+1)
	steps = append(steps, urpath...)
	steps = append(steps, sankeyRevert(lrpath, sankeyLineTo)...)
	steps = append(steps, llpath...)
	steps = append(steps, sankeyRevert(ulpath, sankeyLineTo)...)
	if len(urpath) > 0 {
		steps = append(steps, sankeyPathStep{cmd: sankeyClosePath, pt: urpath[0].pt})
	}

	rotation := cfg.Rotation
	path := sankeyStepsToPath(steps, center, rotation)
	for i := range tips {
		tips[i] = sankeyTransformPoint(tips[i], center, rotation)
		labelLocations[i] = sankeyTransformPoint(labelLocations[i], center, rotation)
	}

	trunkHeight := math.Max(gain, -loss)
	trunk := &Rectangle{
		Patch: Patch{
			EdgeColor: edgeColor,
			EdgeWidth: 0,
			z:         2,
		},
		XY:     geom.Pt{X: center.X - trunkLength/2, Y: center.Y - trunkHeight/2},
		Width:  trunkLength,
		Height: trunkHeight,
		Coords: s.opts.Coords,
	}
	trunk.SetFaceColor(render.Color{})
	extent, haveExtent := sankeyExtent(path.V, labelLocations, showFlow)
	if !haveExtent {
		extent = geom.Rect{Min: center, Max: center}
	}
	diagram := &SankeyDiagram{
		Trunk:  trunk,
		Flows:  append([]float64(nil), flows...),
		Angles: append([]int(nil), angles...),
		Tips:   append([]geom.Pt(nil), tips...),
		Center: center,
		Extent: extent,
	}

	bodyPatch := &PathPatch{
		Patch: Patch{
			FaceColor: patchAlphaColor(defaultFace, alpha),
			EdgeColor: edgeColor,
			EdgeWidth: 1,
			Alpha:     1,
			z:         2.2,
		},
		Path:   path,
		Coords: s.opts.Coords,
	}
	s.ax.AddPatch(bodyPatch)
	diagram.Patch = bodyPatch
	diagram.Ribbons = append(diagram.Ribbons, bodyPatch)

	if cfg.PatchLabel != "" {
		diagram.Text = s.ax.Text(center.X, center.Y, cfg.PatchLabel, TextOptions{
			Coords:   s.opts.Coords,
			FontSize: s.opts.FontSize,
			Color:    s.opts.TextColor,
			HAlign:   TextAlignCenter,
			VAlign:   TextVAlignMiddle,
		})
	}

	labels := sankeyBroadcastStrings(cfg.Labels, len(flows), "")
	for i, flow := range flows {
		if !showFlow[i] || angles[i] == sankeyNone {
			continue
		}
		label := labels[i]
		value := fmt.Sprintf("%g", math.Abs(flow))
		pos := labelLocations[i]
		if label == "" {
			txt := s.ax.Text(pos.X, pos.Y, value, TextOptions{
				Coords:   s.opts.Coords,
				FontSize: s.opts.FontSize,
				Color:    s.opts.TextColor,
				HAlign:   TextAlignCenter,
				VAlign:   TextVAlignMiddle,
			})
			diagram.Values = append(diagram.Values, txt)
			continue
		}
		txt := s.ax.Text(pos.X, pos.Y, label, TextOptions{
			Coords:   s.opts.Coords,
			FontSize: s.opts.FontSize,
			Color:    s.opts.TextColor,
			HAlign:   TextAlignCenter,
			VAlign:   TextVAlignMiddle,
			OffsetY:  -s.opts.FontSize * 0.45,
		})
		diagram.Labels = append(diagram.Labels, txt)
		valueTxt := s.ax.Text(pos.X, pos.Y, value, TextOptions{
			Coords:   s.opts.Coords,
			FontSize: s.opts.FontSize,
			Color:    s.opts.TextColor,
			HAlign:   TextAlignCenter,
			VAlign:   TextVAlignMiddle,
			OffsetY:  s.opts.FontSize * 0.65,
		})
		diagram.Values = append(diagram.Values, valueTxt)
	}

	s.expandExtent(extent)
	s.diagrams = append(s.diagrams, diagram)
	return diagram
}

func sankeyHasVisibleFlow(show []bool) bool {
	for _, visible := range show {
		if visible {
			return true
		}
	}
	return false
}

func sankeyBroadcastInts(values []int, n, fallback int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = fallback
	}
	if len(values) == 1 {
		for i := range out {
			out[i] = values[0]
		}
		return out
	}
	for i := 0; i < n && i < len(values); i++ {
		out[i] = values[i]
	}
	return out
}

func sankeyBroadcastStrings(values []string, n int, fallback string) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fallback
	}
	if len(values) == 1 {
		for i := range out {
			out[i] = values[0]
		}
		return out
	}
	for i := 0; i < n && i < len(values); i++ {
		out[i] = values[i]
	}
	return out
}

func sankeyJustifiedPathLengths(base float64, flows []float64, areInputs, show []bool, angles []int) []float64 {
	pathLengths := make([]float64, len(flows))
	for i, angle := range angles {
		if angle == sankeyRight {
			pathLengths[i] = base
		}
	}

	urlength, ullength := base, base
	lrlength, lllength := base, base
	for i, angle := range angles {
		if !show[i] {
			continue
		}
		if angle == sankeyDown && areInputs[i] {
			pathLengths[i] = ullength
			ullength += flows[i]
		} else if angle == sankeyUp && !areInputs[i] {
			pathLengths[i] = urlength
			urlength -= flows[i]
		}
	}
	for i := len(angles) - 1; i >= 0; i-- {
		if !show[i] {
			continue
		}
		if angles[i] == sankeyUp && areInputs[i] {
			pathLengths[i] = lllength
			lllength += flows[i]
		} else if angles[i] == sankeyDown && !areInputs[i] {
			pathLengths[i] = lrlength
			lrlength -= flows[i]
		}
	}
	hasLeftInput := false
	for i := len(angles) - 1; i >= 0; i-- {
		if angles[i] == sankeyRight && areInputs[i] {
			if hasLeftInput {
				pathLengths[i] = 0
			} else {
				hasLeftInput = true
			}
		}
	}
	hasRightOutput := false
	for i, angle := range angles {
		if angle == sankeyRight && !areInputs[i] && show[i] {
			if hasRightOutput {
				pathLengths[i] = 0
			} else {
				hasRightOutput = true
			}
		}
	}
	return pathLengths
}

func sankeyInitialPaths(gap, trunkLength, gain, loss float64) ([]sankeyPathStep, []sankeyPathStep, []sankeyPathStep, []sankeyPathStep) {
	urpath := []sankeyPathStep{
		{cmd: sankeyMoveTo, pt: geom.Pt{X: gap - trunkLength/2, Y: gain / 2}},
		{cmd: sankeyLineTo, pt: geom.Pt{X: (gap - trunkLength/2) / 2, Y: gain / 2}},
		{cmd: sankeyCurveTo, pt: geom.Pt{X: (gap - trunkLength/2) / 8, Y: gain / 2}},
		{cmd: sankeyCurveTo, pt: geom.Pt{X: (trunkLength/2 - gap) / 8, Y: -loss / 2}},
		{cmd: sankeyCurveTo, pt: geom.Pt{X: (trunkLength/2 - gap) / 2, Y: -loss / 2}},
		{cmd: sankeyLineTo, pt: geom.Pt{X: trunkLength/2 - gap, Y: -loss / 2}},
	}
	llpath := []sankeyPathStep{
		{cmd: sankeyLineTo, pt: geom.Pt{X: trunkLength/2 - gap, Y: loss / 2}},
		{cmd: sankeyLineTo, pt: geom.Pt{X: (trunkLength/2 - gap) / 2, Y: loss / 2}},
		{cmd: sankeyCurveTo, pt: geom.Pt{X: (trunkLength/2 - gap) / 8, Y: loss / 2}},
		{cmd: sankeyCurveTo, pt: geom.Pt{X: (gap - trunkLength/2) / 8, Y: -gain / 2}},
		{cmd: sankeyCurveTo, pt: geom.Pt{X: (gap - trunkLength/2) / 2, Y: -gain / 2}},
		{cmd: sankeyLineTo, pt: geom.Pt{X: gap - trunkLength/2, Y: -gain / 2}},
	}
	lrpath := []sankeyPathStep{
		{cmd: sankeyLineTo, pt: geom.Pt{X: trunkLength/2 - gap, Y: loss / 2}},
	}
	ulpath := []sankeyPathStep{
		{cmd: sankeyLineTo, pt: geom.Pt{X: gap - trunkLength/2, Y: gain / 2}},
	}
	return urpath, llpath, lrpath, ulpath
}

func (s *Sankey) sankeyAddInput(path *[]sankeyPathStep, angle int, flow, length float64) (geom.Pt, geom.Pt) {
	if angle == sankeyNone {
		return geom.Pt{}, geom.Pt{}
	}
	x, y := sankeyLastPoint(*path).X, sankeyLastPoint(*path).Y
	dipDepth := flow / 2 * s.pitch
	if angle == sankeyRight {
		x -= length
		dip := geom.Pt{X: x + dipDepth, Y: y + flow/2}
		*path = append(
			*path,
			sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x, Y: y}},
			sankeyPathStep{cmd: sankeyLineTo, pt: dip},
			sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x, Y: y + flow}},
			sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x + s.opts.Gap, Y: y + flow}},
		)
		return dip, geom.Pt{X: dip.X - s.opts.Offset, Y: dip.Y}
	}

	x -= s.opts.Gap
	sign := -1.0
	quadrant := 2
	if angle == sankeyUp {
		sign = 1
		quadrant = 1
	}
	dip := geom.Pt{X: x - flow/2, Y: y - sign*(length-dipDepth)}
	if s.opts.Radius != 0 {
		*path = append(*path, sankeyArc(quadrant, angle == sankeyUp, s.opts.Radius, geom.Pt{X: x + s.opts.Radius, Y: y - sign*s.opts.Radius})...)
	} else {
		*path = append(*path, sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x, Y: y}})
	}
	*path = append(
		*path,
		sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x, Y: y - sign*length}},
		sankeyPathStep{cmd: sankeyLineTo, pt: dip},
		sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x - flow, Y: y - sign*length}},
	)
	*path = append(*path, sankeyArc(quadrant, angle == sankeyDown, flow+s.opts.Radius, geom.Pt{X: x + s.opts.Radius, Y: y - sign*s.opts.Radius})...)
	*path = append(*path, sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x - flow, Y: y + sign*flow}})
	return dip, geom.Pt{X: dip.X, Y: dip.Y - sign*s.opts.Offset}
}

func (s *Sankey) sankeyAddOutput(path *[]sankeyPathStep, angle int, flow, length float64) (geom.Pt, geom.Pt) {
	if angle == sankeyNone {
		return geom.Pt{}, geom.Pt{}
	}
	x, y := sankeyLastPoint(*path).X, sankeyLastPoint(*path).Y
	tipHeight := (s.opts.Shoulder - flow/2) * s.pitch
	if angle == sankeyRight {
		x += length
		tip := geom.Pt{X: x + tipHeight, Y: y + flow/2}
		*path = append(
			*path,
			sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x, Y: y}},
			sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x, Y: y + s.opts.Shoulder}},
			sankeyPathStep{cmd: sankeyLineTo, pt: tip},
			sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x, Y: y - s.opts.Shoulder + flow}},
			sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x, Y: y + flow}},
			sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x - s.opts.Gap, Y: y + flow}},
		)
		return tip, geom.Pt{X: tip.X + s.opts.Offset, Y: tip.Y}
	}

	x += s.opts.Gap
	sign := -1.0
	quadrant := 0
	if angle == sankeyUp {
		sign = 1
		quadrant = 3
	}
	tip := geom.Pt{X: x - flow/2, Y: y + sign*(length+tipHeight)}
	if s.opts.Radius != 0 {
		*path = append(*path, sankeyArc(quadrant, angle == sankeyUp, s.opts.Radius, geom.Pt{X: x - s.opts.Radius, Y: y + sign*s.opts.Radius})...)
	} else {
		*path = append(*path, sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x, Y: y}})
	}
	*path = append(
		*path,
		sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x, Y: y + sign*length}},
		sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x - s.opts.Shoulder, Y: y + sign*length}},
		sankeyPathStep{cmd: sankeyLineTo, pt: tip},
		sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x + s.opts.Shoulder - flow, Y: y + sign*length}},
		sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x - flow, Y: y + sign*length}},
	)
	*path = append(*path, sankeyArc(quadrant, angle == sankeyDown, s.opts.Radius-flow, geom.Pt{X: x - s.opts.Radius, Y: y + sign*s.opts.Radius})...)
	*path = append(*path, sankeyPathStep{cmd: sankeyLineTo, pt: geom.Pt{X: x - flow, Y: y + sign*flow}})
	return tip, geom.Pt{X: tip.X, Y: tip.Y + sign*s.opts.Offset}
}

func sankeyArc(quadrant int, cw bool, radius float64, center geom.Pt) []sankeyPathStep {
	base := []geom.Pt{
		{X: 1, Y: 0},
		{X: 1, Y: 0.265114773},
		{X: 0.894571235, Y: 0.519642327},
		{X: 0.707106781, Y: 0.707106781},
		{X: 0.519642327, Y: 0.894571235},
		{X: 0.265114773, Y: 1},
		{X: 0, Y: 1},
	}
	pts := make([]geom.Pt, len(base))
	for i, pt := range base {
		switch quadrant {
		case 0, 2:
			if cw {
				pts[i] = pt
			} else {
				pts[i] = geom.Pt{X: pt.Y, Y: pt.X}
			}
		default:
			if cw {
				pts[i] = geom.Pt{X: -pt.Y, Y: pt.X}
			} else {
				pts[i] = geom.Pt{X: -pt.X, Y: pt.Y}
			}
		}
		if quadrant > 1 {
			pts[i].X *= -radius
			pts[i].Y *= -radius
		} else {
			pts[i].X *= radius
			pts[i].Y *= radius
		}
		pts[i].X += center.X
		pts[i].Y += center.Y
	}
	out := make([]sankeyPathStep, len(pts))
	for i, pt := range pts {
		cmd := sankeyCurveTo
		if i == 0 {
			cmd = sankeyLineTo
		}
		out[i] = sankeyPathStep{cmd: cmd, pt: pt}
	}
	return out
}

func sankeyRevert(path []sankeyPathStep, first sankeyPathCmd) []sankeyPathStep {
	out := make([]sankeyPathStep, 0, len(path))
	next := first
	for i := len(path) - 1; i >= 0; i-- {
		out = append(out, sankeyPathStep{cmd: next, pt: path[i].pt})
		next = path[i].cmd
	}
	return out
}

func sankeyStepsToPath(steps []sankeyPathStep, center geom.Pt, rotation float64) geom.Path {
	var path geom.Path
	curves := make([]geom.Pt, 0, 3)
	flushCurve := func() {
		for len(curves) >= 3 {
			path.CubicTo(curves[0], curves[1], curves[2])
			curves = curves[3:]
		}
	}
	for _, step := range steps {
		pt := sankeyTransformPoint(step.pt, center, rotation)
		switch step.cmd {
		case sankeyMoveTo:
			flushCurve()
			path.MoveTo(pt)
		case sankeyLineTo:
			flushCurve()
			path.LineTo(pt)
		case sankeyCurveTo:
			curves = append(curves, pt)
			if len(curves) == 3 {
				flushCurve()
			}
		case sankeyClosePath:
			flushCurve()
			path.Close()
		}
	}
	flushCurve()
	return path
}

func sankeyTransformPoint(pt, center geom.Pt, rotation float64) geom.Pt {
	if rotation != 0 {
		rad := rotation * math.Pi / 180
		cosv, sinv := math.Cos(rad), math.Sin(rad)
		pt = geom.Pt{
			X: pt.X*cosv - pt.Y*sinv,
			Y: pt.X*sinv + pt.Y*cosv,
		}
	}
	return geom.Pt{X: pt.X + center.X, Y: pt.Y + center.Y}
}

func sankeyTrimLast(path []sankeyPathStep) []sankeyPathStep {
	if len(path) == 0 {
		return path
	}
	return path[:len(path)-1]
}

func sankeyLastPoint(path []sankeyPathStep) geom.Pt {
	if len(path) == 0 {
		return geom.Pt{}
	}
	return path[len(path)-1].pt
}

func sankeyExtent(vertices, labels []geom.Pt, show []bool) (geom.Rect, bool) {
	var rect geom.Rect
	have := false
	add := func(pt geom.Pt) {
		if math.IsNaN(pt.X) || math.IsNaN(pt.Y) || math.IsInf(pt.X, 0) || math.IsInf(pt.Y, 0) {
			return
		}
		if !have {
			rect = geom.Rect{Min: pt, Max: pt}
			have = true
			return
		}
		rect = expandRect(rect, pt)
	}
	for _, pt := range vertices {
		add(pt)
	}
	for i, pt := range labels {
		if i < len(show) && show[i] {
			add(pt)
		}
	}
	return rect, have
}

func (s *Sankey) expandExtent(rect geom.Rect) {
	if s == nil {
		return
	}
	if !s.haveExtent {
		s.extent = rect
		s.haveExtent = true
		return
	}
	s.extent = unionRect(s.extent, rect)
}

// Finish returns all diagrams added through the builder and applies Matplotlib's
// final Sankey axes limits/aspect adjustment for data-coordinate diagrams.
func (s *Sankey) Finish() []*SankeyDiagram {
	if s == nil {
		return nil
	}
	if s.ax != nil && s.haveExtent && isDataCoords(s.opts.Coords) {
		margin := s.opts.Margin
		s.ax.SetXLim(s.extent.Min.X-margin, s.extent.Max.X+margin)
		s.ax.SetYLim(s.extent.Min.Y-margin, s.extent.Max.Y+margin)
		_ = s.ax.SetAspect("equal")
	}
	return append([]*SankeyDiagram(nil), s.diagrams...)
}

// NewSankey creates a Matplotlib-style Sankey builder for one axes.
func NewSankey(ax *Axes, opts ...SankeyOptions) *Sankey {
	if ax == nil {
		return nil
	}
	cfg := SankeyOptions{
		Center:      geom.Pt{},
		Coords:      Coords(CoordData),
		Scale:       1,
		TrunkLength: 1,
		PathLength:  0.25,
		Gap:         0.25,
		Radius:      0.1,
		Shoulder:    0.03,
		Offset:      0.15,
		HeadAngle:   100,
		Margin:      0.4,
		Tolerance:   1e-6,
		Alpha:       1,
		TextColor:   ax.resolvedRC().DefaultTextColor(),
		FontSize:    ax.resolvedRC().FontSize,
	}
	if len(opts) > 0 {
		cfg = opts[0]
		if cfg.Scale <= 0 {
			cfg.Scale = 1
		}
		if cfg.TrunkLength <= 0 {
			cfg.TrunkLength = 1
		}
		if cfg.PathLength <= 0 {
			cfg.PathLength = 0.25
		}
		if cfg.Gap <= 0 {
			cfg.Gap = 0.25
		}
		if cfg.Radius <= 0 {
			cfg.Radius = 0.1
		}
		if cfg.Shoulder <= 0 {
			cfg.Shoulder = 0.03
		}
		if cfg.Offset <= 0 {
			cfg.Offset = 0.15
		}
		if cfg.HeadAngle <= 0 {
			cfg.HeadAngle = 100
		}
		if cfg.Margin <= 0 {
			cfg.Margin = 0.4
		}
		if cfg.Tolerance <= 0 {
			cfg.Tolerance = 1e-6
		}
		if cfg.Alpha <= 0 {
			cfg.Alpha = 1
		}
		if cfg.FontSize <= 0 {
			cfg.FontSize = ax.resolvedRC().FontSize
		}
		if cfg.TextColor == (render.Color{}) {
			cfg.TextColor = ax.resolvedRC().DefaultTextColor()
		}
	}
	pitch := math.Tan(math.Pi * (1 - cfg.HeadAngle/180) / 2)
	return &Sankey{
		ax:     ax,
		opts:   cfg,
		cursor: cfg.Center,
		pitch:  pitch,
	}
}
