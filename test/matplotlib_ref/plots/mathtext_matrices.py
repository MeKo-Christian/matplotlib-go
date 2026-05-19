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


def mathtext_matrices(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.05, 0.08, 0.95, 0.92))
    ax.set_xlim(0, 1)
    ax.set_ylim(0, 1)
    ax.axis("off")

    ax.text(0.30, 0.64, r"$\genfrac{(}{)}{0}{0}{a\quad b}{c\quad d}$", transform=ax.transAxes, ha="center", va="center", fontsize=25)
    ax.text(0.70, 0.64, r"$\genfrac{[}{]}{0}{0}{1\quad 0}{0\quad 1}$", transform=ax.transAxes, ha="center", va="center", fontsize=25)
    ax.text(0.30, 0.30, r"$\genfrac{(}{)}{0}{0}{x}{y}$", transform=ax.transAxes, ha="center", va="center", fontsize=24)
    ax.text(0.70, 0.30, r"$\left\langle\genfrac{}{}{0}{0}{u}{v}\right\rangle$", transform=ax.transAxes, ha="center", va="center", fontsize=24)
    save(fig, out_dir, "mathtext_matrices")


PLOT = mathtext_matrices


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
