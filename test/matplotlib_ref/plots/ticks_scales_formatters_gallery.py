#!/usr/bin/env python3
"""Focused tick, scale, formatter, date, category, and unit gallery."""

from __future__ import annotations

from pathlib import Path
import argparse
import datetime as dt
import sys

import matplotlib.dates as mdates
import matplotlib.ticker as mticker

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def panel_rect(row, col):
    x0 = 0.07 + col * 0.46
    y0 = 0.66 - row * 0.29
    return go_rect(x0, y0, x0 + 0.38, y0 + 0.18)


def configure_axes(ax, title, xlabel, ylabel):
    ax.set_title(title)
    ax.set_xlabel(xlabel)
    ax.set_ylabel(ylabel)
    ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    ax.set_axisbelow(True)
    ax.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))


def ticks_scales_formatters_gallery(out_dir):
    fig = make_fig_px(1320, 900)
    fig.text(0.05, 0.955, "Ticks, Scales, and Formatters", fontsize=15)
    fig.text(
        0.05,
        0.928,
        "locators, scale defaults, formatter families, dates, categories, and custom units",
        fontsize=11,
    )

    ax = fig.add_axes(panel_rect(0, 0))
    configure_axes(ax, "Major and Minor Locators", "MultipleLocator + minor ticks", "score")
    ax.plot([0, 1.5, 3, 4.5, 6], [0.12, 0.38, 0.58, 0.74, 0.90], color=TAB10[0], linewidth=lw(2.0))
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 1)
    ax.xaxis.set_major_locator(mticker.MultipleLocator(1.5))
    ax.xaxis.set_minor_locator(mticker.AutoMinorLocator(3))
    ax.grid(True, axis="x", which="minor", color=(0.72, 0.76, 0.80, 0.55), linewidth=lw(0.35), linestyle=(0, (1, 2)))

    ax = fig.add_axes(panel_rect(0, 1))
    configure_axes(ax, "Log Scale and Minor Grid", "base-10 log", "score")
    ax.plot([1, 3, 10, 30, 100, 300, 1000], [0.10, 0.22, 0.38, 0.55, 0.70, 0.82, 0.91], color=TAB10[1], linewidth=lw(2.0))
    ax.set_xscale("log", base=10)
    ax.set_xlim(1, 1000)
    ax.set_ylim(0, 1)
    ax.xaxis.set_minor_locator(mticker.LogLocator(base=10, subs="auto"))
    ax.xaxis.set_major_formatter(mticker.LogFormatterMathtext(base=10))
    ax.grid(True, axis="x", which="minor", color=(0.72, 0.76, 0.80, 0.55), linewidth=lw(0.35), linestyle=(0, (1, 2)))

    ax = fig.add_axes(panel_rect(1, 0))
    configure_axes(ax, "Signed Scale Defaults", "symlog with signed markers", "response")
    x = [-1000, -100, -10, -1, 0, 1, 10, 100, 1000]
    ax.plot(x, [0.12, 0.21, 0.32, 0.44, 0.51, 0.59, 0.70, 0.82, 0.91], color=TAB10[2], linewidth=lw(2.0))
    ax.scatter(x, [0.15, 0.24, 0.35, 0.47, 0.54, 0.62, 0.73, 0.85, 0.94], color=TAB10[4], s=ss(4.5))
    ax.set_xscale("symlog", base=10, linthresh=1)
    ax.set_xlim(-1000, 1000)
    ax.set_ylim(0, 1)

    ax = fig.add_axes(panel_rect(1, 1))
    configure_axes(ax, "Formatter Families", "position", "formatted values")
    ax.plot([0, 1, 2, 3, 4], [0.15, 0.32, 0.48, 0.66, 0.86], color=TAB10[4], linewidth=lw(2.0))
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 1)
    ax.xaxis.set_major_locator(mticker.FixedLocator([0, 1, 2, 3, 4]))
    ax.xaxis.set_major_formatter(mticker.FixedFormatter(["0", "1 kHz", "25%", "1.2e3", "custom"]))
    ax.yaxis.set_major_locator(mticker.FixedLocator([0.15, 0.50, 0.85]))
    ax.yaxis.set_major_formatter(mticker.FuncFormatter(lambda v, _pos: f"y={v:.2f}"))

    ax = fig.add_axes(panel_rect(2, 0))
    configure_axes(ax, "Date and Category Formatters", "date axis with category labels", "requests")
    dates = [
        dt.datetime(2024, 2, 1),
        dt.datetime(2024, 2, 5),
        dt.datetime(2024, 2, 9),
        dt.datetime(2024, 2, 14),
        dt.datetime(2024, 2, 20),
    ]
    ax.plot(dates, [0.08, 0.38, 0.30, 0.48, 0.42], color=TAB10[5], linewidth=lw(2.0))
    ax.xaxis.set_major_locator(mdates.DayLocator(bymonthday=[1, 7, 14, 21], tz=dt.timezone.utc))
    ax.xaxis.set_major_formatter(mdates.DateFormatter("%d %b", tz=dt.timezone.utc))
    ax.set_ylim(0, 1)
    ax.margins(0.04)
    inset = fig.add_axes(go_rect(0.30, 0.16, 0.43, 0.30))
    inset.set_title("Categories")
    inset.bar(["draft", "review", "ship"], [0.35, 0.75, 0.55], color=(0.50, 0.50, 0.50), width=0.72)
    inset.set_ylim(0, 1)
    inset.tick_params(axis="x", labelbottom=False)

    ax = fig.add_axes(panel_rect(2, 1))
    configure_axes(ax, "Custom Unit Converter", "distance", "pace")
    dist = [5, 10, 21.1, 30, 42.2]
    pace = [0.75, 0.69, 0.58, 0.52, 0.60]
    ax.plot(dist, pace, color=TAB10[0], linewidth=lw(2.0))
    ax.scatter(dist, pace, color=TAB10[2], edgecolor=TAB10[0], s=ss(5.0))
    ax.set_xlim(3, 44)
    ax.set_ylim(0, 1)
    ax.xaxis.set_major_formatter(mticker.FuncFormatter(lambda v, _pos: f"{v:g} km"))
    ax.yaxis.set_major_formatter(mticker.PercentFormatter(xmax=1))

    save(fig, out_dir, "ticks_scales_formatters_gallery")


PLOT = ticks_scales_formatters_gallery


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
