from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2] / "matplotlib_ref"))
    from common import *  # noqa: F401,F403


def colorbar_horizontal_ticks(out_dir):
    fig = make_fig_px(640, 360)
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.90, 0.88))
    ax.set_title("Horizontal Colorbar Ticks")
    ax.set_xlabel("x")
    ax.set_ylabel("y")

    mesh = ax.pcolormesh(
        [0, 1, 2, 3, 4],
        [0, 1, 2, 3],
        np.array([
            [-1.0, -0.5, 0.0, 0.5],
            [-0.6, -0.1, 0.4, 0.9],
            [-0.2, 0.3, 0.8, 1.2],
        ]),
        shading="flat",
        cmap="viridis",
        vmin=-1,
        vmax=1.2,
    )
    cbar = fig.colorbar(mesh, ax=ax, location="bottom", ticks=[-1, 0, 1])
    cbar.set_label("horizontal")
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 3)
    save(fig, out_dir, "colorbar_horizontal_ticks")


PLOT = colorbar_horizontal_ticks
