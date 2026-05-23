#!/usr/bin/env python3
"""Matplotlib reference for bilinear imshow parity."""

from __future__ import annotations

from pathlib import Path
import argparse

import matplotlib.pyplot as plt
import numpy as np


def _plot(output_dir: str, name: str, interpolation: str) -> None:
    n = 32
    y, x = np.indices((n, n))
    data = ((x + y) % 2 == 0).astype(float)

    fig = plt.figure(figsize=(2.56, 2.56), dpi=100)
    ax = fig.add_axes([0, 0, 1, 1])
    ax.imshow(
        data,
        cmap="gray",
        vmin=0,
        vmax=1,
        extent=(0, n, 0, n),
        origin="lower",
        interpolation=interpolation,
    )
    Path(output_dir).mkdir(parents=True, exist_ok=True)
    fig.savefig(Path(output_dir) / f"{name}.png")
    plt.close(fig)


def imshow_bilinear(out_dir: str) -> None:
    _plot(out_dir, "imshow_bilinear", "bilinear")


PLOT = imshow_bilinear


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
