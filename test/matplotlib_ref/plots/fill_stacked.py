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

def fill_stacked(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Stacked Fills")
    ax.set_xlim(0, 8)
    ax.set_ylim(0, 8)
    x      = [1, 2, 3, 4, 5, 6, 7]
    layer1 = [1, 1.5, 2, 1.8, 2.2, 1.9, 1.6]
    layer2 = [layer1[i] + 1.5 + 0.3 * math.sin(i) for i in range(len(layer1))]
    layer3 = [layer2[i] + 1.2 + 0.4 * math.cos(i) for i in range(len(layer2))]
    ax.fill_between(x, 0,      layer1, facecolor=(0.8, 0.2, 0.2, 0.8), edgecolor=(0.5, 0,   0,   1), linewidth=1)
    ax.fill_between(x, layer1, layer2, facecolor=(0.2, 0.8, 0.2, 0.8), edgecolor=(0,   0.5, 0,   1), linewidth=1)
    ax.fill_between(x, layer2, layer3, facecolor=(0.2, 0.2, 0.8, 0.8), edgecolor=(0,   0,   0.5, 1), linewidth=1)
    save(fig, out_dir, "fill_stacked")

PLOT = fill_stacked


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
