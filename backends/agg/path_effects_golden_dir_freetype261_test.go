//go:build freetype261

package agg

import (
	"os"
	"path/filepath"
)

// pathEffectGoldenReadPath prefers the FreeType-2.6.1 parity golden
// (testdata/path_effects_golden_freetype/) when present, falling back to the
// shared default golden for fixtures whose rendering does not depend on the
// FreeType version (the non-text path-effect fixtures).
func pathEffectGoldenReadPath(name string) string {
	p := filepath.Join("testdata", "path_effects_golden_freetype", name+".png")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return filepath.Join("testdata", "path_effects_golden", name+".png")
}

// pathEffectGoldenWriteDir is the directory -update-path-effects-golden writes
// to under the parity build.
func pathEffectGoldenWriteDir() string { return "path_effects_golden_freetype" }
