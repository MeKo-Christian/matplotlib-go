#!/usr/bin/env python3
"""Matplotlib reference plot for pattern, gradient, and path-effect rendering."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import numpy as np
import matplotlib.patheffects as pe
import matplotlib.patches as mpatches

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


WIDTH, HEIGHT = 640, 360


def _offset(dx_px, dy_px):
    return (dx_px * 72.0 / DPI, -dy_px * 72.0 / DPI)


def _rgba(c0, c1, n):
    lo = np.array(c0, dtype=float)
    hi = np.array(c1, dtype=float)
    t = np.linspace(0, 1, n)
    return lo[None, :] * (1 - t[:, None]) + hi[None, :] * t[:, None]


def _linear_gradient(ax):
    width = 164
    left = _rgba((0.90, 0.16, 0.18, 1), (0.96, 0.78, 0.20, 1), width // 2)
    right = _rgba((0.96, 0.78, 0.20, 1), (0.10, 0.30, 0.78, 1), width - width // 2)
    row = np.concatenate([left, right], axis=0)
    img = np.repeat(row[None, :, :], 100, axis=0)
    ax.imshow(img, extent=(42, 205, 142, 42), interpolation="bilinear")
    ax.add_patch(mpatches.Rectangle((42, 42), 163, 100, facecolor="none", edgecolor=(0.06, 0.08, 0.10), linewidth=1.8))


def _radial_gradient(ax):
    w, h = 164, 100
    y, x = np.mgrid[0:h, 0:w]
    dist = np.sqrt(((x - w / 2) / 82) ** 2 + ((y - h / 2) / 82) ** 2)
    t = np.clip(dist, 0, 1)
    c0 = np.array((0.98, 0.98, 0.86, 1))
    c1 = np.array((0.08, 0.50, 0.36, 1))
    img = c0[None, None, :] * (1 - t[:, :, None]) + c1[None, None, :] * t[:, :, None]
    ax.imshow(img, extent=(235, 398, 142, 42), interpolation="bilinear")
    ax.add_patch(mpatches.Rectangle((235, 42), 163, 100, facecolor="none", edgecolor=(0.06, 0.08, 0.10), linewidth=1.8))


def _pattern(ax):
    poly = mpatches.Polygon(
        [(455, 44), (594, 54), (570, 145), (430, 128)],
        closed=True,
        facecolor=(0.93, 0.94, 0.98, 1),
        edgecolor=(0.06, 0.08, 0.10, 1),
        linewidth=1.8,
        hatch="///",
    )
    ax.add_patch(poly)


def pattern_gradient_effects(out_dir):
    fig = make_fig_px(WIDTH, HEIGHT)
    ax = fig.add_axes((0, 0, 1, 1))
    ax.set_xlim(0, WIDTH)
    ax.set_ylim(HEIGHT, 0)
    ax.axis("off")

    _linear_gradient(ax)
    _radial_gradient(ax)
    _pattern(ax)

    stroked = mpatches.Rectangle(
        (86, 210),
        142,
        88,
        facecolor=(0.12, 0.56, 0.40, 1),
        edgecolor=(0.02, 0.09, 0.16, 1),
        linewidth=2.2,
        joinstyle="round",
        path_effects=[
            pe.Stroke(linewidth=9, foreground=(1, 0.92, 0.58, 0.95), offset=_offset(4, -4)),
            pe.Normal(),
        ],
    )
    ax.add_patch(stroked)

    filtered = mpatches.Rectangle(
        (362, 210),
        144,
        88,
        facecolor=(0.10, 0.28, 0.74, 1),
        edgecolor="none",
        path_effects=[
            pe.SimplePatchShadow(offset=_offset(8, 8), shadow_rgbFace=(0.08, 0.12, 0.24, 0.45), alpha=0.45, rho=0.35),
            pe.Normal(),
        ],
    )
    ax.add_patch(filtered)

    save(fig, out_dir, "pattern_gradient_effects")


PLOT = pattern_gradient_effects


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
