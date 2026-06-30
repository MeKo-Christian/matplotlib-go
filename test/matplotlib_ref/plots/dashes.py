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

def dashes(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Dash Patterns")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 5)

    specs = [
        (4, [],               (0,   0,   0)),
        (3, [10, 4],          (0.8, 0,   0)),
        (2, [6, 2, 2, 2],     (0,   0.6, 0)),
        (1, [2, 2],           (0,   0,   0.8)),
    ]
    for y_val, pattern, color in specs:
        (line,) = ax.plot([1, 9], [y_val, y_val], color=color, linewidth=3)
        if pattern:
            line.set_dashes(pattern)

    save(fig, out_dir, "dashes")

PLOT = dashes


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
