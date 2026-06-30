package core

import (
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func cloneRenderPathEffects(effects []render.PathEffect) []render.PathEffect {
	if len(effects) == 0 {
		return nil
	}
	out := append([]render.PathEffect(nil), effects...)
	for i := range out {
		if len(out[i].Dashes) > 0 {
			out[i].Dashes = append([]float64(nil), out[i].Dashes...)
		}
	}
	return out
}

// devicePathEffects clones path effects and converts their point-valued
// dimensions to device pixels, mirroring Matplotlib's points→pixels scaling at
// draw time. Stroke linewidth and ticked-stroke geometry are stored in points
// (matplotlib semantics); offsets are already in display pixels and pass
// through unchanged. Call this at the final device sink (where the paint is
// handed to the renderer), not at construction or intermediate stores.
func devicePathEffects(rc style.RC, effects []render.PathEffect) []render.PathEffect {
	out := cloneRenderPathEffects(effects)
	for i := range out {
		out[i].LineWidth = pointsToPixels(rc, out[i].LineWidth)
		out[i].TickSpacing = pointsToPixels(rc, out[i].TickSpacing)
		out[i].TickLength = pointsToPixels(rc, out[i].TickLength)
	}
	return out
}

// linePathEffects converts path effects to device pixels when a draw context
// (and thus a resolved DPI) is available, falling back to a plain clone when it
// is not.
func linePathEffects(ctx *DrawContext, effects []render.PathEffect) []render.PathEffect {
	if ctx == nil {
		return cloneRenderPathEffects(effects)
	}
	return devicePathEffects(ctx.RC, effects)
}
