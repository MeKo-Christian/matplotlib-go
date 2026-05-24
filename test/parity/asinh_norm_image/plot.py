from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def _asinh_norm_fixture_data(rows, cols):
    data = np.zeros((rows, cols))
    for row in range(rows):
        y = row / max(1, rows - 1)
        for col in range(cols):
            x = col / max(1, cols - 1)
            data[row, col] = 160 * (x - 0.5) + 30 * math.sin((y - 0.5) * math.pi)
    return data


def asinh_norm_image(out_dir):
    fig, ax = plt.subplots(figsize=(640 / DPI, 360 / DPI), dpi=DPI)
    ax.set_position([0.12, 0.16, 0.78, 0.72])
    ax.set_title("AsinhNorm Image")
    ax.set_xlabel("x")
    ax.set_ylabel("y")

    im = ax.imshow(
        _asinh_norm_fixture_data(5, 7),
        cmap="viridis",
        norm=mcolors.AsinhNorm(linear_width=2, vmin=-80, vmax=120),
        origin="lower",
        extent=[0, 7, 0, 5],
        aspect="auto",
    )
    ax.set_xlim(0, 7)
    ax.set_ylim(0, 5)
    cbar = fig.colorbar(im, ax=ax)
    cbar.set_label("asinh value")

    save(fig, out_dir, "asinh_norm_image")


PLOT = asinh_norm_image
