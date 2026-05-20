package test

// Golden image regression tests.
//
// TestGolden iterates the example catalog and runs runGoldenTest per case,
// skipping cases that don't have a committed PNG in testdata/golden/. Cases
// listed in optionalVisualGoldenIDs are gated by RUN_OPTIONAL_VISUAL_TESTS=true.
//
// To regenerate goldens for all cases (or a subset via -run):
//
//	go test ./test/... -run TestGolden -update-golden
//
// Per-case invocation:
//
//	go test ./test/... -run TestGolden/basic_line
//	go test ./test/... -run 'TestGolden/.*scatter.*'

import "testing"

// TestGolden runs a byte-identical golden comparison for every catalog case
// that has a committed reference PNG.
func TestGolden(t *testing.T) {
	for _, c := range rendererNeutralCases() {
		c := c
		if !goldenExists(c.ID) {
			continue
		}
		t.Run(c.ID, func(t *testing.T) {
			if optionalVisualGoldenIDs[c.ID] {
				requireOptionalVisualTests(t)
			}
			runGoldenTest(t, c.ID)
		})
	}
}

func TestAGGNativeGolden(t *testing.T) {
	for _, c := range aggNativeCases() {
		c := c
		if !goldenExists(c.ID) {
			continue
		}
		t.Run(c.ID, func(t *testing.T) {
			runGoldenTest(t, c.ID)
		})
	}
}
