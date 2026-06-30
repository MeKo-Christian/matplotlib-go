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


def locator_maxn_edge_labels(out_dir):
    fig = make_fig_px(720, 540)
    axes = [
        fig.add_axes(go_rect(0.12, 0.70, 0.92, 0.92)),
        fig.add_axes(go_rect(0.12, 0.40, 0.92, 0.62)),
        fig.add_axes(go_rect(0.12, 0.10, 0.92, 0.32)),
    ]

    axes[0].set_title("MaxNLocator Degenerate View")
    axes[0].set_xlabel("expanded around 2")
    axes[0].set_ylabel("value")
    axes[0].plot([2, 2], [0.2, 0.9], color=(0.12, 0.47, 0.71), linewidth=2.0)
    axes[0].set_xlim(2, 2)
    axes[0].set_ylim(0, 1)
    axes[0].xaxis.set_major_locator(mticker.MaxNLocator(nbins=2, steps=[1, 2, 2.5, 5, 10]))

    axes[1].set_title("MaxNLocator Prune Both")
    axes[1].set_xlabel("pruned range")
    axes[1].set_ylabel("value")
    axes[1].plot(
        [-3, -1, 1, 3, 5, 7],
        [0.16, 0.28, 0.44, 0.62, 0.78, 0.90],
        color=(0.12, 0.47, 0.71),
        linewidth=2.0,
    )
    axes[1].set_xlim(-3, 7)
    axes[1].set_ylim(0, 1)
    axes[1].xaxis.set_major_locator(mticker.MaxNLocator(nbins=5, prune="both"))

    axes[2].set_title("MaxNLocator Large Offset")
    axes[2].set_xlabel("1e6 + offset")
    axes[2].set_ylabel("value")
    axes[2].plot(
        [1_000_000, 1_000_001, 1_000_002, 1_000_003, 1_000_004],
        [0.14, 0.30, 0.50, 0.72, 0.88],
        color=(0.12, 0.47, 0.71),
        linewidth=2.0,
    )
    axes[2].set_xlim(1_000_000, 1_000_004)
    axes[2].set_ylim(0, 1)
    axes[2].xaxis.set_major_locator(mticker.MaxNLocator(nbins=4))
    axes[2].xaxis.set_major_formatter(mticker.FuncFormatter(lambda v, _pos: f"+{v - 1_000_000:.0f}"))

    for ax in axes:
        ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=0.5)
        ax.set_axisbelow(True)
        ax.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))

    save(fig, out_dir, "locator_maxn_edge_labels")


PLOT = locator_maxn_edge_labels


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
