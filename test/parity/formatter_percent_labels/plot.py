#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import matplotlib.ticker as mticker

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2]))
    from matplotlib_ref.common import *  # noqa: F401,F403


def formatter_percent_labels(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.12, 0.18, 0.92, 0.88))
    ax.set_title("Percent Formatter")
    ax.set_xlabel("share")
    ax.set_ylabel("score")
    ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=0.5)
    ax.set_axisbelow(True)

    x = [-0.5, 0, 0.125, 0.5, 1.0]
    y = [0.15, 0.30, 0.42, 0.68, 0.86]
    ax.plot(x, y, color=(0.12, 0.47, 0.71), linewidth=2.0)
    ax.set_xlim(-0.5, 1.0)
    ax.set_ylim(0, 1.0)
    ax.xaxis.set_major_locator(mticker.FixedLocator(x))
    ax.xaxis.set_major_formatter(mticker.PercentFormatter(xmax=1, decimals=None))
    ax.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))

    save(fig, out_dir, "formatter_percent_labels")


PLOT = formatter_percent_labels


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
