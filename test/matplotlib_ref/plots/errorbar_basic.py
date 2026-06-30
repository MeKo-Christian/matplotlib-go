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

def errorbar_basic(out_dir):
    """Combined line+scatter with symmetric error bars."""
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Error Bars")
    ax.set_xlim(0, 7)
    ax.set_ylim(0, 6)

    x = [1, 2, 3, 4, 5, 6]
    y = [1.8, 2.5, 2.2, 3.1, 2.8, 3.7]
    xerr = [0.20, 0.25, 0.15, 0.22, 0.30, 0.18]
    yerr = [0.28, 0.20, 0.35, 0.24, 0.30, 0.22]

    ax.plot(x, y, color=TAB10[0], linewidth=2)
    ax.scatter(
        x,
        y,
        s=ss(4.5),
        c=[TAB10[2]],
        linewidths=0,
    )
    ax.errorbar(
        x,
        y,
        xerr=xerr,
        yerr=yerr,
        fmt="none",
        ecolor=(0, 0, 0, 1),
        elinewidth=1.2,
        capsize=6,
    )
    save(fig, out_dir, "errorbar_basic")

PLOT = errorbar_basic


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
