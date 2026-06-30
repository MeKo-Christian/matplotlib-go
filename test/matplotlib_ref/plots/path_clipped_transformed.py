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


def path_clipped_transformed(out_dir):
    fig = make_fig_px(720, 420)

    ax = fig.add_axes(go_rect(0.12, 0.16, 0.90, 0.84))
    ax.set_title("Clipped Transformed Path")
    ax.set_xlabel("X")
    ax.set_ylabel("Y")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 6)
    ax.grid(True, color=(0.82, 0.82, 0.82, 1.0), linewidth=0.5)

    vertices = [
        (-1.6, 0.8),
        (1.0, 7.2),
        (4.2, -1.2),
        (6.3, 4.8),
        (8.1, 9.0),
        (12.0, 4.1),
        (10.8, -0.9),
        (8.0, 1.1),
        (5.8, 2.2),
        (3.8, 1.2),
        (1.2, 3.2),
        (-1.6, 0.8),
    ]
    codes = [
        mpath.Path.MOVETO,
        mpath.Path.CURVE4,
        mpath.Path.CURVE4,
        mpath.Path.CURVE4,
        mpath.Path.CURVE4,
        mpath.Path.CURVE4,
        mpath.Path.CURVE4,
        mpath.Path.LINETO,
        mpath.Path.CURVE4,
        mpath.Path.CURVE4,
        mpath.Path.CURVE4,
        mpath.Path.CLOSEPOLY,
    ]
    ax.add_patch(mpatches.PathPatch(
        mpath.Path(vertices, codes),
        facecolor=(0.12, 0.47, 0.71, 0.38),
        edgecolor=(0.05, 0.20, 0.36, 1.0),
        linewidth=1.7,
        transform=ax.transData,
        clip_on=True,
    ))
    ax.plot([0, 10], [0, 6], color=(0.84, 0.15, 0.16, 0.78), linewidth=1.0)

    save(fig, out_dir, "path_clipped_transformed")


PLOT = path_clipped_transformed


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
