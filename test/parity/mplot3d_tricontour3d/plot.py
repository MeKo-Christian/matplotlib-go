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

    # Same polar fan point cloud as mplot3d_trisurf3d; rely on auto-Delaunay so
    # the triangulation matches the Go core.Triangulation auto-mesh.
    n_radii = 8
    n_angles = 36
    radii = np.linspace(0.125, 1.0, n_radii)
    angles = np.linspace(0, 2 * math.pi, n_angles, endpoint=False)[:, np.newaxis]

    x = np.append(0, (radii * np.cos(angles)).flatten())
    y = np.append(0, (radii * np.sin(angles)).flatten())
    z = np.sin(-(x * y))

    ax.tricontour(
        x,
        y,
        z,
        levels=[-0.45, -0.3, -0.15, 0, 0.15, 0.3, 0.45],
        cmap="CMRmap",
        vmin=-0.45,
        vmax=0.45,
    )

    save(fig, out_dir, "mplot3d_tricontour3d")


PLOT = mplot3d_tricontour3d


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
