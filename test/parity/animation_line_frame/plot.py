#!/usr/bin/env python3
"""Matplotlib reference plot for the deterministic animated-line frame fixture.

Reproduces frame GoldenFrame (8) of the traveling-wave FuncAnimation from
examples/animation_gallery (frames.go). The Go side renders the same closed-form
frame statically, so this verifies frame-N parity.
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


def animation_line_frame(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.10, 0.15, 0.93, 0.88))
    ax.set_title("Animated Line")
    ax.set_xlabel("phase")
    ax.set_ylabel("signal")
    ax.set_xlim(0, 2 * np.pi)
    ax.set_ylim(-1.2, 1.2)
    ax.grid(axis="y")

    phase = FRAME * 0.30
    x = np.linspace(0, 2 * np.pi, 200)
    ax.plot(x, np.sin(x + phase), color=TAB10[0], linewidth=lw(2.0))
    ax.plot(x, 0.6 * np.cos(x + phase), color=TAB10[1], linewidth=lw(2.0))

    save(fig, out_dir, "animation_line_frame")


PLOT = animation_line_frame


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
