from __future__ import annotations

import importlib

DEMO_SPECS = [('axes', 'demo_axes'), ('composition', 'demo_composition'), ('mesh', 'demo_mesh'), ('variants', 'demo_variants'), ('statistics', 'demo_statistics'), ('specialty', 'demo_specialty'), ('units', 'demo_units'), ('vectors', 'demo_vectors'), ('polar', 'demo_polar'), ('projections', 'demo_projections'), ('matrix', 'demo_matrix'), ('animation', 'demo_animation')]
DEMO_NAMES = [name for name, _ in DEMO_SPECS]

def load_demo(name: str):
    for demo_name, func_name in DEMO_SPECS:
        if demo_name == name:
            module = importlib.import_module(f"{__name__}.{demo_name}")
            return getattr(module, "DEMO")
    raise KeyError(name)

def all_demos():
    return [(name, load_demo(name)) for name in DEMO_NAMES]
