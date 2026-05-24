#!/usr/bin/env python3
"""Matplotlib reference plot for path-effect semantics."""

from __future__ import annotations

from pathlib import Path
import argparse
import math
import sys

import matplotlib.patheffects as pe
import matplotlib.patches as mpatches

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def _panel(fig, x0, y0, x1, y1):
    ax = fig.add_axes(go_rect(x0, y0, x1, y1))
    ax.set_xlim(0, 1)
    ax.set_ylim(0, 1)
    ax.axis("off")
    return ax


def _offset(dx_px, dy_px):
    return (dx_px * 72.0 / DPI, -dy_px * 72.0 / DPI)


def path_effects(out_dir):
    fig = make_fig()

    text_ax = _panel(fig, 0.06, 0.55, 0.47, 0.93)
    text_ax.text(
        0.5,
        0.52,
        "Path FX",
        ha="center",
        va="center",
        fontsize=26,
        color=(0.08, 0.18, 0.34, 1),
        fontfamily="DejaVu Sans",
        path_effects=[
            pe.SimplePatchShadow(offset=_offset(4, 5), shadow_rgbFace=(0.02, 0.03, 0.04, 0.75), alpha=0.55, rho=0.25),
            pe.Normal(),
        ],
    )

    line_ax = _panel(fig, 0.53, 0.55, 0.94, 0.93)
    xs = [0.08 + 0.84 * i / 48 for i in range(49)]
    ys = [0.50 + 0.24 * math.sin(i / 48 * math.pi * 2.35) for i in range(49)]
    line_ax.plot(
        xs,
        ys,
        color=(0.08, 0.34, 0.66, 1),
        linewidth=lw(3),
        solid_capstyle="butt",
        solid_joinstyle="round",
        path_effects=[pe.withStroke(linewidth=lw(10), foreground=(1, 1, 1, 0.96))],
    )

    scatter_ax = _panel(fig, 0.06, 0.08, 0.47, 0.47)
    scatter_ax.scatter(
        [0.22, 0.50, 0.78, 0.68],
        [0.68, 0.34, 0.70, 0.30],
        s=560,
        c=[
            (0.89, 0.22, 0.24, 1),
            (0.10, 0.55, 0.38, 1),
            (0.13, 0.35, 0.72, 1),
            (0.94, 0.64, 0.18, 1),
        ],
        edgecolors=(0.04, 0.06, 0.08, 1),
        linewidths=lw(1.4),
        path_effects=[
            pe.SimplePatchShadow(offset=_offset(4, 5), shadow_rgbFace=(0.02, 0.03, 0.04, 0.70), alpha=0.5, rho=0.3),
            pe.Normal(),
        ],
    )

    polygon_ax = _panel(fig, 0.53, 0.08, 0.94, 0.47)
    polygon = mpatches.Polygon(
        [(0.12, 0.20), (0.32, 0.82), (0.68, 0.76), (0.90, 0.34), (0.60, 0.12)],
        closed=True,
        facecolor=(0.94, 0.77, 0.28, 1),
        edgecolor=(0.07, 0.20, 0.38, 1),
        linewidth=lw(2.2),
        joinstyle="round",
        path_effects=[
            pe.SimplePatchShadow(offset=_offset(5, 6), shadow_rgbFace=(0.02, 0.03, 0.04, 0.70), alpha=0.45, rho=0.35),
            pe.PathPatchEffect(facecolor=(0.95, 0.92, 0.82, 0.75), edgecolor=(0.83, 0.20, 0.19, 1), linewidth=lw(5.5), joinstyle="round"),
            pe.Normal(),
        ],
    )
    polygon_ax.add_patch(polygon)

    save(fig, out_dir, "path_effects")


PLOT = path_effects


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
