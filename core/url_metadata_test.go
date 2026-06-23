package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// urlMarkerRenderer records the active url/gid at each Path call so tests can
// assert that drawArtist applies and restores artist metadata.
type urlMarkerRenderer struct {
	render.NullRenderer
	curURL, curGID string
	seenURL        []string
	seenGID        []string
}

func (r *urlMarkerRenderer) SetURL(url string) { r.curURL = url }
func (r *urlMarkerRenderer) URL() string       { return r.curURL }
func (r *urlMarkerRenderer) SetGID(gid string) { r.curGID = gid }
func (r *urlMarkerRenderer) GID() string       { return r.curGID }

func (r *urlMarkerRenderer) Path(_ geom.Path, _ *render.Paint) {
	r.seenURL = append(r.seenURL, r.curURL)
	r.seenGID = append(r.seenGID, r.curGID)
}

// urlTestArtist is a minimal artist that emits one Path on Draw and carries
// url/gid via the embedded ArtistRasterization.
type urlTestArtist struct {
	ArtistRasterization
}

func (a *urlTestArtist) Draw(r render.Renderer, _ *DrawContext) {
	r.Path(geom.Path{}, &render.Paint{})
}
func (a *urlTestArtist) Z() float64                    { return 0 }
func (a *urlTestArtist) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

func TestArtistRasterizationURLGIDRoundTrip(t *testing.T) {
	var a urlTestArtist
	if a.URL() != "" || a.GID() != "" {
		t.Fatalf("expected empty url/gid, got %q/%q", a.URL(), a.GID())
	}
	a.SetURL("https://example.com")
	a.SetGID("g1")
	if a.URL() != "https://example.com" || a.GID() != "g1" {
		t.Fatalf("round-trip failed: %q/%q", a.URL(), a.GID())
	}
}

func TestDrawArtistAppliesAndRestoresURLGID(t *testing.T) {
	r := &urlMarkerRenderer{}
	ctx := createTestDrawContext()

	linked := &urlTestArtist{}
	linked.SetURL("https://example.com/a")
	linked.SetGID("node-a")

	plain := &urlTestArtist{} // no metadata

	drawArtist(r, ctx, linked)
	drawArtist(r, ctx, plain)

	if len(r.seenURL) != 2 {
		t.Fatalf("expected 2 path calls, got %d", len(r.seenURL))
	}
	if r.seenURL[0] != "https://example.com/a" || r.seenGID[0] != "node-a" {
		t.Fatalf("metadata not applied to linked artist: %q/%q", r.seenURL[0], r.seenGID[0])
	}
	if r.seenURL[1] != "" || r.seenGID[1] != "" {
		t.Fatalf("metadata leaked to plain artist: %q/%q", r.seenURL[1], r.seenGID[1])
	}
	// After both draws the renderer state must be restored to empty.
	if r.curURL != "" || r.curGID != "" {
		t.Fatalf("renderer state not restored: %q/%q", r.curURL, r.curGID)
	}
}
