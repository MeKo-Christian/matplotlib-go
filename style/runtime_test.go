package style

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateParamsMutatesCurrentDefaults(t *testing.T) {
	ResetDefaults()
	t.Cleanup(ResetDefaults)

	report, err := UpdateParams(Params{
		"figure.dpi":      "144",
		"axes.facecolor":  "#112233",
		"lines.linewidth": "2.5",
	})
	if err != nil {
		t.Fatalf("UpdateParams() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported params: %+v", report.Unsupported)
	}

	rc := CurrentDefaults()
	if rc.DPI != 144 {
		t.Fatalf("DPI = %v, want 144", rc.DPI)
	}
	if got := rc.AxesBackground; got.R != 0x11/255.0 || got.G != 0x22/255.0 || got.B != 0x33/255.0 {
		t.Fatalf("axes facecolor = %+v", got)
	}
	if got, want := rc.LineWidth, 2.5*144.0/72.0; !almostEqual(got, want) {
		t.Fatalf("line width = %v, want %v", got, want)
	}
}

func TestPushContextRestoresPreviousDefaults(t *testing.T) {
	ResetDefaults()
	t.Cleanup(ResetDefaults)

	if _, err := UpdateParams(Params{"figure.dpi": "110"}); err != nil {
		t.Fatalf("UpdateParams() error = %v", err)
	}

	restore, report, err := PushContext(Params{"figure.dpi": "220", "text.color": "red", "text.usetex": "true"})
	if err != nil {
		t.Fatalf("PushContext() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported params: %+v", report.Unsupported)
	}
	if got := CurrentDefaults().DPI; got != 220 {
		t.Fatalf("context DPI = %v, want 220", got)
	}
	if !CurrentDefaults().UseTeX {
		t.Fatal("context UseTeX = false, want true")
	}

	restore()
	if got := CurrentDefaults().DPI; got != 110 {
		t.Fatalf("restored DPI = %v, want 110", got)
	}
	if CurrentDefaults().UseTeX {
		t.Fatal("restored UseTeX = true, want false")
	}
}

// TestCurrentParamsRoundTripsWithoutUnsupported guards the round-trip invariant:
// every key serialized by paramsFromRC must be re-parseable by applyMPLStyleEntry
// (and vice versa), so applying CurrentParams back onto the defaults reports no
// unsupported keys. This single test covers all rcParam groups.
func TestCurrentParamsRoundTripsWithoutUnsupported(t *testing.T) {
	ResetDefaults()
	t.Cleanup(ResetDefaults)

	params := CurrentParams()
	report, err := UpdateParams(params)
	if err != nil {
		t.Fatalf("UpdateParams(CurrentParams()) error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("round-trip produced unsupported keys: %+v", report.Unsupported)
	}
	if len(report.Applied) != len(params) {
		t.Fatalf("applied %d keys, want %d", len(report.Applied), len(params))
	}
}

func TestLoadRCFileReplacesCurrentDefaults(t *testing.T) {
	ResetDefaults()
	t.Cleanup(ResetDefaults)

	path := filepath.Join(t.TempDir(), "matplotlibrc")
	if err := os.WriteFile(path, []byte("figure.dpi: 160\nfigure.facecolor: black\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	report, err := LoadRCFile(path)
	if err != nil {
		t.Fatalf("LoadRCFile() error = %v", err)
	}
	if len(report.Applied) != 2 {
		t.Fatalf("applied count = %d, want 2", len(report.Applied))
	}

	rc := CurrentDefaults()
	if rc.DPI != 160 {
		t.Fatalf("DPI = %v, want 160", rc.DPI)
	}
	if got := rc.FigureBackground(); got.R != 0 || got.G != 0 || got.B != 0 || got.A != 1 {
		t.Fatalf("figure background = %+v, want black", got)
	}
}

func TestLoadDefaultRCFileUsesEnvSearchPath(t *testing.T) {
	ResetDefaults()
	t.Cleanup(ResetDefaults)

	dir := t.TempDir()
	path := filepath.Join(dir, "matplotlibrc")
	if err := os.WriteFile(path, []byte("font.size: 15\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("MATPLOTLIBRC", dir)

	loadedPath, report, err := LoadDefaultRCFile()
	if err != nil {
		t.Fatalf("LoadDefaultRCFile() error = %v", err)
	}
	if loadedPath != path {
		t.Fatalf("loaded path = %q, want %q", loadedPath, path)
	}
	if len(report.Applied) != 1 {
		t.Fatalf("applied count = %d, want 1", len(report.Applied))
	}
	if got := CurrentDefaults().FontSize; got != 15 {
		t.Fatalf("font size = %v, want 15", got)
	}
}

func TestLoadDefaultRCFileReturnsNotExistWhenMissing(t *testing.T) {
	ResetDefaults()
	t.Cleanup(ResetDefaults)
	dir := t.TempDir()
	t.Setenv("MATPLOTLIBRC", "")
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	_, _, err = LoadDefaultRCFile()
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadDefaultRCFile() error = %v, want os.ErrNotExist", err)
	}
}
