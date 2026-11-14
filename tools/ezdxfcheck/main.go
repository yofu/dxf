// Command ezdxfcheck creates DXF files for validation with ezdxf
// (see README.md):
//
//	go run ./tools/ezdxfcheck generate <dir>        sample files written by this package
//	go run ./tools/ezdxfcheck roundtrip <in> <out>  read files and write them back
package main

import (
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
	fmt.Fprintln(os.Stderr, "usage: ezdxfcheck generate <dir> | roundtrip <in> <out>")
	os.Exit(2)
}
