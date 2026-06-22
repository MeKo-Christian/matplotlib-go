// Command vectorshowcase emits a small scene through the PostScript and PGF
// backends to support visual spot-checks of their native vector features:
// linear/radial/multi-stop gradient fills, tiled pattern fills, path clipping,
// hatching, and vertical (stacked) text.
//
// It writes three files into the output directory:
//
//	showcase.ps   - a complete PostScript document (render with Ghostscript)
//	showcase.pgf  - a PGF/TikZ fragment (\input into a LaTeX document)
//	showcase.tex  - a ready-to-compile LaTeX wrapper around showcase.pgf
//
// The `just render-vector` target runs this command and rasterizes the output
// to PNG via gs and pdflatex+pdftoppm. The command itself has no cgo/FreeType
// dependency: it imports only the PS and PGF backends.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/cwbudde/matplotlib-go/backends/pgf"
	"github.com/cwbudde/matplotlib-go/backends/ps"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	canvasW = 420
	canvasH = 200
)

func main() {
	outputDir := flag.String("output-dir", filepath.Join("testdata", "_artifacts", "vector"), "directory to write showcase files")
	flag.Parse()

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		exitf("create output dir %s: %v", *outputDir, err)
	}

	white := render.Color{R: 1, G: 1, B: 1, A: 1}

	psR, err := ps.New(canvasW, canvasH, white)
	if err != nil {
		exitf("ps.New: %v", err)
	}
	drawShowcase(psR)
	psPath := filepath.Join(*outputDir, "showcase.ps")
	if err := psR.SavePS(psPath); err != nil {
		exitf("save PS: %v", err)
	}

	pgfR, err := pgf.New(canvasW, canvasH, white)
	if err != nil {
		exitf("pgf.New: %v", err)
	}
	drawShowcase(pgfR)
	pgfPath := filepath.Join(*outputDir, "showcase.pgf")
	if err := pgfR.SavePGF(pgfPath); err != nil {
		exitf("save PGF: %v", err)
	}

	texPath := filepath.Join(*outputDir, "showcase.tex")
	if err := os.WriteFile(texPath, []byte(latexWrapper), 0o644); err != nil {
		exitf("write LaTeX wrapper: %v", err)
	}

	fmt.Printf("wrote %s, %s, %s\n", psPath, pgfPath, texPath)
}

const latexWrapper = `\documentclass{article}
\usepackage{pgf}
\usepackage[margin=5pt,paperwidth=440pt,paperheight=220pt]{geometry}
\pagestyle{empty}
\begin{document}
\noindent\input{showcase.pgf}
\end{document}
`

// drawShowcase paints the scene via the renderer-neutral interface plus the
// optional gradient/pattern/clip/vertical-text capabilities. It is backend
// agnostic, so the PS and PGF outputs are produced from identical draw calls.
func drawShowcase(r render.Renderer) {
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: canvasW, Y: canvasH}}); err != nil {
		exitf("Begin: %v", err)
	}

	black := render.Color{A: 1}

	// Top row: linear multi-stop gradient, radial gradient, pattern fill.
	r.Path(rect(10, 110, 120, 180), &render.Paint{
		Stroke: black, LineWidth: 1.5,
		FillGradient: render.GradientFill{
			Kind:  render.LinearGradient,
			Start: geom.Pt{X: 10, Y: 145}, End: geom.Pt{X: 120, Y: 145},
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, A: 1}},
				{Offset: 0.5, Color: render.Color{G: 0.7, A: 1}},
				{Offset: 1, Color: render.Color{B: 1, A: 1}},
			},
		},
	})
	r.Path(rect(155, 110, 265, 180), &render.Paint{
		Stroke: black, LineWidth: 1.5,
		FillGradient: render.GradientFill{
			Kind:   render.RadialGradient,
			Center: geom.Pt{X: 210, Y: 145}, Radius: 55,
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, G: 1, A: 1}},
				{Offset: 1, Color: render.Color{B: 0.6, A: 1}},
			},
		},
	})
	r.Path(rect(300, 110, 410, 180), &render.Paint{
		Stroke: black, LineWidth: 1.5,
		FillPattern: dotPattern(),
	})

	// Bottom row: gradient clipped to a circle, vertical text, hatch fill.
	r.Save()
	r.ClipPath(circle(geom.Pt{X: 65, Y: 55}, 35, 48))
	r.Path(rect(10, 20, 120, 90), &render.Paint{
		FillGradient: render.GradientFill{
			Kind:  render.LinearGradient,
			Start: geom.Pt{X: 30, Y: 90}, End: geom.Pt{X: 100, Y: 20},
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 0.9, G: 0.2, B: 0.1, A: 1}},
				{Offset: 1, Color: render.Color{R: 0.1, G: 0.2, B: 0.9, A: 1}},
			},
		},
	})
	r.Restore()

	if vt, ok := r.(render.VerticalTextDrawer); ok {
		vt.DrawTextVertical("VERTICAL", geom.Pt{X: 210, Y: 55}, 16, black)
	}

	r.Path(rect(300, 20, 410, 90), &render.Paint{
		Stroke:      black,
		LineWidth:   1.5,
		FillPattern: linePattern(),
	})

	if err := r.End(); err != nil {
		exitf("End: %v", err)
	}
}

func rect(x0, y0, x1, y1 float64) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: x0, Y: y0})
	p.LineTo(geom.Pt{X: x1, Y: y0})
	p.LineTo(geom.Pt{X: x1, Y: y1})
	p.LineTo(geom.Pt{X: x0, Y: y1})
	p.Close()
	return p
}

func circle(center geom.Pt, radius float64, segments int) geom.Path {
	var p geom.Path
	for i := 0; i < segments; i++ {
		theta := 2 * math.Pi * float64(i) / float64(segments)
		pt := geom.Pt{X: center.X + radius*math.Cos(theta), Y: center.Y + radius*math.Sin(theta)}
		if i == 0 {
			p.MoveTo(pt)
		} else {
			p.LineTo(pt)
		}
	}
	p.Close()
	return p
}

func dotPattern() render.PatternFill {
	var dot geom.Path
	dot.MoveTo(geom.Pt{X: 4, Y: 4})
	dot.LineTo(geom.Pt{X: 10, Y: 4})
	dot.LineTo(geom.Pt{X: 10, Y: 10})
	dot.LineTo(geom.Pt{X: 4, Y: 10})
	dot.Close()
	return render.PatternFill{
		ID:         "dots",
		Cell:       geom.Rect{Max: geom.Pt{X: 14, Y: 14}},
		Path:       dot,
		Foreground: render.Color{R: 0.2, G: 0.4, B: 0.2, A: 1},
		Background: render.Color{R: 0.85, G: 0.92, B: 0.85, A: 1},
	}
}

// linePattern is a stroked (LineWidth>0) pattern: a diagonal line per cell,
// tiled to read as a hatch. Unlike Paint.Hatch it tiles off the path bounds,
// so it renders correctly anywhere on the canvas.
func linePattern() render.PatternFill {
	var line geom.Path
	line.MoveTo(geom.Pt{X: 0, Y: 0})
	line.LineTo(geom.Pt{X: 12, Y: 12})
	return render.PatternFill{
		ID:         "lines",
		Cell:       geom.Rect{Max: geom.Pt{X: 12, Y: 12}},
		Path:       line,
		Foreground: render.Color{R: 0.2, G: 0.2, B: 0.6, A: 1},
		Background: render.Color{R: 0.95, G: 0.95, B: 0.8, A: 1},
		LineWidth:  1.5,
	}
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "vectorshowcase: "+format+"\n", args...)
	os.Exit(1)
}
