package main_test

import (
	"os"
	"strings"
	"testing"
)

func TestPerformanceP0Plumbing(t *testing.T) {
	justfile := readTextFile(t, "Justfile")
	for _, want := range []string{
		"bench-render: freetype261-build",
		"go test ./benchmarks -bench 'BenchmarkCatalogRender|BenchmarkLargeScatter100K'",
		"-benchmem",
		"profile-render: freetype261-build",
		"testdata/_artifacts/perf/catalog_cpu.pprof",
		"testdata/_artifacts/perf/catalog_mem.pprof",
		"testdata/_artifacts/perf/scatter100k_cpu.pprof",
		"testdata/_artifacts/perf/scatter100k_mem.pprof",
	} {
		if !strings.Contains(justfile, want) {
			t.Fatalf("Justfile missing %q", want)
		}
	}

	workflow := readTextFile(t, ".github/workflows/benchmark-report.yml")
	for _, want := range []string{
		"name: Benchmark Report",
		"continue-on-error: true",
		"just bench-render",
		"just profile-render",
		"actions/upload-artifact",
		"testdata/_artifacts/perf",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("benchmark workflow missing %q", want)
		}
	}
}

func readTextFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
