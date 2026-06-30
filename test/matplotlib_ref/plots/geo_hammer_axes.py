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


def geo_hammer_axes(out_dir):
    fig = make_fig_px(720, 420)
    ax = fig.add_axes(go_rect(0.10, 0.14, 0.92, 0.86), projection="hammer")
    ax.set_title("Hammer Projection")
    ax.set_xlabel("longitude")
    ax.set_ylabel("latitude")
    ax.set_xticks(np.arange(-150, 151, 30) * math.pi / 180.0)
    ax.set_yticks(np.arange(-75, 76, 15) * math.pi / 180.0)
    degree_formatter = matplotlib.ticker.FuncFormatter(lambda x, _: f"{round(x * 180.0 / math.pi):.0f}")
    ax.xaxis.set_major_formatter(degree_formatter)
    ax.yaxis.set_major_formatter(degree_formatter)

    ax.xaxis.grid(True, color=(0.78, 0.80, 0.84, 1.0), linewidth=0.8)
    ax.yaxis.grid(True, color=(0.78, 0.80, 0.84, 1.0), linewidth=0.8)

    lon = np.linspace(-math.pi, math.pi, 361)
    lat = 0.35 * np.sin(3.0 * lon)
    ax.plot(lon, lat, color=(0.14, 0.34, 0.70, 1.0), linewidth=2.0)

    save(fig, out_dir, "geo_hammer_axes")


PLOT = geo_hammer_axes


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
