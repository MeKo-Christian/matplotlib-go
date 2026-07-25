#!/usr/bin/env python3
"""Matplotlib reference for the direct PSD spectrogram helper."""

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
    t = np.arange(384, dtype=np.float64) / 64.0
    phase = 2 * np.pi * (3 * t + 0.75 * t * t)
    return 0.8 * np.sin(phase) + 0.25 * np.sin(2 * np.pi * 18 * t)


def plot_specgram_psd(out_dir):
    fig = make_fig_px(640, 360)
    ax = fig.add_axes(go_rect(0.125, 0.11, 0.9, 0.88))
    ax.set_title("PSD Spectrogram")
    ax.set_xlabel("Time")
    ax.set_ylabel("Frequency")
    ax.specgram(signal(), NFFT=64, Fs=64, noverlap=48, pad_to=128,
                window=mlab.window_hanning, cmap="viridis",
                vmin=-55, vmax=-5)
    save(fig, out_dir, "specgram_psd")


PLOT = plot_specgram_psd


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    PLOT(parser.parse_args().output_dir)
