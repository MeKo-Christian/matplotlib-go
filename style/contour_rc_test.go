package style

import "testing"

func TestContourRCParamsParseAndRoundTrip(t *testing.T) {
	src := "contour.algorithm: serial\ncontour.corner_mask: False\ncontour.linewidth: 2.75\ncontour.negative_linestyle: dotted\n"
	theme, report, err := ParseMPLStyle("contour", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported params: %+v", report.Unsupported)
	}
	got := theme.RC.Contour
	if got.Algorithm != "serial" || got.CornerMask || !got.LineWidthSet ||
		!almostEqual(got.LineWidth, 2.75) || got.NegativeLineStyle != "dotted" {
		t.Fatalf("contour rc = %+v", got)
	}

	params := paramsFromRC(theme.RC)
	if params["contour.algorithm"] != "serial" ||
		params["contour.corner_mask"] != "False" ||
		params["contour.linewidth"] != "2.75" ||
		params["contour.negative_linestyle"] != "dotted" {
		t.Fatalf("serialized contour params = %+v", params)
	}

	none, _, err := ParseMPLStyle("contour-none", "contour.linewidth: None\n")
	if err != nil {
		t.Fatalf("parse contour.linewidth=None: %v", err)
	}
	if none.RC.Contour.LineWidthSet {
		t.Fatalf("contour.linewidth=None stored as explicit: %+v", none.RC.Contour)
	}
	if got := paramsFromRC(none.RC)["contour.linewidth"]; got != "None" {
		t.Fatalf("serialized inherited contour linewidth = %q, want None", got)
	}
}

func TestContourRCParamsRejectInvalidValues(t *testing.T) {
	for _, src := range []string{
		"contour.algorithm: imaginary\n",
		"contour.corner_mask: sometimes\n",
		"contour.linewidth: wide\n",
		"contour.negative_linestyle: zigzag\n",
	} {
		if _, _, err := ParseMPLStyle("invalid-contour", src); err == nil {
			t.Fatalf("ParseMPLStyle(%q) succeeded, want error", src)
		}
	}
}

func TestDefaultContourRCMatchesMatplotlib3109(t *testing.T) {
	got := Default.Contour
	if got.Algorithm != "mpl2014" || !got.CornerMask || got.LineWidthSet ||
		got.NegativeLineStyle != "dashed" {
		t.Fatalf("default contour rc = %+v", got)
	}
}
