#!/usr/bin/env python3
"""Matplotlib reference plot for stackplot weighted_wiggle (streamgraph) baseline."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


# Shared with test/parity/stackplot_streamgraph/plot.go.
STREAM_DATA = [
    [1, 2, 4, 6, 7, 6, 4, 3, 2, 2, 1, 1],
    [0.5, 1, 2, 3, 5, 7, 8, 7, 5, 3, 2, 1],
    [2, 2, 1, 1, 2, 3, 4, 6, 7, 6, 4, 2],
    [3, 2, 2, 1, 1, 1, 2, 3, 4, 5, 6, 5],
]


def stackplot_streamgraph(out_dir):
    Path(out_dir).mkdir(parents=True, exist_ok=True)

    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 11)
    ax.set_ylim(-13, 11)
    ax.set_title("Streamgraph (weighted_wiggle)")

    x = list(range(12))
    # Default property-cycle colors (tab10), matching Go's NextColor cycle.
    ax.stackplot(x, *STREAM_DATA, baseline="weighted_wiggle")

    save(fig, out_dir, "stackplot_streamgraph")


PLOT = stackplot_streamgraph


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
