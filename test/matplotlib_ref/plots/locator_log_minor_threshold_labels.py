#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

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


def locator_log_minor_threshold_labels(out_dir):
    fig = make_fig_px(720, 420)
    top = fig.add_axes(go_rect(0.12, 0.58, 0.92, 0.86))
    bottom = fig.add_axes(go_rect(0.12, 0.16, 0.92, 0.44))

    top.set_title("LogLocator Auto Minor Grid")
    top.set_xlabel("base 10")
    top.set_ylabel("value")
    top.plot(
        [1, 10, 100, 1000],
        [0.18, 0.39, 0.67, 0.88],
        color=(0.12, 0.47, 0.71),
        linewidth=lw(2.0),
    )
    top.set_ylim(0, 1)
    top.set_xscale("log", base=10)
    top.set_xlim(1, 1000)
    top.xaxis.set_minor_locator(mticker.LogLocator(base=10, subs="auto"))
    top.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))
    top.grid(True, axis="x", which="major", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    top.grid(
        True,
        axis="x",
        which="minor",
        color=(0.72, 0.76, 0.80, 0.55),
        linewidth=lw(0.35),
        linestyle=(0, (1, 2)),
    )
    top.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    top.set_axisbelow(True)

    bottom.set_title("LogLocator Base 2")
    bottom.set_xlabel("base 2")
    bottom.set_ylabel("value")
    bottom.plot(
        [1, 2, 4, 8, 16, 32, 64],
        [0.14, 0.26, 0.38, 0.52, 0.68, 0.80, 0.91],
        color=(0.12, 0.47, 0.71),
        linewidth=lw(2.0),
    )
    bottom.set_ylim(0, 1)
    bottom.set_xscale("log", base=2)
    bottom.set_xlim(1, 64)
    bottom.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))
    bottom.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    bottom.set_axisbelow(True)

    save(fig, out_dir, "locator_log_minor_threshold_labels")


PLOT = locator_log_minor_threshold_labels


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
