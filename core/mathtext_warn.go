package core

import (
	"sync"

	mt "github.com/cwbudde/mathtext"
	"github.com/cwbudde/matplotlib-go/internal/diag"
)

// seenUnknownMathCommands deduplicates unknown-command warnings: an axis whose
// tick labels repeat the same unrecognized command would otherwise warn on
// every tick and every redraw.
var seenUnknownMathCommands sync.Map

func init() {
	// Matplotlib raises on unknown mathtext commands. The Go engine renders them
	// as literal text; route that event to a one-shot-per-command warning so the
	// divergence is visible instead of silent.
	mt.SetUnknownCommandHandler(func(name string) {
		if _, seen := seenUnknownMathCommands.LoadOrStore(name, struct{}{}); seen {
			return
		}
		diag.Warnf("mathtext: unknown command %q rendered as literal text", `\`+name)
	})
}
