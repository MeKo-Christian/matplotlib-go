#!/usr/bin/env python3
from __future__ import annotations

import argparse
import os
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))
os.environ.setdefault("MPLCONFIGDIR", os.path.join(tempfile.gettempdir(), "matplotlib-go-cache"))

from test.parity import CASE_IDS, load_plot


def requested_ids(raw: list[str] | None) -> list[str]:
    requested: list[str] = []
    for item in raw or []:
        requested.extend(part.strip() for part in item.split(",") if part.strip())
    if not requested or requested == ["all"]:
        return list(CASE_IDS)
    unknown = sorted(set(requested) - set(CASE_IDS))
    if unknown:
        raise SystemExit(f"unknown parity example IDs: {', '.join(unknown)}")
    return requested


def main() -> None:
    parser = argparse.ArgumentParser(description="Render Matplotlib parity examples.")
    parser.add_argument("--output-dir", default=".", help="Directory to write PNG files")
    parser.add_argument("--id", dest="ids", action="append", help="Example ID to render; repeatable or comma-separated")
    parser.add_argument("--all", action="store_true", help="Render every parity example")
    parser.add_argument("--list", action="store_true", help="List available parity example IDs and exit")
    args = parser.parse_args()

    if args.list:
        for name in CASE_IDS:
            print(name)
        return

    ids = list(CASE_IDS) if args.all else requested_ids(args.ids)
    os.makedirs(args.output_dir, exist_ok=True)
    for name in ids:
        load_plot(name)(args.output_dir)


if __name__ == "__main__":
    main()
