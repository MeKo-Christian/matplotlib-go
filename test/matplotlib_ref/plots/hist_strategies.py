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

def hist_strategies(out_dir):
    """Two overlapping probability histograms matching renderHistStrategies."""
    data1 = normal_data(42, 0, 300, 4.0, 1.0)
    data2 = normal_data(7, 0, 300, 7.0, 1.2)
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.12, 0.12, 0.95, 0.90))
    ax.set_title("Histogram Strategies")
    n1, _, _ = ax.hist(
        data1, bins=15, density=False, weights=np.ones(len(data1)) / len(data1),
        color=(0.26, 0.53, 0.80, 0.6),
        edgecolor=(0, 0, 0, 1),
        linewidth=0.5,
        rwidth=1.0,
    )
    n2, _, _ = ax.hist(
        data2, bins=15, density=False, weights=np.ones(len(data2)) / len(data2),
        color=(0.90, 0.50, 0.10, 0.6),
        edgecolor=(0, 0, 0, 1),
        linewidth=0.5,
        rwidth=1.0,
    )
    all_data = np.concatenate([data1, data2])
    margin = 0.05 * (all_data.max() - all_data.min())
    ax.set_xlim(all_data.min() - margin, all_data.max() + margin)
    ax.set_ylim(0, max(n1.max(), n2.max()) * 1.05)
    save(fig, out_dir, "hist_strategies")

PLOT = hist_strategies


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
