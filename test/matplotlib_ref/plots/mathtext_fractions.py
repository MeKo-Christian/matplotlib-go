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


def mathtext_fractions(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.05, 0.08, 0.95, 0.92))
    ax.set_xlim(0, 1)
    ax.set_ylim(0, 1)
    ax.axis("off")

    ax.text(0.50, 0.76, r"$\frac{1}{2} + \frac{a + b}{c + d}$", transform=ax.transAxes, ha="center", va="center", fontsize=24)
    ax.text(0.50, 0.54, r"$\binom{n}{k} = \frac{n!}{k!(n-k)!}$", transform=ax.transAxes, ha="center", va="center", fontsize=23)
    ax.text(0.50, 0.34, r"$\sqrt[3]{x + 1} + \sqrt{y}$", transform=ax.transAxes, ha="center", va="center", fontsize=23)
    ax.text(0.50, 0.16, r"$\left[\frac{1}{1+x}\right]$", transform=ax.transAxes, ha="center", va="center", fontsize=24)
    save(fig, out_dir, "mathtext_fractions")


PLOT = mathtext_fractions


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
