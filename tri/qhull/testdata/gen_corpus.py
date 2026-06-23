#!/usr/bin/env python3
"""Generate the Qhull Delaunay differential-test corpus.

Source of truth: matplotlib.tri.Triangulation (Qhull 8.0.2, `qhull d Qt Qbb Qc Qz`).
Run:  python3 gen_corpus.py   (writes corpus.json next to this file)
Each case stores the ORIGINAL (x, y) plus Qhull's triangles and neighbors arrays
in their exact order/rotation, so the Go port can be diffed bit-for-bit.
"""
import json, os
import numpy as np
from matplotlib.tri import Triangulation

cases = []

def add(name, x, y, category):
    x = np.asarray(x, float); y = np.asarray(y, float)
    tr = Triangulation(x, y)
    cases.append({
        "name": name,
        "category": category,
        "x": x.tolist(),
        "y": y.tolist(),
        "triangles": tr.triangles.tolist(),
        "neighbors": tr.neighbors.tolist(),
    })

# ---- general position ----
add("mplMesh",
    [0.0,1.0,2.0,0.3,1.4,0.8,1.9,2.5,0.1,2.2],
    [0.0,0.2,0.0,1.1,0.9,1.8,1.6,1.0,2.0,2.1], "general")

for seed in (0,1,2,7,42,123):
    for n in (8,15,30,60):
        rng = np.random.RandomState(seed)
        add(f"rand_s{seed}_n{n}", rng.rand(n), rng.rand(n), "general")

# gaussian blob
for seed in (3,9):
    rng = np.random.RandomState(seed)
    add(f"blob_s{seed}", rng.randn(40), rng.randn(40), "general")

# ---- cocircular ----
# square in several point orderings
add("sq_ccw",  [0,1,1,0],[0,0,1,1], "cocircular")
add("sq_cw",   [0,0,1,1],[0,1,1,0], "cocircular")
add("sq_diag", [0,1,0,1],[0,0,1,1], "cocircular")
add("sq_shift",[1,1,0,0],[0,1,1,0], "cocircular")

# regular polygons (all vertices cocircular)
for n in (4,5,6,7,8,12):
    ang = np.arange(n)*2*np.pi/n
    add(f"reg{n}", np.cos(ang), np.sin(ang), "cocircular")
    # with center point
    add(f"reg{n}_c", np.r_[np.cos(ang),0.0], np.r_[np.sin(ang),0.0], "cocircular")

# integer grids (many cocircular unit squares)
for nx in (2,3,4,5,6):
    for ny in (2,3,4):
        gx,gy = np.meshgrid(range(nx), range(ny))
        add(f"grid{nx}x{ny}", gx.ravel(), gy.ravel(), "cocircular")

# concentric rings
for (r1,r2,k) in ((1.0,2.0,6),(1.0,2.0,8),(0.5,1.0,5)):
    a1 = np.arange(k)*2*np.pi/k
    a2 = np.arange(k)*2*np.pi/k + np.pi/k
    x = np.r_[r1*np.cos(a1), r2*np.cos(a2)]
    y = np.r_[r1*np.sin(a1), r2*np.sin(a2)]
    add(f"rings_{r1}_{r2}_{k}", x, y, "cocircular")

out = os.path.join(os.path.dirname(os.path.abspath(__file__)), "corpus.json")
with open(out, "w") as f:
    json.dump({"qhull_version": __import__("matplotlib")._qhull.version(),
               "cases": cases}, f, indent=0)
print(f"wrote {len(cases)} cases to {out}")
ncat = {}
for c in cases: ncat[c["category"]] = ncat.get(c["category"],0)+1
print("by category:", ncat)
