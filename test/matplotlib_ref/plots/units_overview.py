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

def units_overview(out_dir):
    fig = make_fig_px(1200, 420)

    date_ax = fig.add_axes(go_rect(0.06, 0.18, 0.32, 0.86))
    date_ax.set_title("Dates")
    date_ax.set_ylabel("Requests")
    date_ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=0.5)
    date_ax.set_axisbelow(True)
    date_ax.plot(
        [
            dt.datetime(2024, 1, 1),
            dt.datetime(2024, 1, 3),
            dt.datetime(2024, 1, 7),
            dt.datetime(2024, 1, 10),
        ],
        [12, 18, 9, 21],
        color=(0.12, 0.47, 0.71),
        linewidth=2.0,
    )
    date_ax.margins(x=0.05, y=0.05)

    category_ax = fig.add_axes(go_rect(0.38, 0.18, 0.64, 0.86))
    category_ax.set_title("Categories")
    category_ax.set_ylabel("Count")
    category_ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=0.5)
    category_ax.set_axisbelow(True)
    category_ax.bar(
        ["alpha", "beta", "gamma", "delta"],
        [4, 9, 6, 7],
        color=(1.0, 0.50, 0.05),
        edgecolor=(0.60, 0.30, 0.03),
        linewidth=1.0,
        width=0.8,
    )
    category_ax.margins(x=0.10, y=0.10)

    unit_ax = fig.add_axes(go_rect(0.70, 0.18, 0.96, 0.86))
    unit_ax.set_title("Custom Units")
    unit_ax.set_xlabel("Distance")
    unit_ax.set_ylabel("Pace")
    unit_ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
    unit_ax.set_axisbelow(True)
    unit_ax.scatter(
        [5, 10, 21.1, 42.2],
        [6.4, 5.8, 5.2, 5.5],
        s=ss(8),
        c=[(0.17, 0.63, 0.17, 0.92)],
        edgecolors=[(0.09, 0.36, 0.09, 1.0)],
        linewidths=1.0,
    )
    unit_ax.xaxis.set_major_formatter(matplotlib.ticker.FuncFormatter(lambda x, _: f"{x:.0f} km"))
    unit_ax.margins(x=0.08, y=0.08)

    save(fig, out_dir, "units_overview")

PLOT = units_overview


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
