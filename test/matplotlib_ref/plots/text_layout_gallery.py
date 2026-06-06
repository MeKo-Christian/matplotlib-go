#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.parity.text_layout_gallery.plot import text_layout_gallery
except ModuleNotFoundError:
    sys.path.insert(0, str(Path(__file__).resolve().parents[3]))
    from test.parity.text_layout_gallery.plot import text_layout_gallery

PLOT = text_layout_gallery


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
