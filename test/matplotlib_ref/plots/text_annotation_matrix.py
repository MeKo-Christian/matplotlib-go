#!/usr/bin/env python3
"""Focused text, annotation, and offset-box reference fixture."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.parity.text_annotation_matrix.plot import text_annotation_matrix
except ModuleNotFoundError:
    sys.path.insert(0, str(Path(__file__).resolve().parents[3]))
    from test.parity.text_annotation_matrix.plot import text_annotation_matrix

PLOT = text_annotation_matrix


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
