package test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// === from bar_text_diagnostic_test.go =========================
//
// Records DrawText calls from the AGG renderer and compares the resulting
// metrics against Matplotlib's text placement for the same figure. Skipped
// unless MPL_GO_TEXT_DIAG is set; also skipped under purego builds where
// the AGG renderer's FreeType pipeline isn't compiled in (without it the
// metrics would be apples-to-oranges).

type textDrawRecord struct {
	Text    string                   `json:"text"`
	Size    float64                  `json:"size"`
	Origin  geom.Pt                  `json:"origin"`
	Metrics render.TextMetrics       `json:"metrics"`
	Heights render.FontHeightMetrics `json:"heights,omitempty"`
	Bounds  render.TextBounds        `json:"bounds"`
	Ink     geom.Rect                `json:"ink"`
}

type textRecordingRenderer struct {
	*agg.Renderer
	records []textDrawRecord
}

func (r *textRecordingRenderer) DrawText(text string, origin geom.Pt, size float64, color render.Color) {
	bounds, _ := r.MeasureTextBounds(text, size, "")
	metrics := r.MeasureText(text, size, "")
	heights, _ := r.MeasureFontHeights(size, "")
	r.records = append(r.records, textDrawRecord{
		Text:    text,
		Size:    size,
		Origin:  origin,
		Metrics: metrics,
		Heights: heights,
		Bounds:  bounds,
		Ink: geom.Rect{
			Min: geom.Pt{X: origin.X + bounds.X, Y: origin.Y + bounds.Y},
			Max: geom.Pt{X: origin.X + bounds.X + bounds.W, Y: origin.Y + bounds.Y + bounds.H},
		},
	})
	r.Renderer.DrawText(text, origin, size, color)
}

func (r *textRecordingRenderer) DrawTextWithFont(text string, origin geom.Pt, size float64, color render.Color, fontKey string) {
	bounds, _ := r.MeasureTextBounds(text, size, fontKey)
	metrics := r.MeasureText(text, size, fontKey)
	heights, _ := r.MeasureFontHeights(size, fontKey)
	r.records = append(r.records, textDrawRecord{
		Text:    text,
		Size:    size,
		Origin:  origin,
		Metrics: metrics,
		Heights: heights,
		Bounds:  bounds,
		Ink: geom.Rect{
			Min: geom.Pt{X: origin.X + bounds.X, Y: origin.Y + bounds.Y},
			Max: geom.Pt{X: origin.X + bounds.X + bounds.W, Y: origin.Y + bounds.Y + bounds.H},
		},
	})
	r.Renderer.DrawTextWithFont(text, origin, size, color, fontKey)
}

func TestBarBasicTextPlacementDiagnostic(t *testing.T) {
	if os.Getenv("MPL_GO_TEXT_DIAG") == "" {
		t.Skip("set MPL_GO_TEXT_DIAG=1 to log Go vs Matplotlib text placement")
	}
	if agg.NativeFreetypeVersion() == "" {
		t.Skip("bar text diagnostic requires the cgo FreeType pipeline; build without -tags purego")
	}

	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 10)
	ax.SetTitle("Basic Bars")

	base, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("create AGG renderer: %v", err)
	}
	rec := &textRecordingRenderer{Renderer: base}
	core.DrawFigure(fig, rec)

	goPayload, err := json.MarshalIndent(rec.records, "", "  ")
	if err != nil {
		t.Fatalf("marshal Go records: %v", err)
	}
	t.Logf("go text records:\n%s", goPayload)
	for _, size := range []float64{10, 12} {
		metrics := rec.MeasureText("lp", size, "")
		bounds, _ := rec.MeasureTextBounds("lp", size, "")
		heights, _ := rec.MeasureFontHeights(size, "")
		t.Logf("go lp size %.0f: metrics=%+v bounds=%+v heights=%+v", size, metrics, bounds, heights)
	}

	python, err := matplotlibPythonPathForDiag(t)
	if err != nil {
		t.Skipf("Matplotlib Python unavailable: %v", err)
	}
	mplRecords := runMatplotlibBarTextDiagnostic(t, python)
	mplPayload, err := json.MarshalIndent(mplRecords, "", "  ")
	if err != nil {
		t.Fatalf("marshal Matplotlib records: %v", err)
	}
	t.Logf("matplotlib text records:\n%s", mplPayload)
}

func matplotlibPythonPathForDiag(t *testing.T) (string, error) {
	t.Helper()
	candidates := []string{}
	if env := os.Getenv("MATPLOTLIB_GO_PYTHON"); env != "" {
		candidates = append(candidates, env)
	}
	candidates = append(candidates, "/usr/bin/python3", "python3")
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		cmd := exec.Command(candidate, "-c", "import matplotlib")
		cmd.Env = append(os.Environ(), "MPLCONFIGDIR="+t.TempDir())
		if err := cmd.Run(); err == nil {
			return candidate, nil
		}
	}
	return "", exec.ErrNotFound
}

func runMatplotlibBarTextDiagnostic(t *testing.T, python string) []map[string]any {
	t.Helper()
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	script := `
import json
import matplotlib
matplotlib.use("Agg")
from matplotlib.backends.backend_agg import FigureCanvasAgg
from test.matplotlib_ref.common import _bar_basic_scaffold

fig, ax = _bar_basic_scaffold(show_ticks=True, show_tick_labels=True, show_title=True)
canvas = FigureCanvasAgg(fig)
canvas.draw()
renderer = canvas.get_renderer()
height = fig.bbox.height

records = []
for group, labels in [
    ("x", ax.get_xticklabels()),
    ("y", ax.get_yticklabels()),
    ("title", [ax.title]),
]:
    for label in labels:
        if not label.get_visible() or label.get_text() == "":
            continue
        bbox = label.get_window_extent(renderer=renderer)
        anchor = label.get_transform().transform(label.get_position())
        records.append({
            "group": group,
            "text": label.get_text(),
            "fontsize": label.get_fontsize(),
            "anchor": {"x": float(anchor[0]), "y": float(height - anchor[1])},
            "bbox": {
                "min": {"x": float(bbox.x0), "y": float(height - bbox.y1)},
                "max": {"x": float(bbox.x1), "y": float(height - bbox.y0)},
            },
            "ha": label.get_ha(),
            "va": label.get_va(),
        })
print(json.dumps(records))
`
	cmd := exec.Command(python, "-c", script)
	cmd.Env = append(os.Environ(), "MPLCONFIGDIR="+t.TempDir(), "PYTHONPATH="+repoRoot)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("run Matplotlib text diagnostic: %v\n%s", err, exitErr.Stderr)
		}
		t.Fatalf("run Matplotlib text diagnostic: %v", err)
	}
	var records []map[string]any
	if err := json.Unmarshal(out, &records); err != nil {
		t.Fatalf("decode Matplotlib text diagnostic: %v\n%s", err, out)
	}
	return records
}
