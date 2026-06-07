#!/usr/bin/env python3
"""AxisArtist web demo reference module."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.webdemo_common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from webdemo_common import *  # noqa: F401,F403


def demo_axisartist(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect(0.12, 0.14, 0.92, 0.86))
    ax.set_title("AxisArtist-style Spines")
    ax.set_xlabel("centered x")
    ax.set_ylabel("centered y")
    ax.spines["left"].set_position(("data", 0))
    ax.spines["bottom"].set_position(("data", 0))
    ax.spines["right"].set_visible(False)
    ax.spines["top"].set_visible(False)
    ax.grid(True, color=color(0.75, 0.75, 0.78, 0.65))
    x = np.linspace(-3, 3, 180)
    ax.plot(x, np.sin(x), color=color(0.16, 0.42, 0.82), linewidth=lw(2.4), label="sin")
    ax.plot(x, 0.5 * np.cos(2 * x), color=color(0.84, 0.35, 0.18), linewidth=lw(2.0), label="cos")
    ax.legend()
    save(fig, out_dir, "axisartist")


DEMO = demo_axisartist


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
