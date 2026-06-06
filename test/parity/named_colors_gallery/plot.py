#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import argparse
import sys

import matplotlib.colors as mcolors
import matplotlib.patches as mpatches

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


SWATCHES = [
    ("b", "b"),
    ("tab:orange", "tab:orange"),
    ("tab:green", "tab:green"),
    ("rebeccapurple", "rebeccapurple"),
    ("mediumseagreen", "mediumseagreen"),
    ("goldenrod", "goldenrod"),
    ("xkcd:cloudy blue", "xkcd:cloudy blue"),
    ("xkcd:burnt orange", "xkcd:burnt orange"),
    ("0.25", "0.25"),
    ("#66c2a5", "#66c2a5"),
    ("C3", "C3"),
    ("rgba tuple", (0.15, 0.45, 0.65, 1)),
]


def named_colors_gallery(out_dir):
    fig = make_fig_px(900, 520)
    ax = fig.add_axes(go_rect(0.05, 0.08, 0.95, 0.88))
    ax.set_title("Named Color Swatches")
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 3)
    ax.set_xticks([])
    ax.set_yticks([])

    for i, (label, spec) in enumerate(SWATCHES):
        x = i % 4
        y = 2 - i // 4
        rect = mpatches.Rectangle(
            (x + 0.12, y + 0.28),
            0.76,
            0.42,
            facecolor=mcolors.to_rgba(spec),
            edgecolor=(0.15, 0.15, 0.15, 1),
            linewidth=0.8,
        )
        ax.add_patch(rect)
        ax.text(x + 0.50, y + 0.14, label, ha="center", va="center", fontsize=9)

    fig.text(0.05, 0.94, "CSS4, Tableau, xkcd, grayscale, cycle, hex, and tuple specs", fontsize=10)
    save(fig, out_dir, "named_colors_gallery")


PLOT = named_colors_gallery


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
