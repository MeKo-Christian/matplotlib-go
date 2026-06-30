#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

def units_categories(out_dir):
    fig = make_fig_px(760, 360)
    left = fig.add_axes(go_rect(0.08, 0.20, 0.47, 0.86))
    left.set_title("Categorical X")
    left.set_ylabel("Count")
    left.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=0.5)
    left.set_axisbelow(True)
    left.bar(
        ["draft", "review", "ship", "watch"],
        [3, 8, 6, 4],
        color=(1.0, 0.50, 0.05),
        edgecolor=(0.60, 0.30, 0.03),
        linewidth=1.0,
        width=0.8,
    )
    left.margins(x=0.10, y=0.10)

    right = fig.add_axes(go_rect(0.58, 0.20, 0.94, 0.86))
    right.set_title("Categorical Y")
    right.set_xlabel("Hours")
    right.grid(True, axis="x", color=(0.8, 0.8, 0.8), linewidth=0.5)
    right.set_axisbelow(True)
    right.barh(
        ["north", "south", "east"],
        [4, 7, 5],
        color=(0.17, 0.63, 0.17),
        edgecolor=(0.09, 0.36, 0.09),
        linewidth=1.0,
        height=0.8,
    )
    right.margins(x=0.10, y=0.10)

    save(fig, out_dir, "units_categories")

PLOT = units_categories


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
