import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";

test("landing page links to curated documentation entry points", async () => {
  const html = await readFile(new URL("./index.html", import.meta.url), "utf8");

  assert.match(html, /class="docs-links"/);
  for (const target of [
    "https://pkg.go.dev/github.com/cwbudde/matplotlib-go",
    "https://github.com/cwbudde/matplotlib-go#readme",
    "https://github.com/cwbudde/matplotlib-go/blob/main/docs/examples-gallery.md",
    "https://github.com/cwbudde/matplotlib-go/blob/main/docs/backend-selection.md",
    "https://github.com/cwbudde/matplotlib-go/blob/main/docs/matplotlib-migration-notes.md",
    "https://github.com/cwbudde/matplotlib-go/blob/main/docs/matplotlib-parity-status.md",
  ]) {
    assert.match(html, new RegExp(escapeRegExp(target)));
  }
});

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
