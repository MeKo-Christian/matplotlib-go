#!/usr/bin/env python3
"""Matplotlib reference plot for errorbar capthick (thick caps)."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

# Shared with test/parity/errorbar_capthick/plot.go.
X = [1, 2, 3, 4, 5, 6]
Y = [1.8, 2.5, 2.2, 3.1, 2.8, 3.7]
XERR = [0.20, 0.25, 0.15, 0.22, 0.30, 0.18]
YERR = [0.28, 0.20, 0.35, 0.24, 0.30, 0.22]


def errorbar_capthick(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Error Bars (capthick)")
    ax.set_xlim(0, 7)
    ax.set_ylim(0, 6)

    ax.scatter(X, Y, s=ss(4.5), c=[(0.17, 0.63, 0.17)], linewidths=0)
    ax.errorbar(
        X,
        Y,
        xerr=XERR,
        yerr=YERR,
        fmt="none",
        ecolor=(0, 0, 0, 1),
        elinewidth=1.2,
        capsize=6,
        capthick=3.0,
    )
    save(fig, out_dir, "errorbar_capthick")


PLOT = errorbar_capthick


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
