from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def _cyclic_data(rows, cols):
    data = np.zeros((rows, cols))
    cx = (cols - 1) / 2
    cy = (rows - 1) / 2
    for row in range(rows):
        for col in range(cols):
            angle = math.atan2(row - cy, col - cx)
            data[row, col] = (angle + math.pi) / (2 * math.pi)
    return data


def colormap_cyclic(out_dir):
    fig, ax = plt.subplots(figsize=(640 / DPI, 360 / DPI), dpi=DPI)
    ax.set_position([0.12, 0.16, 0.78, 0.72])
    ax.set_title("Cyclic Colormap")
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.imshow(
        _cyclic_data(6, 12),
        cmap="twilight",
        vmin=0,
        vmax=1,
        origin="lower",
        extent=[0, 12, 0, 6],
        aspect="auto",
        interpolation="nearest",
    )
    ax.set_xlim(0, 12)
    ax.set_ylim(0, 6)
    save(fig, out_dir, "colormap_cyclic")


PLOT = colormap_cyclic
