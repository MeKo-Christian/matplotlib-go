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


def configure_axes(ax, title, color):
    ax.set_title(title)
    ax.set_xlim(0, 1)
    ax.set_ylim(0, 1)
    ax.set_xticks([])
    ax.set_yticks([])
    ax.plot(
        [0.10, 0.38, 0.66, 0.90],
        [0.20, 0.76, 0.35, 0.82],
        color=color,
        linewidth=lw(1.8),
    )
    ax.text(
        0.08,
        0.08,
        title,
        transform=ax.transAxes,
        fontsize=9,
        color=color,
        ha="left",
        va="bottom",
    )


def layout_bbox_helpers(out_dir):
    fig = make_fig_px(720, 420)

    left_rect = matplotlib.transforms.Bbox.from_extents(0.10, 0.19, 0.45, 0.78)
    right_rect = matplotlib.transforms.Bbox.from_extents(0.56, 0.34, 0.88, 0.82)
    union_rect = matplotlib.transforms.Bbox.union([left_rect, right_rect])
    padded_rect = union_rect.padded(0.035)
    anchored_rect = matplotlib.transforms.Bbox.from_bounds(0, 0, 0.24, 0.18).anchored("SE", padded_rect)

    configure_axes(fig.add_axes(left_rect.bounds), "left bbox", (0.12, 0.47, 0.71, 1.0))
    configure_axes(fig.add_axes(right_rect.bounds), "right bbox", (0.17, 0.63, 0.17, 1.0))
    configure_axes(fig.add_axes(anchored_rect.bounds), "anchored", (0.84, 0.15, 0.16, 1.0))

    fig.patches.append(mpatches.Rectangle(
        padded_rect.p0,
        padded_rect.width,
        padded_rect.height,
        transform=fig.transFigure,
        facecolor=(0.95, 0.74, 0.20, 0.08),
        edgecolor=(0.42, 0.34, 0.12, 1.0),
        linewidth=lw(1.4),
        linestyle=(0, (6, 4)),
        zorder=10,
    ))

    save(fig, out_dir, "layout_bbox_helpers")


PLOT = layout_bbox_helpers


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
