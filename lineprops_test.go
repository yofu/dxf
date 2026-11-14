package dxf

import (
	"strings"
	"testing"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/table"
)

func TestEntityLineTypeAndWeight(t *testing.T) {
	d := NewDrawing()
	hidden, err := d.LineType("HIDDEN")
	if err != nil {
		t.Fatal(err)
	}
	byblock, _ := d.LineType("ByBlock")
	d.BeginBlock("B", 0, 0, 0)
	d.Line(0, 0, 0, 1, 0, 0)
	d.EndBlock()

	var ents []entity.LineProperties
	l, _ := d.Line(0, 0, 0, 1, 1, 0)
	c, _ := d.Circle(0, 0, 0, 1)
	a, _ := d.Arc(0, 0, 0, 1, 0, 90)
	tx, _ := d.Text("x", 0, 0, 0, 1)
	mt, _ := d.MText("x", 0, 0, 0, 1, 0)
	p, _ := d.Point(0, 0, 0)
	f, _ := d.ThreeDFace([][]float64{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}})
	lw, _ := d.LwPolyline(false, []float64{0, 0}, []float64{1, 0})
	pl, _ := d.Polyline(false, []float64{0, 0, 0}, []float64{1, 1, 1})
	sp, _ := d.Spline([][]float64{{0, 0, 0}, {1, 1, 0}, {2, 0, 0}, {3, 1, 0}})
	ins, _ := d.Insert("B", 0, 0, 0)
	ents = append(ents, l, c, a, tx, mt, p, f, lw, pl, sp, ins)
	for _, e := range ents {
		e.SetLineType(hidden)
		e.SetLineWeight(35)
	}
	pl.Vertices[0].SetLineWeight(table.LineWeightByBlock)
	pl.Vertices[0].SetLineType(byblock)
	l.SetColor(color.Red)

	out := writeString(t, d)
	if !strings.Contains(out, "\n8\n0\n6\nHIDDEN\n62\n1\n370\n35\n") {
		t.Error("6, 62, 370 should follow 8 in this order")
	}

	d2, err := FromStringData(out)
	if err != nil {
		t.Fatal(err)
	}
	es := d2.Entities()
	if len(es) != len(ents) {
		t.Fatalf("want %d entities, got %d", len(ents), len(es))
	}
	for _, e := range es {
		lp := e.(entity.LineProperties)
		if lp.LineTypeName() != "HIDDEN" || lp.LineWeight() != 35 {
			t.Errorf("%T: linetype %s lineweight %d", e, lp.LineTypeName(), lp.LineWeight())
		}
	}
	if lt, _ := d2.LineType("HIDDEN"); es[0].(entity.LineProperties).LineType() != lt {
		t.Error("linetype should refer to the drawing's LTYPE entry")
	}
	v := es[8].(*entity.Polyline).Vertices
	if v[0].LineTypeName() != "ByBlock" || v[0].LineWeight() != table.LineWeightByBlock ||
		v[1].LineType() != nil || v[1].LineWeight() != table.LineWeightByLayer {
		t.Errorf("vertices: %s %d / %s %d", v[0].LineTypeName(), v[0].LineWeight(), v[1].LineTypeName(), v[1].LineWeight())
	}
}

func TestEntityLinePropertiesDefault(t *testing.T) {
	d := NewDrawing()
	l, _ := d.Line(0, 0, 0, 1, 1, 0)
	if l.LineType() != nil || l.LineTypeName() != "ByLayer" || l.LineWeight() != table.LineWeightByLayer {
		t.Errorf("default should be BYLAYER: %s %d", l.LineTypeName(), l.LineWeight())
	}
	s := l.String()
	if strings.Contains(s, "\n6\n") || strings.Contains(s, "\n370\n") {
		t.Error("BYLAYER linetype/lineweight should not be written")
	}
	bylayer, _ := d.LineType("ByLayer")
	l.SetLineType(bylayer)
	if l.LineType() != nil {
		t.Error("ByLayer linetype should be stored as nil")
	}
}

func TestSetLineWeight(t *testing.T) {
	l := entity.NewLine()
	for in, want := range map[int]int{0: 0, 33: 35, 35: 35, 212: 211, -1: -1, -2: -2, -3: -3, -5: -1, 1: 5} {
		if got := l.SetLineWeight(in); got != want {
			t.Errorf("SetLineWeight(%d) = %d, want %d", in, got, want)
		}
	}
}

func TestReadEntityLineProperties(t *testing.T) {
	src := "0\nSECTION\n2\nTABLES\n0\nTABLE\n2\nLTYPE\n70\n1\n" +
		"0\nLTYPE\n2\nDASHED\n70\n0\n3\n__ __\n72\n65\n73\n2\n40\n0.75\n49\n0.5\n49\n-0.25\n0\nENDTAB\n0\nENDSEC\n" +
		entitiesSection("0\nLINE\n8\n0\n6\ndashed\n370\n-3\n10\n0\n20\n0\n11\n1\n21\n1\n"+
			"0\nLINE\n8\n0\n6\nBYLAYER\n370\n-1\n10\n0\n20\n0\n11\n1\n21\n1\n"+
			"0\nLINE\n8\n0\n6\nUNDEFINED\n370\n30\n10\n0\n20\n0\n11\n1\n21\n1\n")
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	es := d.Entities()
	l0 := es[0].(*entity.Line)
	if l0.LineTypeName() != "DASHED" || l0.LineWeight() != table.LineWeightDefault {
		t.Errorf("line 0: %s %d", l0.LineTypeName(), l0.LineWeight())
	}
	if l1 := es[1].(*entity.Line); l1.LineType() != nil || l1.LineWeight() != table.LineWeightByLayer {
		t.Errorf("line 1: %s %d", l1.LineTypeName(), l1.LineWeight())
	}
	if l2 := es[2].(*entity.Line); l2.LineTypeName() != "UNDEFINED" || l2.LineWeight() != 30 {
		t.Errorf("line 2: %s %d", l2.LineTypeName(), l2.LineWeight())
	}
}
