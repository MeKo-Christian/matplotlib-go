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

def stat_variants(out_dir):
    fig = make_fig_px(840, 620)

    axes = {
        "stackplot": fig.add_axes(go_rect(0.08, 0.585, 0.475, 0.93)),
        "ecdf": fig.add_axes(go_rect(0.575, 0.585, 0.97, 0.93)),
        "cumulative": fig.add_axes(go_rect(0.08, 0.10, 0.475, 0.445)),
        "multi": fig.add_axes(go_rect(0.575, 0.10, 0.97, 0.445)),
    }

    stack_ax = axes["stackplot"]
    stack_ax.set_title("StackPlot")
    stack_ax.set_xlim(0, 5)
    stack_ax.set_ylim(0, 7)
    stack_ax.grid(axis="y")
    stack_ax.set_axisbelow(True)
    stack_ax.stackplot(
        [0, 1, 2, 3, 4, 5],
        [1.0, 1.4, 1.3, 1.8, 1.6, 2.0],
        [0.8, 1.1, 1.4, 1.2, 1.6, 1.8],
        [0.5, 0.8, 1.0, 1.4, 1.1, 1.5],
        colors=[
            (0.20, 0.55, 0.75, 0.76),
            (0.90, 0.48, 0.18, 0.76),
            (0.35, 0.66, 0.42, 0.76),
        ],
    )

    ecdf_ax = axes["ecdf"]
    ecdf_ax.set_title("ECDF")
    ecdf_ax.set_xlim(0, 8)
    ecdf_ax.set_ylim(0, 1.05)
    ecdf_ax.grid(axis="y")
    ecdf_ax.set_axisbelow(True)
    ecdf_ax.ecdf(
        [1.2, 1.8, 2.0, 2.0, 3.1, 3.7, 4.3, 5.0, 5.8, 6.6, 7.0],
        compress=True,
        color=(0.18, 0.36, 0.75, 1),
        linewidth=2,
    )

    cumulative_ax = axes["cumulative"]
    cumulative_ax.set_title("Cumulative Step Hist")
    cumulative_ax.set_xlim(0, 6)
    cumulative_ax.set_ylim(0, 1.05)
    cumulative_ax.grid(axis="y")
    cumulative_ax.set_axisbelow(True)
    cumulative_data = [0.4, 0.7, 1.2, 1.4, 2.1, 2.6, 3.1, 3.2, 4.0, 4.8, 5.2]
    cumulative_ax.hist(
        cumulative_data,
        bins=[0, 1, 2, 3, 4, 5, 6],
        weights=np.ones(len(cumulative_data)) / len(cumulative_data),
        cumulative=True,
        histtype="stepfilled",
        facecolor=(0.42, 0.62, 0.90, 0.55),
        edgecolor=(0.12, 0.25, 0.55, 1),
        linewidth=1.4,
    )

    multi_ax = axes["multi"]
    multi_ax.set_title("Stacked Multi-Hist")
    multi_ax.set_xlim(0, 6)
    multi_ax.set_ylim(0, 6)
    multi_ax.grid(axis="y")
    multi_ax.set_axisbelow(True)
    multi_ax.hist(
        [
            [0.3, 0.8, 1.2, 1.7, 2.6, 3.4, 4.1, 5.2],
            [0.5, 1.1, 1.9, 2.3, 2.8, 3.0, 3.7, 4.5, 5.0],
            [1.0, 1.6, 2.2, 2.9, 3.5, 4.2, 4.8, 5.4],
        ],
        bins=[0, 1, 2, 3, 4, 5, 6],
        stacked=True,
        color=[
            (0.22, 0.55, 0.70, 0.8),
            (0.86, 0.42, 0.19, 0.8),
            (0.36, 0.62, 0.36, 0.8),
        ],
        edgecolor=(0.10, 0.10, 0.10, 1),
        linewidth=0.7,
    )

    save(fig, out_dir, "stat_variants")

PLOT = stat_variants


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
