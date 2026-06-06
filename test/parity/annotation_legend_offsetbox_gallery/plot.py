#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import argparse
import math
import sys

import matplotlib.offsetbox as mbox
import matplotlib.patches as mpatches
import matplotlib.path as mpath
import numpy as np
from mpl_toolkits.axes_grid1.anchored_artists import AnchoredSizeBar

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def annotation_legend_offsetbox_gallery(out_dir):
    fig = make_fig_px(1040, 720)
    fig.text(0.05, 0.95, "Annotations, Legends, and Offset Boxes", fontsize=13)

    top = fig.add_axes(go_rect(0.08, 0.55, 0.94, 0.88))
    bottom = fig.add_axes(go_rect(0.08, 0.10, 0.94, 0.41))
    add_annotation_legend_panel(top)
    add_offset_box_panel(bottom)
    save(fig, out_dir, "annotation_legend_offsetbox_gallery")


def add_annotation_legend_panel(ax):
    ax.set_title("Annotation and Legend Layout")
    ax.set_xlim(0, 8)
    ax.set_ylim(-1.25, 1.25)
    ax.set_xlabel("x")
    ax.set_ylabel("signal")

    x = np.linspace(0, 8, 160)
    blue = (0.12, 0.31, 0.68, 1)
    orange = (0.86, 0.43, 0.16, 1)
    ax.plot(x, np.sin(x), color=blue, linewidth=2, label="sin(x)")
    ax.plot(x, 0.65 * np.cos(x * 0.8), color=orange, linewidth=2, label="0.65 cos(0.8x)")

    ax.annotate(
        "curved arrow\nbbox label",
        xy=(math.pi / 2, 1),
        xytext=(68, -42),
        textcoords="offset points",
        fontsize=10,
        ha="center",
        va="center",
        arrowprops={
            "arrowstyle": "-|>,head_length=0.35,head_width=0.20",
            "connectionstyle": "arc3,rad=0.25",
            "color": blue,
            "linewidth": 1.2,
        },
        bbox=text_box(0.28, (0.92, 0.97, 1.00, 0.90), blue),
    )

    text_area = mbox.TextArea("offset box", textprops={"fontsize": 10})
    annotation_box = mbox.AnnotationBbox(
        text_area,
        (5.65, -0.25),
        xybox=(6.75, 0.55),
        xycoords="data",
        boxcoords="data",
        box_alignment=(0.5, 0.5),
        pad=0.3,
        bboxprops={"facecolor": (0.96, 0.92, 1.00, 0.92), "edgecolor": (0.42, 0.25, 0.60, 1)},
        arrowprops={"arrowstyle": "->", "color": (0.42, 0.25, 0.60, 1)},
    )
    ax.add_artist(annotation_box)

    handles, labels = ax.get_legend_handles_labels()
    collected = ax.legend(handles, labels, loc="upper right", title="Handles", ncols=2)
    ax.add_artist(collected)
    proxy = mpatches.Patch(
        label="proxy patch",
        facecolor=(0.94, 0.78, 0.38, 0.95),
        edgecolor=(0.46, 0.30, 0.08, 1),
        linewidth=1.1,
        hatch="//",
    )
    ax.legend(handles=[proxy], labels=["proxy patch"], loc="lower right", frameon=False)


def add_offset_box_panel(ax):
    ax.set_title("Anchored Offset Boxes")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 4)
    ax.set_xticks([])
    ax.set_yticks([])

    anchored = mbox.AnchoredText("AnchoredText\nupper left", loc="upper left", pad=0.4, borderpad=0.8, prop={"size": 10})
    anchored.patch.set_facecolor((0.98, 0.98, 0.92, 0.94))
    anchored.patch.set_edgecolor((0.45, 0.42, 0.16, 1))
    ax.add_artist(anchored)

    drawing = mbox.DrawingArea(58, 34, 0, 0)
    drawing.add_artist(mpatches.PathPatch(local_triangle_path(), facecolor=(0.84, 0.44, 0.18, 0.88), edgecolor=(0.43, 0.20, 0.08, 1), linewidth=1))
    anchored_drawing = mbox.AnchoredOffsetbox(
        loc="upper right",
        child=drawing,
        pad=0.4,
        borderpad=0.8,
        frameon=True,
    )
    anchored_drawing.patch.set_facecolor((0.98, 0.94, 0.88, 0.94))
    anchored_drawing.patch.set_edgecolor((0.55, 0.30, 0.12, 1))
    ax.add_artist(anchored_drawing)

    da = mbox.DrawingArea(18, 18, 0, 0)
    da.add_artist(mpatches.PathPatch(local_diamond_path(), facecolor=(0.25, 0.62, 0.78, 0.9), edgecolor=(0.08, 0.25, 0.35, 1), linewidth=1))
    image_box = mbox.OffsetImage(small_image(), zoom=1.35)
    text_box_artist = mbox.TextArea("HPacker", textprops={"fontsize": 10, "color": (0.10, 0.24, 0.35, 1)})
    packer = mbox.HPacker(children=[da, image_box, text_box_artist], align="center", pad=0, sep=6)
    anchored_packer = mbox.AnchoredOffsetbox(loc="lower left", child=packer, pad=0.4, borderpad=0.8, frameon=True)
    anchored_packer.patch.set_facecolor((0.92, 0.97, 1.00, 0.94))
    anchored_packer.patch.set_edgecolor((0.16, 0.37, 0.54, 1))
    ax.add_artist(anchored_packer)

    sizebar = AnchoredSizeBar(
        ax.transData,
        1.4,
        "1.4 data",
        loc="lower right",
        pad=0.4,
        borderpad=0.8,
        sep=4,
        size_vertical=0.10,
        fill_bar=True,
        frameon=True,
        fontproperties={"size": 10},
    )
    sizebar.patch.set_facecolor((1, 1, 1, 0.86))
    sizebar.patch.set_edgecolor((0.20, 0.20, 0.20, 1))
    ax.add_artist(sizebar)


def text_box(pad, face, edge):
    return {"boxstyle": f"round,pad={pad}", "facecolor": face, "edgecolor": edge, "linewidth": 0.9}


def small_image():
    img = np.zeros((10, 12, 4), dtype=float)
    for y in range(10):
        for x in range(12):
            if (x + y) % 2 == 0:
                img[y, x] = (96 / 255, 150 / 255, 209 / 255, 1)
            else:
                img[y, x] = (231 / 255, 242 / 255, 255 / 255, 1)
    return img


def local_triangle_path():
    return mpath.Path([(7, 28), (50, 25), (30, 6), (7, 28)], [mpath.Path.MOVETO, mpath.Path.LINETO, mpath.Path.LINETO, mpath.Path.CLOSEPOLY])


def local_diamond_path():
    return mpath.Path([(9, 1), (17, 9), (9, 17), (1, 9), (9, 1)], [mpath.Path.MOVETO, mpath.Path.LINETO, mpath.Path.LINETO, mpath.Path.LINETO, mpath.Path.CLOSEPOLY])


PLOT = annotation_legend_offsetbox_gallery


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
