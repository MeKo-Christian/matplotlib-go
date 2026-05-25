# Phase 9A Coverage Audit

Phase 9A is an audit phase. Its purpose is to make Matplotlib parity gaps
visible and actionable before v1.0, not to close every implementation and demo
gap inside the same phase.

The machine-readable source of truth lives in `internal/examplecatalog/`. This
document explains how to read those inventories and how the Phase 9A exit
criteria should be interpreted.

The exit-criteria interpretation is also represented as data in
`CoverageAuditExitCriteria`, so tests can distinguish audit-complete criteria
from criteria that intentionally expose follow-up implementation work.

## What Phase 9A Reveals

Phase 9A answers three questions for each fundamental Matplotlib area:

- Is there a direct Go implementation or idiomatic Go equivalent?
- Is the behavior covered by parity fixtures and Matplotlib reference sources?
- Is there a user-facing demo, and does it exercise meaningful feature breadth?

The audit is intentionally split into smaller inventories:

- `FeatureCoverageMatrix` maps Matplotlib modules and gallery families to Go
  implementation files, parity catalog cases, examples, browser demos, and
  breadth status.
- `FoundationAPIGapAudit` records lower-level API gaps, such as artist
  metadata, ticker/formatter breadth, transform/BBox details, collection
  semantics, patch style registries, text/font behavior, image classes,
  colorbar behavior, norms, pyplot wrappers, and backend lifecycle APIs.
- `DemoBreadthGaps` records public feature families where fixtures exist but
  the user-facing examples are too thin or missing.
- `BrowserDemoCoverageRows` reconciles CLI showcases and Python web reference
  modules with `internal/examplecatalog.WebDemos()`.
- `ReferenceConsistencyClassifications` explains public fixture-only cases
  that are intentionally backend-stress, mesh-shading, signal, or compound
  statistics fixtures rather than standalone demos.
- `CoverageAuditExitCriteria` records the current interpretation, evidence,
  and remaining follow-up for the roadmap exit criteria.

## Current Findings

The current feature matrix classifies 16 fundamental feature areas:

- 5 have implemented Go equivalents for the main tracked behavior.
- 11 are partial: useful Go behavior exists, but important Matplotlib breadth
  remains missing or thinner than upstream.
- 1 area, widgets/events/animation, still has pending visual parity fixture
  coverage.
- Demo breadth is broad for 2 areas, thin for 11 areas, fixture-only for 2
  areas, and pending for 1 area.

The foundation API audit records 17 specific gaps:

- 10 are marked `implement`.
- 7 are marked `idiomatic-equivalent`.
- 0 are currently marked `intentional-omission`.

The demo-breadth audit records 17 demo gaps:

- 13 are high priority.
- 4 are medium priority.
- The highest-risk missing galleries are marker/line grids, advanced scatter,
  bar/fill/histogram variants, colormap and image variants, colorbar norm and
  extension variants, MathText, ticks/scales/formatters, text layout,
  annotation/legend/offset boxes, and mplot3d.

The browser-demo audit records 31 reconciliation rows:

- 30 are planned browser-demo promotions or wiring tasks.
- 1 is reference-only: `radialforce`, which is not yet a catalog case.

The reference-consistency audit now enforces that every catalog case has:

- `test/parity/<id>/plot.go`
- `test/parity/<id>/plot.py`
- `test/matplotlib_ref/plots/<id>.py`

It also requires public fixture-only cases to be accounted for by either a
demo-breadth gap or an explicit reference-consistency classification.

## Exit Criteria Interpretation

The Phase 9A exit criteria should be read as audit criteria, not as a promise
that every discovered gap has already been implemented.

`CoverageAuditExitCriteria` is the machine-readable counterpart to this table.

| Criterion                                                                                                                                                      | Current Interpretation                                                                                                                                                       | Status                                                                                  |
| -------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| Every fundamental Matplotlib feature area is classified as implemented, partially implemented, intentionally omitted, or pending.                              | `FeatureCoverageMatrix` gives each tracked area explicit coverage status values.                                                                                             | Satisfied for the current tracked areas.                                                |
| Every implemented public feature has at least one parity fixture or a documented reason why visual parity testing is not applicable.                           | Catalog and reference-consistency tests enforce parity source/reference coverage for cataloged cases, and implemented matrix rows must have catalog IDs or an omission note. | Satisfied for cataloged/implemented rows.                                               |
| Every major user-facing feature family has a showcase demo, and broad features have demos that exercise meaningful variants rather than only the minimal path. | Phase 9A currently identifies gaps here through `DemoBreadthGaps`; it does not close all of them.                                                                            | Not implementation-complete; audit-complete if the gaps are accepted as follow-up work. |
| Browser demo coverage is aligned with the catalog instead of being a separate, drifting subset.                                                                | `BrowserDemoCoverageRows` accounts for Python web reference modules and CLI-only showcases, but most rows are still planned rather than active.                              | Inventory aligned; implementation still pending for planned rows.                       |

## Practical Outcome

Phase 9A should produce a defensible parity roadmap:

- The project can say which Matplotlib areas have direct Go equivalents.
- Missing fundamentals are captured as named implementation decisions.
- Thin or missing examples are visible as prioritized demo gaps.
- Browser-demo drift is visible and catalog-linked.
- Every catalog case has Go and Python parity sources plus a visible Matplotlib
  reference module.

The remaining `partial`, `pending`, `fixture-only`, and `planned` rows are the
inputs for follow-up implementation phases. Closing those rows should happen as
targeted feature/demo work, not by editing the audit to hide them.
