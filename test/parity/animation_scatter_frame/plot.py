#!/usr/bin/env python3
"""Matplotlib reference plot for the deterministic animated-scatter frame fixture.

Reproduces frame GoldenFrame (8) of the orbiting scatter/collection animation
from examples/animation_gallery (frames.go). The Go side renders the same
closed-form frame statically, so this verifies frame-N parity.
"""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import numpy as np

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

FRAME = 8
N = 24


def animation_scatter_frame(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.10, 0.15, 0.93, 0.88))
    ax.set_title("Animated Scatter")
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.set_xlim(-1.1, 1.1)
    ax.set_ylim(-1.1, 1.1)

    phase = FRAME * 0.20
    xs, ys, scalars, sizes = [], [], [], []
    for i in range(N):
        base = 2 * np.pi * i / N
        r = 0.30 + 0.65 * ((i * 7) % N) / N
        theta = base + phase
        xs.append(r * np.cos(theta))
        ys.append(r * np.sin(theta))
        scalars.append(r)
        sizes.append(40 + 200 * r)

    ax.scatter(
        xs,
        ys,
        c=scalars,
        cmap="viridis",
        vmin=0.30,
        vmax=0.95,
        s=sizes,
        edgecolors=[(0.12, 0.12, 0.14, 1)],
        linewidths=1.2,
    )

    save(fig, out_dir, "animation_scatter_frame")


PLOT = animation_scatter_frame


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
