package core

import (
	"math"
	"strings"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type colorbarExtensionPath struct {
	Path      geom.Path
	OverRange bool
}

func colorbarExtensionPaths(clip geom.Rect, extend, orientation string, extendRect bool) []colorbarExtensionPath {
	extend = normalizeColorbarExtend(extend)
	if extend == "neither" || clip.W() <= 0 || clip.H() <= 0 {
		return nil
	}
	out := make([]colorbarExtensionPath, 0, 2)
	if orientation == "horizontal" {
		width := clip.W() * 0.05
		if extend == "min" || extend == "both" {
			verts := []geom.Pt{
				{X: clip.Min.X, Y: clip.Min.Y},
				{X: clip.Min.X, Y: clip.Max.Y},
				{X: clip.Min.X - width, Y: (clip.Min.Y + clip.Max.Y) * 0.5},
			}
			if extendRect {
				verts = []geom.Pt{
					{X: clip.Min.X, Y: clip.Min.Y},
					{X: clip.Min.X, Y: clip.Max.Y},
					{X: clip.Min.X - width, Y: clip.Max.Y},
					{X: clip.Min.X - width, Y: clip.Min.Y},
				}
			}
			out = append(out, colorbarExtensionPath{
				OverRange: false,
				Path: geom.Path{
					V: verts,
					C: closedPolygonCmds(len(verts)),
				},
			})
		}
		if extend == "max" || extend == "both" {
			verts := []geom.Pt{
				{X: clip.Max.X, Y: clip.Min.Y},
				{X: clip.Max.X + width, Y: (clip.Min.Y + clip.Max.Y) * 0.5},
				{X: clip.Max.X, Y: clip.Max.Y},
			}
			if extendRect {
				verts = []geom.Pt{
					{X: clip.Max.X, Y: clip.Min.Y},
					{X: clip.Max.X + width, Y: clip.Min.Y},
					{X: clip.Max.X + width, Y: clip.Max.Y},
					{X: clip.Max.X, Y: clip.Max.Y},
				}
			}
			out = append(out, colorbarExtensionPath{
				OverRange: true,
				Path: geom.Path{
					V: verts,
					C: closedPolygonCmds(len(verts)),
				},
			})
		}
		return out
	}

	height := clip.H() * 0.05
	if extend == "min" || extend == "both" {
		verts := []geom.Pt{
			{X: clip.Min.X, Y: clip.Min.Y},
			{X: (clip.Min.X + clip.Max.X) * 0.5, Y: clip.Min.Y - height},
			{X: clip.Max.X, Y: clip.Min.Y},
		}
		if extendRect {
			verts = []geom.Pt{
				{X: clip.Min.X, Y: clip.Min.Y - height},
				{X: clip.Max.X, Y: clip.Min.Y - height},
				{X: clip.Max.X, Y: clip.Min.Y},
				{X: clip.Min.X, Y: clip.Min.Y},
			}
		}
		out = append(out, colorbarExtensionPath{
			OverRange: false,
			Path: geom.Path{
				V: verts,
				C: closedPolygonCmds(len(verts)),
			},
		})
	}
	if extend == "max" || extend == "both" {
		verts := []geom.Pt{
			{X: clip.Min.X, Y: clip.Max.Y},
			{X: clip.Max.X, Y: clip.Max.Y},
			{X: (clip.Min.X + clip.Max.X) * 0.5, Y: clip.Max.Y + height},
		}
		if extendRect {
			verts = []geom.Pt{
				{X: clip.Min.X, Y: clip.Max.Y},
				{X: clip.Max.X, Y: clip.Max.Y},
				{X: clip.Max.X, Y: clip.Max.Y + height},
				{X: clip.Min.X, Y: clip.Max.Y + height},
			}
		}
		out = append(out, colorbarExtensionPath{
			OverRange: true,
			Path: geom.Path{
				V: verts,
				C: closedPolygonCmds(len(verts)),
			},
		})
	}
	return out
}

func closedPolygonCmds(n int) []geom.Cmd {
	if n <= 0 {
		return nil
	}
	cmds := make([]geom.Cmd, n)
	for i := range cmds {
		if i == 0 {
			cmds[i] = geom.MoveTo
		} else {
			cmds[i] = geom.LineTo
		}
	}
	return append(cmds, geom.ClosePath)
}

// Draw renders a gradient across the colorbar axes.
func (c *Colorbar) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil || ctx == nil {
		return
	}

	const gradientHeight = 256

	mapping := c.currentMapping()
	c.Mapping = mapping
	cmap := matcolor.GetColormap(mapping.Colormap)
	alpha := c.Alpha
	if alpha <= 0 {
		alpha = 1
	}
	orientation := c.normalizedOrientation()

	if boundaries, values, ok := c.boundaryData(mapping); ok {
		for i := 0; i+1 < len(boundaries); i++ {
			rect := colorbarBoundaryCellRectAt(ctx.Clip, boundaries, i, c.Spacing, orientation)
			path := snappedFillRectPath(rect)
			if len(path.C) == 0 {
				continue
			}
			value := values[i]
			col := mapping.Color(value, alpha)
			r.Path(path, &render.Paint{
				Fill:      col,
				LineJoin:  render.JoinMiter,
				LineCap:   render.CapButt,
				Antialias: render.AntialiasDefault,
			})
		}
		if c.DrawEdges {
			drawColorbarBoundaryDividers(r, ctx.Clip, boundaries, c.Spacing, orientation, c.BorderColor, c.BorderWidth)
		}
	} else {
		for i := 0; i < gradientHeight; i++ {
			t := (float64(i) + 0.5) / float64(gradientHeight)
			col := cmap.AtValue(t)
			col.A *= alpha

			path := snappedFillRectPath(colorbarCellRect(ctx.Clip, i, gradientHeight, orientation))
			if len(path.C) == 0 {
				continue
			}
			r.Path(path, &render.Paint{
				Fill:      col,
				LineJoin:  render.JoinMiter,
				LineCap:   render.CapButt,
				Antialias: render.AntialiasDefault,
			})
		}
	}

	if normalizeColorbarExtend(c.Extend) == "neither" {
		outlinePath := pixelRectPath(ctx.Clip)
		if snapped := snappedStrokeRectPath(ctx.Clip); len(snapped.C) > 0 {
			outlinePath = snapped
		}
		r.Path(outlinePath, &render.Paint{
			Stroke:    c.BorderColor,
			LineWidth: c.BorderWidth,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
		})
	}
}

func drawColorbarBoundaryDividers(r render.Renderer, clip geom.Rect, boundaries []float64, spacing, orientation string, color render.Color, width float64) {
	if r == nil || len(boundaries) < 3 || clip.W() <= 0 || clip.H() <= 0 {
		return
	}
	for i := 1; i+1 < len(boundaries); i++ {
		var path geom.Path
		if orientation == "horizontal" {
			x := colorbarBoundaryCoord(clip.Min.X, clip.Max.X, boundaries, i, spacing)
			x = math.Floor(x) + 0.5
			path = geom.Path{
				V: []geom.Pt{{X: x, Y: clip.Min.Y}, {X: x, Y: clip.Max.Y}},
				C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			}
		} else {
			y := colorbarBoundaryCoord(clip.Max.Y, clip.Min.Y, boundaries, i, spacing)
			y = math.Floor(y) + 0.5
			path = geom.Path{
				V: []geom.Pt{{X: clip.Min.X, Y: y}, {X: clip.Max.X, Y: y}},
				C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			}
		}
		r.Path(path, &render.Paint{
			Stroke:    color,
			LineWidth: width,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
		})
	}
}

// DrawOverlay renders colorbar extension patches outside the axes clip.
func (c *Colorbar) DrawOverlay(r render.Renderer, ctx *DrawContext) {
	if c == nil || ctx == nil {
		return
	}
	if normalizeColorbarExtend(c.Extend) == "neither" {
		return
	}

	mapping := c.currentMapping()
	c.Mapping = mapping
	cmap := matcolor.GetColormap(mapping.Colormap)
	alpha := c.Alpha
	if alpha <= 0 {
		alpha = 1
	}

	orientation := c.normalizedOrientation()
	for _, ext := range colorbarExtensionPaths(ctx.Clip, c.Extend, orientation, c.ExtendRect) {
		col := render.Color{}
		if value, ok := c.boundaryExtensionValue(mapping, ext.OverRange); ok {
			col = mapping.Color(value, alpha)
		} else {
			t := -1.0
			if ext.OverRange {
				t = 2
			}
			col = cmap.AtValue(t)
			col.A *= alpha
		}
		r.Path(ext.Path, &render.Paint{
			Fill:      col,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
			Antialias: render.AntialiasOff,
		})
	}

	outline := colorbarExtendedOutlinePath(ctx.Clip, c.Extend, orientation, c.ExtendRect)
	if len(outline.C) > 0 {
		r.Path(outline, &render.Paint{
			Stroke:    c.BorderColor,
			LineWidth: c.BorderWidth,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
		})
	}
}

func colorbarExtendedOutlinePath(clip geom.Rect, extend, orientation string, extendRect bool) geom.Path {
	extend = normalizeColorbarExtend(extend)
	if extend == "neither" || clip.W() <= 0 || clip.H() <= 0 {
		return geom.Path{}
	}
	if extendRect {
		outlineRect := clip
		if orientation == "horizontal" {
			width := clip.W() * 0.05
			if extend == "min" || extend == "both" {
				outlineRect.Min.X -= width
			}
			if extend == "max" || extend == "both" {
				outlineRect.Max.X += width
			}
		} else {
			height := clip.H() * 0.05
			if extend == "min" || extend == "both" {
				outlineRect.Min.Y -= height
			}
			if extend == "max" || extend == "both" {
				outlineRect.Max.Y += height
			}
		}
		return snappedStrokeRectPath(outlineRect)
	}
	if orientation == "horizontal" {
		width := clip.W() * 0.05
		left := clip.Min.X
		leftTip := left
		if extend == "min" || extend == "both" {
			leftTip -= width
		}
		right := clip.Max.X
		rightTip := right
		if extend == "max" || extend == "both" {
			rightTip += width
		}
		midY := (clip.Min.Y + clip.Max.Y) * 0.5
		verts := []geom.Pt{{X: left, Y: clip.Max.Y}}
		if extend == "min" || extend == "both" {
			verts = append(verts, geom.Pt{X: leftTip, Y: midY})
		}
		verts = append(verts, geom.Pt{X: left, Y: clip.Min.Y}, geom.Pt{X: right, Y: clip.Min.Y})
		if extend == "max" || extend == "both" {
			verts = append(verts, geom.Pt{X: rightTip, Y: midY})
		}
		verts = append(verts, geom.Pt{X: right, Y: clip.Max.Y})
		return geom.Path{V: verts, C: closedPolygonCmds(len(verts))}
	}

	height := clip.H() * 0.05
	bottom := clip.Min.Y
	bottomTip := bottom
	if extend == "min" || extend == "both" {
		bottomTip -= height
	}
	top := clip.Max.Y
	topTip := top
	if extend == "max" || extend == "both" {
		topTip += height
	}
	midX := (clip.Min.X + clip.Max.X) * 0.5
	verts := []geom.Pt{{X: clip.Min.X, Y: bottom}}
	if extend == "min" || extend == "both" {
		verts = append(verts, geom.Pt{X: midX, Y: bottomTip})
	}
	verts = append(verts, geom.Pt{X: clip.Max.X, Y: bottom}, geom.Pt{X: clip.Max.X, Y: top})
	if extend == "max" || extend == "both" {
		verts = append(verts, geom.Pt{X: midX, Y: topTip})
	}
	verts = append(verts, geom.Pt{X: clip.Min.X, Y: top})
	return geom.Path{V: verts, C: closedPolygonCmds(len(verts))}
}

func colorbarCellRect(clip geom.Rect, index, count int, orientation string) geom.Rect {
	if count <= 0 {
		return geom.Rect{}
	}
	if orientation == "horizontal" {
		x0 := clip.Min.X + clip.W()*float64(index)/float64(count)
		x1 := clip.Min.X + clip.W()*float64(index+1)/float64(count)
		return geom.Rect{
			Min: geom.Pt{X: x0, Y: clip.Min.Y},
			Max: geom.Pt{X: x1, Y: clip.Max.Y},
		}
	}
	// Display space is y-up: index 0 (the lowest scalar value) belongs at the
	// bottom of the bar (clip.Min.Y) and the last index at the top (clip.Max.Y),
	// matching the tick labels the axis system now lays out y-up.
	y0 := clip.Min.Y + clip.H()*float64(index)/float64(count)
	y1 := clip.Min.Y + clip.H()*float64(index+1)/float64(count)
	return geom.Rect{
		Min: geom.Pt{X: clip.Min.X, Y: y0},
		Max: geom.Pt{X: clip.Max.X, Y: y1},
	}
}

func colorbarBoundaryCellRect(clip geom.Rect, low, high, vmin, vmax float64, orientation string) geom.Rect {
	span := vmax - vmin
	if span == 0 {
		return geom.Rect{}
	}
	if orientation == "horizontal" {
		x0 := clip.Min.X + clip.W()*((low-vmin)/span)
		x1 := clip.Min.X + clip.W()*((high-vmin)/span)
		return geom.Rect{
			Min: geom.Pt{X: x0, Y: clip.Min.Y},
			Max: geom.Pt{X: x1, Y: clip.Max.Y},
		}
	}
	// y-up: the low boundary maps to the bottom (clip.Min.Y), the high boundary
	// to the top (clip.Max.Y).
	y0 := clip.Min.Y + clip.H()*((low-vmin)/span)
	y1 := clip.Min.Y + clip.H()*((high-vmin)/span)
	return geom.Rect{
		Min: geom.Pt{X: clip.Min.X, Y: y0},
		Max: geom.Pt{X: clip.Max.X, Y: y1},
	}
}

func (c *Colorbar) normalizedOrientation() string {
	if c != nil && strings.ToLower(strings.TrimSpace(c.Orientation)) == "horizontal" {
		return "horizontal"
	}
	return "vertical"
}
