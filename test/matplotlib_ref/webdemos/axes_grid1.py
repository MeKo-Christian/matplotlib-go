#!/usr/bin/env python3
"""Axes Grid1 web demo reference module."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.webdemo_common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from webdemo_common import *  # noqa: F401,F403


def demo_axes_grid1(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(2, 2, 0.08, 0.92, 0.12, 0.88, 0.08, 0.12)
    data = heatmap_data(28, 32)
    for row in range(2):
        for col in range(2):
            ax = fig.add_axes(rects[row][col])
            ax.set_title(f"Image {row * 2 + col + 1}")
            im = ax.imshow(np.roll(data, row * 4 + col * 7, axis=1), cmap="viridis", origin="lower")
            if row == 0 and col == 1:
                inset = inset_axes(ax, width="38%", height="38%", loc="lower left")
                inset.imshow(data[8:20, 10:24], cmap="magma", origin="lower")
                inset.set_xticks([])
                inset.set_yticks([])
    fig.colorbar(im, ax=fig.axes, label="value", shrink=0.82)
    save(fig, out_dir, "axes_grid1")


DEMO = demo_axes_grid1


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
