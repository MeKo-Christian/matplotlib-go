#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def text_layout_gallery(out_dir):
    fig = make_fig_px(900, 560)
    ax = fig.add_axes(go_rect(0.08, 0.10, 0.94, 0.88))
    ax.set_title("Text Layout Gallery")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 4)
    ax.set_xticks([])
    ax.set_yticks([])

    add_alignment_samples(ax)
    add_rotated_samples(ax)
    add_multiline_samples(ax)
    add_wrapped_sample(ax)
    fig.text(0.05, 0.94, "Alignment, rotation, multiline, wrapping, and bbox text", fontsize=11)
    save(fig, out_dir, "text_layout_gallery")


def add_alignment_samples(ax):
    cross_color = (0.70, 0.70, 0.74, 1)
    for x in [0.9, 1.6, 2.3]:
        ax.plot([x, x], [2.6, 3.5], color=cross_color)
    for y in [2.75, 3.05, 3.35]:
        ax.plot([0.5, 2.7], [y, y], color=cross_color)
    ax.text(0.9, 3.35, "left/top", ha="left", va="top", fontsize=10)
    ax.text(1.6, 3.05, "center", ha="center", va="center", fontsize=10)
    ax.text(2.3, 2.75, "right/bottom", ha="right", va="bottom", fontsize=10)


def add_rotated_samples(ax):
    ax.text(
        3.45,
        3.35,
        "rotation\nmode",
        ha="center",
        va="center",
        rotation=-32,
        bbox=text_box(0.25, (1.00, 0.92, 0.78, 0.78), (0.68, 0.38, 0.12, 1)),
    )
    ax.text(
        4.50,
        3.18,
        "anchor",
        ha="center",
        va="center",
        rotation=34,
        rotation_mode="anchor",
        fontsize=12,
        bbox=text_box(0.22, (0.88, 0.94, 1.00, 0.78), (0.20, 0.38, 0.68, 1)),
    )


def add_multiline_samples(ax):
    ax.text(
        0.75,
        1.75,
        "multi-line\nleft aligned\ntext",
        ha="left",
        va="top",
        multialignment="left",
        linespacing=1.3,
        fontsize=11,
        bbox=text_box(0.28, (0.94, 0.97, 0.92, 0.86), (0.24, 0.50, 0.20, 1)),
    )
    ax.text(
        2.55,
        1.65,
        "right\naligned\nblock",
        ha="right",
        va="top",
        multialignment="right",
        linespacing=1.15,
        fontsize=11,
        bbox=text_box(0.28, (0.98, 0.94, 1.00, 0.86), (0.45, 0.26, 0.58, 1)),
    )


def add_wrapped_sample(ax):
    ax.text(
        3.45,
        1.78,
        "wrapped text uses a fixed display width inside a rounded bbox",
        ha="left",
        va="top",
        wrap=True,
        fontsize=11,
        bbox=text_box(0.28, (1.00, 0.98, 0.88, 0.88), (0.52, 0.45, 0.16, 1)),
    )


def text_box(pad, face, edge):
    return {
        "boxstyle": f"round,pad={pad}",
        "facecolor": face,
        "edgecolor": edge,
        "linewidth": 0.9,
    }


PLOT = text_layout_gallery


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
