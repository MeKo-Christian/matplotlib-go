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

def plot_variants(out_dir):
    fig = make_fig_px(840, 620)

    axes = {
        "step": fig.add_axes(go_rect(0.08, 0.585, 0.475, 0.93)),
        "fill": fig.add_axes(go_rect(0.575, 0.585, 0.97, 0.93)),
        "broken": fig.add_axes(go_rect(0.08, 0.10, 0.475, 0.445)),
        "stack": fig.add_axes(go_rect(0.575, 0.10, 0.97, 0.445)),
    }

    step_ax = axes["step"]
    step_ax.set_title("Step + Stairs")
    step_ax.set_xlim(0, 6)
    step_ax.set_ylim(0, 5.2)
    step_ax.grid(axis="y")
    step_ax.set_axisbelow(True)
    step_ax.step(
        [0.6, 1.4, 2.2, 3.0, 3.8, 4.6, 5.4],
        [1.1, 2.5, 1.7, 3.4, 2.9, 4.1, 3.6],
        where="post",
        color=(0.15, 0.39, 0.78),
        linewidth=2.0,
    )
    step_ax.stairs(
        [0.9, 1.7, 1.4, 2.6, 1.8, 2.2],
        [0.4, 1.1, 2.0, 2.9, 3.7, 4.6, 5.5],
        baseline=0.35,
        fill=True,
        facecolor=(0.91, 0.49, 0.20, 0.72),
        edgecolor=(0.58, 0.26, 0.08, 1.0),
        linewidth=1.5,
    )

    fill_ax = axes["fill"]
    fill_ax.set_title("FillBetweenX + Refs")
    fill_ax.set_xlim(0, 7)
    fill_ax.set_ylim(0, 6)
    fill_ax.grid(axis="x")
    fill_ax.set_axisbelow(True)
    fill_ax.fill_betweenx(
        [0.4, 1.2, 2.0, 2.8, 3.6, 4.4, 5.2],
        [1.3, 2.1, 1.7, 2.8, 2.2, 3.1, 2.6],
        [3.4, 4.1, 4.8, 5.1, 5.6, 6.0, 6.3],
        facecolor=(0.24, 0.68, 0.54, 0.72),
        edgecolor=(0.12, 0.38, 0.28, 1.0),
        linewidth=1.2,
    )
    fill_ax.axvspan(2.2, 3.1, color=(0.92, 0.75, 0.18), alpha=0.20)
    fill_ax.axhline(4.0, color=(0.52, 0.18, 0.18), linewidth=1.2, dashes=[4 * 36.0 / DPI, 3 * 36.0 / DPI])
    fill_ax.axvline(5.3, color=(0.18, 0.22, 0.55), linewidth=1.2, dashes=[2 * 36.0 / DPI, 2 * 36.0 / DPI])
    fill_ax.axline((0.9, 0.3), (6.4, 5.6), color=(0.22, 0.22, 0.22), linewidth=1.1)

    broken_ax = axes["broken"]
    broken_ax.set_title("broken_barh")
    broken_ax.set_xlim(0, 10)
    broken_ax.set_ylim(0, 4.4)
    broken_ax.grid(axis="x")
    broken_ax.set_axisbelow(True)
    broken_ax.broken_barh([(0.8, 1.6), (3.1, 2.2), (6.5, 1.3)], (0.7, 0.9), facecolors=(0.21, 0.51, 0.76))
    broken_ax.broken_barh([(1.6, 1.0), (4.0, 1.4), (7.1, 1.7)], (2.1, 0.9), facecolors=(0.86, 0.38, 0.16))
    for x, y, label in [(1.6, 1.15, "prep"), (4.2, 1.15, "run"), (7.15, 1.15, "cool"), (2.1, 2.55, "IO"), (4.7, 2.55, "fit"), (7.95, 2.55, "ship")]:
        broken_ax.text(x, y, label, ha="center", va="center", fontsize=10, color="white")

    stack_ax = axes["stack"]
    stack_ax.set_title("Stacked Bars + Labels")
    stack_ax.set_xlim(0.4, 4.6)
    stack_ax.set_ylim(0, 7.6)
    stack_ax.grid(axis="y")
    stack_ax.set_axisbelow(True)
    xs = [1, 2, 3, 4]
    series_a = [1.4, 2.2, 1.8, 2.5]
    series_b = [2.1, 1.6, 2.4, 1.7]
    bottom = stack_ax.bar(xs, series_a, color=(0.16, 0.59, 0.49), width=0.8)
    top = stack_ax.bar(xs, series_b, bottom=series_a, color=(0.88, 0.47, 0.16), width=0.8)
    stack_ax.bar_label(bottom, labels=["A1", "A2", "A3", "A4"], label_type="center", color="white", fontsize=10)
    stack_ax.bar_label(top, fmt="%.1f", color=(0.20, 0.20, 0.20), fontsize=10, padding=4)

    save(fig, out_dir, "plot_variants")

PLOT = plot_variants


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
