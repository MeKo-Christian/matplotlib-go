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


def locator_linear_labels(out_dir):
    fig = make_fig_px(720, 540)
    axes = [
        fig.add_axes(go_rect(0.12, 0.70, 0.92, 0.92)),
        fig.add_axes(go_rect(0.12, 0.40, 0.92, 0.62)),
        fig.add_axes(go_rect(0.12, 0.10, 0.92, 0.32)),
    ]
    titles = ["Default Linear Locator", "LinearLocator(5)", "MultipleLocator(1.5)"]
    x = [0, 1.5, 3, 4.5, 6]
    y = [0.20, 0.38, 0.64, 0.78, 0.90]

    for ax, title in zip(axes, titles):
        ax.set_title(title)
        ax.set_xlabel("x")
        ax.set_ylabel("y")
        ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
        ax.set_axisbelow(True)
        ax.plot(x, y, color=(0.12, 0.47, 0.71), linewidth=lw(2.0))
        ax.set_xlim(0, 6)
        ax.set_ylim(0, 1)
        ax.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))

    axes[1].xaxis.set_major_locator(mticker.LinearLocator(numticks=5))
    axes[2].xaxis.set_major_locator(mticker.MultipleLocator(1.5))

    save(fig, out_dir, "locator_linear_labels")


PLOT = locator_linear_labels


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
