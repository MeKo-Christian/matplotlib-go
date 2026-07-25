#!/usr/bin/env python3
"""Matplotlib reference for a slope-defined infinite line."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def axline_slope(out_dir):
    Path(out_dir).mkdir(parents=True, exist_ok=True)

    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 6)
    ax.set_title("Slope-defined axline")
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.axline((2, 3), slope=0.75, color=(31 / 255, 119 / 255, 180 / 255), linewidth=1.5)

    save(fig, out_dir, "axline_slope")


PLOT = axline_slope


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
