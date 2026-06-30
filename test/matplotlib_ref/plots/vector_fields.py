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

def vector_fields(out_dir):
    fig = make_fig_px(919, 620)
    axes = {
        "quiver": fig.add_axes(go_rect(0.07, 0.58, 0.47, 0.92)),
        "barbs": fig.add_axes(go_rect(0.57, 0.58, 0.97, 0.92)),
        "stream": fig.add_axes(go_rect(0.07, 0.10, 0.47, 0.44)),
        "xy": fig.add_axes(go_rect(0.57, 0.10, 0.97, 0.44)),
    }

    quiver_ax = axes["quiver"]
    quiver_ax.set_title("Quiver + Key")
    quiver_ax.set_xlim(0, 6)
    quiver_ax.set_ylim(0, 5)
    quiver_ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
    quiver_ax.set_axisbelow(True)
    qx, qy, qu, qv = [], [], [], []
    for row in range(4):
        for col in range(5):
            x = 0.8 + col * 1.0
            y = 0.8 + row * 0.95
            qx.append(x)
            qy.append(y)
            qu.append(0.55 + 0.08 * math.sin(y * 0.9))
            qv.append(0.22 * math.cos(x * 0.8))
    q = quiver_ax.quiver(
        qx, qy, qu, qv,
        color=(0.14, 0.42, 0.73),
        scale=10.0,
        scale_units="width",
        units="dots",
        width=2.2,
    )
    quiver_ax.quiverkey(q, 0.78, 0.12, 0.5, "0.5", coordinates="axes", labelpos="E")

    barb_ax = axes["barbs"]
    barb_ax.set_title("Barbs")
    barb_ax.set_xlim(0, 6)
    barb_ax.set_ylim(0, 5)
    barb_ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
    barb_ax.set_axisbelow(True)
    bx, by, bu, bv = [], [], [], []
    for row in range(4):
        for col in range(5):
            x = 0.9 + col * 0.95
            y = 0.8 + row * 0.95
            bx.append(x)
            by.append(y)
            bu.append(14 + 5 * math.sin(y * 0.8))
            bv.append(8 * math.cos(x * 0.7))
    barb_ax.barbs(
        bx, by, bu, bv,
        barbcolor=(0.47, 0.23, 0.12),
        flagcolor=(0.86, 0.52, 0.24),
        length=6.0,
        linewidth=1.0,
    )

    stream_ax = axes["stream"]
    stream_ax.set_title("Streamplot")
    stream_ax.set_xlim(0, 6)
    stream_ax.set_ylim(0, 5)
    stream_ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
    stream_ax.set_axisbelow(True)
    sx = np.array([0, 1, 2, 3, 4, 5, 6], dtype=float)
    sy = np.array([0, 1, 2, 3, 4, 5], dtype=float)
    su = np.zeros((len(sy), len(sx)))
    sv = np.zeros((len(sy), len(sx)))
    for yi, y in enumerate(sy):
        for xi, x in enumerate(sx):
            su[yi, xi] = 1.0 + 0.12 * math.cos(y * 0.7)
            sv[yi, xi] = 0.35 * math.sin((x - 3) * 0.8) - 0.10 * (y - 2.5)
    stream_ax.streamplot(
        sx, sy, su, sv,
        start_points=np.array([[0.4, 0.8], [0.4, 2.2], [0.4, 3.6]], dtype=float),
        broken_streamlines=False,
        integration_direction="forward",
        color=(0.13, 0.53, 0.39),
        linewidth=1.5,
        arrowsize=1.0,
    )

    xy_ax = axes["xy"]
    xy_ax.set_title("Quiver XY")
    xy_ax.set_xlim(0, 6)
    xy_ax.set_ylim(0, 5)
    xy_ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=0.5)
    xy_ax.set_axisbelow(True)
    xg = np.array([0.8, 1.8, 2.8, 3.8, 4.8], dtype=float)
    yg = np.array([0.8, 1.8, 2.8, 3.8], dtype=float)
    ugu = np.zeros((len(yg), len(xg)))
    ugv = np.zeros((len(yg), len(xg)))
    for yi, y in enumerate(yg):
        for xi, x in enumerate(xg):
            ugu[yi, xi] = -(y - 2.4) * 0.35
            ugv[yi, xi] = (x - 2.8) * 0.35
    xy_ax.quiver(
        xg, yg, ugu, ugv,
        color=(0.74, 0.23, 0.27),
        pivot="middle",
        angles="xy",
        scale=9.0,
        scale_units="width",
        units="dots",
        width=1.9,
    )

    save(fig, out_dir, "vector_fields")

PLOT = vector_fields


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
