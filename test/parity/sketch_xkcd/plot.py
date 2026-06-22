#!/usr/bin/env python3
"""Matplotlib reference plot for the sketch/xkcd path perturbation.

Sets rcParams['path.sketch'] = (1, 100, 2) directly — the same parameters as
pyplot.xkcd() — but keeps the default font, so only the path wiggle is compared
(the xkcd handwriting font is not bundled with matplotlib-go)."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import numpy as np
import matplotlib as mpl

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def sketch_xkcd(out_dir):
    with mpl.rc_context({"path.sketch": (1, 100, 2)}):
        fig = make_fig()
        ax = fig.add_axes(go_rect(0.12, 0.14, 0.94, 0.88))
        ax.set_xlim(0, 10)
        ax.set_ylim(-1.2, 1.2)

        x = np.linspace(0, 10, 200)
        ax.plot(x, np.sin(x), color=(0.12, 0.47, 0.71, 1), linewidth=lw(2),
                solid_capstyle="butt")
        ax.plot([0, 10], [0, 0], color=(0.84, 0.15, 0.16, 1), linewidth=lw(2),
                solid_capstyle="butt")

        save(fig, out_dir, "sketch_xkcd")


PLOT = sketch_xkcd


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
