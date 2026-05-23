package webagg

import "github.com/cwbudde/matplotlib-go/core"

type staleRegistration struct {
	artist core.CallbackArtist
	id     core.ArtistCallbackID
}

func (m *Manager) attachStaleCallbacks() {
	forEachFigureArtist(m.figure, func(artist core.Artist) {
		callbackArtist, ok := artist.(core.CallbackArtist)
		if !ok {
			return
		}
		id := callbackArtist.AddCallback(func(core.Artist) {
			_ = m.DrawIdle()
		})
		if id == 0 {
			return
		}
		m.staleRegs = append(m.staleRegs, staleRegistration{artist: callbackArtist, id: id})
	})
}

func (m *Manager) detachStaleCallbacks() {
	m.mu.Lock()
	regs := append([]staleRegistration(nil), m.staleRegs...)
	m.staleRegs = nil
	m.mu.Unlock()
	for _, reg := range regs {
		if reg.artist != nil {
			reg.artist.RemoveCallback(reg.id)
		}
	}
}

func clearStaleArtists(fig *core.Figure) {
	forEachFigureArtist(fig, func(artist core.Artist) {
		if stale, ok := artist.(core.StaleArtist); ok {
			stale.ClearStale()
		}
	})
}

func forEachFigureArtist(fig *core.Figure, visit func(core.Artist)) {
	if fig == nil || visit == nil {
		return
	}
	for _, artist := range fig.Artists {
		if artist != nil {
			visit(artist)
		}
	}
	for _, ax := range fig.Children {
		if ax == nil {
			continue
		}
		for _, artist := range ax.Artists {
			if artist != nil {
				visit(artist)
			}
		}
	}
}
