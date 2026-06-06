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

def bar_variants(out_dir):
    fig = make_fig_px(840, 620)
    axes = {
        "vertical": fig.add_axes(go_rect(0.07, 0.585, 0.46, 0.93)),
        "horizontal": fig.add_axes(go_rect(0.56, 0.585, 0.96, 0.93)),
        "grouped": fig.add_axes(go_rect(0.07, 0.10, 0.46, 0.445)),
        "stacked": fig.add_axes(go_rect(0.56, 0.10, 0.96, 0.445)),
    }

    # Vertical bars with categorical tick labels.
    v_ax = axes["vertical"]
    v_ax.set_title("Vertical + Tick Labels")
    v_ax.set_xlim(0.4, 5.6)
    v_ax.set_ylim(0, 10)
    v_ax.bar([1, 2, 3, 4, 5], [3, 7, 2, 8, 5], width=0.6, color=(0.20, 0.60, 0.80))
    v_ax.set_xticks([1, 2, 3, 4, 5])
    v_ax.set_xticklabels(["alpha", "beta", "gamma", "delta", "eps"])

    # Horizontal bars.
    h_ax = axes["horizontal"]
    h_ax.set_title("Horizontal Bars")
    h_ax.set_xlim(0, 10)
    h_ax.set_ylim(0, 6)
    h_ax.barh([1, 2, 3, 4, 5], [3, 7, 2, 8, 5], height=0.6, color=(0.8, 0.4, 0.2))

    # Grouped bars.
    g_ax = axes["grouped"]
    g_ax.set_title("Grouped Bars")
    g_ax.set_xlim(0, 7)
    g_ax.set_ylim(0, 10)
    g_ax.bar([1.2, 2.2, 3.2, 4.2, 5.2], [3, 7, 2, 8, 5], width=0.35,
             color=(0.8, 0.2, 0.2), edgecolor=(0.5, 0, 0), linewidth=lw(1))
    g_ax.bar([1.8, 2.8, 3.8, 4.8, 5.8], [5, 4, 6, 3, 7], width=0.35,
             color=(0.2, 0.8, 0.2), edgecolor=(0, 0.5, 0), linewidth=lw(1))

    # Stacked bars with labels.
    s_ax = axes["stacked"]
    s_ax.set_title("Stacked + Bar Labels")
    s_ax.set_xlim(0.4, 4.6)
    s_ax.set_ylim(0, 7.6)
    s_ax.grid(axis="y")
    s_ax.set_axisbelow(True)
    xs = [1, 2, 3, 4]
    series_a = [1.4, 2.2, 1.8, 2.5]
    series_b = [2.1, 1.6, 2.4, 1.7]
    bottom = s_ax.bar(xs, series_a, color=(0.16, 0.59, 0.49), width=0.8)
    top = s_ax.bar(xs, series_b, bottom=series_a, color=(0.88, 0.47, 0.16), width=0.8)
    s_ax.bar_label(bottom, labels=["A1", "A2", "A3", "A4"], label_type="center", color="white", fontsize=10)
    s_ax.bar_label(top, fmt="%.1f", color=(0.20, 0.20, 0.20), fontsize=10, padding=4)

    save(fig, out_dir, "bar_variants")

PLOT = bar_variants


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
