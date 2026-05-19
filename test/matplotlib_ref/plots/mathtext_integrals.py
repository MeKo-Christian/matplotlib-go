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


def mathtext_integrals(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.05, 0.08, 0.95, 0.92))
    ax.set_xlim(0, 1)
    ax.set_ylim(0, 1)
    ax.axis("off")

    ax.text(0.50, 0.77, r"$\int_0^\infty e^{-x}\,dx = 1$", transform=ax.transAxes, ha="center", va="center", fontsize=24)
    ax.text(0.50, 0.55, r"$\sum_{i=1}^{n} i^2$", transform=ax.transAxes, ha="center", va="center", fontsize=26)
    ax.text(0.50, 0.34, r"$\prod_{k=1}^{m} k$", transform=ax.transAxes, ha="center", va="center", fontsize=26)
    ax.text(0.50, 0.15, r"$\lim_{x\to 0} \frac{\sin x}{x} = 1$", transform=ax.transAxes, ha="center", va="center", fontsize=23)
    save(fig, out_dir, "mathtext_integrals")


PLOT = mathtext_integrals


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
