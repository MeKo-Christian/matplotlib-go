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


def date_month_year_labels(out_dir):
    fig = make_fig_px(720, 360)
    ax = fig.add_axes(go_rect(0.10, 0.18, 0.94, 0.88))
    ax.set_title("Monthly + Yearly Date Labels")
    ax.set_xlabel("month")
    ax.set_ylabel("index")
    ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    ax.set_axisbelow(True)

    dates = [
        dt.datetime(2023, 1, 1),
        dt.datetime(2023, 7, 1),
        dt.datetime(2024, 1, 1),
        dt.datetime(2024, 7, 1),
    ]
    ax.plot(dates, [2, 4, 3, 7], color=(0.12, 0.47, 0.71), linewidth=lw(2.0))
    ax.set_xlim(dates[0], dates[-1])
    ax.set_ylim(0, 8)
    ax.xaxis.set_major_locator(mdates.MonthLocator(bymonth=[1, 7]))
    ax.xaxis.set_major_formatter(mdates.DateFormatter("%b %Y"))
    ax.yaxis.set_major_locator(matplotlib.ticker.FixedLocator([0, 4, 8]))

    save(fig, out_dir, "date_month_year_labels")


PLOT = date_month_year_labels


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
