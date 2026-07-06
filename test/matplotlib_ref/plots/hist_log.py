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

def hist_log(out_dir):
    """Log-scale count histogram matching examples/hist_log."""
    data = normal_data(42, 0, 500, 5.0, 1.5)
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.12, 0.12, 0.95, 0.90))
    ax.set_title("Logarithmic Histogram")
    ax.hist(
        data,
        bins="sturges",
        color=(0.26, 0.53, 0.80, 0.8),
        edgecolor=(0, 0, 0, 1),
        linewidth=0.8,
        rwidth=1.0,
        log=True,
    )
    # match AutoScale(0.05) x margin; the log y axis uses a fixed positive window.
    margin = 0.05 * (data.max() - data.min())
    ax.set_xlim(data.min() - margin, data.max() + margin)
    ax.set_ylim(1, 200)
    save(fig, out_dir, "hist_log")

PLOT = hist_log


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
