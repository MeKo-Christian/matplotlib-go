from __future__ import annotations

import importlib

DEMO_SPECS = [
    ("lines", "demo_lines"),
    ("scatter", "demo_scatter"),
    ("bars", "demo_bars"),
    ("fills", "demo_fills"),
    ("errorbars", "demo_errorbars"),
    ("histogram", "demo_histogram"),
    ("heatmap", "demo_heatmap"),
    ("axes", "demo_axes"),
    ("composition", "demo_composition"),
    ("subplots", "demo_subplots"),
    ("colorbars", "demo_colorbars"),
    ("annotations", "demo_annotations"),
    ("patches", "demo_patches"),
    ("mesh", "demo_mesh"),
    ("variants", "demo_variants"),
    ("statistics", "demo_statistics"),
    ("specialty", "demo_specialty"),
    ("units", "demo_units"),
    ("vectors", "demo_vectors"),
    ("polar", "demo_polar"),
    ("projections", "demo_projections"),
    ("toolkit", "demo_toolkit"),
    ("mplot3d", "demo_mplot3d"),
    ("triangulation", "demo_triangulation"),
    ("matrix", "demo_matrix"),
    ("animation", "demo_animation"),
    ("axisartist", "demo_axisartist"),
    ("axes_grid1", "demo_axes_grid1"),
]
DEMO_NAMES = [name for name, _ in DEMO_SPECS]

def load_demo(name: str):
    for demo_name, func_name in DEMO_SPECS:
        if demo_name == name:
            module = importlib.import_module(f"{__name__}.{demo_name}")
            return getattr(module, "DEMO")
    raise KeyError(name)

def all_demos():
    return [(name, load_demo(name)) for name in DEMO_NAMES]
