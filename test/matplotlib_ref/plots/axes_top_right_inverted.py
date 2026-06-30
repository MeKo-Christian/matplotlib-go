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

def axes_top_right_inverted(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.12, 0.9, 0.9))
    ax.set_title("Top/Right Axes + Inversion")
    ax.set_xlabel("Bottom X")
    ax.set_ylabel("Left Y")
    ax.set_xlim(10, 0)
    ax.set_ylim(10, 0)
    ax.tick_params(top=True, labeltop=True, right=True, labelright=True)
    ax.minorticks_off()

    ax.plot(
        [1, 3, 6, 8.5],
        [2, 4, 6.5, 8],
        color=(0.15, 0.35, 0.75),
        linewidth=2.2,
    )
    ax.scatter(
        [2, 5, 8],
        [8, 5, 2],
        s=ss(9),
        marker="D",
        c=[(0.85, 0.35, 0.20, 0.9)],
        edgecolors=[(0.45, 0.15, 0.05, 1.0)],
        linewidths=1.0,
    )
    save(fig, out_dir, "axes_top_right_inverted")

PLOT = axes_top_right_inverted


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
