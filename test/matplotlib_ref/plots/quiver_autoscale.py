#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import math
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def quiver_autoscale(out_dir):
    """Default-scale quiver matching examples/quiver_autoscale.

    No scale= is passed, so matplotlib's default autoscale applies
    (scale = 1.8 * amean * sn / span, sn = max(10, sqrt(N)), span = 1 for the
    default units='width'). The cos(0.5x)/sin(0.5y) grid mirrors the Go side.
    """
    xs = []
    ys = []
    us = []
    vs = []
    for iy in range(7):
        for ix in range(9):
            xs.append(float(ix))
            ys.append(float(iy))
            us.append(math.cos(0.5 * ix))
            vs.append(math.sin(0.5 * iy))

    fig = make_fig()
    ax = fig.add_axes(go_rect(0.10, 0.12, 0.95, 0.90))
    ax.set_title("Quiver Autoscale (default scale)")
    ax.set_xlim(-0.5, 8.5)
    ax.set_ylim(-0.5, 6.5)
    ax.quiver(xs, ys, us, vs, color=(0.15, 0.35, 0.65))
    save(fig, out_dir, "quiver_autoscale")


PLOT = quiver_autoscale


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
