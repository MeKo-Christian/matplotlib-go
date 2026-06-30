#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

GRID_N = 41


def _dipole_field():
    """z = x * exp(-x^2 - y^2) over [-3, 3]^2, matching examples/contour_styles."""
    t = np.array([-3.0 + 6.0 * i / (GRID_N - 1) for i in range(GRID_N)], dtype=float)
    xs = t
    ys = t
    xx, yy = np.meshgrid(xs, ys)
    z = xx * np.exp(-xx * xx - yy * yy)
    return xs, ys, z


def contour_styles(out_dir):
    fig = make_fig_px(640, 360)
    xs, ys, z = _dipole_field()

    # Left: monochrome contour lines. Negative levels dash by default
    # (rcParams["contour.negative_linestyle"]).
    line_ax = fig.add_axes(go_rect(0.08, 0.12, 0.46, 0.90))
    line_ax.set_title("Contour: negative dashing")
    line_ax.set_xlim(-3, 3)
    line_ax.set_ylim(-3, 3)
    cs = line_ax.contour(
        xs, ys, z,
        levels=[-0.3, -0.2, -0.1, 0.1, 0.2, 0.3],
        colors=[(0.0, 0.0, 0.0, 1.0)],
        linewidths=1.5,
    )
    line_ax.clabel(
        cs, fmt="%.2f", inline=False,
        manual=[(-0.7, 0.6), (-0.7, 1.21), (0.7, 0.6), (0.7, 1.21)],
    )

    # Right: filled contour with extend="both" (under/over bands) and hatches
    # cycled across the bands.
    fill_ax = fig.add_axes(go_rect(0.56, 0.12, 0.94, 0.90))
    fill_ax.set_title("Contourf: extend + hatches")
    fill_ax.set_xlim(-3, 3)
    fill_ax.set_ylim(-3, 3)
    fill_ax.contourf(
        xs, ys, z,
        levels=[-0.3, -0.2, -0.1, 0, 0.1, 0.2, 0.3],
        extend="both",
        hatches=["//", "\\\\", "xx"],
    )

    save(fig, out_dir, "contour_styles")


PLOT = contour_styles


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
