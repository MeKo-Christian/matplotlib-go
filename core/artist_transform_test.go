package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

type artistTransformRecordingRenderer struct {
	render.NullRenderer
	paths []geom.Path
}

func (r *artistTransformRecordingRenderer) Path(path geom.Path, _ *render.Paint) {
	r.paths = append(r.paths, path)
}

func TestArtistTransformMetadataDefaultsAndStale(t *testing.T) {
	var metadata ArtistRasterization
	got := metadata.ArtistTransform()
	if got.HasTransform || got.Transform != nil || got.HasCoords {
		t.Fatalf("zero-value artist transform = %+v, want no explicit transform metadata", got)
	}

	metadata.SetTransformCoords(Coords(CoordAxes))
	got = metadata.ArtistTransform()
	if !got.HasCoords || got.Coords != Coords(CoordAxes) {
		t.Fatalf("artist coords transform = %+v, want axes coords", got)
	}
	if !metadata.Stale() {
		t.Fatal("transform coordinate metadata should mark artist stale")
	}
	metadata.SetStale(false)

	explicit := transform.NewAffine(geom.Affine{A: 2, D: 3, E: 4, F: 5})
	metadata.SetTransform(explicit)
	got = metadata.ArtistTransform()
	if !got.HasTransform || got.Transform == nil || got.Transform.Apply(geom.Pt{X: 1, Y: 1}) != (geom.Pt{X: 6, Y: 8}) {
		t.Fatalf("artist explicit transform = %+v, want affine display transform", got)
	}
	if !metadata.Stale() {
		t.Fatal("explicit transform metadata should mark artist stale")
	}

	metadata.ClearTransform()
	metadata.ClearTransformCoords()
	got = metadata.ArtistTransform()
	if got.HasTransform || got.Transform != nil || got.HasCoords {
		t.Fatalf("cleared artist transform = %+v, want no explicit transform metadata", got)
	}
}

func TestLine2DUsesArtistCoordinateTransform(t *testing.T) {
	ctx := transformTestContext()
	line := &Line2D{
		XY:  []geom.Pt{{X: 0.25, Y: 0.75}, {X: 0.5, Y: 0.25}},
		W:   1,
		Col: render.Color{A: 1},
	}
	line.SetTransformCoords(Coords(CoordAxes))

	ren := &artistTransformRecordingRenderer{}
	line.Draw(ren, ctx)

	if len(ren.paths) != 1 {
		t.Fatalf("path call count = %d, want 1", len(ren.paths))
	}
	if got := ren.paths[0].V[0]; got != (geom.Pt{X: 25, Y: 25}) {
		t.Fatalf("line first point = %+v, want axes-coordinate display point", got)
	}
	if got := ren.paths[0].V[1]; got != (geom.Pt{X: 50, Y: 75}) {
		t.Fatalf("line second point = %+v, want axes-coordinate display point", got)
	}
}

func TestPathPatchUsesExplicitArtistTransform(t *testing.T) {
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 1, Y: 2})
	path.LineTo(geom.Pt{X: 3, Y: 4})
	patch := &PathPatch{
		Patch: Patch{
			FaceColor: render.Color{A: 1},
		},
		Path:   path,
		Coords: Coords(CoordData),
	}
	patch.SetTransform(transform.NewAffine(geom.Affine{A: 2, D: 3, E: 4, F: 5}))

	ren := &artistTransformRecordingRenderer{}
	patch.Draw(ren, transformTestContext())

	if len(ren.paths) != 1 {
		t.Fatalf("path call count = %d, want 1", len(ren.paths))
	}
	if got := ren.paths[0].V[0]; got != (geom.Pt{X: 6, Y: 11}) {
		t.Fatalf("patch first point = %+v, want explicit transform output", got)
	}
	if got := ren.paths[0].V[1]; got != (geom.Pt{X: 10, Y: 17}) {
		t.Fatalf("patch second point = %+v, want explicit transform output", got)
	}
}

func TestScatterCopiesArtistTransformToPathCollection(t *testing.T) {
	ctx := transformTestContext()
	scatter := &Scatter2D{
		XY:        []geom.Pt{{X: 0.2, Y: 0.8}},
		Size:      36,
		Color:     render.Color{A: 1},
		Marker:    MarkerCircle,
		EdgeWidth: 0,
	}
	scatter.SetTransformCoords(Coords(CoordFigure))

	collection := scatter.toPathCollection(&render.NullRenderer{}, ctx)
	tr := artistTransformFor(ctx, collection, collection.Coords)
	if tr == nil {
		t.Fatal("scatter path collection did not retain artist transform metadata")
	}
	if got := tr.Apply(collection.Offsets[0]); got != (geom.Pt{X: 40, Y: 40}) {
		t.Fatalf("scatter transformed offset = %+v, want figure-coordinate display point", got)
	}
}

func transformTestContext() *DrawContext {
	axesRect := geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 100, Y: 100}}
	figureRect := geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 200, Y: 200}}
	return &DrawContext{
		DataToPixel: Transform2D{
			DataToAxes:  transform.NewScaleTransform(transform.NewLinear(0, 10), transform.NewLinear(0, 10)),
			AxesToPixel: transform.NewDisplayRectTransform(axesRect),
		},
		Clip:       axesRect,
		FigureRect: figureRect,
	}
}
