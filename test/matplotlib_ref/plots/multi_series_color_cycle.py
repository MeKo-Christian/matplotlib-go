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

def multi_series_color_cycle(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Color Cycle")
    ax.set_xlim(0, 2 * math.pi)
    ax.set_ylim(-1.2, 1.2)
    n = 50
    x = [2 * math.pi * i / (n - 1) for i in range(n)]
    for i, freq in enumerate([1, 2, 3, 4]):
        y = [math.sin(freq * v) for v in x]
        ax.plot(x, y, color=TAB10[i], linewidth=2)
    save(fig, out_dir, "multi_series_color_cycle")

PLOT = multi_series_color_cycle


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
