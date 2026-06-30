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

def colorbar_composition(out_dir):
    fig, ax = plt.subplots(figsize=(1000 / DPI, 700 / DPI), dpi=DPI, constrained_layout=True)
    rows, cols = 80, 120
    data = np.zeros((rows, cols))
    for row in range(rows):
        for col in range(cols):
            x = (col / (cols - 1)) * 4 - 2
            y = (row / (rows - 1)) * 4 - 2
            r = math.hypot(x, y)
            data[row, col] = math.sin(3 * r) * math.exp(-0.6 * r)

    im = ax.imshow(data, cmap="inferno", origin="lower", extent=[0, cols, 0, rows], aspect="auto")
    ax.set_title("Heatmap with Colorbar")
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.set_xlim(0, cols)
    ax.set_ylim(0, rows)
    ax.set_yticks(np.arange(0, rows + 1, 20))
    ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
    cbar = fig.colorbar(im, ax=ax)
    cbar.set_label("Intensity")

    save(fig, out_dir, "colorbar_composition")

PLOT = colorbar_composition


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
