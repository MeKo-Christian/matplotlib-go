# Justfile for common tasks

set shell := ["bash", "-cu"]

# Root of a built Skia checkout for the native (cgo) Skia backend. Override with
# `SKIA_ROOT=/path/to/skia just test-skia-native`. The directory must contain
# Skia's `include/` headers and a built shared library under `out/Shared/`
# (see the docs in backends/skia/skia_cwrap.cpp / native_cgo.go).
skia_root := env_var_or_default("SKIA_ROOT", "/mnt/projekte/Code/skia")
skia_cgo_cxxflags := "-I" + skia_root
skia_cgo_ldflags := "-L" + skia_root + "/out/Shared -lskia -Wl,-rpath," + skia_root + "/out/Shared"

default: build

all: build

fmt:
    if command -v treefmt >/dev/null 2>&1; then \
      treefmt --allow-missing-formatter; \
    else \
      echo "treefmt not installed; skipping"; \
    fi

lint: freetype261-build
    if command -v golangci-lint >/dev/null 2>&1; then \
      golangci-lint run ./... --timeout=5m --new-from-merge-base=origin/main; \
    else \
      echo "golangci-lint not installed; run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
      exit 1; \
    fi

lint-full: freetype261-build
    if command -v golangci-lint >/dev/null 2>&1; then \
      golangci-lint run ./... --timeout=5m; \
    else \
      echo "golangci-lint not installed; run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
      exit 1; \
    fi

lint-fix: freetype261-build
    if command -v golangci-lint >/dev/null 2>&1; then \
      golangci-lint run ./... --fix --timeout=5m; \
    else \
      echo "golangci-lint not installed; run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
      exit 1; \
    fi

build: freetype261-build
    # Use freetype-aware rendering paths (including GoBasic text-family/size handling).
    CGO_ENABLED=1 go build -tags freetype ./...

web-build:
    bash ./web/build-wasm.sh

build-skia: freetype261-build
    CGO_ENABLED=1 go build -tags "skia freetype" ./...

test: freetype261-build
    CGO_ENABLED=1 go test -tags freetype ./...

test-optional-visual: freetype261-build
    RUN_OPTIONAL_VISUAL_TESTS=true CGO_ENABLED=1 go test -tags freetype ./...

bench-render: freetype261-build
    mkdir -p testdata/_artifacts/perf
    set -o pipefail; CGO_ENABLED=1 go test ./benchmarks -bench 'BenchmarkCatalogRender|BenchmarkLargeScatter100K' -benchtime="${BENCHTIME:-1x}" -run '^$$' -count=1 -benchmem | tee testdata/_artifacts/perf/render-bench.txt

profile-render: freetype261-build
    mkdir -p testdata/_artifacts/perf
    set -o pipefail; CGO_ENABLED=1 go test ./benchmarks -bench 'BenchmarkCatalogRender' -benchtime="${CATALOG_BENCHTIME:-10x}" -run '^$$' -count=1 -benchmem -cpuprofile testdata/_artifacts/perf/catalog_cpu.pprof -memprofile testdata/_artifacts/perf/catalog_mem.pprof | tee testdata/_artifacts/perf/catalog-bench.txt
    set -o pipefail; CGO_ENABLED=1 go test ./benchmarks -bench 'BenchmarkLargeScatter100KDraw$$' -benchtime="${SCATTER_BENCHTIME:-5x}" -run '^$$' -count=1 -benchmem -cpuprofile testdata/_artifacts/perf/scatter100k_cpu.pprof -memprofile testdata/_artifacts/perf/scatter100k_mem.pprof | tee testdata/_artifacts/perf/scatter100k-bench.txt

large-file-audit:
    @echo "Large tracked Go files (>= 1000 lines)"
    @while IFS= read -r f; do wc -l "$f"; done < <(git ls-files '*.go') | awk '$1 >= 1000 {printf "%7d %s\n", $1, $2}' | sort -nr
    @echo
    @echo "Large tracked non-Go artifacts (>= 256 KiB)"
    @while IFS= read -r f; do \
      case "$f" in *.go) continue;; esac; \
      size=$(du -k "$f" | cut -f1); \
      if [ "$size" -ge 256 ]; then printf "%7dK %s\n" "$size" "$f"; fi; \
    done < <(git ls-files) | sort -nr

# --- FreeType 2.6.1 (matplotlib's pinned version) ---------------------------
# matplotlib generates every reference image with FreeType 2.6.1. The AGG
# backend statically links that same vendored FreeType by DEFAULT (the cgo
# flags in backends/agg/freetype_native.go), so text rasterization byte-matches
# the references — title_strict and the strict-text cases hit RMSE ~0. Every
# cgo target below depends on `freetype261-build` so the prefix exists before
# the first compile. (A `-tags systemfreetype` compile fallback links the
# system FreeType for environments without the vendored prefix, but it is not
# parity-exact and golden/reference tests are expected to diverge under it.)

# Build & cache the vendored static FreeType 2.6.1 (idempotent).
freetype261-build:
    bash third_party/freetype/build.sh
# ---------------------------------------------------------------------------

test-skia: freetype261-build
    CGO_ENABLED=1 go test -tags "skia freetype" ./...

# --- Native Skia backend (cgo, links a real Skia library) ------------------
# These targets add the `skiacgo` tag, which compiles backends/skia/skia_cwrap.cpp
# and links {{skia_root}}/out/Shared/libskia.so via the C-ABI wrapper. Set
# SKIA_ROOT to point at your Skia checkout if it is not at the default path.
#
# NOTE: the `freetype` tag is intentionally omitted. The native Skia backend
# uses Skia only for geometry (paths/markers/vertices/gradients), not text, so
# it does not need FreeType. If libskia.so was built with
# skia_use_freetype2/system_freetype disabled it statically bundles its own
# FreeType; linking agg's vendored FreeType 2.6.1 (the `freetype` tag) into the
# same binary causes duplicate FT_* symbols and a runtime crash. To run native
# Skia AND agg's native-FreeType text in one binary, rebuild Skia with
# `skia_use_freetype=false` first.

build-skia-native:
    CGO_ENABLED=1 \
      CGO_CXXFLAGS="{{skia_cgo_cxxflags}}" \
      CGO_LDFLAGS="{{skia_cgo_ldflags}}" \
      go build -tags "skia skiacgo" ./backends/skia/...

test-skia-native:
    CGO_ENABLED=1 \
      CGO_CXXFLAGS="{{skia_cgo_cxxflags}}" \
      CGO_LDFLAGS="{{skia_cgo_ldflags}}" \
      go test -tags "skia skiacgo" ./backends/skia/... -v

golden-update TEST="": freetype261-build
    if [ -n "{{TEST}}" ]; then \
      CGO_ENABLED=1 go test -tags freetype -count=1 -run "{{TEST}}" ./test -update-golden; \
    else \
      CGO_ENABLED=1 go test -tags freetype -count=1 -run '^TestGolden$$' ./test -update-golden; \
    fi

text-parity-backend: freetype261-build
    CGO_ENABLED=1 go test -tags freetype ./backends/agg -run "TestUsesDejaVuSansWithoutFallback|TestRasterTextWidthTracksRendererDPI|TestMeasureTextUsesStableFontLineMetrics|TestTrailingSpaceDoesNotRenderDuplicateGlyph|TestInternalSpaceDoesNotReplayPreviousGlyph" -count=1 -v

text-parity-core:
    CGO_ENABLED=1 go test ./core -run "TestTitleFontSizeUsesTitleOnlyCompensation|TestDrawAxesLabels_YLabelUsesTickBoundsAndLabelPad|TestTickLabelPositionUsesBoundsForBottomXAxis|TestTickLabelPositionUsesBoundsForLeftYAxis|TestTickLabelPositionUsesFontHeightMetricsForBottomXAxis|TestTickLabelPositionUsesBottomAlignmentForTopXAxis|TestTickLabelPositionUsesCenterBaselineForRightYAxis|TestAlignedTextOrigin|TestAxesTextDrawsNormalizedContent|TestAnnotationDrawOverlayRendersArrowAndText|TestAxesTextSupportsAxesAndBlendedCoordinates" -count=1 -v

text-parity-canaries: freetype261-build
    RUN_OPTIONAL_VISUAL_TESTS=true CGO_ENABLED=1 go test -tags freetype ./test -run 'TestMatplotlibRef/(bar_basic_tick_labels|bar_basic_title|hist_strategies|text_labels_strict|title_strict)$$' -count=1 -v

text-parity-golden: freetype261-build
    CGO_ENABLED=1 go test -tags freetype ./test -run 'TestGolden/(bar_basic_tick_labels|bar_basic_title|hist_strategies|text_labels_strict|title_strict)$$' -count=1 -update-golden -v

text-parity-compare: freetype261-build
    RUN_OPTIONAL_VISUAL_TESTS=true CGO_ENABLED=1 go test -tags freetype ./test -run 'TestReferenceCompare/(bar_basic_tick_labels|bar_basic_title|hist_strategies|text_labels_strict|title_strict)$$' -count=1 -v

backend-info:
    @go run ./examples/backends/info/main.go 2>/dev/null || echo "Backend info example not yet available"

cli:
    go run ./main.go --help

# Render the vector-backend showcase (gradient/pattern/clip/vertical-text) to
# PNG for visual spot-checks of the PS and PGF backends. Emits showcase.ps,
# showcase.pgf and showcase.tex, then rasterizes them. Requires ghostscript
# (gs) for PS and a LaTeX with pgf (pdflatex) + pdftoppm (poppler) for PGF;
# any missing tool is skipped with a note. No cgo/FreeType needed.
render-vector OUT="testdata/_artifacts/vector":
    mkdir -p "{{OUT}}"
    go run ./cmd/vectorshowcase --output-dir "{{OUT}}"
    if command -v gs >/dev/null 2>&1; then \
      gs -q -dSAFER -dBATCH -dNOPAUSE -sDEVICE=png16m -r150 -o "{{OUT}}/showcase_ps.png" "{{OUT}}/showcase.ps"; \
      echo "wrote {{OUT}}/showcase_ps.png"; \
    else \
      echo "ghostscript (gs) not found; skipping PS->PNG (install: apt-get install ghostscript)"; \
    fi
    if command -v pdflatex >/dev/null 2>&1 && command -v pdftoppm >/dev/null 2>&1; then \
      ( cd "{{OUT}}" && pdflatex -interaction=nonstopmode -halt-on-error showcase.tex >showcase_pdflatex.log 2>&1 ); \
      pdftoppm -png -r 150 -singlefile "{{OUT}}/showcase.pdf" "{{OUT}}/showcase_pgf"; \
      echo "wrote {{OUT}}/showcase_pgf.png"; \
    else \
      echo "pdflatex/pdftoppm not found; skipping PGF->PNG (install: apt-get install texlive-pictures poppler-utils)"; \
    fi

# Start parity comparison viewer for matplotlib-go golden vs reference images.
parity-viewer PORT="8090" FILTER="": freetype261-build
    PORT={{PORT}} CGO_ENABLED=1 go run -tags freetype ./cmd/parityviewer --port {{PORT}} --name-filter "{{FILTER}}"

# Start parity viewer with standard golden/reference cases and web demo cases.
parity-viewer-all PORT="8090" FILTER="":
    PORT={{PORT}} CGO_ENABLED=1 go run -tags freetype ./cmd/parityviewer --port {{PORT}} --include-webdemo --name-filter "{{FILTER}}"

# Print parity comparison rows for filtered cases (no server) and exit.
parity-viewer-print PORT="8090" FILTER="" PREFIX="":
    PORT={{PORT}} CGO_ENABLED=1 go run -tags freetype ./cmd/parityviewer --port {{PORT}} --name-filter "{{FILTER}}" --name-prefix "{{PREFIX}}" --print

# Print standard and web demo parity comparison rows without starting a server.
parity-viewer-all-print PORT="8090" FILTER="" PREFIX="":
    PORT={{PORT}} CGO_ENABLED=1 go run -tags freetype ./cmd/parityviewer --port {{PORT}} --include-webdemo --name-filter "{{FILTER}}" --name-prefix "{{PREFIX}}" --print

# Generate Go and Matplotlib PNGs for the browser demo catalog.
web-parity-update DEMOS="all" WIDTH="960" HEIGHT="540":
    mkdir -p testdata/_artifacts/webdemo/go testdata/_artifacts/webdemo/matplotlib
    CGO_ENABLED=1 go run -tags freetype ./cmd/webdemoexport --backend agg --output-dir testdata/_artifacts/webdemo/go --demos "{{DEMOS}}" --width {{WIDTH}} --height {{HEIGHT}}
    if command -v uv >/dev/null 2>&1; then \
      uv run test/matplotlib_ref/webdemo.py --output-dir testdata/_artifacts/webdemo/matplotlib --width {{WIDTH}} --height {{HEIGHT}} --plots {{DEMOS}}; \
    else \
      python3 test/matplotlib_ref/webdemo.py --output-dir testdata/_artifacts/webdemo/matplotlib --width {{WIDTH}} --height {{HEIGHT}} --plots {{DEMOS}}; \
    fi

# Generate Skia-tagged Go PNGs and Matplotlib PNGs for the browser demo catalog.
web-parity-update-skia DEMOS="all" WIDTH="960" HEIGHT="540":
    mkdir -p testdata/_artifacts/webdemo/skia testdata/_artifacts/webdemo/matplotlib
    CGO_ENABLED=1 go run -tags "skia freetype" ./cmd/webdemoexport --backend skia --output-dir testdata/_artifacts/webdemo/skia --demos "{{DEMOS}}" --width {{WIDTH}} --height {{HEIGHT}}
    if command -v uv >/dev/null 2>&1; then \
      uv run test/matplotlib_ref/webdemo.py --output-dir testdata/_artifacts/webdemo/matplotlib --width {{WIDTH}} --height {{HEIGHT}} --plots {{DEMOS}}; \
    else \
      python3 test/matplotlib_ref/webdemo.py --output-dir testdata/_artifacts/webdemo/matplotlib --width {{WIDTH}} --height {{HEIGHT}} --plots {{DEMOS}}; \
    fi

# Start parity viewer for web demo Matplotlib references vs direct Go PNG exports.
web-parity-viewer PORT="8090" FILTER="":
    PORT={{PORT}} CGO_ENABLED=1 go run -tags freetype ./cmd/parityviewer --port {{PORT}} --baseline-dir testdata/_artifacts/webdemo/matplotlib --artifact-dir testdata/_artifacts/webdemo/go --name-filter "{{FILTER}}"

# Start parity viewer for web demo Matplotlib references vs Skia-tagged Go PNG exports.
web-parity-viewer-skia PORT="8090" FILTER="":
    PORT={{PORT}} CGO_ENABLED=1 go run -tags freetype ./cmd/parityviewer --port {{PORT}} --baseline-dir testdata/_artifacts/webdemo/matplotlib --artifact-dir testdata/_artifacts/webdemo/skia --name-filter "{{FILTER}}"

# Print web demo parity comparison rows without starting a server.
web-parity-print PORT="8090" FILTER="" PREFIX="":
    PORT={{PORT}} CGO_ENABLED=1 go run -tags freetype ./cmd/parityviewer --port {{PORT}} --baseline-dir testdata/_artifacts/webdemo/matplotlib --artifact-dir testdata/_artifacts/webdemo/go --name-filter "{{FILTER}}" --name-prefix "{{PREFIX}}" --print

# Print Skia web demo parity comparison rows without starting a server.
web-parity-print-skia PORT="8090" FILTER="" PREFIX="":
    PORT={{PORT}} CGO_ENABLED=1 go run -tags freetype ./cmd/parityviewer --port {{PORT}} --baseline-dir testdata/_artifacts/webdemo/matplotlib --artifact-dir testdata/_artifacts/webdemo/skia --name-filter "{{FILTER}}" --name-prefix "{{PREFIX}}" --print

examples:
    @echo "Running examples..."
    @for dir in examples/*/; do \
        if [ -f "$$dir/main.go" ]; then \
            echo "Running $$dir"; \
            cd "$$dir" && go run main.go; \
            cd - > /dev/null; \
        elif [ -f "$$dir/basic.go" ]; then \
            echo "Running $$dir/basic.go"; \
            cd "$$dir" && go run basic.go; \
            cd - > /dev/null; \
        fi; \
    done
    @for subdir in examples/*/*/; do \
        if [ -f "$$subdir/main.go" ]; then \
            echo "Running $$subdir"; \
            cd "$$subdir" && go run main.go; \
            cd - > /dev/null; \
        fi; \
    done

clean-examples:
    @echo "Cleaning PNG files from examples..."
    find examples/ -name "*.png" -type f -delete
    @echo "PNG files removed."

fix:
    just lint-fix
    just fmt
