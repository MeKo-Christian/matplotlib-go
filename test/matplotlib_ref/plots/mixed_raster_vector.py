#!/usr/bin/env python3
"""Matplotlib reference for mixed raster/vector output."""

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


def mixed_raster_vector(out_dir):
    fig = make_fig_px(640, 640)
    ax = fig.add_axes(go_rect(0.12, 0.12, 0.88, 0.88), projection="polar")
    ax.set_title("Mixed Raster / Vector")
    ax.set_ylim(0, 1.0)
    ax.xaxis.grid(True, color=(0.82, 0.84, 0.88, 1.0), linewidth=lw(0.8))
    ax.yaxis.grid(True, color=(0.84, 0.86, 0.90, 0.9), linewidth=lw(0.8))

    n = 240
    theta = []
    radius = []
    sizes = []
    colors = []
    edges = []
    for i in range(n):
        t = float(i) / float(n - 1)
        th = 8.0 * math.pi * t + 0.18 * math.sin(11 * t)
        r = 0.12 + 0.98 * t + 0.08 * math.sin(7 * th)
        theta.append(th)
        radius.append(r)
        marker_radius = 3.2 + 2.4 * math.sin(math.pi * t) * math.sin(math.pi * t)
        sizes.append(ss(marker_radius))
        colors.append((0.08 + 0.70 * t, 0.30 + 0.45 * (1 - t), 0.86 - 0.48 * t, 0.56))
        edges.append((0.02, 0.08, 0.18, 0.42))

    ax.scatter(
        theta,
        radius,
        s=sizes,
        c=colors,
        edgecolors=edges,
        linewidths=lw(0.45),
        marker="o",
        label="raster cloud",
        rasterized=True,
    )

    line_theta = np.linspace(0.0, 2.0 * math.pi, 180)
    line_radius = 0.58 + 0.16 * np.cos(5 * line_theta)
    ax.plot(
        line_theta,
        line_radius,
        color=(0.08, 0.16, 0.30, 1),
        linewidth=lw(1.8),
        label="vector line",
    )
    ax.legend()
    save(fig, out_dir, "mixed_raster_vector")


PLOT = mixed_raster_vector


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
