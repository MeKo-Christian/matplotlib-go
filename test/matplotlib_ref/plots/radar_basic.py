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

from matplotlib.projections import register_projection
from matplotlib.projections.polar import PolarAxes
from matplotlib.spines import Spine
from matplotlib.transforms import Affine2D


def radar_factory(num_vars):
    theta = np.linspace(0, 2 * math.pi, num_vars, endpoint=False)

    class RadarTransform(PolarAxes.PolarTransform):
        def transform_path_non_affine(self, path):
            if path._interpolation_steps > 1:
                path = path.interpolated(num_vars)
            return mpath.Path(self.transform(path.vertices), path.codes)

    class RadarAxes(PolarAxes):
        name = "radar"
        PolarTransform = RadarTransform

        def __init__(self, *args, **kwargs):
            super().__init__(*args, **kwargs)
            self.set_theta_zero_location("N")

        def fill(self, *args, closed=True, **kwargs):
            return super().fill(closed=closed, *args, **kwargs)

        def plot(self, *args, **kwargs):
            lines = super().plot(*args, **kwargs)
            for line in lines:
                self._close_line(line)
            return lines

        def _close_line(self, line):
            x, y = line.get_data()
            if x[0] != x[-1]:
                x = np.append(x, x[0])
                y = np.append(y, y[0])
                line.set_data(x, y)

        def set_varlabels(self, labels):
            self.set_thetagrids(np.degrees(theta), labels)

        def _gen_axes_patch(self):
            return mpatches.RegularPolygon((0.5, 0.5), num_vars, radius=0.5, edgecolor="k")

        def _gen_axes_spines(self):
            spine = Spine(axes=self, spine_type="circle", path=mpath.Path.unit_regular_polygon(num_vars))
            spine.set_transform(Affine2D().scale(0.5).translate(0.5, 0.5) + self.transAxes)
            return {"polar": spine}

    register_projection(RadarAxes)
    return theta


def radar_basic(out_dir):
    labels = ["Speed", "Power", "Range", "Handling", "Comfort"]
    values = np.array([0.72, 0.88, 0.64, 0.79, 0.58])
    angles = radar_factory(len(labels))
    closed_angles = np.r_[angles, angles[0]]
    closed_values = np.r_[values, values[0]]

    fig = make_fig_px(720, 720)
    ax = fig.add_axes(go_rect(0.12, 0.10, 0.88, 0.88), projection="radar")
    ax.set_theta_zero_location("N")
    ax.set_title("Radar Projection")
    ax.set_varlabels(labels)
    ax.set_ylim(0, 1)
    ax.set_yticks([0.25, 0.5, 0.75, 1.0])
    ax.set_yticklabels(["25%", "50%", "75%", "100%"])
    ax.xaxis.grid(True, color=(0.78, 0.80, 0.84, 1.0), linewidth=0.8)
    ax.yaxis.grid(True, color=(0.80, 0.83, 0.88, 1.0), linewidth=0.8)

    ax.fill(closed_angles, closed_values, color=(0.18, 0.50, 0.82, 0.22))
    ax.plot(closed_angles, closed_values, color=(0.15, 0.35, 0.70, 1.0), linewidth=2.2, label="model A")

    save(fig, out_dir, "radar_basic")


PLOT = radar_basic


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
