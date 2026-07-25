#!/usr/bin/env python3
"""Matplotlib reference for the direct Welch PSD helper."""

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

import matplotlib.mlab as mlab


def signal():
    t = np.arange(256, dtype=np.float64) / 64.0
    return np.sin(2 * np.pi * 8 * t) + 0.35 * np.sin(2 * np.pi * 15 * t + 0.2) + 0.1 * np.cos(2 * np.pi * 3 * t)


def plot_psd_welch(out_dir):
    fig = make_fig_px(640, 360)
    ax = fig.add_axes(go_rect(0.125, 0.11, 0.9, 0.88))
    ax.set_title("Welch Power Spectral Density")
    ax.psd(signal(), NFFT=64, Fs=64, Fc=2, noverlap=32, pad_to=128,
           window=mlab.window_hanning, detrend=mlab.detrend_mean,
           color=(0.12, 0.47, 0.71), linewidth=1.5)
    save(fig, out_dir, "psd_welch")


PLOT = plot_psd_welch


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    PLOT(parser.parse_args().output_dir)
