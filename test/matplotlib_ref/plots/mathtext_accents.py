#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


# MathText accent expressions supported by Matplotlib's mathtext. The best-effort,
# non-parity extensions (\overbrace, \underbrace, \stackrel, \not) are excluded.
LABELS = [
    (0.20, 0.85, r"$\hat{x}$"),
    (0.50, 0.85, r"$\bar{x}$"),
    (0.80, 0.85, r"$\vec{v}$"),
    (0.20, 0.62, r"$\dot{x}$"),
    (0.50, 0.62, r"$\ddot{y}$"),
    (0.80, 0.62, r"$\tilde{n}$"),
    (0.20, 0.39, r"$\widehat{AB}$"),
    (0.50, 0.39, r"$\widetilde{xy}$"),
    (0.80, 0.39, r"$\overline{x+y}$"),
    (0.20, 0.16, r"$\overset{a}{X}$"),
    (0.50, 0.16, r"$\underset{b}{Y}$"),
    (0.80, 0.16, r"$X_{\substack{i \\ j}}$"),
]


def mathtext_accents(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.05, 0.08, 0.95, 0.92))
    ax.set_xlim(0, 1)
    ax.set_ylim(0, 1)
    ax.axis("off")

    for x, y, expr in LABELS:
        ax.text(x, y, expr, transform=ax.transAxes, ha="center", va="center", fontsize=26)
    save(fig, out_dir, "mathtext_accents")


PLOT = mathtext_accents


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
