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


def mplot3d_scatter3d(out_dir):
    fig = make_fig_px(720, 560)
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.88, 0.88), projection="3d")

    n = 100
    values = np.array(pcg_float64_values(19680801, 0, 3 * n))
    xs = 23 + values[:n] * (32 - 23)
    ys = values[n : 2 * n] * 100
    zs = -50 + values[2 * n :] * 25

    ax.scatter(xs, ys, zs)

    save(fig, out_dir, "mplot3d_scatter3d")


PLOT = mplot3d_scatter3d


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
