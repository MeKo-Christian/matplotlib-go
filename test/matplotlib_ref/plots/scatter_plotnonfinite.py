#!/usr/bin/env python3
"""Matplotlib reference plot for scatter plotnonfinite=True."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import math

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def _data():
    n = 12
    x = [0.5 + 0.5 * i for i in range(n)]
    y = [3 + 2 * math.sin(i * 0.6) for i in range(n)]
    c = [i / (n - 1) for i in range(n)]
    c[3] = float("nan")
    c[8] = float("nan")
    x[6] = float("nan")
    return x, y, c


def scatter_plotnonfinite(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.12, 0.12, 0.9, 0.9))
    ax.set_title("Scatter (plotnonfinite)")
    ax.set_xlim(0, 7)
    ax.set_ylim(0, 6)

    x, y, c = _data()
    ax.scatter(
        x,
        y,
        c=c,
        cmap="viridis",
        vmin=0.0,
        vmax=1.0,
        s=ss(7),
        linewidths=0,
        plotnonfinite=True,
    )
    save(fig, out_dir, "scatter_plotnonfinite")


PLOT = scatter_plotnonfinite


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
