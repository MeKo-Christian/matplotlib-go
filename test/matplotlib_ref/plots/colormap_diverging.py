from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def _diverging_data(rows, cols):
    data = np.zeros((rows, cols))
    for row in range(rows):
        y = row / max(1, rows - 1)
        for col in range(cols):
            x = col / max(1, cols - 1)
            data[row, col] = 2 * x - 1 + 0.22 * (y - 0.5)
    return data


def colormap_diverging(out_dir):
    fig, ax = plt.subplots(figsize=(640 / DPI, 360 / DPI), dpi=DPI)
    ax.set_position([0.12, 0.16, 0.78, 0.72])
    ax.set_title("Diverging Colormap")
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.imshow(
        _diverging_data(5, 9),
        cmap="RdBu",
        vmin=-1,
        vmax=1,
        origin="lower",
        extent=[0, 9, 0, 5],
        aspect="auto",
        interpolation="nearest",
    )
    ax.set_xlim(0, 9)
    ax.set_ylim(0, 5)
    save(fig, out_dir, "colormap_diverging")


PLOT = colormap_diverging
