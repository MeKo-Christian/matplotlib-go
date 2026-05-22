#!/usr/bin/env python3
# /// script
# requires-python = ">=3.10"
# dependencies = ["matplotlib>=3.7", "numpy"]
# ///
"""Generate Matplotlib reference images for the browser demo catalog.

The functions in this file intentionally mirror internal/webdemo/demo.go. They
are not imported by Go tests; they feed the parity viewer with PNG baselines for
the charts currently exposed by the WASM demo.
"""

import argparse
import datetime as dt
import math
import os
import sys

import matplotlib

matplotlib.use("Agg")
import matplotlib.dates as mdates
import matplotlib.pyplot as plt
import numpy as np
from matplotlib.patches import Circle, Ellipse, FancyArrow, Polygon, Rectangle
from matplotlib.sankey import Sankey
from matplotlib.ticker import LogLocator, MultipleLocator, ScalarFormatter
from mpl_toolkits.axes_grid1.inset_locator import inset_axes, mark_inset

DPI = 100
DEFAULT_WIDTH = 960
DEFAULT_HEIGHT = 540
MASK64 = (1 << 64) - 1


# Shared utilities used by split web demo modules.

def color(r, g, b, a=1.0):
    return (r, g, b, a)


def lw(px):
    return px * 72.0 / DPI


def ss(radius_px):
    radius_pt = radius_px * 72.0 / DPI
    return math.pi * radius_pt * radius_pt


def make_fig(width_px, height_px):
    fig = plt.figure(figsize=(width_px / DPI, height_px / DPI), dpi=DPI)
    fig.patch.set_facecolor("white")
    return fig


def rect(min_x=0.10, min_y=0.12, max_x=0.96, max_y=0.90):
    return [min_x, min_y, max_x - min_x, max_y - min_y]


def grid_rects(nrows, ncols, left, right, bottom, top, wspace, hspace, width_ratios=None, height_ratios=None):
    """Return axes rectangles using matplotlib-go's figure-normalized GridSpec math."""
    width_ratios = list(width_ratios or [1.0] * ncols)
    height_ratios = list(height_ratios or [1.0] * nrows)
    inner_w = right - left
    inner_h = top - bottom
    available_w = inner_w - wspace * (ncols - 1)
    available_h = inner_h - hspace * (nrows - 1)
    widths = [available_w * r / sum(width_ratios) for r in width_ratios]
    heights = [available_h * r / sum(height_ratios) for r in height_ratios]

    out = []
    for row in range(nrows):
        row_rects = []
        y_top = top - sum(heights[:row]) - hspace * row
        y0 = y_top - heights[row]
        for col in range(ncols):
            x0 = left + sum(widths[:col]) + wspace * col
            row_rects.append([x0, y0, widths[col], heights[row]])
        out.append(row_rects)
    return out


def span_rect(rects, row, col, row_span, col_span):
    x0 = rects[row][col][0]
    y0 = rects[row + row_span - 1][col][1]
    x1 = rects[row][col + col_span - 1][0] + rects[row][col + col_span - 1][2]
    y1 = rects[row][col][1] + rects[row][col][3]
    return [x0, y0, x1 - x0, y1 - y0]


def linspace(start, stop, n):
    return np.linspace(start, stop, n)


def save(fig, out_dir, name):
    path = os.path.join(out_dir, f"{name}.png")
    fig.savefig(path, dpi=DPI, facecolor="white", bbox_inches=None)
    plt.close(fig)
    print(f"wrote {path}")


def _pcg_step(hi, lo):
    mul_hi = 2549297995355413924
    mul_lo = 4865540595714422341
    inc_hi = 6364136223846793005
    inc_lo = 1442695040888963407
    mul_lo_lo = (lo * mul_lo) & MASK64
    mul_lo_hi = (lo * mul_lo) >> 64
    hi_new = (mul_lo_hi + (hi * mul_lo) + (lo * mul_hi)) & MASK64
    lo_new = mul_lo_lo + inc_lo
    carry = 1 if lo_new > MASK64 else 0
    lo_new &= MASK64
    hi_new = (hi_new + inc_hi + carry) & MASK64
    return hi_new, lo_new


def _pcg_float64(hi, lo):
    hi, lo = _pcg_step(hi, lo)
    cheap_mul = 0xDA942042E4DD58B5
    val = hi ^ (hi >> 32)
    val = (val * cheap_mul) & MASK64
    val ^= val >> 48
    val = (val * (lo | 1)) & MASK64
    return (val & ((1 << 53) - 1)) / float(1 << 53), hi, lo


def normal_sample(state):
    hi, lo = state
    u1, hi, lo = _pcg_float64(hi, lo)
    u2, hi, lo = _pcg_float64(hi, lo)
    if u1 == 0:
        u1 = sys.float_info.min
    return math.sqrt(-2 * math.log(u1)) * math.cos(2 * math.pi * u2), (hi, lo)


def scatter_cluster(seed1, seed2, center_x, center_y, n):
    state = (seed1 & MASK64, seed2 & MASK64)
    x, y = [], []
    for _ in range(n):
        sx, state = normal_sample(state)
        sy, state = normal_sample(state)
        x.append(center_x + 0.65 * sx)
        y.append(center_y + 0.55 * sy)
    return np.array(x), np.array(y)


def deterministic_normal(n, mean, sigma):
    state = (42, 7)
    out = []
    for _ in range(n):
        sample, state = normal_sample(state)
        out.append(mean + sigma * sample)
    return np.array(out)


def vector_grid(x, y):
    u = np.zeros((len(y), len(x)))
    v = np.zeros((len(y), len(x)))
    for yi, yv in enumerate(y):
        for xi, xv in enumerate(x):
            u[yi, xi] = 1.0 + 0.12 * math.cos(yv * 0.7)
            v[yi, xi] = 0.35 * math.sin((xv - 3) * 0.8) - 0.10 * (yv - 2.5)
    return u, v


def heatmap_data(rows=28, cols=36):
    data = np.zeros((rows, cols))
    for row in range(rows):
        y = -1 + 2 * row / float(rows - 1)
        for col in range(cols):
            x = -1 + 2 * col / float(cols - 1)
            r1 = math.hypot(x + 0.35, y - 0.15)
            r2 = math.hypot(x - 0.4, y + 0.2)
            data[row, col] = math.sin(8 * r1) / (1 + 3 * r1) + 0.8 * math.cos(7 * r2)
    return data

__all__ = [name for name in globals() if not name.startswith("__")]
