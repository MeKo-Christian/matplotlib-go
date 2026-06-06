#!/usr/bin/env python3
"""Matplotlib reference for Phase 17.6.2 convenience axes helpers."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def axes_convenience_helpers(out_dir):
    fig = make_fig_px(840, 540)

    box_ax = fig.add_axes(go_rect(0.07, 0.58, 0.30, 0.92))
    box_ax.set_title("Bxp")
    box_ax.set_xlim(0.4, 2.6)
    box_ax.set_ylim(0, 6)
    box_ax.bxp(
        [
            {"med": 2.2, "q1": 1.4, "q3": 3.1, "whislo": 0.8, "whishi": 4.1, "mean": 2.35, "fliers": [5.0], "label": "A"},
            {"med": 3.4, "q1": 2.5, "q3": 4.2, "whislo": 1.5, "whishi": 5.0, "mean": 3.6, "fliers": [], "label": "B"},
        ],
        showmeans=True,
        boxprops={"color": (0.18, 0.36, 0.70, 1), "linewidth": lw(1)},
        whiskerprops={"color": (0.18, 0.36, 0.70, 1), "linewidth": lw(1)},
        capprops={"color": (0.18, 0.36, 0.70, 1), "linewidth": lw(1)},
    )

    violin_ax = fig.add_axes(go_rect(0.38, 0.58, 0.61, 0.92))
    violin_ax.set_title("Violin")
    violin_ax.set_xlim(0.4, 2.6)
    violin_ax.set_ylim(0, 6)
    violin_ax.violin(
        [
            {"coords": [1, 2, 3, 4, 5], "vals": [0.2, 0.7, 1.0, 0.5, 0.15], "mean": 3.0, "median": 3.0, "min": 1, "max": 5, "quantiles": [2, 4]},
            {"coords": [1, 2, 3, 4, 5], "vals": [0.15, 0.5, 1.0, 0.7, 0.2], "mean": 3.2, "median": 3.3, "min": 1, "max": 5},
        ],
        showmeans=True,
        showmedians=True,
    )

    line_ax = fig.add_axes(go_rect(0.69, 0.58, 0.96, 0.92))
    line_ax.set_title("H/VLines")
    line_ax.set_xlim(0, 5)
    line_ax.set_ylim(0, 5)
    line_ax.hlines([1, 2.5, 4], [0.5], [4.5], colors=[(0.12, 0.30, 0.72, 1)], linewidth=lw(1.4))
    line_ax.vlines([1, 2.5, 4], [0.6], [4.4], colors=[(0.75, 0.18, 0.16, 1)], linewidth=lw(1.4))

    contour_ax = fig.add_axes(go_rect(0.18, 0.08, 0.82, 0.43))
    contour_ax.set_title("Clabel")
    contour_ax.set_xlim(0, 3)
    contour_ax.set_ylim(0, 3)
    contour = contour_ax.contour(
        [[0, 1, 2, 3], [1, 2, 3, 4], [2, 3, 4, 5], [3, 4, 5, 6]],
        levels=[2, 3, 4],
        colors=[(0.13, 0.20, 0.35, 1)],
        linewidths=lw(1),
    )
    contour_ax.clabel(contour, levels=[3], fontsize=9, colors=[(0.05, 0.05, 0.05, 1)])

    save(fig, out_dir, "axes_convenience_helpers")


PLOT = axes_convenience_helpers


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
