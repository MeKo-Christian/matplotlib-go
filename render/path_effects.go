package render

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

// DrawPathWithEffects applies paint.PathEffects using renderer-neutral path
// replay. It returns true when it consumed the draw; callers should otherwise
// continue with their normal direct path drawing.
func DrawPathWithEffects(_ Renderer, path geom.Path, paint *Paint, draw func(geom.Path, *Paint)) bool {
	if paint == nil || len(paint.PathEffects) == 0 || draw == nil {
		return false
	}

	base := clonePaint(*paint)
	base.PathEffects = nil

	for _, effect := range paint.PathEffects {
		switch effect.Kind {
		case PathEffectNormal:
			draw(path, &base)
		case PathEffectStroke:
			effectPaint := strokeEffectPaint(base, effect)
			draw(offsetPath(path, effect.Offset), &effectPaint)
		case PathEffectShadow:
			effectPaint := shadowEffectPaint(base, effect)
			draw(offsetPath(path, effect.Offset), &effectPaint)
		case PathEffectPathPatch:
			effectPaint := patchEffectPaint(base, effect)
			draw(offsetPath(path, effect.Offset), &effectPaint)
		case PathEffectTickedStroke:
			ticks := tickedStrokePath(offsetPath(path, effect.Offset), effect)
			if len(ticks.C) == 0 {
				continue
			}
			effectPaint := tickedStrokePaint(base, effect)
			draw(ticks, &effectPaint)
		case PathEffectFilter:
			// Filter-backed effects require backend-specific capture/replay.
			// Until a renderer consumes Filter natively, preserve the artist.
			draw(path, &base)
		default:
			draw(path, &base)
		}
	}
	return true
}

func strokeEffectPaint(base Paint, effect PathEffect) Paint {
	out := lineEffectPaint(base, effect)
	out.Fill = effect.Fill
	out.FillPattern = PatternFill{}
	out.FillGradient = GradientFill{}
	out.Hatch = ""
	out.HatchColor = Color{}
	return out
}

func patchEffectPaint(base Paint, effect PathEffect) Paint {
	out := base
	out.PathEffects = nil
	out.FillPattern = PatternFill{}
	out.FillGradient = GradientFill{}
	out.Hatch = ""
	out.HatchColor = Color{}
	out.Fill = effect.Fill
	out.Stroke = effect.Stroke
	if effect.LineWidth > 0 {
		out.LineWidth = effect.LineWidth
	}
	applyEffectLineState(&out, effect)
	applyEffectComposite(&out, effect)
	return out
}

func shadowEffectPaint(base Paint, effect PathEffect) Paint {
	out := base
	out.PathEffects = nil
	out.FillPattern = PatternFill{}
	out.FillGradient = GradientFill{}
	out.Hatch = ""
	out.HatchColor = Color{}
	out.Fill = Color{}
	out.Stroke = Color{}
	if effect.Fill.A > 0 {
		out.Fill = effect.Fill
	} else if effect.Stroke.A <= 0 && base.Fill.A > 0 {
		out.Fill = shadowColor(base.Fill, effect)
	}
	if effect.Stroke.A > 0 {
		out.Stroke = effect.Stroke
	} else if effect.Fill.A <= 0 && base.Stroke.A > 0 {
		out.Stroke = shadowColor(base.Stroke, effect)
	}
	if effect.LineWidth > 0 {
		out.LineWidth = effect.LineWidth
	}
	applyEffectLineState(&out, effect)
	applyEffectComposite(&out, effect)
	return out
}

func tickedStrokePaint(base Paint, effect PathEffect) Paint {
	out := lineEffectPaint(base, effect)
	out.Fill = Color{}
	out.FillPattern = PatternFill{}
	out.FillGradient = GradientFill{}
	out.Hatch = ""
	out.HatchColor = Color{}
	return out
}

func lineEffectPaint(base Paint, effect PathEffect) Paint {
	out := base
	out.PathEffects = nil
	if effect.Stroke.A > 0 {
		out.Stroke = effect.Stroke
	}
	if effect.LineWidth > 0 {
		out.LineWidth = effect.LineWidth
	}
	applyEffectLineState(&out, effect)
	applyEffectComposite(&out, effect)
	return out
}

func applyEffectLineState(out *Paint, effect PathEffect) {
	if effect.LineJoin != JoinMiter {
		out.LineJoin = effect.LineJoin
	}
	if effect.LineCap != CapButt {
		out.LineCap = effect.LineCap
	}
	if effect.MiterLimit > 0 {
		out.MiterLimit = effect.MiterLimit
	}
	if len(effect.Dashes) > 0 {
		out.Dashes = cloneFloat64s(effect.Dashes)
	}
}

func applyEffectComposite(out *Paint, effect PathEffect) {
	if effect.CompositeMode != CompositeSourceOver {
		out.CompositeMode = effect.CompositeMode
	}
}

func shadowColor(base Color, effect PathEffect) Color {
	alpha := effect.ShadowAlpha
	if alpha <= 0 {
		alpha = 0.3
	}
	rho := effect.ShadowRho
	if rho <= 0 {
		rho = 0.3
	}
	return Color{
		R: base.R * rho,
		G: base.G * rho,
		B: base.B * rho,
		A: alpha,
	}
}

func offsetPath(path geom.Path, offset geom.Pt) geom.Path {
	if offset == (geom.Pt{}) || len(path.V) == 0 {
		return path
	}
	out := geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: make([]geom.Pt, len(path.V)),
	}
	for i, pt := range path.V {
		out.V[i] = geom.Pt{X: pt.X + offset.X, Y: pt.Y + offset.Y}
	}
	return out
}

func tickedStrokePath(path geom.Path, effect PathEffect) geom.Path {
	spacing := effect.TickSpacing
	if spacing <= 0 {
		spacing = 10
	}
	length := effect.TickLength
	if length <= 0 {
		length = math.Sqrt2
	}
	angle := effect.TickAngle
	if angle == 0 {
		angle = 45
	}
	segments := flattenPath(path)
	if len(segments) == 0 {
		return geom.Path{}
	}

	theta := -angle * math.Pi / 180
	cosTheta := math.Cos(theta)
	sinTheta := math.Sin(theta)
	tickLength := length * spacing

	out := geom.Path{}
	for _, poly := range segments {
		if len(poly) < 2 {
			continue
		}
		dist := make([]float64, len(poly))
		total := 0.0
		for i := 1; i < len(poly); i++ {
			total += math.Hypot(poly[i].X-poly[i-1].X, poly[i].Y-poly[i-1].Y)
			dist[i] = total
		}
		if total <= spacing {
			continue
		}
		for s := spacing / 2; s < total; s += spacing {
			pt, dir, ok := interpolatePolyline(poly, dist, s)
			if !ok {
				continue
			}
			end := geom.Pt{
				X: pt.X + (dir.X*cosTheta-dir.Y*sinTheta)*tickLength,
				Y: pt.Y + (dir.X*sinTheta+dir.Y*cosTheta)*tickLength,
			}
			out.MoveTo(pt)
			out.LineTo(end)
		}
	}
	return out
}

func interpolatePolyline(poly []geom.Pt, dist []float64, s float64) (geom.Pt, geom.Pt, bool) {
	for i := 1; i < len(poly); i++ {
		if s > dist[i] {
			continue
		}
		segLen := dist[i] - dist[i-1]
		if segLen <= 0 {
			continue
		}
		t := (s - dist[i-1]) / segLen
		a, b := poly[i-1], poly[i]
		pt := geom.Pt{
			X: a.X + (b.X-a.X)*t,
			Y: a.Y + (b.Y-a.Y)*t,
		}
		dir := geom.Pt{
			X: (b.X - a.X) / segLen,
			Y: (b.Y - a.Y) / segLen,
		}
		return pt, dir, true
	}
	return geom.Pt{}, geom.Pt{}, false
}

func flattenPath(path geom.Path) [][]geom.Pt {
	var out [][]geom.Pt
	var current []geom.Pt
	var cursor, start geom.Pt
	haveCursor := false
	vi := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if len(current) > 0 {
				out = append(out, current)
			}
			if vi >= len(path.V) {
				return out
			}
			cursor = path.V[vi]
			start = cursor
			haveCursor = true
			current = []geom.Pt{cursor}
			vi++
		case geom.LineTo:
			if vi >= len(path.V) || !haveCursor {
				return out
			}
			cursor = path.V[vi]
			current = append(current, cursor)
			vi++
		case geom.QuadTo:
			if vi+1 >= len(path.V) || !haveCursor {
				return out
			}
			c, end := path.V[vi], path.V[vi+1]
			for i := 1; i <= 12; i++ {
				t := float64(i) / 12
				current = append(current, quadPoint(cursor, c, end, t))
			}
			cursor = end
			vi += 2
		case geom.CubicTo:
			if vi+2 >= len(path.V) || !haveCursor {
				return out
			}
			c1, c2, end := path.V[vi], path.V[vi+1], path.V[vi+2]
			for i := 1; i <= 16; i++ {
				t := float64(i) / 16
				current = append(current, cubicPoint(cursor, c1, c2, end, t))
			}
			cursor = end
			vi += 3
		case geom.ClosePath:
			if haveCursor {
				current = append(current, start)
				cursor = start
			}
		}
	}
	if len(current) > 0 {
		out = append(out, current)
	}
	return out
}
