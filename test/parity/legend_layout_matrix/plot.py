#!/usr/bin/env python3
"""Focused legend layout, proxy, and handler reference fixture."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def legend_layout_matrix(out_dir):
    fig = make_fig_px(720, 420)
    ax = fig.add_axes(go_rect(0.10, 0.14, 0.94, 0.90))
    ax.set_title("Legend Layout Matrix")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5)

    blue = (0.12, 0.31, 0.68, 1.0)
    orange = (0.86, 0.43, 0.16, 1.0)
    green = (0.21, 0.58, 0.27, 1.0)
    red = (0.72, 0.20, 0.18, 1.0)
    purple = (0.47, 0.31, 0.70, 1.0)

    ax.plot(
        [0.4, 1.4, 2.4, 3.4, 4.4, 5.4],
        [1.0, 1.7, 1.4, 2.2, 2.0, 2.7],
        color=blue,
        linewidth=2.0,
        label="line",
    )
    ax.scatter(
        [0.7, 1.7, 2.7, 3.7, 4.7],
        [3.3, 3.8, 3.1, 3.6, 3.2],
        color=orange,
        edgecolor=red,
        linewidths=1.0,
        marker="o",
        label="scatter",
    )
    ax.errorbar(
        [1.0, 2.3, 3.6, 4.9],
        [0.85, 1.15, 0.95, 1.25],
        yerr=[0.22, 0.18, 0.25, 0.20],
        color=green,
        linewidth=2.0,
        capsize=5,
        marker="s",
        markersize=5,
        label="errorbar",
    )
    (handler_line,) = ax.plot(
        [0.5, 1.6, 2.7, 3.8, 4.9],
        [4.5, 4.1, 4.35, 4.0, 4.25],
        color=purple,
        linewidth=2.0,
        label="handler patch",
    )

    handles, labels = ax.get_legend_handles_labels()
    handler_proxy = mpatches.Patch(
        facecolor=(0.74, 0.52, 0.83, 0.85),
        edgecolor=purple,
        linewidth=1.2,
        hatch="/",
    )
    handles = [handler_proxy if handle is handler_line else handle for handle in handles]

    collected = ax.legend(
        handles=handles,
        labels=labels,
        loc="upper left",
        title="Collected",
        ncols=2,
        columnspacing=2.0,
        markerscale=1.8,
        scatterpoints=3,
        frameon=True,
    )
    ax.add_artist(collected)

    proxy = mpatches.Patch(
        facecolor=(0.93, 0.77, 0.33, 0.92),
        edgecolor=(0.45, 0.30, 0.08, 1.0),
        linewidth=1.2,
        hatch="xx",
        label="proxy patch",
    )
    ax.legend(handles=[proxy], loc="lower right", frameon=False)

    save(fig, out_dir, "legend_layout_matrix")


PLOT = legend_layout_matrix


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
