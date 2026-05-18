package core

import "github.com/cwbudde/matplotlib-go/render"

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
