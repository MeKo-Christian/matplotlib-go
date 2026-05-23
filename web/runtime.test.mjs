import test from "node:test";
import assert from "node:assert/strict";

import {
  REQUIRED_API_METHODS,
  buildAPICompatibilityMessage,
  buildRuntimeExitMessage,
  missingAPIMethods,
} from "./runtime.mjs";

test("missingAPIMethods reports the required API surface in order", () => {
  const partialAPI = {
    listDemos() {},
    defaultDemoID() {},
  };

  assert.deepEqual(missingAPIMethods(partialAPI), [
    "listBackends",
    "mountDemo",
    "resizeDemo",
    "unmountDemo",
    "defaultBackendID",
    "renderDemoPNG",
    "demoSource",
    "setNavigationMode",
    "navigationMode",
    "triggerToolbar",
  ]);
  assert.deepEqual(missingAPIMethods(Object.fromEntries(REQUIRED_API_METHODS.map((name) => [name, () => {}]))), []);
});

test("buildAPICompatibilityMessage points developers to a rebuild", () => {
  const message = buildAPICompatibilityMessage(["mountDemo", "resizeDemo"]);

  assert.match(message, /Incompatible web build detected/);
  assert.match(message, /just web-build/);
  assert.match(message, /mountDemo, resizeDemo/);
});

test("buildRuntimeExitMessage includes optional error detail", () => {
  assert.match(buildRuntimeExitMessage(), /WASM runtime exited unexpectedly/);
  assert.match(
    buildRuntimeExitMessage(new Error("panic in callback")),
    /Details: panic in callback/,
  );
});
