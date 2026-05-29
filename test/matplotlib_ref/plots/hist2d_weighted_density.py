#!/usr/bin/env python3
"""Matplotlib reference plot for weighted density hist2d coverage."""

from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

try:
    from test.matplotlib_ref.plots.mesh_fixture_common import mesh_fixture_axes, hist2d_weighted_data
except ModuleNotFoundError:
    from mesh_fixture_common import mesh_fixture_axes, hist2d_weighted_data


def hist2d_weighted_density(out_dir):
    fig, ax = mesh_fixture_axes("hist2d weighted density")
    x, y, weights = hist2d_weighted_data()
    _, _, _, mesh = ax.hist2d(
        x,
        y,
        bins=[[-2, -1, 0, 1, 2, 3], [-1.5, -0.5, 0.5, 1.5, 2.5]],
        weights=weights,
        density=True,
        cmap="magma",
    )
    fig.colorbar(mesh, ax=ax, label="density")
    ax.set_xlim(-2, 3)
    ax.set_ylim(-1.5, 2.5)
    save(fig, out_dir, "hist2d_weighted_density")


PLOT = hist2d_weighted_density
