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


def formatter_engineering_labels(out_dir):
    fig = make_fig_px(720, 400)
    add_panel(
        fig,
        go_rect(0.12, 0.58, 0.94, 0.88),
        "Micro Engineering Labels",
        [-2e-6, -1e-6, 0, 1e-6, 2e-6],
        mticker.EngFormatter(unit="V"),
    )
    add_panel(
        fig,
        go_rect(0.12, 0.16, 0.94, 0.46),
        "Kilohertz Engineering Labels",
        [0, 1000, 1500, 2000],
        mticker.EngFormatter(unit="Hz", places=1),
    )
    save(fig, out_dir, "formatter_engineering_labels")


def add_panel(fig, rect, title, ticks, formatter):
    ax = fig.add_axes(rect)
    ax.set_title(title)
    ax.set_xlabel("value")
    ax.set_ylabel("score")
    ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    ax.set_axisbelow(True)

    y = [0.2 + 0.15 * i for i in range(len(ticks))]
    ax.plot(ticks, y, color=(0.12, 0.47, 0.71), linewidth=lw(2.0))
    ax.set_xlim(ticks[0], ticks[-1])
    ax.set_ylim(0, 1)
    ax.xaxis.set_major_locator(mticker.FixedLocator(ticks))
    ax.xaxis.set_major_formatter(formatter)
    ax.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))


PLOT = formatter_engineering_labels


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
