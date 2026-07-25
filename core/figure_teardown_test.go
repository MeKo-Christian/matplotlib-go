package core

import (
	"testing"
)

func TestFigureDelAxesRemovesChild(t *testing.T) {
	fig := NewFigure(400, 300)
	ax0 := fig.AddAxes(unitRect())
	ax1 := fig.AddAxes(unitRect())

	fig.DelAxes(ax0)

	if len(fig.Children) != 1 || fig.Children[0] != ax1 {
		t.Fatalf("DelAxes did not remove ax0: Children=%v", fig.Children)
	}
	if ax0.figure != nil {
		t.Fatalf("DelAxes did not detach ax0.figure")
	}
}

func TestFigureDelAxesBreaksShareLinks(t *testing.T) {
	fig := NewFigure(400, 300)
	ax0 := fig.AddAxes(unitRect())
	ax1 := fig.AddAxes(unitRect())
	ax1.shareX = ax0
	ax1.shareY = ax0

	fig.DelAxes(ax0)

	if ax1.shareX != nil || ax1.shareY != nil {
		t.Fatalf("DelAxes did not break share links: shareX=%v shareY=%v", ax1.shareX, ax1.shareY)
	}
}

func TestAxesRemoveRoutesThroughFigure(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())

	ax.Remove()

	if len(fig.Children) != 0 {
		t.Fatalf("Remove did not detach axes: Children=%v", fig.Children)
	}
}

func TestFigureClearEmptiesEverything(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	_, _ = ax.Plot([]float64{0, 1}, []float64{0, 1})
	fig.Add(&Line2D{})
	fig.SupTitle = "super"
	fig.SupXLabel = "sx"
	fig.SupYLabel = "sy"

	fig.Clear()

	if len(fig.Children) != 0 {
		t.Fatalf("Clear left %d axes", len(fig.Children))
	}
	if len(fig.Artists) != 0 {
		t.Fatalf("Clear left %d figure artists", len(fig.Artists))
	}
	if fig.SupTitle != "" || fig.SupXLabel != "" || fig.SupYLabel != "" {
		t.Fatalf("Clear did not reset sup-labels")
	}
}

func TestFigureClfAliasesClear(t *testing.T) {
	fig := NewFigure(400, 300)
	fig.AddAxes(unitRect())
	fig.Clf()
	if len(fig.Children) != 0 {
		t.Fatalf("Clf left %d axes", len(fig.Children))
	}
}
