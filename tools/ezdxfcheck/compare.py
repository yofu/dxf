import sys, os, ezdxf
from collections import Counter
ez, rt = sys.argv[1], sys.argv[2]

def r(v):
    if isinstance(v, float):
        return round(v, 6)
    if isinstance(v, (tuple, list)):
        return tuple(r(x) for x in v)
    try:
        return tuple(round(x, 6) for x in v)
    except TypeError:
        return v

def ltype_pattern(doc, lt):
    """pattern tags without the optional (74, 0); 340 as the style's (name, font)"""
    out = []
    for t in (lt.pattern_tags.tags if lt.pattern_tags else []):
        v = t.value
        if (t.code == 74 and v == 0) or t.code == 5:  # R12 files have the handle here
            continue
        if t.code == 340:
            st = doc.entitydb.get(v)
            v = (st.dxf.name, st.dxf.font) if st else v
        out.append((t.code, r(v)))
    return tuple(out)

def summary(doc):
    s = {}
    for layout in doc.layouts:
        s["layout " + layout.name] = dict(sorted(Counter(e.dxftype() for e in layout).items()))
    s["layers"] = sorted((l.dxf.name, l.dxf.color, l.dxf.linetype, l.dxf.get("lineweight"), l.dxf.get("plot", 1), l.is_on(),
                          l.dxf.get("true_color"), round(l.transparency, 3)) for l in doc.layers)
    s["version"] = doc.dxfversion
    s["linetypes"] = sorted((lt.dxf.name, ltype_pattern(doc, lt)) for lt in doc.linetypes)
    s["styles"] = sorted((st.dxf.name, st.dxf.font) for st in doc.styles)
    s["dimstyles"] = sorted(ds.dxf.name for ds in doc.dimstyles)
    s["blocks"] = sorted((b.name, len(b)) for b in doc.blocks if not b.name.lower().startswith("*paper_space"))
    s["groups"] = sorted((n, len(g)) for n, g in doc.groups)
    ents = []
    for e in doc.modelspace():
        t = e.dxftype()
        a = e.dxf
        d = [t, a.layer, a.get("color", 256), a.get("linetype", "BYLAYER"), a.get("lineweight", -1),
             a.get("true_color"), a.get("color_name"), a.get("transparency")]
        if t == "LINE": d += [r(a.start), r(a.end), a.get("thickness", 0), e.get_xdata("MYAPP") if e.has_xdata("MYAPP") else None]
        elif t == "CIRCLE": d += [r(a.center), a.radius]
        elif t == "ARC": d += [r(a.center), a.radius, a.start_angle, a.end_angle]
        elif t == "TEXT": d += [a.text, r(a.insert), r(a.get("align_point")), a.halign, a.valign, a.style]
        elif t == "MTEXT": d += [e.text, r(a.insert), a.char_height, a.width, a.get("bg_fill"), a.get("bg_fill_color"), a.get("bg_fill_true_color"), r(a.get("box_fill_scale"))]
        elif t == "LWPOLYLINE": d += [r(list(e.get_points("xyseb"))), e.closed, a.get("elevation", 0)]
        elif t == "POLYLINE": d += [e.get_mode(), len(e.vertices), r(a.elevation), [r(v.dxf.location) for v in e.vertices], [tuple(v.dxf.get(k, 0) for k in ("vtx0", "vtx1", "vtx2", "vtx3")) for v in e.vertices]]
        elif t == "SPLINE": d += [r(list(e.control_points)), r(list(e.weights)), r(list(e.fit_points)), r(list(e.knots)), a.flags]
        elif t == "ELLIPSE": d += [r(a.center), r(a.major_axis), a.ratio, a.start_param, a.end_param]
        elif t == "POINT": d += [r(a.location), a.get("angle", 0)]
        elif t in ("SOLID", "TRACE", "3DFACE"): d += [r(a.vtx0), r(a.vtx1), r(a.vtx2), r(a.vtx3)]
        elif t in ("XLINE", "RAY"): d += [r(a.start), r(a.unit_vector)]
        elif t == "HATCH": d += [a.pattern_name, a.solid_fill, len(e.paths), [type(p).__name__ for p in e.paths], r([(l.angle, *l.offset) for l in e.pattern.lines]) if e.pattern else None,
                                (e.gradient.name, e.gradient.color1, e.gradient.color2, r(e.gradient.rotation), e.gradient.centered, e.gradient.one_color, r(e.gradient.tint)) if e.gradient else None]
        elif t == "DIMENSION": d += [e.dimtype, r(e.get_measurement()), len(e.get_geometry_block()) if e.get_geometry_block() else None, r(a.defpoint)]
        elif t == "LEADER": d += [r(list(e.vertices))]
        elif t == "INSERT": d += [a.name, r(a.insert), a.xscale, a.rotation, [(x.dxf.tag, x.dxf.text) for x in e.attribs]]
        elif t in ("PDFUNDERLAY", "DWFUNDERLAY", "DGNUNDERLAY"):
            ud = e.get_underlay_def()
            d += [r(a.insert), r(e.scaling), a.rotation, a.flags, a.contrast, a.fade, r(list(e.boundary_path)),
                  (ud.dxftype(), ud.dxf.filename, ud.dxf.name) if ud else None]
        elif t in ("IMAGE", "WIPEOUT"):
            d += [r(a.insert), r(a.u_pixel), r(a.v_pixel), r(a.image_size), a.flags, a.clipping, a.fade, r(list(e.boundary_path)),
                  e.image_def.dxf.filename if t == "IMAGE" and e.image_def else None]
        elif t == "WIPEOUT": d += [r(a.insert), len(e.boundary_path)]
        elif t == "MLINE": d += [len(e.vertices), e.style.dxf.name if e.style else None]
        if e.has_extension_dict:
            xd = e.get_extension_dict()
            d += [sorted((k, v.dxftype()) for k, v in xd.items())]
        ents.append(tuple(str(x) for x in d))
    s["entities"] = ents
    vps = []
    for layout in doc.layouts:
        for v in layout.query("VIEWPORT"):
            vps.append((layout.name, r(v.dxf.center), v.dxf.width, v.dxf.height, r(v.dxf.view_center_point), v.dxf.view_height))
    s["viewports"] = sorted(vps)
    return s

for orig in sorted(os.listdir(ez)):
    base = orig[:-4]
    o = ezdxf.readfile(os.path.join(ez, orig))
    so = summary(o)
    for suffix in ("_rt", "_rtbin"):
        fn = os.path.join(rt, base + suffix + ".dxf")
        print("=" * 60)
        print(base + suffix)
        try:
            doc = ezdxf.readfile(fn)
        except Exception as ex:
            print("  READ FAILED:", ex)
            continue
        aud = doc.audit()
        print(f"  audit errors={len(aud.errors)} fixes={len(aud.fixes)}")
        for e in aud.errors[:10]:
            print("    ERROR", e.code, e.message)
        for e in aud.fixes[:10]:
            print("    FIX", e.code, e.message)
        sr = summary(doc)
        for k in so:
            if k == "entities":
                if len(so[k]) != len(sr.get(k, [])):
                    print(f"  entities: {len(so[k])} -> {len(sr.get(k, []))}")
                for a, b in zip(so[k], sr.get(k, [])):
                    if a != b:
                        print("  DIFF", ascii(a))
                        print("    ->", ascii(b))
            elif so[k] != sr.get(k):
                print(f"  DIFF {k}:\n    {ascii(so[k])}\n    -> {ascii(sr.get(k))}")
