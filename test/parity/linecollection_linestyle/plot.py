#!/usr/bin/env python3
"""Matplotlib reference plot for LineCollection string linestyles via hlines."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

# Shared with test/parity/linecollection_linestyle/plot.go.
YS = [1, 2, 3, 4]
LINESTYLES = ["solid", "dashed", "dashdot", "dotted"]


def linecollection_linestyle(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Line Styles")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 5)

    ax.hlines(
        YS,
        0.5,
        9.5,
        colors=(0, 0, 0, 1),
        linestyles=LINESTYLES,
        linewidth=lw(2),
    )
    save(fig, out_dir, "linecollection_linestyle")


PLOT = linecollection_linestyle


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
