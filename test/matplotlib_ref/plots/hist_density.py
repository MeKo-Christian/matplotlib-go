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

def hist_density(out_dir):
    """Density histogram matching renderHistDensity in golden_test.go."""
    data = normal_data(42, 0, 500, 5.0, 1.5)
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.12, 0.12, 0.95, 0.90))
    ax.set_title("Density Histogram")
    n, bins, patches = ax.hist(
        data,
        bins=20,
        density=True,
        color=(0.20, 0.65, 0.30, 0.8),
        edgecolor=(0, 0, 0, 1),
        linewidth=0.8,
        rwidth=1.0,
    )
    margin = 0.05 * (data.max() - data.min())
    ax.set_xlim(data.min() - margin, data.max() + margin)
    density_max = n.max()
    ax.set_ylim(0, density_max * 1.05)
    save(fig, out_dir, "hist_density")

PLOT = hist_density


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
