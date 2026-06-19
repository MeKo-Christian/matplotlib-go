#!/usr/bin/env python3
"""Matplotlib reference plot for spectrum variant helpers."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def spectrum_signal():
    t = np.arange(128, dtype=np.float64) / 64.0
    return np.sin(2 * np.pi * 5.25 * t) + 0.35 * np.cos(2 * np.pi * 12.4 * t + 0.4)


def plot_spectrum_variants(out_dir):
    fig = make_fig_px(900, 640)
    x = spectrum_signal()
    window = np.ones(len(x), dtype=np.float64)

    mag_ax = fig.add_axes(go_rect(0.08, 0.72, 0.96, 0.93))
    mag_ax.set_title("Magnitude Spectrum")
    mag_ax.magnitude_spectrum(
        x,
        Fs=64,
        window=window,
        scale="dB",
        color=(0.12, 0.47, 0.71),
        linewidth=lw(1.8),
    )
    mag_ax.grid(axis="y")

    angle_ax = fig.add_axes(go_rect(0.08, 0.41, 0.96, 0.62))
    angle_ax.set_title("Angle Spectrum")
    angle_ax.angle_spectrum(
        x,
        Fs=64,
        Fc=4,
        window=window,
        sides="twosided",
        color=(1.00, 0.50, 0.05),
        linewidth=lw(1.8),
    )
    angle_ax.grid(axis="y")

    phase_ax = fig.add_axes(go_rect(0.08, 0.10, 0.96, 0.31))
    phase_ax.set_title("Phase Spectrum")
    phase_ax.phase_spectrum(
        x,
        Fs=64,
        window=window,
        sides="onesided",
        color=(0.17, 0.63, 0.17),
        linewidth=lw(1.8),
    )
    phase_ax.grid(axis="y")

    save(fig, out_dir, "spectrum_variants")


PLOT = plot_spectrum_variants


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
