#!/usr/bin/env python3
"""Matplotlib reference plot for masked pcolormesh coverage."""

from __future__ import annotations

from pathlib import Path
import sys

import matplotlib
import numpy as np

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

try:
    from test.matplotlib_ref.plots.mesh_fixture_common import mesh_fixture_data, mesh_fixture_axes
except ModuleNotFoundError:
    from mesh_fixture_common import mesh_fixture_data, mesh_fixture_axes


def pcolormesh_masked(out_dir):
    fig, ax = mesh_fixture_axes("masked mesh")
    cmap = matplotlib.colormaps["viridis"].copy()
    cmap.set_bad((0.62, 0.62, 0.62, 0.78))
    data = np.ma.array(
        mesh_fixture_data(4, 5),
        mask=[
            [False, True, False, False, False],
            [False, False, False, True, False],
            [True, False, False, False, False],
            [False, False, True, False, False],
        ],
    )
    ax.pcolormesh(
        [0, 1, 2, 3, 4, 5],
        [0, 1, 2, 3, 4],
        data,
        cmap=cmap,
        edgecolors=[(0.98, 0.98, 0.98, 1)],
        linewidth=0.7,
    )
    ax.set_xlim(0, 5)
    ax.set_ylim(0, 4)
    save(fig, out_dir, "pcolormesh_masked")


PLOT = pcolormesh_masked
