package canvas

import (
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
)

// PickResult is one hit from a [Pick] call.
type PickResult struct {
	Axes   *Axes
	Artist core.Artist
	Info   core.PickInfo
}

// Pick walks every axes in the figure and returns artists that contain p
// (figure pixels). Results are returned in front-to-back order so the first
// element is the topmost artist under the cursor. Only artists implementing
// [core.ArtistPicker] participate; the others are silently skipped.
//
// This mirrors Matplotlib's pick_event collection. Backends typically take the
// first hit and translate it into a [PickEvent].
func Pick(fig *Figure, p geom.Pt) []PickResult {
	if fig == nil {
		return nil
	}

	var hits []PickResult
	// Figure-level artists draw last, so they hit first.
	for i := len(fig.Artists) - 1; i >= 0; i-- {
		art := fig.Artists[i]
		if hit, info, ok := pickArtist(art, p, core.FigureDrawContext(fig)); ok {
			hits = append(hits, PickResult{Artist: hit, Info: info})
		}
	}

	for i := len(fig.Children) - 1; i >= 0; i-- {
		ax := fig.Children[i]
		if ax == nil || !ax.ContainsDisplayPoint(p) {
			continue
		}
		ctx := core.AxesDrawContext(ax, fig)
		artists := ax.Artists
		for j := len(artists) - 1; j >= 0; j-- {
			art := artists[j]
			if hit, info, ok := pickArtist(art, p, ctx); ok {
				hits = append(hits, PickResult{Axes: ax, Artist: hit, Info: info})
			}
		}
	}
	return hits
}

func pickArtist(art core.Artist, p geom.Pt, ctx *core.DrawContext) (core.Artist, core.PickInfo, bool) {
	picker, ok := art.(core.ArtistPicker)
	if !ok {
		return nil, core.PickInfo{}, false
	}
	hit, info := picker.Contains(p, ctx)
	if !hit {
		return nil, core.PickInfo{}, false
	}
	return art, info, true
}

// EmitPick performs a [Pick] and dispatches the topmost hit, if any, as a
// [PickEvent] through dispatcher.
func EmitPick(dispatcher *Dispatcher, fig *Figure, mouse MouseEvent) (PickResult, bool) {
	if dispatcher == nil {
		return PickResult{}, false
	}
	hits := Pick(fig, mouse.Position)
	if len(hits) == 0 {
		return PickResult{}, false
	}
	top := hits[0]
	event := NewPickEvent(fig, top.Artist, mouse)
	event.Axes = top.Axes
	_ = dispatcher.Emit(event.Event)
	return top, true
}
