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

def units_custom_converter(out_dir):
    fig = make_fig_px(680, 380)
    ax = fig.add_axes(go_rect(0.10, 0.18, 0.94, 0.88))
    ax.set_title("Custom Distance Units")
    ax.set_xlabel("Distance")
    ax.set_ylabel("Pace")
    ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
    ax.set_axisbelow(True)

    distances = [5, 10, 21.1, 30, 42.2]
    pace = [6.4, 5.9, 5.3, 5.1, 5.4]
    ax.plot(distances, pace, color=(0.55, 0.34, 0.29), linewidth=1.4)
    ax.scatter(
        distances,
        pace,
        s=ss(8),
        c=[(0.17, 0.63, 0.17, 0.92)],
        edgecolors=[(0.09, 0.36, 0.09, 1.0)],
        linewidths=1.0,
    )
    ax.xaxis.set_major_formatter(matplotlib.ticker.FuncFormatter(lambda x, _: f"{x:.0f} km"))
    ax.margins(x=0.08, y=0.08)

    save(fig, out_dir, "units_custom_converter")

PLOT = units_custom_converter


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
