#!/usr/bin/env python3
"""Focused patch style, connection style, and hatch-density reference fixture."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def patch_style_matrix(out_dir):
    fig = make_fig_px(720, 420)
    ax = fig.add_axes(go_rect(0.03, 0.06, 0.97, 0.96))
    ax.set_xlim(0, 12)
    ax.set_ylim(0, 8)
    ax.set_axis_off()

    styles = [
        "square,pad=0.10",
        "round,pad=0.10,rounding_size=0.16",
        "round4,pad=0.10,rounding_size=0.16",
        "sawtooth,pad=0.10,tooth_size=0.12",
        "roundtooth,pad=0.10,tooth_size=0.12",
        "circle,pad=0.10",
        "ellipse,pad=0.10",
        "larrow,pad=0.10",
        "rarrow,pad=0.10",
        "darrow,pad=0.10",
    ]
    colors = [
        (0.34, 0.66, 0.82, 0.82),
        (0.90, 0.52, 0.28, 0.82),
        (0.47, 0.72, 0.43, 0.82),
        (0.78, 0.48, 0.73, 0.82),
        (0.86, 0.70, 0.34, 0.82),
    ]
    for i, style in enumerate(styles):
        col = i % 5
        row = i // 5
        ax.add_patch(mpatches.FancyBboxPatch(
            (0.65 + col * 2.25, 6.55 - row * 1.15), 1.35, 0.62,
            boxstyle=style,
            facecolor=colors[col],
            edgecolor=(0.13, 0.15, 0.18, 1.0),
            linewidth=1.0,
            mutation_scale=1.0,
        ))

    for i, hatch in enumerate(["/", "//", "o", "oo", ".", "..", "*", "**"]):
        rect = mpatches.Rectangle(
            (0.75 + i * 1.38, 3.58), 0.9, 0.78,
            facecolor=(0.92, 0.91, 0.84, 1.0),
            edgecolor=(0.18, 0.22, 0.25, 1.0),
            linewidth=0.85,
            hatch=hatch,
        )
        ax.add_patch(rect)

    arrows = [
        ((0.9, 2.25), (3.1, 2.6), "->,head_length=0.35,head_width=0.22",
         "arc,armA=0.9,armB=0.65,rad=0.18",
         (0.28, 0.48, 0.82, 1.0), (0.12, 0.24, 0.47, 1.0)),
        ((4.0, 2.62), (6.25, 1.88), "|-|",
         "bar,fraction=0.25,angle=0",
         (0.0, 0.0, 0.0, 0.0), (0.66, 0.28, 0.23, 1.0)),
        ((7.05, 2.08), (10.8, 2.58), "wedge,tail_width=0.26,shrink_factor=0.35",
         "arc3,rad=0.22",
         (0.40, 0.66, 0.35, 0.82), (0.20, 0.38, 0.18, 1.0)),
    ]
    for pos_a, pos_b, arrowstyle, connectionstyle, face, edge in arrows:
        ax.add_patch(mpatches.FancyArrowPatch(
            pos_a,
            pos_b,
            arrowstyle=arrowstyle,
            connectionstyle=connectionstyle,
            mutation_scale=15,
            facecolor=face,
            edgecolor=edge,
            linewidth=1.25,
        ))

    save(fig, out_dir, "patch_style_matrix")


PLOT = patch_style_matrix


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
