package style

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func TestPatchRCParseSerializeAndRuntime(t *testing.T) {
	src := `
patch.linewidth: 2.75
patch.facecolor: C1
patch.edgecolor: "123456"
patch.force_edgecolor: True
patch.antialiased: False
axes.prop_cycle: cycler('color', ['red', 'lime'])
`
	theme, report, err := ParseMPLStyle("patch.mplstyle", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unsupported = %+v", report.Unsupported)
	}
	got := theme.RC.Patch
	if got.LineWidth != 2.75 || got.FaceColorRaw != "C1" ||
		got.FaceColor != (render.Color{G: 1, A: 1}) ||
		got.EdgeColor != (render.Color{R: 0x12 / 255.0, G: 0x34 / 255.0, B: 0x56 / 255.0, A: 1}) ||
		!got.ForceEdgeColor || got.Antialiased {
		t.Fatalf("patch rc = %+v", got)
	}

	params := paramsFromRC(theme.RC)
	if params["patch.linewidth"] != "2.75" || params["patch.facecolor"] != "C1" ||
		params["patch.edgecolor"] != "#123456" ||
		params["patch.force_edgecolor"] != "True" || params["patch.antialiased"] != "False" {
		t.Fatalf("serialized patch params = %+v", params)
	}

	ResetDefaults()
	t.Cleanup(ResetDefaults)
	if _, err := UpdateParams(Params{
		"patch.facecolor": "C1",
		"axes.prop_cycle": "cycler('color', ['black', 'yellow'])",
	}); err != nil {
		t.Fatalf("UpdateParams() error = %v", err)
	}
	current := CurrentDefaults()
	if current.Patch.FaceColorRaw != "C1" ||
		current.Patch.FaceColor != (render.Color{R: 1, G: 1, A: 1}) {
		t.Fatalf("runtime patch face = raw %q resolved %+v", current.Patch.FaceColorRaw, current.Patch.FaceColor)
	}
	if CurrentParams()["patch.facecolor"] != "C1" {
		t.Fatalf("CurrentParams patch.facecolor = %q", CurrentParams()["patch.facecolor"])
	}
}

func TestPatchFaceCycleReferenceReResolvesWhenCycleChanges(t *testing.T) {
	base := Apply(Default)
	base.Patch.FaceColorRaw = "C1"
	next, _, err := applyMPLStyleParams(base, Params{
		"axes.prop_cycle": "cycler('color', ['blue', 'magenta'])",
	})
	if err != nil {
		t.Fatalf("applyMPLStyleParams() error = %v", err)
	}
	if next.Patch.FaceColor != (render.Color{R: 1, B: 1, A: 1}) {
		t.Fatalf("C1 resolved to %+v, want magenta from updated cycle", next.Patch.FaceColor)
	}
}
