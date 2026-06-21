package main

import (
	"bytes"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestNewGoldenUpdateCommandIncludesFreetypeTag(t *testing.T) {
	t.Setenv("GOCACHE", "")

	cmd := newGoldenUpdateCommand("/tmp/repo", "^TestCase$")

	if cmd.Dir != "/tmp/repo" {
		t.Fatalf("Dir = %q, want %q", cmd.Dir, "/tmp/repo")
	}
	if !slices.Contains(cmd.Args, "-tags") {
		t.Fatalf("Args missing -tags: %v", cmd.Args)
	}
	if !slices.Contains(cmd.Args, goldenUpdateBuildTag) {
		t.Fatalf("Args missing %q tag: %v", goldenUpdateBuildTag, cmd.Args)
	}
	if !slices.Contains(cmd.Args, "-update-golden") {
		t.Fatalf("Args missing -update-golden: %v", cmd.Args)
	}
	if !slices.Contains(cmd.Args, "-timeout") {
		t.Fatalf("Args missing -timeout: %v", cmd.Args)
	}
	if !slices.Contains(cmd.Args, goldenUpdateTimeout.String()) {
		t.Fatalf("Args missing timeout value %q: %v", goldenUpdateTimeout, cmd.Args)
	}
	if !slices.Contains(cmd.Args, "./test") {
		t.Fatalf("Args missing ./test package: %v", cmd.Args)
	}
	pkgIdx := slices.Index(cmd.Args, "./test")
	updateIdx := slices.Index(cmd.Args, "-update-golden")
	if pkgIdx == -1 || updateIdx == -1 || updateIdx <= pkgIdx {
		t.Fatalf("expected -update-golden after ./test package so it reaches the test binary: %v", cmd.Args)
	}
	if !slices.Contains(cmd.Args, "^TestCase$") {
		t.Fatalf("Args missing run pattern: %v", cmd.Args)
	}
	if !slices.Contains(cmd.Env, "CGO_ENABLED=1") {
		t.Fatalf("Env missing CGO_ENABLED=1: %v", cmd.Env)
	}
	if !slices.Contains(cmd.Env, goldenUpdateOptionalVisualTestsEnv+"=true") {
		t.Fatalf("Env missing %s=true: %v", goldenUpdateOptionalVisualTestsEnv, cmd.Env)
	}
	if !slices.Contains(cmd.Env, "GOCACHE="+goldenUpdateDefaultGoCache()) {
		t.Fatalf("Env missing configured GOCACHE: %v", cmd.Env)
	}
}

func TestRerenderAllArtifactsUsesGoldenOnlyPattern(t *testing.T) {
	cmd := newGoldenUpdateCommand("/tmp/repo", goldenUpdateRunPatternAll)
	if !slices.Contains(cmd.Args, goldenUpdateRunPatternAll) {
		t.Fatalf("Args missing all-goldens run pattern: %v", cmd.Args)
	}
}

func TestNewParityRenderCommandRendersSingleIDToGoldenDir(t *testing.T) {
	t.Setenv("GOCACHE", "")

	cmd := newParityRenderCommand("/tmp/repo", "errorbar_basic")

	if cmd.Dir != "/tmp/repo" {
		t.Fatalf("Dir = %q, want %q", cmd.Dir, "/tmp/repo")
	}
	requiredArgs := []string{
		"run",
		"-tags",
		goldenUpdateBuildTag,
		"./test/parity/cmd",
		"--id",
		"errorbar_basic",
		"--output-dir",
		filepath.Join("testdata", "golden"),
	}
	for _, arg := range requiredArgs {
		if !slices.Contains(cmd.Args, arg) {
			t.Fatalf("Args missing %q: %v", arg, cmd.Args)
		}
	}
	if !slices.Contains(cmd.Env, "CGO_ENABLED=1") {
		t.Fatalf("Env missing CGO_ENABLED=1: %v", cmd.Env)
	}
	if !slices.Contains(cmd.Env, "GOCACHE="+goldenUpdateDefaultGoCache()) {
		t.Fatalf("Env missing configured GOCACHE: %v", cmd.Env)
	}
}

func TestPageFooterPersistsViewerStateAcrossRerender(t *testing.T) {
	requiredSnippets := []string{
		"var viewerStateStorageKey = 'mpl-parity-viewer-state-v1';",
		"var viewerStateControlIDs = ['search', 'sort-select', 'diff-mode', 'resample-mode', 'original-size'];",
		"state.openCards = Array.from(document.querySelectorAll('.card.open')).map(cardStateKey);",
		"window.sessionStorage.setItem(viewerStateStorageKey, JSON.stringify(state));",
		"restoreViewerState();",
		"bindViewerStatePersistence();",
		"setOriginalSize(document.getElementById('original-size').checked);",
		"saveViewerState();",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(pageFooter, snippet) {
			t.Fatalf("pageFooter missing snippet %q", snippet)
		}
	}
}

func TestPageHeaderDoesNotRenderMatteControl(t *testing.T) {
	forbiddenSnippets := []string{
		`id="matte-mode"`,
		"Matte:",
		"setMatteMode",
		"matte-target",
	}
	for _, snippet := range forbiddenSnippets {
		if strings.Contains(pageHeader, snippet) || strings.Contains(pageFooter, snippet) {
			t.Fatalf("page markup still contains removed matte UI snippet %q", snippet)
		}
	}
}

func TestPageFooterInitializesSliderAtCenteredPercentage(t *testing.T) {
	if !strings.Contains(pageFooter, "function applyPos(pct)") {
		t.Fatalf("pageFooter missing applyPos helper")
	}
	if !strings.Contains(pageFooter, "applyPos(0.5);") {
		t.Fatalf("pageFooter missing centered slider initialization")
	}
	if !strings.Contains(pageFooter, "overlayLayer.style.width = (100 / pct) + '%'") {
		t.Fatalf("pageFooter missing slider overlay width-lock logic")
	}
	if !strings.Contains(pageFooter, "if (pct <= 0) {") {
		t.Fatalf("pageFooter missing slider overlay zero-percent guard")
	}
	if strings.Contains(pageFooter, "setPos(wrap.getBoundingClientRect().width / 2);") {
		t.Fatalf("pageFooter still initializes slider with width-based clientX")
	}
}

func TestPageFooterPersistsOpenCardsAcrossReload(t *testing.T) {
	requiredSnippets := []string{
		"function cardStateKey(card) {",
		"state.openCards = Array.from(document.querySelectorAll('.card.open')).map(cardStateKey);",
		"var openCards = Array.isArray(state.openCards) ? state.openCards : [];",
		"card.classList.toggle('open', openCardSet.has(cardStateKey(card)));",
		"card.classList.toggle('open');",
		"loadCardImages(card);",
		"saveViewerState();",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(pageFooter, snippet) {
			t.Fatalf("pageFooter missing snippet %q", snippet)
		}
	}
}

func TestPageFooterUsesCacheBustingNavigationAfterRerender(t *testing.T) {
	requiredSnippets := []string{
		"function navigateToFreshPage() {",
		"url.searchParams.set('_pv', String(Date.now()));",
		"window.location.assign(url.toString());",
		"startRerender([name]).then(function(job) {",
		"navigateToFreshPage();",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(pageFooter, snippet) {
			t.Fatalf("pageFooter missing rerender freshness snippet %q", snippet)
		}
	}
	if strings.Contains(pageFooter, "window.location.reload();") {
		t.Fatalf("pageFooter still uses plain reload after rerender")
	}
}

func TestPageRerenderButtonsSeparateDisabledAndBusyCursors(t *testing.T) {
	requiredSnippets := []string{
		"button.rerender-btn:disabled { opacity: 0.6; cursor: not-allowed; }",
		"button.rerender-btn.is-rerendering { cursor: wait; }",
		"button.classList.toggle('is-rerendering', disabled);",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(pageHeader, snippet) && !strings.Contains(pageFooter, snippet) {
			t.Fatalf("page markup missing rerender cursor snippet %q", snippet)
		}
	}
	if strings.Contains(pageHeader, "button.rerender-btn:disabled { opacity: 0.6; cursor: wait; }") {
		t.Fatalf("disabled rerender buttons still use wait cursor")
	}
}

func TestPageFooterIncludesSynchronizedCardZoom(t *testing.T) {
	requiredSnippets := []string{
		"var minZoomDragPixels = 4;",
		"function setCardZoomFromSelection(card, rect) {",
		"function resetCardZoom(card) {",
		"surface.querySelectorAll('.zoom-transform').forEach(function(layer) {",
		"if (width > minZoomDragPixels && height > minZoomDragPixels) {",
		"surface.addEventListener('contextmenu', function(e) {",
		"resetCardZoom(surface.closest('.card'));",
		"if (drag.surface.classList.contains('slider-wrap') && typeof drag.surface.__setSliderClientX === 'function') {",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(pageFooter, snippet) {
			t.Fatalf("pageFooter missing synchronized zoom snippet %q", snippet)
		}
	}
}

func TestBuildViewOptionsDisablesRerenderForParityDir(t *testing.T) {
	opts := buildViewOptions(true, "/repo", "/repo/testdata/golden")
	if opts.CanRerender {
		t.Fatal("expected rerender to be disabled in parity mode")
	}
	if !strings.Contains(opts.RerenderDisabledMsg, "--parity-dir") {
		t.Fatalf("unexpected disabled message: %q", opts.RerenderDisabledMsg)
	}
}

func TestBuildViewOptionsDisablesRerenderForCustomArtifactDir(t *testing.T) {
	opts := buildViewOptions(false, "/repo", "/repo/custom-artifacts")
	if opts.CanRerender {
		t.Fatal("expected rerender to be disabled for custom artifact dir")
	}
	if !strings.Contains(opts.RerenderDisabledMsg, "/repo/custom-artifacts") {
		t.Fatalf("unexpected disabled message: %q", opts.RerenderDisabledMsg)
	}
}

func TestRenderPageDisablesRerenderButtonsWhenUnsupported(t *testing.T) {
	var out bytes.Buffer
	renderPage(&out, loadResult{}, viewOptions{
		CanRerender:         false,
		RerenderDisabledMsg: "rerender disabled here",
	})
	html := out.String()
	if !strings.Contains(html, `id="rerender-all-btn" class="rerender-btn" type="button" disabled`) {
		t.Fatalf("expected bulk rerender button to be disabled: %s", html)
	}
	if !strings.Contains(html, `header-warning`) {
		t.Fatalf("expected disabled warning in header: %s", html)
	}
}
