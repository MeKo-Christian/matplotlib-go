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


def mplot3d_contour3d(out_dir):
    fig = make_fig_px(720, 560)
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.88, 0.88), projection="3d")

    # Mirrors mpl_toolkits.mplot3d.axes3d.get_test_data(0.25).
    x = y = np.arange(-3.0, 3.0, 0.25)
    X, Y = np.meshgrid(x, y)
    Z1 = np.exp(-(X**2 + Y**2) / 2) / (2 * np.pi)
    Z2 = np.exp(-(((X - 1) / 1.5) ** 2 + ((Y - 1) / 0.5) ** 2) / 2) / (
        2 * np.pi * 0.5 * 1.5
    )
    Z = (Z2 - Z1) * 500
    X = X * 10
    Y = Y * 10
    ax.contour(
        X,
        Y,
        Z,
        levels=[-50, -25, 0, 25, 50, 75],
        cmap="coolwarm",
        vmin=-50,
        vmax=75,
    )

    save(fig, out_dir, "mplot3d_contour3d")


PLOT = mplot3d_contour3d


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
