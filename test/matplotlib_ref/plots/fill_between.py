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

def fill_between(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Fill Between Curves")
    ax.set_xlim(0, 6.28)
    ax.set_ylim(-1.5, 1.5)
    n  = 50
    x  = [6.28 * i / (n - 1) for i in range(n)]
    y1 = [math.sin(v) for v in x]
    y2 = [0.8 * math.cos(v) for v in x]
    ax.fill_between(
        x, y1, y2,
        facecolor=(0.8, 0.3, 0.3, 0.6),
        edgecolor=(0.5, 0.1, 0.1, 1.0),
        linewidth=1.5,
    )
    ax.plot(x, y1, color=(1, 0, 0), linewidth=2)
    ax.plot(x, y2, color=(0, 0, 1), linewidth=2)
    save(fig, out_dir, "fill_between")

PLOT = fill_between


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
