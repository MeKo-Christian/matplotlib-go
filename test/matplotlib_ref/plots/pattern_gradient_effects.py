#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/parity/pattern_gradient_effects/plot.py."""

from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.parity.pattern_gradient_effects.plot import PLOT, pattern_gradient_effects
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[2]))
    from parity.pattern_gradient_effects.plot import PLOT, pattern_gradient_effects  # noqa: F401
