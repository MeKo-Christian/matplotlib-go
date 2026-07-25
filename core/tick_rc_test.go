package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/ticker"
)

func TestAxesApplyRCTickGeometryAndVisibility(t *testing.T) {
	rc := style.Default
	rc.DPI = 144
	rc.XTick.Direction = "in"
	rc.XTick.Alignment = "right"
	rc.XTick.Secondary = true
	rc.XTick.LabelPrimary = false
	rc.XTick.LabelSecondary = true
	rc.XTick.Major = style.TickLevelRC{
		Size:      7,
		Width:     1.25,
		Pad:       6,
		Primary:   true,
		Secondary: false,
		Visible:   true,
	}
	rc.XTick.Minor = style.TickLevelRC{
		Size:      4,
		Width:     0.5,
		Pad:       2,
		Primary:   false,
		Secondary: true,
		Visible:   true,
		NDivs:     3,
	}
	rc.YTick.Direction = "inout"
	rc.YTick.Alignment = "top"
	rc.YTick.Primary = false
	rc.YTick.Secondary = true
	rc.YTick.LabelPrimary = false
	rc.YTick.LabelSecondary = true
	rc.YTick.Minor.Visible = true
	rc.YTick.Minor.NDivs = 4
	rc.YTick.Minor.Secondary = false

	fig := NewFigure(200, 120)
	fig.RC = rc
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})

	if got, want := ax.XAxis.TickSize, 14.0; got != want {
		t.Fatalf("bottom major tick size = %v px, want %v", got, want)
	}
	if got, want := ax.XAxis.MinorTickSize, 8.0; got != want {
		t.Fatalf("bottom minor tick size = %v px, want %v", got, want)
	}
	if got, want := ax.XAxis.tickLineWidth(), 1.25; got != want {
		t.Fatalf("bottom major tick width = %v pt, want %v", got, want)
	}
	if got, want := ax.XAxis.minorTickLineWidth(), 0.5; got != want {
		t.Fatalf("bottom minor tick width = %v pt, want %v", got, want)
	}
	if ax.XAxis.TickDirection != TickDirectionIn {
		t.Fatalf("bottom tick direction = %v, want in", ax.XAxis.TickDirection)
	}
	if !ax.XAxis.ShowTicks || ax.XAxis.ShowMinorTicks ||
		ax.XAxis.ShowLabels || ax.XAxis.ShowMinorLabels {
		t.Fatalf("bottom x visibility = %+v", ax.XAxis)
	}
	if got := ax.XAxis.MajorLabelStyle.PadPt; got != 6 {
		t.Fatalf("bottom major label pad = %v pt, want 6", got)
	}
	if !ax.XAxis.MajorLabelStyle.padPtSet || !ax.XAxis.MinorLabelStyle.padPtSet {
		t.Fatal("rc tick pads were not marked explicit")
	}
	if ax.XAxis.MajorLabelStyle.AutoAlign ||
		ax.XAxis.MajorLabelStyle.HAlign != TextAlignRight ||
		ax.XAxis.MajorLabelStyle.VAlign != TextVAlignTop {
		t.Fatalf("bottom x alignment = %+v", ax.XAxis.MajorLabelStyle)
	}
	auto, ok := ax.XAxis.MinorLocator.(ticker.AutoMinorLocator)
	if !ok || auto.N != 3 {
		t.Fatalf("bottom minor locator = %#v, want AutoMinorLocator N=3", ax.XAxis.MinorLocator)
	}

	if ax.XAxisTop == nil {
		t.Fatal("secondary x tick settings did not create a top axis")
	}
	if ax.XAxisTop.ShowTicks || !ax.XAxisTop.ShowMinorTicks ||
		ax.XAxisTop.ShowLabels || !ax.XAxisTop.ShowMinorLabels {
		t.Fatalf("top x visibility = %+v", ax.XAxisTop)
	}
	if ax.XAxisTop.MajorLabelStyle.HAlign != TextAlignRight ||
		ax.XAxisTop.MajorLabelStyle.VAlign != TextVAlignBottom {
		t.Fatalf("top x alignment = %+v", ax.XAxisTop.MajorLabelStyle)
	}

	if ax.YAxis.ShowTicks || ax.YAxis.ShowMinorTicks ||
		ax.YAxis.ShowLabels || ax.YAxis.ShowMinorLabels {
		t.Fatalf("left y visibility = %+v", ax.YAxis)
	}
	if ax.YAxisRight == nil || !ax.YAxisRight.ShowTicks || ax.YAxisRight.ShowMinorTicks ||
		!ax.YAxisRight.ShowLabels || ax.YAxisRight.ShowMinorLabels {
		t.Fatalf("right y visibility = %+v", ax.YAxisRight)
	}
	if ax.YAxisRight.TickDirection != TickDirectionInOut {
		t.Fatalf("right tick direction = %v, want inout", ax.YAxisRight.TickDirection)
	}
	if ax.YAxisRight.MajorLabelStyle.HAlign != TextAlignLeft ||
		ax.YAxisRight.MajorLabelStyle.VAlign != TextVAlignTop {
		t.Fatalf("right y alignment = %+v", ax.YAxisRight.MajorLabelStyle)
	}
}

func TestTickParamsMajorMinorVisibilityRemainIndependent(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{5}
	axis.MinorLocator = staticLocator{5.5}
	axis.ShowTicks = false
	axis.ShowMinorTicks = true
	axis.minorTickVisibilitySet = true

	ctx := createTestDrawContext()
	r := &recordingRenderer{}
	axis.DrawTicks(r, ctx)
	if got := len(r.pathCalls); got != 1 {
		t.Fatalf("minor-only DrawTicks path count = %d, want 1", got)
	}

	ax := &Axes{XAxis: axis}
	show := true
	hide := false
	if err := ax.TickParams(TickParams{
		Axis:      "bottom",
		Which:     "major",
		ShowTicks: &show,
		Bottom:    &show,
	}); err != nil {
		t.Fatalf("TickParams(major): %v", err)
	}
	if err := ax.TickParams(TickParams{
		Axis:      "bottom",
		Which:     "minor",
		ShowTicks: &hide,
		Bottom:    &hide,
	}); err != nil {
		t.Fatalf("TickParams(minor): %v", err)
	}
	if !axis.ShowTicks || axis.ShowMinorTicks {
		t.Fatalf("major/minor visibility = %v/%v, want true/false", axis.ShowTicks, axis.ShowMinorTicks)
	}
}
