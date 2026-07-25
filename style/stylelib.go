package style

import (
	"embed"
	"io/fs"
	"path"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/diag"
)

// bundledStyleLib embeds Matplotlib's standard .mplstyle library (mpl-data/
// stylelib, version 3.10.9) so themes such as "fivethirtyeight", "bmh",
// "solarize_light2" and the "seaborn-v0_8*" family are available out of the box.
//
// Files whose names begin with "_" (Matplotlib's private sheets) are excluded by
// the embed directive automatically.
//
//go:embed stylelib/*.mplstyle
var bundledStyleLib embed.FS

// BundledStyleNames lists the embedded stylesheet themes in registration order.
// It is populated by registerBundledStyles at init.
var bundledStyleNames []string

func init() {
	registerBundledStyles()
}

// registerBundledStyles parses each embedded .mplstyle sheet and registers it as
// a Theme, but never overrides a theme that is already registered. This keeps
// the hand-tuned built-ins (default, ggplot, dark_background, publication) — and
// the golden images that depend on them — authoritative, while still shipping
// the broader Matplotlib library.
func registerBundledStyles() {
	// The bundled sheets are a fixed, known set; warning about their unparsed
	// or unhonored rcParams on every process start would print ~50 stderr
	// lines of pure noise to every consumer before their own code runs.
	// Silence diagnostics for the duration of the bundled parse, then reset
	// the one-shot warning dedup so the same keys in user-supplied sheets
	// parsed later still warn.
	restore := diag.SetHandler(nil)
	defer func() {
		restore()
		resetUnhonoredRCParamWarnings()
	}()

	entries, err := fs.ReadDir(bundledStyleLib, "stylelib")
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !isMPLStylePath(entry.Name()) {
			continue
		}
		// Skip Matplotlib's private sheets (names beginning with "_"), which an
		// explicit embed glob includes but are not meant to be public themes.
		if strings.HasPrefix(entry.Name(), "_") {
			continue
		}
		data, err := bundledStyleLib.ReadFile(path.Join("stylelib", entry.Name()))
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".mplstyle")
		theme, _, err := ParseMPLStyle(name, string(data))
		if err != nil {
			continue
		}
		if _, exists := LookupTheme(theme.Name); exists {
			continue
		}
		RegisterTheme(theme)
		bundledStyleNames = append(bundledStyleNames, theme.Name)
	}
}

// BundledStyleNames returns the names of the embedded stylesheet themes that were
// registered (those not shadowed by a built-in theme), in registration order.
func BundledStyleNames() []string {
	return append([]string(nil), bundledStyleNames...)
}
