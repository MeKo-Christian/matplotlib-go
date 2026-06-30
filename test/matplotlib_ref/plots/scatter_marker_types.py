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

def scatter_marker_types(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.06, 0.08, 0.98, 0.92))
    ax.set_title("Marker Style Grid")
    ax.set_xlim(0.5, 7.5)
    ax.set_ylim(0.5, 6.5)
    from matplotlib.markers import (
        MarkerStyle,
        TICKLEFT, TICKRIGHT, TICKUP, TICKDOWN,
        CARETLEFT, CARETRIGHT, CARETUP, CARETDOWN,
        CARETLEFTBASE, CARETRIGHTBASE, CARETUPBASE, CARETDOWNBASE,
    )
    markers = [
        ".", ",", "o", "v", "^", "<", ">",
        "1", "2", "3", "4", "8", "s", "p",
        "P", "*", "h", "H", "+", "x", "X",
        "D", "d", "|", "_", TICKLEFT, TICKRIGHT,
        TICKUP, TICKDOWN, CARETLEFT, CARETRIGHT, CARETUP,
        CARETDOWN, CARETLEFTBASE, CARETRIGHTBASE, CARETUPBASE,
        CARETDOWNBASE, (5, 0, 18), (5, 1, 18), (6, 2, 0),
        "$f$", MarkerStyle("o", fillstyle="none"),
    ]
    edge = (0.05, 0.05, 0.05)
    for i, marker in enumerate(markers):
        x = i % 7 + 1
        y = 6 - i // 7
        color = TAB10[i % len(TAB10)]
        ax.scatter([x], [y], s=ss(8), c=[color], marker=marker,
                   linewidths=1.2, edgecolors=[edge])
    save(fig, out_dir, "scatter_marker_types")

PLOT = scatter_marker_types


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
