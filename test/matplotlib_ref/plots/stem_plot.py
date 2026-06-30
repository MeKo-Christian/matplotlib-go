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

def stem_plot(out_dir):
    fig = make_fig_px(720, 420)
    ax = fig.add_axes(go_rect(0.10, 0.16, 0.94, 0.86))
    ax.set_title("Stem")
    ax.set_xlabel("Sample")
    ax.set_ylabel("Amplitude")
    ax.set_xlim(0.5, 7.5)
    ax.set_ylim(-0.2, 4.2)
    ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=0.5)
    ax.set_axisbelow(True)
    markerline, stemlines, baseline = ax.stem(
        [1, 2, 3, 4, 5, 6, 7],
        [0.9, 2.2, 1.6, 3.3, 2.4, 3.7, 2.1],
        basefmt="-",
        bottom=0.3,
    )
    stem_color = (0.15, 0.42, 0.73)
    # Stem and baseline line widths are left at Matplotlib's default
    # (rcParams["lines.linewidth"] = 1.5 points), matching the Go example which
    # does not override StemOptions.LineWidth / BaselineWidth.
    plt.setp(stemlines, color=stem_color)
    plt.setp(markerline, color=stem_color, markerfacecolor=stem_color, markeredgecolor=stem_color, markersize=7)
    plt.setp(baseline, color=(0.32, 0.32, 0.32))

    save(fig, out_dir, "stem_plot")

PLOT = stem_plot


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
