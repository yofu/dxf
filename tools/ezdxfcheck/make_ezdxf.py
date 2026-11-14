import sys, os, ezdxf
from ezdxf.math import Vec3
out = sys.argv[1]

doc = ezdxf.new("R2000", setup=True)
doc.layers.add("WALL", color=1, linetype="DASHED", lineweight=35)
nl = doc.layers.add("NOPLOT", color=3); nl.dxf.plot = 0; nl.off()
# complex linetypes: text and shape
doc.styles.add_shx("ltypeshp.shx")
doc.linetypes.add("GAS_LINE", pattern='A,.5,-.2,["GAS",STANDARD,S=.1,U=0.0,X=-0.1,Y=-.05],-.25',
                  description="Gas line ----GAS----GAS----", length=0.95)
doc.linetypes.add("TRACKS", pattern="A,.15,[132,ltypeshp.shx,S=.25],.15",
                  description="Tracks -|-|-|-|-|-|-|-", length=0.3)
msp = doc.modelspace()
msp.add_line((0, -5), (10, -5), dxfattribs={"linetype": "GAS_LINE"})
msp.add_line((0, 0), (10, 5), dxfattribs={"layer": "WALL", "color": 3, "thickness": 2})
msp.add_circle((5, 5), 3, dxfattribs={"linetype": "CENTER", "lineweight": 50})
msp.add_arc((0, 0), 5, 30, 150)
msp.add_text("Hello", dxfattribs={"height": 2.5, "style": "OpenSans"}).set_placement((1, 2), align=ezdxf.enums.TextEntityAlignment.MIDDLE_CENTER)
msp.add_mtext(r"line1\Pline2 {\fArial|b1;bold}", dxfattribs={"char_height": 2, "insert": (0, 20), "width": 30})
msp.add_lwpolyline([(0, 0, 0.1, 0.2, 0.5), (10, 0, 0, 0, 0), (10, 10)], format="xyseb", close=True, dxfattribs={"elevation": 3})
msp.add_polyline2d([(0, 0), (5, 5), (10, 0)], dxfattribs={"elevation": (0, 0, 1)})
msp.add_polyline3d([(0, 0, 0), (1, 2, 3), (4, 5, 6)])
pf = msp.add_polyface()
pf.append_face([(0, 0, 0), (1, 0, 0), (1, 1, 0), (0, 1, 0)])
pf.append_face([(0, 0, 1), (1, 0, 1), (1, 1, 1)])
msp.add_rational_spline([(0, 0), (1, 2), (3, 1), (4, 0)], [1, 0.5, 2, 1], degree=3)
msp.add_spline(fit_points=[(0, 10), (2, 12), (4, 10), (6, 12)])
msp.add_ellipse((20, 20), major_axis=(5, 1), ratio=0.4, start_param=0.5, end_param=4)
msp.add_point((3, 3), dxfattribs={"angle": 15})
msp.add_solid([(0, 0), (1, 0), (0, 1), (1, 1)])
msp.add_trace([(0, 0), (1, 0), (0, 1), (1, 1)])
msp.add_3dface([(0, 0, 0), (1, 0, 0), (1, 1, 1), (0, 1, 1)])
msp.add_xline((0, 0), (1, 1))
msp.add_ray((0, 0), (1, 0))
h = msp.add_hatch(color=2)
h.paths.add_polyline_path([(0, 0), (10, 0), (10, 10), (0, 10)], is_closed=True)
h.set_pattern_fill("ANSI37", scale=0.5, angle=15)
h2 = msp.add_hatch()
ep = h2.paths.add_edge_path()
ep.add_line((0, 0), (10, 0))
ep.add_arc((10, 5), 5, -90, 90)
ep.add_ellipse((5, 10), (5, 0), 0.5, 0, 180)
ep.add_line((0, 10), (0, 0))
dim = msp.add_linear_dim(base=(5, -5), p1=(0, 0), p2=(10, 0))
dim.render()
dim2 = msp.add_radius_dim((50, 50), radius=10, angle=45)
dim2.render()
dim3 = msp.add_aligned_dim(p1=(0, 0), p2=(10, 10), distance=3)
dim3.render()
msp.add_leader([(0, 0), (5, 5), (10, 5)])
blk = doc.blocks.new("DOOR")
blk.add_line((0, 0), (1, 0))
blk.add_arc((0, 0), 1, 0, 90)
blk.add_attdef("TAG", (0, 1), dxfattribs={"height": 0.25, "prompt": "Tag?"})
ins = msp.add_blockref("DOOR", (20, 0), dxfattribs={"xscale": 2, "rotation": 30})
ins.add_auto_attribs({"TAG": "D-01"})
line = msp.add_line((0, 0), (0, 10))
line.set_xdata("MYAPP", [(1000, "text"), (1040, 3.14), (1070, 7)])
doc.groups.new("GRP").extend([line])
psp = doc.layouts.get("Layout1")
psp.add_viewport(center=(100, 100), size=(150, 100), view_center_point=(5, 5), view_height=50)
lay2 = doc.layouts.new("Second")
lay2.add_line((0, 0), (100, 100))
lay2.add_viewport(center=(100, 100), size=(100, 100), view_center_point=(0, 0), view_height=30)
# objects referred to by handle: image, wipeout, mline, extension dictionary
image_def = doc.add_image_def(filename="plan.png", size_in_pixel=(640, 480))
msp.add_image(insert=(0, 0), size_in_units=(6.4, 4.8), image_def=image_def, rotation=0)
clipped = msp.add_image(insert=(10, 20), size_in_units=(64, 48), image_def=image_def, rotation=90)
clipped.set_boundary_path([(-0.5, -0.5), (319.5, 239.5)])
clipped.dxf.fade = 30
msp.add_wipeout([(0, 0), (10, 0), (10, 5), (0, 5)])
pdf_def = doc.add_underlay_def(filename="plans/site.pdf", fmt="pdf", name="2")
pu = msp.add_underlay(pdf_def, insert=(10, 20), scale=2, rotation=30)
pu.clipping = True
pu.boundary_path = [(0, 0), (100, 50)]
pu.dxf.fade = 25
dgn_def = doc.add_underlay_def(filename="model.dgn", fmt="dgn", name="default")
msp.add_underlay(dgn_def, insert=(0, 0))
msp.add_wipeout([(0, 0), (5, 0), (5, 5)])
msp.add_mline([(0, 0), (10, 0), (10, 10)])
xl = msp.add_line((0, 0), (1, 1))
xdict = xl.new_extension_dict()
xrec = xdict.add_xrecord("MYDATA")
xrec.extend([(1, "extension data"), (40, 1.5)])
doc.saveas(os.path.join(out, "ez_r2000.dxf"))
doc.saveas(os.path.join(out, "ez_r2000_bin.dxf"), fmt="bin")

# R2010: true color, color names and transparency
c10 = ezdxf.new("R2010")
lay = c10.layers.add("GLASS")
lay.rgb = (10, 200, 250)
lay.transparency = 0.3
m10 = c10.modelspace()
ln = m10.add_line((0, 0), (10, 0), dxfattribs={"layer": "GLASS"})
ln.rgb = (255, 128, 0)
ln.transparency = 0.5
m10.add_circle((0, 0), 5, dxfattribs={"true_color": ezdxf.colors.rgb2int((1, 2, 3)), "color_name": "BOOK$NAME"})
# gradient hatch and MTEXT background fill
gh = m10.add_hatch()
gh.paths.add_polyline_path([(0, 0), (10, 0), (10, 5), (0, 5)])
gh.set_gradient((255, 0, 0), (0, 0, 255), rotation=30, name="CYLINDER")
gh1 = m10.add_hatch()
gh1.paths.add_polyline_path([(20, 0), (30, 0), (30, 5)])
gh1.set_gradient((0, 128, 0), (191, 191, 191), rotation=0, centered=1, one_color=1, tint=0.75, name="SPHERICAL")
m10.add_mtext("bg color").set_bg_color(2, scale=1.2)
m10.add_mtext("bg rgb").set_bg_color((10, 20, 30), scale=2, text_frame=True)
m10.add_mtext("bg canvas").set_bg_color("canvas")
m10.add_mtext("frame").set_bg_color(None, text_frame=True)
c10.saveas(os.path.join(out, "ez_r2010.dxf"))

# other versions: Japanese text (UTF-8 for R2007 and later), true color, image, mline
for v in ["R2004", "R2007", "R2013", "R2018"]:
    vd = ezdxf.new(v, setup=True)
    vd.layers.add("壁", color=1)
    vm = vd.modelspace()
    vm.add_text("日本語", dxfattribs={"layer": "壁", "height": 2.5})
    vl = vm.add_line((0, 0), (10, 5))
    vl.rgb = (12, 34, 56)
    if v != "R2004":
        vl.transparency = 0.4
    vdef = vd.add_image_def(filename="plan.png", size_in_pixel=(640, 480))
    vm.add_image(insert=(0, 0), size_in_units=(6.4, 4.8), image_def=vdef, rotation=0)
    vm.add_mline([(0, 0), (10, 0)])
    vm.add_mtext(r"複数行\P文字", dxfattribs={"char_height": 2})
    vd.saveas(os.path.join(out, f"ez_{v.lower()}.dxf"))

# Japanese drawing in Shift_JIS (cp932)
jdoc = ezdxf.new("R2000")
jdoc.encoding = "cp932"
jdoc.header["$DWGCODEPAGE"] = "ANSI_932"
jdoc.layers.add("壁")
jdoc.modelspace().add_text("日本語", dxfattribs={"layer": "壁", "height": 3})
jdoc.saveas(os.path.join(out, "ez_sjis.dxf"))

# R12 drawing
r12 = ezdxf.new("R12")
m = r12.modelspace()
m.add_line((0, 0), (1, 1))
m.add_polyline2d([(0, 0), (1, 0), (1, 1)], close=True)
m.add_text("R12")
b = r12.blocks.new("B")
b.add_circle((0, 0), 1)
m.add_blockref("B", (5, 5))
r12.saveas(os.path.join(out, "ez_r12.dxf"))
print("ok")

# phase A entities: XLINE, RAY, TRACE, SHAPE, TOLERANCE, LIGHT, HELIX, MESH,
# MPOLYGON, 3DSOLID (SAT in R2010, SAB in ACDSDATA for R2018), ARC_DIMENSION
from ezdxf.render import forms
from ezdxf.acis import api as acis
for v in ["R2010", "R2018"]:
    md = ezdxf.new(v, setup=True)
    mm = md.modelspace()
    mm.add_xline((1, 2), (3, 4))
    mm.add_ray((0, 0), (0, 2))
    mm.add_trace([(0, 0), (1, 0), (0, 1), (1, 1)])
    md.styles.add_shx("ltypeshp.shx")
    mm.add_shape("TRACK1", insert=(5, 5), size=2, dxfattribs={"rotation": 30})
    tol = mm.new_entity("TOLERANCE", {"insert": (0, 10), "content": r"{\Fgdt;j}%%v0.1", "dimstyle": "Standard"})
    light = mm.new_entity("LIGHT", {"name": "L1", "type": 3, "location": (0, 0, 10), "target": (0, 0, 9)})
    hx = mm.add_helix(2, 1, 3)
    hx.translate(10, 0, 0)
    mesh = mm.add_mesh()
    with mesh.edit_data() as data:
        data.vertices = [(0, 0, 0), (1, 0, 0), (1, 1, 0), (0, 1, 1)]
        data.faces = [(0, 1, 2), (0, 2, 3)]
    mp = mm.add_mpolygon(color=3)
    mp.paths.add_polyline_path([(0, 0), (10, 0), (10, 10), (0, 10)], flags=1)
    mp.paths.add_polyline_path([(2, 2), (4, 2), (4, 4)], flags=0)
    solid = mm.add_3dsolid()
    body = acis.body_from_mesh(forms.cube())
    if v < "R2013":
        solid.sat = acis.export_sat([body])
    else:
        solid.sab = acis.export_sab([body])
    dim = mm.add_arc_dim_3p(base=(0, 6), center=(0, 0), p1=(5, 0), p2=(0, 5))
    dim.render()
    md.saveas(os.path.join(out, f"ez_more_{v}.dxf"))

# phase B: MLINE, MULTILEADER (MTEXT and block content)
from ezdxf.math import Vec2
from ezdxf.render import mleader as mlr
for v in ["R2010", "R2018"]:
    bd = ezdxf.new(v, setup=True)
    bm = bd.modelspace()
    bm.add_mline([(0, 0), (10, 0), (10, 10)], dxfattribs={"scale_factor": 2, "justification": 1})
    st = bd.mline_styles.new("WALL")
    st.elements.append(0.1, 1)
    st.elements.append(-0.1, 5)
    bm.add_mline([(0, 0), (4, 0), (4, 4), (0, 4)], close=True, dxfattribs={"style_name": "WALL"})
    ml = bm.add_multileader_mtext("Standard")
    ml.set_content(r"Note\Psecond line", char_height=2.5)
    ml.add_leader_line(mlr.ConnectionSide.left, [Vec2(0, 0), Vec2(5, 8)])
    ml.add_leader_line(mlr.ConnectionSide.right, [Vec2(30, 0)])
    ml.build(insert=Vec2(10, 10))
    blk = bd.blocks.new("MARK")
    blk.add_circle((0, 0), 1)
    mb = bm.add_multileader_block("Standard")
    mb.set_content(name="MARK", scale=2)
    mb.add_leader_line(mlr.ConnectionSide.left, [Vec2(-10, -10)])
    mb.build(insert=Vec2(20, 20))
    bd.saveas(os.path.join(out, f"ez_phaseb_{v}.dxf"))
