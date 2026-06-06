#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


PANELS = [
    ("Fractions and Roots", [r"$\frac{1}{2} + \frac{a+b}{c+d}$", r"$\sqrt{x^2 + y^2}$"]),
    ("Operators", [r"$\int_0^\infty e^{-x}\,dx = 1$", r"$\sum_{i=1}^{n} i^2$"]),
    ("Fences and Matrices", [r"$\left[\frac{1}{1+x}\right]$", r"$\genfrac{(}{)}{0}{0}{a\quad b}{c\quad d}$"]),
    ("Inline Labels", [r"phase $\alpha_i^2$ peak", r"ratio $\frac{a}{b}$"]),
]


def mathtext_gallery(out_dir):
    fig = make_fig_px(900, 560)
    fig.text(0.05, 0.94, "MathText Gallery", fontsize=13)

    for i, (title, exprs) in enumerate(PANELS):
        col = i % 2
        row = i // 2
        x0 = 0.07 + col * 0.46
        y0 = 0.52 - row * 0.42
        ax = fig.add_axes([x0, y0, 0.39, 0.31])
        ax.set_title(title)
        ax.set_xlim(0, 1)
        ax.set_ylim(0, 1)
        ax.set_xticks([])
        ax.set_yticks([])
        ax.text(0.50, 0.64, exprs[0], ha="center", va="center", fontsize=18)
        ax.text(0.50, 0.34, exprs[1], ha="center", va="center", fontsize=17)

    save(fig, out_dir, "mathtext_gallery")


PLOT = mathtext_gallery


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
