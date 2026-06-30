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

def patch_showcase(out_dir):
    fig = make_fig_px(930, 340)

    left = fig.add_axes(go_rect(0.05, 0.16, 0.31, 0.88))
    left.set_title("Patch Primitives")
    left.set_xlim(0, 6)
    left.set_ylim(0, 4)
    left.add_patch(mpatches.Rectangle(
        (0.6, 0.7), 1.5, 1.0,
        facecolor=(0.95, 0.70, 0.23, 0.86),
        edgecolor=(0.48, 0.27, 0.08, 1.0),
        linewidth=1.1,
        hatch="/",
    ))
    left.add_patch(mpatches.Circle(
        (3.0, 1.25), 0.56,
        facecolor=(0.22, 0.57, 0.82, 0.82),
        edgecolor=(0.11, 0.29, 0.44, 1.0),
        linewidth=1.0,
    ))
    left.add_patch(mpatches.Ellipse(
        (4.8, 2.75), 1.55, 0.95, angle=28,
        facecolor=(0.23, 0.72, 0.51, 0.80),
        edgecolor=(0.10, 0.36, 0.24, 1.0),
        linewidth=1.0,
    ))
    left.add_patch(mpatches.Polygon(
        [[2.15, 3.2], [2.85, 2.25], [1.35, 2.45]],
        closed=True,
        facecolor=(0.84, 0.34, 0.34, 0.82),
        edgecolor=(0.48, 0.14, 0.14, 1.0),
        linewidth=1.0,
    ))

    middle = fig.add_axes(go_rect(0.37, 0.16, 0.63, 0.88))
    middle.set_title("Fancy Arrow + Path")
    middle.set_xlim(0, 6)
    middle.set_ylim(0, 4)
    middle.add_patch(mpatches.FancyArrow(
        0.9, 3.2, 2.2, -1.0,
        width=0.18,
        head_width=0.62,
        head_length=0.62,
        facecolor=(0.91, 0.42, 0.22, 0.88),
        edgecolor=(0.58, 0.22, 0.10, 1.0),
        linewidth=1.0,
        length_includes_head=True,
    ))
    star_vertices = [
        (4.15, 0.95), (4.45, 1.70), (5.22, 1.75), (4.63, 2.22), (4.84, 2.96),
        (4.15, 2.54), (3.46, 2.96), (3.67, 2.22), (3.08, 1.75), (3.85, 1.70), (4.15, 0.95),
    ]
    star_codes = [mpath.Path.MOVETO] + [mpath.Path.LINETO] * 9 + [mpath.Path.CLOSEPOLY]
    middle.add_patch(mpatches.PathPatch(
        mpath.Path(star_vertices, star_codes),
        facecolor=(0.76, 0.76, 0.86, 0.72),
        edgecolor=(0.29, 0.29, 0.45, 1.0),
        linewidth=1.0,
        hatch="x",
    ))

    right = fig.add_axes(go_rect(0.69, 0.16, 0.95, 0.88))
    right.set_title("Fancy Boxes")
    right.set_xlim(0, 6)
    right.set_ylim(0, 4)
    right.add_patch(mpatches.FancyBboxPatch(
        (0.9, 0.8), 2.1, 1.25,
        boxstyle="round,pad=0.14,rounding_size=0.24",
        facecolor=(0.29, 0.67, 0.78, 0.28),
        edgecolor=(0.10, 0.37, 0.45, 1.0),
        linewidth=1.0,
        hatch="/",
    ))
    right.add_patch(mpatches.FancyBboxPatch(
        (3.35, 1.55), 1.65, 1.05,
        boxstyle="square,pad=0.10",
        facecolor=(0.96, 0.87, 0.60, 0.82),
        edgecolor=(0.50, 0.39, 0.12, 1.0),
        linewidth=1.0,
    ))

    save(fig, out_dir, "patch_showcase")

PLOT = patch_showcase


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
