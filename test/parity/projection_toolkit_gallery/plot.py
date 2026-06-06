#!/usr/bin/env python3
"""Parity wrapper for the projection toolkit gallery Matplotlib reference."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.plots.projection_toolkit_gallery import PLOT
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[3]))
    from test.matplotlib_ref.plots.projection_toolkit_gallery import PLOT


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
