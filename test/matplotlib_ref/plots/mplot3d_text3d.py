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


def mplot3d_text3d(out_dir):
    fig = make_fig_px(720, 560)
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.88, 0.88), projection="3d")

    ax.set_xlim(0, 10)
    ax.set_ylim(0, 10)
    ax.set_zlim(0, 10)

    labels = [
        (1, 1, 1, "alpha"),
        (8, 2, 3, "beta"),
        (3, 7, 5, "gamma"),
        (6, 6, 8, "delta"),
        (2, 4, 9, "epsilon"),
        (9, 9, 2, "zeta"),
    ]
    for x, y, z, s in labels:
        ax.text(x, y, z, s)

    ax.set_xlabel("X")
    ax.set_ylabel("Y")
    ax.set_zlabel("Z")

    save(fig, out_dir, "mplot3d_text3d")


PLOT = mplot3d_text3d


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
