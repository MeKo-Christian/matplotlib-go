#!/usr/bin/env python3
"""Generate the Phase 3 tolerance audit from the committed parity fixtures.

The audit answers three questions for every ``internal/examplecatalog`` case:

1. How far is ``testdata/golden/{id}.png`` from ``testdata/matplotlib_ref/{id}.png``
   under the harness metric (``test/imagecmp.ComparePNG``)?
2. What *shape* does the residual have -- a handful of shifted glyph clusters, a
   broad low-amplitude wash, or nothing at all? Whole-image RMSE cannot tell
   these apart, and only the first kind is a fixable core-library divergence.
3. How much slack does the case's catalog tolerance have over its measured
   value, i.e. how large a regression could land without failing the gate?

The script is deliberately standalone: it reads only committed PNGs and the
catalog source, needs no cgo, and reproduces ``imagecmp``'s RMSE/MeanAbs exactly
(RMSE = sqrt(mean of squared per-channel differences over RGBA)).

Usage:

    python3 docs/plans/generate_phase3_tolerance_audit.py

Writes ``docs/plans/phase3-tolerance-audit.md`` and
``docs/plans/phase3-tolerance-audit.json``.
"""

from __future__ import annotations

import json
import os
import re
import sys
from dataclasses import asdict, dataclass, field

import numpy as np
from PIL import Image
from scipy import ndimage

REPO = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
CATALOG = os.path.join(REPO, "internal", "examplecatalog", "catalog.go")
GOLDEN = os.path.join(REPO, "testdata", "golden")
MPLREF = os.path.join(REPO, "testdata", "matplotlib_ref")
OUT_MD = os.path.join(REPO, "docs", "plans", "phase3-tolerance-audit.md")
OUT_JSON = os.path.join(REPO, "docs", "plans", "phase3-tolerance-audit.json")

# Harness defaults from test/helpers_test.go.
DEFAULT_MAX_MEAN_ABS = 2.50
DIFF_TOLERANCE = 1  # per-channel LSBs ignored, matching ComparePNG

# A case is "loose" when a regression this many times its measured residual
# would still pass, and "fragile" when almost any change trips the gate.
LOOSE_SLACK = 3.0
FRAGILE_SLACK = 1.15

# Amplitude at which a broad residual stops being antialiasing-scale noise.
LOW_AMPLITUDE_P99 = 32
# Fraction of the frame above which a residual is "dense" rather than local.
DENSE_PCT = 1.0
# Share of clustered pixels that must agree on one offset to call it a shift.
SHIFT_DOMINANCE = 0.8
MAX_CLUSTERS_ANALYZED = 12


@dataclass
class Row:
    id: str
    max_mean_abs: float
    max_rmse: float
    rmse: float = 0.0
    mean_abs: float = 0.0
    psnr: float = 0.0
    max_diff: int = 0
    diff_px: int = 0
    pct_px: float = 0.0
    clusters: int = 0
    largest_cluster: int = 0
    p50_amp: int = 0
    p99_amp: int = 0
    shifts: dict[str, int] = field(default_factory=dict)
    family: str = ""
    slack: float | None = None
    verdict: str = ""


def parse_catalog() -> list[Row]:
    """Read the one-line ``examplecatalog.Case`` literals and their tolerances."""
    src = open(CATALOG, encoding="utf-8").read()
    rows: list[Row] = []
    for id_, rest in re.findall(r'^\t\{ID: "([^"]+)",(.*)\},\s*$', src, re.M):

        def num(name: str) -> float:
            m = re.search(name + r":\s*([0-9.]+)", rest)
            return float(m.group(1)) if m else 0.0

        rows.append(
            Row(
                id=id_,
                max_mean_abs=num("MaxMeanAbs"),
                max_rmse=num("MaxRMSE"),
            )
        )
    return rows


def load(path: str) -> np.ndarray:
    return np.asarray(Image.open(path).convert("RGBA")).astype(np.int16)


def dominant_shifts(golden: np.ndarray, ref: np.ndarray, mask: np.ndarray) -> dict[str, int]:
    """For the largest diff clusters, find the integer offset that best explains
    each one. A cluster that a +0+1 shift cancels is a placement bug, not an
    amplitude difference."""
    labels, count = ndimage.label(mask, structure=np.ones((3, 3)))
    if count == 0:
        return {}
    sizes = ndimage.sum(mask, labels, range(1, count + 1))
    boxes = ndimage.find_objects(labels)
    a = golden[:, :, :3].astype(np.float64)
    b = ref[:, :, :3].astype(np.float64)
    height, width = mask.shape

    shifts: dict[str, int] = {}
    for k in np.argsort(-sizes)[:MAX_CLUSTERS_ANALYZED]:
        ys, xs = boxes[k]
        y0, y1 = max(0, ys.start - 3), min(height, ys.stop + 3)
        x0, x1 = max(0, xs.start - 3), min(width, xs.stop + 3)
        best = (0, 0, np.abs(a[y0:y1, x0:x1] - b[y0:y1, x0:x1]).mean())
        for dy in (-2, -1, 0, 1, 2):
            for dx in (-2, -1, 0, 1, 2):
                if y0 + dy < 0 or x0 + dx < 0 or y1 + dy > height or x1 + dx > width:
                    continue
                residual = np.abs(a[y0 + dy : y1 + dy, x0 + dx : x1 + dx] - b[y0:y1, x0:x1]).mean()
                if residual < best[2] - 1e-9:
                    best = (dx, dy, residual)
        key = "%+d%+d" % (best[0], best[1])
        shifts[key] = shifts.get(key, 0) + int(sizes[k])
    return shifts


def classify(row: Row) -> str:
    if row.diff_px == 0:
        return "identical"
    clustered = sum(row.shifts.values())
    if clustered:
        for key, weight in row.shifts.items():
            if key != "+0+0" and weight >= SHIFT_DOMINANCE * clustered:
                return "shift " + key
    if row.pct_px < DENSE_PCT:
        return "sparse-local"
    if row.p99_amp <= LOW_AMPLITUDE_P99:
        return "dense-lowamp"
    return "dense-highamp"


def measure(row: Row) -> None:
    golden = load(os.path.join(GOLDEN, row.id + ".png"))
    ref = load(os.path.join(MPLREF, row.id + ".png"))
    if golden.shape != ref.shape:
        raise SystemExit(f"{row.id}: golden {golden.shape} vs reference {ref.shape}")

    diff = np.abs(golden - ref)
    per_channel = diff.astype(np.float64)
    row.rmse = round(float(np.sqrt((per_channel**2).mean())), 3)
    row.mean_abs = round(float(per_channel.mean()), 3)
    # Matches imagecmp since Phase 3.1: PSNR is a restatement of RMSE.
    row.psnr = round(float(20 * np.log10(255 / row.rmse)), 2) if row.rmse > 0 else float("inf")

    channel_max = diff.max(axis=2)
    row.max_diff = int(channel_max.max())
    mask = channel_max > DIFF_TOLERANCE
    row.diff_px = int(mask.sum())
    row.pct_px = round(100.0 * row.diff_px / mask.size, 4)

    amplitudes = channel_max[mask]
    if amplitudes.size:
        row.p50_amp = int(np.median(amplitudes))
        row.p99_amp = int(np.percentile(amplitudes, 99))

    labels, count = ndimage.label(mask, structure=np.ones((3, 3)))
    row.clusters = int(count)
    if count:
        sizes = ndimage.sum(mask, labels, range(1, count + 1))
        row.largest_cluster = int(sizes.max())
    row.shifts = dominant_shifts(golden, ref, mask)
    row.family = classify(row)

    if row.max_rmse > 0 and row.rmse > 0:
        row.slack = round(row.max_rmse / row.rmse, 2)
    if row.rmse == 0:
        # Slack is undefined against a zero residual, and no ratchet applies.
        row.verdict = "exact"
    elif row.slack is None:
        row.verdict = "no-rmse-gate"
    elif row.slack >= LOOSE_SLACK:
        row.verdict = "loose"
    elif row.slack < FRAGILE_SLACK:
        row.verdict = "fragile"
    else:
        row.verdict = "ok"


def render_markdown(rows: list[Row]) -> str:
    total = len(rows)
    identical = [r for r in rows if r.family == "identical"]
    shifted = [r for r in rows if r.family.startswith("shift ")]
    sparse = [r for r in rows if r.family == "sparse-local"]
    dense_low = [r for r in rows if r.family == "dense-lowamp"]
    dense_high = [r for r in rows if r.family == "dense-highamp"]
    loose = [r for r in rows if r.verdict == "loose"]
    fragile = [r for r in rows if r.verdict == "fragile"]

    out: list[str] = []
    w = out.append
    w("# Phase 3 tolerance audit\n")
    w(
        "Generated by `docs/plans/generate_phase3_tolerance_audit.py` from the "
        "committed fixtures. Every number below is `testdata/golden/{id}.png` "
        "against `testdata/matplotlib_ref/{id}.png` under the harness metric.\n"
    )
    w("## Summary\n")
    w(f"- Cases audited: **{total}** (every catalog row has both a golden and a reference).")
    w(f"- Pixel-identical (within {DIFF_TOLERANCE} LSB): **{len(identical)}**.")
    w(
        f"- Residual is a dominant integer pixel offset: **{len(shifted)}** "
        "— placement bugs, not amplitude noise."
    )
    w(f"- Sparse local residual (<{DENSE_PCT:g}% of frame): **{len(sparse)}**.")
    w(f"- Dense low-amplitude wash (p99 <= {LOW_AMPLITUDE_P99}): **{len(dense_low)}**.")
    w(f"- Dense high-amplitude residual: **{len(dense_high)}**.")
    w(f"- `MaxRMSE` at least {LOOSE_SLACK:g}x the measured value: **{len(loose)}**.")
    w(f"- `MaxRMSE` under {FRAGILE_SLACK:g}x the measured value: **{len(fragile)}**.")
    w(
        "- PSNR is no longer a gate anywhere. Phase 3.1 fixed the `uint8` "
        "overflow in `imagecmp`'s PSNR accumulator, which made `PSNR == "
        "20*log10(255/RMSE)` exact and every `MinPSNR` floor a restatement of "
        "`MaxRMSE`. `Case.MinPSNR` is gone; the 60 floors that were both binding "
        "and reachable were folded into `MaxRMSE`, and the 65 that the overflow "
        "had made unreachable were dropped for 3.6 to re-derive.\n"
    )
    w("## Why RMSE alone is not a gate\n")
    w(
        "`basic_line` is the smallest plot in the catalog, and it is why this "
        "table exists. Until Phase 3.3 its residual was 139 pixels out of "
        "230,400 — a single y tick label rendered one pixel low, with glyph "
        "rasterization otherwise byte-identical. Those 139 pixels reached a "
        "per-channel difference of 249 and still produced RMSE 2.49, "
        "comfortably inside the case's 2.8 allowance. Whole-image RMSE dilutes "
        "a fully-wrong glyph by frame area, so it cannot bound a localized "
        "placement error; the case now measures zero differing pixels, and "
        "only the shape gate would have caught the regression that hid there. "
        "The `family` column below exists to separate such cases from genuine "
        "antialiasing washes.\n"
    )
    w("## Shift families\n")
    w(
        "Cases whose largest residual clusters are cancelled by one integer "
        "offset. Each offset is a candidate shared root cause.\n"
    )
    by_shift: dict[str, list[Row]] = {}
    for r in shifted:
        by_shift.setdefault(r.family, []).append(r)
    for key in sorted(by_shift, key=lambda k: -len(by_shift[k])):
        ids = ", ".join(f"`{r.id}`" for r in sorted(by_shift[key], key=lambda r: -r.rmse))
        w(f"- **{key}** ({len(by_shift[key])}): {ids}")
    w("")
    w("## Per-case table\n")
    w(
        "`slack` is `MaxRMSE / RMSE`: how many times the measured residual a "
        "regression could grow before the gate fires.\n"
    )
    w(
        "| case | RMSE | MaxRMSE | slack | verdict | diff px | % frame | clusters | "
        "p50 amp | p99 amp | max diff | family |"
    )
    w("| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |")
    for r in sorted(rows, key=lambda r: (-r.rmse, r.id)):
        slack = "-" if r.slack is None else f"{r.slack:.2f}x"
        w(
            f"| `{r.id}` | {r.rmse:.2f} | {r.max_rmse:.2f} | {slack} | {r.verdict} | "
            f"{r.diff_px} | {r.pct_px:.3f} | {r.clusters} | {r.p50_amp} | "
            f"{r.p99_amp} | {r.max_diff} | {r.family} |"
        )
    w("")
    return "\n".join(out)


def main() -> int:
    rows = parse_catalog()
    for row in rows:
        measure(row)
    with open(OUT_JSON, "w", encoding="utf-8") as fh:
        json.dump([asdict(r) for r in rows], fh, indent=1, sort_keys=True)
        fh.write("\n")
    with open(OUT_MD, "w", encoding="utf-8") as fh:
        fh.write(render_markdown(rows))
    print(f"wrote {OUT_MD} and {OUT_JSON} ({len(rows)} cases)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
