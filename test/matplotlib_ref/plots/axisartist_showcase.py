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

def axisartist_showcase(out_dir):
    fig = make_fig_px(980, 640)

    host = fig.add_axes(go_rect(0.08, 0.14, 0.56, 0.88))
    host.set_title("AxisArtist / Parasite")
    host.set_xlabel("phase")
    host.set_ylabel("signal")
    host.set_xlim(-3.5, 3.5)
    host.set_ylim(-1.3, 1.3)
    host.grid(axis="y", color=(0.78, 0.80, 0.84, 1.0), linewidth=0.8)

    x = np.linspace(-3.5, 3.5, 240)
    sine = np.sin(x)
    cos_scaled = 55 + 35 * np.cos(x * 0.8)

    host.plot(x, sine, color=(0.14, 0.34, 0.72, 1.0), linewidth=2.2, label="sin(x)")
    host.axhline(0.0, color=(0.26, 0.26, 0.30, 1.0), linewidth=1.4, dashes=[5 * 36.0 / DPI, 3 * 36.0 / DPI])
    host.axvline(0.0, color=(0.26, 0.26, 0.30, 1.0), linewidth=1.4, dashes=[5 * 36.0 / DPI, 3 * 36.0 / DPI])
    host.tick_params(direction="inout")

    right = host.twinx()
    right.set_ylim(0, 100)
    right.plot(x, cos_scaled, color=(0.74, 0.28, 0.18, 1.0), linewidth=1.8, label="55 + 35 cos(0.8x)")
    right.spines["right"].set_color((0.74, 0.28, 0.18, 1.0))
    right.tick_params(axis="y", colors=(0.74, 0.28, 0.18, 1.0))
    right.spines["top"].set_visible(False)
    right.spines["left"].set_visible(False)
    right.spines["bottom"].set_visible(False)

    host.text(
        0.02,
        0.98,
        "floating axes at x=0 / y=0\nparasite right scale",
        transform=host.transAxes,
        ha="left",
        va="top",
        fontsize=10,
        bbox=dict(boxstyle="round,pad=0.3", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
    )
    host.legend(loc="upper center")

    save(fig, out_dir, "axisartist_showcase")

PLOT = axisartist_showcase


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
