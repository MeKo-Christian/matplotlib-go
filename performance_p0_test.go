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

func TestPerformanceP1ScatterThresholdIsDocumented(t *testing.T) {
	notes := readTextFile(t, "docs/performance-profiling.md")
	for _, want := range []string{
		"## Regression Budgets",
		"`BenchmarkLargeScatter100KDraw`",
		"700 ms/op",
		"400 MB/op",
		"4,000,000 allocs/op",
	} {
		if !strings.Contains(notes, want) {
			t.Fatalf("performance notes missing %q", want)
		}
	}
}

func TestPerformanceP2RendererReuseIsDocumented(t *testing.T) {
	notes := readTextFile(t, "docs/performance-profiling.md")
	for _, want := range []string{
		"`BenchmarkLargeScatter100KRedrawReuseRenderer`",
		"`agg.Renderer.Clear`",
		"`agg.Renderer.ImageView`",
		"reused renderer",
	} {
		if !strings.Contains(notes, want) {
			t.Fatalf("performance notes missing %q", want)
		}
	}
}

func TestPerformanceP2ScalarMappingCacheIsDocumented(t *testing.T) {
	notes := readTextFile(t, "docs/performance-profiling.md")
	for _, want := range []string{
		"`BenchmarkScalarMappedImageColors`",
		"`BenchmarkScalarMappedScatterColors`",
		"`BenchmarkScalarMappedQuadMeshColors`",
		"`ScalarMapInfo.Resolved`",
		"caches the resolved colormap",
	} {
		if !strings.Contains(notes, want) {
			t.Fatalf("performance notes missing %q", want)
		}
	}
}

func TestPerformanceP2MemoryTargetsAndTuningGuideIsDocumented(t *testing.T) {
	notes := readTextFile(t, "docs/performance-profiling.md")
	for _, want := range []string{
		"## Memory Targets And Tuning Guide",
		"Typical catalog plots",
		"`BenchmarkLargeScatter100KDraw`",
		"`BenchmarkLargeScatter100KRedrawReuseRenderer`",
		"Repeated redraw",
		"Avoid `GetImage`",
		"Batch markers",
		"Text-heavy tick labels",
		"Backend selection",
	} {
		if !strings.Contains(notes, want) {
			t.Fatalf("performance notes missing %q", want)
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
