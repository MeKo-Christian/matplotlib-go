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

def fill_basic(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Fill to Baseline")
    ax.set_xlim(0, 10)
    ax.set_ylim(-1, 3)
    x  = [1, 2, 3, 4, 5, 6, 7, 8, 9]
    y  = [0.5, 1.8, 2.3, 1.2, 2.8, 1.9, 2.1, 1.5, 0.8]
    ax.fill_between(
        x, 0, y,
        facecolor=(0.3, 0.7, 0.9, 0.7),
        edgecolor=(0.1, 0.3, 0.5, 1.0),
        linewidth=2,
    )
    save(fig, out_dir, "fill_basic")

PLOT = fill_basic


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
