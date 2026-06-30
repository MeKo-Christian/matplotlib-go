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

def scatter_gallery(out_dir):
    fig = make_fig_px(840, 620)
    axes = {
        "cmap": fig.add_axes(go_rect(0.07, 0.585, 0.46, 0.93)),
        "size": fig.add_axes(go_rect(0.56, 0.585, 0.96, 0.93)),
        "alpha": fig.add_axes(go_rect(0.07, 0.10, 0.46, 0.445)),
        "families": fig.add_axes(go_rect(0.56, 0.10, 0.96, 0.445)),
    }

    # Colormapped scalar mapping.
    cmap_ax = axes["cmap"]
    cmap_ax.set_title("Colormapped")
    cmap_ax.set_xlim(0, 6)
    cmap_ax.set_ylim(0, 5)
    cmap_ax.scatter(
        [0.8, 1.8, 2.8, 3.8, 4.8],
        [1.2, 3.2, 2.2, 3.8, 1.8],
        c=[-1.0, -0.2, 0.35, 0.8, 1.2],
        cmap="viridis",
        s=[ss(7), ss(10), ss(13), ss(16), ss(19)],
        edgecolors=[(0.08, 0.08, 0.08, 1)],
        linewidths=1.5,
    )

    # Variable size, single color.
    size_ax = axes["size"]
    size_ax.set_title("Variable Size")
    size_ax.set_xlim(0, 10)
    size_ax.set_ylim(0, 9)
    size_ax.scatter(
        [1.5, 3.0, 4.5, 6.0, 7.5, 9.0],
        [5.0, 3.0, 6.0, 4.0, 7.0, 5.0],
        c=[(0.12, 0.47, 0.71)],
        s=[ss(5), ss(9), ss(13), ss(17), ss(21), ss(25)],
        edgecolors=[(0.08, 0.08, 0.08, 1)],
        linewidths=1.0,
    )

    # Alpha blending of overlapping clusters.
    alpha_ax = axes["alpha"]
    alpha_ax.set_title("Alpha Blending")
    alpha_ax.set_xlim(0, 8)
    alpha_ax.set_ylim(0, 8)
    alpha_ax.scatter(
        [2.5, 3.5, 3.0, 3.0],
        [4.0, 4.0, 4.8, 3.2],
        c=[(0.85, 0.20, 0.20)],
        s=ss(28),
        alpha=0.45,
        linewidths=0,
    )
    alpha_ax.scatter(
        [4.5, 5.5, 5.0, 5.0],
        [4.0, 4.0, 4.8, 3.2],
        c=[(0.20, 0.30, 0.85)],
        s=ss(28),
        alpha=0.45,
        linewidths=0,
    )

    # Marker families.
    fam_ax = axes["families"]
    fam_ax.set_title("Marker Families")
    fam_ax.set_xlim(0.5, 3.5)
    fam_ax.set_ylim(0.5, 2.5)
    markers = ["o", "s", "^", "D", "p", "*"]
    edge = (0.08, 0.08, 0.08, 1)
    for i, marker in enumerate(markers):
        x = i % 3 + 1
        y = 2 - i // 3
        color = TAB10[i % len(TAB10)]
        fam_ax.scatter([x], [y], s=ss(14), c=[color], marker=marker,
                       edgecolors=[edge], linewidths=1.2)

    save(fig, out_dir, "scatter_gallery")

PLOT = scatter_gallery


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
