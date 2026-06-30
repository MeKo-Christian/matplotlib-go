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


def formatter_fixed_null_labels(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.12, 0.18, 0.92, 0.88))
    ax.set_title("Fixed + Null Formatters")
    ax.set_xlabel("stage")
    ax.set_ylabel("hidden values")
    ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=0.5)
    ax.set_axisbelow(True)

    x = [0, 1, 2, 3]
    y = [0.20, 0.62, 0.44, 0.82]
    ax.plot(x, y, color=(0.12, 0.47, 0.71), linewidth=2.0)
    ax.set_xlim(-0.25, 3.25)
    ax.set_ylim(0, 1)
    ax.xaxis.set_major_locator(mticker.FixedLocator(x))
    ax.xaxis.set_major_formatter(mticker.FixedFormatter(["draft", "review", "ship", ""]))
    ax.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))
    ax.yaxis.set_major_formatter(mticker.NullFormatter())

    save(fig, out_dir, "formatter_fixed_null_labels")


PLOT = formatter_fixed_null_labels


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
