package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestAxesSpineVisibilityConsumesRCDefaults(t *testing.T) {
	fig := NewFigure(200, 120)
	fig.RC.Axes.Spines.Top = false
	fig.RC.Axes.Spines.Bottom = true
	fig.RC.Axes.Spines.Left = false
	fig.RC.Axes.Spines.Right = true

	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	if !ax.XAxis.ShowSpine {
		t.Fatal("bottom spine hidden, want visible")
	}
	if ax.YAxis.ShowSpine {
		t.Fatal("left spine visible, want hidden")
	}
	if ax.fallbackSpineVisible(AxisTop) {
		t.Fatal("top spine visible, want hidden")
	}
	if !ax.fallbackSpineVisible(AxisRight) {
		t.Fatal("right spine hidden, want visible")
	}

	if top := ax.TopAxis(); top == nil || top.ShowSpine {
		t.Fatalf("explicit top axis = %+v, want rc-hidden spine", top)
	}
	if right := ax.RightAxis(); right == nil || !right.ShowSpine {
		t.Fatalf("explicit right axis = %+v, want rc-visible spine", right)
	}
}

func TestAxesSpineExplicitChangesWinUntilClear(t *testing.T) {
	fig := NewFigure(200, 120)
	fig.RC.Axes.Spines.Top = false
	fig.RC.Axes.Spines.Bottom = false
	fig.RC.Axes.Spines.Left = false
	fig.RC.Axes.Spines.Right = false
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})

	ax.XAxis.ShowSpine = true
	ax.YAxis.ShowSpine = true
	ax.TopAxis().ShowSpine = true
	ax.RightAxis().ShowSpine = true
	if !ax.XAxis.ShowSpine || !ax.YAxis.ShowSpine ||
		!ax.XAxisTop.ShowSpine || !ax.YAxisRight.ShowSpine {
		t.Fatal("explicit spine visibility changes were not retained")
	}

	ax.Clear()
	if ax.XAxis.ShowSpine || ax.YAxis.ShowSpine ||
		ax.fallbackSpineVisible(AxisTop) || ax.fallbackSpineVisible(AxisRight) {
		t.Fatal("Clear did not restore rc-hidden spine defaults")
	}
	if ax.XAxisTop != nil || ax.YAxisRight != nil {
		t.Fatal("Clear did not restore fallback top/right spines")
	}
}

func TestProjectionSpinesIgnoreRectilinearRCVisibility(t *testing.T) {
	fig := NewFigure(200, 120)
	fig.RC.Axes.Spines.Top = false
	fig.RC.Axes.Spines.Bottom = false
	fig.RC.Axes.Spines.Left = true
	fig.RC.Axes.Spines.Right = true

	ax, err := fig.AddAxesProjection(
		geom.Rect{Max: geom.Pt{X: 1, Y: 1}},
		"polar",
	)
	if err != nil {
		t.Fatalf("AddAxesProjection(polar): %v", err)
	}
	if !ax.XAxis.ShowSpine || ax.YAxis.ShowSpine {
		t.Fatalf("polar spine visibility changed by rectilinear rcParams: x=%v y=%v",
			ax.XAxis.ShowSpine, ax.YAxis.ShowSpine)
	}
}

func TestCartesianProjectionConsumesRCSpineVisibility(t *testing.T) {
	fig := NewFigure(200, 120)
	fig.RC.Axes.Spines.Bottom = false
	fig.RC.Axes.Spines.Left = false

	ax, err := fig.AddAxesProjection(
		geom.Rect{Max: geom.Pt{X: 1, Y: 1}},
		"skewx",
	)
	if err != nil {
		t.Fatalf("AddAxesProjection(skewx): %v", err)
	}
	if ax.XAxis.ShowSpine || ax.YAxis.ShowSpine {
		t.Fatalf("Cartesian projection ignored rc spine visibility: x=%v y=%v",
			ax.XAxis.ShowSpine, ax.YAxis.ShowSpine)
	}
}
