from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2] / "matplotlib_ref"))
    from common import *  # noqa: F401,F403


def collection_mutable_scalarmap(out_dir):
    fig = make_fig_px(640, 360)
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.90, 0.88))
    ax.set_title("Mutable ScalarMap Collection")
    ax.set_xlabel("x")
    ax.set_ylabel("y")

    initial = np.array([
        [-0.8, -0.4, 0.0, 0.4],
        [-0.6, -0.1, 0.3, 0.8],
        [-0.2, 0.2, 0.6, 1.0],
    ])
    mesh = ax.pcolormesh(
        [0, 1, 2, 3, 4],
        [0, 1, 2, 3],
        initial,
        shading="flat",
        cmap="viridis",
    )
    cbar = fig.colorbar(mesh, ax=ax)
    cbar.set_label("updated")

    mesh.set_array(np.array([
        [1.00, 0.70, 0.35, 0.05],
        [0.80, 0.45, 0.10, -0.20],
        [0.55, 0.20, -0.15, -0.50],
    ]).ravel())
    mesh.set_cmap("plasma")
    mesh.set_clim(-0.5, 1.0)
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 3)
    save(fig, out_dir, "collection_mutable_scalarmap")


PLOT = collection_mutable_scalarmap
