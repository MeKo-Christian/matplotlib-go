#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import argparse
import math
import sys

import matplotlib.colors as mcolors

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def log_data(rows, cols):
    return [[10 ** (3 * (x + y) / (rows + cols - 2)) for x in range(cols)] for y in range(rows)]


def two_slope_data(rows, cols):
    out = []
    for y in range(rows):
        yy = y / (rows - 1)
        row = []
        for x in range(cols):
            xx = x / (cols - 1)
            row.append(6 * xx - 3 + 1.5 * (yy - 0.5))
        out.append(row)
    return out


def colorbar_variants_gallery(out_dir):
    fig = make_fig_px(1040, 720)

    ax = fig.add_axes(go_rect(0.07, 0.60, 0.34, 0.88))
    ax.set_title("LogNorm")
    ax.set_xticks([])
    ax.set_yticks([])
    im = ax.imshow(log_data(6, 8), cmap="magma", norm=mcolors.LogNorm(vmin=1, vmax=1000), origin="lower", extent=(0, 8, 0, 6), aspect="auto")
    fig.colorbar(im, ax=ax, label="log value", pad=0.03)

    ax = fig.add_axes(go_rect(0.57, 0.60, 0.84, 0.88))
    ax.set_title("TwoSlopeNorm")
    ax.set_xticks([])
    ax.set_yticks([])
    im = ax.imshow(two_slope_data(6, 8), cmap="RdBu", norm=mcolors.TwoSlopeNorm(vmin=-3, vcenter=0, vmax=6), origin="lower", extent=(0, 8, 0, 6), aspect="auto")
    fig.colorbar(im, ax=ax, label="anomaly", pad=0.03)

    ax = fig.add_axes(go_rect(0.07, 0.16, 0.34, 0.44))
    ax.set_title("BoundaryNorm")
    boundaries = [0, 1, 2, 3, 4]
    mesh = ax.pcolormesh(
        [0, 1, 2, 3, 4],
        [0, 1, 2, 3],
        [[0.2, 0.8, 1.2, 1.8], [2.2, 2.8, 3.2, 3.8], [0.5, 1.5, 2.5, 3.5]],
        cmap="viridis",
        norm=mcolors.BoundaryNorm(boundaries, 256),
    )
    fig.colorbar(mesh, ax=ax, label="band", pad=0.03, drawedges=True, ticks=boundaries)
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 3)

    ax = fig.add_axes(go_rect(0.57, 0.16, 0.84, 0.44))
    ax.set_title("extensions")
    cmap = plt.get_cmap("viridis").copy()
    cmap.set_under((0.08, 0.16, 0.72, 1))
    cmap.set_over((0.78, 0.12, 0.08, 1))
    mesh = ax.pcolormesh(
        [0, 1, 2, 3],
        [0, 1, 2],
        [[-0.35, 0.15, 0.35], [0.55, 0.85, 1.35]],
        cmap=cmap,
        vmin=0,
        vmax=1,
    )
    fig.colorbar(mesh, ax=ax, label="extended", pad=0.03, extend="both")
    ax.set_xlim(0, 3)
    ax.set_ylim(0, 2)

    fig.text(0.06, 0.95, "Colorbar Norms and Extensions", fontsize=13)
    save(fig, out_dir, "colorbar_variants_gallery")


PLOT = colorbar_variants_gallery


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
