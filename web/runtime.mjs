export const REQUIRED_API_METHODS = Object.freeze([
  "listDemos",
  "listBackends",
  "mountDemo",
  "resizeDemo",
  "unmountDemo",
  "defaultDemoID",
  "defaultBackendID",
  "renderDemoPNG",
  "demoSource",
  "setNavigationMode",
  "navigationMode",
  "triggerToolbar",
]);

export function missingAPIMethods(api) {
  return REQUIRED_API_METHODS.filter((name) => typeof api?.[name] !== "function");
}

export function buildAPICompatibilityMessage(missingMethods) {
  const missing = Array.isArray(missingMethods) ? missingMethods : [];
  const details =
    missing.length > 0 ? ` Missing methods: ${missing.join(", ")}.` : "";
  return `Incompatible web build detected; rebuild with just web-build and refresh the page.${details}`;
}

export function buildRuntimeExitMessage(error) {
  const detail =
    error instanceof Error && error.message
      ? ` Details: ${error.message}.`
      : "";
  return `WASM runtime exited unexpectedly. Reload the page or rebuild with just web-build if assets are stale.${detail}`;
}
