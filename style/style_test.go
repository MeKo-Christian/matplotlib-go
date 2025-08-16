package style

import "testing"

func TestDefaults(t *testing.T) {
	d := Default
	if d.DPI != 96 || d.FontKey == "" || d.FontSize <= 0 {
		t.Fatalf("unexpected defaults: %+v", d)
	}
	if d.TickCountX != 5 || d.TickCountY != 5 {
		t.Fatalf("unexpected tick defaults: %+v", d)
	}
}

func TestOptionsApplyAndOrder(t *testing.T) {
	rc := Apply(Default,
		WithDPI(144),
		WithFont("TestFont", 14),
		WithLineWidth(2.0),
		WithTextColor(0.1, 0.2, 0.3, 0.4),
		WithLineColor(0.5, 0.6, 0.7, 0.8),
		WithBackground(0.9, 0.9, 0.9, 1.0),
		WithTickCounts(7, 9),
	)
	if rc.DPI != 144 || rc.FontKey != "TestFont" || rc.FontSize != 14 {
		t.Fatalf("font/dpi options not applied: %+v", rc)
	}
	if rc.LineWidth != 2.0 || rc.TextColor != [4]float64{0.1, 0.2, 0.3, 0.4} {
		t.Fatalf("style color/width not applied: %+v", rc)
	}
	if rc.TickCountX != 7 || rc.TickCountY != 9 {
		t.Fatalf("tick counts not applied: %+v", rc)
	}

	// Order: last wins
	rc2 := Apply(Default, WithDPI(110), WithDPI(72))
	if rc2.DPI != 72 {
		t.Fatalf("expected last option to win, got %v", rc2.DPI)
	}
}

func TestPrecedence_SimulatedFigureAxes(t *testing.T) {
	// Simulate precedence: global(Default) -> figure overrides -> axes overrides
	figRC := Apply(Default, WithDPI(110), WithFont("FigFont", 11))
	axRC := Apply(figRC, WithFont("AxesFont", 9))

	// Axes font overrides figure/global
	if axRC.FontKey != "AxesFont" || axRC.FontSize != 9 {
		t.Fatalf("axes font override failed: %+v", axRC)
	}
	// Inherit figure DPI
	if axRC.DPI != 110 {
		t.Fatalf("expected DPI inherit from figure, got %v", axRC.DPI)
	}
	// Inherit defaults for untouched fields
	if axRC.LineWidth != Default.LineWidth {
		t.Fatalf("expected default line width inherit, got %v", axRC.LineWidth)
	}
}
