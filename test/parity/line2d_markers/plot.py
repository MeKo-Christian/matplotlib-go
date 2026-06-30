#!/usr/bin/env python3
"""Matplotlib reference plot for Line2D marker semantics."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import matplotlib.path as mpath

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def line2d_markers(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.09, 0.14, 0.94, 0.88))
    ax.set_title("Line2D Markers")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 6)

    ax.plot(
        [0.7, 2.0, 3.3],
        [5.35, 5.65, 5.35],
        color=(0.12, 0.47, 0.71, 1),
        linewidth=1.5,
        marker="o",
        markersize=9,
        markerfacecolor=(1.00, 0.50, 0.05, 0.55),
        markeredgecolor="auto",
        markeredgewidth=1.2,
        label="filled auto edge",
        solid_capstyle="butt",
    )

    ax.plot(
        [4.2, 5.5, 6.8],
        [5.35, 5.65, 5.35],
        color=(0.17, 0.63, 0.17, 1),
        linewidth=1.5,
        marker="s",
        markersize=9,
        markerfacecolor="none",
        markeredgecolor=(0.08, 0.28, 0.08, 1),
        markeredgewidth=1.4,
        label="face none",
        solid_capstyle="butt",
    )

    ax.plot(
        [7.6, 8.7, 9.8],
        [5.35, 5.65, 5.35],
        color=(0.84, 0.15, 0.16, 1),
        linewidth=1.5,
        marker="+",
        markersize=11,
        label="line-only",
        solid_capstyle="butt",
    )

    custom = mpath.Path(
        [(0, -0.55), (0.48, 0.36), (-0.48, 0.36), (0, -0.55)],
        [mpath.Path.MOVETO, mpath.Path.LINETO, mpath.Path.LINETO, mpath.Path.CLOSEPOLY],
    )
    ax.plot(
        [0.8, 2.1, 3.4],
        [3.85, 4.2, 3.85],
        color=(0.58, 0.40, 0.74, 1),
        linewidth=1.4,
        marker=custom,
        markersize=10,
        markerfacecolor=(0.58, 0.40, 0.74, 0.55),
        markeredgecolor=(0.25, 0.12, 0.40, 1),
        label="custom path",
        solid_capstyle="butt",
    )

    ax.plot(
        [4.3, 5.5, 6.7],
        [3.85, 4.2, 3.85],
        color=(0.55, 0.34, 0.29, 1),
        linewidth=1.4,
        marker=(5, 1, 18),
        markersize=11,
        markerfacecolor=(0.55, 0.34, 0.29, 0.65),
        markeredgecolor=(0.25, 0.12, 0.08, 1),
        label="tuple star",
        solid_capstyle="butt",
    )

    ax.plot(
        [7.4, 8.6, 9.8],
        [3.85, 4.2, 3.85],
        color=(0.50, 0.50, 0.50, 1),
        linewidth=1.4,
        marker="$f$",
        markersize=12,
        markerfacecolor=(0.89, 0.47, 0.76, 0.65),
        markeredgecolor=(0.35, 0.18, 0.32, 1),
        label="mathtext",
        solid_capstyle="butt",
    )

    for x, fillstyle, label in [
        (1.4, "left", "left"),
        (3.4, "right", "right"),
        (5.4, "top", "top"),
        (7.4, "bottom", "bottom"),
    ]:
        ax.plot(
            [x - 0.4, x, x + 0.4],
            [2.0, 2.32, 2.0],
            color=(0.09, 0.75, 0.81, 1),
            linewidth=1.2,
            marker="o",
            markersize=12,
            fillstyle=fillstyle,
            markerfacecolor=(0.09, 0.75, 0.81, 1),
            markerfacecoloralt=(0.95, 0.78, 0.18, 1),
            markeredgecolor=(0.05, 0.25, 0.28, 1),
            label=f"half {label}",
            solid_capstyle="butt",
        )

    ax.legend(loc="upper left")
    save(fig, out_dir, "line2d_markers")


PLOT = line2d_markers


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
