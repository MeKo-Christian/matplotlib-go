from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def colorbar_extendfrac(out_dir):
    fig, ax = plt.subplots(figsize=(640 / DPI, 360 / DPI), dpi=DPI)
    ax.set_position([0.12, 0.16, 0.78, 0.72])
    ax.set_title("Colorbar ExtendFrac")
    ax.set_xlabel("x")
    ax.set_ylabel("y")

    cmap = matplotlib.colormaps["viridis"].copy()
    cmap.set_under((0.08, 0.16, 0.72, 1.0))
    cmap.set_over((0.78, 0.12, 0.08, 1.0))
    data = np.array([
        [-0.35, 0.15, 0.35],
        [0.55, 0.85, 1.35],
    ])
    mesh = ax.pcolormesh([0, 1, 2, 3], [0, 1, 2], data, cmap=cmap, vmin=0, vmax=1, shading="flat")
    ax.set_xlim(0, 3)
    ax.set_ylim(0, 2)
    cbar = fig.colorbar(mesh, ax=ax, extend="both", extendfrac=[0.1, 0.3])
    cbar.set_label("extendfrac")

    save(fig, out_dir, "colorbar_extendfrac")


PLOT = colorbar_extendfrac
