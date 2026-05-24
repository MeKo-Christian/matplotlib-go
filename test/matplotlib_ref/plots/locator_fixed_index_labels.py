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


def locator_fixed_index_labels(out_dir):
    fig = make_fig_px(720, 420)
    top = fig.add_axes(go_rect(0.12, 0.58, 0.92, 0.86))
    bottom = fig.add_axes(go_rect(0.12, 0.16, 0.92, 0.44))

    top.set_title("FixedLocator Subsampling")
    top.set_xlabel("fixed positions")
    top.set_ylabel("value")
    top.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    top.set_axisbelow(True)
    top.plot(
        [-6, -4, -2, 0, 2, 4, 6],
        [0.18, 0.34, 0.47, 0.62, 0.74, 0.86, 0.92],
        color=(0.12, 0.47, 0.71),
        linewidth=lw(2.0),
    )
    top.set_xlim(-6, 6)
    top.set_ylim(0, 1)
    top.xaxis.set_major_locator(mticker.FixedLocator([-6, -4, -2, 0, 2, 4, 6], nbins=4))
    top.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))

    bottom.set_title("IndexLocator Base + Offset")
    bottom.set_xlabel("index")
    bottom.set_ylabel("value")
    bottom.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    bottom.set_axisbelow(True)
    bottom.plot(
        [0, 1, 2, 3, 4, 5, 6, 7, 8],
        [0.15, 0.28, 0.36, 0.51, 0.63, 0.70, 0.82, 0.88, 0.93],
        color=(0.12, 0.47, 0.71),
        linewidth=lw(2.0),
    )
    bottom.set_xlim(0, 8)
    bottom.set_ylim(0, 1)
    bottom.xaxis.set_major_locator(mticker.IndexLocator(base=2, offset=1))
    bottom.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))

    save(fig, out_dir, "locator_fixed_index_labels")


PLOT = locator_fixed_index_labels


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
