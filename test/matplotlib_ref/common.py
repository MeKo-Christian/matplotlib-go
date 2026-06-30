#!/usr/bin/env python3
# /// script
# requires-python = ">=3.10"
# dependencies = ["matplotlib>=3.7"]
# ///
"""Generate matplotlib reference images for matplotlib-go test comparisons.

Each function renders the same plot as the corresponding golden test in
test/golden_test.go, using real matplotlib as the reference renderer.

Usage:
    uv run generate.py --output-dir /path/to/output
    python3 generate.py --output-dir /path/to/output
"""

import argparse
import datetime as dt
import json
import math
import os
import sys

import numpy as np
import matplotlib

matplotlib.use("Agg")
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
import matplotlib.colors as mcolors
import matplotlib.patches as mpatches
import matplotlib.path as mpath
import matplotlib.tri as mtri
from mpl_toolkits.axes_grid1 import ImageGrid, make_axes_locatable

DPI = 100
W_PX, H_PX = 640, 360

# Tab10 palette — must match color/palette.go
TAB10 = [
    (0.12, 0.47, 0.71),  # blue
    (1.00, 0.50, 0.05),  # orange
    (0.17, 0.63, 0.17),  # green
    (0.84, 0.15, 0.16),  # red
    (0.58, 0.40, 0.74),  # purple
    (0.55, 0.34, 0.29),  # brown
    (0.89, 0.47, 0.76),  # pink
    (0.50, 0.50, 0.50),  # gray
    (0.74, 0.74, 0.13),  # olive
    (0.09, 0.75, 0.81),  # cyan
]

MASK64 = (1 << 64) - 1


# Shared utilities used by split plot modules.

def _pcg_step(hi: int, lo: int) -> tuple[int, int]:
    """Advance Go's math/rand/v2.PCG state and return the next tuple.

    Mirrors the exact state transition from /usr/local/go1.25.0/src/math/rand/v2/pcg.go.
    """
    mul_hi = 2549297995355413924
    mul_lo = 4865540595714422341
    inc_hi = 6364136223846793005
    inc_lo = 1442695040888963407

    # bits.Mul64(p.lo, mulLo)
    mul_lo_lo = (lo * mul_lo) & MASK64
    mul_lo_hi = (lo * mul_lo) >> 64

    hi_new = (mul_lo_hi + (hi * mul_lo) + (lo * mul_hi)) & MASK64

    lo_new = mul_lo_lo + inc_lo
    carry = 1 if lo_new > MASK64 else 0
    lo_new &= MASK64
    hi_new = (hi_new + inc_hi + carry) & MASK64
    return hi_new, lo_new


def _pcg_uint64(hi: int, lo: int) -> tuple[int, int, int]:
    """Generate one PCG output value and return (value, next_hi, next_lo)."""
    hi, lo = _pcg_step(hi, lo)
    cheap_mul = 0xDA942042E4DD58B5
    val = hi ^ (hi >> 32)
    val = (val * cheap_mul) & MASK64
    val ^= val >> 48
    val = (val * (lo | 1)) & MASK64
    return val, hi, lo


def _pcg_float64(hi: int, lo: int) -> tuple[float, int, int]:
    """Return a Go-compatible rand.Float64 and updated PCG state.

    Go's Float64 uses:
    float64(r.Uint64()<<11>>11) / (1 << 53)
    """
    value, hi, lo = _pcg_uint64(hi, lo)
    return float((value & ((1 << 53) - 1)) / float(1 << 53)), hi, lo


def normal_data(seed1: int, seed2: int, n: int, mean: float, std: float) -> np.ndarray:
    """Generate normally distributed samples using the Go Box-Muller path."""
    hi, lo = seed1 & MASK64, seed2 & MASK64
    out = np.empty(n, dtype=np.float64)
    for _ in range(n):
        u1, hi, lo = _pcg_float64(hi, lo)
        u2, hi, lo = _pcg_float64(hi, lo)
        out[_] = math.sqrt(-2 * math.log(u1)) * math.cos(2 * math.pi * u2) * std + mean
    return out


def pcg_float64_values(seed1: int, seed2: int, n: int) -> list[float]:
    """Generate n Go-compatible Float64 values from the PCG stream."""
    hi, lo = seed1 & MASK64, seed2 & MASK64
    out = []
    for _ in range(n):
        value, hi, lo = _pcg_float64(hi, lo)
        out.append(value)
    return out


def histogram_payload(data: list[float], bins, density: bool = False, weights: np.ndarray | None = None) -> dict[str, list[float]]:
    """Return histogram bin edges and bin heights for parity comparisons."""
    counts, edges = np.histogram(data, bins=bins, density=density, weights=weights)
    return {
        "edges": edges.tolist(),
        "heights": counts.tolist(),
    }


def _to_list(values: np.ndarray | list[float]) -> list[float]:
    if isinstance(values, np.ndarray):
        return values.tolist()
    return values


def make_fig_px(width_px, height_px):
    """Figure with explicit pixel dimensions at the shared DPI."""
    fig = plt.figure(figsize=(width_px / DPI, height_px / DPI), dpi=DPI)
    fig.patch.set_facecolor("white")
    return fig


def make_fig():
    """640×360 figure at 100 DPI with white background."""
    return make_fig_px(W_PX, H_PX)


def go_rect(min_x, min_y, max_x, max_y):
    """Convert Go RectFraction to matplotlib add_axes rect.

    Go RectFraction and matplotlib add_axes both use bottom-left figure
    fractions. Matplotlib wants [left, bottom, width, height].
    """
    return [min_x, min_y, max_x - min_x, max_y - min_y]


def lw(width_pt):
    """Deprecated identity passthrough.

    The Go port now stores line widths in points (matplotlib semantics) and
    converts to device pixels at draw time, so references use matplotlib's
    native point values directly. Kept as a no-op safety net for any
    remaining call site.
    """
    return width_pt


def px2pt(px):
    """Convert a Go pixel-space length to matplotlib points.

    A few quantities (notably tick *length*) are still stored in device pixels
    by the Go port and coupled to its pixel-based label layout, so references
    bridge them px→pt. (Line/edge widths are now points end-to-end and need no
    bridge.) TODO: make tick length points-based and drop this.
    """
    return px * 72.0 / DPI


def ss(go_radius_px):
    """Desired scatter radius (pixels) → matplotlib scatter s (points²).

    Matplotlib scatter s ≈ π·r² where r is the radius in points.
    """
    r_pt = go_radius_px * 72.0 / DPI
    return math.pi * r_pt * r_pt


def save(fig, out_dir, name):
    path = os.path.join(out_dir, f"{name}.png")
    fig.savefig(path, dpi=DPI, facecolor="white", bbox_inches=None)
    plt.close(fig)
    print(f"  wrote {path}")


def _bar_basic_scaffold(show_ticks: bool, show_tick_labels: bool, show_title: bool):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 10)
    ax.tick_params(
        axis="both",
        which="both",
        bottom=show_ticks,
        left=show_ticks,
        labelbottom=show_tick_labels,
        labelleft=show_tick_labels,
    )
    if show_title:
        ax.set_title("Basic Bars")
    return fig, ax


def _composition_configure_axes(ax, title, x, y, color):
    ax.set_title(title, fontsize=12)
    ax.set_xlabel("x", fontsize=10)
    ax.set_ylabel("y", fontsize=10)
    ax.tick_params(labelsize=10)
    ax.plot(x, y, color=color, linewidth=2.0, label=title)
    ax.margins(0.10)

__all__ = [name for name in globals() if not name.startswith("__")]
