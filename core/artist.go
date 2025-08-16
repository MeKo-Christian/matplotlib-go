package core

import (
    "sort"

    "matplotlib-go/internal/geom"
    "matplotlib-go/render"
    "matplotlib-go/style"
    "matplotlib-go/transform"
)

// Artist is anything that can draw itself with a z-order and optional bounds.
type Artist interface {
    Draw(r render.Renderer, ctx *DrawContext)
    Z() float64
    Bounds(ctx *DrawContext) geom.Rect
}

// ArtistFunc adapts a function to an Artist.
type ArtistFunc func(r render.Renderer, ctx *DrawContext)

func (f ArtistFunc) Draw(r render.Renderer, ctx *DrawContext) { f(r, ctx) }
func (f ArtistFunc) Z() float64                                  { return 0 }
func (f ArtistFunc) Bounds(_ *DrawContext) geom.Rect              { return geom.Rect{} }

// DrawContext carries per-draw state like transforms and style.
type DrawContext struct {
    // DataToPixel maps data coordinates to pixels.
    DataToPixel Transform2D
    // Styling configuration in effect.
    RC style.RC
    // Clip is the axes pixel rectangle.
    Clip geom.Rect
}

// Transform2D wires x/y scales with an axes->pixel affine transform.
type Transform2D struct {
    XScale       transform.Scale
    YScale       transform.Scale
    AxesToPixel  transform.AffineT
}

// Figure is the root of the Artist tree. It contains Axes children.
type Figure struct {
    SizePx  geom.Pt
    RC      style.RC
    Children []*Axes
}

// NewFigure creates a new figure with pixel dimensions and optional style overrides.
func NewFigure(w, h int, opts ...style.Option) *Figure {
    rc := style.Apply(style.Default, opts...)
    return &Figure{
        SizePx:  geom.Pt{X: float64(w), Y: float64(h)},
        RC:      rc,
        Children: nil,
    }
}

// Axes represents an axes region inside a figure.
type Axes struct {
    RectFraction geom.Rect // [0..1] fraction in figure coords
    RC           *style.RC // nil => inherit figure RC
    XScale       transform.Scale
    YScale       transform.Scale
    Artists      []Artist
    zsorted      bool
}

// AddAxes appends an Axes to the Figure. If opts are provided, the Axes gets its
// own RC copy; otherwise it inherits from the Figure.
func (f *Figure) AddAxes(r geom.Rect, opts ...style.Option) *Axes {
    var rc *style.RC
    if len(opts) > 0 {
        v := style.Apply(f.RC, opts...)
        rc = &v
    }
    ax := &Axes{
        RectFraction: r,
        RC:           rc,
        XScale:       transform.NewLinear(0, 1),
        YScale:       transform.NewLinear(0, 1),
    }
    f.Children = append(f.Children, ax)
    return ax
}

// Add registers an Artist with the Axes.
func (a *Axes) Add(art Artist) { a.Artists = append(a.Artists, art); a.zsorted = false }

// layout computes the pixel rectangle for this Axes inside the Figure.
func (a *Axes) layout(f *Figure) (pixelRect geom.Rect) {
    // Map fraction [0..1] to pixel coordinates
    min := geom.Pt{X: f.SizePx.X*a.RectFraction.Min.X, Y: f.SizePx.Y*a.RectFraction.Min.Y}
    max := geom.Pt{X: f.SizePx.X*a.RectFraction.Max.X, Y: f.SizePx.Y*a.RectFraction.Max.Y}
    return geom.Rect{Min: min, Max: max}
}

// effectiveRC resolves the RC for this axes, inheriting from the Figure if needed.
func (a *Axes) effectiveRC(f *Figure) style.RC {
    if a.RC != nil { return *a.RC }
    return f.RC
}

// DrawFigure performs a traversal and draws the figure into the renderer.
func DrawFigure(fig *Figure, r render.Renderer) {
    vp := geom.Rect{Min: geom.Pt{0, 0}, Max: geom.Pt{fig.SizePx.X, fig.SizePx.Y}}
    _ = r.Begin(vp)
    defer r.End()

    for _, ax := range fig.Children {
        px := ax.layout(fig)
        r.Save()
        r.ClipRect(px)

        // Build DrawContext with composed transform
        ctx := &DrawContext{
            DataToPixel: Transform2D{
                XScale:      ax.XScale,
                YScale:      ax.YScale,
                AxesToPixel: transform.NewAffine(axesToPixel(px)),
            },
            RC:   ax.effectiveRC(fig),
            Clip: px,
        }

        if !ax.zsorted {
            sort.SliceStable(ax.Artists, func(i, j int) bool {
                zi, zj := ax.Artists[i].Z(), ax.Artists[j].Z()
                if zi == zj { return i < j }
                return zi < zj
            })
            ax.zsorted = true
        }
        for _, art := range ax.Artists { art.Draw(r, ctx) }
        r.Restore()
    }
}

// axesToPixel returns an affine mapping [0..1]^2 (axes space) -> pixel rect.
func axesToPixel(px geom.Rect) geom.Affine {
    sx := px.W()
    sy := px.H()
    tx := px.Min.X
    ty := px.Min.Y
    return geom.Affine{A: sx, D: sy, E: tx, F: ty}
}

