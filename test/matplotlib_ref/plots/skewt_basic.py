#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
from contextlib import ExitStack
import sys

from matplotlib.axes import Axes
import matplotlib.axis as maxis
from matplotlib.projections import register_projection
import matplotlib.spines as mspines
from matplotlib import transforms
from matplotlib.ticker import FixedLocator, MultipleLocator, NullFormatter, ScalarFormatter

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def interval_contains(interval, val):
    contains = getattr(transforms, "interval_contains", None)
    if contains is None:
        contains = transforms._interval_contains
    return contains(interval, val)


class SkewXTick(maxis.XTick):
    def draw(self, renderer):
        with ExitStack() as stack:
            for artist in [self.gridline, self.tick1line, self.tick2line, self.label1, self.label2]:
                stack.callback(artist.set_visible, artist.get_visible())
            needs_lower = interval_contains(self.axes.lower_xlim, self.get_loc())
            needs_upper = interval_contains(self.axes.upper_xlim, self.get_loc())
            self.tick1line.set_visible(self.tick1line.get_visible() and needs_lower)
            self.label1.set_visible(self.label1.get_visible() and needs_lower)
            self.tick2line.set_visible(self.tick2line.get_visible() and needs_upper)
            self.label2.set_visible(self.label2.get_visible() and needs_upper)
            super().draw(renderer)

    def get_view_interval(self):
        return self.axes.xaxis.get_view_interval()


class SkewXAxis(maxis.XAxis):
    def _get_tick(self, major):
        return SkewXTick(self.axes, None, major=major)

    def get_view_interval(self):
        return self.axes.upper_xlim[0], self.axes.lower_xlim[1]


class SkewSpine(mspines.Spine):
    def _adjust_location(self):
        pts = self._path.vertices
        if self.spine_type == "top":
            pts[:, 0] = self.axes.upper_xlim
        else:
            pts[:, 0] = self.axes.lower_xlim


class SkewXAxes(Axes):
    name = "skewx"

    def _init_axis(self):
        self.xaxis = SkewXAxis(self)
        self.spines.top.register_axis(self.xaxis)
        self.spines.bottom.register_axis(self.xaxis)
        self.yaxis = maxis.YAxis(self)
        self.spines.left.register_axis(self.yaxis)
        self.spines.right.register_axis(self.yaxis)

    def _gen_axes_spines(self):
        return {
            "top": SkewSpine.linear_spine(self, "top"),
            "bottom": mspines.Spine.linear_spine(self, "bottom"),
            "left": mspines.Spine.linear_spine(self, "left"),
            "right": mspines.Spine.linear_spine(self, "right"),
        }

    def _set_lim_and_transforms(self):
        super()._set_lim_and_transforms()
        skew = transforms.Affine2D().skew_deg(30, 0)
        self.transDataToAxes = self.transScale + self.transLimits + skew
        self.transData = self.transDataToAxes + self.transAxes
        self._xaxis_transform = (
            transforms.blended_transform_factory(
                self.transScale + self.transLimits,
                transforms.IdentityTransform(),
            )
            + skew
            + self.transAxes
        )

    @property
    def lower_xlim(self):
        return self.axes.viewLim.intervalx

    @property
    def upper_xlim(self):
        pts = [[0.0, 1.0], [1.0, 1.0]]
        return self.transDataToAxes.inverted().transform(pts)[:, 0]


register_projection(SkewXAxes)


def skewt_basic(out_dir):
    fig = make_fig_px(720, 640)
    ax = fig.add_axes(go_rect(0.16, 0.14, 0.88, 0.88), projection="skewx")
    ax.set(title="Skew-T Style Projection", xlabel="Temperature (deg C)", ylabel="Pressure (hPa)")
    ax.set_yscale("log")
    ax.set_xlim(-70, 35)
    ax.set_ylim(1050, 180)
    ax.xaxis.set_major_locator(MultipleLocator(10))
    ax.xaxis.set_minor_locator(MultipleLocator(5))
    ax.yaxis.set_major_locator(FixedLocator([100, 200, 300, 500, 700, 850, 1000]))
    ax.yaxis.set_major_formatter(ScalarFormatter())
    ax.yaxis.set_minor_formatter(NullFormatter())
    ax.grid(color=(0.82, 0.84, 0.88, 1.0), linewidth=0.8)

    pressure = np.array([1000, 925, 850, 700, 600, 500, 400, 300, 250, 200])
    temperature = np.array([24, 20, 15, 5, -4, -14, -28, -43, -51, -58])
    dewpoint = np.array([18, 14, 8, -4, -14, -25, -38, -50, -57, -64])
    ax.plot(temperature, pressure, color=(0.78, 0.13, 0.16, 1.0), linewidth=2.4, label="temperature")
    ax.plot(dewpoint, pressure, color=(0.05, 0.48, 0.28, 1.0), linewidth=2.4, label="dewpoint")
    ax.legend()

    save(fig, out_dir, "skewt_basic")


PLOT = skewt_basic


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
