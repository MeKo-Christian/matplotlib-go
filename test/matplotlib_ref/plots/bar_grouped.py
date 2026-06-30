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

def bar_grouped(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Grouped Bars")
    ax.set_xlim(0, 7)
    ax.set_ylim(0, 10)
    ax.bar(
        [1.2, 2.2, 3.2, 4.2, 5.2], [3, 7, 2, 8, 5],
        width=0.35, color=(0.8, 0.2, 0.2), edgecolor=(0.5, 0, 0),
        linewidth=1,
    )
    ax.bar(
        [1.8, 2.8, 3.8, 4.8, 5.8], [5, 4, 6, 3, 7],
        width=0.35, color=(0.2, 0.8, 0.2), edgecolor=(0, 0.5, 0),
        linewidth=1,
    )
    save(fig, out_dir, "bar_grouped")

PLOT = bar_grouped


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
