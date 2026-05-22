from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def named_colors(out_dir):
    fig, ax = plt.subplots(figsize=(640 / DPI, 360 / DPI), dpi=DPI)
    ax.set_position([0.10, 0.18, 0.80, 0.70])
    ax.set_title("Named Colors")
    ax.set_xlabel("color spec")
    ax.set_ylabel("value")
    ax.set_xlim(0, 8)
    ax.set_ylim(0, 8)

    x = np.array([1, 2, 3, 4, 5, 6, 7], dtype=float)
    heights = [2.4, 3.4, 4.3, 5.2, 6.1, 4.8, 3.7]
    colors = [
        "#66c2a5",
        "0.35",
        "tab:orange",
        "rebeccapurple",
        "xkcd:cloudy blue",
        "C3",
        (0.15, 0.45, 0.65, 1),
    ]
    ax.bar(x, heights, width=0.68, color=colors, edgecolor=(0.12, 0.12, 0.12), linewidth=1)
    ax.set_xticks(x, ["hex", "gray", "tab", "css", "xkcd", "C3", "tuple"])
    ax.set_yticks([0, 2, 4, 6, 8])
    save(fig, out_dir, "named_colors")


PLOT = named_colors
