#!/usr/bin/env python3
"""Matplotlib reference for a direct single-series box plot."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def boxplot_single_series(out_dir):
    Path(out_dir).mkdir(parents=True, exist_ok=True)

    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 2)
    ax.set_ylim(0, 10)
    ax.set_title("Single-series box plot")
    ax.set_xlabel("Series")
    ax.set_ylabel("Value")
    ax.boxplot([1.0, 1.4, 1.8, 2.2, 2.8, 3.5, 4.2, 4.8, 8.5], manage_ticks=False)

    save(fig, out_dir, "boxplot_single_series")


PLOT = boxplot_single_series


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
