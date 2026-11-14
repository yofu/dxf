package dxf

import (
	"strings"
	"testing"

	"github.com/yofu/dxf/entity"
)

func TestThicknessRoundTrip(t *testing.T) {
	d := NewDrawing()
	l, _ := d.Line(0, 0, 0, 1, 0, 0)
	c, _ := d.Circle(0, 0, 0, 1)
	a, _ := d.Arc(0, 0, 0, 1, 0, 90)
	p, _ := d.Point(0, 0, 0)
	tx, _ := d.Text("x", 0, 0, 0, 1)
	lw, _ := d.LwPolyline(false, []float64{0, 0}, []float64{1, 0})
	pl, _ := d.Polyline(false, []float64{0, 0, 0}, []float64{1, 1, 0})
	s := entity.NewSolid()
	s.Points[1] = []float64{1, 0, 0}
	s.Points[2] = []float64{0, 1, 0}
	s.Points[3] = []float64{0, 1, 0}
	d.AddEntity(s)
	d.BeginBlock("B", 0, 0, 0)
	d.AttributeDefinition("TAG", "", "", 0, 0, 0, 1)
	d.EndBlock()
	ins, _ := d.Insert("B", 0, 0, 0)

	things := []entity.Thick{l, c, a, p, tx, lw, pl, s, ins.Attributes[0]}
	for i, th := range things {
		th.SetThickness(float64(i + 1))
	}
	b, _ := d.Block("B")
	b.Entities[0].(*entity.AttributeDefinition).SetThickness(-2)

	if str := l.String(); !strings.Contains(str, "AcDbLine\n39\n1.0") {
		t.Error("LINE: 39 should follow the subclass marker")
	}
	if str := p.String(); !strings.Contains(str, "\n30\n0.000000\n39\n4.0") {
		t.Error("POINT: 39 should follow the point")
	}

	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	es := d2.Entities()
	want := []float64{1, 2, 3, 4, 5, 6, 7, 8}
	got := []float64{
		es[0].(*entity.Line).Thickness, es[1].(*entity.Circle).Thickness, es[2].(*entity.Arc).Thickness,
		es[3].(*entity.Point).Thickness, es[4].(*entity.Text).Thickness, es[5].(*entity.LwPolyline).Thickness,
		es[6].(*entity.Polyline).Thickness, es[7].(*entity.Solid).Thickness,
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entity %d (%T): thickness %v, want %v", i, es[i], got[i], want[i])
		}
	}
	if th := es[8].(*entity.Insert).Attributes[0].Thickness; th != 9 {
		t.Errorf("ATTRIB thickness: %v", th)
	}
	b2, _ := d2.Block("B")
	if th := b2.Entities[0].(*entity.AttributeDefinition).Thickness; th != -2 {
		t.Errorf("ATTDEF thickness: %v", th)
	}

	// bounding boxes are extruded along Z
	if mins, maxs := es[0].BBox(); mins[2] != 0 || maxs[2] != 1 {
		t.Errorf("line bbox z: %v - %v", mins[2], maxs[2])
	}
	l.Thickness = -3
	if mins, maxs := l.BBox(); mins[2] != -3 || maxs[2] != 0 {
		t.Errorf("negative thickness bbox z: %v - %v", mins[2], maxs[2])
	}
}
