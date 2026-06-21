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
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def formatter_scalar_scientific_labels(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.14, 0.18, 0.92, 0.88))
    ax.set_title("Scalar Scientific Formatter")
    ax.set_xlabel("value")
    ax.set_ylabel("score")
    ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    ax.set_axisbelow(True)

    x = [-1200, 0, 1200]
    y = [0.28, 0.62, 0.78]
    ax.plot(x, y, color=(0.12, 0.47, 0.71), linewidth=lw(2.0))
    ax.set_xlim(-1400, 1400)
    ax.set_ylim(0, 1)
    ax.xaxis.set_major_locator(mticker.FixedLocator(x))
    ax.xaxis.set_major_formatter(scalar_scientific_formatter())
    ax.yaxis.set_major_locator(mticker.FixedLocator([0, 0.5, 1.0]))

    save(fig, out_dir, "formatter_scalar_scientific_labels")


def scalar_scientific_formatter():
    # The Go port's core.ScalarFormatter now mirrors Matplotlib's ScalarFormatter
    # offset / order-of-magnitude factoring (Phase 7), so the reference uses the
    # real ScalarFormatter rather than a per-tick FuncFormatter workaround.
    formatter = mticker.ScalarFormatter(useMathText=True)
    formatter.set_scientific(True)
    formatter.set_powerlimits((0, 0))
    return formatter


PLOT = formatter_scalar_scientific_labels


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
