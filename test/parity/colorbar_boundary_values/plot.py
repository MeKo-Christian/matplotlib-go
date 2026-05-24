from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def colorbar_boundary_values(out_dir):
    fig, ax = plt.subplots(figsize=(640 / DPI, 360 / DPI), dpi=DPI)
    ax.set_position([0.12, 0.16, 0.78, 0.72])
    ax.set_title("Boundary Colorbar Values")
    ax.set_xlabel("x")
    ax.set_ylabel("y")

    data = np.array([
        [-0.7, -0.2, 0.2, 0.8],
        [-0.5, 0.0, 0.6, 1.2],
        [-0.1, 0.3, 0.9, 1.4],
    ])
    im = ax.imshow(
        data,
        cmap="viridis",
        norm=mcolors.Normalize(vmin=-0.5, vmax=1.2),
        origin="lower",
        extent=[0, 4, 0, 3],
        aspect="auto",
    )
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 3)
    cbar = fig.colorbar(
        im,
        ax=ax,
        extend="both",
        extendrect=True,
        boundaries=[-0.5, -0.1, 0.4, 1.2],
        values=[-0.35, 0.15, 0.8],
        spacing="uniform",
        drawedges=True,
        ticks=[-0.5, -0.1, 0.4, 1.2],
    )
    cbar.set_label("bands")

    save(fig, out_dir, "colorbar_boundary_values")


PLOT = colorbar_boundary_values
