#!/usr/bin/env python3
"""Matplotlib reference plot for Line2D data, gaps, and markevery semantics."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import numpy as np

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def line2d_semantics(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.14, 0.94, 0.88))
    ax.set_title("Line2D Semantics")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 5)

    (mutated,) = ax.plot(
        [0.6, 2.2, 3.8],
        [4.65, 4.65, 4.65],
        color=(0.10, 0.26, 0.78, 1),
        linewidth=3,
        label="set_data",
        solid_capstyle="butt",
    )
    mutated.set_data([0.6, 1.8, 3.0, 4.2], [4.55, 4.25, 4.65, 4.35])

    ax.plot(
        [0.6, 1.6, np.nan, 2.4, 3.5, np.inf, 4.3, 5.2],
        [3.45, 3.85, 3.1, 3.15, 3.75, 3.4, 3.25, 3.65],
        color=(0.84, 0.15, 0.16, 1),
        linewidth=3,
        solid_capstyle="butt",
    )

    (gap,) = ax.plot(
        [5.6, 9.4],
        [3.55, 3.55],
        color=(0.08, 0.45, 0.18, 1),
        linewidth=4,
        gapcolor=(0.92, 0.48, 0.10, 1),
        solid_capstyle="butt",
        dash_capstyle="butt",
    )
    gap.set_dashes([4, 2])

    x = [0.7, 1.5, 2.3, 3.1, 3.9, 4.7, 5.5, 6.3, 7.1, 7.9, 8.7, 9.5]
    y1 = [1.1, 1.45, 1.2, 1.65, 1.35, 1.75, 1.45, 1.85, 1.55, 1.95, 1.65, 2.05]
    ax.plot(
        x,
        y1,
        color=(0.58, 0.40, 0.74, 1),
        linewidth=1.5,
        marker="o",
        markersize=8,
        markevery=(1, 3),
        solid_capstyle="butt",
    )

    y2 = [0.55, 0.85, 0.65, 0.95, 0.75, 1.05, 0.85, 1.15, 0.95, 1.25, 1.05, 1.35]
    ax.plot(
        x,
        y2,
        color=(0.55, 0.34, 0.29, 1),
        linewidth=1.5,
        marker="s",
        markersize=7,
        markevery=[0, 4, 8, 11],
        solid_capstyle="butt",
    )

    save(fig, out_dir, "line2d_semantics")


PLOT = line2d_semantics


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
