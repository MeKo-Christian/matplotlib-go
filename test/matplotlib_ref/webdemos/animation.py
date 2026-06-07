#!/usr/bin/env python3
"""Animation web demo reference module."""

from __future__ import annotations

from pathlib import Path
import argparse
import math
import sys

try:
    from test.matplotlib_ref.webdemo_common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from webdemo_common import *  # noqa: F401,F403


def _sine_frame(frame):
    xs = np.linspace(0, 2 * math.pi, 96)
    phase = frame * 0.35
    ys = np.sin(xs + phase) * (0.75 + 0.02 * frame)
    return xs, ys


def demo_animation(out_dir, width, height):
    fig = make_fig(width, height)

    func_ax = fig.add_axes([0.08, 0.16, 0.40, 0.74])
    func_ax.set_title("FuncAnimation frames")
    func_ax.set_xlabel("phase")
    func_ax.set_ylabel("signal")
    func_ax.set_xlim(0, 2 * math.pi)
    func_ax.set_ylim(-1.35, 1.35)
    func_ax.grid(True, axis="y")
    for frame in [0, 4, 8]:
        xs, ys = _sine_frame(frame)
        if frame == 8:
            func_ax.plot(xs, ys, color=color(0.12, 0.34, 0.68), linewidth=lw(2.3))
        else:
            func_ax.plot(xs, ys, color=color(0.28, 0.28, 0.28, 0.45), linewidth=lw(1.2))

    artist_ax = fig.add_axes([0.58, 0.16, 0.36, 0.74])
    artist_ax.set_title("ArtistAnimation frames")
    artist_ax.set_xlabel("frame")
    artist_ax.set_ylabel("value")
    artist_ax.set_xlim(-0.5, 3.5)
    artist_ax.set_ylim(0, 3.5)
    artist_ax.grid(True, axis="y")
    rows = [
        ([0, 1, 2, 3], [0.7, 1.4, 0.9, 1.8], color(0.84, 0.35, 0.18), "frame 0"),
        ([0, 1, 2, 3], [1.3, 2.2, 1.6, 2.7], color(0.20, 0.58, 0.34), "frame 1"),
        ([0, 1, 2, 3], [2.0, 2.8, 2.2, 3.1], color(0.55, 0.32, 0.72), "frame 2"),
    ]
    for xs, ys, col, label in rows:
        artist_ax.plot(xs, ys, color=col, linewidth=lw(2.1), label=label)
    artist_ax.legend()

    save(fig, out_dir, "animation")


DEMO = demo_animation


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
