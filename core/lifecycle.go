package core

// ArtistCallbackID identifies a registered artist lifecycle callback.
type ArtistCallbackID int64

// ArtistCallback is called when an artist marks itself stale.
type ArtistCallback func(Artist)

// ArtistLifecycle is an embeddable helper for Matplotlib-style stale state and
// callbacks. Concrete artists can opt in without forcing every Artist to carry
// mutable base-class state. It intentionally does not wire parent Figure/Axes
// propagation yet; see docs/adr/0002-artist-stale-state-scope.md.
type ArtistLifecycle struct {
	owner     Artist
	stale     bool
	next      ArtistCallbackID
	callbacks map[ArtistCallbackID]ArtistCallback
}

// BindArtist records the concrete artist passed to callbacks.
func (l *ArtistLifecycle) BindArtist(owner Artist) {
	if l == nil {
		return
	}
	l.owner = owner
}

// AddCallback registers a stale-state callback.
func (l *ArtistLifecycle) AddCallback(callback ArtistCallback) ArtistCallbackID {
	if l == nil || callback == nil {
		return 0
	}
	l.next++
	if l.callbacks == nil {
		l.callbacks = make(map[ArtistCallbackID]ArtistCallback)
	}
	l.callbacks[l.next] = callback
	return l.next
}

// RemoveCallback removes a registered stale-state callback.
func (l *ArtistLifecycle) RemoveCallback(id ArtistCallbackID) {
	if l == nil || id == 0 || l.callbacks == nil {
		return
	}
	delete(l.callbacks, id)
}

// MarkStale marks the artist stale and notifies callbacks.
func (l *ArtistLifecycle) MarkStale() {
	if l == nil {
		return
	}
	l.stale = true
	for _, callback := range l.callbacks {
		callback(l.owner)
	}
}

// ClearStale marks the artist clean after a draw or explicit synchronization.
func (l *ArtistLifecycle) ClearStale() {
	if l == nil {
		return
	}
	l.stale = false
}

// Stale reports whether the artist has been marked dirty since the last clear.
func (l *ArtistLifecycle) Stale() bool {
	if l == nil {
		return false
	}
	return l.stale
}

// StaleArtist is an optional extension for artists that expose dirty-state
// lifecycle hooks.
type StaleArtist interface {
	Stale() bool
	ClearStale()
	MarkStale()
}

// CallbackArtist is an optional extension for artists that expose lifecycle
// callback registration.
type CallbackArtist interface {
	AddCallback(ArtistCallback) ArtistCallbackID
	RemoveCallback(ArtistCallbackID)
}
