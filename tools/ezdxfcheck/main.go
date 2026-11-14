// Command ezdxfcheck creates DXF files for validation with ezdxf
// (see README.md); generate also writes each sample as R12 (SaveAsR12):
//
//	go run ./tools/ezdxfcheck generate <dir>        sample files written by this package
//	go run ./tools/ezdxfcheck roundtrip <in> <out>  read files and write them back
//	go run ./tools/ezdxfcheck transform <dir>       entities before and after transformations
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/yofu/dxf"
	"github.com/yofu/dxf/codepage"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/object"
	"github.com/yofu/dxf/path"
	"github.com/yofu/dxf/table"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func basic() *drawing.Drawing {
	d := dxf.NewDrawing()
	must(d.AddLayer("壁", color.Red, dxf.DefaultLineType, true))
	hidden := must(d.LineType("HIDDEN"))
	l := must(d.Line(0, 0, 0, 100, 50, 0))
	l.SetLineType(hidden)
	l.SetLineWeight(35)
	l.SetColor(color.Blue)
	l.Thickness = 5
	must(d.Circle(10, 10, 0, 5))
	must(d.Arc(20, 20, 0, 5, 0, 90))
	must(d.Point(1, 2, 3))
	t := must(d.Text("日本語テキスト", 0, -10, 0, 2.5))
	t.Anchor(entity.CENTER_CENTER)
	must(d.MText("1行目\n2行目", 0, -20, 0, 2.5, 50))
	lw := must(d.LwPolyline(true, []float64{0, 0}, []float64{10, 0}, []float64{10, 10}))
	lw.Bulges[0] = 0.5
	lw.SetWidth(1, 0.2, 0.4)
	lw.Elevation = 2
	must(d.Polyline(false, []float64{0, 0, 0}, []float64{1, 1, 1}, []float64{2, 0, 2}))
	must(d.ThreeDFace([][]float64{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}}))
	must(d.Spline([][]float64{{0, 0, 0}, {1, 1, 0}, {2, 0, 0}, {3, 1, 0}, {4, 0, 0}}))
	e := must(d.Ellipse([]float64{50, 50, 0}, []float64{10, 0, 0}, 0.5))
	e.Start, e.End = 0, math.Pi
	s := entity.NewSolid()
	s.Points = [][]float64{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}, {0, 1, 0}}
	d.AddEntity(s)
	g := must(d.Group("G1", "group", l))
	_ = g
	l.SetXData([][2]string{{"1001", "MYAPP"}, {"1000", "line data"}, {"1040", "1.5"}})
	return d
}

func blocks() *drawing.Drawing {
	d := dxf.NewDrawing()
	must(d.BeginBlock("TITLE", 0, 0, 0))
	must(d.Line(0, 0, 0, 100, 0, 0))
	must(d.Circle(50, 10, 0, 5))
	must(d.AttributeDefinition("NAME", "図面名", "未定", 10, 5, 0, 3))
	must(d.AttributeDefinition("DATE", "日付", "2026-01-01", 60, 5, 0, 2.5))
	d.EndBlock()
	ins := must(d.Insert("TITLE", 100, 100, 0))
	ins.SetAttribute("NAME", "平面図")
	ins2 := must(d.Insert("TITLE", 0, 0, 0))
	ins2.Scale = []float64{2, 2, 1}
	ins2.Rotation = 45
	d.SyncAttributes(ins2)
	ins3 := must(d.Insert("TITLE", 0, 300, 0))
	ins3.Columns, ins3.ColumnSpacing = 3, 150
	return d
}

// more uses XLINE, RAY, TRACE, SHAPE, TOLERANCE, LIGHT, HELIX, MESH,
// MPOLYGON, 3DSOLID (SAT) and ARC_DIMENSION (AC1024).
func more() *drawing.Drawing {
	d := dxf.NewDrawing()
	must(0, d.SetVersion(format.R2010))
	must(d.XLine([]float64{1, 2, 0}, []float64{3, 4, 0}))
	must(d.Ray([]float64{0, 0, 0}, []float64{0, 2, 0}))
	must(d.Trace([][]float64{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}, {1, 1, 0}}))
	d.AddShapeFile("ltypeshp.shx")
	sh := must(d.Shape("TRACK1", []float64{5, 5, 0}, 2))
	sh.Rotation = 30
	must(d.Tolerance(`{\Fgdt;j}%%v0.1`, []float64{0, 10, 0}))
	must(d.Light("L1", entity.LIGHT_SPOT, []float64{0, 0, 10}))
	must(d.Helix([]float64{10, 0, 0}, 2, 1, 3, true))
	m := must(d.Mesh([][]float64{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 1}}, [][]int{{0, 1, 2}, {0, 2, 3}}))
	m.Edges = [][2]int{{0, 1}}
	m.Creases = []float64{0.5}
	mp := must(d.MPolygon([][]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}, [][]float64{{2, 2}, {4, 2}, {4, 4}}))
	mp.FillColor = 3
	must(d.Body("3DSOLID", cubeSAT))
	return d
}

// phaseB uses MLINE (Standard and a style), MULTILEADER with MTEXT
// content (AC1027).
func phaseB() *drawing.Drawing {
	d := dxf.NewDrawing()
	must(0, d.SetVersion(format.R2013))
	must(d.MLine("", [][]float64{{0, 0}, {10, 0}, {10, 10}}, false, 2, entity.MLINE_ZERO))
	w := must(d.AddMLineStyle("WALL", entity.MLineStyleElement{Offset: 0.1, Color: 1, LineType: "BYLAYER"},
		entity.MLineStyleElement{Offset: -0.1, Color: 5, LineType: "BYLAYER"}))
	w.Flags = entity.MLINESTYLE_FILL
	must(d.MLine("WALL", [][]float64{{0, 0}, {4, 0}, {4, 4}, {0, 4}}, true, 1, entity.MLINE_TOP))
	must(d.MultiLeaderText(`Note\Psecond line`, []float64{10, 10, 0}, 2.5,
		[][]float64{{0, 0, 0}, {5, 8, 0}}, [][]float64{{30, 0, 0}}))
	return d
}

// phaseC uses XRECORD (entity and layer), DICTIONARYVAR, draw order,
// XCLIP, a page setup and image/wipeout variables.
func phaseC() *drawing.Drawing {
	d := dxf.NewDrawing()
	l := must(d.Line(0, 0, 0, 1, 0, 0))
	must(d.SetXRecord(l, "MYAPP", [][2]string{{"1", "data"}, {"40", "1.5"}, {"90", "7"}}))
	must(d.SetXRecord(d.Layers["0"], "LAYERDATA", [][2]string{{"1", "on layer"}}))
	d.SetVariable("XCLIPFRAME", "2")
	c := must(d.Circle(0, 0, 0, 1))
	must(d.Point(0, 0, 0))
	must(0, d.SetDrawOrder("", l))
	_ = c
	d.BeginBlock("B", 0, 0, 0)
	must(d.Line(0, 0, 0, 10, 10, 0))
	d.EndBlock()
	ins := must(d.Insert("B", 100, 0, 0))
	ins.Rotation = 90
	must(d.ClipBlockReference(ins, [][]float64{{99, 1}, {101, 5}}))
	d.Layouts()[1].PageSetup(420, 297, object.PLOT_MILLIMETERS, [4]float64{10, 10, 10, 20}, 1, 50, "")
	d.SetRasterVariables(1, 1, 1)
	d.SetWipeoutVariables(2)
	return d
}

// phaseD uses the creation APIs: dimensions, MTEXT columns, patterns,
// associative hatch, polygon mesh, splines, arrows, linetypes, XDATA and
// a revision cloud (AC1032).
// phaseF: comments, entities imported from another drawing and an
// external reference
func phaseF() *drawing.Drawing {
	src := phaseD()
	d := dxf.NewDrawing()
	d.Comments = []string{"phase F sample"}
	must(dxf.ImportEntities(d, src, src.Entities()...))
	must(dxf.AttachXref(d, "XPLAN", "basic.dxf", []float64{100, 0, 0}, false))
	return d
}

func phaseD() *drawing.Drawing {
	d := dxf.NewDrawing()
	must(0, d.SetVersion(format.R2018))
	must(d.Angular2LineDimension([]float64{0, 0}, []float64{10, 0}, []float64{0, 0}, []float64{10, 10}, []float64{8, 3}))
	must(d.OrdinateDimension([]float64{0, 0}, []float64{30, 20}, []float64{35, 40}, true))
	must(d.OrdinateDimension([]float64{0, 0}, []float64{30, 20}, []float64{50, 25}, false))
	must(d.ArcDimension([]float64{0, 0}, []float64{5, 0}, []float64{0, 5}, []float64{4, 4}))
	must(d.LinearDimensions([][]float64{{0, -10}, {10, -10}, {25, -10}}, []float64{0, -15}, 0))
	ed := entity.NewMTextEditor("Title").Paragraph().Font("Arial", true, false, "bold").ColumnBreak().Append("second column")
	m := must(d.MText(ed.String(), 0, 30, 0, 1, 20))
	m.Columns = entity.NewStaticColumns(2, 20, 2, 30)
	h := must(d.Hatch([]float64{50, 0}, []float64{60, 0}, []float64{60, 10}))
	must(0, h.SetPattern("BRICK", 0, 1))
	lw := must(d.LwPolyline(true, []float64{70, 0}, []float64{90, 0}, []float64{90, 20}, []float64{70, 20}))
	c := must(d.Circle(80, 10, 0, 3))
	must(d.HatchFromEntities(lw, c))
	must(d.PolygonMesh(2, 3, [][]float64{{0, 50, 0}, {1, 50, 0}, {2, 50, 1}, {0, 51, 0}, {1, 51, 1}, {2, 51, 0}}, false, true))
	must(d.SplineThrough([][]float64{{0, 60, 0}, {2, 63, 0}, {5, 61, 0}, {8, 64, 0}, {10, 60, 0}}))
	must(d.SplineFit([][]float64{{0, 70, 0}, {2, 73, 0}, {5, 71, 0}}))
	must(d.InsertArrow("_DOT", []float64{100, 0}, 2.5, 45))
	must(d.LoadLineTypes("*GAS2,Gas ----GAS----\nA,.5,-.2,[\"GAS\",STANDARD,S=.1,R=0.0,X=-0.1,Y=-.05],-.25\n"))
	must(0, dxf.SetHeaderVariable(d, "$MIRRTEXT", [2]string{"70", "1"}))
	l := must(d.Line(0, 80, 0, 10, 80, 0))
	l.SetXDataValues("MYAPP", entity.XString("text"), entity.XReal(1.5), entity.XPoint([]float64{1, 2, 3}))
	must(d.RevisionCloud([][]float64{{0, 90}, {10, 90}, {10, 95}, {0, 95}}, 2, true))
	return d
}

// pathSamples writes entities of all kinds with paths and the flattened
// points and extents of this package (paths.json) for path_check.py.
func pathSamples(out string) {
	d := transformEntities(true)
	sp := must(d.Spline([][]float64{{0, 10, 0}, {1, 13, 0}, {3, 14, 0}, {5, 11, 0}, {7, 13, 0}, {9, 10, 0}}))
	sp.Knots = []float64{0, 0, 0, 0, 1, 2, 3, 3, 3, 3}
	h := must(d.Hatch([]float64{20, 20}, []float64{30, 20}, []float64{30, 25}))
	h.Boundaries[0].Bulges = []float64{0.5, 0, 0}
	must(d.PolygonMesh(2, 3, [][]float64{{0, 50, 0}, {1, 50, 0}, {2, 50, 1}, {0, 51, 0}, {1, 51, 1}, {2, 51, 0}}, false, false))
	must(d.Helix([]float64{40, 0, 0}, 2, 1, 2, true))
	must(0, d.SaveAs(filepath.Join(out, "paths.dxf")))
	type rec struct {
		Type   string        `json:"type"`
		Points [][][]float64 `json:"points"`
		Min    []float64     `json:"min"`
		Max    []float64     `json:"max"`
	}
	var recs []rec
	for _, e := range d.Entities() {
		r := rec{Type: dxf.TypeName(e)}
		if ps, err := path.FromEntity(e); err == nil {
			for _, p := range ps {
				r.Points = append(r.Points, p.Flatten(0.0001))
			}
		}
		r.Min, r.Max = dxf.Extents(d, entity.Entities{e}, 0.0001)
		recs = append(recs, r)
	}
	b, _ := json.Marshal(recs)
	must(0, os.WriteFile(filepath.Join(out, "paths.json"), b, 0644))
	fmt.Println("written in", out)
}

// cubeSAT is minimal SAT data (the geometry is not checked).
var cubeSAT = []string{"700 0 1 0", "@33 Open Design Alliance ACIS Builder @12 ACIS 32.0 NT @24 Thu Jan 01 00:00:00 2026 ", "1 9.9999999999999995e-07 1e-10 ", "End-of-ACIS-data"}

// images uses an image, a clipped rotated image and a wipeout.
func images() *drawing.Drawing {
	d := dxf.NewDrawing()
	def := must(d.AddImageDefinition("images/plan.png", 640, 480))
	must(d.Image(def, 0, 0, 0, 6.4, 4.8, 0))
	im := must(d.Image(def, 10, 20, 0, 64, 48, 90))
	im.SetBoundary([]float64{-0.5, -0.5}, []float64{319.5, 239.5})
	im.Fade = 30
	must(d.Wipeout([]float64{0, 0}, []float64{10, 0}, []float64{10, 5}, []float64{0, 5}))
	return d
}

// transformations checked by transform_check.py (the same parameters)
var transforms = []struct {
	name    string
	t       entity.Transform
	uniform bool // conformal: all entities
}{
	{"mirror", entity.MirrorBy([]float64{1, 1}, []float64{3, 2}), true},
	{"mirror_scale", entity.Transform{Base: []float64{2, 0, 0}, Angle: 40, ScaleX: -2, ScaleY: 2, ScaleZ: 2}, true},
	{"scalexyz", entity.ScaleXYZ([]float64{1, 2, 0}, 2, 0.5, 1), false},
	{"general", entity.Transform{Angle: 30, ScaleX: 1.5, ScaleY: -0.7, Offset: []float64{5, 5, 0}}, false},
}

// transformEntities returns a drawing with entities to transform; uniform
// adds the entities that need a conformal transformation.
func transformEntities(uniform bool) *drawing.Drawing {
	d := dxf.NewDrawing()
	d.BeginBlock("B", 0, 0, 0)
	must(d.Line(0, 0, 0, 2, 0, 0))
	must(d.Arc(1, 0, 0, 1, 0, 90))
	d.EndBlock()
	must(d.Line(0, 0, 0, 3, 1, 0))
	must(d.Circle(2, 2, 0, 1.5))
	must(d.Arc(-1, 1, 0, 2, 30, 200))
	e := must(d.Ellipse([]float64{4, -2, 0}, []float64{2, 1, 0}, 0.5))
	e.Start, e.End = 0.5, 4
	must(d.LwPolyline(true, []float64{0, 0}, []float64{4, 0}, []float64{4, 3}))
	must(d.Spline([][]float64{{0, 0, 0}, {1, 2, 0}, {3, 1, 0}, {4, 3, 0}}))
	must(d.ThreeDFace([][]float64{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}}))
	h := must(d.Hatch([]float64{5, 5}, []float64{8, 5}, []float64{8, 7}))
	h.SetPattern("ANSI31", 0, 1)
	ins := must(d.Insert("B", 2, -3, 0))
	ins.Rotation = 90
	// extrusion -Z (OCS X axis is -X)
	cz := must(d.Circle(-3, 4, 0, 1.2))
	cz.Direction = []float64{0, 0, -1}
	az := must(d.Arc(2, 6, 0, 1.5, 20, 160))
	az.Direction = []float64{0, 0, -1}
	lz := must(d.LwPolyline(false, []float64{-6, 0}, []float64{-4, 0}, []float64{-4, 2}))
	lz.Extrusion = []float64{0, 0, -1}
	if uniform {
		lb := must(d.LwPolyline(false, []float64{-6, 3}, []float64{-4, 3}, []float64{-4, 5}))
		lb.Bulges[0] = 0.4
		lb.Extrusion = []float64{0, 0, -1}
		// tilted extrusion
		ct := must(d.Circle(1, 2, 3, 1))
		ct.Direction = []float64{0.6, 0, 0.8}
		at := must(d.Arc(1, 2, 3, 2, 0, 120))
		at.Direction = []float64{0, -0.6, 0.8}
		lp := must(d.LwPolyline(false, []float64{0, 5}, []float64{3, 5}, []float64{3, 8}))
		lp.Bulges[0], lp.Bulges[1] = 0.5, -0.3
		ha := entity.NewHatch()
		ha.AddEdgePath(
			&entity.HatchLine{Start: []float64{10, 0}, End: []float64{12, 0}},
			&entity.HatchArc{Center: []float64{12, 1}, Radius: 1, Start: 270, End: 90, CCW: true},
			&entity.HatchArc{Center: []float64{11, 2}, Radius: 1, Start: 0, End: 180, CCW: false},
			&entity.HatchLine{Start: []float64{10, 2}, End: []float64{10, 0}},
		)
		d.AddEntity(ha)
		ins2 := must(d.Insert("B", -3, -3, 0))
		ins2.Rotation = 25
		ins2.Scale = []float64{1, 2, 1}
		must(d.Text("ABC", 1, 1, 0, 1))
	}
	return d
}

func transformSamples(out string) {
	for _, tr := range transforms {
		d := transformEntities(tr.uniform)
		if err := d.SaveAs(filepath.Join(out, "xf_"+tr.name+"_orig.dxf")); err != nil {
			panic(err)
		}
		d = transformEntities(tr.uniform)
		if err := d.TransformEntities(tr.t, d.Entities()...); err != nil {
			panic(fmt.Sprintf("%s: %v", tr.name, err))
		}
		if err := d.SaveAs(filepath.Join(out, "xf_"+tr.name+".dxf")); err != nil {
			panic(err)
		}
	}
	fmt.Println("written in", out)
}

// linetypes uses complex linetypes with a text string and a shape.
func linetypes() *drawing.Drawing {
	d := dxf.NewDrawing()
	shx := d.AddShapeFile("ltypeshp.shx")
	gas := must(d.AddComplexLineType("GAS_LINE", "Gas line ----GAS----GAS----",
		&table.LineTypeElement{Length: 0.5},
		table.TextElement(-0.2, "GAS", d.CurrentStyle, 0.1, 0, -0.1, -0.05),
		&table.LineTypeElement{Length: -0.25}))
	tracks := must(d.AddComplexLineType("TRACKS", "Tracks -|-|-|-|-|-|-|-",
		table.ShapeElement(0.15, 132, shx, 0.25, 0, 0, 0),
		&table.LineTypeElement{Length: 0.15}))
	l := must(d.Line(0, 0, 0, 10, 0, 0))
	l.SetLineType(gas)
	c := must(d.Circle(0, 0, 0, 5))
	c.SetLineType(tracks)
	return d
}

// underlays uses PDF and DGN underlays.
func underlays() *drawing.Drawing {
	d := dxf.NewDrawing()
	def := must(d.AddUnderlayDefinition(entity.UNDERLAY_PDF, "plans/site.pdf", "2"))
	u := must(d.Underlay(def, 10, 20, 0, 2))
	u.Rotation = 30
	u.SetClipping([]float64{0, 0}, []float64{100, 50})
	u.Fade = 25
	dgn := must(d.AddUnderlayDefinition(entity.UNDERLAY_DGN, "model.dgn", ""))
	must(d.Underlay(dgn, 0, 0, 0, 1))
	return d
}

func hatchDims() *drawing.Drawing {
	d := dxf.NewDrawing()
	h := must(d.Hatch([]float64{0, 0}, []float64{100, 0}, []float64{100, 50}, []float64{0, 50}))
	h.AddPolylinePath([][]float64{{10, 10}, {20, 10}, {20, 20}}, nil)
	h2 := must(d.Hatch([]float64{200, 0}, []float64{300, 0}, []float64{300, 50}))
	must(0, h2.SetPattern("ANSI31", 0, 1))
	h3 := entity.NewHatch()
	h3.AddEdgePath(
		&entity.HatchLine{Start: []float64{400, 0}, End: []float64{450, 0}},
		&entity.HatchArc{Center: []float64{450, 25}, Radius: 25, Start: 270, End: 90, CCW: true},
		&entity.HatchLine{Start: []float64{450, 50}, End: []float64{400, 50}},
		&entity.HatchLine{Start: []float64{400, 50}, End: []float64{400, 0}},
	)
	h3.SetUserPattern(30, 5, true)
	d.AddEntity(h3)
	must(d.LinearDimension([]float64{0, 0, 0}, []float64{100, 0, 0}, []float64{50, -20, 0}, 0))
	must(d.LinearDimension([]float64{0, 0, 0}, []float64{100, 50, 0}, []float64{-20, 0, 0}, 90))
	must(d.AlignedDimension([]float64{200, 0, 0}, []float64{300, 50, 0}, []float64{240, 60, 0}))
	must(d.RadiusDimension([]float64{500, 0, 0}, []float64{530, 40, 0}))
	must(d.DiameterDimension([]float64{600, 0, 0}, []float64{650, 0, 0}))
	must(d.AngularDimension([]float64{700, 0, 0}, []float64{750, 0, 0}, []float64{700, 50, 0}, []float64{730, 30, 0}))
	must(d.Leader([]float64{0, 100, 0}, []float64{20, 120, 0}, []float64{40, 120, 0}))
	ld, _ := must2(d.LeaderText("注記", 0, []float64{100, 100, 0}, []float64{120, 120, 0}))
	_ = ld
	return d
}

func must2[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(err)
	}
	return a, b
}

// colors uses true colors, transparency, gradient hatches and MTEXT
// background fill (AC1024).
func colors() *drawing.Drawing {
	d := dxf.NewDrawing()
	must(0, d.SetVersion(format.R2010))
	l := must(d.AddLayer("GLASS", color.Cyan, dxf.DefaultLineType, true))
	l.SetTrueColor(10, 200, 250)
	l.SetTransparency(0.3)
	ln := must(d.Line(0, 0, 0, 10, 0, 0))
	ln.SetTrueColor(255, 128, 0)
	ln.SetTransparency(0.5)
	c := must(d.Circle(0, 0, 0, 5))
	c.SetTrueColor(1, 2, 3)
	c.SetColorName("BOOK$NAME")
	c.SetTransparencyByBlock()

	// gradient hatches and MTEXT background fill
	h := must(d.Hatch([]float64{0, 0}, []float64{10, 0}, []float64{10, 5}, []float64{0, 5}))
	h.SetGradient(entity.GRADIENT_CYLINDER, color.RGB(255, 0, 0), color.RGB(0, 0, 255), 30)
	h = must(d.Hatch([]float64{20, 0}, []float64{30, 0}, []float64{30, 5}))
	h.SetOneColorGradient(entity.GRADIENT_SPHERICAL, color.RGB(0, 128, 0), 0.75, 0).Centered = 1
	m := must(d.MText("bg color", 0, 10, 0, 1, 10))
	m.SetBackgroundColor(color.Yellow, 1.2)
	m = must(d.MText("bg rgb", 0, 15, 0, 1, 10))
	m.SetBackgroundTrueColor(10, 20, 30, 2)
	m.SetTextFrame(true)
	m = must(d.MText("bg canvas", 0, 20, 0, 1, 10))
	m.SetBackgroundCanvas(1.5)
	m = must(d.MText("frame", 0, 25, 0, 1, 10))
	m.SetTextFrame(true)
	return d
}

// edited uses the editing API: explode, copy, transform, remove, rename.
func edited() *drawing.Drawing {
	d := blocks()
	ins := d.Entities()[0].(*entity.Insert)
	if _, err := dxf.Explode(d, ins); err != nil {
		panic(err)
	}
	must(d.AddLayer("A", color.Red, dxf.DefaultLineType, true))
	l := must(d.Line(0, 0, 0, 10, 0, 0))
	c := must(d.Circle(5, 5, 0, 2))
	h := must(d.Hatch([]float64{0, 0}, []float64{10, 0}, []float64{10, 10}))
	must(0, h.SetPattern("ANSI37", 0, 1))
	dim := must(d.LinearDimension([]float64{0, 0, 0}, []float64{10, 0, 0}, []float64{5, -5, 0}, 0))
	must(dxf.CopyEntities(d, entity.Rotate([]float64{0, 0, 0}, 30), l, c, h, dim))
	must(0, d.Scale([]float64{0, 0, 0}, 2, c, dim))
	must(0, d.RenameLayer("A", "B"))
	d.RemoveEntity(l)
	must(0, d.RenameBlock("TITLE", "TITLE2"))
	return d
}

func layouts() *drawing.Drawing {
	d := dxf.NewDrawing()
	must(d.Line(0, 0, 0, 1000, 500, 0))
	d.PaperSpace()
	must(d.LwPolyline(true, []float64{10, 10}, []float64{410, 10}, []float64{410, 287}, []float64{10, 287}))
	must(d.Viewport([]float64{210, 148.5, 0}, 380, 250, []float64{500, 250}, 600))
	d.ModelSpace()
	must(0, d.RenameLayout("Layout1", "配置1"))
	must(d.AddLayout("A3"))
	must(0, d.UseLayout("A3"))
	must(d.Text("A3 layout", 20, 20, 0, 5))
	must(d.Viewport([]float64{210, 148.5, 0}, 200, 150, []float64{0, 0}, 300))
	must(0, d.UseLayout("Model"))
	return d
}

func generate(out string) {
	files := map[string]func() *drawing.Drawing{
		"basic": basic, "blocks": blocks, "hatchdims": hatchDims, "layouts": layouts, "edited": edited, "colors": colors,
		"linetypes": linetypes, "underlays": underlays, "images": images, "more": more, "phaseb": phaseB, "phasec": phaseC, "phased": phaseD,
		"phasef": phaseF,
	}
	// the basic drawing with true colors in each version
	for _, v := range format.Versions()[1:] {
		v := v
		files["basic_"+v] = func() *drawing.Drawing {
			d := basic()
			must(0, d.SetVersion(v))
			l := d.Entities()[0].(*entity.Line)
			l.SetTrueColor(255, 128, 0)
			l.SetTransparency(0.25)
			return d
		}
	}
	for name, f := range files {
		d := f()
		d.SetExt()
		if err := d.SaveAs(filepath.Join(out, name+".dxf")); err != nil {
			panic(err)
		}
		d = f()
		if err := d.SaveAsBinary(filepath.Join(out, name+"_bin.dxf")); err != nil {
			panic(err)
		}
		d = f()
		d.Header().CodePage = codepage.Japanese
		if err := d.SaveAs(filepath.Join(out, name+"_sjis.dxf")); err != nil {
			panic(err)
		}
		if !strings.HasPrefix(name, "basic_") {
			if err := dxf.SaveAsR12(f(), filepath.Join(out, name+"_r12.dxf")); err != nil {
				panic(err)
			}
		}
	}
	fmt.Println("generated in", out)
}

func roundtrip(in, out string) {
	files, _ := filepath.Glob(filepath.Join(in, "*.dxf"))
	for _, fn := range files {
		d, err := dxf.FromFile(fn)
		if err != nil {
			fmt.Println(filepath.Base(fn), "READ ERROR:", err)
			continue
		}
		base := strings.TrimSuffix(filepath.Base(fn), ".dxf")
		if err := d.SaveAs(filepath.Join(out, base+"_rt.dxf")); err != nil {
			fmt.Println(base, "WRITE ERROR:", err)
		}
		d2, _ := dxf.FromFile(fn)
		if err := d2.SaveAsBinary(filepath.Join(out, base+"_rtbin.dxf")); err != nil {
			fmt.Println(base, "WRITE BIN ERROR:", err)
		}
		fmt.Println(base, "entities", len(d.Entities()), "blocks", len(d.Blocks()), "layouts", len(d.Layouts()), "groups", len(d.Groups))
	}
}

func main() {
	if len(os.Args) >= 3 && os.Args[1] == "generate" {
		generate(os.Args[2])
		return
	}
	if len(os.Args) >= 4 && os.Args[1] == "roundtrip" {
		roundtrip(os.Args[2], os.Args[3])
		return
	}
	if len(os.Args) >= 3 && os.Args[1] == "paths" {
		pathSamples(os.Args[2])
		return
	}
	if len(os.Args) >= 3 && os.Args[1] == "transform" {
		transformSamples(os.Args[2])
		return
	}
	fmt.Fprintln(os.Stderr, "usage: ezdxfcheck generate <dir> | roundtrip <in> <out> | transform <dir>")
	os.Exit(2)
}
