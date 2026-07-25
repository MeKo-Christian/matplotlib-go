#!/usr/bin/env python3
"""Focused Matplotlib parity fixture for logarithmic plot wrappers."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import matplotlib.ticker as mticker

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2]))
    from matplotlib_ref.common import *  # noqa: F401,F403


def axes_log_plot_wrappers(out_dir):
    fig = make_fig_px(960, 360)
    color = (0.12, 0.47, 0.71)

    semilog_x = fig.add_axes(go_rect(0.07, 0.18, 0.31, 0.86))
    semilog_x.set_title("SemilogX")
    semilog_x.set_xlabel("log x")
    semilog_x.set_ylabel("linear y")
    semilog_x.semilogx(
        [1, 3, 10, 30, 100, 300, 1000],
        [0.5, 1.0, 1.8, 2.6, 3.2, 4.1, 4.6],
        color=color,
        linewidth=2.0,
    )
    semilog_x.set_xlim(1, 1000)
    semilog_x.set_ylim(0, 5)
    semilog_x.yaxis.set_major_locator(mticker.FixedLocator([0, 2.5, 5]))
    semilog_x.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
    semilog_x.set_axisbelow(True)

    semilog_y = fig.add_axes(go_rect(0.39, 0.18, 0.63, 0.86))
    semilog_y.set_title("SemilogY")
    semilog_y.set_xlabel("linear x")
    semilog_y.set_ylabel("log y")
    semilog_y.semilogy(
        [0, 1, 2, 3, 4, 5, 6],
        [1, 3, 10, 30, 100, 300, 1000],
        color=color,
        linewidth=2.0,
    )
    semilog_y.set_xlim(0, 6)
    semilog_y.set_ylim(1, 1000)
    semilog_y.xaxis.set_major_locator(mticker.FixedLocator([0, 3, 6]))
    semilog_y.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
    semilog_y.set_axisbelow(True)

    log_log = fig.add_axes(go_rect(0.71, 0.18, 0.95, 0.86))
    log_log.set_title("LogLog")
    log_log.set_xlabel("log x")
    log_log.set_ylabel("log y")
    log_log.loglog(
        [1, 3, 10, 30, 100, 300, 1000],
        [1, 2, 7, 18, 70, 240, 900],
        color=color,
        linewidth=2.0,
    )
    log_log.set_xlim(1, 1000)
    log_log.set_ylim(1, 1000)
    log_log.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
    log_log.set_axisbelow(True)

    save(fig, out_dir, "axes_log_plot_wrappers")


PLOT = axes_log_plot_wrappers


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
