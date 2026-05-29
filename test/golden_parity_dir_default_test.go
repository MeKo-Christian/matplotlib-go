//go:build !freetype261

package test

// goldenParityDir returns the FreeType-parity golden directory, or "" for the
// default build (no parity golden set; use the shared testdata/golden/).
func goldenParityDir() string { return "" }
