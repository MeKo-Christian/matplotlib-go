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


def mplot3d_bar2d_zdir(out_dir):
    fig = make_fig_px(720, 560)
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.88, 0.88), projection="3d")

    xs = np.arange(8)
    heights = [
        [0.9, 0.5, 0.7, 0.3, 0.8, 0.4, 0.6, 0.5],
        [0.4, 0.8, 0.5, 0.9, 0.3, 0.7, 0.5, 0.6],
        [0.7, 0.3, 0.9, 0.4, 0.6, 0.8, 0.4, 0.7],
        [0.5, 0.7, 0.4, 0.8, 0.5, 0.3, 0.9, 0.6],
    ]
    colors = ["r", "g", "b", "y"]
    planes = [3, 2, 1, 0]
    for color, k, ys in zip(colors, planes, heights):
        ax.bar(xs, ys, zs=k, zdir="y", color=color, alpha=0.8)

    ax.set_xlabel("X")
    ax.set_ylabel("Y")
    ax.set_zlabel("Z")

    save(fig, out_dir, "mplot3d_bar2d_zdir")


PLOT = mplot3d_bar2d_zdir


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
