#!/usr/bin/env python3
"""Extract a stable public-surface inventory from vendored Matplotlib."""

from __future__ import annotations

import ast
import json
from dataclasses import dataclass, asdict
from pathlib import Path


SOURCE_ROOT = Path("third_party/matplotlib/lib/matplotlib")

MODULES = [
    "artist.py",
    "axis.py",
    "ticker.py",
    "scale.py",
    "transforms.py",
    "lines.py",
    "markers.py",
    "collections.py",
    "patches.py",
    "text.py",
    "legend.py",
    "offsetbox.py",
    "image.py",
    "colorbar.py",
    "cm.py",
    "colors.py",
    "pyplot.py",
    "backend_bases.py",
    "backend_tools.py",
    "widgets.py",
    "animation.py",
]


@dataclass(frozen=True)
class Row:
    id: str
    module: str
    kind: str
    name: str


def main() -> None:
    rows: list[Row] = []
    for module in MODULES:
        tree = ast.parse((SOURCE_ROOT / module).read_text(), filename=module)
        rows.extend(top_level_rows(module, tree))
        rows.extend(registry_rows(module, tree))

    rows = sorted(set(rows), key=lambda row: row.id)
    artifact = {
        "schema_version": 1,
        "source_root": str(SOURCE_ROOT),
        "modules": MODULES,
        "rows": [asdict(row) for row in rows],
    }
    print(json.dumps(artifact, indent=2, sort_keys=True))


def top_level_rows(module: str, tree: ast.Module) -> list[Row]:
    rows: list[Row] = []
    for node in tree.body:
        if isinstance(node, ast.ClassDef) and is_public(node.name):
            rows.append(row(module, "class", node.name))
        elif isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)) and is_public(node.name):
            rows.append(row(module, "function", node.name))
        elif isinstance(node, (ast.Assign, ast.AnnAssign)):
            for name in assignment_names(node):
                if is_public_constant(name):
                    rows.append(row(module, "constant", name))
    return rows


def registry_rows(module: str, tree: ast.Module) -> list[Row]:
    if module == "image.py":
        return dict_registry_rows(module, tree, "_interpd_", "interpolation")
    if module == "scale.py":
        return dict_registry_rows(module, tree, "_scale_mapping", "scale")
    if module == "backend_tools.py":
        return dict_registry_rows(module, tree, "default_tools", "tool")
    if module == "markers.py":
        return marker_registry_rows(module, tree)
    if module == "lines.py":
        return line_registry_rows(module, tree)
    if module == "patches.py":
        return patch_style_rows(module, tree)
    return []


def marker_registry_rows(module: str, tree: ast.Module) -> list[Row]:
    rows: list[Row] = []
    marker_style = find_class(tree, "MarkerStyle")
    if marker_style is None:
        return rows
    for node in marker_style.body:
        if isinstance(node, ast.Assign):
            names = assignment_names(node)
            if "markers" in names:
                rows.extend(registry_values(module, "marker", dict_keys(node.value)))
            elif "fillstyles" in names:
                rows.extend(registry_values(module, "fillstyle", sequence_values(node.value)))
    return rows


def line_registry_rows(module: str, tree: ast.Module) -> list[Row]:
    rows: list[Row] = []
    line2d = find_class(tree, "Line2D")
    if line2d is None:
        return rows
    for node in line2d.body:
        if isinstance(node, ast.Assign):
            names = assignment_names(node)
            if "_lineStyles" in names or "lineStyles" in names:
                rows.extend(registry_values(module, "linestyle", dict_keys(node.value)))
            elif "drawStyles" in names:
                rows.extend(registry_values(module, "drawstyle", dict_keys(node.value)))
    return rows


def patch_style_rows(module: str, tree: ast.Module) -> list[Row]:
    rows: list[Row] = []
    registry_by_parent = {
        "BoxStyle": "boxstyle",
        "ConnectionStyle": "connectionstyle",
        "ArrowStyle": "arrowstyle",
    }
    for parent in tree.body:
        if not isinstance(parent, ast.ClassDef):
            continue
        registry = registry_by_parent.get(parent.name)
        if registry is None:
            continue
        for child in parent.body:
            if not isinstance(child, ast.ClassDef):
                continue
            style_name = registered_style_name(child) or child.name.lower()
            rows.append(row(module, f"registry:{registry}", style_name))
    return rows


def dict_registry_rows(module: str, tree: ast.Module, variable: str, registry: str) -> list[Row]:
    for node in tree.body:
        if isinstance(node, ast.Assign) and variable in assignment_names(node):
            return registry_values(module, registry, dict_keys(node.value))
    return []


def registry_values(module: str, registry: str, values: list[str]) -> list[Row]:
    return [row(module, f"registry:{registry}", value) for value in values]


def row(module: str, kind: str, name: str) -> Row:
    return Row(
        id=f"{module}:{kind}:{name}",
        module=module,
        kind=kind,
        name=name,
    )


def find_class(tree: ast.Module, name: str) -> ast.ClassDef | None:
    for node in tree.body:
        if isinstance(node, ast.ClassDef) and node.name == name:
            return node
    return None


def assignment_names(node: ast.Assign | ast.AnnAssign) -> list[str]:
    targets = node.targets if isinstance(node, ast.Assign) else [node.target]
    names: list[str] = []
    for target in targets:
        if isinstance(target, ast.Name):
            names.append(target.id)
    return names


def dict_keys(node: ast.AST) -> list[str]:
    if isinstance(node, ast.Dict):
        return [literal_name(key) for key in node.keys if key is not None and literal_name(key) != ""]
    return []


def sequence_values(node: ast.AST) -> list[str]:
    if isinstance(node, (ast.Tuple, ast.List, ast.Set)):
        return [literal_name(elt) for elt in node.elts if literal_name(elt) != ""]
    return []


def registered_style_name(node: ast.ClassDef) -> str:
    for decorator in node.decorator_list:
        if not isinstance(decorator, ast.Call):
            continue
        if not isinstance(decorator.func, ast.Name) or decorator.func.id != "_register_style":
            continue
        for keyword in decorator.keywords:
            if keyword.arg == "name":
                return literal_name(keyword.value)
    return ""


def literal_name(node: ast.AST) -> str:
    if isinstance(node, ast.Constant):
        return str(node.value)
    if isinstance(node, ast.Name):
        return node.id
    return ""


def is_public(name: str) -> bool:
    return not name.startswith("_")


def is_public_constant(name: str) -> bool:
    return is_public(name) and (name.isupper() or name in {"lineStyles", "drawStyles"})


if __name__ == "__main__":
    main()
