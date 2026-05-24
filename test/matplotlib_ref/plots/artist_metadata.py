#!/usr/bin/env python3
"""Matplotlib reference plot for shared artist metadata parity."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import matplotlib.path as mpath
import matplotlib.patches as mpatches
import matplotlib.transforms as mtransforms

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def artist_metadata(out_dir):
    fig = make_fig_px(540, 360)
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.88, 0.86))
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 10)
    ax.set_xticks([])
    ax.set_yticks([])

    background_path = mpath.Path(
        [(0.08, 0.12), (0.92, 0.12), (0.92, 0.88), (0.08, 0.88), (0.08, 0.12)],
        [mpath.Path.MOVETO, mpath.Path.LINETO, mpath.Path.LINETO, mpath.Path.LINETO, mpath.Path.CLOSEPOLY],
    )
    ax.add_patch(
        mpatches.PathPatch(
            background_path,
            transform=ax.transAxes,
            facecolor=(0.16, 0.68, 0.43, 0.65),
            edgecolor=(0.02, 0.25, 0.14, 1),
            linewidth=lw(1),
        )
    )

    alpha_line, = ax.plot(
        [0.6, 9.4],
        [1.4, 8.6],
        color=(0.10, 0.26, 0.78, 1),
        linewidth=lw(8),
        alpha=0.45,
        solid_capstyle="butt",
    )
    alpha_line.set_alpha(0.45)

    hidden, = ax.plot(
        [0.6, 9.4],
        [8.7, 1.3],
        color=(1.0, 0.0, 0.75, 1),
        linewidth=lw(16),
        solid_capstyle="butt",
    )
    hidden.set_visible(False)

    clipped, = ax.plot(
        [0.4, 9.6],
        [5.0, 5.0],
        color=(0.86, 0.12, 0.10, 0.70),
        linewidth=lw(18),
        solid_capstyle="butt",
    )
    lo = ax.transData.transform((2.2, 4.45))
    hi = ax.transData.transform((7.8, 5.55))
    clipped.set_clip_box(mtransforms.Bbox.from_extents(lo[0], lo[1], hi[0], hi[1]))

    save(fig, out_dir, "artist_metadata")


PLOT = artist_metadata


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
