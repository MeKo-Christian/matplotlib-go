# Justfile for common tasks

set shell := ["bash", "-cu"]

default: build

all: build

fmt:
    if command -v treefmt >/dev/null 2>&1; then \
      treefmt --allow-missing-formatter; \
    else \
      echo "treefmt not installed; skipping"; \
    fi

lint:
    if command -v golangci-lint >/dev/null 2>&1; then \
      golangci-lint run --timeout=5m; \
    else \
      echo "golangci-lint not installed; run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
      exit 1; \
    fi

lint-fix:
    if command -v golangci-lint >/dev/null 2>&1; then \
      golangci-lint run --fix --timeout=5m; \
    else \
      echo "golangci-lint not installed; run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
      exit 1; \
    fi

build:
    # Use freetype-aware rendering paths (including GoBasic text-family/size handling).
    CGO_ENABLED=1 go build -tags freetype ./...

web-build:
    bash ./web/build-wasm.sh

build-skia:
    CGO_ENABLED=1 go build -tags "skia freetype" ./...

test:
    CGO_ENABLED=1 go test -tags freetype ./...

test-optional-visual:
    RUN_OPTIONAL_VISUAL_TESTS=true CGO_ENABLED=1 go test -tags freetype ./...

test-skia:
    CGO_ENABLED=1 go test -tags "skia freetype" ./...

golden-update TEST="":
    if [ -n "{{TEST}}" ]; then \
      CGO_ENABLED=1 go test -tags freetype -count=1 -run "{{TEST}}" ./test -update-golden; \
    else \
      CGO_ENABLED=1 go test -tags freetype -count=1 -run '^Test.*_Golden$$' ./test -update-golden; \
    fi

text-parity-backend:
    CGO_ENABLED=1 go test -tags freetype ./backends/agg -run "TestUsesDejaVuSansWithoutFallback|TestRasterTextWidthTracksRendererDPI|TestMeasureTextUsesStableFontLineMetrics|TestTrailingSpaceDoesNotRenderDuplicateGlyph|TestInternalSpaceDoesNotReplayPreviousGlyph" -count=1 -v

text-parity-core:
    CGO_ENABLED=1 go test ./core -run "TestTitleFontSizeUsesTitleOnlyCompensation|TestDrawAxesLabels_YLabelUsesTickBoundsAndLabelPad|TestTickLabelPositionUsesBoundsForBottomXAxis|TestTickLabelPositionUsesBoundsForLeftYAxis|TestTickLabelPositionUsesFontHeightMetricsForBottomXAxis|TestTickLabelPositionUsesBottomAlignmentForTopXAxis|TestTickLabelPositionUsesCenterBaselineForRightYAxis|TestAlignedTextOrigin|TestAxesTextDrawsNormalizedContent|TestAnnotationDrawOverlayRendersArrowAndText|TestAxesTextSupportsAxesAndBlendedCoordinates" -count=1 -v

text-parity-canaries:
    CGO_ENABLED=1 go test -tags freetype ./test -run "TestMpl_BarBasicTickLabels|TestMpl_BarBasicTitle|TestMpl_HistStrategies|TestTextLabelsStrict_MatplotlibRef|TestTitleStrict_MatplotlibRef" -count=1 -v

text-parity-golden:
    CGO_ENABLED=1 go test -tags freetype ./test -run "TestBarBasicTickLabels_Golden|TestBarBasicTitle_Golden|TestHistStrategies_Golden|TestTextLabelsStrict_Golden|TestTitleStrict_Golden" -count=1 -update-golden -v

text-parity-compare:
    CGO_ENABLED=1 go test -tags freetype ./test -run "TestReferenceImages_GoldenVsMatplotlibRef/bar_basic_tick_labels|TestReferenceImages_GoldenVsMatplotlibRef/bar_basic_title|TestReferenceImages_GoldenVsMatplotlibRef/hist_strategies|TestTextLabelsStrict_MatplotlibRef|TestTitleStrict_MatplotlibRef" -count=1 -v

backend-info:
    @go run ./examples/backends/info/main.go 2>/dev/null || echo "Backend info example not yet available"

cli:
    go run ./main.go --help

# Start parity comparison viewer for matplotlib-go golden vs reference images.
parity-viewer PORT="8090" FILTER="":
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
