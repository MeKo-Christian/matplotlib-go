#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/parity/path_effects/plot.py."""

from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.parity.path_effects.plot import PLOT, path_effects
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2]))
    from parity.path_effects.plot import PLOT, path_effects  # noqa: F401
