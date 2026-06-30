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

def polar_axes(out_dir):
    fig = make_fig_px(720, 720)
    ax = fig.add_axes(go_rect(0.12, 0.10, 0.88, 0.88), projection="polar")
    ax.set_title("Polar Axes")
    ax.set_xlabel("theta")
    ax.set_ylabel("radius")
    ax.set_ylim(0, 1.1)

    thetas = np.linspace(0.0, 2.0 * math.pi, 720)
    radii = 0.55 + 0.35 * np.cos(5.0 * thetas)

    ax.xaxis.grid(True, color=(0.8, 0.82, 0.86, 1.0), linewidth=0.9)
    ax.yaxis.grid(True, color=(0.82, 0.84, 0.88, 0.9), linewidth=0.8)
    ax.plot(
        thetas,
        radii,
        color=(0.16, 0.33, 0.73, 1.0),
        linewidth=2.2,
        label="r = 0.55 + 0.35 cos(5theta)",
    )
    ax.fill_between(
        thetas,
        radii,
        0,
        color=(0.36, 0.56, 0.92, 0.2),
    )
    save(fig, out_dir, "polar_axes")

PLOT = polar_axes


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
