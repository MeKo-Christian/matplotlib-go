#!/usr/bin/env python3
"""Matplotlib reference plot for pcolormesh nearest coverage."""

from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

try:
    from test.matplotlib_ref.plots.mesh_fixture_common import mesh_fixture_data, mesh_fixture_axes
except ModuleNotFoundError:
    from mesh_fixture_common import mesh_fixture_data, mesh_fixture_axes


def pcolormesh_nearest(out_dir):
    fig, ax = mesh_fixture_axes("nearest centers")
    ax.pcolormesh(
        [0.0, 0.8, 2.1, 3.5, 5.0],
        [-1.0, 0.2, 1.4, 2.7],
        mesh_fixture_data(4, 5),
        shading="nearest",
        cmap="viridis",
        vmin=-0.75,
        vmax=0.95,
    )
    ax.set_xlim(-0.4, 5.7)
    ax.set_ylim(-1.6, 3.35)
    save(fig, out_dir, "pcolormesh_nearest")


PLOT = pcolormesh_nearest
