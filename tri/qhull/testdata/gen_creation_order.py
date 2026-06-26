#!/usr/bin/env python3
"""Capture Qhull's vertex CREATION ORDER for every corpus case.

Ground truth for the Qhull-faithful cocircular Delaunay port (PLAN.md Phase 12):
each Delaunay cell is fanned from its last-created (highest vertex->id) vertex, so
reproducing Qhull's diagonal choice reduces to reproducing the vertex creation
order. This script pipes every case from corpus.json through the `introspect` tool
(built from the vendored qhull 8.0.2 source) and records, per case, the real input
point ids in ascending vertex->id order (the Qz infinity point is dropped).

Build the tool first (gitignored):
    cc -O2 -I third_party/qhull-8.0.2/src third_party/qhull-8.0.2/introspect.c \
       third_party/qhull-8.0.2/src/libqhull_r/*.c -lm -o /tmp/introspect

Run (writes creation_order.json next to this file):
    QHULL_INTROSPECT=/tmp/introspect python3 gen_creation_order.py
"""
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))


def introspect_order(tool, x, y):
    """Return the real-point ids in ascending Qhull vertex->id (creation) order."""
    n = len(x)
    stdin = f"{n}\n" + "\n".join(f"{xi!r} {yi!r}" for xi, yi in zip(x, y, strict=True)) + "\n"
    out = subprocess.run([tool], input=stdin, capture_output=True, text=True, check=True).stdout
    line = next(row for row in out.splitlines() if row.startswith("VERTICES"))
    # tokens are "vid:pointid"; the infinity point has pointid == n.
    pairs = [tok.split(":") for tok in line.split()[1:]]
    pairs = [(int(vid), int(pid)) for vid, pid in pairs]
    pairs.sort()  # ascending vertex->id == creation order
    return [pid for _, pid in pairs if pid != n]


def main():
    tool = os.environ.get("QHULL_INTROSPECT", "/tmp/introspect")
    if not os.path.exists(tool):
        sys.exit(f"introspect tool not found at {tool}; set QHULL_INTROSPECT (see header)")
    corpus = json.load(open(os.path.join(HERE, "corpus.json")))
    order = {c["name"]: introspect_order(tool, c["x"], c["y"]) for c in corpus["cases"]}
    out = os.path.join(HERE, "creation_order.json")
    with open(out, "w") as f:
        json.dump({"qhull_version": corpus["qhull_version"], "order": order}, f, indent=0)
    print(f"wrote creation order for {len(order)} cases to {out}")


if __name__ == "__main__":
    main()
