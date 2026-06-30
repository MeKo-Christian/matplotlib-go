#!/usr/bin/env python3
"""Matplotlib reference for image alpha parity."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def wave_image_data(rows, cols):
    y = np.linspace(0, 1, rows)
    x = np.linspace(0, 1, cols)
    yy, xx = np.meshgrid(y, x, indexing="ij")
    return 0.5 + 0.25 * np.sin((xx * 2.4 + 0.15) * np.pi) + 0.22 * np.cos((yy * 2.7 - 0.2) * np.pi)


def image_alpha(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.92, 0.88))
    ax.set_title("Image Alpha")
    ax.set_xlabel("column")
    ax.set_ylabel("row")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 6)
    ax.plot([0, 6], [0, 6], color=(0.08, 0.08, 0.10, 1), linewidth=2.0)
    ax.plot([0, 6], [6, 0], color=(0.08, 0.08, 0.10, 1), linewidth=2.0)
    ax.imshow(
        wave_image_data(6, 6),
        cmap="plasma",
        vmin=0,
        vmax=1,
        alpha=0.45,
        interpolation="bilinear",
        origin="lower",
        extent=(0, 6, 0, 6),
        aspect="auto",
    )
    save(fig, out_dir, "image_alpha")


PLOT = image_alpha


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
