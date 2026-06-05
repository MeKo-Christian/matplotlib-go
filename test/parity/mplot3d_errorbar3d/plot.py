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


def mplot3d_errorbar3d(out_dir):
    fig = make_fig_px(720, 560)
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.88, 0.88), projection="3d")

    x = np.array([-0.8, -0.25, 0.35, 0.9])
    y = np.array([0.1, 0.65, -0.25, 0.35])
    z = np.array([0.2, 0.85, 0.45, 1.1])
    xerr = np.array([0.12, 0.08, 0.16, 0.10])
    yerr = np.array([0.10, 0.14, 0.09, 0.13])
    zerr = np.array([0.18, 0.12, 0.16, 0.10])
    ax.errorbar(x, y, z, xerr=xerr, yerr=yerr, zerr=zerr, fmt="none", color="tab:blue")

    save(fig, out_dir, "mplot3d_errorbar3d")


PLOT = mplot3d_errorbar3d


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
