#!/usr/bin/env python3
"""Projection and toolkit web demo reference module."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.webdemo_common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from webdemo_common import *  # noqa: F401,F403


def demo_toolkit(out_dir, width, height):
    fig = make_fig(width, height)
    ax_polar = fig.add_axes([0.07, 0.12, 0.28, 0.76], projection="polar")
    theta = np.linspace(0, 2 * math.pi, 160)
    ax_polar.set_title("Polar")
    ax_polar.plot(theta, 1.0 + 0.25 * np.sin(4 * theta), color=color(0.16, 0.42, 0.82), linewidth=lw(2.0))

    ax_geo = fig.add_axes([0.40, 0.52, 0.24, 0.34], projection="aitoff")
    ax_geo.set_title("Aitoff")
    ax_geo.grid(True)

    ax_radar = fig.add_axes([0.70, 0.12, 0.26, 0.76], projection="polar")
    ax_radar.set_title("Radar-style")
    values = np.array([0.7, 0.95, 0.62, 0.84, 0.76, 0.7])
    angles = np.linspace(0, 2 * math.pi, len(values))
    ax_radar.plot(angles, values, color=color(0.84, 0.35, 0.18), linewidth=lw(2.0))
    ax_radar.fill(angles, values, color=color(0.84, 0.35, 0.18, 0.25))

    ax_inset = fig.add_axes([0.42, 0.14, 0.22, 0.24])
    ax_inset.set_title("Inset")
    x = np.linspace(0, 8, 120)
    ax_inset.plot(x, np.sin(x), color=color(0.20, 0.58, 0.34), linewidth=lw(2.0))
    ax_inset.grid(True)

    save(fig, out_dir, "toolkit")


DEMO = demo_toolkit


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
