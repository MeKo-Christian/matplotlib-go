//go:build freetype261

package test

// goldenParityDir returns the golden directory for the FreeType-2.6.1 parity
// build. Cases whose text rendering differs from the default build commit a
// PNG here (testdata/golden_freetype/); all others fall back to testdata/golden/
// via goldenReadPath.
func goldenParityDir() string { return "golden_freetype" }
