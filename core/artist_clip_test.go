package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type clipTestArtist struct {
	ArtistRasterization
	draw func(render.Renderer)
}

func (a *clipTestArtist) Draw(r render.Renderer, _ *DrawContext) {
	if a.draw != nil {
		a.draw(r)
	}
}

func (a *clipTestArtist) Z() float64 { return 0 }

func (a *clipTestArtist) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

func (a *clipTestArtist) DrawOverlay(r render.Renderer, _ *DrawContext) {
	if a.draw != nil {
		a.draw(r)
	}
}

type clipRecordingRenderer struct {
	render.NullRenderer
	events     []string
	rects      []geom.Rect
	paths      []geom.Path
	transforms []geom.Affine
}

func (r *clipRecordingRenderer) Save() {
	r.events = append(r.events, "save")
	r.NullRenderer.Save()
}

func (r *clipRecordingRenderer) Restore() {
	r.events = append(r.events, "restore")
	r.NullRenderer.Restore()
}

func (r *clipRecordingRenderer) ClipRect(rect geom.Rect) {
	r.events = append(r.events, "clipRect")
	r.rects = append(r.rects, rect)
	r.NullRenderer.ClipRect(rect)
}

func (r *clipRecordingRenderer) ClipPath(path geom.Path) {
	r.events = append(r.events, "clipPath")
	r.paths = append(r.paths, path)
	r.NullRenderer.ClipPath(path)
}

type transformedClipRecordingRenderer struct {
	clipRecordingRenderer
}

func (r *transformedClipRecordingRenderer) ClipPathTransformed(path geom.Path, transform geom.Affine) {
	r.events = append(r.events, "clipPathTransformed")
	r.paths = append(r.paths, path)
	r.transforms = append(r.transforms, transform)
}

func TestArtistClipMetadataDefaultsAndClonesPath(t *testing.T) {
	var metadata ArtistRasterization
	if !metadata.ClipOn() {
		t.Fatal("zero-value artist clip metadata should be enabled")
	}
	if got := metadata.ArtistClip(); !got.ClipOn || got.HasClipRect || got.HasClipPath || got.HasClipPathTrans {
		t.Fatalf("zero-value artist clip = %+v, want clip-on with no explicit clips", got)
	}

	metadata.SetClipOn(false)
	if metadata.ClipOn() {
		t.Fatal("SetClipOn(false) did not disable explicit artist clipping")
	}
	if !metadata.Stale() {
		t.Fatal("clip metadata changes should mark the artist stale")
	}
	metadata.SetStale(false)

	rect := geom.Rect{Min: geom.Pt{X: 1, Y: 2}, Max: geom.Pt{X: 3, Y: 4}}
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 10, Y: 20})
	path.LineTo(geom.Pt{X: 30, Y: 40})
	transform := geom.Affine{A: 2, D: 3, E: 4, F: 5}

	metadata.SetClipOn(true)
	metadata.SetClipRect(rect)
	metadata.SetClipPath(path)
	metadata.SetClipPathTransform(transform)
	path.V[0].X = 999

	got := metadata.ArtistClip()
	if !got.ClipOn || !got.HasClipRect || !got.HasClipPath || !got.HasClipPathTrans {
		t.Fatalf("artist clip = %+v, want all explicit clip metadata set", got)
	}
	if got.ClipRect != rect || got.ClipPathTransform != transform {
		t.Fatalf("artist clip = %+v, want rect %v and transform %v", got, rect, transform)
	}
	if got.ClipPath.V[0].X == 999 {
		t.Fatal("SetClipPath retained caller-owned path storage")
	}

	got.ClipPath.V[0].X = 777
	if metadata.ArtistClip().ClipPath.V[0].X == 777 {
		t.Fatal("ArtistClip returned mutable internal path storage")
	}

	metadata.ClearClipRect()
	metadata.ClearClipPath()
	got = metadata.ArtistClip()
	if got.HasClipRect || got.HasClipPath || got.HasClipPathTrans {
		t.Fatalf("cleared artist clip = %+v, want no explicit clip metadata", got)
	}
}

func TestDrawArtistAppliesArtistClipRectAndPath(t *testing.T) {
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 1, Y: 1})
	path.LineTo(geom.Pt{X: 2, Y: 2})
	rect := geom.Rect{Min: geom.Pt{X: 5, Y: 6}, Max: geom.Pt{X: 7, Y: 8}}

	art := &clipTestArtist{
		draw: func(r render.Renderer) {
			r.(*clipRecordingRenderer).events = append(r.(*clipRecordingRenderer).events, "draw")
		},
	}
	art.SetClipRect(rect)
	art.SetClipPath(path)

	ren := &clipRecordingRenderer{}
	drawArtist(ren, &DrawContext{}, art)

	wantEvents := []string{"save", "clipRect", "clipPath", "draw", "restore"}
	assertStringSliceEqual(t, ren.events, wantEvents)
	if len(ren.rects) != 1 || ren.rects[0] != rect {
		t.Fatalf("clip rects = %v, want [%v]", ren.rects, rect)
	}
	if len(ren.paths) != 1 || len(ren.paths[0].V) != len(path.V) || ren.paths[0].V[0] != path.V[0] {
		t.Fatalf("clip paths = %+v, want path %+v", ren.paths, path)
	}
}

func TestDrawOverlayArtistAppliesArtistClip(t *testing.T) {
	rect := geom.Rect{Min: geom.Pt{X: 1, Y: 2}, Max: geom.Pt{X: 3, Y: 4}}
	art := &clipTestArtist{
		draw: func(r render.Renderer) {
			r.(*clipRecordingRenderer).events = append(r.(*clipRecordingRenderer).events, "overlay")
		},
	}
	art.SetClipRect(rect)

	ren := &clipRecordingRenderer{}
	drawOverlayArtist(ren, &DrawContext{}, art, art)

	wantEvents := []string{"save", "clipRect", "overlay", "restore"}
	assertStringSliceEqual(t, ren.events, wantEvents)
}

func TestDrawOverlayArtistWithoutArtistClipRemainsUnclipped(t *testing.T) {
	art := &clipTestArtist{
		draw: func(r render.Renderer) {
			r.(*clipRecordingRenderer).events = append(r.(*clipRecordingRenderer).events, "overlay")
		},
	}

	ren := &clipRecordingRenderer{}
	drawOverlayArtist(ren, &DrawContext{}, art, art)

	assertStringSliceEqual(t, ren.events, []string{"overlay"})
}

func TestArtistClipPathUsesTransformCapability(t *testing.T) {
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 1, Y: 1})
	path.LineTo(geom.Pt{X: 2, Y: 2})
	transform := geom.Affine{A: 2, D: 3, E: 4, F: 5}
	art := &clipTestArtist{
		draw: func(r render.Renderer) {
			r.(*transformedClipRecordingRenderer).events = append(r.(*transformedClipRecordingRenderer).events, "draw")
		},
	}
	art.SetClipPath(path)
	art.SetClipPathTransform(transform)

	ren := &transformedClipRecordingRenderer{}
	drawArtist(ren, &DrawContext{}, art)

	wantEvents := []string{"save", "clipPathTransformed", "draw", "restore"}
	assertStringSliceEqual(t, ren.events, wantEvents)
	if len(ren.transforms) != 1 || ren.transforms[0] != transform {
		t.Fatalf("clip path transforms = %v, want [%v]", ren.transforms, transform)
	}
}

func TestArtistClipPathTransformFallsBackToTransformedPath(t *testing.T) {
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 1, Y: 1})
	path.LineTo(geom.Pt{X: 2, Y: 2})
	transform := geom.Affine{A: 2, D: 3, E: 4, F: 5}
	art := &clipTestArtist{
		draw: func(r render.Renderer) {
			r.(*clipRecordingRenderer).events = append(r.(*clipRecordingRenderer).events, "draw")
		},
	}
	art.SetClipPath(path)
	art.SetClipPathTransform(transform)

	ren := &clipRecordingRenderer{}
	drawArtist(ren, &DrawContext{}, art)

	wantEvents := []string{"save", "clipPath", "draw", "restore"}
	assertStringSliceEqual(t, ren.events, wantEvents)
	if len(ren.paths) != 1 {
		t.Fatalf("clip path count = %d, want 1", len(ren.paths))
	}
	wantFirst := transform.Apply(path.V[0])
	if ren.paths[0].V[0] != wantFirst {
		t.Fatalf("transformed first clip point = %v, want %v", ren.paths[0].V[0], wantFirst)
	}
}

func TestArtistClipOnFalseSkipsExplicitArtistClip(t *testing.T) {
	rect := geom.Rect{Min: geom.Pt{X: 1, Y: 2}, Max: geom.Pt{X: 3, Y: 4}}
	art := &clipTestArtist{
		draw: func(r render.Renderer) {
			r.(*clipRecordingRenderer).events = append(r.(*clipRecordingRenderer).events, "draw")
		},
	}
	art.SetClipRect(rect)
	art.SetClipOn(false)

	ren := &clipRecordingRenderer{}
	drawArtist(ren, &DrawContext{}, art)

	assertStringSliceEqual(t, ren.events, []string{"draw"})
}

func assertStringSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("events = %v, want %v", got, want)
		}
	}
}
