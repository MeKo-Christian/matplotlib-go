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


def scale_logit_ticks(out_dir):
    fig = make_fig_px(720, 400)
    ax = fig.add_axes(go_rect(0.12, 0.18, 0.92, 0.88))
    ax.set_title("Logit Scale Ticks")
    ax.set_xlabel("probability")
    ax.set_ylabel("score")
    ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=0.5)
    ax.set_axisbelow(True)

    x = [0.001, 0.01, 0.05, 0.2, 0.5, 0.8, 0.95, 0.99, 0.999]
    y = [0.12, 0.21, 0.34, 0.46, 0.55, 0.67, 0.78, 0.86, 0.91]
    ax.plot(x, y, color=(0.12, 0.47, 0.71), linewidth=2.0)
    ax.set_xlim(0.001, 0.999)
    ax.set_ylim(0, 1)
    ax.set_xscale("logit")
    ax.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))

    save(fig, out_dir, "scale_logit_ticks")


PLOT = scale_logit_ticks


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
