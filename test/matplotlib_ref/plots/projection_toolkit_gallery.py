#!/usr/bin/env python3
"""Grouped Matplotlib projection/toolkit reference gallery."""

from __future__ import annotations

from pathlib import Path
import argparse
from contextlib import ExitStack
import sys

from matplotlib.axes import Axes
import matplotlib.axis as maxis
from matplotlib.projections import register_projection
from matplotlib.projections.polar import PolarAxes
from matplotlib.spines import Spine
import matplotlib.spines as mspines
from matplotlib import transforms
from matplotlib.transforms import Affine2D
from matplotlib.ticker import FixedLocator, MultipleLocator, NullFormatter

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def panel(row, col):
    left, right = 0.055, 0.965
    bottom, top = 0.075, 0.935
    hgap, vgap = 0.060, 0.095
    w = (right - left - 2 * hgap) / 3
    h = (top - bottom - 2 * vgap) / 3
    x0 = left + col * (w + hgap)
    y1 = top - row * (h + vgap)
    return go_rect(x0, y1 - h, x0 + w, y1)


def radar_factory(num_vars):
    theta = np.linspace(0, 2 * math.pi, num_vars, endpoint=False)

    class RadarTransform(PolarAxes.PolarTransform):
        def transform_path_non_affine(self, path):
            if path._interpolation_steps > 1:
                path = path.interpolated(num_vars)
            return mpath.Path(self.transform(path.vertices), path.codes)

    class RadarAxes(PolarAxes):
        name = "radar_gallery"
        PolarTransform = RadarTransform

        def __init__(self, *args, **kwargs):
            super().__init__(*args, **kwargs)
            self.set_theta_zero_location("N")

        def fill(self, *args, closed=True, **kwargs):
            return super().fill(closed=closed, *args, **kwargs)

        def plot(self, *args, **kwargs):
            lines = super().plot(*args, **kwargs)
            for line in lines:
                x, y = line.get_data()
                if x[0] != x[-1]:
                    line.set_data(np.append(x, x[0]), np.append(y, y[0]))
            return lines

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
    name = "skewx_gallery"

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
            transforms.blended_transform_factory(self.transScale + self.transLimits, transforms.IdentityTransform())
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


def add_polar(fig):
    ax = fig.add_axes(panel(0, 0), projection="polar")
    ax.set_title("Polar")
    ax.set_theta_zero_location("N")
    ax.set_theta_direction(-1)
    ax.set_ylim(0, 1.15)
    ax.set_yticks([0.25, 0.5, 0.75, 1.0])
    ax.set_yticklabels(["25%", "50%", "75%", "100%"])
    ax.grid(color=(0.80, 0.82, 0.86, 1.0), linewidth=lw(0.8))
    theta = np.linspace(0, 2 * math.pi, 241)
    radius = 0.62 + 0.28 * np.sin(3 * theta + 0.35)
    ax.fill(theta, radius, color=(0.18, 0.50, 0.82, 0.22))
    ax.plot(theta, radius, color=(0.14, 0.34, 0.70, 1.0), linewidth=lw(2.0))


def add_geo(fig, row, col, projection, title, lon_min, lon_max):
    ax = fig.add_axes(panel(row, col), projection=projection)
    ax.set_title(title)
    ax.set_xlabel("lon")
    ax.set_ylabel("lat")
    if projection == "lambert":
        ax.set_xticks(np.arange(-120, 121, 30) * math.pi / 180.0)
    ax.grid(color=(0.80, 0.82, 0.86, 1.0), linewidth=lw(0.7))
    lon = np.linspace(lon_min, lon_max, 241)
    lat = 0.35 * np.sin(3.0 * lon)
    ax.plot(lon, lat, color=(0.14, 0.34, 0.70, 1.0), linewidth=lw(1.8))


def add_radar(fig):
    labels = ["Speed", "Power", "Range", "Handling", "Comfort"]
    values = np.array([0.72, 0.88, 0.64, 0.79, 0.58])
    angles = radar_factory(len(labels))
    closed_angles = np.r_[angles, angles[0]]
    closed_values = np.r_[values, values[0]]
    ax = fig.add_axes(panel(1, 2), projection="radar_gallery")
    ax.set_title("Radar")
    ax.set_varlabels(labels)
    ax.set_ylim(0, 1)
    ax.set_yticks([0.25, 0.5, 0.75, 1.0])
    ax.set_yticklabels(["25%", "50%", "75%", "100%"])
    ax.grid(color=(0.80, 0.83, 0.88, 1.0), linewidth=lw(0.75))
    ax.fill(closed_angles, closed_values, color=(0.18, 0.50, 0.82, 0.22))
    ax.plot(closed_angles, closed_values, color=(0.14, 0.34, 0.70, 1.0), linewidth=lw(2.0))


def add_skewt(fig):
    ax = fig.add_axes(panel(2, 0), projection="skewx_gallery")
    ax.set(title="Skew-T", xlabel="temp", ylabel="pressure")
    ax.set_yscale("log")
    ax.set_xlim(-70, 35)
    ax.set_ylim(1050, 180)
    ax.xaxis.set_major_locator(MultipleLocator(20))
    ax.xaxis.set_minor_locator(MultipleLocator(10))
    ax.yaxis.set_major_locator(FixedLocator([200, 300, 500, 700, 850, 1000]))
    ax.yaxis.set_minor_formatter(NullFormatter())
    ax.grid(color=(0.80, 0.82, 0.86, 1.0), linewidth=lw(0.75))
    pressure = np.array([1000, 925, 850, 700, 600, 500, 400, 300, 250, 200])
    temperature = np.array([24, 20, 15, 5, -4, -14, -28, -43, -51, -58])
    dewpoint = np.array([18, 14, 8, -4, -14, -25, -38, -50, -57, -64])
    ax.plot(temperature, pressure, color=(0.78, 0.13, 0.16, 1.0), linewidth=lw(2.0), label="temp")
    ax.plot(dewpoint, pressure, color=(0.05, 0.48, 0.28, 1.0), linewidth=lw(2.0), label="dew")
    ax.legend()


def add_axisartist(fig):
    ax = fig.add_axes(panel(2, 1))
    ax.set_title("AxisArtist / Twin")
    ax.set_xlabel("phase")
    ax.set_ylabel("signal")
    ax.set_xlim(-3.5, 3.5)
    ax.set_ylim(-1.3, 1.3)
    ax.grid(axis="y", color=(0.80, 0.82, 0.86, 1.0), linewidth=lw(0.75))
    x = np.linspace(-3.5, 3.5, 180)
    ax.plot(x, np.sin(x), color=(0.14, 0.34, 0.70, 1.0), linewidth=lw(2.0), label="sin")
    ax.axhline(0, color=(0.26, 0.26, 0.30, 1.0), linewidth=lw(1.2), dashes=[5 * 36.0 / DPI, 3 * 36.0 / DPI])
    ax.axvline(0, color=(0.26, 0.26, 0.30, 1.0), linewidth=lw(1.2), dashes=[5 * 36.0 / DPI, 3 * 36.0 / DPI])
    right = ax.twinx()
    right.set_ylim(0, 100)
    right.plot(x, 55 + 35 * np.cos(x * 0.8), color=(0.74, 0.28, 0.18, 1.0), linewidth=lw(2.0), label="scaled cos")
    right.spines["right"].set_color((0.74, 0.28, 0.18, 1.0))
    right.tick_params(axis="y", colors=(0.74, 0.28, 0.18, 1.0))
    ax.text(
        0.03,
        0.97,
        "floating reference\nparasite scale",
        transform=ax.transAxes,
        va="top",
        fontsize=8,
        bbox=dict(boxstyle="round,pad=0.25", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
    )


def surface(rows, cols, phase):
    data = np.zeros((rows, cols))
    for y in range(rows):
        yy = y / float(rows - 1)
        for x in range(cols):
            xx = x / float(cols - 1)
            data[y, x] = 0.5 + 0.25 * math.sin((xx + phase) * 2 * math.pi) + 0.25 * math.cos((yy + phase * 0.3) * 3 * math.pi)
    return data


def add_axes_grid(fig):
    outer = fig.add_axes(panel(2, 2))
    outer.set_title("axes_grid1")
    outer.set_xticks([])
    outer.set_yticks([])
    outer.set_frame_on(False)
    rect = panel(2, 2)
    x0, y0, w, h = rect
    grid = ImageGrid(
        fig,
        (x0 + 0.01, y0 + 0.02, w - 0.02, h - 0.07),
        nrows_ncols=(2, 2),
        axes_pad=(0.012 * 13.2, 0.018 * 9.0),
        share_all=False,
    )
    for idx, ax in enumerate(grid):
        row = idx // 2
        col = idx % 2
        ax.set_title("Tile")
        ax.set_xticks([0, 12, 23])
        ax.set_yticks([0, 12, 23])
        ax.imshow(surface(24, 24, float(row * 2 + col)), origin="upper")


def projection_toolkit_gallery(out_dir):
    fig = make_fig_px(1320, 900)
    add_polar(fig)
    add_geo(fig, 0, 1, "mollweide", "Mollweide", -math.pi, math.pi)
    add_geo(fig, 0, 2, "aitoff", "Aitoff", -math.pi, math.pi)
    add_geo(fig, 1, 0, "hammer", "Hammer", -math.pi, math.pi)
    add_geo(fig, 1, 1, "lambert", "Lambert", -math.pi / 2, math.pi / 2)
    add_radar(fig)
    add_skewt(fig)
    add_axisartist(fig)
    add_axes_grid(fig)
    save(fig, out_dir, "projection_toolkit_gallery")


PLOT = projection_toolkit_gallery


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
