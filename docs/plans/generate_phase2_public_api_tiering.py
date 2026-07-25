#!/usr/bin/env python3
"""Generate the exhaustive Phase 2 public-API tiering artifact."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
FROZEN_API = ROOT / "docs/plans/phase2-prebreak-public-api.json"
OUTPUT = ROOT / "docs/plans/phase2-public-api-tiering.json"


INTROSPECTION_DELETE = {
    "func Findobj": (
        "The Python spelling and any-typed root hide the artist ownership model.",
        "Use explicit artist ownership; internal tests may use an unexported traversal helper.",
    ),
    "func FindobjType": (
        "The Python-style generic helper depends on the same hidden any-typed traversal contract.",
        "Use explicit artist ownership and ordinary Go type assertions.",
    ),
    "func Getp": (
        "String-keyed any-valued property reads duplicate typed Go getters and cover only a small subset of artists.",
        "Use the artist's typed getter.",
    ),
    "func GetpAll": (
        "A map[string]any snapshot is incomplete, weakly typed, and duplicates the typed artist API.",
        "Use the artist's typed getters.",
    ),
    "func Setp": (
        "String-keyed any-valued mutation warns and ignores rejected input instead of returning an error.",
        "Use the artist's typed setter.",
    ),
    "method ArtistRasterization.Property": (
        "This method exists only to support the deleted Python-style property bag.",
        "Use typed ArtistRasterization getters.",
    ),
    "method ArtistRasterization.PropertyNames": (
        "This method exists only to support the deleted Python-style property bag.",
        "Use the documented typed ArtistRasterization API.",
    ),
    "method ArtistRasterization.SetProperty": (
        "This method exists only to support the deleted Python-style property bag.",
        "Use typed ArtistRasterization setters.",
    ),
    "method Axes.Findobj": (
        "The Python spelling exposes traversal without a stable public child-container contract.",
        "Use explicit Axes artist ownership; internal tests may use an unexported traversal helper.",
    ),
    "method Figure.Findobj": (
        "The Python spelling exposes traversal without a stable public child-container contract.",
        "Use explicit Figure artist ownership; internal tests may use an unexported traversal helper.",
    ),
    "method Line2D.Property": (
        "This method exists only to support the deleted Python-style property bag.",
        "Use typed Line2D getters.",
    ),
    "method Line2D.PropertyNames": (
        "This method exists only to support the deleted Python-style property bag.",
        "Use the documented typed Line2D API.",
    ),
    "method Line2D.SetProperty": (
        "This method exists only to support the deleted Python-style property bag.",
        "Use typed Line2D setters.",
    ),
    "type PropertyBag": (
        "The string-keyed any-valued protocol is incomplete and conflicts with the typed Go API.",
        "Implement and call typed artist methods.",
    ),
}

UNITS_DELETE = {
    "method Axes.BarUnits": (
        "Unit conversion belongs in the primary plotting entry point under the Phase 2 error convention.",
        "Axes.Bar with unit-capable input returning (*Bar2D, error).",
    ),
    "method Axes.FillBetweenUnits": (
        "Unit conversion belongs in the primary plotting entry point under the Phase 2 error convention.",
        "Axes.FillBetween with unit-capable input returning (*Fill2D, error).",
    ),
    "method Axes.PlotUnits": (
        "Unit conversion belongs in the primary plotting entry point under the Phase 2 error convention.",
        "Axes.Plot with unit-capable input returning (*Line2D, error).",
    ),
    "method Axes.ScatterUnits": (
        "Unit conversion belongs in the primary plotting entry point under the Phase 2 error convention.",
        "Axes.Scatter with unit-capable input returning (*Scatter2D, error).",
    ),
}

DELETE = {
    **{("core", symbol): decision for symbol, decision in INTROSPECTION_DELETE.items()},
    **{("core", symbol): decision for symbol, decision in UNITS_DELETE.items()},
    (
        "render",
        "type CapabilityBridgeReporter",
    ): (
        "String-named bridge reporting is superseded by typed runtime capability status reporting.",
        "Use backends runtime capability status reporting with backends.Capability and CapabilityStatus.",
    ),
}

DEMOTE = {
    (
        "render",
        "type RendererModeReporter",
    ): (
        "Renderer mode labels are backend-registry reporting metadata, not a drawing contract.",
        "backends.RendererModeReporter",
        "backends",
    ),
}


def generate() -> bytes:
    frozen_bytes = FROZEN_API.read_bytes()
    frozen = json.loads(frozen_bytes)
    rows: list[dict[str, str]] = []

    for package in frozen["packages"]:
        package_name = package["dir"]
        for symbol in package["symbols"]:
            key = (package_name, symbol["id"])
            row = {
                "package": package_name,
                "id": symbol["id"],
                "disposition": "keep",
            }
            if key in DELETE:
                rationale, replacement = DELETE[key]
                row.update(
                    disposition="delete",
                    rationale=rationale,
                    replacement=replacement,
                )
            elif key in DEMOTE:
                rationale, replacement, target_package = DEMOTE[key]
                row.update(
                    disposition="demote",
                    rationale=rationale,
                    replacement=replacement,
                    target_package=target_package,
                )
            elif package_name == "render" and symbol["declaration"].startswith(
                "type "
            ) and " interface{" in symbol["declaration"]:
                row["rationale"] = (
                    "Retain the named renderer contract as part of the public "
                    "backend extension SPI."
                )
            rows.append(row)

    artifact = {
        "schema_version": 1,
        "generated_by": "docs/plans/generate_phase2_public_api_tiering.py",
        "baseline": {
            "path": "docs/plans/phase2-prebreak-public-api.json",
            "sha256": hashlib.sha256(frozen_bytes).hexdigest(),
            "symbol_count": len(rows),
        },
        "symbols": rows,
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
