#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import numpy as np

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

def histogram_variants(out_dir):
    fig = make_fig_px(840, 620)
    axes = {
        "counts": fig.add_axes(go_rect(0.08, 0.585, 0.46, 0.93)),
        "density": fig.add_axes(go_rect(0.57, 0.585, 0.96, 0.93)),
        "cumulative": fig.add_axes(go_rect(0.08, 0.10, 0.46, 0.445)),
        "multiple": fig.add_axes(go_rect(0.57, 0.10, 0.96, 0.445)),
    }
    black = (0, 0, 0, 1)

    # Raw counts.
    c_ax = axes["counts"]
    c_ax.set_title("Counts")
    c_ax.set_xlim(0, 10)
    c_ax.set_ylim(0, 70)
    data_c = normal_data(1, 0, 400, 5.0, 1.5)
    c_ax.hist(data_c, bins=18, color=(0.26, 0.53, 0.80, 0.85),
              edgecolor=black, linewidth=0.8, rwidth=1.0)

    # Density.
    d_ax = axes["density"]
    d_ax.set_title("Density")
    d_ax.set_xlim(0, 10)
    d_ax.set_ylim(0, 0.35)
    data_d = normal_data(42, 0, 500, 5.0, 1.5)
    d_ax.hist(data_d, bins=20, density=True, color=(0.20, 0.65, 0.30, 0.8),
              edgecolor=black, linewidth=0.8, rwidth=1.0)

    # Cumulative counts.
    cu_ax = axes["cumulative"]
    cu_ax.set_title("Cumulative")
    cu_ax.set_xlim(0, 10)
    cu_ax.set_ylim(0, 420)
    data_cu = normal_data(1, 0, 400, 5.0, 1.5)
    cu_ax.hist(data_cu, bins=18, cumulative=True, color=(0.55, 0.35, 0.75, 0.85),
               edgecolor=black, linewidth=0.8, rwidth=1.0)

    # Multiple overlapping probability histograms.
    m_ax = axes["multiple"]
    m_ax.set_title("Multiple (Probability)")
    m_ax.set_xlim(0, 11)
    m_ax.set_ylim(0, 0.25)
    data1 = normal_data(42, 0, 300, 4.0, 1.0)
    data2 = normal_data(7, 0, 300, 7.0, 1.2)
    m_ax.hist(data1, bins=15, density=False, weights=np.ones(len(data1)) / len(data1),
              color=(0.26, 0.53, 0.80, 0.6), edgecolor=black, linewidth=0.5, rwidth=1.0)
    m_ax.hist(data2, bins=15, density=False, weights=np.ones(len(data2)) / len(data2),
              color=(0.90, 0.50, 0.10, 0.6), edgecolor=black, linewidth=0.5, rwidth=1.0)

    save(fig, out_dir, "histogram_variants")

PLOT = histogram_variants


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
