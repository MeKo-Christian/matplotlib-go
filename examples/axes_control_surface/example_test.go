package axes_control_surface

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/ticker"
)

func TestPlotKeepsTopLabelReadableAndLeftMinorTicksSparse(t *testing.T) {
	fig := Plot()
	if len(fig.Children) == 0 {
		t.Fatal("Plot() did not create the left axes")
	}
	left := fig.Children[0]

	if got, want := left.RectFraction.Max.Y, 0.78; got != want {
		t.Fatalf("left axes top = %v, want %v", got, want)
	}

	assertMinorStep := func(name string, axis *core.Axis) {
		t.Helper()
		loc, ok := axis.MinorLocator.(ticker.MultipleLocator)
		if !ok {
			t.Fatalf("%s minor locator = %T, want ticker.MultipleLocator", name, axis.MinorLocator)
		}
		if got, want := loc.Base, 0.2; got != want {
			t.Fatalf("%s minor tick step = %v, want %v", name, got, want)
		}
	}
	assertMinorStep("top x-axis", left.XAxisTop)
	assertMinorStep("right y-axis", left.YAxisRight)
}
