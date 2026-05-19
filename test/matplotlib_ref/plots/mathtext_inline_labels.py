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
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def mathtext_inline_labels(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.92, 0.88))

    n = 90
    xs = [i / (n - 1) * 5 for i in range(n)]
    y1 = [0.55 + 0.35 * math.sin(1.5 * t) for t in xs]
    y2 = [0.48 + 0.28 * math.cos(1.5 * t + 0.45) for t in xs]

    ax.plot(xs, y1, linewidth=lw(2), color=TAB10[0], label=r"state $x_i(t)$")
    ax.plot(xs, y2, linewidth=lw(2), color=TAB10[1], label=r"state $y_i(t)$")
    ax.set_title(r"Inline labels: $\omega_n$ response")
    ax.set_xlabel(r"time $t$")
    ax.set_ylabel(r"state $x_i(t)$")
    ax.legend()
    ax.text(0.03, 0.88, r"peak $\alpha_i^2$", transform=ax.transAxes, ha="left", va="top", fontsize=12)
    ax.text(0.97, 0.08, r"ratio $\frac{a}{b}$", transform=ax.transAxes, ha="right", va="bottom", fontsize=12)
    save(fig, out_dir, "mathtext_inline_labels")


PLOT = mathtext_inline_labels


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
