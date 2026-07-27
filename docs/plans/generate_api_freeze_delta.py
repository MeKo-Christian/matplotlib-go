#!/usr/bin/env python3
"""Reconcile the live public-API freeze against the Phase 2 tiering decisions.

The tiering artifact classifies every symbol in the pre-break baseline as keep,
demote, or delete. The live freeze is what Phase 4 tags. This script diffs the
two at (package, symbol) granularity and requires that every difference is
accounted for by an explicit decision:

* every ``delete`` row is absent from the freeze;
* the ``demote`` row appears only in its target package;
* every ``keep`` row that left its baseline package either moved with the Phase
  2.2 package split or is named in ``REMOVED`` with a replacement;
* every name the freeze gained is named in ``ADDED`` with a category.

Anything unaccounted for is a hard failure: the point of the artifact is that
the delta can never grow silently.
"""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TIERING = ROOT / "docs/plans/api-tiering.json"
FREEZE = ROOT / "test/testdata/public_api/stable_public_api.json"
OUTPUT = ROOT / "docs/plans/api-freeze-delta.json"

# Packages carved out of `core` by Phase 2.2, plus the new `optional` package.
# A baseline `core` symbol that reappears unchanged in one of these is an
# accepted move and needs no hand-written row.
SPLIT_PACKAGES = ("plot3d", "ticker", "dates", "widgets", "optional")

_GETTER = "Dropped the Python-style Get prefix for the idiomatic Go noun."
_SPLIT = "Moved to the package that owns the feature after the Phase 2.2 split."

# (package, id) -> (category, replacement prose, target (package, id) or None).
#
# Rows for symbols the tiering artifact classified `keep` that nevertheless left
# their baseline package. Each names where the capability went. A target that is
# None means the replacement already existed in the baseline.
REMOVED: dict[tuple[str, str], tuple[str, str, tuple[str, str] | None]] = {
    ("backends", "func GetBestBackend"): (
        "renamed",
        _GETTER,
        ("backends", "func BestBackend"),
    ),
    ("backends", "func GetRecommendedBackend"): (
        "renamed",
        _GETTER,
        ("backends", "func RecommendedBackend"),
    ),
    ("backends/agg", "method Renderer.GetImage"): (
        "renamed",
        "Dropped the Get prefix. The draw method that previously held the "
        "Renderer.Image name became Renderer.DrawImage, so this symbol id "
        "survives the freeze with a different meaning.",
        None,
    ),
    ("backends/agg", "method Renderer.GetImageNRGBA"): (
        "renamed",
        _GETTER,
        ("backends/agg", "method Renderer.ImageNRGBA"),
    ),
    ("backends/gobasic", "method Renderer.GetImage"): (
        "renamed",
        "Dropped the Get prefix; the former Renderer.Image draw method became "
        "Renderer.DrawImage.",
        None,
    ),
    ("color", "func GetColormap"): (
        "renamed",
        "Renamed to the comma-ok lookup spelling used across the registries.",
        ("color", "func LookupColormap"),
    ),
    ("color", "func GetColormapStrict"): (
        "renamed",
        "Renamed to the comma-ok lookup spelling used across the registries.",
        ("color", "func LookupColormapStrict"),
    ),
    ("core", "func GetEpoch"): (
        "moved-renamed",
        "Moved with the date tick API and lost the Get prefix.",
        ("dates", "func Epoch"),
    ),
    ("core", "func NewAxes3D"): (
        "moved-renamed",
        "3D axes construction moved to plot3d.",
        ("plot3d", "func NewAxes"),
    ),
    ("core", "method Figure.AddAxes3D"): (
        "moved-renamed",
        "3D axes construction moved to plot3d and became a free function so "
        "core keeps no 3D dependency.",
        ("plot3d", "func AddAxes"),
    ),
    ("core", "method Axes3D.GetZLabel"): (
        "moved-renamed",
        "Moved with Axes3D and lost the Get prefix.",
        ("plot3d", "method Axes3D.ZLabel"),
    ),
    ("core", "method Collection.GetArray"): (
        "renamed",
        _GETTER,
        ("core", "method Collection.Array"),
    ),
    ("core", "method Scatter2D.GetArray"): (
        "renamed",
        _GETTER,
        ("core", "method Scatter2D.Array"),
    ),
    ("core", "method Line2D.SetMarkEvery"): (
        "renamed",
        "Renamed to match the MarkEverySpec field it actually writes; it never "
        "touched MarkEvery.",
        ("core", "method Line2D.SetMarkEverySpec"),
    ),
    ("core", "method Figure.SetSuptitle"): (
        "folded",
        "Duplicate spelling removed; the Go-cased Figure.SetSupTitle remains.",
        None,
    ),
    ("core", "method Figure.SetSupxlabel"): (
        "folded",
        "Duplicate spelling removed; the Go-cased Figure.SetSupXLabel remains.",
        None,
    ),
    ("core", "method Figure.SetSupylabel"): (
        "folded",
        "Duplicate spelling removed; the Go-cased Figure.SetSupYLabel remains.",
        None,
    ),
    ("core", "method Legend.SetLocator"): (
        "folded",
        "Pure alias for the exported Legend.Locator field. Assign the field.",
        None,
    ),
    ("core", "method AnchoredTextBox.SetLocator"): (
        "folded",
        "Pure alias for the exported AnchoredTextBox.Locator field. Assign the "
        "field.",
        None,
    ),
    ("core", "method Axis.SetTickDirection"): (
        "renamed",
        "Raw-string setter replaced by a parse function feeding the typed "
        "Axis.TickDirection field.",
        ("core", "func ParseTickDirection"),
    ),
    ("core", "method Axes.AddWidget"): (
        "folded",
        "Axes.Add now routes artists that implement the widget interface to the "
        "widget layer, so the separate entry point is redundant.",
        None,
    ),
    ("core", "type WidgetArtist"): (
        "unexported",
        "The widget marker interface became the unexported core.widgetArtist. "
        "Widgets satisfy it by implementing Draw, Z, Bounds, and WidgetLayer; "
        "core.Axes.Add dispatches on it.",
        None,
    ),
    ("core", "method Figure.AddWidgetAxes"): (
        "moved-renamed",
        _SPLIT,
        ("widgets", "func NewAxes"),
    ),
    ("core", "method SubFigure.AddWidgetAxes"): (
        "moved-renamed",
        _SPLIT,
        ("widgets", "func NewSubFigureAxes"),
    ),
    ("core", "method SubplotSpec.AddWidgetAxes"): (
        "moved-renamed",
        _SPLIT,
        ("widgets", "func NewSubplotAxes"),
    ),
    ("geom", "func GetCosSin"): ("renamed", _GETTER, ("geom", "func CosSin")),
    ("geom", "func GetIntersection"): (
        "renamed",
        _GETTER,
        ("geom", "func Intersection"),
    ),
    ("geom", "func GetNormalPoints"): (
        "renamed",
        _GETTER,
        ("geom", "func NormalPoints"),
    ),
    ("geom", "func GetParallels"): ("renamed", _GETTER, ("geom", "func Parallels")),
    ("pyplot", "func GetCMap"): ("renamed", _GETTER, ("pyplot", "func CMap")),
    ("pyplot", "func GetCurrentFigManager"): (
        "renamed",
        _GETTER,
        ("pyplot", "func CurrentFigManager"),
    ),
    ("style", "func GetTheme"): (
        "renamed",
        "Renamed to the comma-ok lookup spelling used across the registries.",
        ("style", "func LookupTheme"),
    ),
    ("tri", "method TriAnalyzer.GetFlatTriMask"): (
        "renamed",
        _GETTER,
        ("tri", "method TriAnalyzer.FlatTriMask"),
    ),
    ("render", "method NullRenderer.Image"): (
        "renamed",
        "The renderer draw method was renamed so raster exporters could expose "
        "their buffer as Image().",
        ("render", "method NullRenderer.DrawImage"),
    ),
}

# Every 3D widget-family method that moved from core.Axes to a widgets
# constructor. Same shape for all of them, so build the rows.
for _method, _ctor in (
    ("Button", "NewButton"),
    ("CheckButtons", "NewCheckButtons"),
    ("Cursor", "NewCursor"),
    ("EllipseSelector", "NewEllipseSelector"),
    ("LassoSelector", "NewLassoSelector"),
    ("MultiCursor", "NewMultiCursor"),
    ("MultiCursorWithOptions", "NewMultiCursorWithOptions"),
    ("PolygonSelector", "NewPolygonSelector"),
    ("RadioButtons", "NewRadioButtons"),
    ("RangeSlider", "NewRangeSlider"),
    ("RectangleSelector", "NewRectangleSelector"),
    ("Slider", "NewSlider"),
    ("SpanSelector", "NewSpanSelector"),
    ("TextBox", "NewTextBox"),
):
    REMOVED[("core", f"method Axes.{_method}")] = (
        "moved-renamed",
        _SPLIT,
        ("widgets", f"func {_ctor}"),
    )

# The vector backends only ever had the draw spelling, so their rename is a
# straight id swap.
for _backend in ("backends/pdf", "backends/pgf", "backends/ps", "backends/svg"):
    REMOVED[(_backend, "method Renderer.Image")] = (
        "renamed",
        "The renderer draw method was renamed to DrawImage.",
        (_backend, "method Renderer.DrawImage"),
    )

# (package, id) -> (category, note). Names the freeze gained that are not the
# target of a REMOVED row and did not move with the package split.
ADDED: dict[tuple[str, str], tuple[str, str]] = {}


def _add(category: str, note: str, package: str, *ids: str) -> None:
    for symbol in ids:
        ADDED[(package, symbol)] = (category, note)


_add(
    "typed-enum",
    "Typed replacement for a raw-string option enum. String literals still "
    "compile; the type and its constants are new names, and no field was "
    "removed, so the freeze count only grows.",
    "core",
    "type ColorbarExtend",
    "type ColorbarLocation",
    "type ImageAspect",
    "type PlotOrientation",
    "type VectorPivot",
    "type ViolinSide",
    "const AspectAuto",
    "const AspectEqual",
    "const ColorbarBottom",
    "const ColorbarLeft",
    "const ColorbarRight",
    "const ColorbarTop",
    "const ExtendBoth",
    "const ExtendMax",
    "const ExtendMin",
    "const ExtendNeither",
    "const OrientationHorizontal",
    "const OrientationVertical",
    "const PivotMiddle",
    "const PivotTail",
    "const PivotTip",
    "const ViolinSideBoth",
    "const ViolinSideHigh",
    "const ViolinSideLow",
)

_add(
    "options-model",
    "Folded value type or options struct from the Phase 2.3 options model. "
    "Each one replaces struct fields, which the audit does not count as "
    "symbols, so folding trades uncounted fields for counted names.",
    "core",
    "type DashPattern",
    "method DashPattern.Scaled",
    "func MatplotlibDashes",
    "func PixelDashes",
    "type LineCollectionOptions",
)

_add(
    "optional-package",
    "The optional package introduced by the Phase 2.3 options model. It "
    "replaced 408 pointer-to-primitive option fields, none of which the audit "
    "counted as symbols.",
    "optional",
    "type Value",
    "func Of",
    "func None",
    "func FromPtr",
)

_add(
    "figure-output",
    "The Figure.Save/WriteTo/Image output surface and its renderer registry.",
    "core",
    "method Figure.Save",
    "method Figure.WriteTo",
    "method Figure.Image",
    "func RegisterFigureOutputRenderer",
    "type FigureOutputRendererFactory",
)

_add(
    "encapsulation-reader",
    "Reader introduced when the matching exported field was unexported so its "
    "setter became the only writer. The field it replaces was never counted as "
    "a symbol, so this is a pure count increase for a net reduction in surface.",
    "core",
    "method Axes.Title",
    "method Axes.XLabel",
    "method Axes.YLabel",
    "method Figure.SupTitle",
    "method Figure.SupXLabel",
    "method Figure.SupYLabel",
)
_add(
    "encapsulation-reader",
    "Reader introduced when the matching exported field was unexported so its "
    "clamping, caret-repositioning, and on-change setter became the only writer.",
    "widgets",
    "method Slider.Value",
    "method TextBox.Value",
    "method RadioButtons.Active",
    "method RangeSlider.Low",
    "method RangeSlider.High",
)

_add(
    "split-export",
    "Formerly package-private helper that had to be exported so the package "
    "split could keep working. plot3d and widgets live outside core and need "
    "these to draw, order, and lay out artists the way core does.",
    "core",
    "func DrawAlignedText",
    "func DrawAlignedTextWithFont",
    "func FontKeyWithWeight",
    "method Axes.InvalidateArtistOrder",
    "method Axes.ResolvedRC",
    "method Axes.XLabelFontKey",
    "method Axes.XLabelPad",
    "method Axes.YLabelFontKey",
    "method Axes.YLabelPad",
    "method Collection.SetZ",
    "method Line2D.SetZ",
    "method Scatter2D.SetZ",
    "method Scatter2D.PrototypePath",
)
_add(
    "split-export",
    "Formerly package-private tick helper that had to be exported so core could "
    "keep calling it after locators and formatters moved to ticker.",
    "ticker",
    "func FormatTick",
    "func LabelFormatter",
    "func WithMajorContext",
    "method ScalarFormatter.FormatData",
    "method ScalarFormatter.FormatStep",
    "method ScalarFormatter.WithOffsetThreshold",
)

_add(
    "shared-helper",
    "Shared path extracted during Phase 2.3 so core and the split packages "
    "resolve alpha and scalar mapping identically.",
    "render",
    "method Color.WithAlphaMultiplier",
)
_add(
    "shared-helper",
    "Shared path extracted during Phase 2.3 so core and the split packages "
    "resolve alpha and scalar mapping identically.",
    "core",
    "method PlotOptions.ScalarMapConfig",
)

_add(
    "audit-scope",
    "Pre-existing API the baseline collector never parsed. The FFmpeg writer "
    "sits behind //go:build ffmpeg and predates Phase 2; the final collector "
    "reads every non-test source file regardless of build constraints.",
    "animation",
    "func FFmpegAvailable",
    "func NewFFmpegWriter",
    "func NewFFmpegWebMWriter",
    "type FFmpegWriter",
    "method FFmpegWriter.Setup",
    "method FFmpegWriter.GrabFrame",
    "method FFmpegWriter.Finish",
    "method FFmpegWriter.FrameSize",
)
_add(
    "audit-scope",
    "Native Skia surface parsed only by the final collector, which merges the "
    "unavailable-backend stub type with its skiacgo/skiagpu implementation. The "
    "GPU accessors are Phase 1 work, which closed alongside the baseline.",
    "backends/skia",
    "method Renderer.BridgeInfo",
    "method Renderer.ColorType",
    "method Renderer.DrawGouraudTriangles",
    "method Renderer.DrawMarkers",
    "method Renderer.DrawPathCollection",
    "method Renderer.DrawQuadMesh",
    "method Renderer.FlushGPU",
    "method Renderer.GPU",
    "method Renderer.GPUModeRequested",
    "method Renderer.ImageTransformed",
    "method Renderer.RendererModeLabel",
    "method Renderer.RuntimeCapabilityStatus",
    "method Renderer.SampleCount",
    "method Renderer.SetDefaultSketch",
    "method Renderer.SetResolution",
    "method Renderer.StartFilter",
    "method Renderer.StopFilter",
    "method Renderer.SupportsGradientFill",
    "method Renderer.SupportsNativeHatch",
    "method Renderer.SupportsPatternFill",
    "method Renderer.Surface",
)
_add(
    "audit-scope",
    "The demoted render.RendererModeReporter in its target package. The tiering "
    "artifact records the demotion; this is where it landed.",
    "backends",
    "type RendererModeReporter",
)
for _backend in ("backends/agg", "backends/gobasic", "backends/skia"):
    _add(
        "renamed-draw-method",
        "The Renderer.Image draw method was renamed to DrawImage. On the raster "
        "backends the vacated Image name was then taken by the buffer accessor, "
        "so the freeze keeps the id while its meaning changed.",
        _backend,
        "method Renderer.DrawImage",
    )

_add(
    "post-phase2",
    "Added by Phase 2 follow-up 1, which routed the 2D and 3D contour paths "
    "through one levels-autoscaling scalar-map resolver. plot3d cannot reach "
    "core's unexported helpers, so the shared pair had to be exported.",
    "core",
    "func ResolveContourScalarMap",
    "func ContourFillScalarMap",
)

_add(
    "post-phase2",
    "Added by Phase 3.3.4, which needed a Scale outside the transform package "
    "to declare that its forward map is affine so AsAffine can flatten the "
    "data->pixel graph into a single matrix. core.invertedScale is the first "
    "implementor; without it matshow's inverted y axis fell back to staged "
    "evaluation and placed one tick a pixel low. Purely additive -- the "
    "built-in Linear path is unchanged.",
    "transform",
    "type AffineScale",
    "func IsAffineScale",
)


def load(path: Path) -> tuple[bytes, dict]:
    raw = path.read_bytes()
    return raw, json.loads(raw)


def generate() -> bytes:
    tiering_bytes, tiering = load(TIERING)
    freeze_bytes, freeze = load(FREEZE)

    baseline = {
        (row["package"], row["id"]): row["disposition"] for row in tiering["symbols"]
    }
    live = {
        (package["dir"], symbol["id"])
        for package in freeze["packages"]
        for symbol in package["symbols"]
    }
    live_ids: dict[str, set[str]] = {}
    for package in freeze["packages"]:
        for symbol in package["symbols"]:
            live_ids.setdefault(symbol["id"], set()).add(package["dir"])

    rows: list[dict] = []
    problems: list[str] = []

    # Deletions and the demotion first: these are the decisions the tiering
    # artifact made explicitly, so they are checked, not classified.
    for key, disposition in sorted(baseline.items()):
        package, symbol = key
        if disposition == "delete":
            if key in live:
                problems.append(f"{package}.{symbol}: tiered delete is still frozen")
            rows.append(
                {
                    "direction": "removed",
                    "package": package,
                    "id": symbol,
                    "category": "tiered-delete",
                    "note": "Deleted by the Phase 2.1 tiering decision.",
                }
            )
            continue
        if disposition == "demote":
            if key in live:
                problems.append(
                    f"{package}.{symbol}: tiered demote still frozen in its source package"
                )
            rows.append(
                {
                    "direction": "removed",
                    "package": package,
                    "id": symbol,
                    "category": "tiered-demote",
                    "note": "Demoted to the consumer-owned package by the Phase 2.1 "
                    "tiering decision.",
                }
            )
            continue
        if key in live:
            continue

        # A `keep` row that left its baseline package.
        moved_to = sorted(live_ids.get(symbol, set()) & set(SPLIT_PACKAGES))
        if package == "core" and moved_to and key not in REMOVED:
            rows.append(
                {
                    "direction": "removed",
                    "package": package,
                    "id": symbol,
                    "category": "moved",
                    "note": _SPLIT,
                    "target_package": moved_to[0],
                    "target_id": symbol,
                }
            )
            continue
        decision = REMOVED.get(key)
        if decision is None:
            problems.append(
                f"{package}.{symbol}: kept by tiering but absent from the freeze "
                "with no REMOVED row"
            )
            continue
        category, note, target = decision
        row = {
            "direction": "removed",
            "package": package,
            "id": symbol,
            "category": category,
            "note": note,
        }
        if target is not None:
            if target not in live:
                problems.append(
                    f"{package}.{symbol}: replacement {target[0]}.{target[1]} is not frozen"
                )
            row["target_package"], row["target_id"] = target
        rows.append(row)

    rename_targets = {
        target for _, _, target in REMOVED.values() if target is not None
    }

    for key in sorted(live - set(baseline)):
        package, symbol = key
        if key in rename_targets:
            rows.append(
                {
                    "direction": "added",
                    "package": package,
                    "id": symbol,
                    "category": "rename-target",
                    "note": "Replacement for a renamed or relocated baseline symbol.",
                }
            )
            continue
        if package in SPLIT_PACKAGES and ("core", symbol) in baseline:
            rows.append(
                {
                    "direction": "added",
                    "package": package,
                    "id": symbol,
                    "category": "moved",
                    "note": _SPLIT,
                    "source_package": "core",
                    "source_id": symbol,
                }
            )
            continue
        decision = ADDED.get(key)
        if decision is None:
            problems.append(
                f"{package}.{symbol}: frozen but not classified; add an ADDED row"
            )
            continue
        category, note = decision
        rows.append(
            {
                "direction": "added",
                "package": package,
                "id": symbol,
                "category": category,
                "note": note,
            }
        )

    stale = sorted(set(ADDED) - live)
    for package, symbol in stale:
        problems.append(f"{package}.{symbol}: ADDED row no longer matches the freeze")
    for key in sorted(set(REMOVED) - set(baseline)):
        problems.append(f"{key[0]}.{key[1]}: REMOVED row is not a baseline symbol")

    if problems:
        raise SystemExit(
            "public API freeze delta is unreconciled:\n  " + "\n  ".join(problems)
        )

    counts: dict[str, dict[str, int]] = {}
    for row in rows:
        counts.setdefault(row["direction"], {})
        counts[row["direction"]][row["category"]] = (
            counts[row["direction"]].get(row["category"], 0) + 1
        )

    artifact = {
        "schema_version": 1,
        "generated_by": "docs/plans/generate_api_freeze_delta.py",
        "baseline": {
            "path": "docs/plans/api-tiering.json",
            "sha256": hashlib.sha256(tiering_bytes).hexdigest(),
            "symbol_count": len(baseline),
        },
        "freeze": {
            "path": "test/testdata/public_api/stable_public_api.json",
            "sha256": hashlib.sha256(freeze_bytes).hexdigest(),
            "symbol_count": len(live),
        },
        "counts": counts,
        "rows": rows,
    }
    return (json.dumps(artifact, indent=2, ensure_ascii=False) + "\n").encode()


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--check",
        action="store_true",
        help="fail instead of writing when the committed artifact is stale",
    )
    args = parser.parse_args()
    generated = generate()

    if args.check:
        if not OUTPUT.exists() or OUTPUT.read_bytes() != generated:
            print(f"{OUTPUT.relative_to(ROOT)} is stale")
            return 1
        return 0

    OUTPUT.write_bytes(generated)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
