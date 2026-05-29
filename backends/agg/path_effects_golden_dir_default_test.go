//go:build !freetype261

package agg

import "path/filepath"

// pathEffectGoldenReadPath returns the golden PNG path for the default build.
func pathEffectGoldenReadPath(name string) string {
	return filepath.Join("testdata", "path_effects_golden", name+".png")
}

// pathEffectGoldenWriteDir is the directory -update-path-effects-golden writes to.
func pathEffectGoldenWriteDir() string { return "path_effects_golden" }
