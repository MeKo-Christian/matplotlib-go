#!/usr/bin/env python3
"""Split Matplotlib web demo module generated from test/matplotlib_ref/webdemo.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.webdemo_common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from webdemo_common import *  # noqa: F401,F403


def demo_axes(out_dir, width, height):
    fig = make_fig(width, height)

    left = fig.add_axes(rect(0.08, 0.14, 0.42, 0.84))
    left.set_title("Top/Right Axes and 0.1 Minor Ticks")
    left.set_xlabel("x")
    left.set_ylabel("y")
    left.set_xlim(0, 5)
    left.set_ylim(0, 5)
    left.grid(True)
    left.xaxis.set_label_position("top")
    left.xaxis.tick_top()
    left.yaxis.set_label_position("right")
    left.yaxis.tick_right()
    left.xaxis.set_major_locator(MultipleLocator(1))
    left.yaxis.set_major_locator(MultipleLocator(1))
    left.xaxis.set_major_formatter(ScalarFormatter())
    left.yaxis.set_major_formatter(ScalarFormatter())
    left.xaxis.set_minor_locator(MultipleLocator(0.1))
    left.yaxis.set_minor_locator(MultipleLocator(0.1))
    left.set_aspect("equal", adjustable="box")
    left.set_box_aspect(1)
    left.plot([0.2, 1.1, 2.2, 3.3, 4.5], [0.1, 1.0, 1.9, 3.1, 4.4], color=color(0.10, 0.32, 0.76), linewidth=lw(2))
    left.scatter([0.3, 1.5, 2.7, 4.0], [0.2, 1.4, 2.6, 3.9], color=color(0.92, 0.48, 0.20, 0.92), edgecolor=color(0.52, 0.22, 0.08), linewidth=lw(1), s=ss(8))

    right = fig.add_axes(rect(0.56, 0.14, 0.94, 0.82))
    right.set_title("Growth, Twin Rate, and Weeks")
    right.set_xlabel("days")
    right.set_ylabel("active accounts")
    right.set_xlim(0, 28)
    right.set_ylim(1, 100)
    right.set_yscale("log")
    right.xaxis.set_major_locator(MultipleLocator(4))
    right.xaxis.set_major_formatter(ScalarFormatter())
    right.yaxis.set_major_formatter(ScalarFormatter())
    right.yaxis.set_minor_locator(LogLocator(10, subs=[2, 3, 4, 5, 6, 7, 8, 9]))
    right.grid(True)
    right.plot([0, 4, 8, 12, 16, 20, 24, 28], [1.5, 2.6, 4.8, 9.5, 18, 33, 61, 96], color=color(0.12, 0.45, 0.72), linewidth=lw(2))

    twin = right.twinx()
    twin.set_ylim(0, 100)
    twin.set_ylabel("conversion rate (%)")
    twin.yaxis.set_label_position("right")
    twin.tick_params(axis="y", colors=color(0.80, 0.22, 0.22, 1.0))
    twin.spines["right"].set_color((0.80, 0.22, 0.22, 1.0))
    twin.plot([0, 4, 8, 12, 16, 20, 24, 28], [12, 18, 24, 35, 49, 61, 74, 88], color=color(0.80, 0.22, 0.22), linewidth=lw(1.8))

    sec = right.secondary_xaxis("top", functions=(lambda x: x / 7, lambda x: x * 7))
    sec.xaxis.set_major_locator(MultipleLocator(1))
    sec.xaxis.set_major_formatter(ScalarFormatter())
    sec.spines["top"].set_color((0.16, 0.42, 0.30, 1.0))

    save(fig, out_dir, "axes")


DEMO = demo_axes


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    DEMO(args.output_dir)


if __name__ == "__main__":
    main()
