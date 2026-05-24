#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2]))
    from matplotlib_ref.common import *  # noqa: F401,F403


def date_concise_intraday_labels(out_dir):
    fig = make_fig_px(720, 360)
    ax = fig.add_axes(go_rect(0.10, 0.18, 0.94, 0.88))
    ax.set_title("Concise Intraday Dates")
    ax.set_xlabel("time")
    ax.set_ylabel("requests")
    ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    ax.set_axisbelow(True)

    start = dt.datetime(2024, 1, 2, 0, 0, 0)
    times = [start + dt.timedelta(hours=h) for h in [0, 6, 12, 18]]
    ax.plot(times, [8, 11, 9, 14], color=(0.12, 0.47, 0.71), linewidth=lw(2.0))
    ax.set_xlim(times[0], times[-1])
    ax.set_ylim(0, 16)
    locator = mdates.HourLocator(byhour=[0, 6, 12, 18])
    ax.xaxis.set_major_locator(locator)
    ax.xaxis.set_major_formatter(mdates.ConciseDateFormatter(locator))
    ax.yaxis.set_major_locator(matplotlib.ticker.FixedLocator([0, 8, 16]))

    save(fig, out_dir, "date_concise_intraday_labels")


PLOT = date_concise_intraday_labels


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
