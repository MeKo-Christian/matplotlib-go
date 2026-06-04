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

def gridspec_composition(out_dir):
    fig = make_fig_px(960, 640)
    outer = fig.add_gridspec(
        2,
        2,
        left=0.08,
        right=0.96,
        bottom=0.10,
        top=0.92,
        wspace=0.06,
        hspace=0.28,
        width_ratios=[2, 1],
    )

    main_ax = fig.add_subplot(outer[:, 0])
    _composition_configure_axes(main_ax, "Main Span", [0, 1, 2, 3, 4], [1.2, 2.8, 2.1, 3.6, 3.1], (0.15, 0.35, 0.72))
    main_ax.set_xticks([0, 1, 2, 3, 4])
    main_ax.set_yticks([1.0, 1.5, 2.0, 2.5, 3.0, 3.5])

    nested = outer[0, 1].subgridspec(2, 1, hspace=0.75)
    top_right = fig.add_subplot(nested[0, 0])
    _composition_configure_axes(top_right, "Nested Top", [0, 1, 2, 3], [3.4, 2.6, 2.9, 1.8], (0.72, 0.32, 0.18))
    top_right.set_xticks([0, 1, 2, 3])
    top_right.set_yticks([2, 3])

    bottom_right = fig.add_subplot(nested[1, 0], sharex=top_right)
    _composition_configure_axes(bottom_right, "Nested Bottom", [0, 1, 2, 3], [1.0, 1.6, 1.3, 2.2], (0.18, 0.55, 0.34))
    bottom_right.set_xticks([0, 1, 2, 3])
    bottom_right.set_yticks([1, 2])

    sub = fig.add_subfigure(outer[1, 1])
    inset = sub.add_subplot(1, 1, 1)
    _composition_configure_axes(inset, "SubFigure", [0, 1, 2, 3], [2.0, 2.4, 1.9, 2.7], (0.55, 0.22, 0.50))
    inset.set_xticks([0, 1, 2, 3])
    inset.set_yticks([2.0, 2.2, 2.4, 2.6])

    save(fig, out_dir, "gridspec_composition")

PLOT = gridspec_composition


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
