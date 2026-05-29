#!/usr/bin/env python3
"""Matplotlib reference plot for pcolormesh Gouraud coverage."""

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


def pcolormesh_gouraud(out_dir):
    fig, ax = mesh_fixture_axes("Gouraud mesh")
    mesh = ax.pcolormesh(
        [-2.5, -1.5, -0.4, 0.8, 1.7, 2.8],
        [-1.8, -0.8, 0.1, 1.1, 2.0],
        mesh_fixture_data(5, 6),
        shading="gouraud",
        cmap="viridis",
        vmin=-0.85,
        vmax=0.85,
    )
    fig.colorbar(mesh, ax=ax, label="value")
    ax.set_xlim(-2.5, 2.8)
    ax.set_ylim(-1.8, 2.0)
    save(fig, out_dir, "pcolormesh_gouraud")


PLOT = pcolormesh_gouraud
