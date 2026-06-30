#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import matplotlib.ticker as mticker
import numpy as np

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2]))
    from matplotlib_ref.common import *  # noqa: F401,F403


def scale_function_defaults(out_dir):
    fig = make_fig_px(720, 480)
    top = fig.add_axes(go_rect(0.12, 0.58, 0.92, 0.90))
    bottom = fig.add_axes(go_rect(0.12, 0.15, 0.92, 0.47))

    top.set_title("Function Scale Defaults")
    top.set_xlabel("sqrt-scaled x")
    top.set_ylabel("response")
    top.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=0.5)
    top.set_axisbelow(True)
    x = [0, 4, 16, 36, 64, 100]
    y = [0.12, 0.25, 0.41, 0.58, 0.76, 0.90]
    top.plot(x, y, color=(0.12, 0.47, 0.71), linewidth=2.0)
    top.set_xlim(0, 100)
    top.set_ylim(0, 1)
    top.set_xscale("function", functions=(np.sqrt, np.square))
    top.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))

    bottom.set_title("Functionlog Scale Defaults")
    bottom.set_xlabel("sqrt-scaled log x")
    bottom.set_ylabel("response")
    bottom.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=0.5)
    bottom.set_axisbelow(True)
    x = [1, 10, 100, 1000, 10000]
    y = [0.12, 0.31, 0.52, 0.72, 0.90]
    bottom.plot(x, y, color=(0.12, 0.47, 0.71), linewidth=2.0)
    bottom.set_xlim(1, 10000)
    bottom.set_ylim(0, 1)
    bottom.set_xscale("functionlog", functions=(np.sqrt, np.square), base=10)
    bottom.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))

    save(fig, out_dir, "scale_function_defaults")


PLOT = scale_function_defaults


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
