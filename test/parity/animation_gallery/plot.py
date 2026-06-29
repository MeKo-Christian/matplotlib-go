#!/usr/bin/env python3
"""Matplotlib reference for the animation gallery preview."""

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


def _sine_frame(frame):
    x = np.linspace(0, 2 * math.pi, 96)
    y = np.sin(x + frame * 0.35) * (0.75 + 0.02 * frame)
    return x, y


def animation_gallery(out_dir):
    fig = make_fig_px(1000, 560)

    ax = fig.add_axes([0.08, 0.16, 0.40, 0.74])
    ax.set_title("FuncAnimation frames")
    ax.set_xlabel("phase")
    ax.set_ylabel("signal")
    ax.set_xlim(0, 2 * math.pi)
    ax.set_ylim(-1.35, 1.35)
    ax.grid(axis="y")
    for frame in [0, 4, 8]:
        x, y = _sine_frame(frame)
        color = (0.28, 0.28, 0.28, 0.45)
        width = lw(1.2)
        if frame == 8:
            color = (0.12, 0.34, 0.68, 1.0)
            width = lw(2.3)
        ax.plot(x, y, color=color, linewidth=width)

    ax2 = fig.add_axes([0.58, 0.16, 0.36, 0.74])
    ax2.set_title("ArtistAnimation frames")
    ax2.set_xlabel("frame")
    ax2.set_ylabel("value")
    ax2.set_xlim(-0.5, 3.5)
    ax2.set_ylim(0, 3.5)
    ax2.grid(axis="y")
    series = [
        ([(0, 0.7), (1, 1.4), (2, 0.9), (3, 1.8)], (0.84, 0.35, 0.18), "frame 0"),
        ([(0, 1.3), (1, 2.2), (2, 1.6), (3, 2.7)], (0.20, 0.58, 0.34), "frame 1"),
        ([(0, 2.0), (1, 2.8), (2, 2.2), (3, 3.1)], (0.55, 0.32, 0.72), "frame 2"),
    ]
    for points, color, label in series:
        pts = np.array(points, dtype=float)
        ax2.plot(pts[:, 0], pts[:, 1], color=color, linewidth=lw(2.1), label=label)
    ax2.legend(loc="upper left", frameon=True)

    save(fig, out_dir, "animation_gallery")


PLOT = animation_gallery


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
