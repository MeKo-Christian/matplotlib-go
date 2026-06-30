#!/usr/bin/env python3
"""Matplotlib reference plot for RendererAgg clip-path batch coverage."""

from __future__ import annotations

from pathlib import Path
import argparse
import math
import sys

import matplotlib.patches as mpatches
import matplotlib.path as mpath
import matplotlib.ticker as mticker

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def clip_batch_path():
    points = [
        (0.55, 1.10),
        (2.05, 0.50),
        (3.10, 1.05),
        (5.35, 0.80),
        (4.70, 2.45),
        (5.50, 4.05),
        (3.70, 3.80),
        (2.55, 5.05),
        (1.75, 3.55),
        (0.55, 3.85),
        (1.20, 2.35),
    ]
    vertices = points + [points[0]]
    codes = [mpath.Path.MOVETO] + [mpath.Path.LINETO] * (len(points) - 1) + [mpath.Path.CLOSEPOLY]
    return mpath.Path(vertices, codes)


def clip_path_batch(out_dir):
    fig = make_fig_px(980, 620)
    ax = fig.add_axes(go_rect(0.10, 0.14, 0.94, 0.88))
    ax.set_title("RendererAgg clip path batch")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5.4)
    ax.xaxis.set_major_locator(mticker.MultipleLocator(1))
    ax.yaxis.set_major_locator(mticker.MultipleLocator(1))
    ax.grid(axis="y")

    xedges = np.array([0, 0.75, 1.5, 2.35, 3.1, 4.0, 4.85, 5.45, 6.0], dtype=float)
    yedges = np.array([0, 0.7, 1.55, 2.4, 3.2, 4.15, 5.4], dtype=float)
    data = np.empty((len(yedges) - 1, len(xedges) - 1), dtype=float)
    for yi in range(data.shape[0]):
        for xi in range(data.shape[1]):
            cx = (xedges[xi] + xedges[xi + 1]) * 0.5
            cy = (yedges[yi] + yedges[yi + 1]) * 0.5
            data[yi, xi] = 0.45 + 0.42 * math.sin(cx * 1.15) + 0.33 * math.cos(cy * 1.35) + 0.06 * ((xi + yi) % 3)

    mesh = ax.pcolormesh(
        xedges,
        yedges,
        data,
        shading="flat",
        cmap="viridis",
        vmin=-0.35,
        vmax=1.15,
        alpha=0.84,
        edgecolors=[(0.97, 0.97, 0.97, 0.72)],
        linewidth=0.55,
        antialiased=True,
    )
    clip_path = clip_batch_path()
    mesh.set_clip_path(clip_path, transform=ax.transData)

    outline = mpatches.PathPatch(
        clip_path,
        transform=ax.transData,
        facecolor="none",
        edgecolor=(0.05, 0.08, 0.12, 1),
        linewidth=2.0,
        joinstyle="miter",
    )
    ax.add_patch(outline)
    save(fig, out_dir, "clip_path_batch")


PLOT = clip_path_batch


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
