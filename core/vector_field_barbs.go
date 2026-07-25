package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/optarg"
	"github.com/cwbudde/matplotlib-go/render"
)

// Barbs adds a wind-barb artist to the axes.
func (a *Axes) Barbs(x, y, u, v []float64, opts ...BarbsOptions) *Barbs {
	if a == nil {
		return nil
	}
	opt, supplied := optarg.Optional("barbs", opts)
	anchors, uu, vv, scalars, ok := flattenVectorSamples(x, y, u, v, barbsScalarOptions(opt, supplied))
	if !ok || len(anchors) == 0 {
		return nil
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
	supplied, ok := optarg.Optional("barbs grid", opts)
	anchors, uu, vv, scalars, valid := flattenVectorGrid(x, y, u, v, barbsScalarOptions(supplied, ok))
	if !valid {
		return nil
	}

	var opt BarbsOptions
	if ok {
		opt = supplied
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
	if normalizeVectorUnits(b.Units, "points") == "points" {
		lengthPx *= b.Length / 2
	}
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

		angle := math.Atan2(v, u)
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
	dir := 1.0
	if flip {
		dir = -1.0
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
	path.Close()

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
