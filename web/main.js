"use strict";

import {
  buildAPICompatibilityMessage,
  buildRuntimeExitMessage,
  missingAPIMethods,
} from "./runtime.mjs";
import {
  THEME_KEY,
  nextThemeMode,
  normalizeThemeMode,
  resolveThemeMode,
  themeIcon,
  themeLabel,
} from "./theme.mjs";

const go = new Go();
const DEFAULT_WIDTH = 960;
const DEFAULT_HEIGHT = 540;
const GO_KEYWORDS = new Set([
  "break",
  "case",
  "chan",
  "const",
  "continue",
  "default",
  "defer",
  "else",
  "fallthrough",
  "for",
  "func",
  "go",
  "goto",
  "if",
  "import",
  "interface",
  "map",
  "package",
  "range",
  "return",
  "select",
  "struct",
  "switch",
  "type",
  "var",
]);
const GO_PREDECLARED = new Set([
  "append",
  "bool",
  "byte",
  "cap",
  "complex64",
  "complex128",
  "copy",
  "delete",
  "error",
  "false",
  "float32",
  "float64",
  "imag",
  "int",
  "int8",
  "int16",
  "int32",
  "int64",
  "iota",
  "len",
  "make",
  "new",
  "nil",
  "panic",
  "print",
  "println",
  "real",
  "recover",
  "rune",
  "string",
  "true",
  "uint",
  "uint8",
  "uint16",
  "uint32",
  "uint64",
  "uintptr",
]);

let api;
let runtimeExited = false;
let runtimeExitMessage = buildRuntimeExitMessage();
let currentDemoID = "";
let currentBackendID = "";
let renderFrame = 0;
let currentThemeMode = "auto";
const backendNames = new Map();
const sourceCache = new Map();

async function init() {
  initTheme();

  const canvas = document.getElementById("plotCanvas");
  if (!canvas) {
    updateStatus("Failed to load WASM: canvas element was not found");
    return;
  }

  try {
    let result;
    try {
      result = await WebAssembly.instantiateStreaming(
        fetch("main.wasm"),
        go.importObject,
      );
    } catch (streamError) {
      const response = await fetch("main.wasm");
      const bytes = await response.arrayBuffer();
      result = await WebAssembly.instantiate(bytes, go.importObject);
      console.warn(
        "instantiateStreaming failed, fell back to ArrayBuffer",
        streamError,
      );
    }

    monitorRuntime(go.run(result.instance));
    api = await waitForAPI();
    ensureCompatibleAPI(api);

    const demos = api.listDemos();
    const backends = api.listBackends();
    populateSelector(demos, api.defaultDemoID());
    populateBackendSelector(backends, api.defaultBackendID());

    document
      .getElementById("downloadBtn")
      .addEventListener("click", () => downloadCurrentPlot());
    document
      .getElementById("sourceBtn")
      .addEventListener("click", () => toggleSourcePanel());
    document
      .getElementById("closeSourceBtn")
      .addEventListener("click", () => hideSourcePanel());
    document
      .getElementById("demoSelector")
      .addEventListener("change", () => scheduleRender());
    document
      .getElementById("backendSelector")
      .addEventListener("change", () => scheduleRender());

    document
      .getElementById("navHomeBtn")
      .addEventListener("click", () => triggerNavAction("home"));
    document
      .getElementById("navPanBtn")
      .addEventListener("click", () => toggleNavMode("pan"));
    document
      .getElementById("navZoomBtn")
      .addEventListener("click", () => toggleNavMode("zoom"));

    window.addEventListener("resize", () => scheduleRender());

    scheduleRender();
  } catch (error) {
    updateStatus(`Failed to load WASM: ${error.message}`);
    console.error(error);
  }
}

function initTheme() {
  currentThemeMode = readStoredThemeMode();
  applyTheme(currentThemeMode);
  updateThemeToggleButton();

  document.getElementById("themeToggle")?.addEventListener("click", () => {
    cycleTheme();
  });

  const systemScheme = window.matchMedia("(prefers-color-scheme: dark)");
  systemScheme.addEventListener("change", () => {
    if (currentThemeMode === "auto") {
      applyTheme("auto");
    }
  });
}

function readStoredThemeMode() {
  try {
    return normalizeThemeMode(localStorage.getItem(THEME_KEY));
  } catch {
    return "auto";
  }
}

function storeThemeMode(mode) {
  try {
    localStorage.setItem(THEME_KEY, mode);
  } catch {
    // Theme persistence is optional; the active DOM theme still changes.
  }
}

function getSystemTheme() {
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

function applyTheme(mode) {
  const normalizedMode = normalizeThemeMode(mode);
  const resolvedTheme = resolveThemeMode(normalizedMode, getSystemTheme());
  document.documentElement.dataset.themeMode = normalizedMode;
  if (resolvedTheme === "dark") {
    document.documentElement.dataset.theme = "dark";
    return;
  }
  delete document.documentElement.dataset.theme;
}

function updateThemeToggleButton() {
  const toggle = document.getElementById("themeToggle");
  const label = toggle?.querySelector(".theme-toggle-label");
  const icon = toggle?.querySelector(".theme-toggle-icon");
  if (!toggle || !label || !icon) {
    return;
  }

  label.textContent = themeLabel(currentThemeMode);
  icon.innerHTML = themeIcon(currentThemeMode);
  toggle.title = `Theme: ${themeLabel(currentThemeMode)}`;
  toggle.setAttribute("aria-label", `Theme: ${themeLabel(currentThemeMode)}`);
}

function cycleTheme() {
  currentThemeMode = nextThemeMode(currentThemeMode);
  applyTheme(currentThemeMode);
  updateThemeToggleButton();
  storeThemeMode(currentThemeMode);
}

function waitForAPI() {
  return new Promise((resolve, reject) => {
    const startedAt = Date.now();

    function poll() {
      if (runtimeExited) {
        reject(new Error(runtimeExitMessage));
        return;
      }
      if (window.matplotlibGoWASM) {
        resolve(window.matplotlibGoWASM);
        return;
      }
      if (Date.now() - startedAt > 5000) {
        reject(new Error("matplotlibGoWASM API was not registered"));
        return;
      }
      window.setTimeout(poll, 20);
    }

    poll();
  });
}

function ensureCompatibleAPI(candidate) {
  const missingMethods = missingAPIMethods(candidate);
  if (missingMethods.length > 0) {
    throw new Error(buildAPICompatibilityMessage(missingMethods));
  }
}

function monitorRuntime(runPromise) {
  Promise.resolve(runPromise)
    .then(() => {
      runtimeExited = true;
      runtimeExitMessage = buildRuntimeExitMessage();
      updateStatus(runtimeExitMessage);
    })
    .catch((error) => {
      runtimeExited = true;
      runtimeExitMessage = buildRuntimeExitMessage(error);
      updateStatus(runtimeExitMessage);
      console.error(error);
    });
}

function populateSelector(demos, selectedID) {
  const selector = document.getElementById("demoSelector");
  selector.innerHTML = "";

  for (const demo of demos) {
    const option = document.createElement("option");
    option.value = demo.id;
    option.textContent = demo.title;
    if (demo.id === selectedID) {
      option.selected = true;
    }
    selector.appendChild(option);
  }
}

function populateBackendSelector(backends, selectedID) {
  const selector = document.getElementById("backendSelector");
  selector.innerHTML = "";
  backendNames.clear();

  for (const backend of backends) {
    const option = document.createElement("option");
    option.value = backend.id;
    option.textContent = backend.name;
    option.title = backend.description;
    if (backend.id === selectedID) {
      option.selected = true;
    }
    selector.appendChild(option);
    backendNames.set(backend.id, backend.name);
  }
}

function mountSelectedDemo() {
  if (runtimeExited) {
    updateStatus(runtimeExitMessage);
    return;
  }

  const selector = document.getElementById("demoSelector");
  const demoID = selector.value || api.defaultDemoID();
  const backendSelector = document.getElementById("backendSelector");
  const backendID = backendSelector.value || api.defaultBackendID();
  const backendName = backendNames.get(backendID) || backendID;
  const size = plotSize();

  updateStatus(`Rendering ${demoID} with ${backendName}…`);
  setDownloadEnabled(false);
  setSourceEnabled(false);

  const startedAt = performance.now();
  const result = api.mountDemo(
    "plotCanvas",
    demoID,
    size.width,
    size.height,
    backendID,
  );
  if (result.error) {
    updateStatus(result.error);
    return;
  }
  const elapsedMs = performance.now() - startedAt;

  currentDemoID = result.id;
  currentBackendID = result.backendID || backendID;
  document.getElementById("demoTitle").textContent = result.title;
  document.getElementById("demoDescription").textContent = result.description;
  setDownloadEnabled(true);
  setSourceEnabled(true);
  setNavToolbarEnabled(true);
  syncNavToolbar(result.mode || api.navigationMode());
  if (!document.getElementById("sourcePanel").hidden) {
    showSourceForDemo(currentDemoID, false);
  }
  updateStatus(
    `Rendered ${result.id} with ${backendName} in ${elapsedMs.toFixed(1)} ms at ${result.width}×${result.height}`,
  );
}

function triggerNavAction(action) {
  if (runtimeExited || !api) {
    return;
  }
  const result = api.triggerToolbar(action);
  if (result && result.error) {
    updateStatus(result.error);
    return;
  }
  syncNavToolbar(result && result.mode);
}

function toggleNavMode(mode) {
  if (runtimeExited || !api) {
    return;
  }
  const current = api.navigationMode();
  const next = current === mode ? "none" : mode;
  const result = api.setNavigationMode(next);
  if (result && result.error) {
    updateStatus(result.error);
    return;
  }
  syncNavToolbar(result && result.mode);
}

function syncNavToolbar(mode) {
  const active = typeof mode === "string" ? mode : "none";
  const panBtn = document.getElementById("navPanBtn");
  const zoomBtn = document.getElementById("navZoomBtn");
  if (panBtn) {
    panBtn.setAttribute("aria-pressed", String(active === "pan"));
  }
  if (zoomBtn) {
    zoomBtn.setAttribute("aria-pressed", String(active === "zoom"));
  }
}

function setNavToolbarEnabled(enabled) {
  for (const id of ["navHomeBtn", "navPanBtn", "navZoomBtn"]) {
    const btn = document.getElementById(id);
    if (btn) {
      btn.disabled = !enabled;
    }
  }
}

function scheduleRender() {
  if (renderFrame !== 0) {
    return;
  }

  renderFrame = window.requestAnimationFrame(() => {
    renderFrame = 0;
    mountSelectedDemo();
  });
}

function plotSize() {
  const canvas = document.getElementById("plotCanvas");
  const width = Math.max(1, Math.round(canvas.clientWidth || DEFAULT_WIDTH));
  const height = Math.max(1, Math.round(canvas.clientHeight || DEFAULT_HEIGHT));
  return { width, height };
}

function downloadCurrentPlot() {
  const canvas = document.getElementById("plotCanvas");
  if (!canvas || !currentDemoID || !currentBackendID) {
    return;
  }

  canvas.toBlob((blob) => {
    if (!blob) {
      updateStatus("Failed to prepare PNG download");
      return;
    }

    const link = document.createElement("a");
    const url = URL.createObjectURL(blob);
    link.href = url;
    link.download = `matplotlib-go-${currentDemoID}-${currentBackendID}.png`;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(url);
  }, "image/png");
}

function updateStatus(message) {
  document.getElementById("statusMsg").textContent = message;
}

function setDownloadEnabled(enabled) {
  document.getElementById("downloadBtn").disabled = !enabled;
}

function setSourceEnabled(enabled) {
  document.getElementById("sourceBtn").disabled = !enabled;
}

function toggleSourcePanel() {
  const panel = document.getElementById("sourcePanel");
  if (panel.hidden) {
    showSourceForDemo(currentDemoID, true);
    return;
  }
  hideSourcePanel();
}

function showSourceForDemo(demoID, scrollIntoView) {
  if (!demoID) {
    return;
  }

  const source = loadDemoSource(demoID);
  if (!source) {
    return;
  }

  document.getElementById("sourceTitle").textContent =
    `Code for ${source.title}`;
  document.getElementById("sourceCode").innerHTML = highlightGoSource(
    source.source,
  );

  const panel = document.getElementById("sourcePanel");
  panel.hidden = false;

  const button = document.getElementById("sourceBtn");
  button.textContent = "Hide code";
  button.setAttribute("aria-expanded", "true");

  if (scrollIntoView) {
    panel.scrollIntoView({ behavior: "smooth", block: "start" });
  }
}

function hideSourcePanel() {
  document.getElementById("sourcePanel").hidden = true;

  const button = document.getElementById("sourceBtn");
  button.textContent = "View code";
  button.setAttribute("aria-expanded", "false");
}

function loadDemoSource(demoID) {
  if (sourceCache.has(demoID)) {
    return sourceCache.get(demoID);
  }

  const result = api.demoSource(demoID);
  if (result.error) {
    updateStatus(result.error);
    return null;
  }

  sourceCache.set(demoID, result);
  return result;
}

function highlightGoSource(source) {
  let html = "";
  let index = 0;
  let expectFunctionName = false;

  while (index < source.length) {
    const char = source[index];
    const next = source[index + 1];

    if (char === "/" && next === "/") {
      const end = source.indexOf("\n", index);
      const tokenEnd = end === -1 ? source.length : end;
      html += tokenSpan("comment", source.slice(index, tokenEnd));
      index = tokenEnd;
      continue;
    }

    if (char === "/" && next === "*") {
      const end = source.indexOf("*/", index + 2);
      const tokenEnd = end === -1 ? source.length : end + 2;
      html += tokenSpan("comment", source.slice(index, tokenEnd));
      index = tokenEnd;
      continue;
    }

    if (char === "`" || char === '"' || char === "'") {
      const tokenEnd = readStringEnd(source, index, char);
      html += tokenSpan("string", source.slice(index, tokenEnd));
      index = tokenEnd;
      expectFunctionName = false;
      continue;
    }

    if (isDigit(char)) {
      const tokenEnd = readNumberEnd(source, index);
      html += tokenSpan("number", source.slice(index, tokenEnd));
      index = tokenEnd;
      expectFunctionName = false;
      continue;
    }

    if (isIdentifierStart(char)) {
      const tokenEnd = readIdentifierEnd(source, index);
      const token = source.slice(index, tokenEnd);
      if (GO_KEYWORDS.has(token)) {
        html += tokenSpan("keyword", token);
        expectFunctionName = token === "func";
      } else if (expectFunctionName && nextNonSpace(source, tokenEnd) === "(") {
        html += tokenSpan("function", token);
        expectFunctionName = false;
      } else if (GO_PREDECLARED.has(token)) {
        html += tokenSpan("builtin", token);
        expectFunctionName = false;
      } else {
        html += escapeHTML(token);
      }
      index = tokenEnd;
      continue;
    }

    html += escapeHTML(char);
    if (!isWhitespace(char)) {
      expectFunctionName = false;
    }
    index += 1;
  }

  return html;
}

function readStringEnd(source, start, quote) {
  let index = start + 1;
  while (index < source.length) {
    if (quote !== "`" && source[index] === "\\") {
      index += 2;
      continue;
    }
    if (source[index] === quote) {
      return index + 1;
    }
    index += 1;
  }
  return source.length;
}

function readNumberEnd(source, start) {
  let index = start + 1;
  while (index < source.length && /[0-9A-Fa-f_.xXoOpP]/.test(source[index])) {
    index += 1;
  }
  return index;
}

function readIdentifierEnd(source, start) {
  let index = start + 1;
  while (index < source.length && isIdentifierPart(source[index])) {
    index += 1;
  }
  return index;
}

function nextNonSpace(source, start) {
  let index = start;
  while (index < source.length && isWhitespace(source[index])) {
    index += 1;
  }
  return source[index];
}

function isIdentifierStart(char) {
  return /[A-Za-z_]/.test(char);
}

function isIdentifierPart(char) {
  return /[A-Za-z0-9_]/.test(char);
}

function isDigit(char) {
  return /[0-9]/.test(char);
}

function isWhitespace(char) {
  return /\s/.test(char);
}

function tokenSpan(type, value) {
  return `<span class="tok-${type}">${escapeHTML(value)}</span>`;
}

function escapeHTML(value) {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;");
}

init();
