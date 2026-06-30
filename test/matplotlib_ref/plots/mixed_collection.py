#!/usr/bin/env python3
"""Matplotlib reference plot for mixed path collection coverage."""

from __future__ import annotations

from pathlib import Path
import argparse
import math
import sys

import matplotlib.collections as mcoll
import matplotlib.ticker as mticker

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def rect_path(x, y, w, h):
    return mpath.Path([(x, y), (x + w, y), (x + w, y + h), (x, y + h), (x, y)])


def triangle_path(cx, cy, r):
    verts = []
    for i in range(3):
        angle = -math.pi / 2 + float(i) * 2 * math.pi / 3
        verts.append((cx + r * math.cos(angle), cy + r * math.sin(angle)))
    verts.append(verts[0])
    return mpath.Path(verts)


def diamond_path(cx, cy, r):
    return mpath.Path([(cx, cy + r), (cx + r, cy), (cx, cy - r), (cx - r, cy), (cx, cy + r)])


def star_path(cx, cy, r):
    verts = []
    for i in range(10):
        radius = r if i % 2 == 0 else r * 0.45
        angle = -math.pi / 2 + float(i) * math.pi / 5
        verts.append((cx + radius * math.cos(angle), cy + radius * math.sin(angle)))
    verts.append(verts[0])
    return mpath.Path(verts)


def mixed_collection(out_dir):
    fig = make_fig_px(980, 620)
    ax = fig.add_axes(go_rect(0.10, 0.14, 0.94, 0.88))
    ax.set_title("RendererAgg mixed path collection")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 6)
    ax.xaxis.set_major_locator(mticker.MultipleLocator(2))
    ax.yaxis.set_major_locator(mticker.MultipleLocator(1))

    paths = [
        rect_path(0.8, 0.7, 1.4, 1.2),
        triangle_path(2.9, 1.0, 0.9),
        diamond_path(4.6, 1.2, 0.7),
        star_path(6.4, 1.0, 0.75),
        rect_path(7.8, 0.7, 1.1, 1.7),
        triangle_path(1.7, 4.0, 1.1),
        diamond_path(3.8, 4.1, 0.8),
        star_path(5.8, 4.0, 0.85),
        rect_path(7.5, 3.3, 1.5, 1.1),
    ]
    faces = [
        (0.13, 0.47, 0.70, 0.65),
        (1.00, 0.50, 0.05, 0.72),
        (0.17, 0.63, 0.17, 0.70),
        (0.84, 0.15, 0.16, 0.62),
        (0.58, 0.40, 0.74, 0.70),
        (0.55, 0.34, 0.29, 0.66),
        (0.89, 0.47, 0.76, 0.66),
        (0.50, 0.50, 0.50, 0.70),
        (0.74, 0.74, 0.13, 0.70),
    ]
    edges = [
        (0.02, 0.14, 0.23, 1),
        (0.46, 0.21, 0.02, 1),
        (0.02, 0.30, 0.06, 1),
        (0.45, 0.04, 0.05, 1),
        (0.28, 0.17, 0.42, 1),
        (0.31, 0.17, 0.14, 1),
        (0.44, 0.19, 0.37, 1),
        (0.20, 0.20, 0.20, 1),
        (0.36, 0.36, 0.04, 1),
    ]
    widths = [v for v in [1.1, 1.6, 1.0, 1.8, 1.2, 1.4, 1.0, 1.6, 1.2]]
    patches = [mpatches.PathPatch(path) for path in paths]
    collection = mcoll.PatchCollection(
        patches,
        facecolors=faces,
        edgecolors=edges,
        linewidths=widths,
        match_original=False,
        label="mixed collection",
    )
    ax.add_collection(collection)
    save(fig, out_dir, "mixed_collection")


PLOT = mixed_collection


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
