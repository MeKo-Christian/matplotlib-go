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


def basic_line_labels(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.15, 0.95, 0.9))
    ax.set_title("Basic Line")
    ax.set_xlabel("Time")
    ax.set_ylabel("Signal")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 1)
    ax.plot(
        [0, 1, 3, 6, 10],
        [0, 0.2, 0.9, 0.4, 0.8],
        color="black",
        linewidth=2,
    )
    save(fig, out_dir, "basic_line_labels")


PLOT = basic_line_labels


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
