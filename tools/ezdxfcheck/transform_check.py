"""Check the transformations of this package against ezdxf.

usage: transform_check.py <dir>   (files written by "ezdxfcheck transform <dir>")

For each transformation, the entities of xf_<name>_orig.dxf are converted
to paths (block references to the paths of their block entities),
transformed by the matrix of the transformation, and compared with the
paths of the entities of xf_<name>.dxf, transformed by this package.
TEXT is compared by its alignment point. Paths of full circles are
cubic Bezier approximations (relative error about 3e-4) starting at
different points: distances below 1e-3 are accepted.
"""
import math
import os
import sys

import ezdxf
from ezdxf import path
from ezdxf.math import Matrix44, Vec3

# the same parameters as in main.go (transforms)
def mirror(p1, p2):
    a = math.atan2(p2[1] - p1[1], p2[0] - p1[0])
    return (Matrix44.translate(-p1[0], -p1[1], 0) @ Matrix44.scale(1, -1, 1)
            @ Matrix44.z_rotate(2 * a) @ Matrix44.translate(p1[0], p1[1], 0))

def affine(base=(0, 0, 0), angle=0, sx=1, sy=1, sz=1, offset=(0, 0, 0)):
    return (Matrix44.translate(-base[0], -base[1], -base[2]) @ Matrix44.scale(sx, sy, sz)
            @ Matrix44.z_rotate(math.radians(angle)) @ Matrix44.translate(base[0] + offset[0], base[1] + offset[1], base[2] + offset[2]))

TRANSFORMS = {
    "mirror": mirror((1, 1), (3, 2)),
    "mirror_scale": affine(base=(2, 0, 0), angle=40, sx=-2, sy=2, sz=2),
    "scalexyz": affine(base=(1, 2, 0), sx=2, sy=0.5),
    "general": affine(angle=30, sx=1.5, sy=-0.7, offset=(5, 5, 0)),
}

def points(entity, m=None):
    """flattened points of the paths of entity (transformed by m)"""
    if entity.dxftype() == "INSERT":
        es = list(entity.virtual_entities())
    else:
        es = [entity]
    pts = []
    for e in es:
        if e.dxftype() in ("ATTDEF", "ATTRIB"):
            continue
        if e.dxftype() == "HATCH":
            ps = path.from_hatch(e)
        else:
            ps = [path.make_path(e)]
        for p in ps:
            if m is not None:
                p = p.transform(m)
            pts.extend(p.flattening(0.0001))
    return pts

def distance(p, polyline):
    best = math.inf
    for a, b in zip(polyline, polyline[1:]):
        ab = b - a
        t = 0 if ab.magnitude == 0 else max(0, min(1, (p - a).dot(ab) / ab.magnitude_square))
        best = min(best, (p - (a + ab * t)).magnitude)
    return best

def compare(ref, got):
    """largest distance between the point sets (both ways)"""
    if not ref or not got:
        return 0 if not ref and not got else math.inf
    if len(ref) == 1 or len(got) == 1:
        return max(min((p - q).magnitude for q in ref) for p in got)
    return max(max(distance(p, ref) for p in got), max(distance(p, got) for p in ref))

def main(d):
    failed = 0
    for name, m in TRANSFORMS.items():
        orig = ezdxf.readfile(os.path.join(d, f"xf_{name}_orig.dxf"))
        res = ezdxf.readfile(os.path.join(d, f"xf_{name}.dxf"))
        audit = res.audit()
        print(f"{name}: audit errors={len(audit.errors)} fixes={len(audit.fixes)}")
        for a, b in zip(orig.modelspace(), res.modelspace()):
            if a.dxftype() == "TEXT":
                p = m.transform(a.dxf.insert)
                q = b.dxf.align_point if b.dxf.get("halign", 0) or b.dxf.get("valign", 0) else b.dxf.insert
                err = (p - Vec3(q)).magnitude
            else:
                err = compare(points(a, m), points(b))
            status = "ok" if err < 1e-3 else "NG"
            if status == "NG":
                failed += 1
            print(f"  {status} {a.dxftype():10s} -> {b.dxftype():10s} max distance {err:.2e}")
    print("failed:", failed)
    return failed

if __name__ == "__main__":
    sys.exit(1 if main(sys.argv[1]) else 0)
