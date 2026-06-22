#!/usr/bin/env python3
"""Matplotlib reference for native RGB/RGBA imshow parity."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def rgb_gradient(ny, nx):
    img = np.zeros((ny, nx, 3))
    for row in range(ny):
        for col in range(nx):
            img[row, col] = (col / (nx - 1), row / (ny - 1), 0.30)
    return img


def rgba_blocks(ny, nx):
    img = np.zeros((ny, nx, 4))
    for row in range(ny):
        for col in range(nx):
            img[row, col] = (0.85, 0.10, 0.10, col / (nx - 1))
    return img


def imshow_rgb(out_dir):
    fig = make_fig()

    ax_rgb = fig.add_axes(go_rect(0.07, 0.12, 0.47, 0.90))
    ax_rgb.set_title("RGB")
    ax_rgb.imshow(rgb_gradient(8, 8), interpolation="nearest", aspect="auto")

    ax_rgba = fig.add_axes(go_rect(0.55, 0.12, 0.95, 0.90))
    ax_rgba.set_title("RGBA")
    ax_rgba.imshow(rgba_blocks(8, 8), interpolation="nearest", aspect="auto")

    save(fig, out_dir, "imshow_rgb")


PLOT = imshow_rgb


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
