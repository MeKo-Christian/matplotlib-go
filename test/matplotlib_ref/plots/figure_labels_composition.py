#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

def figure_labels_composition(out_dir):
    fig, axs = plt.subplots(2, 2, figsize=(1100 / DPI, 720 / DPI), dpi=DPI, constrained_layout=True)
    fig.suptitle("Shared-Axis Figure Labels")
    fig.supxlabel("time [s]")
    fig.supylabel("amplitude")

    handles = []
    labels = []
    for row in range(2):
        for col in range(2):
            ax = axs[row, col]
            x = np.linspace(0, 2 * math.pi, 180)
            y = np.sin(x + row * 0.5) * (1 + col * 0.2)
            (line,) = ax.plot(x, y, label=f"series {row * 2 + col + 1}")
            handles.append(line)
            labels.append(line.get_label())
            ax.set_title(f"Panel {row * 2 + col + 1}")
            ax.set_xlabel("local x")
            ax.set_ylabel("local y")
            ax.set_xlim(0, 2 * math.pi)
            ax.set_ylim(-1.6, 1.6)
            ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
            ax.set_axisbelow(True)

    axs[0, 0].text(
        0.02,
        0.92,
        "upper-left\nnote",
        transform=axs[0, 0].transAxes,
        va="top",
        bbox=dict(facecolor="white", edgecolor=(0.5, 0.5, 0.5)),
    )
    axs[1, 1].text(
        0.98,
        0.08,
        "lower-right",
        transform=axs[1, 1].transAxes,
        ha="right",
        va="bottom",
        bbox=dict(facecolor="white", edgecolor=(0.5, 0.5, 0.5)),
    )
    fig.text(
        0.985,
        0.94,
        "Figure note",
        ha="right",
        va="top",
        bbox=dict(facecolor="white", edgecolor=(0.5, 0.5, 0.5)),
    )
    fig.legend(handles, labels, loc="upper right", bbox_to_anchor=(0.99, 0.90))

    save(fig, out_dir, "figure_labels_composition")

PLOT = figure_labels_composition


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
