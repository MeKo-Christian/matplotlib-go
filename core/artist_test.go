package core

import (
    "testing"
    "matplotlib-go/internal/geom"
    "matplotlib-go/render"
    "matplotlib-go/style"
)

func TestRCEffectivePrecedence(t *testing.T) {
    fig := NewFigure(800, 600, style.WithDPI(110))
    axInherit := fig.AddAxes(geom.Rect{Min: geom.Pt{0.1, 0.1}, Max: geom.Pt{0.9, 0.9}})
    if got := axInherit.effectiveRC(fig).DPI; got != 110 {
        t.Fatalf("expected axes inherit figure DPI, got %v", got)
    }
    axOverride := fig.AddAxes(geom.Rect{Min: geom.Pt{0.2, 0.2}, Max: geom.Pt{0.8, 0.8}}, style.WithDPI(200))
    if got := axOverride.effectiveRC(fig).DPI; got != 200 {
        t.Fatalf("expected axes override DPI=200, got %v", got)
    }
}

// artist with custom z
type zArtist struct{ z float64; id int; hit *[]int }
func (a zArtist) Draw(_ render.Renderer, _ *DrawContext) { *a.hit = append(*a.hit, a.id) }
func (a zArtist) Z() float64 { return a.z }
func (a zArtist) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

func TestZOrderStableSortAndTraversal(t *testing.T) {
    fig := NewFigure(100, 100)
    ax := fig.AddAxes(geom.Rect{Min: geom.Pt{0,0}, Max: geom.Pt{1,1}})
    var order []int
    // Insertion order: ids 1..5, with equal z for 2 and 3
    ax.Add(zArtist{z: 0, id: 1, hit: &order})
    ax.Add(zArtist{z: 1, id: 2, hit: &order})
    ax.Add(zArtist{z: 1, id: 3, hit: &order})
    ax.Add(zArtist{z: -1, id: 4, hit: &order})
    ax.Add(zArtist{z: 2, id: 5, hit: &order})

    var r render.NullRenderer
    DrawFigure(fig, &r)

    // Expected draw order: z=-1 (id4), z=0 (id1), z=1 (ids 2 then 3), z=2 (id5)
    want := []int{4,1,2,3,5}
    if len(order) != len(want) { t.Fatalf("draw count mismatch: got %d want %d", len(order), len(want)) }
    for i := range want {
        if order[i] != want[i] {
            t.Fatalf("order mismatch at %d: got %v want %v (full=%v)", i, order[i], want[i], order)
        }
    }
}

