package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/transform"
)

// pathsBitEqual reports byte-level path equality: vertices are compared by IEEE
// bit pattern so a NaN matches a NaN at the same position (NaN != NaN under ==).
func pathsBitEqual(a, b geom.Path) bool {
	if len(a.C) != len(b.C) || len(a.V) != len(b.V) {
		return false
	}
	for i := range a.C {
		if a.C[i] != b.C[i] {
			return false
		}
	}
	for i := range a.V {
		if math.Float64bits(a.V[i].X) != math.Float64bits(b.V[i].X) ||
			math.Float64bits(a.V[i].Y) != math.Float64bits(b.V[i].Y) {
			return false
		}
	}
	return true
}

// TestBuildArtistDisplayPathCacheParity asserts the cached patch/collection draw
// path is byte-identical to the direct per-vertex transform across affine and
// non-affine legs. The reference is applyTransformPath on the same prepared
// source path and resolved transform (the historical algorithm).
func TestBuildArtistDisplayPathCacheParity(t *testing.T) {
	rect := geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 410, Y: 320}}

	linear := transform.NewScaleTransform(transform.NewLinear(0, 4), transform.NewLinear(0, 3))
	logLeg := transform.NewScaleTransform(transform.NewLog(1, 1000, 10), transform.NewLog(1, 1000, 10))
	counting := countingTransform{}

	legs := []struct {
		name string
		leg  transform.T
	}{
		{"linear-affine", linear},
		{"log-nonaffine", logLeg},
		{"counting-nonaffine", counting},
	}

	shapes := []struct {
		name          string
		local         geom.Path
		localToCoords geom.Affine
	}{
		{"rect-rotated", rectanglePath(2, 1.5), patchAffine(geom.Pt{X: 1, Y: 1}, 30)},
		{"polygon", polygonPath([]geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}}, true), geom.Identity()},
		{"path-bezier", roundedRectPath(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 3, Y: 2}}, 0.5), geom.Identity()},
	}

	for _, leg := range legs {
		for _, sh := range shapes {
			t.Run(leg.name+"/"+sh.name, func(t *testing.T) {
				_, ctx := newCacheTestAxesContext(leg.leg, rect)
				// A Patch-embedding artist drawing in data coordinates routes
				// through the shared cache; a fresh instance starts uncached.
				r := &PathPatch{Path: sh.local, Coords: Coords(CoordData)}

				got := buildArtistDisplayPath(ctx, r, r.Coords, sh.local, sh.localToCoords)
				tr := artistTransformFor(ctx, r, r.Coords)
				want := applyTransformPath(applyAffinePath(sh.local, sh.localToCoords), tr)

				if !pathsBitEqual(got, want) {
					t.Fatalf("cached path diverged from direct path\n got: %+v\nwant: %+v", got, want)
				}
			})
		}
	}
}

// TestBuildArtistDisplayPathReusesProjection drives a Patch-embedding artist
// through the full every-draw refreshDataTransform with a non-affine leg, then
// resizes the bbox and redraws with a freshly-rebuilt (value-equal) source. The
// projection must be reused (Apply count unchanged) because the leg is unchanged
// and the source compares equal, while the output reflects the new affine.
func TestBuildArtistDisplayPathReusesProjection(t *testing.T) {
	applies := 0
	counting := countingTransform{applies: &applies}
	ax, ctx := newCacheTestAxesContext(counting, geom.Rect{Max: geom.Pt{X: 100, Y: 100}})

	local := rectanglePath(2, 2)
	r := &Rectangle{Width: 2, Height: 2, Coords: Coords(CoordData)}
	n := len(applyAffinePath(local, geom.Identity()).V)

	first := buildArtistDisplayPath(ctx, r, r.Coords, rectanglePath(2, 2), geom.Identity())
	if applies != n {
		t.Fatalf("first draw: non-affine pass ran %d times, want %d", applies, n)
	}

	// Resize the bbox and re-run refreshDataTransform with an unchanged leg, then
	// rebuild the source fresh exactly as a real draw does.
	ax.axesBbox.Set(geom.Rect{Max: geom.Pt{X: 200, Y: 150}})
	ax.refreshDataTransform(counting)
	second := buildArtistDisplayPath(ctx, r, r.Coords, rectanglePath(2, 2), geom.Identity())

	if applies != n {
		t.Fatalf("after resize + refresh: non-affine pass ran %d times total, want %d (projection should be reused)", applies, n)
	}
	if pathsEqualExact(first, second) {
		t.Fatal("resized redraw produced an identical path; trailing affine was not refreshed")
	}

	// A genuine source change must re-project.
	_ = buildArtistDisplayPath(ctx, r, r.Coords, rectanglePath(3, 3), geom.Identity())
	if applies != n+n {
		t.Fatalf("after source change: non-affine pass ran %d times total, want %d", applies, n+n)
	}
}

// TestCollectionPerElementCacheReuse asserts each collection element keeps an
// independent projection cache: a redraw with freshly-rebuilt value-equal element
// sources (under an unchanged non-affine leg, resized bbox) reuses every
// element's projection.
func TestCollectionPerElementCacheReuse(t *testing.T) {
	applies := 0
	counting := countingTransform{applies: &applies}
	ax, ctx := newCacheTestAxesContext(counting, geom.Rect{Max: geom.Pt{X: 100, Y: 100}})

	c := &PathCollection{Collection: Collection{Coords: Coords(CoordData)}}
	elem0 := func() geom.Path { return polygonPath([]geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}}, true) }
	elem1 := func() geom.Path {
		return polygonPath([]geom.Pt{{X: 2, Y: 2}, {X: 3, Y: 2}, {X: 3, Y: 3}, {X: 2, Y: 3}}, true)
	}
	n0 := len(elem0().V)
	n1 := len(elem1().V)

	buildCachedDisplayPath(ctx, c.pathCacheSlot(0), c, c.Coords, elem0(), geom.Identity())
	buildCachedDisplayPath(ctx, c.pathCacheSlot(1), c, c.Coords, elem1(), geom.Identity())
	if applies != n0+n1 {
		t.Fatalf("first draw: non-affine pass ran %d times, want %d", applies, n0+n1)
	}

	// Redraw: resize + unchanged leg + freshly-rebuilt equal sources → both reuse.
	ax.axesBbox.Set(geom.Rect{Max: geom.Pt{X: 200, Y: 150}})
	ax.refreshDataTransform(counting)
	buildCachedDisplayPath(ctx, c.pathCacheSlot(0), c, c.Coords, elem0(), geom.Identity())
	buildCachedDisplayPath(ctx, c.pathCacheSlot(1), c, c.Coords, elem1(), geom.Identity())
	if applies != n0+n1 {
		t.Fatalf("after resize redraw: non-affine pass ran %d times total, want %d (per-element projections should be reused)", applies, n0+n1)
	}

	// Changing only element 0's source must re-project element 0 alone.
	buildCachedDisplayPath(ctx, c.pathCacheSlot(0), c, c.Coords, polygonPath([]geom.Pt{{X: 0, Y: 0}, {X: 5, Y: 0}, {X: 5, Y: 5}}, true), geom.Identity())
	buildCachedDisplayPath(ctx, c.pathCacheSlot(1), c, c.Coords, elem1(), geom.Identity())
	if applies != n0+n1+n0 {
		t.Fatalf("after element-0 source change: non-affine pass ran %d times total, want %d (only element 0 reprojects)", applies, n0+n1+n0)
	}
}
