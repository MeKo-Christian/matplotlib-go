#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import sys


try:
    from test.matplotlib_ref.plots.ticks_scales_formatters_gallery import PLOT
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2]))
    from test.matplotlib_ref.plots.ticks_scales_formatters_gallery import PLOT


if __name__ == "__main__":
    PLOT(".")
