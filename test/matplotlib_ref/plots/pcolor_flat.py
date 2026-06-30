#!/usr/bin/env python3
"""Matplotlib reference plot for pcolor flat coverage."""

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


def pcolor_flat(out_dir):
    fig, ax = mesh_fixture_axes("pcolor flat")
    ax.pcolor(
        [0.0, 0.9, 2.0, 3.4, 4.1, 5.2],
        [-0.2, 0.8, 1.6, 2.9, 4.0],
        mesh_fixture_data(4, 5),
        shading="flat",
        cmap="plasma",
        edgecolors=[(0.96, 0.96, 0.96, 1)],
        linewidth=0.75,
    )
    ax.set_xlim(0, 5.2)
    ax.set_ylim(-0.2, 4.0)
    save(fig, out_dir, "pcolor_flat")


PLOT = pcolor_flat
