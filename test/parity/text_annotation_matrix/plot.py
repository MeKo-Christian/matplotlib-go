#!/usr/bin/env python3
"""Focused text, annotation, and offset-box reference fixture."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def text_annotation_matrix(out_dir):
    from matplotlib.offsetbox import (
        AnnotationBbox,
        AnchoredText,
        AnchoredOffsetbox,
        DrawingArea,
        HPacker,
        OffsetImage,
        TextArea,
    )
    from mpl_toolkits.axes_grid1.anchored_artists import AnchoredDrawingArea, AnchoredSizeBar

    fig = make_fig_px(720, 420)
    ax = fig.add_axes(go_rect(0.08, 0.12, 0.94, 0.90))
    ax.set_title("Text + Annotation Matrix")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5)

    ax.text(
        0.55,
        4.25,
        "bold italic\nbbox",
        fontsize=13,
        color=(0.12, 0.20, 0.42, 1.0),
        ha="left",
        va="top",
        family="DejaVu Sans",
        fontstyle="italic",
        fontweight="bold",
        bbox={
            "boxstyle": "round,pad=0.3",
            "facecolor": (0.88, 0.92, 1.00, 0.85),
            "edgecolor": (0.20, 0.32, 0.68, 1.0),
            "linewidth": lw(1.0),
        },
    )
    ax.text(
        2.25,
        4.45,
        "rotated label",
        fontsize=12,
        color=(0.56, 0.24, 0.16, 1.0),
        ha="center",
        va="center",
        rotation=-28,
        bbox={
            "boxstyle": "square,pad=0.25",
            "facecolor": (1.00, 0.90, 0.78, 0.72),
            "edgecolor": (0.65, 0.35, 0.12, 1.0),
            "linewidth": lw(0.8),
        },
    )

    ax.annotate(
        "bbox arrow",
        xy=(2.1, 2.15),
        xytext=(72, -40),
        textcoords="offset points",
        fontsize=11,
        color=(0.15, 0.18, 0.22, 1.0),
        ha="center",
        va="center",
        bbox={
            "boxstyle": "square,pad=0.25",
            "facecolor": (0.92, 0.97, 0.92, 0.92),
            "edgecolor": (0.23, 0.48, 0.20, 1.0),
            "linewidth": lw(1.0),
        },
        arrowprops={
            "arrowstyle": "-|>,head_length=0.35,head_width=0.20",
            "connectionstyle": "arc3,rad=0.25",
            "color": (0.18, 0.39, 0.70, 1.0),
            "linewidth": lw(1.2),
            "mutation_scale": 9,
        },
    )
    ax.annotate(
        "clipped",
        xy=(-0.4, 4.7),
        xytext=(40, -20),
        textcoords="offset points",
        annotation_clip=True,
        arrowprops={"connectionstyle": "arc3,rad=0.25", "color": (0.7, 0.1, 0.1, 1.0)},
    )

    ab_text = AnnotationBbox(
        TextArea("TextArea box", textprops={"fontsize": 10, "color": (0.23, 0.15, 0.35, 1.0)}),
        (3.75, 3.05),
        xybox=(4.7, 3.9),
        boxcoords="data",
        box_alignment=(0.5, 0.5),
        pad=0.35,
        frameon=True,
        arrowprops={"arrowstyle": "-|>", "color": (0.45, 0.24, 0.62, 1.0), "linewidth": lw(1.25)},
        bboxprops={
            "facecolor": (0.96, 0.90, 1.00, 0.92),
            "edgecolor": (0.45, 0.24, 0.62, 1.0),
            "linewidth": lw(1.0),
        },
    )
    ax.add_artist(ab_text)

    img = np.zeros((10, 12, 4), dtype=float)
    for y in range(10):
        for x in range(12):
            img[y, x] = (0.90, 0.94, 1.0, 1.0) if (x + y) % 2 else (0.38, 0.59, 0.82, 1.0)
    ab_img = AnnotationBbox(
        OffsetImage(img, zoom=2.4),
        (3.75, 1.55),
        xybox=(4.7, 1.25),
        boxcoords="data",
        pad=0.25,
        frameon=True,
        arrowprops={"arrowstyle": "-|>", "color": (0.25, 0.25, 0.25, 1.0), "linewidth": lw(1.25)},
        bboxprops={
            "facecolor": (1.0, 1.0, 1.0, 0.95),
            "edgecolor": (0.25, 0.25, 0.25, 1.0),
            "linewidth": lw(1.0),
        },
    )
    ax.add_artist(ab_img)

    anchored_text = AnchoredText(
        "anchored\ntext",
        loc="upper left",
        pad=0.35,
        borderpad=0.6,
        prop={"size": 9, "color": (0.18, 0.18, 0.10, 1.0)},
        frameon=True,
    )
    anchored_text.patch.set_facecolor((0.98, 0.98, 0.94, 0.92))
    anchored_text.patch.set_edgecolor((0.45, 0.45, 0.24, 1.0))
    ax.add_artist(anchored_text)

    drawing = AnchoredDrawingArea(46, 28, 0, 0, loc="upper right", pad=0.4, borderpad=0.7, frameon=True)
    drawing.drawing_area.add_artist(mpatches.Polygon(
        [(6, 24), (40, 21), (24, 5)],
        closed=True,
        facecolor=(0.83, 0.44, 0.18, 0.85),
        edgecolor=(0.43, 0.20, 0.08, 1.0),
        linewidth=lw(1.0),
    ))
    drawing.patch.set_facecolor((0.98, 0.96, 0.88, 0.92))
    drawing.patch.set_edgecolor((0.45, 0.36, 0.14, 1.0))
    ax.add_artist(drawing)

    da = DrawingArea(18, 18)
    da.add_artist(mpatches.Polygon(
        [(9, 1), (17, 9), (9, 17), (1, 9)],
        closed=True,
        facecolor=(0.25, 0.62, 0.78, 0.9),
        edgecolor=(0.08, 0.25, 0.35, 1.0),
        linewidth=lw(1.0),
    ))
    packed = HPacker(
        children=[da, TextArea("HPack", textprops={"fontsize": 9, "color": (0.10, 0.24, 0.35, 1.0)})],
        align="center",
        pad=0,
        sep=5,
    )
    anchored = AnchoredOffsetbox(loc="lower left", child=packed, pad=0.4, borderpad=0.6, frameon=True)
    anchored.patch.set_facecolor((0.93, 0.97, 1.00, 0.92))
    anchored.patch.set_edgecolor((0.16, 0.37, 0.54, 1.0))
    ax.add_artist(anchored)

    sizebar = AnchoredSizeBar(
        ax.transData,
        1.25,
        "1.25 data",
        loc="lower right",
        pad=0.4,
        borderpad=0.6,
        sep=4,
        frameon=True,
        size_vertical=0.09,
        fill_bar=True,
        color=(0.08, 0.08, 0.08, 1.0),
        fontproperties={"size": 9},
    )
    sizebar.patch.set_facecolor((1.0, 1.0, 1.0, 0.82))
    sizebar.patch.set_edgecolor((0.20, 0.20, 0.20, 1.0))
    ax.add_artist(sizebar)

    save(fig, out_dir, "text_annotation_matrix")


PLOT = text_annotation_matrix


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
