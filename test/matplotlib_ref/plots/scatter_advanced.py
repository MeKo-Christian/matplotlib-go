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

def scatter_advanced(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Advanced Scatter")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 10)
    xs = [2, 4, 6, 8, 2, 4, 6, 8]
    ys = [2, 4, 6, 8, 8, 6, 4, 2]
    radii = [6, 10, 14, 18, 8, 12, 16, 20]
    fills = [
        (1, 0.5, 0.5), (0.5, 1, 0.5), (0.5, 0.5, 1), (1, 1, 0.5),
        (1, 0.5, 1),   (0.5, 1, 1),   (0.8, 0.8, 0.8), (0.3, 0.3, 0.3),
    ]
    edges = [
        (0.5, 0, 0),   (0, 0.5, 0),   (0, 0, 0.5),   (0.5, 0.5, 0),
        (0.5, 0, 0.5), (0, 0.5, 0.5), (0.4, 0.4, 0.4), (0, 0, 0),
    ]
    ax.scatter(
        xs, ys,
        s=[ss(r) for r in radii],
        c=fills,
        edgecolors=edges,
        linewidths=2,
        alpha=0.8,
    )
    save(fig, out_dir, "scatter_advanced")

PLOT = scatter_advanced


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
