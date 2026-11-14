"""Compare the paths and extents of this package with ezdxf.

usage: path_check.py <dir>   (files written by "ezdxfcheck paths <dir>")

For each entity of paths.dxf, the points flattened by this package
(paths.json) are compared with ezdxf's path.make_path flattening (largest
distance between the polylines, both ways), and the extents with ezdxf's
bbox.extents.
"""
import json
import math
import os
import sys

import ezdxf
from ezdxf import bbox, path
from ezdxf.math import Vec3


def distance(p, polyline):
    best = math.inf
    for a, b in zip(polyline, polyline[1:]):
        ab = b - a
        t = 0 if ab.magnitude == 0 else max(0, min(1, (p - a).dot(ab) / ab.magnitude_square))
        best = min(best, (p - (a + ab * t)).magnitude)
    return best


def compare(ref, got):
    if not ref or not got:
        return 0 if not ref and not got else math.inf
    if len(ref) == 1 or len(got) == 1:
        return max(min((p - q).magnitude for q in ref) for p in got)
    return max(max(distance(p, ref) for p in got), max(distance(p, got) for p in ref))


def ezpaths(e):
    if e.dxftype() == "HATCH":
        return list(path.from_hatch(e))
    if e.dxftype() == "POLYLINE" and (e.is_polygon_mesh or e.is_poly_face_mesh):
        return []
    return [path.make_path(e)]


def main(d):
    doc = ezdxf.readfile(os.path.join(d, "paths.dxf"))
    recs = json.load(open(os.path.join(d, "paths.json")))
    failed = 0
    for e, rec in zip(doc.modelspace(), recs):
        if e.dxftype() in ("TEXT", "MTEXT"):
            continue  # no paths without fonts
        if e.dxftype() == "INSERT":
            ps = [path.make_path(v) for v in e.virtual_entities()]
        else:
            ps = ezpaths(e)
        ref = [Vec3(q) for p in ps for q in p.flattening(0.0001)]
        got = [Vec3(q) for poly in (rec["points"] or []) for q in poly]
        err = compare(ref, got) if ps and rec["points"] else 0
        box = bbox.extents([e], fast=False)
        berr = max((box.extmin - Vec3(rec["min"])).magnitude, (box.extmax - Vec3(rec["max"])).magnitude) if box.has_data else 0
        ok = err < 1e-3 and berr < 1e-3
        failed += not ok
        print(f"  {'ok' if ok else 'NG'} {e.dxftype():10s} path {err:.2e} extents {berr:.2e}")
    print("failed:", failed)
    return failed


if __name__ == "__main__":
    sys.exit(1 if main(sys.argv[1]) else 0)
