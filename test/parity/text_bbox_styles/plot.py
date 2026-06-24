#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2]))
    from matplotlib_ref.common import *  # noqa: F401,F403


FACE = (0.85, 0.92, 1.0)
EDGE = (0.20, 0.30, 0.60)

# Each FancyBboxPatch style reachable through a Text bbox, addressed by the same
# boxstyle spec strings the Go bridge understands. Layout is a 2×5 grid.
BOXES = [
    (0.27, 0.86, "square", "square,pad=0.4"),
    (0.73, 0.86, "circle", "circle,pad=0.4"),
    (0.27, 0.68, "round", "round,pad=0.4"),
    (0.73, 0.68, "round4", "round4,pad=0.4"),
    (0.27, 0.50, "ellipse", "ellipse,pad=0.4"),
    (0.73, 0.50, "sawtooth", "sawtooth,pad=0.4"),
    (0.27, 0.32, "roundtooth", "roundtooth,pad=0.4"),
    (0.73, 0.32, "rarrow", "rarrow,pad=0.4"),
    (0.27, 0.14, "larrow", "larrow,pad=0.4"),
    (0.73, 0.14, "darrow", "darrow,pad=0.4"),
]


def text_bbox_styles(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.05, 0.05, 0.95, 0.95))
    ax.set_xlim(0, 1)
    ax.set_ylim(0, 1)
    ax.axis("off")

    for x, y, label, style in BOXES:
        ax.text(
            x, y, label,
            transform=ax.transAxes, ha="center", va="center", fontsize=16,
            bbox=dict(boxstyle=style, facecolor=FACE, edgecolor=EDGE),
        )
    save(fig, out_dir, "text_bbox_styles")


PLOT = text_bbox_styles


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
