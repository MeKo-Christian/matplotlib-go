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

def arrays_showcase(out_dir):
    fig = make_fig_px(1240, 620)

    annotated = np.array([
        [0.12, 0.28, 0.46, 0.64, 0.82],
        [0.18, 0.34, 0.58, 0.74, 0.88],
        [0.24, 0.42, 0.63, 0.79, 0.91],
        [0.16, 0.38, 0.61, 0.83, 0.97],
    ])

    ax_heat = fig.add_axes(go_rect(0.05, 0.14, 0.31, 0.88))
    ax_heat.set_title("Annotated Heatmap")
    ax_heat.set_xlabel("column")
    ax_heat.set_ylabel("row")
    img = ax_heat.imshow(annotated, cmap="viridis", aspect="equal", origin="upper")
    ax_heat.set_xticks(np.arange(annotated.shape[1]))
    ax_heat.set_yticks(np.arange(annotated.shape[0]))
    ax_heat.xaxis.tick_top()
    ax_heat.xaxis.set_label_position("top")
    threshold = (annotated.min() + annotated.max()) / 2.0
    for row in range(annotated.shape[0]):
        for col in range(annotated.shape[1]):
            color = (1.0, 1.0, 1.0, 1.0) if annotated[row, col] >= threshold else (0.12, 0.12, 0.14, 1.0)
            ax_heat.text(col, row, f"{annotated[row, col]:.2f}", ha="center", va="center", fontsize=10, color=color)

    rows, cols = 8, 10
    xx = np.linspace(0.0, 1.0, cols)
    yy = np.linspace(0.0, 1.0, rows)
    mesh_data = np.zeros((rows, cols))
    for y in range(rows):
        for x in range(cols):
            mesh_data[y, x] = 0.55 + 0.25 * math.sin((xx[x] * 2.3 + 0.35) * math.pi) + 0.20 * math.cos((yy[y] * 2.8 - 0.35 * 0.4) * math.pi)

    ax_mesh = fig.add_axes(go_rect(0.37, 0.14, 0.63, 0.88))
    ax_mesh.set_title("PColorMesh + Contour")
    ax_mesh.set_xlabel("x bin")
    ax_mesh.set_ylabel("y bin")
    x_edges = np.arange(cols + 1)
    y_edges = np.arange(rows + 1)
    quad = ax_mesh.pcolormesh(
        x_edges,
        y_edges,
        mesh_data,
        cmap="plasma",
        edgecolors="white",
        linewidth=0.65,
        shading="flat",
    )
    contour = ax_mesh.contour(
        np.arange(cols),
        np.arange(rows),
        mesh_data,
        levels=6,
        colors=[(0.14, 0.10, 0.16, 0.95)],
        linewidths=1.1,
    )
    ax_mesh.clabel(contour, inline=True, fmt="%.3g", fontsize=10, colors=[(0.14, 0.10, 0.16, 0.95)])

    spy = np.zeros((18, 18))
    for y in range(18):
        for x in range(18):
            if x == y or x + y == 17 or (x + 2 * y) % 7 == 0 or (2 * x + y) % 11 == 0:
                spy[y, x] = 1

    ax_spy = fig.add_axes(go_rect(0.69, 0.14, 0.95, 0.88))
    ax_spy.set_title("Spy")
    ax_spy.set_xlabel("column")
    ax_spy.set_ylabel("row")
    ax_spy.spy(spy, precision=0.1, marker="s", markersize=10, color=(0.16, 0.38, 0.72, 1.0))
    ax_spy.text(
        0.98,
        0.02,
        "sparse structure view",
        transform=ax_spy.transAxes,
        ha="right",
        va="bottom",
        fontsize=10,
        bbox=dict(boxstyle="round,pad=0.3", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
    )

    fig.text(
        0.98,
        0.98,
        "arrays gallery family\nheatmap, quad mesh, sparse matrix",
        ha="right",
        va="top",
        fontsize=11,
        bbox=dict(boxstyle="round,pad=0.35", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
    )

    save(fig, out_dir, "arrays_showcase")

PLOT = arrays_showcase


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
