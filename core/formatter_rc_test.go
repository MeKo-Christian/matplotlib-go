package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/ticker"
)

func TestAxesScalarFormatterConsumesRCDefaults(t *testing.T) {
	t.Setenv("LC_ALL", "de_DE.UTF-8")

	fig := NewFigure(200, 120)
	fig.RC.Axes.Formatter.Limits = [2]int{-3, 7}
	fig.RC.Axes.Formatter.OffsetThreshold = 6
	fig.RC.Axes.Formatter.UseLocale = true
	fig.RC.Axes.Formatter.UseMathText = true
	fig.RC.Axes.Formatter.UseOffset = false

	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	for _, axis := range []*Axis{ax.XAxis, ax.YAxis} {
		formatter, ok := axis.Formatter.(ticker.ScalarFormatter)
		if !ok {
			t.Fatalf("default formatter = %T, want ScalarFormatter", axis.Formatter)
		}
		if formatter.PowerLimits != [2]int{-3, 7} || !formatter.UsePowerLimits {
			t.Fatalf("formatter power limits = %+v (enabled=%v), want [-3 7]", formatter.PowerLimits, formatter.UsePowerLimits)
		}
		if formatter.OffsetThreshold != 6 || !formatter.DisableOffset {
			t.Fatalf("formatter offset defaults not consumed: %+v", formatter)
		}
		if !formatter.UseLocale || !formatter.UseMathText {
			t.Fatalf("formatter locale/mathtext defaults not consumed: %+v", formatter)
		}
		if got, want := formatter.Format(1234.5), "1.234,5"; got != want {
			t.Fatalf("localized scalar label = %q, want %q", got, want)
		}
		if got, want := formatter.FormatStep(1234.5, 0.5), `$\mathdefault{1.234{,}5}$`; got != want {
			t.Fatalf("localized MathText label = %q, want %q", got, want)
		}
		if got, want := formatter.FormatData(1.2e6), `1{,}2 \times 10^{6}`; got != want {
			t.Fatalf("localized MathText offset component = %q, want %q", got, want)
		}
		if got := formatter.OffsetText([]float64{1_000_100, 1_000_200, 1_000_300}); got != "" {
			t.Fatalf("useoffset=False produced offset text %q", got)
		}
	}
}

func TestAxesFormatterExplicitReplacementWinsUntilClear(t *testing.T) {
	fig := NewFigure(200, 120)
	fig.RC.Axes.Formatter.Limits = [2]int{-2, 4}
	fig.RC.Axes.Formatter.UseMathText = true
	fig.RC.Axes.Formatter.UseOffset = false
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})

	explicit := ticker.ScalarFormatter{Prec: 1, DisableScientific: true}
	ax.XAxis.Formatter = explicit
	if got := ax.XAxis.Formatter; got != explicit {
		t.Fatalf("explicit formatter changed unexpectedly: %#v", got)
	}

	ax.Clear()
	formatter, ok := ax.XAxis.Formatter.(ticker.ScalarFormatter)
	if !ok {
		t.Fatalf("formatter after Clear = %T, want ScalarFormatter", ax.XAxis.Formatter)
	}
	if formatter.PowerLimits != [2]int{-2, 4} || !formatter.UseMathText || !formatter.DisableOffset {
		t.Fatalf("Clear did not restore rc formatter defaults: %+v", formatter)
	}
}

func TestAxesFormatterMinExponentReachesNewLogFormatter(t *testing.T) {
	fig := NewFigure(200, 120)
	fig.RC.Axes.Formatter.MinExponent = 2
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})

	if err := ax.SetXScale("log"); err != nil {
		t.Fatalf("SetXScale(log): %v", err)
	}
	formatter, ok := ax.XAxis.Formatter.(ticker.LogFormatterMathText)
	if !ok {
		t.Fatalf("log formatter = %T, want LogFormatterMathText", ax.XAxis.Formatter)
	}
	if formatter.MinExponent != 2 {
		t.Fatalf("log formatter min exponent = %d, want 2", formatter.MinExponent)
	}
	if got, want := formatter.Format(10), `$\mathdefault{10}$`; got != want {
		t.Fatalf("small exponent label = %q, want %q", got, want)
	}
	if got, want := formatter.Format(100), `$\mathdefault{10^{2}}$`; got != want {
		t.Fatalf("threshold exponent label = %q, want %q", got, want)
	}
}
