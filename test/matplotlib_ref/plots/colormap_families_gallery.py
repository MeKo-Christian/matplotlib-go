#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import argparse
import math
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def gradient_strip(rows, cols):
    return [[x / (cols - 1) for x in range(cols)] for _ in range(rows)]


def diverging_strip(rows, cols):
    return [[2 * x / (cols - 1) - 1 for x in range(cols)] for _ in range(rows)]


def categorical_strip(rows, cols, n):
    return [[math.floor(x * n / cols) for x in range(cols)] for _ in range(rows)]


def cyclic_strip(rows, cols):
    return [[(x / (cols - 1) + 0.08 * math.sin((x / (cols - 1)) * 2 * math.pi)) % 1 for x in range(cols)] for _ in range(rows)]


def colormap_families_gallery(out_dir):
    fig = make_fig_px(900, 520)
    rows = [
        ("sequential: viridis", "viridis", gradient_strip(3, 80), 0, 1),
        ("sequential reversed: viridis_r", "viridis_r", gradient_strip(3, 80), 0, 1),
        ("perceptual: plasma", "plasma", gradient_strip(3, 80), 0, 1),
        ("diverging: RdBu", "RdBu", diverging_strip(3, 80), -1, 1),
        ("qualitative: tab10", "tab10", categorical_strip(3, 80, 10), 0, 9),
        ("cyclic: twilight", "twilight", cyclic_strip(3, 80), 0, 1),
    ]

    left, right, top, height, gap = 0.35, 0.95, 0.88, 0.09, 0.045
    for i, (title, cmap, data, vmin, vmax) in enumerate(rows):
        y1 = top - i * (height + gap)
        ax = fig.add_axes(go_rect(left, y1 - height, right, y1))
        ax.set_xticks([])
        ax.set_yticks([])
        ax.imshow(data, cmap=cmap, vmin=vmin, vmax=vmax, origin="lower", extent=(0, len(data[0]), 0, len(data)), aspect="auto", interpolation="nearest")
        fig.text(0.06, y1 - height / 2, title, ha="left", va="center", fontsize=10)

    fig.text(0.06, 0.94, "Colormap Family Gallery", fontsize=13)
    save(fig, out_dir, "colormap_families_gallery")


PLOT = colormap_families_gallery


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
