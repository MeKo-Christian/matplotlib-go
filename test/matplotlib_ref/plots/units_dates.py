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

def units_dates(out_dir):
    fig = make_fig_px(720, 380)
    ax = fig.add_axes(go_rect(0.10, 0.18, 0.94, 0.88))
    ax.set_title("Date Units")
    ax.set_xlabel("Date")
    ax.set_ylabel("Requests")
    ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=0.5)
    ax.set_axisbelow(True)

    dates = [
        dt.datetime(2024, 2, 1),
        dt.datetime(2024, 2, 5),
        dt.datetime(2024, 2, 9),
        dt.datetime(2024, 2, 14),
        dt.datetime(2024, 2, 20),
    ]
    lower = [6, 7, 5, 8, 7]
    upper = [10, 15, 13, 18, 16]
    ax.fill_between(dates, lower, upper, color=(0.85, 0.91, 0.96), linewidth=0)
    ax.plot(
        dates,
        [8, 12, 9, 15, 13],
        color=(0.12, 0.47, 0.71),
        linewidth=2.0,
    )
    ax.xaxis.set_major_locator(mdates.DayLocator(bymonthday=[5, 12, 19]))
    ax.xaxis.set_major_formatter(mdates.DateFormatter("%d %b"))
    ax.margins(x=0.06, y=0.06)

    save(fig, out_dir, "units_dates")

PLOT = units_dates


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
