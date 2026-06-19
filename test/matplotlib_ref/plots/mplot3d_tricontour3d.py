#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

from mpl_toolkits.mplot3d import Axes3D  # noqa: F401

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def mplot3d_tricontour3d(out_dir):
    fig = make_fig_px(720, 560)
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.88, 0.88), projection="3d")

    n_angles = 48
    n_radii = 8
    min_radius = 0.25

    radii = np.linspace(min_radius, 0.95, n_radii)
    angles = np.linspace(0, 2 * math.pi, n_angles, endpoint=False)
    angles = np.repeat(angles[..., np.newaxis], n_radii, axis=1)
    angles[:, 1::2] += math.pi / n_angles

    x = (radii * np.cos(angles)).flatten()
    y = (radii * np.sin(angles)).flatten()
    z = (np.cos(radii) * np.cos(3 * angles)).flatten()

    triang = mtri.Triangulation(x, y)
    triang.set_mask(
        np.hypot(x[triang.triangles].mean(axis=1), y[triang.triangles].mean(axis=1))
        < min_radius
    )

    ax.tricontour(triang, z, cmap="CMRmap")
    ax.view_init(elev=45.0)

    save(fig, out_dir, "mplot3d_tricontour3d")


PLOT = mplot3d_tricontour3d


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
