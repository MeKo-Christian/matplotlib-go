#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

from matplotlib.ticker import MultipleLocator

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

def axes_control_surface(out_dir):
    fig = make_fig_px(760, 360)

    left = fig.add_axes(go_rect(0.07, 0.14, 0.47, 0.78))
    left.set_title("Moved Axes + Aspect")
    left.set_xlabel("Top X")
    left.set_ylabel("Right Y")
    left.set_xlim(-1, 5)
    left.set_ylim(-1, 5)
    left.xaxis.set_label_position("top")
    left.xaxis.tick_top()
    left.yaxis.set_label_position("right")
    left.yaxis.tick_right()
    left.set_aspect("equal", adjustable="box")
    left.set_box_aspect(1)
    left.minorticks_on()
    left.locator_params(axis="both", nbins=6)
    left.xaxis.set_minor_locator(MultipleLocator(0.2))
    left.yaxis.set_minor_locator(MultipleLocator(0.2))
    tick_color = (0.18, 0.42, 0.55, 1.0)
    left.tick_params(
        axis="both",
        which="major",
        colors=tick_color,
        length=px2pt(7),
        width=1.2,
    )
    left.tick_params(
        axis="both",
        which="minor",
        colors=tick_color,
        length=px2pt(4),
        width=0.9,
    )
    left.plot(
        [-0.5, 0.8, 2.2, 4.2],
        [-0.2, 1.0, 2.1, 4.4],
        color=(0.10, 0.32, 0.76),
        linewidth=2.0,
    )
    left.scatter(
        [0.0, 1.5, 3.5, 4.5],
        [0.0, 1.8, 3.2, 4.6],
        s=ss(8),
        c=[(0.92, 0.48, 0.20, 0.92)],
        edgecolors=[(0.52, 0.22, 0.08, 1.0)],
        linewidths=1.0,
    )

    right = fig.add_axes(go_rect(0.58, 0.14, 0.95, 0.78))
    right.set_title("Twin + Secondary")
    right.set_xlim(0, 10)
    right.set_ylim(0, 20)
    right.plot(
        [0, 2, 4, 6, 8, 10],
        [2, 6, 9, 13, 16, 19],
        color=(0.12, 0.45, 0.72),
        linewidth=2.0,
    )

    twin = right.twinx()
    twin.set_ylim(0, 100)
    twin.tick_params(axis="y", colors=(0.80, 0.22, 0.22, 1.0))
    twin.spines["right"].set_color((0.80, 0.22, 0.22, 1.0))
    twin.plot(
        [0, 2, 4, 6, 8, 10],
        [10, 22, 38, 58, 81, 96],
        color=(0.80, 0.22, 0.22),
        linewidth=1.8,
    )

    sec = right.secondary_xaxis("top", functions=(lambda x: x * 10, lambda x: x / 10))
    sec.tick_params(axis="x", colors=(0.16, 0.42, 0.30, 1.0))
    sec.spines["top"].set_color((0.16, 0.42, 0.30, 1.0))

    save(fig, out_dir, "axes_control_surface")

PLOT = axes_control_surface


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
