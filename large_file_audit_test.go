package main_test

import (
	"strings"
	"testing"
)

func TestLargeFileAuditPlumbingIsDocumented(t *testing.T) {
	justfile := readTextFile(t, "Justfile")
	for _, want := range []string{
		"large-file-audit:",
		"git ls-files '*.go'",
		"Large tracked Go files (>= 1000 lines)",
		"Large tracked non-Go artifacts (>= 256 KiB)",
	} {
		if !strings.Contains(justfile, want) {
			t.Fatalf("Justfile missing %q", want)
		}
	}

	doc := readTextFile(t, "docs/large-file-decomposition.md")
	for _, want := range []string{
		"# Large File Decomposition",
		"`just large-file-audit`",
		"## Baseline Inventory",
		"core/axes3d_test.go",
		"core/contour.go",
		"docs/matplotlib-parity-status.md",
		"testdata/svg_golden/mathtext_basic.svg",
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("large-file decomposition doc missing %q", want)
		}
	}

	plan := readTextFile(t, "PLAN.md")
	for _, want := range []string{
		"[x] **L1 — Add a repeatable large-file audit.**",
		"`docs/large-file-decomposition.md`",
	} {
		if !strings.Contains(plan, want) {
			t.Fatalf("plan missing %q", want)
		}
	}
}
