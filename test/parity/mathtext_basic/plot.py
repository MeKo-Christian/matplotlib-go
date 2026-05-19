#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import math
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2]))
    from matplotlib_ref.common import *  # noqa: F401,F403


def mathtext_basic(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.10, 0.14, 0.92, 0.88))

    n = 120
    xs = [i / (n - 1) * 4 * math.pi for i in range(n)]
    ys = [math.sin(t) * math.exp(-0.08 * t) for t in xs]

    ax.plot(xs, ys, linewidth=lw(2), color=TAB10[0])
    ax.set_title(r"MathText $\alpha^2 + \beta_i$")
    ax.set_xlabel(r"phase $\theta$")
    ax.set_ylabel(r"amplitude $\frac{1}{\sqrt{2}}$")
    ax.text(
        0.98,
        0.92,
        r"$x_{\mathrm{max}}$",
        transform=ax.transAxes,
        ha="right",
        va="top",
        fontsize=12,
    )
    ax.annotate(
        r"$\Delta y \approx \frac{1}{2}$",
        xy=(3.2, 0.35),
        xytext=(34, -26),
        textcoords="offset points",
        fontsize=12,
        arrowprops={"arrowstyle": "->", "linewidth": lw(1), "color": "black"},
    )
    ax.text(
        0.03,
        0.93,
        r"$\omega_n = 2\pi f_n$",
        transform=ax.transAxes,
        ha="left",
        va="top",
        fontsize=11,
        bbox={"boxstyle": "square,pad=0.3", "facecolor": "white", "edgecolor": "0.8"},
    )
    save(fig, out_dir, "mathtext_basic")


PLOT = mathtext_basic


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
