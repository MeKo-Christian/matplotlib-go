package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func TestFigureRCSeedsLayoutWithExplicitPrecedence(t *testing.T) {
	fig := NewFigure(400, 300, func(rc *style.RC) {
		rc.Figure.AutoLayout = true
		rc.Figure.Constrained.Use = true
		rc.Figure.Subplot = style.FigureSubplotRC{
			Left: 0.2, Right: 0.8, Bottom: 0.15, Top: 0.85,
			WSpace: 0.4, HSpace: 0.3,
		}
	})
	if got := fig.LayoutEngine(); got != LayoutEngineTight {
		t.Fatalf("layout engine = %v, want tight (autolayout takes precedence)", got)
	}
	grid := fig.Subplots(1, 2)
	options := grid[0][0].subplotSpec.grid.options
	if options.Left != 0.2 || options.Right != 0.8 || options.Bottom != 0.15 || options.Top != 0.85 {
		t.Fatalf("subplot margins = %+v", options)
	}
	if !floatApprox(options.WSpace, 0.1/0.6, 1e-12) || !floatApprox(options.HSpace, 0.09/0.7, 1e-12) {
		t.Fatalf("subplot spacing = w=%v h=%v", options.WSpace, options.HSpace)
	}

	override := fig.Subplots(
		1, 1,
		WithSubplotPadding(0.1, 0.9, 0.12, 0.88),
		WithSubplotSpacing(0.07, 0.08),
	)
	explicit := override[0][0].subplotSpec.grid.options
	if explicit.Left != 0.1 || explicit.Right != 0.9 || explicit.Bottom != 0.12 || explicit.Top != 0.88 ||
		!floatApprox(explicit.WSpace, 0.07/0.8, 1e-12) || !floatApprox(explicit.HSpace, 0.08/0.76, 1e-12) {
		t.Fatalf("explicit subplot options did not win: %+v", explicit)
	}
	fig.ConstrainedLayout()
	if got := fig.LayoutEngine(); got != LayoutEngineConstrained {
		t.Fatalf("explicit layout engine = %v", got)
	}
}

func TestFigureRCConstrainedControlsAndZeroWidthPatchEdge(t *testing.T) {
	fig := NewFigure(200, 100, func(rc *style.RC) {
		rc.Figure.Constrained.HPad = 0.1
		rc.Figure.Constrained.WPad = 0.2
		rc.Figure.Constrained.HSpace = 0.03
		rc.Figure.Constrained.WSpace = 0.04
		rc.Figure.EdgeColor = render.Color{R: 1, A: 1}
	})
	if got, want := layoutPadPx(fig, LayoutEngineConstrained), 20.0; got != want {
		t.Fatalf("w_pad pixels = %v, want %v", got, want)
	}
	if got, want := constrainedLayoutPadPx(fig), 10.0; got != want {
		t.Fatalf("h_pad pixels = %v, want %v", got, want)
	}
	if got, want := constrainedLayoutDefaultSpacePx(fig, 200, 2, true), 4.0; got != want {
		t.Fatalf("wspace pixels = %v, want %v", got, want)
	}
	if got, want := constrainedLayoutDefaultSpacePx(fig, 100, 2, false), 1.5; got != want {
		t.Fatalf("hspace pixels = %v, want %v", got, want)
	}

	var r recordingRenderer
	vp := geom.Rect{Max: geom.Pt{X: 200, Y: 100}}
	drawFigureBackground(&r, vp, DrawOptions{}, fig)
	// Matplotlib's Figure constructor has linewidth=0.0. The rc edge color is
	// retained as patch state but must not invent a visible frame.
	if len(r.pathCalls) != 0 {
		t.Fatalf("zero-width figure edge unexpectedly drew: %+v", r.pathCalls)
	}
	fig.RC.Figure.FrameOn = false
	drawFigureBackground(&r, vp, DrawOptions{}, fig)
	if len(r.pathCalls) != 0 {
		t.Fatalf("frameon=false emitted a patch: %+v", r.pathCalls)
	}
}
