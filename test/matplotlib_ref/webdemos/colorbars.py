#!/usr/bin/env python3
"""Colorbar web demo reference module."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.webdemo_common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from webdemo_common import *  # noqa: F401,F403


def demo_colorbars(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(2, 2, 0.08, 0.90, 0.12, 0.90, 0.12, 0.16)
    data = heatmap_data(24, 28)
    specs = [
        ("Viridis", "viridis", None),
        ("Plasma", "plasma", None),
        ("Diverging", "coolwarm", (-1.0, 1.0)),
        ("Discrete", "tab10", None),
    ]
    for idx, (title, cmap, limits) in enumerate(specs):
        row, col = divmod(idx, 2)
        ax = fig.add_axes(rects[row][col])
        ax.set_title(title)
        kwargs = {"cmap": cmap, "origin": "lower"}
        if limits is not None:
            kwargs.update({"vmin": limits[0], "vmax": limits[1]})
        im = ax.imshow(data, **kwargs)
        fig.colorbar(im, ax=ax, label="value")
    save(fig, out_dir, "colorbars")


DEMO = demo_colorbars


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
