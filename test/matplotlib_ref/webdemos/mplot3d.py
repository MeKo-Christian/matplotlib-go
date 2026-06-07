#!/usr/bin/env python3
"""mplot3d web demo reference module."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.webdemo_common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from webdemo_common import *  # noqa: F401,F403


def demo_mplot3d(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes([0.08, 0.10, 0.56, 0.82], projection="3d")
    x = np.linspace(-2.5, 2.5, 36)
    y = np.linspace(-2.5, 2.5, 32)
    xx, yy = np.meshgrid(x, y)
    zz = np.sin(xx * yy) / (1.0 + 0.2 * (xx * xx + yy * yy))
    ax.set_title("3D Surface")
    ax.plot_surface(xx, yy, zz, cmap="viridis", linewidth=0, antialiased=True)
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.set_zlabel("z")

    ax2 = fig.add_axes([0.70, 0.18, 0.25, 0.64], projection="3d")
    t = np.linspace(0, 4 * math.pi, 90)
    ax2.set_title("3D Line")
    ax2.plot(np.cos(t), np.sin(t), t / (4 * math.pi), color=color(0.84, 0.35, 0.18), linewidth=lw(2.0))
    ax2.scatter(np.cos(t[::12]), np.sin(t[::12]), t[::12] / (4 * math.pi), color=color(0.16, 0.42, 0.82), s=ss(5))

    save(fig, out_dir, "mplot3d")


DEMO = demo_mplot3d


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
