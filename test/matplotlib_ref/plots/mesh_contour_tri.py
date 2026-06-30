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

def mesh_contour_tri(out_dir):
    fig = make_fig_px(980, 620)
    axes = {
        "mesh": fig.add_axes(go_rect(0.07, 0.57, 0.46, 0.93)),
        "contour": fig.add_axes(go_rect(0.57, 0.57, 0.96, 0.93)),
        "hist2d": fig.add_axes(go_rect(0.07, 0.10, 0.46, 0.46)),
        "tri": fig.add_axes(go_rect(0.57, 0.10, 0.96, 0.46)),
    }

    mesh_ax = axes["mesh"]
    mesh_ax.set_title("PColorMesh")
    mesh_ax.set_xlim(0, 4)
    mesh_ax.set_ylim(0, 3)
    data = np.array([
        [0.2, 0.6, 0.3, 0.9],
        [0.4, 0.8, 0.5, 0.7],
        [0.1, 0.3, 0.9, 0.6],
    ])
    xedges = np.array([0, 1, 2, 3, 4], dtype=float)
    yedges = np.array([0, 1, 2, 3], dtype=float)
    mesh_ax.pcolormesh(
        xedges, yedges, data,
        shading="flat",
        edgecolors=[(0.95, 0.95, 0.95, 1.0)],
        linewidth=0.8,
    )

    contour_ax = axes["contour"]
    contour_ax.set_title("Contour + Contourf")
    contour_ax.set_xlim(0, 4)
    contour_ax.set_ylim(0, 4)
    contour_data = np.array([
        [0.0, 0.4, 0.8, 0.4, 0.0],
        [0.2, 0.8, 1.3, 0.8, 0.2],
        [0.3, 1.0, 1.7, 1.0, 0.3],
        [0.2, 0.8, 1.3, 0.8, 0.2],
        [0.0, 0.4, 0.8, 0.4, 0.0],
    ])
    xx = np.arange(5, dtype=float)
    yy = np.arange(5, dtype=float)
    contour_levels = [0.2, 0.6, 1.0, 1.4, 1.8]
    contour_ax.contourf(xx, yy, contour_data, levels=contour_levels)
    lines = contour_ax.contour(
        xx, yy, contour_data,
        levels=[0.4, 0.8, 1.2, 1.6],
        colors=[(0.18, 0.18, 0.18, 1.0)],
        linewidths=1.0,
    )
    contour_ax.clabel(lines, fmt="%g", fontsize=10)

    hist_ax = axes["hist2d"]
    hist_ax.set_title("Hist2D")
    hist_ax.set_xlim(0, 4)
    hist_ax.set_ylim(0, 4)
    hx = [0.4, 0.7, 1.1, 1.4, 1.8, 2.1, 2.3, 2.6, 2.9, 3.2, 3.4, 3.6]
    hy = [0.6, 1.0, 1.2, 1.6, 1.4, 2.0, 2.3, 2.1, 2.8, 3.0, 3.2, 3.4]
    hist_ax.hist2d(
        hx, hy,
        bins=[np.array([0, 1, 2, 3, 4], dtype=float), np.array([0, 1, 2, 3, 4], dtype=float)],
        edgecolor=(0.95, 0.95, 0.95, 1.0),
        linewidth=0.8,
    )

    tri_ax = axes["tri"]
    tri_ax.set_title("Triangulation")
    tri_ax.set_xlim(0, 4)
    tri_ax.set_ylim(0, 4)
    tx = np.array([0.4, 1.6, 3.0, 0.8, 2.1, 3.5], dtype=float)
    ty = np.array([0.5, 0.4, 0.7, 2.2, 2.8, 2.1], dtype=float)
    tri = mtri.Triangulation(tx, ty, triangles=[[0, 1, 3], [1, 4, 3], [1, 2, 4], [2, 5, 4]])
    values = np.array([0.2, 0.8, 1.0, 1.5, 1.1, 0.6], dtype=float)
    tri_ax.tripcolor(tri, values, shading="flat")
    tri_ax.triplot(tri, color=(0.15, 0.15, 0.15), linewidth=1.0)
    tri_ax.tricontour(tri, values, levels=[0.7, 1.1], colors=[(0.98, 0.98, 0.98, 1.0)], linewidths=1.0)

    save(fig, out_dir, "mesh_contour_tri")

PLOT = mesh_contour_tri


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
