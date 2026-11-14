package dxf

import (
	"reflect"
	"strings"
	"testing"

	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/object"
)

func TestXRecordAndVariables(t *testing.T) {
	d := NewDrawing()
	l, _ := d.Line(0, 0, 0, 1, 0, 0)
	tags := [][2]string{{"1", "data"}, {"40", "1.5"}, {"90", "7"}}
	if _, err := d.SetXRecord(l, "MYAPP", tags); err != nil {
		t.Fatal(err)
	}
	layer := d.Layers["0"]
	d.SetXRecord(layer, "LAYERDATA", [][2]string{{"1", "on layer"}})
	d.SetVariable("XCLIPFRAME", "2")
	out := writeString(t, d)
	checkStructure(t, d)
	for _, s := range []string{"\n102\n{ACAD_XDICTIONARY\n360\n", "\n0\nXRECORD\n", "\n100\nAcDbXrecord\n280\n1\n1\ndata\n",
		"\n3\nAcDbVariableDictionary\n", "\n0\nDICTIONARYVAR\n", "\n100\nDictionaryVariables\n280\n0\n1\n2\n"} {
		if !strings.Contains(out, s) {
			t.Errorf("%q not found", s)
		}
	}
	d2 := readPreserve(t, out)
	x, err := d2.XRecord(d2.Entities()[0], "MYAPP")
	if err != nil || x == nil || !reflect.DeepEqual(x.Tags(), tags) {
		t.Fatalf("xrecord: %v %v", err, x)
	}
	if lx, _ := d2.XRecord(d2.Layers["0"], "LAYERDATA"); lx == nil || lx.Tags()[0][1] != "on layer" {
		t.Errorf("layer xrecord: %v", lx)
	}
	if v, ok := d2.Variable("XCLIPFRAME"); !ok || v != "2" {
		t.Errorf("variable: %q %v", v, ok)
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
	// change a record read
	d2.SetXRecord(d2.Entities()[0], "MYAPP", [][2]string{{"1", "new"}})
	d2.SetVariable("XCLIPFRAME", "1")
	d3 := readPreserve(t, writeString(t, d2))
	if x, _ := d3.XRecord(d3.Entities()[0], "MYAPP"); x == nil || len(x.Tags()) != 1 || x.Tags()[0][1] != "new" {
		t.Errorf("changed xrecord: %v", x)
	}
	if v, _ := d3.Variable("XCLIPFRAME"); v != "1" {
		t.Errorf("changed variable: %q", v)
	}
}

func TestDrawOrderAndXClip(t *testing.T) {
	d := NewDrawing()
	a, _ := d.Line(0, 0, 0, 1, 0, 0)
	b, _ := d.Circle(0, 0, 0, 1)
	c, _ := d.Point(0, 0, 0)
	d.SetDrawOrder("", a) // a on top
	if got := d.DrawOrder(""); len(got) != 3 || got[0] != b || got[2] != a {
		t.Fatalf("draw order: %v", got)
	}
	d.BeginBlock("B", 0, 0, 0)
	d.Line(0, 0, 0, 10, 10, 0)
	d.EndBlock()
	ins, _ := d.Insert("B", 100, 0, 0)
	ins.Rotation = 90
	if _, err := d.ClipBlockReference(ins, [][]float64{{99, 1}, {101, 5}}); err != nil {
		t.Fatal(err)
	}
	out := writeString(t, d)
	checkStructure(t, d)
	for _, s := range []string{"\n0\nSORTENTSTABLE\n", "\n3\nACAD_SORTENTS\n", "\n100\nAcDbSortentsTable\n",
		"\n3\nACAD_FILTER\n", "\n3\nSPATIAL\n", "\n100\nAcDbFilter\n100\nAcDbSpatialFilter\n70\n4\n"} {
		if !strings.Contains(out, s) {
			t.Errorf("%q not found", s)
		}
	}
	d2 := readPreserve(t, out)
	es := d2.Entities()
	if got := d2.DrawOrder(""); got[0] != es[1] || got[1] != es[2] || got[2] != es[0] {
		t.Errorf("draw order read: %v", got)
	}
	f := d2.BlockReferenceClipping(es[3].(*entity.Insert))
	if f == nil || len(f.Boundary()) != 4 || !near(f.Boundary()[0], []float64{1, 1}) || !near(f.Boundary()[2], []float64{5, -1}) || !f.Enabled() {
		t.Fatalf("clipping: %+v", f)
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
	_ = c
}

func TestPageSetupAndImageVariables(t *testing.T) {
	d := NewDrawing()
	l := d.Layouts()[1]
	l.PageSetup(420, 297, object.PLOT_MILLIMETERS, [4]float64{10, 10, 10, 20}, 1, 50, "")
	d.SetRasterVariables(1, 1, 1)
	d.SetWipeoutVariables(2)
	out := writeString(t, d)
	checkStructure(t, d)
	for _, s := range []string{"\n2\nDWG To PDF.pc3\n4\nUSER_(420.00_x_297.00_MM)\n", "\n73\n1\n", "\n3\nACAD_IMAGE_VARS\n",
		"\n0\nRASTERVARIABLES\n", "\n0\nWIPEOUTVARIABLES\n"} {
		if !strings.Contains(out, s) {
			t.Errorf("%q not found", s)
		}
	}
	d2 := readPreserve(t, out)
	ps := d2.Layouts()[1].PlotSettings
	if ps == nil || ps.PaperWidth != 420 || ps.MarginLeft != 20 || ps.Rotation != 1 || ps.ScaleDenominator != 50 || d2.Layouts()[1].LimMax[0] != 400*50 {
		t.Fatalf("plot settings: %+v %v", ps, d2.Layouts()[1].LimMax)
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
}
