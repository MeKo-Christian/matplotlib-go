#!/usr/bin/env python3
"""Matplotlib reference for the compact mplot3d gallery."""

from __future__ import annotations

from pathlib import Path
import argparse
import math
import sys

from mpl_toolkits.mplot3d import Axes3D  # noqa: F401
from mpl_toolkits.mplot3d import axes3d

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def radial_surface(count, min_val, max_val):
    x = np.linspace(min_val, max_val, count)
    y = np.linspace(min_val, max_val, count)
    xx, yy = np.meshgrid(x, y)
    z = np.sin(np.hypot(xx, yy))
    return x, y, xx, yy, z


def fan_mesh(n_radii, n_angles):
    x = [0.0]
    y = [0.0]
    z = [0.0]
    for angle_index in range(n_angles):
        angle = 2 * math.pi * angle_index / n_angles
        for radius_index in range(n_radii):
            radius = 0.125 + (radius_index / (n_radii - 1)) * (1.0 - 0.125)
            xv = radius * math.cos(angle)
            yv = radius * math.sin(angle)
            x.append(xv)
            y.append(yv)
            z.append(math.sin(-xv * yv))
    return np.array(x), np.array(y), np.array(z)


def draw_line3d(ax):
    t = np.linspace(0.0, 1.0, 72)
    ax.plot(t, np.sin(6 * math.pi * t), np.cos(6 * math.pi * t), linewidth=1.6)


def draw_scatter3d(ax):
    n = 55
    values = np.array(pcg_float64_values(19680801, 0, 3 * n))
    x = 23 + values[0::3] * 9
    y = values[1::3] * 100
    z = -50 + values[2::3] * 25
    ax.scatter(x, y, z)


def draw_surface(ax):
    x, y, xx, yy, z = radial_surface(28, -4, 4)
    ax.plot_surface(xx, yy, z, cmap="Blues", vmin=2 * z.min(), linewidth=0)


def draw_wireframe(ax):
    x, y, z = axes3d.get_test_data(0.12)
    ax.plot_wireframe(x, y, z, rstride=4, cstride=4)


def draw_trisurf(ax):
    x, y, z = fan_mesh(7, 20)
    ax.plot_trisurf(x, y, z, cmap="viridis", vmin=2 * z.min())


def draw_bar3d(ax):
    ax.bar3d([1, 1, 2, 2], [1, 2, 1, 2], [0, 0, 0, 0], [0.5] * 4, [0.5] * 4, [2, 3, 1, 4])


def draw_voxels(ax):
    x, y, z = np.indices((6, 6, 6))
    voxelarray = ((x < 2) & (y < 2) & (z < 2)) | ((x >= 4) & (y >= 4) & (z >= 4))
    ax.voxels(voxelarray, edgecolor="k")


def draw_quiver3d(ax):
    n = 3
    values = np.linspace(-1, 1, n)
    x, y, z, u, v, w = [], [], [], [], [], []
    for yv in values:
        for xv in values:
            for zv in values:
                x.append(xv)
                y.append(yv)
                z.append(zv)
                u.append((xv + yv) / 5)
                v.append((yv - xv) / 5)
                w.append(0)
    ax.quiver(x, y, z, u, v, w)


def draw_stem3d(ax):
    t = np.linspace(0, 2 * math.pi, 16)
    ax.stem(np.sin(t), np.cos(t), np.linspace(0, 1, len(t)))


def draw_fill_between3d(ax):
    theta = np.linspace(0, 2 * math.pi, 38)
    z = np.linspace(0, 1, len(theta))
    x1 = np.cos(theta)
    y1 = np.sin(theta)
    x2 = np.cos(theta + math.pi)
    y2 = np.sin(theta + math.pi)
    ax.plot(x1, y1, z, linewidth=1.4, color="C0")
    ax.plot(x2, y2, z, linewidth=1.4, color="C0")
    ax.fill_between(x1, y1, z, x2, y2, z, alpha=0.5)


def mplot3d_gallery(out_dir):
    fig = make_fig_px(1320, 840)
    panels = [
        ("3D line", draw_line3d),
        ("3D scatter", draw_scatter3d),
        ("surface", draw_surface),
        ("wireframe", draw_wireframe),
        ("trisurf", draw_trisurf),
        ("bar3d", draw_bar3d),
        ("voxels", draw_voxels),
        ("quiver3d", draw_quiver3d),
        ("stem3d", draw_stem3d),
        ("fill_between3d", draw_fill_between3d),
    ]
    for i, (title, draw) in enumerate(panels):
        row, col = divmod(i, 5)
        left = 0.035 + col * 0.193
        bottom = 0.54 if row == 0 else 0.08
        ax = fig.add_axes(go_rect(left, bottom, left + 0.17, bottom + 0.36), projection="3d")
        ax.set_title(title)
        ax.view_init(elev=30, azim=-60)
        draw(ax)

    fig.text(0.035, 0.975, "mplot3d Gallery", ha="left", va="top", fontsize=13)
    save(fig, out_dir, "mplot3d_gallery")


PLOT = mplot3d_gallery


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
