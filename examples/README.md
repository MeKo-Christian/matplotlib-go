Examples live here. The curated gallery workflow is documented in
[`../docs/examples-gallery.md`](../docs/examples-gallery.md).

Each cataloged example is a library package at `examples/<id>/example.go`
exporting `func Render() image.Image`. The same `Render` is consumed by:

- The browser gallery (source display via `internal/webdemo`).
- Parity testing — `test/parity/<id>/plot.go` is a thin wrapper that imports
  the showcase package, so golden tests share a single rendering body.
- The unified CLI runner, `go run ./cmd/example -name <id> -o <file>`.

Parity test fixtures (one folder per case, Go + Python) live in
`test/parity/<id>/`. The matching Matplotlib reference Python plots live in
`test/matplotlib_ref/plots/<id>.py`. Use `go run ./test/parity/cmd --list`
to inspect canonical case IDs, or `python3 test/parity/run.py --list` for
the Python side.

Renderer/backend stress fixtures (`FixtureOnly: true` in
`internal/examplecatalog`) are deliberately _not_ surfaced as user-facing
showcase examples — their bodies live only under `test/parity/<id>/plot.go`.

A few legacy topic-named directories (`examples/lines/`, `examples/scatter/`,
etc.) still exist as `package main` runners and as host packages for
delegating wrappers (e.g. `examples/arrays_showcase` imports
`examples/arrays/showcase`). Inlining and retiring those is tracked in
`PLAN.md`.
