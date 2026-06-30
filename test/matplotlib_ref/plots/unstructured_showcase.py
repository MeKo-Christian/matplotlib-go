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

def unstructured_showcase(out_dir):
    fig = make_fig_px(1320, 520)

    tri_x = np.array([0.0, 0.85, 1.75, 2.85, 0.2, 1.1, 2.1, 0.55, 1.55, 2.55])
    tri_y = np.array([0.0, 0.2, 0.05, 0.3, 1.0, 1.15, 1.25, 2.15, 2.3, 2.05])
    triangles = np.array([
        [0, 1, 4],
        [1, 5, 4],
        [1, 2, 5],
        [2, 6, 5],
        [2, 3, 6],
        [4, 5, 7],
        [5, 8, 7],
        [5, 6, 8],
        [6, 9, 8],
    ])
    tri = mtri.Triangulation(tri_x, tri_y, triangles)
    values = np.sin(tri_x * 1.4) + 0.7 * np.cos((tri_y + 0.15) * 2.1)

    ax_mesh = fig.add_axes(go_rect(0.05, 0.16, 0.31, 0.88))
    ax_mesh.set_title("Triangulation")
    ax_mesh.set_xlabel("x")
    ax_mesh.set_ylabel("y")
    ax_mesh.set_xlim(-0.1, 3.1)
    ax_mesh.set_ylim(-0.15, 2.65)
    ax_mesh.set_aspect("equal")
    ax_mesh.triplot(tri, color=(0.18, 0.24, 0.34, 1.0), linewidth=1.35)
    ax_mesh.text(
        0.98,
        0.02,
        "explicit triangular mesh",
        transform=ax_mesh.transAxes,
        ha="right",
        va="bottom",
        fontsize=10,
        bbox=dict(boxstyle="round,pad=0.3", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
    )

    ax_color = fig.add_axes(go_rect(0.37, 0.16, 0.63, 0.88))
    ax_color.set_title("Tripcolor + Tricontour")
    ax_color.set_xlabel("x")
    ax_color.set_ylabel("y")
    ax_color.set_xlim(-0.1, 3.1)
    ax_color.set_ylim(-0.15, 2.65)
    ax_color.set_aspect("equal")
    ax_color.tripcolor(
        tri,
        values,
        cmap="viridis",
        edgecolors="white",
        linewidth=0.6,
        shading="flat",
    )
    contour = ax_color.tricontour(
        tri,
        values,
        levels=6,
        colors=[(0.08, 0.12, 0.18, 0.95)],
        linewidths=1.15,
    )
    ax_color.clabel(contour, inline=True, fmt="%.3g", fontsize=10, colors=[(0.08, 0.12, 0.18, 0.95)])

    ax_fill = fig.add_axes(go_rect(0.69, 0.16, 0.95, 0.88))
    ax_fill.set_title("Filled Tricontour")
    ax_fill.set_xlabel("x")
    ax_fill.set_ylabel("y")
    ax_fill.set_xlim(-0.1, 3.1)
    ax_fill.set_ylim(-0.15, 2.65)
    ax_fill.set_aspect("equal")
    ax_fill.tricontourf(tri, values, levels=7, cmap="plasma")
    ax_fill.tricontour(
        tri,
        values,
        levels=7,
        colors=[(1.0, 1.0, 1.0, 0.88)],
        linewidths=0.95,
    )

    fig.text(
        0.98,
        0.98,
        "unstructured gallery family\ntriangulation, tripcolor, tricontour",
        ha="right",
        va="top",
        fontsize=11,
        bbox=dict(boxstyle="round,pad=0.35", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
    )

    save(fig, out_dir, "unstructured_showcase")

PLOT = unstructured_showcase


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
