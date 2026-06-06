package mplot3d_gallery

import (
	"image"
	"math"
	"math/rand/v2"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	mplcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 1320
	Height = 840
	DPI    = 100
)

// Plot builds a compact mplot3d feature gallery.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	panels := []struct {
		title string
		draw  func(*core.Axes3D)
	}{
		{title: "3D line", draw: drawLine3D},
		{title: "3D scatter", draw: drawScatter3D},
		{title: "surface", draw: drawSurface},
		{title: "wireframe", draw: drawWireframe},
		{title: "trisurf", draw: drawTrisurf},
		{title: "bar3d", draw: drawBar3D},
		{title: "voxels", draw: drawVoxels},
		{title: "quiver3d", draw: drawQuiver3D},
		{title: "stem3d", draw: drawStem3D},
		{title: "fill_between3d", draw: drawFillBetween3D},
	}

	for i, panel := range panels {
		row := i / 5
		col := i % 5
		left := 0.035 + float64(col)*0.193
		bottom := 0.54
		if row == 1 {
			bottom = 0.08
		}
		ax := mustAxes3D(fig, geom.Rect{
			Min: geom.Pt{X: left, Y: bottom},
			Max: geom.Pt{X: left + 0.17, Y: bottom + 0.36},
		})
		ax.SetTitle(panel.title)
		ax.SetView(30, -60)
		panel.draw(ax)
	}

	fig.Text(0.035, 0.975, "mplot3d Gallery", core.TextOptions{
		Coords:   core.Coords(core.CoordFigure),
		HAlign:   core.TextAlignLeft,
		VAlign:   core.TextVAlignTop,
		FontSize: 13,
	})
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(DPI)
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func mustAxes3D(fig *core.Figure, rect geom.Rect) *core.Axes3D {
	ax, err := fig.AddAxes3D(rect)
	if err != nil {
		panic(err)
	}
	return ax
}

func drawLine3D(ax *core.Axes3D) {
	const n = 72
	x := make([]float64, n)
	y := make([]float64, n)
	z := make([]float64, n)
	for i := range n {
		t := float64(i) / float64(n-1)
		x[i] = t
		y[i] = math.Sin(6 * math.Pi * t)
		z[i] = math.Cos(6 * math.Pi * t)
	}
	width := 1.6
	ax.Plot3D(x, y, z, core.PlotOptions{LineWidth: &width})
}

func drawScatter3D(ax *core.Axes3D) {
	rng := rand.New(rand.NewPCG(19680801, 0))
	const n = 55
	x := make([]float64, n)
	y := make([]float64, n)
	z := make([]float64, n)
	for i := range n {
		x[i] = 23 + rng.Float64()*9
		y[i] = rng.Float64() * 100
		z[i] = -50 + rng.Float64()*25
	}
	ax.Scatter3D(x, y, z)
}

func drawSurface(ax *core.Axes3D) {
	x, y, z := radialSurface(28, -4, 4)
	cmap := "Blues"
	vmin := 2 * common.MinInGrid(z)
	zero := 0.0
	ax.Surface(x, y, z, core.PlotOptions{
		Colormap:  &cmap,
		VMin:      &vmin,
		LineWidth: &zero,
	})
}

func drawWireframe(ax *core.Axes3D) {
	x, y, z := common.Get3DWireframeTestData(0.12)
	rStride := 4
	cStride := 4
	ax.Wireframe(x, y, z, core.PlotOptions{RStride: &rStride, CStride: &cStride})
}

func drawTrisurf(ax *core.Axes3D) {
	tri, z := fanMesh(7, 20)
	cmap := "viridis"
	vmin := 2 * common.MinInSlice(z)
	ax.Trisurf(tri, z, core.PlotOptions{Colormap: &cmap, VMin: &vmin})
}

func drawBar3D(ax *core.Axes3D) {
	x := []float64{1, 1, 2, 2}
	y := []float64{1, 2, 1, 2}
	z := []float64{0, 0, 0, 0}
	dx := []float64{0.5, 0.5, 0.5, 0.5}
	dy := []float64{0.5, 0.5, 0.5, 0.5}
	dz := []float64{2, 3, 1, 4}
	ax.Bar3D(x, y, z, dx, dy, dz)
}

func drawVoxels(ax *core.Axes3D) {
	const n = 6
	filled := make([][][]bool, n)
	for i := 0; i < n; i++ {
		filled[i] = make([][]bool, n)
		for j := 0; j < n; j++ {
			filled[i][j] = make([]bool, n)
			for k := 0; k < n; k++ {
				filled[i][j][k] = (i < 2 && j < 2 && k < 2) || (i >= 4 && j >= 4 && k >= 4)
			}
		}
	}
	edgeColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	ax.Voxels(filled, core.VoxelOptions{EdgeColor: &edgeColor})
}

func drawQuiver3D(ax *core.Axes3D) {
	const n = 3
	step := 2.0 / float64(n-1)
	x := make([]float64, 0, n*n*n)
	y := make([]float64, 0, n*n*n)
	z := make([]float64, 0, n*n*n)
	u := make([]float64, 0, n*n*n)
	v := make([]float64, 0, n*n*n)
	w := make([]float64, 0, n*n*n)
	for j := 0; j < n; j++ {
		yv := -1 + float64(j)*step
		for i := 0; i < n; i++ {
			xv := -1 + float64(i)*step
			for k := 0; k < n; k++ {
				zv := -1 + float64(k)*step
				x = append(x, xv)
				y = append(y, yv)
				z = append(z, zv)
				u = append(u, (xv+yv)/5)
				v = append(v, (yv-xv)/5)
				w = append(w, 0)
			}
		}
	}
	ax.Quiver(x, y, z, u, v, w)
}

func drawStem3D(ax *core.Axes3D) {
	const n = 16
	x := make([]float64, n)
	y := make([]float64, n)
	z := make([]float64, n)
	for i := range n {
		t := 2 * math.Pi * float64(i) / float64(n-1)
		x[i] = math.Sin(t)
		y[i] = math.Cos(t)
		z[i] = float64(i) / float64(n-1)
	}
	ax.Stem(x, y, z)
}

func drawFillBetween3D(ax *core.Axes3D) {
	const n = 38
	x1 := make([]float64, n)
	y1 := make([]float64, n)
	z1 := make([]float64, n)
	x2 := make([]float64, n)
	y2 := make([]float64, n)
	z2 := make([]float64, n)
	for i := range n {
		t := 2 * math.Pi * float64(i) / float64(n-1)
		x1[i] = math.Cos(t)
		y1[i] = math.Sin(t)
		z1[i] = float64(i) / float64(n-1)
		x2[i] = math.Cos(t + math.Pi)
		y2[i] = math.Sin(t + math.Pi)
		z2[i] = z1[i]
	}
	width := 1.4
	blue := mplcolor.Tab10[0]
	ax.Plot3D(x1, y1, z1, core.PlotOptions{LineWidth: &width, Color: &blue})
	ax.Plot3D(x2, y2, z2, core.PlotOptions{LineWidth: &width, Color: &blue})
	alpha := 0.5
	ax.FillBetween(x1, y1, z1, x2, y2, z2, core.FillBetween3DOptions{Alpha: &alpha})
}

func radialSurface(count int, minVal, maxVal float64) ([]float64, []float64, [][]float64) {
	step := (maxVal - minVal) / float64(count-1)
	x := make([]float64, count)
	y := make([]float64, count)
	z := make([][]float64, count)
	for i := range count {
		x[i] = minVal + step*float64(i)
		y[i] = x[i]
		z[i] = make([]float64, count)
	}
	for yi := range count {
		for xi := range count {
			r := math.Hypot(x[xi], y[yi])
			z[yi][xi] = math.Sin(r)
		}
	}
	return x, y, z
}

func fanMesh(nRadii, nAngles int) (core.Triangulation, []float64) {
	x := make([]float64, 1+nRadii*nAngles)
	y := make([]float64, len(x))
	z := make([]float64, len(x))
	index := 1
	for angleIndex := 0; angleIndex < nAngles; angleIndex++ {
		angle := 2 * math.Pi * float64(angleIndex) / float64(nAngles)
		for radiusIndex := 0; radiusIndex < nRadii; radiusIndex++ {
			radius := 0.125 + (float64(radiusIndex)/float64(nRadii-1))*(1.0-0.125)
			x[index] = radius * math.Cos(angle)
			y[index] = radius * math.Sin(angle)
			z[index] = math.Sin(-x[index] * y[index])
			index++
		}
	}
	return core.Triangulation{X: x, Y: y}, z
}
