#!/usr/bin/env python3
"""Matplotlib reference plot for boxplot default (patch_artist=False) styling."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def boxplot_default(out_dir):
    Path(out_dir).mkdir(parents=True, exist_ok=True)

    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 10)
    ax.set_title("Box Plots (default styling)")
    ax.set_xlabel("Group")
    ax.set_ylabel("Value")

    datasets = [
        [0.9, 1.0, 1.1, 1.2, 1.3, 1.45, 1.5, 1.7, 1.8],
        [4.0, 4.2, 4.3, 4.5, 4.8, 5.0, 5.4, 5.8, 9.4],
        [2.0, 2.1, 2.1, 2.2, 2.3, 2.4, 2.4, 2.6, 3.8],
    ]
    positions = [1.0, 2.0, 3.0]

    # Pure Matplotlib defaults: patch_artist=False (unfilled boxes), C1 medians,
    # black whiskers/caps, and unfilled-circle fliers (markersize 6).
    ax.boxplot(
        datasets,
        positions=positions,
        widths=0.55,
        manage_ticks=False,
    )
    ax.set_axisbelow(True)
    ax.yaxis.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)

    save(fig, out_dir, "boxplot_default")


PLOT = boxplot_default


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
