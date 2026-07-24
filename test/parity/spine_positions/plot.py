#!/usr/bin/env python3
"""Axes-fraction and outward-point spine-position parity fixture."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2]))
    from matplotlib_ref.common import *  # noqa: F401,F403


def configure_axes(ax, title):
    ax.set_title(title)
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.set_xlim(-2, 2)
    ax.set_ylim(-2, 2)
    ax.set_xticks([-2, -1, 0, 1, 2])
    ax.set_yticks([-2, -1, 0, 1, 2])
    ax.spines["top"].set_visible(False)
    ax.spines["right"].set_visible(False)
    ax.plot(
        [-2, -1, 0, 1, 2],
        [-1.4, -0.4, 0.1, 0.9, 1.6],
        color=(0.12, 0.47, 0.71, 1.0),
        linewidth=2.0,
    )


def spine_positions(out_dir):
    fig = make_fig_px(720, 400)

    centered = fig.add_axes(go_rect(0.08, 0.16, 0.47, 0.84))
    configure_axes(centered, "axes = 0.5")
    centered.spines["bottom"].set_position(("axes", 0.5))
    centered.spines["left"].set_position(("axes", 0.5))

    outward = fig.add_axes(go_rect(0.58, 0.16, 0.97, 0.84))
    configure_axes(outward, "outward = 10 pt")
    outward.spines["bottom"].set_position(("outward", 10))
    outward.spines["left"].set_position(("outward", 10))

    save(fig, out_dir, "spine_positions")


PLOT = spine_positions


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
