# Phase 2 Public API Tiering

This document records the pre-v1 decision for every declaration in the
preserved `phase2-prebreak-public-api.json` snapshot. The exhaustive,
reviewable rows live in `phase2-public-api-tiering.json`; the adjacent
generator applies the decisions below to that immutable baseline
deterministically. The live API freeze remains
`test/testdata/public_api/stable_public_api.json` and may therefore be
regenerated without rewriting this historical decision record.

The baseline contains 3,102 declarations across 24 audited packages: 3,082 are
kept, 19 are deleted, and one is demoted. `keep` means the public capability is
retained, although Phase 2 may move its import path. `demote` means the
capability remains available through a narrower consumer-owned API. `delete`
means the old declaration has no compatibility promise. Deleted and demoted
rows always name a replacement and rationale.

## Decisions

- Delete the Python-style introspection surface: `Setp`, `Getp`, `GetpAll`,
  `Findobj`, `FindobjType`, the `Axes.Findobj` and `Figure.Findobj` methods,
  `PropertyBag`, and its six implementations on `ArtistRasterization` and
  `Line2D`. These APIs use strings and `any`, cover only a small subset of
  artists, and can warn-and-ignore invalid mutations. Callers should use typed
  artist methods and explicit ownership.
- Delete `Axes.PlotUnits`, `ScatterUnits`, `BarUnits`, and `FillBetweenUnits`
  after folding unit conversion into the corresponding primary methods. The
  primary methods will return `(T, error)` for rejected input. Conversion and
  validation must be transactional: a rejected call must not add an artist,
  advance a property cycle, or leave partially configured axis units.
- Delete `render.CapabilityBridgeReporter`. Typed
  `backends.CapabilityStatus` reporting supersedes its string-based protocol.
- Demote `render.RendererModeReporter` to the consumer-owned `backends`
  package. Mode labels describe registry reports, not drawing behavior.
- Keep every other renderer interface. The repository documents third-party
  backend registration, and these named optional contracts make advertised
  `backends.Capability` values verifiable.
- Keep all remaining declarations. This is intentionally conservative:
  package moves and later convention work may change their package or
  signature, but those capabilities are not deleted by the surface-tiering
  decision.

## Updating and validation

Run:

```bash
python3 docs/plans/generate_phase2_public_api_tiering.py
go test -run TestPublicAPITieringMatchesPreBreakSnapshot .
```

The root test verifies the frozen file hash and count, exact duplicate-free
symbol coverage, the allowed dispositions, replacement/rationale rules, the
landmark decisions above, and retention of the remaining renderer interfaces.
The artifact preserves the Phase 2.1 baseline; later API regeneration should
update this design record deliberately rather than silently accepting new
symbols. `phase2-freeze-delta.md` is where that reconciliation lives: it walks
every difference between these decisions and the live freeze, and
`TestPublicAPIFreezeDeltaIsReconciled` keeps the walk current.

The Phase 2 API collector parses every non-test Go source file in each audited
package, independent of build constraints. It deduplicates equivalent function
signatures (ignoring parameter names), merges empty unavailable-backend stub
types with their implementation type, and rejects other conflicting
build-variant declarations. The explicit package list in
`apidoc_coverage_test.go` now includes `dates`, `plot3d`, `ticker`, and
`widgets`, so the final freeze covers both the new package boundaries and
build-tagged backend surfaces.
