#!/usr/bin/env python3
"""Triangulation web demo reference module."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import matplotlib.tri as mtri

try:
    from test.matplotlib_ref.webdemo_common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from webdemo_common import *  # noqa: F401,F403


def demo_triangulation(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(1, 3, 0.07, 0.94, 0.14, 0.86, 0.08, 0.0)
    rng = np.random.default_rng(42)
    x = rng.uniform(-2, 2, 90)
    y = rng.uniform(-2, 2, 90)
    z = np.sin(x * 1.8) * np.cos(y * 1.4)
    tri = mtri.Triangulation(x, y)

    ax = fig.add_axes(rects[0][0])
    ax.set_title("Triplot")
    ax.triplot(tri, color=color(0.24, 0.24, 0.28), linewidth=lw(0.8))
    ax.scatter(x, y, s=ss(2.5), color=color(0.16, 0.42, 0.82))

    ax = fig.add_axes(rects[0][1])
    ax.set_title("Tripcolor")
    im = ax.tripcolor(tri, z, shading="gouraud", cmap="viridis")
    fig.colorbar(im, ax=ax)

    ax = fig.add_axes(rects[0][2])
    ax.set_title("TriContour")
    ax.tricontour(tri, z, colors=[color(0.16, 0.42, 0.82)], linewidths=lw(1.0))
    ax.tricontourf(tri, z, levels=8, cmap="magma", alpha=0.85)

    save(fig, out_dir, "triangulation")


DEMO = demo_triangulation


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
