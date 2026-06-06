#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import numpy as np

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

def fill_variants(out_dir):
    fig = make_fig_px(840, 620)
    axes = {
        "between": fig.add_axes(go_rect(0.07, 0.585, 0.46, 0.93)),
        "betweenx": fig.add_axes(go_rect(0.56, 0.585, 0.96, 0.93)),
        "stacked": fig.add_axes(go_rect(0.07, 0.10, 0.46, 0.445)),
        "alpha": fig.add_axes(go_rect(0.56, 0.10, 0.96, 0.445)),
    }

    # Fill between two curves.
    b_ax = axes["between"]
    b_ax.set_title("Fill Between")
    b_ax.set_xlim(0, 6.28)
    b_ax.set_ylim(-1.5, 1.5)
    n = 50
    x = [6.28 * i / (n - 1) for i in range(n)]
    y1 = [np.sin(t) for t in x]
    y2 = [0.8 * np.cos(t) for t in x]
    b_ax.fill_between(x, y1, y2, facecolor=(0.8, 0.3, 0.3, 0.6),
                      edgecolor=(0.5, 0.1, 0.1, 1.0), linewidth=lw(1.5))
    b_ax.plot(x, y1, color=(1, 0, 0), linewidth=lw(2))
    b_ax.plot(x, y2, color=(0, 0, 1), linewidth=lw(2))

    # Fill betweenx.
    bx_ax = axes["betweenx"]
    bx_ax.set_title("Fill BetweenX")
    bx_ax.set_xlim(0, 7)
    bx_ax.set_ylim(0, 6)
    bx_ax.fill_betweenx(
        [0.4, 1.2, 2.0, 2.8, 3.6, 4.4, 5.2],
        [1.3, 2.1, 1.7, 2.8, 2.2, 3.1, 2.6],
        [3.4, 4.1, 4.8, 5.1, 5.6, 6.0, 6.3],
        facecolor=(0.24, 0.68, 0.54, 0.72),
        edgecolor=(0.12, 0.38, 0.28, 1.0),
        linewidth=lw(1.2),
    )

    # Stacked fills.
    s_ax = axes["stacked"]
    s_ax.set_title("Stacked Fills")
    s_ax.set_xlim(0, 8)
    s_ax.set_ylim(0, 8)
    xs = [1, 2, 3, 4, 5, 6, 7]
    layer1 = [1, 1.5, 2, 1.8, 2.2, 1.9, 1.6]
    layer2 = [layer1[i] + 1.5 + 0.3 * np.sin(i) for i in range(len(layer1))]
    layer3 = [layer2[i] + 1.2 + 0.4 * np.cos(i) for i in range(len(layer1))]
    base = [0] * len(xs)
    s_ax.fill_between(xs, base, layer1, facecolor=(0.8, 0.2, 0.2, 0.8),
                      edgecolor=(0.5, 0, 0, 1), linewidth=lw(1.0))
    s_ax.fill_between(xs, layer1, layer2, facecolor=(0.2, 0.8, 0.2, 0.8),
                      edgecolor=(0, 0.5, 0, 1), linewidth=lw(1.0))
    s_ax.fill_between(xs, layer2, layer3, facecolor=(0.2, 0.2, 0.8, 0.8),
                      edgecolor=(0, 0, 0.5, 1), linewidth=lw(1.0))

    # Alpha overlap.
    a_ax = axes["alpha"]
    a_ax.set_title("Alpha Overlap")
    a_ax.set_xlim(0, 6)
    a_ax.set_ylim(0, 5)
    xa = [6.0 * i / (n - 1) for i in range(n)]
    ya = [2.5 + 1.6 * np.sin(t) for t in xa]
    yb = [2.5 + 1.6 * np.cos(t) for t in xa]
    zero = [0] * n
    a_ax.fill_between(xa, zero, ya, facecolor=(0.85, 0.25, 0.25, 0.45), linewidth=0)
    a_ax.fill_between(xa, zero, yb, facecolor=(0.20, 0.35, 0.85, 0.45), linewidth=0)

    save(fig, out_dir, "fill_variants")

PLOT = fill_variants


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
