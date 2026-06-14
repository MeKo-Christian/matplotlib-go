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

	plan := readTextFile(t, "PLAN.md")
	for _, want := range []string{
		"[x] **P2 — Surface and image-copy reuse for repeated renders.**",
		"[x] Add a benchmark that redraws the same figure into a reused renderer.",
		"[x] Document or expose a supported renderer-reuse path",
		"[x] Avoid `GetImage` copies in benchmark/save paths",
	} {
		if !strings.Contains(plan, want) {
			t.Fatalf("plan missing %q", want)
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

	plan := readTextFile(t, "PLAN.md")
	for _, want := range []string{
		"[x] **P2 — Cache scalar mapping setup.**",
		"[x] Cache resolved colormap and norm state on scalar-mapped artists",
		"[x] Add focused benchmarks for scalar-mapped image, scatter, and mesh rows.",
	} {
		if !strings.Contains(plan, want) {
			t.Fatalf("plan missing %q", want)
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

	plan := readTextFile(t, "PLAN.md")
	for _, want := range []string{
		"[x] **Memory targets and tuning guide.**",
		"[x] Define v1.0 memory targets for typical catalog plots, 100k scatter, and",
		"[x] Document practical tuning advice: renderer reuse, avoiding unnecessary",
	} {
		if !strings.Contains(plan, want) {
			t.Fatalf("plan missing %q", want)
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
