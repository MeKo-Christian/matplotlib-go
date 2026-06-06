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

def lines_markers_gallery(out_dir):
    from matplotlib.markers import MarkerStyle

    fig = make_fig_px(840, 620)
    axes = {
        "dash": fig.add_axes(go_rect(0.07, 0.585, 0.46, 0.93)),
        "joins": fig.add_axes(go_rect(0.56, 0.585, 0.96, 0.93)),
        "markers": fig.add_axes(go_rect(0.07, 0.10, 0.46, 0.445)),
        "legend": fig.add_axes(go_rect(0.56, 0.10, 0.96, 0.445)),
    }

    # Dash patterns.
    dash_ax = axes["dash"]
    dash_ax.set_title("Dash Patterns")
    dash_ax.set_xlim(0, 10)
    dash_ax.set_ylim(0, 5)
    specs = [
        (4, [],            (0,   0,   0)),
        (3, [10, 4],       (0.8, 0,   0)),
        (2, [6, 2, 2, 2],  (0,   0.6, 0)),
        (1, [2, 2],        (0,   0,   0.8)),
    ]
    for y_val, pattern, color in specs:
        (line,) = dash_ax.plot([1, 9], [y_val, y_val], color=color, linewidth=lw(3))
        if pattern:
            line.set_dashes(pattern)

    # Line joins and caps.
    joins_ax = axes["joins"]
    joins_ax.set_title("Line Joins and Caps")
    joins_ax.set_xlim(0, 10)
    joins_ax.set_ylim(0, 6)
    joins_ax.plot([1, 3, 3, 5], [5, 5, 3, 3], color=(0.8, 0.2, 0.2), linewidth=lw(8))
    joins_ax.plot([7, 9], [5, 5], color=(0.2, 0.2, 0.8), linewidth=lw(8))

    # Marker grid + open-fill marker.
    marker_ax = axes["markers"]
    marker_ax.set_title("Marker Grid + Fill Styles")
    marker_ax.set_xlim(0.5, 6.5)
    marker_ax.set_ylim(0.5, 2.5)
    markers = [
        "o", "s", "^", "D", "p", "*",
        "h", "8", "P", "X", "d", MarkerStyle("o", fillstyle="none"),
    ]
    edge = (0.05, 0.05, 0.05)
    for i, marker in enumerate(markers):
        x = i % 6 + 1
        y = 2 - i // 6
        color = TAB10[i % len(TAB10)]
        marker_ax.scatter([x], [y], s=ss(9), c=[color], marker=marker,
                          linewidths=lw(1.2), edgecolors=[edge])

    # Multi-series legend.
    legend_ax = axes["legend"]
    legend_ax.set_title("Multi-Series Legend")
    legend_ax.set_xlim(0, 6)
    legend_ax.set_ylim(0, 6)
    xs = [0.5, 1.5, 2.5, 3.5, 4.5, 5.5]
    series = [
        ("rising",  TAB10[0], [0.8, 1.6, 2.5, 3.3, 4.2, 5.0]),
        ("falling", TAB10[1], [5.2, 4.4, 3.7, 2.9, 2.1, 1.3]),
        ("wave",    TAB10[2], [2.6, 3.6, 2.8, 3.8, 3.0, 4.0]),
    ]
    for label, color, ys in series:
        legend_ax.plot(xs, ys, color=color, linewidth=lw(2), label=label)
    legend_ax.legend(loc="upper right")

    save(fig, out_dir, "lines_markers_gallery")

PLOT = lines_markers_gallery


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
