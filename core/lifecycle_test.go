package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type lifecycleArtist struct {
	ArtistLifecycle
}

func newLifecycleArtist() *lifecycleArtist {
	artist := &lifecycleArtist{}
	artist.BindArtist(artist)
	return artist
}

func (a *lifecycleArtist) Draw(render.Renderer, *DrawContext) {}

func (a *lifecycleArtist) Z() float64 { return 0 }

func (a *lifecycleArtist) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

func TestArtistLifecycleCallbacksAndStaleState(t *testing.T) {
	artist := newLifecycleArtist()
	called := 0
	id := artist.AddCallback(func(got Artist) {
		called++
		if got != artist {
			t.Fatalf("callback artist = %T, want lifecycle artist", got)
		}
	})

	artist.MarkStale()
	if !artist.Stale() {
		t.Fatal("artist not stale after MarkStale")
	}
	if called != 1 {
		t.Fatalf("called = %d, want 1", called)
	}

	artist.ClearStale()
	if artist.Stale() {
		t.Fatal("artist stale after ClearStale")
	}

	artist.RemoveCallback(id)
	artist.MarkStale()
	if called != 1 {
		t.Fatalf("called after removal = %d, want 1", called)
	}
}
