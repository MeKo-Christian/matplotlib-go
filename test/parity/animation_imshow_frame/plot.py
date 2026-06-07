#!/usr/bin/env python3
"""Matplotlib reference plot for the deterministic animated-heatmap frame fixture.

Reproduces frame GoldenFrame (8) of the ripple imshow animation from
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


def animation_imshow_frame(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.10, 0.15, 0.93, 0.88))
    ax.set_title("Animated Heatmap")
    ax.set_xlabel("column")
    ax.set_ylabel("row")

    ax.imshow(
        ripple_matrix(FRAME),
        cmap="viridis",
        vmin=-1,
        vmax=1,
        origin="lower",
        extent=[0, COLS, 0, ROWS],
        aspect="auto",
    )
    ax.set_xlim(0, COLS)
    ax.set_ylim(0, ROWS)

    save(fig, out_dir, "animation_imshow_frame")


PLOT = animation_imshow_frame


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
