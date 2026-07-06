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

def bar_yerr(out_dir):
    """Categorical bars with symmetric y error bars (bar(yerr=…, capsize=…))."""
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Bars with Error Bars")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 12)

    x = [1, 2, 3, 4, 5]
    heights = [3, 7, 2, 8, 5]
    yerr = [0.8, 1.2, 0.5, 1.5, 0.9]

    ax.bar(
        x,
        heights,
        width=0.6,
        color=(0.2, 0.6, 0.8),
        yerr=yerr,
        ecolor=(0, 0, 0, 1),
        capsize=5,
        error_kw={"elinewidth": 1.2},
    )
    save(fig, out_dir, "bar_yerr")

PLOT = bar_yerr


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
