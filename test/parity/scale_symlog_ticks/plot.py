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


def scale_symlog_ticks(out_dir):
    fig = make_fig_px(720, 400)
    ax = fig.add_axes(go_rect(0.12, 0.18, 0.92, 0.88))
    ax.set_title("Symlog Scale Ticks")
    ax.set_xlabel("signed value")
    ax.set_ylabel("response")
    ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=0.5)
    ax.set_axisbelow(True)

    x = [-1000, -100, -10, -1, 0, 1, 10, 100, 1000]
    y = [0.14, 0.23, 0.34, 0.46, 0.52, 0.59, 0.69, 0.80, 0.88]
    ax.plot(x, y, color=(0.12, 0.47, 0.71), linewidth=2.0)
    ax.set_xlim(-1000, 1000)
    ax.set_ylim(0, 1)
    ax.set_xscale("symlog", base=10, linthresh=1)
    ax.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))

    save(fig, out_dir, "scale_symlog_ticks")


PLOT = scale_symlog_ticks


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
