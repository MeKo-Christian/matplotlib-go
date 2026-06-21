package core

import (
	"strings"
	"testing"

	mt "github.com/cwbudde/mathtext"
	"github.com/cwbudde/matplotlib-go/internal/diag"
)

type zeroMeasurer struct{}

func (zeroMeasurer) MeasureText(string, float64, string) mt.Metrics { return mt.Metrics{} }

// Laying out a mathtext expression with an unrecognized command must surface a
// warning (routed from the mathtext engine through core into diag) rather than
// silently rendering the command as literal text. The warning is deduplicated
// per command name.
func TestUnknownMathCommandWarnsThroughCore(t *testing.T) {
	var warnings []string
	restore := diag.SetHandler(func(m string) { warnings = append(warnings, m) })
	defer restore()

	// Unique command name so the package-level dedup cache does not hide it.
	const cmd = `zzqxunknownmathcmd`
	mt.LayoutMathText(zeroMeasurer{}, `$x\`+cmd+`$`, 12, "base", mt.Options{})
	mt.LayoutMathText(zeroMeasurer{}, `$x\`+cmd+`$`, 12, "base", mt.Options{})

	if len(warnings) != 1 {
		t.Fatalf("expected exactly one deduplicated warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0], cmd) {
		t.Fatalf("warning %q should name the unknown command %q", warnings[0], cmd)
	}
}
