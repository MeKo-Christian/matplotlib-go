#!/usr/bin/env python3
"""Matplotlib reference for the grouped triangulation gallery."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import matplotlib

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def panel(row, col):
    left, right = 0.065, 0.955
    bottom, top = 0.11, 0.91
    hgap, vgap = 0.085, 0.12
    w = (right - left - hgap) / 2
    h = (top - bottom - vgap) / 2
    x0 = left + col * (w + hgap)
    y1 = top - row * (h + vgap)
    return go_rect(x0, y1 - h, x0 + w, y1)


def sample_triangulation():
    x = np.array([0.0, 0.85, 1.75, 2.85, 0.2, 1.1, 2.1, 0.55, 1.55, 2.55])
    y = np.array([0.0, 0.2, 0.05, 0.3, 1.0, 1.15, 1.25, 2.15, 2.3, 2.05])
    triangles = np.array(
        [
            [0, 1, 4],
            [1, 5, 4],
            [1, 2, 5],
            [2, 6, 5],
            [2, 3, 6],
            [4, 5, 7],
            [5, 8, 7],
            [5, 6, 8],
            [6, 9, 8],
        ]
    )
    values = np.sin(x * 1.4) + 0.7 * np.cos((y + 0.15) * 2.1)
    return mtri.Triangulation(x, y, triangles), values


def configure_tri_axes(ax, title):
    ax.set_title(title)
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.set_xlim(-0.1, 3.1)
    ax.set_ylim(-0.15, 2.65)
    ax.set_aspect("equal")


def configure_mesh_axes(ax, title):
    ax.set_title(title)
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.set_xlim(0, 5)
    ax.set_ylim(0, 4)
    ax.set_aspect("equal")


def masked_mesh_data():
    return np.array(
        [
            [0.15, 0.35, 0.62, 0.48, 0.82],
            [0.30, 0.58, 0.76, 0.52, 0.68],
            [0.46, 0.72, 0.55, 0.28, 0.41],
            [0.22, 0.50, 0.88, 0.65, 0.34],
        ]
    )


def masked_mesh_mask():
    return [
        [False, True, False, False, False],
        [False, False, False, True, False],
        [True, False, False, False, False],
        [False, False, True, False, False],
    ]


def triangulation_gallery(out_dir):
    fig = make_fig_px(1320, 760)
    tri, values = sample_triangulation()

    ax = fig.add_axes(panel(0, 0))
    configure_tri_axes(ax, "Triplot")
    ax.triplot(tri, color=(0.18, 0.24, 0.34, 1.0), linewidth=1.25, label="triplot")

    ax = fig.add_axes(panel(0, 1))
    configure_tri_axes(ax, "Tripcolor + Tricontour")
    ax.tripcolor(
        tri,
        values,
        shading="flat",
        cmap="viridis",
        edgecolors="white",
        linewidth=0.55,
    )
    contour = ax.tricontour(
        tri,
        values,
        levels=6,
        colors=[(0.07, 0.10, 0.16, 0.95)],
        linewidths=1.05,
    )
    ax.clabel(contour, inline=True, fontsize=8, colors=[(0.07, 0.10, 0.16, 0.95)])

    ax = fig.add_axes(panel(1, 0))
    configure_tri_axes(ax, "Tricontourf")
    ax.tricontourf(tri, values, levels=7, cmap="plasma")
    ax.tricontour(
        tri,
        values,
        levels=7,
        colors=[(1, 1, 1, 0.88)],
        linewidths=0.9,
    )

    ax = fig.add_axes(panel(1, 1))
    configure_mesh_axes(ax, "Masked PColorMesh")
    cmap = matplotlib.colormaps["viridis"].copy()
    cmap.set_bad((0.62, 0.62, 0.62, 0.78))
    data = np.ma.array(masked_mesh_data(), mask=masked_mesh_mask())
    ax.pcolormesh(
        [0, 1, 2, 3, 4, 5],
        [0, 1, 2, 3, 4],
        data,
        cmap=cmap,
        edgecolors="white",
        linewidth=0.55,
    )

    fig.text(
        0.98,
        0.975,
        "triangulation gallery\ntriplot, tripcolor, tricontour, tricontourf, masked mesh",
        ha="right",
        va="top",
        fontsize=11,
        bbox=dict(boxstyle="round,pad=0.35", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
    )

    save(fig, out_dir, "triangulation_gallery")


PLOT = triangulation_gallery


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
