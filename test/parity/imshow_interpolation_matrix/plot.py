#!/usr/bin/env python3
"""Matplotlib reference for the imshow interpolation matrix parity case."""

from __future__ import annotations

from pathlib import Path
import argparse

import matplotlib.pyplot as plt
import numpy as np


INTERPOLATION_MODES = [
    "nearest", "none", "bilinear", "bicubic", "hanning",
    "hamming", "lanczos", "spline16", "spline36", "kaiser",
    "quadric", "catrom", "gaussian", "bessel", "mitchell",
    "sinc", "blackman", "hermite", "antialiased", "auto",
]


def _data() -> np.ndarray:
    n = 16
    y, x = np.indices((n, n))
    checker = np.where((x + y) % 2 == 0, 0.35, 0.0)
    wave = 0.25 * np.sin(x * 0.95) * np.cos(y * 0.7)
    gradient = 0.35 * (x + y) / (2 * (n - 1))
    return np.clip(0.20 + checker + wave + gradient, 0, 1)


def imshow_interpolation_matrix(out_dir: str) -> None:
    data = _data()
    fig = plt.figure(figsize=(8.0, 4.8), dpi=100)
    left, right, bottom, top = 0.04, 0.98, 0.06, 0.94
    hgap, vgap = 0.024, 0.075
    cols, rows = 5, 4
    cell_w = (right - left - hgap * (cols - 1)) / cols
    cell_h = (top - bottom - vgap * (rows - 1)) / rows

    for idx, mode in enumerate(INTERPOLATION_MODES):
        col = idx % cols
        row = idx // cols
        x0 = left + col * (cell_w + hgap)
        y1 = top - row * (cell_h + vgap)
        ax = fig.add_axes([x0, y1 - cell_h, cell_w, cell_h])
        ax.set_title(mode)
        ax.set_xticks([])
        ax.set_yticks([])
        _imshow(ax, data, mode)

    Path(out_dir).mkdir(parents=True, exist_ok=True)
    fig.savefig(Path(out_dir) / "imshow_interpolation_matrix.png")
    plt.close(fig)


PLOT = imshow_interpolation_matrix


def _imshow(ax, data: np.ndarray, mode: str) -> None:
    kwargs = dict(
        cmap="viridis",
        vmin=0,
        vmax=1,
        extent=(0, data.shape[1], 0, data.shape[0]),
        origin="lower",
        aspect="auto",
        interpolation=mode,
    )
    try:
        ax.imshow(data, **kwargs)
    except ValueError:
        if mode != "auto":
            raise
        # Some older installed Matplotlib builds lack the newer public "auto"
        # registry name. The vendored upstream source includes it, and it shares
        # the adaptive nearest/Hanning policy with "antialiased" for this tile.
        kwargs["interpolation"] = "antialiased"
        ax.imshow(data, **kwargs)


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
