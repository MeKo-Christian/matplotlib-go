package core

import (
	"slices"
	"testing"
)

func TestGetpSetpRoundtrip(t *testing.T) {
	ax := testAxes()
	line := ax.Plot([]float64{0, 1}, []float64{0, 1})

	Setp(line, map[string]any{
		"visible": false,
		"alpha":   0.5,
		"label":   "mine",
	})

	if v, ok := Getp(line, "visible"); !ok || v != false {
		t.Fatalf("Getp visible = %v, %v; want false, true", v, ok)
	}
	if v, ok := Getp(line, "alpha"); !ok || v != 0.5 {
		t.Fatalf("Getp alpha = %v, %v; want 0.5, true", v, ok)
	}
	if v, ok := Getp(line, "label"); !ok || v != "mine" {
		t.Fatalf("Getp label = %v, %v; want \"mine\", true", v, ok)
	}
	if line.Visible() {
		t.Fatalf("Setp visible=false did not hide the line")
	}
}

func TestLine2DPropertyExtension(t *testing.T) {
	line := &Line2D{}

	Setp(line, map[string]any{
		"linewidth": 2.5,
		"visible":   false, // delegated to embedded ArtistRasterization
	})

	if v, ok := Getp(line, "linewidth"); !ok || v != 2.5 {
		t.Fatalf("Getp linewidth = %v, %v; want 2.5, true", v, ok)
	}
	if line.W != 2.5 {
		t.Fatalf("Setp linewidth did not set Line2D.W: got %v", line.W)
	}
	if v, ok := Getp(line, "visible"); !ok || v != false {
		t.Fatalf("Getp visible (delegated) = %v, %v; want false, true", v, ok)
	}
}

func TestGetpUnknownProperty(t *testing.T) {
	line := &Line2D{}
	if v, ok := Getp(line, "nope"); ok {
		t.Fatalf("Getp unknown = %v, %v; want _, false", v, ok)
	}
}

func TestFindobjAllAndPredicate(t *testing.T) {
	fig := NewFigure(400, 300)
	ax0 := fig.AddAxes(unitRect())
	ax1 := fig.AddAxes(unitRect())
	l0 := ax0.Plot([]float64{0, 1}, []float64{0, 1})
	l1 := ax1.Plot([]float64{0, 1}, []float64{1, 0})

	all := Findobj(fig, nil)
	if !slices.Contains(all, Artist(l0)) || !slices.Contains(all, Artist(l1)) {
		t.Fatalf("Findobj(nil) missing lines: %d found", len(all))
	}

	onlyL0 := Findobj(fig, func(a Artist) bool { return a == Artist(l0) })
	if len(onlyL0) != 1 || onlyL0[0] != Artist(l0) {
		t.Fatalf("Findobj predicate = %v, want [l0]", onlyL0)
	}
}

func TestFindobjTypeReturnsTyped(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	ax.Plot([]float64{0, 1}, []float64{0, 1})
	ax.Plot([]float64{0, 1}, []float64{1, 0})

	lines := FindobjType[*Line2D](fig)
	if len(lines) != 2 {
		t.Fatalf("FindobjType[*Line2D] = %d, want 2", len(lines))
	}
}
