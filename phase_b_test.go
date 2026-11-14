package dxf

import (
	"math"
	"strings"
	"testing"

	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
)

func TestMLine(t *testing.T) {
	d := NewDrawing()
	m, err := d.MLine("", [][]float64{{0, 0}, {10, 0}, {10, 10}}, false, 2, entity.MLINE_ZERO)
	if err != nil {
		t.Fatal(err)
	}
	// Standard: elements at +-0.5, scale 2, zero justification: +-1 from the line
	if e := m.ElementPoints(0); !near(e[0], []float64{0, 1, 0}) || !near(e[1], []float64{9, 1, 0}) || !near(e[2], []float64{9, 10, 0}) {
		t.Errorf("element 0: %v", e)
	}
	if e := m.ElementPoints(1); !near(e[1], []float64{11, -1, 0}) {
		t.Errorf("element 1: %v", e)
	}
	wall, _ := d.AddMLineStyle("WALL", entity.MLineStyleElement{Offset: 0.1, Color: 1, LineType: "BYLAYER"},
		entity.MLineStyleElement{Offset: -0.1, Color: 5, LineType: "BYLAYER"})
	wall.Flags = entity.MLINESTYLE_FILL
	closed, _ := d.MLine("WALL", [][]float64{{0, 0}, {4, 0}, {4, 4}, {0, 4}}, true, 1, entity.MLINE_TOP)
	if _, err := d.MLine("NONE", [][]float64{{0, 0}, {1, 0}}, false, 1, 0); err == nil {
		t.Error("unknown style should be an error")
	}
	out := writeString(t, d)
	checkStructure(t, d)
	for _, s := range []string{"\n0\nMLINE\n", "\n100\nAcDbMline\n2\nStandard\n340\n", "\n3\nACAD_MLINESTYLE\n",
		"\n0\nMLINESTYLE\n", "\n100\nAcDbMlineStyle\n2\nWALL\n70\n1\n"} {
		if !strings.Contains(out, s) {
			t.Errorf("%q not found", s)
		}
	}
	d2 := readPreserve(t, out)
	r := d2.Entities()[0].(*entity.MLine)
	rc := d2.Entities()[1].(*entity.MLine)
	if r.Style == nil || r.Style.Name() != "Standard" || len(r.Vertices) != 3 || !near(r.ElementPoints(0)[1], []float64{9, 1, 0}) ||
		rc.Style == nil || rc.Style.Name() != "WALL" || !rc.IsClosed() || len(rc.Style.Elements) != 2 || rc.Style.Elements[1].Color != 5 {
		t.Fatalf("mlines: %+v %+v", r, rc)
	}
	if raws := d2.RawObjects(); len(raws) != 0 {
		t.Errorf("raw objects: %v", raws)
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
	// transformation: the element points are transformed
	want := entity.MirrorBy([]float64{0, 0}, []float64{1, 1}).Point(closed.ElementPoints(1)[2])
	d.Mirror([]float64{0, 0}, []float64{1, 1}, closed)
	if got := closed.ElementPoints(1)[2]; !near(got, want) {
		t.Errorf("mirrored element point %v, want %v", got, want)
	}
	want = entity.ScaleXYZ(nil, 2, 0.5, 1).Point(m.ElementPoints(0)[1])
	d.ScaleXYZ(nil, 2, 0.5, 1, m)
	if got := m.ElementPoints(0)[1]; !near(got, want) {
		t.Errorf("scaled element point %v, want %v", got, want)
	}
}

func TestMultiLeader(t *testing.T) {
	d := NewDrawing()
	d.SetVersion(format.R2013)
	m, err := d.MultiLeaderText("Note\\Psecond line", []float64{10, 10, 0}, 2.5,
		[][]float64{{0, 0, 0}, {5, 8, 0}}, [][]float64{{30, 0, 0}})
	if err != nil {
		t.Fatal(err)
	}
	c := m.Context
	if len(c.Leaders) != 2 || c.Leaders[0].Dogleg[0] != 1 || c.Leaders[1].Dogleg[0] != -1 ||
		!near(c.Leaders[0].LastPoint, []float64{10 - 2 - 8, 10 - 1.25, 0}) {
		t.Errorf("leaders: %+v %+v", c.Leaders[0], c.Leaders[1])
	}
	out := writeString(t, d)
	checkStructure(t, d)
	for _, s := range []string{"\n0\nMULTILEADER\n", "\n100\nAcDbMLeader\n270\n2\n300\nCONTEXT_DATA{\n", "\n302\nLEADER{\n",
		"\n304\nLEADER_LINE{\n", "\n304\nNote\\Psecond line\n", "\n3\nACAD_MLEADERSTYLE\n", "\n0\nMLEADERSTYLE\n",
		"\n1\nMLEADERSTYLE\n", "\n295\n0\n"} {
		if !strings.Contains(out, s) {
			t.Errorf("%q not found", s)
		}
	}
	d2 := readPreserve(t, out)
	r := d2.Entities()[0].(*entity.MultiLeader)
	rc := r.Context
	if rc.MText == nil || rc.MText.Content != "Note\\Psecond line" || len(rc.Leaders) != 2 || len(rc.Leaders[1].Lines[0].Vertices) != 1 ||
		!near(rc.Leaders[0].Lines[0].Vertices[1], []float64{5, 8, 0}) || rc.CharHeight != 2.5 || !r.Style.IsSet() {
		t.Fatalf("multileader: %+v", rc)
	}
	if len(d2.MLeaderStyles()) != 1 || len(d2.RawObjects()) != 0 {
		t.Errorf("styles %d, raw objects %v", len(d2.MLeaderStyles()), d2.RawObjects())
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
	if err := d2.Rotate([]float64{0, 0, 0}, 90, r); err != nil || !near(rc.MText.Insert, []float64{-10, 10, 0}) ||
		math.Abs(rc.MText.Rotation-math.Pi/2) > 1e-12 {
		t.Errorf("rotated: %v %v %v", err, rc.MText.Insert, rc.MText.Rotation)
	}
	if err := d2.Mirror([]float64{0, 0}, []float64{1, 0}, r); err == nil {
		t.Error("mirrored multileaders are not supported")
	}
}

func TestAcadTable(t *testing.T) {
	src := "0\nSECTION\n2\nHEADER\n9\n$ACADVER\n1\nAC1021\n0\nENDSEC\n" +
		"0\nSECTION\n2\nBLOCKS\n0\nBLOCK\n8\n0\n2\n*T1\n70\n1\n10\n0\n20\n0\n30\n0\n0\nLINE\n8\n0\n10\n0\n20\n0\n30\n0\n11\n10\n21\n0\n31\n0\n0\nENDBLK\n8\n0\n0\nENDSEC\n" +
		"0\nSECTION\n2\nENTITIES\n0\nACAD_TABLE\n5\n40\n100\nAcDbEntity\n8\nT\n100\nAcDbBlockReference\n2\n*T1\n10\n5\n20\n6\n30\n0\n" +
		"100\nAcDbTable\n280\n0\n342\n0\n343\n0\n11\n1\n21\n0\n31\n0\n90\n22\n91\n2\n92\n2\n93\n0\n94\n0\n95\n0\n96\n0\n141\n5\n141\n5\n142\n10\n142\n10\n" +
		"171\n1\n172\n0\n301\nCELL_VALUE\n302\nA1\n304\nACVALUE_END\n171\n1\n172\n0\n301\nCELL_VALUE\n302\nB1\n304\nACVALUE_END\n" +
		"171\n1\n172\n0\n301\nCELL_VALUE\n302\nA2\n304\nACVALUE_END\n171\n1\n172\n0\n301\nCELL_VALUE\n302\nB2\n304\nACVALUE_END\n" +
		"0\nENDSEC\n0\nEOF\n"
	d := readPreserve(t, src)
	tb := d.Entities()[0].(*entity.AcadTable)
	if tb.BlockName != "*T1" || !near(tb.Insert, []float64{5, 6, 0}) {
		t.Fatalf("table: %+v", tb)
	}
	content := tb.Content()
	if len(content) != 2 || content[0][1] != "B1" || content[1][0] != "A2" {
		t.Errorf("content: %v", content)
	}
	out := writeString(t, d)
	if writeString(t, readPreserve(t, out)) != out || !strings.Contains(out, "\n301\nCELL_VALUE\n302\nB2\n") {
		t.Error("table data should be kept")
	}
	d.Move(1, 0, 0, tb)
	if !near(tb.Insert, []float64{6, 6, 0}) {
		t.Errorf("moved: %v", tb.Insert)
	}
	ins, err := d.TableToBlockReference(tb)
	if err != nil || d.Entities()[0] != ins || ins.BlockName != "*T1" || ins.Layer() != tb.Layer() {
		t.Errorf("block reference: %v %+v", err, ins)
	}
}
