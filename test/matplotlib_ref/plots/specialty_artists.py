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

def specialty_artists(out_dir):
    fig = make_fig_px(980, 720)
    axes = {
        "event": fig.add_axes(go_rect(0.07, 0.57, 0.34, 0.94)),
        "hexbin": fig.add_axes(go_rect(0.39, 0.57, 0.66, 0.94)),
        "pie": fig.add_axes(go_rect(0.73, 0.57, 0.98, 0.94)),
        "violin": fig.add_axes(go_rect(0.07, 0.08, 0.34, 0.45)),
        "table": fig.add_axes(go_rect(0.39, 0.08, 0.66, 0.45)),
        "sankey": fig.add_axes(go_rect(0.73, 0.08, 0.98, 0.45)),
    }

    event_ax = axes["event"]
    event_ax.set_title("Eventplot")
    event_ax.set_xlim(0, 10)
    event_ax.set_ylim(0.4, 3.6)
    event_ax.grid(True, axis="x", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    event_ax.set_axisbelow(True)
    event_ax.eventplot(
        [
            [0.8, 1.4, 3.1, 4.6, 7.3],
            [1.2, 2.9, 4.0, 6.4, 8.6],
            [0.5, 2.2, 5.4, 6.8, 9.1],
        ],
        lineoffsets=[1, 2, 3],
        linelengths=[0.6, 0.7, 0.8],
        colors=[
            (0.18, 0.44, 0.74, 1.0),
            (0.84, 0.38, 0.16, 1.0),
            (0.20, 0.63, 0.42, 1.0),
        ],
        linewidths=lw(1.5),
    )

    hex_ax = axes["hexbin"]
    hex_ax.set_title("Hexbin")
    hex_ax.set_xlim(0, 1)
    hex_ax.set_ylim(0, 1)
    hx = [0.08, 0.15, 0.21, 0.25, 0.34, 0.41, 0.48, 0.56, 0.63, 0.66, 0.74, 0.82, 0.88]
    hy = [0.14, 0.19, 0.24, 0.31, 0.46, 0.52, 0.61, 0.44, 0.73, 0.81, 0.68, 0.86, 0.58]
    hc = [1, 2, 1.5, 2.3, 2.8, 3.1, 3.6, 2.1, 4.5, 4.9, 3.8, 5.2, 4.1]
    hex_ax.hexbin(hx, hy, C=hc, gridsize=7, reduce_C_function=np.mean, mincnt=1, extent=(0, 1, 0, 1))

    pie_ax = axes["pie"]
    pie_ax.set_title("Pie")
    pie_ax.pie(
        [28, 22, 18, 32],
        labels=["Core", "I/O", "Render", "Docs"],
        autopct="%.0f%%",
        startangle=90,
        labeldistance=1.08,
        explode=[0, 0.04, 0, 0.02],
        colors=TAB10[:4],
        wedgeprops={"linewidth": lw(1.0), "edgecolor": "white"},
    )

    violin_ax = axes["violin"]
    violin_ax.set_title("Violin")
    violin_ax.set_xlim(0.4, 3.6)
    violin_ax.set_ylim(0.8, 5.2)
    violin_ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    violin_ax.set_axisbelow(True)
    parts = violin_ax.violinplot(
        [
            [1.2, 1.5, 1.7, 2.1, 2.4, 2.6, 2.9, 3.0, 3.2],
            [1.8, 2.0, 2.2, 2.5, 2.7, 3.0, 3.4, 3.8, 4.0],
            [2.4, 2.5, 2.7, 2.9, 3.1, 3.4, 3.7, 4.1, 4.6],
        ],
        showmeans=True,
        showmedians=False,
        showextrema=True,
    )
    for body in parts["bodies"]:
        body.set_facecolor(TAB10[0])
        body.set_edgecolor((0.20, 0.20, 0.20))
        body.set_alpha(0.45)

    table_ax = axes["table"]
    table_ax.set_title("Table")
    table_ax.axis("off")
    table = table_ax.table(
        cellText=[["Latency", "18ms", "14ms"], ["Throughput", "220/s", "265/s"]],
        rowLabels=["A", "B"],
        colLabels=["Metric", "Q1", "Q2"],
        bbox=[0.04, 0.18, 0.92, 0.64],
        cellLoc="center",
    )
    table.auto_set_font_size(False)
    table.set_fontsize(10)

    sankey_ax = axes["sankey"]
    sankey_ax.set_title("Sankey")
    sankey_ax.axis("off")
    sankey_ax.set_xlim(0, 1)
    sankey_ax.set_ylim(0, 1)
    trunk = mpatches.Rectangle(
        (0.18, 0.47),
        0.18,
        0.06,
        transform=sankey_ax.transAxes,
        facecolor=(0.12, 0.47, 0.71, 0.75),
        edgecolor=(0.10, 0.10, 0.10, 1.0),
        linewidth=lw(1.0),
    )
    sankey_ax.add_patch(trunk)
    flows = [
        ("Waste", -2, (0.84, 0.15, 0.16, 0.75)),
        ("CPU", 3, (0.17, 0.63, 0.17, 0.75)),
        ("Cache", 1.5, (1.00, 0.50, 0.05, 0.75)),
    ]
    for idx, (label, flow, color) in enumerate(flows):
        width = abs(flow) * 0.018
        y = 0.40 + idx * 0.095
        x0 = 0.36
        x1 = 0.66
        verts = [
            (x0, 0.50 - width / 2),
            (x0 + 0.10, y - width / 2),
            (x1, y - width / 2),
            (x1, y + width / 2),
            (x0 + 0.10, y + width / 2),
            (x0, 0.50 + width / 2),
            (x0, 0.50 - width / 2),
        ]
        path = mpath.Path(verts)
        sankey_ax.add_patch(mpatches.PathPatch(path, transform=sankey_ax.transAxes, facecolor=color, edgecolor=(0.10, 0.10, 0.10), linewidth=lw(1.0)))
        sankey_ax.text(0.70, y, label, transform=sankey_ax.transAxes, va="center", fontsize=10)

    save(fig, out_dir, "specialty_artists")

PLOT = specialty_artists


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
