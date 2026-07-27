#!/usr/bin/env python3
"""Matplotlib reference plot for the specialty-depth coverage cases."""

from __future__ import annotations

from pathlib import Path
import argparse
import inspect
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def specialty_depth(out_dir):
    fig = make_fig_px(980, 720)

    err_ax = fig.add_axes(go_rect(0.07, 0.60, 0.34, 0.94))
    err_ax.set_title("ErrorBar limits")
    err_ax.set_xlim(0, 5)
    err_ax.set_ylim(0, 5)
    err_ax.grid(axis="y")
    err_ax.errorbar(
        [1, 2, 3, 4],
        [1.2, 2.5, 3.1, 3.7],
        xerr=[[0.25, 0.35, 0.20, 0.30], [0.45, 0.25, 0.35, 0.20]],
        yerr=[[0.35, 0.50, 0.30, 0.60], [0.55, 0.30, 0.65, 0.40]],
        lolims=[False, True, False, False],
        uplims=[False, False, False, True],
        xlolims=[True, False, False, False],
        xuplims=[False, False, True, False],
        fmt="o",
        color=(0.12, 0.35, 0.70),
        ecolor=(0.12, 0.35, 0.70),
        elinewidth=1.4,
        capsize=8,
        markersize=4,
    )

    box_ax = fig.add_axes(go_rect(0.39, 0.60, 0.66, 0.94))
    box_ax.set_title("BoxPlot depth")
    box_ax.set_xlim(0.4, 2.6)
    box_ax.set_ylim(0, 8)
    box_ax.grid(axis="y")
    box = box_ax.boxplot(
        [
            [1.1, 1.8, 2.2, 2.6, 2.9, 3.1, 3.7, 6.8],
            [2.4, 3.1, 3.7, 4.3, 4.8, 5.2, 5.9, 7.2],
        ],
        notch=True,
        whis=(5, 95),
        conf_intervals=[[2.45, 3.10], [4.25, 5.00]],
        usermedians=[2.8, 4.6],
        patch_artist=True,
        flierprops={"marker": "D", "markersize": 4},
    )
    for patch, color in zip(box["boxes"], [(0.45, 0.65, 0.90, 0.78), (0.90, 0.55, 0.28, 0.78)]):
        patch.set_facecolor(color)

    violin_ax = fig.add_axes(go_rect(0.73, 0.60, 0.97, 0.94))
    violin_ax.set_title("Violin side")
    violin_ax.set_xlim(0.5, 5.5)
    violin_ax.set_ylim(0.5, 2.4)
    violin_ax.grid(axis="x")
    violin_kwargs = {
        "side": "high",
    } if "side" in inspect.signature(violin_ax.violinplot).parameters else {}
    parts = violin_ax.violinplot(
        [
            [1.0, 1.3, 1.7, 2.2, 2.5, 3.1, 3.7, 4.1, 4.7],
            [1.4, 1.8, 2.2, 2.8, 3.2, 3.5, 4.2, 4.8, 5.1],
        ],
        vert=False,
        quantiles=[[0.25, 0.75], [0.25, 0.75]],
        bw_method="scott",
        showmedians=True,
        showextrema=True,
        **violin_kwargs,
    )
    if not violin_kwargs:
        _clip_horizontal_violin_high(parts, [1, 2])
    for body in parts["bodies"]:
        body.set_facecolor((0.30, 0.60, 0.78, 0.58))
        body.set_edgecolor((0.12, 0.12, 0.12, 0.9))

    pie_ax = fig.add_axes(go_rect(0.08, 0.10, 0.34, 0.45))
    pie_ax.set_title("Pie labels")
    wedges, texts = pie_ax.pie(
        [0.22, 0.18, 0.30],
        labels=["Alpha", "Beta", "Gamma"],
        normalize=False,
        rotatelabels=True,
        shadow=True,
        startangle=30,
        colors=[(0.22, 0.55, 0.75), (0.90, 0.45, 0.18), (0.32, 0.64, 0.34)],
        wedgeprops={"linewidth": 1.0, "edgecolor": "white"},
    )
    for wedge, hatch in zip(wedges, ["/", "x", "\\"]):
        wedge.set_hatch(hatch)
    for wedge, label in zip(wedges, ["22%", "18%", "30%"]):
        angle = 0.5 * (wedge.theta1 + wedge.theta2)
        x = 0.62 * math.cos(math.radians(angle))
        y = 0.62 * math.sin(math.radians(angle))
        pie_ax.text(x, y, label, ha="center", va="center")

    hex_ax = fig.add_axes(go_rect(0.43, 0.10, 0.93, 0.45))
    hex_ax.set_title("Hexbin log + marginals")
    hex_ax.set_xlim(1, 120)
    hex_ax.set_ylim(1, 120)
    hex_ax.hexbin(
        [1.2, 1.8, 2.6, 4.0, 6.5, 9.0, 14, 22, 35, 58, 92],
        [1.1, 2.2, 3.0, 5.5, 7.0, 12, 20, 28, 48, 80, 105],
        C=[1, 3, 2, 5, 7, 6, 11, 14, 18, 23, 30],
        gridsize=6,
        reduce_C_function=np.max,
        bins="log",
        xscale="log",
        yscale="log",
        marginals=True,
    )

    save(fig, out_dir, "specialty_depth")


PLOT = specialty_depth


def _clip_horizontal_violin_high(parts, positions):
    for body, pos in zip(parts.get("bodies", []), positions):
        vertices = body.get_paths()[0].vertices
        vertices[:, 1] = np.maximum(vertices[:, 1], pos)
    for key in ("cmedians", "cmins", "cmaxes", "cquantiles"):
        collection = parts.get(key)
        if collection is None:
            continue
        segments = collection.get_segments()
        clipped = []
        for segment in segments:
            pos = min(positions, key=lambda p: abs(np.mean(segment[:, 1]) - p))
            segment = segment.copy()
            segment[:, 1] = np.maximum(segment[:, 1], pos)
            clipped.append(segment)
        collection.set_segments(clipped)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
