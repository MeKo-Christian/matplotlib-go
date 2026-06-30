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

def multi_series_basic(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Multi-Series Plot")
    ax.set_xlim(0, 8)
    ax.set_ylim(0, 6)
    ax.plot([1, 2, 3, 4, 5, 6], [1.5, 2.8, 2.2, 3.5, 3.8, 4.2], color=TAB10[0], linewidth=2)
    ax.scatter([1.5, 2.5, 3.5, 4.5, 5.5], [2.2, 3.1, 2.9, 4.1, 4.5],
               s=ss(8), c=[TAB10[1]], linewidths=0)
    ax.bar([2, 3, 4, 5], [3.8, 2.5, 4.8, 3.2], width=0.4, color=TAB10[2])
    save(fig, out_dir, "multi_series_basic")

PLOT = multi_series_basic


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
