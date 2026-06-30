#!/usr/bin/env python3
"""Matplotlib reference for the widgets and selectors gallery."""

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

from matplotlib.patches import Ellipse, Polygon, Rectangle
from matplotlib.widgets import Button, CheckButtons, RadioButtons, RangeSlider, Slider, TextBox


def _small_axes(fig, left, bottom, width, height):
    ax = fig.add_axes([left, bottom, width, height])
    return ax


def widgets_gallery(out_dir):
    width, height = 1100, 760
    fig = make_fig_px(width, height)

    main_ax = fig.add_axes([0.06, 0.36, 0.64, 0.56])
    main_ax.set_title("Widgets and selectors")
    main_ax.set_xlabel("phase")
    main_ax.set_ylabel("amplitude")
    main_ax.set_xlim(0, 2 * math.pi)
    main_ax.set_ylim(-1.35, 1.35)
    main_ax.grid(axis="y", color=(0.85, 0.85, 0.85), linewidth=0.8)

    x = np.linspace(0, 2 * math.pi, 220)
    signal = np.sin(x)
    modulation = 0.55 * np.cos(1.7 * x)
    blue = (0.12, 0.34, 0.68)
    orange = (0.84, 0.35, 0.18)
    main_ax.plot(x, signal, color=blue, linewidth=2.2, label="signal")
    main_ax.plot(x, modulation, color=orange, linewidth=1.7, label="modulation")
    main_ax.legend(loc="upper right", frameon=True)

    main_ax.axvspan(0.7, 1.35, color=blue, alpha=0.18)
    main_ax.add_patch(Rectangle((2.25, -0.95), 0.8, 0.7, facecolor=blue, edgecolor=blue, alpha=0.20))
    main_ax.add_patch(Ellipse((4.0, 0.625), 0.9, 0.75, facecolor=orange, edgecolor=orange, alpha=0.22))
    main_ax.add_patch(
        Polygon(
            [(4.95, -0.9), (5.75, -0.75), (5.45, -0.15)],
            closed=True,
            facecolor=(0.17, 0.63, 0.17, 0.22),
            edgecolor=(0.17, 0.63, 0.17),
        )
    )
    lasso = np.array([(1.2, 0.75), (1.45, 1.0), (1.85, 0.92), (2.05, 0.63), (1.75, 0.42), (1.34, 0.5)])
    main_ax.plot(lasso[:, 0], lasso[:, 1], color=(0.58, 0.40, 0.74), linewidth=1.4)
    main_ax.axvline(2.8, color=(0.2, 0.2, 0.2), linewidth=1.0, alpha=0.75)
    main_ax.axhline(0.35, color=(0.2, 0.2, 0.2), linewidth=1.0, alpha=0.75)

    aux_ax = fig.add_axes([0.76, 0.36, 0.18, 0.56], sharex=main_ax)
    aux_ax.set_title("shared cursor")
    aux_ax.set_xlim(0, 2 * math.pi)
    aux_ax.set_ylim(-1, 1)
    aux_ax.grid(axis="y", color=(0.85, 0.85, 0.85), linewidth=0.8)
    aux_ax.plot(x, modulation, color=orange, linewidth=1.7)
    aux_ax.axvline(4.45, color=(0.2, 0.2, 0.2), linewidth=1.0, alpha=0.75)
    aux_ax.axhline(0.0, color=(0.2, 0.2, 0.2), linewidth=1.0, alpha=0.75)

    Button(_small_axes(fig, 0.06, 0.23, 0.16, 0.07), "Apply")
    Slider(_small_axes(fig, 0.28, 0.23, 0.26, 0.07), "gain", 0, 1, valinit=0.68)
    RangeSlider(_small_axes(fig, 0.59, 0.23, 0.23, 0.07), "window", 0, 1, valinit=(0.22, 0.78))
    TextBox(_small_axes(fig, 0.86, 0.23, 0.10, 0.07), "label", initial="phase scan")
    CheckButtons(_small_axes(fig, 0.06, 0.07, 0.36, 0.12), ["signal", "modulation", "grid"], [True, True, False])
    RadioButtons(_small_axes(fig, 0.55, 0.07, 0.28, 0.12), ["blue", "amber", "mono"], active=1)

    save(fig, out_dir, "widgets_gallery")


PLOT = widgets_gallery


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
