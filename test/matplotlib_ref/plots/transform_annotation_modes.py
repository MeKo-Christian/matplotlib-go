#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2]))
    from matplotlib_ref.common import *  # noqa: F401,F403


def transform_annotation_modes(out_dir):
    fig = make_fig_px(720, 420)
    ax = fig.add_axes(go_rect(0.12, 0.15, 0.88, 0.82))
    ax.set_title("Annotation Coordinate Modes")
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 10)
    ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
    ax.set_axisbelow(True)

    ax.plot([1, 3, 5, 7, 9], [1.5, 4, 3, 7, 8.5], color=(0.12, 0.47, 0.71), linewidth=2.0)
    arrowprops = dict(arrowstyle="->", color=(0.10, 0.10, 0.10), lw=1.1)
    text = dict(fontsize=10, color=(0.10, 0.10, 0.10))

    ax.annotate("data", xy=(3, 4), xycoords="data", xytext=(34, -30), textcoords="offset pixels", arrowprops=arrowprops, **text)
    ax.annotate("axes", xy=(0.78, 0.74), xycoords="axes fraction", xytext=(-46, 28), textcoords="offset pixels", ha="right", arrowprops=arrowprops, **text)
    ax.annotate("figure", xy=(0.72, 0.24), xycoords="figure fraction", xytext=(42, 24), textcoords="offset pixels", arrowprops=arrowprops, **text)

    save(fig, out_dir, "transform_annotation_modes")


PLOT = transform_annotation_modes


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
