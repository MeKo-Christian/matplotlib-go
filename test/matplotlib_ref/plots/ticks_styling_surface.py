#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import matplotlib.ticker as mticker

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2]))
    from matplotlib_ref.common import *  # noqa: F401,F403


def ticks_styling_surface(out_dir):
    fig = make_fig_px(720, 420)
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.90, 0.78))
    ax.set_title("Tick Styling Surface")
    ax.set_xlabel("top labels")
    ax.set_ylabel("right labels")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 12)
    ax.xaxis.set_label_position("top")
    ax.yaxis.set_label_position("right")
    ax.tick_params(
        axis="x",
        top=True,
        bottom=False,
        labeltop=True,
        labelbottom=False,
    )
    ax.tick_params(
        axis="y",
        right=True,
        left=False,
        labelright=True,
        labelleft=False,
    )
    ax.xaxis.set_major_locator(mticker.FixedLocator([0, 2, 4, 6]))
    ax.yaxis.set_major_locator(mticker.FixedLocator([0, 4, 8, 12]))
    ax.xaxis.set_minor_locator(mticker.MultipleLocator(1))
    ax.yaxis.set_minor_locator(mticker.MultipleLocator(2))
    ax.tick_params(
        axis="both",
        which="major",
        colors=(0.18, 0.42, 0.55, 1.0),
        length=px2pt(8),
        width=1.4,
        labelrotation=35,
    )
    ax.tick_params(
        axis="both",
        which="minor",
        colors=(0.45, 0.45, 0.45, 1.0),
        length=px2pt(4),
        width=0.8,
    )
    ax.grid(
        True,
        axis="both",
        which="major",
        color=(0.50, 0.60, 0.70, 0.65),
        linewidth=0.8,
        linestyle=(0, (4, 2)),
    )
    ax.grid(
        True,
        axis="both",
        which="minor",
        color=(0.70, 0.74, 0.78, 0.45),
        linewidth=0.5,
        linestyle=(0, (1, 2)),
    )
    ax.plot(
        [0, 1.5, 3, 4.5, 6],
        [1, 4, 5.5, 9, 11],
        color=(0.12, 0.47, 0.71),
        linewidth=2.0,
    )

    save(fig, out_dir, "ticks_styling_surface")


PLOT = ticks_styling_surface


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
