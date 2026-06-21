package main

import (
	"errors"
	"fmt"
	"html"
	"io"
	"path/filepath"
	"strings"
)

const (
	rerenderAllButtonPlaceholder = "__RERENDER_ALL_BUTTON__"
)

type viewOptions struct {
	CanRerender         bool
	RerenderDisabledMsg string
}

func buildViewOptions(useParity bool, repoRoot, artifactDir string) viewOptions {
	return buildViewOptionsWithExtra(useParity, false, repoRoot, artifactDir)
}

func buildViewOptionsWithExtra(useParity, includeExtra bool, repoRoot, artifactDir string) viewOptions {
	if useParity {
		return viewOptions{
			CanRerender:         false,
			RerenderDisabledMsg: "Re-render is disabled in --parity-dir mode because the button only regenerates ./test goldens.",
		}
	}
	if includeExtra {
		return viewOptions{
			CanRerender:         false,
			RerenderDisabledMsg: "Re-render is disabled in unified mode because web demo artifacts need a different regeneration command.",
		}
	}
	defaultArtifactDir := filepath.Join(repoRoot, "testdata", "golden")
	if !samePath(defaultArtifactDir, artifactDir) {
		return viewOptions{
			CanRerender:         false,
			RerenderDisabledMsg: fmt.Sprintf("Re-render is disabled because the viewer is reading artifacts from %s, but rerender only updates %s.", artifactDir, defaultArtifactDir),
		}
	}
	return viewOptions{CanRerender: true}
}

func ensureRerenderSupported(useParity bool, repoRoot, artifactDir string) error {
	return ensureRerenderSupportedWithExtra(useParity, false, repoRoot, artifactDir)
}

func ensureRerenderSupportedWithExtra(useParity, includeExtra bool, repoRoot, artifactDir string) error {
	opts := buildViewOptionsWithExtra(useParity, includeExtra, repoRoot, artifactDir)
	if opts.CanRerender {
		return nil
	}
	return errors.New(opts.RerenderDisabledMsg)
}

func samePath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

func renderPage(w io.Writer, result loadResult, opts viewOptions) {
	fmt.Fprint(w, strings.Replace(pageHeader, rerenderAllButtonPlaceholder, buildRerenderAllButton(opts), 1))
	fmt.Fprintf(
		w,
		`<div class="header-meta">%d comparisons loaded`,
		result.ComparedCount,
	)
	if result.SkippedCount > 0 {
		fmt.Fprintf(w, `, %d baselines skipped because no matching artifact exists`, result.SkippedCount)
	}
	if !opts.CanRerender && opts.RerenderDisabledMsg != "" {
		fmt.Fprintf(w, ` <span class="header-warning">%s</span>`, htmlEscape(opts.RerenderDisabledMsg))
	}
	fmt.Fprint(w, `.</div></div>`)

	fmt.Fprint(w, `<div class="container" id="cards-container">`)
	for i := range result.Cases {
		renderCard(w, &result.Cases[i], opts)
	}
	if len(result.Cases) == 0 {
		fmt.Fprint(w, `<div class="empty-state">No parity comparisons found. Point --baseline-dir/--artifact-dir at existing PNG sets.</div>`)
	}
	fmt.Fprint(w, pageFooter)
}

func buildRerenderAllButton(opts viewOptions) string {
	if opts.CanRerender {
		return `<button id="rerender-all-btn" class="rerender-btn" type="button">Re-render Artifacts</button>`
	}
	return fmt.Sprintf(
		`<button id="rerender-all-btn" class="rerender-btn" type="button" disabled title=%q>Re-render Artifacts</button>`,
		htmlEscape(opts.RerenderDisabledMsg),
	)
}

func renderCard(w io.Writer, entry *caseEntry, opts viewOptions) {
	if entry == nil {
		return
	}

	fmt.Fprintf(
		w,
		`<div class="card" data-name="%s" data-suite="%s" data-baseline="%s" data-rmse="%.4f" data-avg-diff="%.4f" data-max-diff="%d" data-diff-pixels="%d" data-diff-ratio="%.6f">`,
		htmlEscape(entry.Name),
		htmlEscape(entry.Suite),
		htmlEscape(entry.Baseline),
		entry.RMSE,
		entry.AvgDiff,
		entry.MaxDiff,
		entry.DiffPixels,
		entry.DiffRatio,
	)
	fmt.Fprint(w, `<div class="card-header">`)
	fmt.Fprint(w, `<span class="badge badge-neutral sort-metric-badge" style="display:none"></span>`)
	fmt.Fprintf(w, `<span class="card-title">%s</span>`, htmlEscape(entry.Name))
	if opts.CanRerender {
		fmt.Fprintf(
			w,
			`<button class="rerender-btn" data-suite="%s" data-name="%s" type="button">Re-render Artifact</button>`,
			htmlEscape(entry.Suite),
			htmlEscape(entry.Name),
		)
	} else {
		fmt.Fprintf(
			w,
			`<button class="rerender-btn" data-suite="%s" data-name="%s" type="button" disabled title="%s">Re-render Artifact</button>`,
			htmlEscape(entry.Suite),
			htmlEscape(entry.Name),
			htmlEscape(opts.RerenderDisabledMsg),
		)
	}
	fmt.Fprint(w, `<div class="right-badges">`)
	fmt.Fprintf(w, `<span class="badge badge-neutral">%s</span>`, htmlEscape(entry.Suite))
	fmt.Fprintf(w, `<span class="badge badge-neutral">%s</span>`, htmlEscape(entry.Baseline))
	fmt.Fprintf(w, `<span class="badge %s">RMSE %.2f</span>`, badgeClassRMSE(entry.RMSE), entry.RMSE)
	fmt.Fprintf(w, `<span class="badge %s">avg %.2f</span>`, badgeClassAvgDiff(entry.AvgDiff), entry.AvgDiff)
	fmt.Fprintf(w, `<span class="badge %s">max %d</span>`, badgeClassMaxDiff(entry.MaxDiff), entry.MaxDiff)
	fmt.Fprintf(w, `<span class="badge %s">diff %.2f%%</span>`, badgeClassDiffRatio(entry.DiffRatio), entry.DiffRatio*100)
	fmt.Fprint(w, `</div></div>`)

	fmt.Fprint(w, `<div class="card-body"><div class="card-meta">`)
	fmt.Fprintf(w, `size baseline %dx%d, artifact %dx%d`, entry.RefWidth, entry.RefHeight, entry.ActWidth, entry.ActHeight)
	fmt.Fprint(w, `</div><div class="img-grid">`)

	fmt.Fprint(w, `<div class="img-col col-ref">`)
	fmt.Fprint(w, `<label>Baseline</label>`)
	fmt.Fprint(w, `<div class="zoom-surface">`)
	fmt.Fprint(w, `<div class="zoom-transform">`)
	fmt.Fprintf(w, `<img class="parity-image" data-src="%s" alt="baseline">`, htmlEscape(entry.RefImageURL))
	fmt.Fprint(w, `</div><div class="zoom-selection"></div></div>`)
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `<div class="img-col col-artifact">`)
	fmt.Fprint(w, `<label>Artifact</label>`)
	fmt.Fprint(w, `<div class="zoom-surface">`)
	fmt.Fprint(w, `<div class="zoom-transform">`)
	fmt.Fprintf(w, `<img class="parity-image" data-src="%s" alt="artifact">`, htmlEscape(entry.ActImageURL))
	fmt.Fprint(w, `</div><div class="zoom-selection"></div></div>`)
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `<div class="img-col col-overlay">`)
	fmt.Fprint(w, `<label>Overlay</label>`)
	fmt.Fprint(w, `<div class="slider-wrap zoom-surface">`)
	fmt.Fprint(w, `<div class="zoom-transform zoom-base-layer">`)
	fmt.Fprintf(w, `<img class="base" data-src="%s" alt="base">`, htmlEscape(entry.RefImageURL))
	fmt.Fprint(w, `</div>`)
	fmt.Fprint(w, `<div class="slider-overlay">`)
	fmt.Fprint(w, `<div class="zoom-transform zoom-overlay-layer">`)
	fmt.Fprintf(w, `<img data-src="%s" alt="overlay">`, htmlEscape(entry.ActImageURL))
	fmt.Fprint(w, `</div></div><div class="slider-divider"></div><div class="zoom-selection"></div></div></div>`)

	fmt.Fprint(w, `<div class="img-col col-amp">`)
	fmt.Fprint(w, `<label>Diff amplified</label>`)
	fmt.Fprint(w, `<div class="zoom-surface">`)
	fmt.Fprint(w, `<div class="zoom-transform">`)
	fmt.Fprintf(w, `<img class="parity-image" data-src="%s" alt="amplified-diff">`, htmlEscape(entry.AmpDiffURL))
	fmt.Fprint(w, `</div><div class="zoom-selection"></div></div>`)
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `<div class="img-col col-raw">`)
	fmt.Fprint(w, `<label>Diff raw</label>`)
	fmt.Fprint(w, `<div class="zoom-surface">`)
	fmt.Fprint(w, `<div class="zoom-transform">`)
	fmt.Fprintf(w, `<img class="parity-image" data-src="%s" alt="raw-diff">`, htmlEscape(entry.RawDiffURL))
	fmt.Fprint(w, `</div><div class="zoom-selection"></div></div>`)
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `</div></div></div>`)
}

func htmlEscape(s string) string {
	return html.EscapeString(s)
}

const (
	badgeClassOK   = "badge-ok"
	badgeClassWarn = "badge-warn"
	badgeClassBad  = "badge-bad"
)

func badgeClassRMSE(v float64) string {
	if v <= 5 {
		return badgeClassOK
	}
	if v <= 20 {
		return badgeClassWarn
	}
	return badgeClassBad
}

func badgeClassAvgDiff(v float64) string {
	if v <= 2 {
		return badgeClassOK
	}
	if v <= 8 {
		return badgeClassWarn
	}
	return badgeClassBad
}

func badgeClassMaxDiff(v uint8) string {
	if v <= 10 {
		return badgeClassOK
	}
	if v <= 40 {
		return badgeClassWarn
	}
	return badgeClassBad
}

func badgeClassDiffRatio(v float64) string {
	if v <= 0.01 {
		return badgeClassOK
	}
	if v <= 0.05 {
		return badgeClassWarn
	}
	return badgeClassBad
}

const pageHeader = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Matplotlib-Go Parity Viewer</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { background: #101216; color: #d7dce4; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 13px; }
.sticky-header { position: sticky; top: 0; z-index: 100; background: rgba(16,18,22,0.96); backdrop-filter: blur(10px); border-bottom: 1px solid #2d3440; padding: 10px 12px; }
.controls { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.controls h1 { font-size: 15px; color: #f4f7fb; margin-right: 8px; }
.controls input, .controls select { background: #171b22; color: #d7dce4; border: 1px solid #334155; padding: 5px 8px; font-family: inherit; font-size: 12px; border-radius: 6px; }
.controls label { display: flex; align-items: center; gap: 5px; font-size: 12px; color: #aeb8c8; }
.header-meta { margin-top: 8px; color: #93a1b5; font-size: 12px; }
.rerender-progress { display: none; align-items: center; gap: 8px; margin-top: 8px; color: #93a1b5; font-size: 12px; }
.rerender-progress.is-active { display: flex; }
.rerender-progress progress { width: 220px; height: 10px; accent-color: #7cf0a5; }
#summary { margin-left: auto; color: #93a1b5; font-size: 12px; }
.container { padding: 12px; }
.card { background: #141922; border: 1px solid #293241; margin-bottom: 10px; border-radius: 10px; overflow: hidden; box-shadow: 0 8px 24px rgba(0,0,0,0.2); }
.card-header { padding: 9px 12px; cursor: pointer; display: flex; align-items: center; gap: 8px; background: #181f2a; user-select: none; }
.card-header:hover { background: #1d2531; }
.card-body { display: none; padding: 12px; }
.card.open .card-body { display: block; }
.card-title { font-size: 13px; color: #f4f7fb; flex: 1; }
.card-meta { color: #93a1b5; margin-bottom: 10px; }
.right-badges { display: flex; gap: 6px; align-items: center; flex-wrap: wrap; }
.badge { padding: 2px 7px; border-radius: 999px; font-size: 11px; font-weight: bold; }
.badge-neutral { background: #202734; color: #d7dce4; border: 1px solid #425168; }
.badge-ok { background: #163323; color: #7cf0a5; border: 1px solid #2e7250; }
.badge-warn { background: #3a2d13; color: #ffc66d; border: 1px solid #6f5521; }
.badge-bad { background: #3b1619; color: #ff8a8f; border: 1px solid #7c2d35; }
.img-grid { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 10px; align-items: start; overflow-x: auto; }
.img-col { display: flex; flex-direction: column; gap: 6px; min-width: 0; overflow: auto; }
.img-col label { font-size: 11px; color: #93a1b5; text-align: center; }
.zoom-surface { position: relative; overflow: hidden; border-radius: 6px; background-color: #fff; cursor: crosshair; max-width: 100%; }
.zoom-transform { position: relative; transform-origin: 0 0; will-change: transform; }
.zoom-selection { display: none; position: absolute; border: 1px solid #93c5fd; background: rgba(96, 165, 250, 0.18); pointer-events: none; z-index: 4; }
.parity-image { display: block; image-rendering: auto; width: 100%; height: auto; background-color: #fff; background-image: none; max-width: 100%; }
.resample-pixelated .parity-image { image-rendering: pixelated; }
.original-size .img-grid { grid-template-columns: repeat(5, max-content); }
.original-size .img-col { min-width: max-content; }
.original-size .zoom-surface, .original-size .slider-wrap { align-self: flex-start; }
.original-size .parity-image { width: auto; height: auto; max-width: none; }
.col-raw, .col-amp { display: none; }
.slider-wrap { width: 100%; background-color: #fff; background-image: none; }
.slider-wrap img.base, .slider-wrap .zoom-overlay-layer img { display: block; image-rendering: auto; width: 100%; height: auto; }
.resample-pixelated .slider-wrap img.base { image-rendering: pixelated; }
.resample-pixelated .slider-wrap .zoom-overlay-layer img { image-rendering: pixelated; }
.original-size .slider-wrap, .original-size .zoom-surface { width: auto; }
.original-size .slider-wrap img.base { width: auto; height: auto; max-width: none; }
.original-size .slider-wrap .zoom-overlay-layer img { width: auto; height: auto; max-width: none; }
.slider-overlay { position: absolute; top: 0; left: 0; height: 100%; overflow: hidden; width: 50%; z-index: 2; pointer-events: none; }
.slider-divider { position: absolute; top: 0; left: 50%; height: 100%; width: 3px; background: #f8fafc; cursor: col-resize; transform: translateX(-50%); }
.slider-divider::before {
  content: ''; position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%);
  width: 18px; height: 18px; border-radius: 50%; background: #f8fafc; border: 2px solid #111827;
}
.zoom-dragging, .zoom-dragging * { user-select: none; }
.empty-state { border: 1px dashed #425168; border-radius: 10px; padding: 24px; color: #93a1b5; background: #141922; }
button.rerender-btn { background: #0f172a; color: #e2e8f0; border: 1px solid #334155; border-radius: 6px; padding: 4px 8px; cursor: pointer; font-family: inherit; font-size: 11px; }
button.rerender-btn:hover { background: #273244; }
button.rerender-btn:disabled { opacity: 0.6; cursor: not-allowed; }
button.rerender-btn.is-rerendering { cursor: wait; }
code { color: #f4f7fb; }
@media (max-width: 1400px) { .img-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 800px) { .img-grid { grid-template-columns: 1fr; } }
</style>
</head>
<body>
<div class="sticky-header">
  <div class="controls">
    <h1>Matplotlib-Go Parity Viewer</h1>
    <input type="text" id="search" placeholder="Search case…" oninput="filterCards()" style="width:180px">
    <select id="sort-select" onchange="sortCards()">
      <option value="rmse-desc">Sort: RMSE ↓</option>
      <option value="rmse-asc">Sort: RMSE ↑</option>
      <option value="diff-pixels-desc">Sort: Different pixels ↓</option>
      <option value="diff-pixels-asc">Sort: Different pixels ↑</option>
      <option value="diff-ratio-desc">Sort: Diff ratio ↓</option>
      <option value="diff-ratio-asc">Sort: Diff ratio ↑</option>
      <option value="avg-diff-desc">Sort: Avg diff ↓</option>
      <option value="avg-diff-asc">Sort: Avg diff ↑</option>
      <option value="name-asc">Sort: Name ↑</option>
    </select>
    <select id="diff-mode" onchange="setDiffMode(this.value)">
      <option value="amp">Diff: amplified</option>
      <option value="raw">Diff: raw</option>
      <option value="both">Diff: both</option>
    </select>
    <select id="resample-mode" onchange="setResampleMode(this.value)">
      <option value="smooth">Scaling: smooth</option>
      <option value="pixelated">Scaling: pixelated</option>
    </select>
    __RERENDER_ALL_BUTTON__
    <label><input type="checkbox" id="original-size" onchange="setOriginalSize(this.checked)"> Original size</label>
    <span id="summary"></span>
  </div>
  <div id="rerender-progress-row" class="rerender-progress">
    <progress id="rerender-progress" value="0" max="1"></progress>
    <span id="rerender-progress-text">Idle</span>
  </div>
`

const pageFooter = `</div>
<script>
(function() {
  var viewerStateStorageKey = 'mpl-parity-viewer-state-v1';
  var viewerStateControlIDs = ['search', 'sort-select', 'diff-mode', 'resample-mode', 'original-size'];
  var minZoomDragPixels = 4;
  var activeSlider = null;
  var activeZoomSelection = null;

  function cardStateKey(card) {
    return [card.dataset.suite || '', card.dataset.baseline || '', card.dataset.name || ''].join('::');
  }

  function loadViewerState() {
    try {
      var raw = window.sessionStorage.getItem(viewerStateStorageKey);
      if (!raw) return {};
      var parsed = JSON.parse(raw);
      if (!parsed || typeof parsed !== 'object') return {};
      return parsed;
    } catch (err) {
      return {};
    }
  }

  function saveViewerState() {
    var state = {};
    viewerStateControlIDs.forEach(function(id) {
      var el = document.getElementById(id);
      if (!el) return;
      state[id] = el.type === 'checkbox' ? el.checked : el.value;
    });
    state.openCards = Array.from(document.querySelectorAll('.card.open')).map(cardStateKey);
    try {
      window.sessionStorage.setItem(viewerStateStorageKey, JSON.stringify(state));
    } catch (err) {
    }
  }

  function restoreViewerState() {
    var state = loadViewerState();
    viewerStateControlIDs.forEach(function(id) {
      if (!Object.prototype.hasOwnProperty.call(state, id)) return;
      var el = document.getElementById(id);
      if (!el) return;
      if (el.type === 'checkbox') {
        el.checked = !!state[id];
        return;
      }
      if (typeof state[id] === 'string') {
        el.value = state[id];
      }
    });
    var openCards = Array.isArray(state.openCards) ? state.openCards : [];
    var openCardSet = new Set(openCards);
    document.querySelectorAll('.card').forEach(function(card) {
      card.classList.toggle('open', openCardSet.has(cardStateKey(card)));
    });
  }

  function bindViewerStatePersistence() {
    viewerStateControlIDs.forEach(function(id) {
      var el = document.getElementById(id);
      if (!el) return;
      el.addEventListener(el.type === 'text' ? 'input' : 'change', saveViewerState);
    });
  }

  function metric(card, attr) {
    return parseFloat(card.dataset[attr] || 0);
  }

  function clamp(value, minValue, maxValue) {
    return Math.min(Math.max(value, minValue), maxValue);
  }

  function filterCards() {
    var q = document.getElementById('search').value.toLowerCase();
    document.querySelectorAll('.card').forEach(function(card) {
      var name = (card.dataset.name || '').toLowerCase();
      card.style.display = name.indexOf(q) >= 0 ? '' : 'none';
    });
    updateSummary();
  }

  function sortCards() {
    var mode = document.getElementById('sort-select').value;
    var container = document.getElementById('cards-container');
    var cards = Array.from(container.querySelectorAll('.card'));
    cards.sort(function(a, b) {
      if (mode === 'rmse-desc') return metric(b, 'rmse') - metric(a, 'rmse');
      if (mode === 'rmse-asc') return metric(a, 'rmse') - metric(b, 'rmse');
      if (mode === 'diff-pixels-desc') return metric(b, 'diffPixels') - metric(a, 'diffPixels');
      if (mode === 'diff-pixels-asc') return metric(a, 'diffPixels') - metric(b, 'diffPixels');
      if (mode === 'diff-ratio-desc') return metric(b, 'diffRatio') - metric(a, 'diffRatio');
      if (mode === 'diff-ratio-asc') return metric(a, 'diffRatio') - metric(b, 'diffRatio');
      if (mode === 'avg-diff-desc') return metric(b, 'avgDiff') - metric(a, 'avgDiff');
      if (mode === 'avg-diff-asc') return metric(a, 'avgDiff') - metric(b, 'avgDiff');
      return (a.dataset.name || '').localeCompare(b.dataset.name || '');
    });
    cards.forEach(function(card) { container.appendChild(card); });
    updateSortMetricBadges(mode);
  }

  function badgeColorClass(attr, value, diffRatio) {
    if (attr === 'rmse') {
      if (value <= 5) return 'badge-ok';
      if (value <= 20) return 'badge-warn';
      return 'badge-bad';
    }
    if (attr === 'avgDiff') {
      if (value <= 2) return 'badge-ok';
      if (value <= 8) return 'badge-warn';
      return 'badge-bad';
    }
    if (attr === 'maxDiff') {
      if (value <= 10) return 'badge-ok';
      if (value <= 40) return 'badge-warn';
      return 'badge-bad';
    }
    if (attr === 'diffPixels' || attr === 'diffRatio') {
      var ratio = parseFloat(diffRatio || 0);
      if (ratio <= 0.01) return 'badge-ok';
      if (ratio <= 0.05) return 'badge-warn';
      return 'badge-bad';
    }
    return 'badge-neutral';
  }

  function updateSortMetricBadges(mode) {
    var label = '';
    var attr = '';
    var format = function(v) { return String(v); };
    if (mode === 'rmse-desc' || mode === 'rmse-asc') {
      label = 'RMSE'; attr = 'rmse'; format = function(v) { return Number(v).toFixed(2); };
    } else if (mode === 'diff-pixels-desc' || mode === 'diff-pixels-asc') {
      label = 'diff px'; attr = 'diffPixels'; format = function(v) { return String(Math.round(Number(v))); };
    } else if (mode === 'diff-ratio-desc' || mode === 'diff-ratio-asc') {
      label = 'diff %'; attr = 'diffRatio'; format = function(v) { return (Number(v) * 100).toFixed(2); };
    } else if (mode === 'avg-diff-desc' || mode === 'avg-diff-asc') {
      label = 'avg'; attr = 'avgDiff'; format = function(v) { return Number(v).toFixed(2); };
    }
    document.querySelectorAll('.card').forEach(function(card) {
      var badge = card.querySelector('.sort-metric-badge');
      if (!badge) return;
      if (!attr) {
        badge.style.display = 'none';
        return;
      }
      var value = parseFloat(card.dataset[attr] || 0);
      badge.className = 'badge ' + badgeColorClass(attr, value, card.dataset.diffRatio) + ' sort-metric-badge';
      badge.textContent = label + ' ' + format(value);
      badge.style.display = '';
    });
  }

  function setDiffMode(mode) {
    document.querySelectorAll('.col-amp').forEach(function(el) {
      el.style.display = (mode === 'amp' || mode === 'both') ? 'flex' : 'none';
    });
    document.querySelectorAll('.col-raw').forEach(function(el) {
      el.style.display = (mode === 'raw' || mode === 'both') ? 'flex' : 'none';
    });
    refreshCardZooms();
  }

  function updateSummary() {
    var all = document.querySelectorAll('.card');
    var visible = Array.from(all).filter(function(card) { return card.style.display !== 'none'; });
    document.getElementById('summary').textContent = visible.length + ' / ' + all.length + ' cases';
  }

  function setResampleMode(mode) {
    var container = document.getElementById('cards-container');
    container.classList.remove('resample-smooth', 'resample-pixelated');
    container.classList.add(mode === 'pixelated' ? 'resample-pixelated' : 'resample-smooth');
  }

  function setOriginalSize(on) {
    var container = document.getElementById('cards-container');
    if (on) container.classList.add('original-size'); else container.classList.remove('original-size');
    refreshCardZooms();
  }

  function setRerenderButtonsDisabled(disabled) {
    document.querySelectorAll('.rerender-btn').forEach(function(button) {
      button.disabled = disabled;
      button.classList.toggle('is-rerendering', disabled);
    });
  }

  function loadCardImages(card) {
    if (!card) return;
    card.querySelectorAll('img[data-src]').forEach(function(img) {
      if (img.getAttribute('src')) return;
      img.setAttribute('src', img.dataset.src);
    });
  }

  function updateRerenderProgress(status) {
    var row = document.getElementById('rerender-progress-row');
    var bar = document.getElementById('rerender-progress');
    var text = document.getElementById('rerender-progress-text');
    if (!row || !bar || !text) return;
    row.classList.add('is-active');
    var total = Number(status.total || 0);
    var completed = Number(status.completed || 0);
    bar.max = total > 0 ? total : 1;
    bar.value = Math.min(completed, bar.max);
    var current = status.current ? ' · ' + status.current : '';
    if (status.status === 'failed') {
      text.textContent = 'Failed after ' + completed + ' / ' + total + current;
      return;
    }
    if (status.status === 'done') {
      text.textContent = 'Completed ' + completed + ' / ' + total;
      return;
    }
    text.textContent = 'Rendering ' + completed + ' / ' + total + current;
  }

  function startRerender(names) {
    var params = new URLSearchParams();
    names.forEach(function(name) {
      params.append('name', name);
    });
    return fetch('/rerender', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8' },
      body: params.toString()
    }).then(function(response) {
      if (!response.ok) {
        return response.text().then(function(text) {
          throw new Error(text || 'rerender failed');
        });
      }
      return response.json();
    });
  }

  function pollRerenderJob(jobID) {
    return fetch('/rerender/status?id=' + encodeURIComponent(jobID), {
      headers: { 'Accept': 'application/json' }
    }).then(function(response) {
      if (!response.ok) {
        return response.text().then(function(text) {
          throw new Error(text || 'rerender status failed');
        });
      }
      return response.json();
    }).then(function(status) {
      updateRerenderProgress(status);
      if (status.status === 'done') {
        return status;
      }
      if (status.status === 'failed') {
        throw new Error(status.error || 'rerender failed');
      }
      return new Promise(function(resolve) {
        window.setTimeout(resolve, 500);
      }).then(function() {
        return pollRerenderJob(jobID);
      });
    });
  }

  function navigateToFreshPage() {
    var url = new URL(window.location.href);
    url.searchParams.set('_pv', String(Date.now()));
    window.location.assign(url.toString());
  }

  document.querySelectorAll('.card-header').forEach(function(header) {
    header.addEventListener('click', function() {
      var card = header.closest('.card');
      card.classList.toggle('open');
      if (card.classList.contains('open')) {
        loadCardImages(card);
      }
      saveViewerState();
    });
  });

  document.querySelectorAll('.rerender-btn').forEach(function(button) {
    if (button.id === 'rerender-all-btn') {
      return;
    }
    button.addEventListener('click', function(event) {
      event.stopPropagation();
      var name = button.dataset.name || '';
      if (!name) return;
      saveViewerState();
      setRerenderButtonsDisabled(true);
      startRerender([name]).then(function(job) {
        updateRerenderProgress(job);
        return pollRerenderJob(job.job_id || job.id);
      }).then(function() {
        navigateToFreshPage();
      }).catch(function(err) {
        window.alert(err.message);
        setRerenderButtonsDisabled(false);
      });
    });
  });

  var bulkButton = document.getElementById('rerender-all-btn');
  bulkButton.addEventListener('click', function() {
    saveViewerState();
    setRerenderButtonsDisabled(true);
    var names = Array.from(document.querySelectorAll('.card')).map(function(card) {
      return card.dataset.name || '';
    }).filter(function(name) {
      return !!name;
    });
    return startRerender(names).then(function(job) {
      updateRerenderProgress(job);
      return pollRerenderJob(job.job_id || job.id);
    }).then(function() {
      navigateToFreshPage();
    }).catch(function(err) {
      window.alert(err.message);
      setRerenderButtonsDisabled(false);
    });
  });

  function ensureCardZoomState(card) {
    if (!card.__zoomState) {
      card.__zoomState = { scale: 1, x: 0, y: 0 };
    }
    return card.__zoomState;
  }

  function applyCardZoom(card) {
    var state = ensureCardZoomState(card);
    card.querySelectorAll('.zoom-surface').forEach(function(surface) {
      var width = surface.clientWidth;
      var height = surface.clientHeight;
      surface.querySelectorAll('.zoom-transform').forEach(function(layer) {
        if (!width || !height || state.scale <= 1) {
          layer.style.transform = '';
          return;
        }
        var tx = -state.x * width * state.scale;
        var ty = -state.y * height * state.scale;
        layer.style.transform = 'matrix(' + state.scale + ',0,0,' + state.scale + ',' + tx + ',' + ty + ')';
      });
    });
  }

  function refreshCardZooms() {
    document.querySelectorAll('.card').forEach(applyCardZoom);
  }

  function setCardZoomFromSelection(card, rect) {
    var width = clamp(rect.width, 0, 1);
    var height = clamp(rect.height, 0, 1);
    if (width <= 0 || height <= 0) {
      return;
    }
    var scale = 1 / Math.max(width, height);
    var visibleWidth = 1 / scale;
    var visibleHeight = 1 / scale;
    var centerX = rect.x + width / 2;
    var centerY = rect.y + height / 2;
    var state = ensureCardZoomState(card);
    state.scale = scale;
    state.x = clamp(centerX - visibleWidth / 2, 0, Math.max(0, 1 - visibleWidth));
    state.y = clamp(centerY - visibleHeight / 2, 0, Math.max(0, 1 - visibleHeight));
    applyCardZoom(card);
  }

  function resetCardZoom(card) {
    var state = ensureCardZoomState(card);
    state.scale = 1;
    state.x = 0;
    state.y = 0;
    applyCardZoom(card);
  }

  function showSelectionBox(selection, x0, y0, x1, y1) {
    var left = Math.min(x0, x1);
    var top = Math.min(y0, y1);
    selection.style.display = 'block';
    selection.style.left = left + 'px';
    selection.style.top = top + 'px';
    selection.style.width = Math.abs(x1 - x0) + 'px';
    selection.style.height = Math.abs(y1 - y0) + 'px';
  }

  function hideSelectionBox(selection) {
    if (!selection) return;
    selection.style.display = 'none';
    selection.style.left = '0';
    selection.style.top = '0';
    selection.style.width = '0';
    selection.style.height = '0';
  }

  document.querySelectorAll('.slider-wrap').forEach(function(wrap) {
    var divider = wrap.querySelector('.slider-divider');
    var overlay = wrap.querySelector('.slider-overlay');
    function applyPos(pct) {
      pct = Math.max(0, Math.min(1, pct));
      wrap.__sliderPct = pct;
      overlay.style.width = (pct * 100) + '%';
      divider.style.left = (pct * 100) + '%';
      var overlayLayer = wrap.querySelector('.zoom-overlay-layer');
      if (overlayLayer) {
        if (pct <= 0) {
          overlayLayer.style.width = '0%';
          return;
        }
        overlayLayer.style.width = (100 / pct) + '%';
      }
    }
    function setPos(x) {
      var rect = wrap.getBoundingClientRect();
      if (rect.width <= 0) {
        applyPos(0.5);
        return;
      }
      applyPos((x - rect.left) / rect.width);
    }
    wrap.__setSliderClientX = setPos;
    divider.addEventListener('mousedown', function(e) {
      activeSlider = { wrap: wrap, setPos: setPos };
      e.preventDefault();
      e.stopPropagation();
    });
    applyPos(0.5);
  });

  document.addEventListener('mousemove', function(e) {
    if (activeSlider) {
      activeSlider.setPos(e.clientX);
    }
    if (!activeZoomSelection) {
      return;
    }
    var rect = activeZoomSelection.rect;
    var x = clamp(e.clientX - rect.left, 0, rect.width);
    var y = clamp(e.clientY - rect.top, 0, rect.height);
    activeZoomSelection.currentX = x;
    activeZoomSelection.currentY = y;
    showSelectionBox(activeZoomSelection.selection, activeZoomSelection.startX, activeZoomSelection.startY, x, y);
  });

  document.addEventListener('mouseup', function(e) {
    if (activeSlider && e.button === 0) {
      activeSlider = null;
    }
    if (!activeZoomSelection || e.button !== 0) {
      return;
    }
    var drag = activeZoomSelection;
    var width = Math.abs(drag.currentX - drag.startX);
    var height = Math.abs(drag.currentY - drag.startY);
    hideSelectionBox(drag.selection);
    document.body.classList.remove('zoom-dragging');
    activeZoomSelection = null;

    if (width > minZoomDragPixels && height > minZoomDragPixels) {
      setCardZoomFromSelection(drag.card, {
        x: Math.min(drag.startX, drag.currentX) / drag.rect.width,
        y: Math.min(drag.startY, drag.currentY) / drag.rect.height,
        width: width / drag.rect.width,
        height: height / drag.rect.height,
      });
      return;
    }

    if (drag.surface.classList.contains('slider-wrap') && typeof drag.surface.__setSliderClientX === 'function') {
      drag.surface.__setSliderClientX(e.clientX);
    }
  });

  document.querySelectorAll('.zoom-surface').forEach(function(surface) {
    surface.addEventListener('mousedown', function(e) {
      if (e.button !== 0) {
        return;
      }
      if (e.target.closest('.slider-divider')) {
        return;
      }
      var rect = surface.getBoundingClientRect();
      if (rect.width <= 0 || rect.height <= 0) {
        return;
      }
      var selection = surface.querySelector('.zoom-selection');
      if (!selection) {
        return;
      }
      var startX = clamp(e.clientX - rect.left, 0, rect.width);
      var startY = clamp(e.clientY - rect.top, 0, rect.height);
      activeZoomSelection = {
        card: surface.closest('.card'),
        surface: surface,
        rect: rect,
        selection: selection,
        startX: startX,
        startY: startY,
        currentX: startX,
        currentY: startY,
      };
      showSelectionBox(selection, startX, startY, startX, startY);
      document.body.classList.add('zoom-dragging');
      e.preventDefault();
    });
    surface.addEventListener('contextmenu', function(e) {
      e.preventDefault();
      resetCardZoom(surface.closest('.card'));
    });
  });

  window.addEventListener('resize', refreshCardZooms);

  window.filterCards = filterCards;
  window.sortCards = sortCards;
  window.setDiffMode = setDiffMode;
  window.setOriginalSize = setOriginalSize;
  window.setResampleMode = setResampleMode;

  restoreViewerState();
  document.querySelectorAll('.card.open').forEach(loadCardImages);
  bindViewerStatePersistence();
  sortCards();
  setDiffMode(document.getElementById('diff-mode').value);
  setResampleMode(document.getElementById('resample-mode').value);
  setOriginalSize(document.getElementById('original-size').checked);
  refreshCardZooms();
  filterCards();
})();
</script>
</body>
</html>
`
