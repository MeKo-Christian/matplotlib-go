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


def boxplot_basic(out_dir):
    Path(out_dir).mkdir(parents=True, exist_ok=True)

    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 10)
    ax.set_title("Box Plots")
    ax.set_xlabel("Group")
    ax.set_ylabel("Value")

    datasets = [
        [0.9, 1.0, 1.1, 1.2, 1.3, 1.45, 1.5, 1.7, 1.8],
        [4.0, 4.2, 4.3, 4.5, 4.8, 5.0, 5.4, 5.8, 9.4],
        [2.0, 2.1, 2.1, 2.2, 2.3, 2.4, 2.4, 2.6, 3.8],
    ]
    positions = [1.0, 2.0, 3.0]
    colors = [
        (0.25, 0.55, 0.82, 0.75),
        (0.80, 0.45, 0.20, 0.75),
        (0.35, 0.70, 0.35, 0.75),
    ]

    bp = ax.boxplot(
        datasets,
        positions=positions,
        widths=0.55,
        patch_artist=True,
        showfliers=True,
        boxprops=dict(linewidth=1.2, color=(0, 0, 0, 1)),
        whiskerprops=dict(linewidth=1.2, color=(0, 0, 0, 1)),
        capprops=dict(linewidth=1.2, color=(0, 0, 0, 1)),
        medianprops=dict(linewidth=1.8, color=(0, 0, 0, 1)),
        flierprops=dict(
            marker="o",
            markerfacecolor=(0, 0, 0, 1),
            markeredgecolor=(0, 0, 0, 1),
            markersize=4,
        ),
        manage_ticks=False,
    )
    ax.set_axisbelow(True)
    ax.yaxis.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
    for patch, color in zip(bp["boxes"], colors):
        patch.set_facecolor(color)
        patch.set_alpha(color[3])

    save(fig, out_dir, "boxplot_basic")


PLOT = boxplot_basic


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
