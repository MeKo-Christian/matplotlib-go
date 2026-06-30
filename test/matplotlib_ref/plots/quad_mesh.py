#!/usr/bin/env python3
"""Matplotlib reference plot for quad mesh batch coverage."""

from __future__ import annotations

from pathlib import Path
import argparse
import math
import sys

import matplotlib.ticker as mticker

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def quad_mesh(out_dir):
    fig = make_fig_px(980, 620)
    ax = fig.add_axes(go_rect(0.10, 0.14, 0.94, 0.88))
    ax.set_title("RendererAgg quad mesh")
    ax.set_xlim(0, 9)
    ax.set_ylim(0, 6)
    ax.xaxis.set_major_locator(mticker.MultipleLocator(1))
    ax.yaxis.set_major_locator(mticker.MultipleLocator(1))

    data = np.empty((6, 9), dtype=float)
    for y in range(data.shape[0]):
        for x in range(data.shape[1]):
            data[y, x] = 0.45 + 0.38 * math.sin(float(x) * 0.7) + 0.22 * math.cos(float(y) * 1.1)
    xedges = np.array([0, 1.1, 1.9, 3.0, 3.7, 4.9, 5.8, 6.7, 7.9, 9.0], dtype=float)
    yedges = np.array([0, 0.8, 1.7, 2.9, 3.6, 4.8, 6.0], dtype=float)
    ax.pcolormesh(
        xedges,
        yedges,
        data,
        shading="flat",
        cmap="viridis",
        vmin=-0.15,
        vmax=1.1,
        edgecolors=[(0.96, 0.96, 0.96, 1)],
        linewidth=0.65,
    )
    save(fig, out_dir, "quad_mesh")


PLOT = quad_mesh


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
