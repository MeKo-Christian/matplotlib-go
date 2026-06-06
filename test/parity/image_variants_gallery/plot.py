#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import argparse
import math
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def clamp_unit(v):
    return max(0, min(1, v))


def image_data(n):
    out = []
    for y in range(n):
        row = []
        for x in range(n):
            checker = 0.35 if (x + y) % 2 == 0 else 0.0
            wave = 0.25 * math.sin(x * 0.95) * math.cos(y * 0.7)
            gradient = 0.35 * (x + y) / (2 * (n - 1))
            row.append(clamp_unit(0.20 + checker + wave + gradient))
        out.append(row)
    return out


def checker_data(n):
    return [[0.25 if (x // 3 + y // 3) % 2 == 0 else 0.85 for x in range(n)] for y in range(n)]


def radial_data(n):
    out = []
    for y in range(n):
        yy = 2 * y / (n - 1) - 1
        row = []
        for x in range(n):
            xx = 2 * x / (n - 1) - 1
            row.append(math.exp(-3.5 * (xx * xx + yy * yy)) + 0.2 * math.sin(10 * xx))
        out.append(row)
    return out


def matshow_data():
    return [
        [0.2, 0.3, 0.5, 0.8, 0.6],
        [0.1, 0.6, 0.7, 0.4, 0.3],
        [0.9, 0.8, 0.2, 0.3, 0.5],
        [0.4, 0.2, 0.6, 0.9, 0.7],
    ]


def sparse_data(n):
    out = []
    for y in range(n):
        row = []
        for x in range(n):
            row.append(1 if x == y or x + y == n - 1 or (x + 2 * y) % 7 == 0 or (2 * x + y) % 11 == 0 else 0)
        out.append(row)
    return out


def image_variants_gallery(out_dir):
    fig = make_fig_px(1080, 720)
    for i, mode in enumerate(["nearest", "bilinear", "bicubic"]):
        left, gap, w = 0.06, 0.035, 0.27
        ax = fig.add_axes(go_rect(left + i * (w + gap), 0.58, left + i * (w + gap) + w, 0.88))
        ax.set_title(mode)
        ax.set_xticks([])
        ax.set_yticks([])
        ax.imshow(image_data(16), cmap="viridis", vmin=0, vmax=1, origin="lower", extent=(0, 16, 0, 16), aspect="auto", interpolation=mode)

    alpha_ax = fig.add_axes(go_rect(0.06, 0.12, 0.31, 0.43))
    alpha_ax.set_title("alpha overlay")
    alpha_ax.set_xlim(0, 18)
    alpha_ax.set_ylim(0, 18)
    alpha_ax.imshow(checker_data(18), cmap="Greys", origin="lower", extent=(0, 18, 0, 18), aspect="auto")
    alpha_ax.imshow(radial_data(18), cmap="magma", alpha=0.58, origin="lower", extent=(0, 18, 0, 18), aspect="auto")

    mat_ax = fig.add_axes(go_rect(0.39, 0.12, 0.61, 0.43))
    mat_ax.set_title("matshow")
    mat_ax.matshow(matshow_data(), cmap="plasma")

    marker_ax = fig.add_axes(go_rect(0.68, 0.12, 0.80, 0.43))
    marker_ax.set_title("spy markers")
    marker_ax.spy(sparse_data(18), precision=0.1, marker="s", markersize=7, color=(0.12, 0.38, 0.70, 1))

    image_ax = fig.add_axes(go_rect(0.84, 0.12, 0.96, 0.43))
    image_ax.set_title("spy image")
    image_ax.spy(sparse_data(18), precision=0.1)

    fig.text(0.05, 0.96, "Image Variants Gallery", fontsize=13)
    save(fig, out_dir, "image_variants_gallery")


PLOT = image_variants_gallery


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
