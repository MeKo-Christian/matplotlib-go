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

def transform_coordinates(out_dir):
    fig = make_fig_px(720, 420)
    ax = fig.add_axes(go_rect(0.13, 0.16, 0.90, 0.84))
    ax.set_title("Transform Coordinates")
    ax.set_xlabel("X")
    ax.set_ylabel("Y")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 10)
    ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
    ax.set_axisbelow(True)

    ax.plot(
        [1.0, 2.5, 4.5, 7.0, 8.8],
        [1.5, 3.2, 5.6, 6.4, 8.2],
        color=(0.14, 0.37, 0.74),
        linewidth=2.2,
    )
    ax.scatter(
        [2.5, 7.0, 8.8],
        [3.2, 6.4, 8.2],
        s=ss(8),
        marker="D",
        c=[(0.88, 0.42, 0.16, 0.92)],
        edgecolors=[(0.45, 0.18, 0.05, 1.0)],
        linewidths=1.0,
    )

    text_kwargs = dict(fontsize=11, color=(0.10, 0.10, 0.10))
    ax.text(1.3, 1.1, "data", transform=ax.transData, ha="left", va="baseline", **text_kwargs)
    ax.text(0.03, 0.97, "axes", transform=ax.transAxes, ha="left", va="top", **text_kwargs)
    fig.text(0.07, 0.08, "figure", ha="left", va="bottom", **text_kwargs)
    ax.text(
        0.50,
        0.22,
        "blend",
        transform=matplotlib.transforms.blended_transform_factory(fig.transFigure, ax.transAxes),
        ha="center",
        va="bottom",
        **text_kwargs,
    )
    ax.annotate(
        "axes note",
        xy=(0.82, 0.78),
        xycoords="axes fraction",
        xytext=(-48, -26),
        textcoords="offset pixels",
        fontsize=10,
        color=(0.10, 0.10, 0.10),
        ha="right",
        va="top",
        arrowprops=dict(arrowstyle="->", color=(0.10, 0.10, 0.10), lw=1.25),
    )

    save(fig, out_dir, "transform_coordinates")

PLOT = transform_coordinates


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
