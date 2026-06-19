package pyplot

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/style"
)

func TestRCUpdatesActiveDefaultsForNewFigures(t *testing.T) {
	resetForTests()

	if err := RC("figure", style.Params{"dpi": "144"}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}
	if err := RC("axes", style.Params{"facecolor": "#ddeeff"}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}

	fig := Figure()
	if fig.RC.DPI != 144 {
		t.Fatalf("figure DPI = %v, want 144", fig.RC.DPI)
	}
	if got := fig.RC.AxesBackground; got.R != 0xdd/255.0 || got.G != 0xee/255.0 || got.B != 0xff/255.0 {
		t.Fatalf("axes facecolor = %+v", got)
	}
}

func TestRCContextTemporarilyOverridesDefaults(t *testing.T) {
	resetForTests()

	if err := RC("figure", style.Params{"dpi": "120"}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}

	restore, err := RCContext(style.Params{"figure.dpi": "220"})
	if err != nil {
		t.Fatalf("RCContext() error = %v", err)
	}

	if got := Figure().RC.DPI; got != 220 {
		t.Fatalf("context figure DPI = %v, want 220", got)
	}

	restore()
	if got := Figure().RC.DPI; got != 120 {
		t.Fatalf("restored figure DPI = %v, want 120", got)
	}
}

func TestRCDefaultsResetsActiveDefaults(t *testing.T) {
	resetForTests()

	if err := RC("figure", style.Params{"dpi": "144"}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}
	RCDefaults()

	if got := Figure().RC.DPI; got != style.Default.DPI {
		t.Fatalf("figure DPI = %v, want default %v", got, style.Default.DPI)
	}
}

func TestLoadRCFileUpdatesDefaults(t *testing.T) {
	resetForTests()

	path := filepath.Join(t.TempDir(), "matplotlibrc")
	if err := os.WriteFile(path, []byte("figure.dpi: 175\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := LoadRCFile(path); err != nil {
		t.Fatalf("LoadRCFile() error = %v", err)
	}
	if got := Figure().RC.DPI; got != 175 {
		t.Fatalf("figure DPI = %v, want 175", got)
	}
}

func TestFigureUsesConfiguredFigureSize(t *testing.T) {
	resetForTests()

	if err := RC("figure", style.Params{
		"dpi":     "120",
		"figsize": "7.5, 5",
	}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}

	fig := Figure()
	if fig.SizePx.X != 900 || fig.SizePx.Y != 600 {
		t.Fatalf("figure size = %.0fx%.0f, want 900x600", fig.SizePx.X, fig.SizePx.Y)
	}
}
