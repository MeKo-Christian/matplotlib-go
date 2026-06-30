from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2] / "matplotlib_ref"))
    from common import *  # noqa: F401,F403


def axes_option_breadth(out_dir):
    fig = make_fig_px(840, 540)

    scatter_ax = fig.add_axes(go_rect(0.07, 0.58, 0.46, 0.92))
    scatter_ax.set_title("Scatter options")
    scatter_ax.set_xlim(0, 6)
    scatter_ax.set_ylim(0, 5)
    scatter_ax.scatter(
        [0.8, 1.8, 2.8, 3.8, 4.8],
        [1.2, 3.2, 2.2, 3.8, 1.8],
        c=[-1.0, -0.2, 0.35, 0.8, 1.2],
        cmap="viridis",
        s=[ss(7), ss(10), ss(13), ss(16), ss(19)],
        edgecolors=[(0.08, 0.08, 0.08, 1)],
        linewidths=1.5,
    )

    bar_ax = fig.add_axes(go_rect(0.56, 0.58, 0.96, 0.92))
    bar_ax.set_title("Bar options")
    bar_ax.set_xlim(0, 4.2)
    bar_ax.set_ylim(0, 6.2)
    x = [0.6, 1.7, 2.9]
    widths = [0.42, 0.58, 0.36]
    base = bar_ax.bar(
        x,
        [2.0, 3.1, 1.6],
        width=widths,
        align="edge",
        color=[(0.12, 0.47, 0.71, 1), (1.00, 0.50, 0.05, 1), (0.17, 0.63, 0.17, 1)],
        edgecolor=(0.10, 0.10, 0.10, 1),
        linewidth=1.0,
    )
    bar_ax.bar_label(base, fmt="%.1f", padding=2, fontsize=8)
    top = bar_ax.bar(
        x,
        [1.3, 1.0, 1.8],
        width=widths,
        bottom=[2.0, 3.1, 1.6],
        align="edge",
        color=(0.58, 0.40, 0.74, 1),
        edgecolor=(0.10, 0.10, 0.10, 1),
        linewidth=1.0,
    )
    bar_ax.bar_label(top, fmt="%.1f", padding=2, fontsize=8)

    fill_ax = fig.add_axes(go_rect(0.07, 0.10, 0.46, 0.44))
    fill_ax.set_title("Fill options")
    fill_ax.set_xlim(0, 5)
    fill_ax.set_ylim(-1.4, 1.4)
    fill_ax.fill_between(
        [0, 1, 2, 3, 4, 5],
        [-0.6, 0.8, 0.2, 1.0, -0.2, 0.7],
        [0.4, -0.4, 0.6, -0.3, 0.5, -0.5],
        where=[False, True, True, False, True, True],
        interpolate=True,
        step="post",
        facecolor=(0.84, 0.15, 0.16, 0.55),
        edgecolor=(0.50, 0.05, 0.06, 0.55),
        linewidth=1.0,
    )

    error_ax = fig.add_axes(go_rect(0.56, 0.10, 0.96, 0.44))
    error_ax.set_title("Errorbar options")
    error_ax.set_xlim(0, 6)
    error_ax.set_ylim(0, 5)
    error_ax.errorbar(
        [0.8, 1.6, 2.4, 3.2, 4.0, 4.8],
        [1.0, 2.0, 1.5, 3.3, 2.6, 4.0],
        xerr=[0.18, 0.24, 0.20, 0.30, 0.22, 0.26],
        yerr=[0.35, 0.42, 0.25, 0.50, 0.38, 0.45],
        fmt="o-",
        color=(0.09, 0.75, 0.81, 1),
        capsize=4.0,
        markersize=4.5,
        errorevery=(1, 2),
    )

    save(fig, out_dir, "axes_option_breadth")


PLOT = axes_option_breadth
