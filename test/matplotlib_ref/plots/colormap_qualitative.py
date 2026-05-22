from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def _qualitative_data():
    return np.array(
        [
            [0, 1, 2, 3, 4, 5, 6, 7, 8, 9],
            [10, 11, 12, 13, 14, 15, 16, 17, 18, 19],
        ],
        dtype=float,
    )


def colormap_qualitative(out_dir):
    fig, ax = plt.subplots(figsize=(640 / DPI, 360 / DPI), dpi=DPI)
    ax.set_position([0.12, 0.16, 0.78, 0.72])
    ax.set_title("Qualitative Colormap")
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.imshow(
        _qualitative_data(),
        cmap="tab20",
        vmin=0,
        vmax=19,
        origin="lower",
        extent=[0, 10, 0, 2],
        aspect="auto",
        interpolation="nearest",
    )
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 2)
    save(fig, out_dir, "colormap_qualitative")


PLOT = colormap_qualitative
