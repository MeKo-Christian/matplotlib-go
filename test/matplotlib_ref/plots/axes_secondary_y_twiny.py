#!/usr/bin/env python3
"""Focused Matplotlib parity fixture for twiny and secondary_yaxis."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import matplotlib.ticker as mticker

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2]))
    from matplotlib_ref.common import *  # noqa: F401,F403


def axes_secondary_y_twiny(out_dir):
    fig = make_fig_px(760, 400)
    blue = (0.12, 0.47, 0.71)
    orange = (1.0, 0.50, 0.05)

    host = fig.add_axes(go_rect(0.09, 0.17, 0.45, 0.82))
    host.set_title("TwinY")
    host.set_xlabel("bottom x")
    host.set_ylabel("shared y")
    host.plot([0, 1, 2, 3, 4], [0, 1, 2, 3, 4], color=blue, linewidth=2.0)
    host.set_xlim(0, 4)
    host.set_ylim(0, 4)
    host.xaxis.set_major_locator(mticker.FixedLocator([0, 2, 4]))
    host.yaxis.set_major_locator(mticker.FixedLocator([0, 2, 4]))
    host.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=0.5)
    host.set_axisbelow(True)

    twin = host.twiny()
    twin.plot([10, 20, 30, 40, 50], [4, 3, 2, 1, 0], color=orange, linewidth=2.0)
    twin.set_xlim(10, 50)
    twin.xaxis.set_major_locator(mticker.FixedLocator([10, 30, 50]))
    twin.set_xlabel("top x")

    primary = fig.add_axes(go_rect(0.57, 0.17, 0.88, 0.82))
    primary.set_title("SecondaryYAxis")
    primary.set_xlabel("sample")
    primary.set_ylabel("Celsius")
    primary.plot([0, 1, 2, 3, 4], [0, 20, 40, 60, 100], color=blue, linewidth=2.0)
    primary.set_xlim(0, 4)
    primary.set_ylim(0, 100)
    primary.xaxis.set_major_locator(mticker.FixedLocator([0, 2, 4]))
    primary.yaxis.set_major_locator(mticker.FixedLocator([0, 50, 100]))
    primary.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=0.5)
    primary.set_axisbelow(True)

    secondary = primary.secondary_yaxis(
        "right",
        functions=(lambda celsius: celsius * 1.8 + 32, lambda fahrenheit: (fahrenheit - 32) / 1.8),
    )
    secondary.yaxis.set_major_locator(mticker.FixedLocator([32, 122, 212]))
    secondary.set_ylabel("Fahrenheit")

    save(fig, out_dir, "axes_secondary_y_twiny")


PLOT = axes_secondary_y_twiny


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
