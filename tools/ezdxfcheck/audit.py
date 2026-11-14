import sys, glob, os, io, contextlib
import ezdxf
from ezdxf import recover

def describe(doc):
    lines = []
    msp = doc.modelspace()
    from collections import Counter
    lines.append("  modelspace: " + str(dict(Counter(e.dxftype() for e in msp))))
    for layout in doc.layouts:
        if layout.name == "Model":
            continue
        lines.append(f"  layout {layout.name!r} (block {layout.block_record.dxf.name}): " + str(dict(Counter(e.dxftype() for e in layout))))
    for e in msp:
        t = e.dxftype()
        if t == "TEXT":
            lines.append(f"    TEXT {e.dxf.text!r} layer={e.dxf.layer!r}")
        elif t == "MTEXT":
            lines.append(f"    MTEXT {e.plain_text()!r}")
        elif t == "INSERT":
            lines.append(f"    INSERT {e.dxf.name} attribs={[(a.dxf.tag, a.dxf.text) for a in e.attribs]}")
        elif t == "DIMENSION":
            lines.append(f"    DIMENSION type={e.dimtype} block={e.dxf.geometry} measurement={e.get_measurement()!r} blockents={len(e.get_geometry_block()) if e.get_geometry_block() else None}")
        elif t == "HATCH":
            lines.append(f"    HATCH pattern={e.dxf.pattern_name} solid={e.dxf.solid_fill} paths={len(e.paths)}")
        elif t == "LEADER":
            lines.append(f"    LEADER vertices={len(e.vertices)} annotation={e.dxf.get('annotation_handle')}")
        elif t == "LINE":
            lines.append(f"    LINE color={e.dxf.color} ltype={e.dxf.linetype} lw={e.dxf.lineweight} thick={e.dxf.thickness} xdata={e.has_xdata('MYAPP')}")
    lines.append(f"  groups: {[(n, len(g)) for n, g in doc.groups]}")
    return "\n".join(lines)

for fn in sorted(glob.glob(os.path.join(sys.argv[1], "*.dxf"))):
    print("=" * 70)
    print(os.path.basename(fn))
    try:
        doc = ezdxf.readfile(fn)
        print("  readfile: ok", doc.dxfversion, "encoding", doc.encoding)
    except Exception as ex:
        print("  readfile FAILED:", type(ex).__name__, ex)
        try:
            doc, auditor = recover.readfile(fn)
            print("  recover: ok; errors", len(auditor.errors), "fixes", len(auditor.fixes))
        except Exception as ex2:
            print("  recover FAILED:", ex2)
            continue
    auditor = doc.audit()
    print(f"  audit: errors={len(auditor.errors)} fixes={len(auditor.fixes)}")
    for e in auditor.errors[:15]:
        print("    ERROR", e.code, e.message)
    for e in auditor.fixes[:15]:
        print("    FIX", e.code, e.message)
    print(describe(doc))
