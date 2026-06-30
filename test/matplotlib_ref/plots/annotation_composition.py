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

def annotation_composition(out_dir):
    fig = make_fig_px(1040, 720)
    ax = fig.add_axes(go_rect(0.10, 0.14, 0.90, 0.88))
    ax.set_title("Text and Arrow Annotations")
    ax.set_xlabel("phase")
    ax.set_ylabel("response")
    ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
    ax.set_axisbelow(True)

    x = np.linspace(0, 6 * math.pi, 240)
    y = np.sin(x) * np.exp(-0.015 * x) + 0.2 * np.cos(0.5 * x)
    ax.plot(x, y, label="signal")
    ax.set_xlim(0, 6 * math.pi)
    ax.set_ylim(-1.2, 1.2)
    ax.legend(loc="upper right")

    peak_x = math.pi / 2
    peak_y = math.sin(peak_x) * math.exp(-0.015 * peak_x) + 0.2 * math.cos(0.5 * peak_x)
    ax.annotate(
        "Peak\n= 0.42",
        xy=(peak_x, peak_y),
        xytext=(48, -42),
        textcoords="offset pixels",
        fontsize=12,
        arrowprops=dict(arrowstyle="->", color="black", lw=1.0),
    )
    ax.text(0.20, 0.90, "m∫T  φ x =  λ/4", transform=ax.transAxes, fontsize=12)

    save(fig, out_dir, "annotation_composition")

PLOT = annotation_composition


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
