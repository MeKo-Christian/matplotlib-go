#!/usr/bin/env python3
"""Matplotlib reference plot for the deterministic two-panel animation fixture.

Reproduces frame GoldenFrame (8) of the line + heatmap subplot composition from
examples/animation_gallery (frames.go). The Go side renders the same closed-form
frame statically, so this verifies frame-N parity.
"""

from __future__ import annotations

from pathlib import Path
import argparse
import math
import sys

import numpy as np

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

FRAME = 8
ROWS, COLS = 36, 56


def ripple_matrix(frame):
    cx = (COLS - 1) / 2
    cy = (ROWS - 1) / 2
    t = frame * 0.40
    z = np.empty((ROWS, COLS))
    for j in range(ROWS):
        for i in range(COLS):
            d = math.hypot(i - cx, j - cy)
            z[j, i] = math.sin(d * 0.5 - t)
    return z


def animation_subplots_frame(out_dir):
    fig = make_fig_px(800, 360)

    line_ax = fig.add_axes(go_rect(0.07, 0.15, 0.45, 0.88))
    line_ax.set_title("Animated Line")
    line_ax.set_xlabel("phase")
    line_ax.set_ylabel("signal")
    line_ax.set_xlim(0, 2 * np.pi)
    line_ax.set_ylim(-1.2, 1.2)
    line_ax.grid(axis="y")

    phase = FRAME * 0.30
    x = np.linspace(0, 2 * np.pi, 200)
    line_ax.plot(x, np.sin(x + phase), color=TAB10[0], linewidth=lw(2.0))
    line_ax.plot(x, 0.6 * np.cos(x + phase), color=TAB10[1], linewidth=lw(2.0))

    heat_ax = fig.add_axes(go_rect(0.55, 0.15, 0.93, 0.88))
    heat_ax.set_title("Animated Heatmap")
    heat_ax.set_xlabel("column")
    heat_ax.set_ylabel("row")
    heat_ax.imshow(
        ripple_matrix(FRAME),
        cmap="viridis",
        vmin=-1,
        vmax=1,
        origin="lower",
        extent=[0, COLS, 0, ROWS],
        aspect="auto",
    )
    heat_ax.set_xlim(0, COLS)
    heat_ax.set_ylim(0, ROWS)

    save(fig, out_dir, "animation_subplots_frame")


PLOT = animation_subplots_frame


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
