#!/usr/bin/env python3
"""Matplotlib reference plot for large scatter batch coverage."""

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


def large_scatter(out_dir):
    fig = make_fig_px(980, 620)
    ax = fig.add_axes(go_rect(0.09, 0.13, 0.95, 0.88))
    ax.set_title("RendererAgg marker batch")
    ax.set_xlim(-0.5, 14.5)
    ax.set_ylim(-1.5, 11.5)
    ax.grid(True, axis="y")

    xs, ys, sizes, fills, edges = [], [], [], [], []
    for i in range(180):
        x = float(i % 15) + 0.24 * math.sin(float(i) * 0.73)
        y = float((i * 7) % 12) + 0.28 * math.cos(float(i) * 0.41)
        radius = 4.0 + float((i * 11) % 9)
        t = float(i % 30) / 29.0
        xs.append(x)
        ys.append(y)
        sizes.append(ss(radius))
        fills.append((0.12 + 0.70 * t, 0.58 - 0.25 * t, 0.88 - 0.56 * t, 0.72))
        edges.append((0.08, 0.10 + 0.28 * t, 0.18, 0.95))

    ax.scatter(
        xs,
        ys,
        s=sizes,
        c=fills,
        edgecolors=edges,
        linewidths=0.75,
        marker="o",
        label="batched markers",
    )
    ax.legend()
    save(fig, out_dir, "large_scatter")


PLOT = large_scatter


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
