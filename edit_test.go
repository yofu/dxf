package dxf

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/table"
)

func near(a, b []float64) bool {
	for i := range b {
		if i >= len(a) || math.Abs(a[i]-b[i]) > 1e-9 {
			return false
		}
	}
	return true
}

func TestTransformEntities(t *testing.T) {
	d := NewDrawing()
	l, _ := d.Line(1, 0, 0, 2, 0, 0)
	c, _ := d.Circle(1, 1, 0, 1)
	a, _ := d.Arc(0, 0, 0, 1, 0, 90)
	tx, _ := d.Text("abc", 1, 0, 0, 2)
	if err := d.Rotate([]float64{0, 0, 0}, 90, l, c, a, tx); err != nil {
		t.Fatal(err)
	}
	if !near(l.Start, []float64{0, 1, 0}) || !near(l.End, []float64{0, 2, 0}) {
		t.Errorf("line: %v %v", l.Start, l.End)
	}
	if !near(c.Center, []float64{-1, 1, 0}) || a.Angle[0] != 90 || a.Angle[1] != 180 {
		t.Errorf("circle %v arc %v", c.Center, a.Angle)
	}
	if !near(tx.Coord1, []float64{0, 1, 0}) || tx.Rotation != 90 {
		t.Errorf("text: %v %v", tx.Coord1, tx.Rotation)
	}
	if err := d.Scale([]float64{0, 0, 0}, 2, c, tx); err != nil {
		t.Fatal(err)
	}
	if c.Radius != 2 || !near(c.Center, []float64{-2, 2, 0}) || tx.Height != 4 {
		t.Errorf("scale: circle %v r %v, text height %v", c.Center, c.Radius, tx.Height)
	}
	lw, _ := d.LwPolyline(false, []float64{0, 0}, []float64{1, 0})
	lw.SetWidth(0, 1, 1)
	if err := d.Move(10, 20, 5, lw); err != nil {
		t.Fatal(err)
	}
	if !near(lw.Vertices[1], []float64{11, 20}) || lw.Elevation != 5 || len(lw.Vertices[1]) != 2 {
		t.Errorf("lwpolyline: %v elevation %v", lw.Vertices, lw.Elevation)
	}
	if err := d.Scale(nil, 0, l); err == nil {
		t.Error("scale 0 should be an error")
	}

	// errors
	// extrusion -Z: the OCS X axis is -X
	ext := entity.NewCircle()
	ext.Direction = []float64{0, 0, -1}
	if err := d.Move(1, 0, 0, ext); err != nil || !near(ext.Center, []float64{-1, 0, 0}) || ext.Direction[2] != -1 {
		t.Errorf("extruded circle: %v %v %v", err, ext.Center, ext.Direction)
	}
	tilted := entity.NewCircle()
	tilted.Direction = []float64{1, 0, 0}
	if err := d.ScaleXYZ(nil, 2, 1, 1, tilted); !errors.Is(err, entity.ErrTransform) {
		t.Errorf("non-uniform scaling of a tilted circle: %v", err)
	}
	if err := d.Move(1, 0, 0, entity.NewUnknown("XLINE")); !errors.Is(err, entity.ErrTransform) {
		t.Errorf("unknown entity: %v", err)
	}
}

func TestTransformInsertHatchDimension(t *testing.T) {
	d := NewDrawing()
	d.BeginBlock("B", 0, 0, 0)
	d.AttributeDefinition("TAG", "", "v", 1, 0, 0, 1)
	d.EndBlock()
	ins, _ := d.Insert("B", 10, 0, 0)
	h, _ := d.Hatch([]float64{0, 0}, []float64{10, 0}, []float64{10, 10})
	h.SetPattern("ANSI31", 0, 1)
	dim, _ := d.LinearDimension([]float64{0, 0, 0}, []float64{100, 0, 0}, []float64{50, 20, 0}, 0)
	b, _ := d.Block(dim.BlockName)
	ext := b.Entities[0].(*entity.Line)
	before := append([]float64(nil), ext.Start...)

	if err := d.Rotate([]float64{0, 0, 0}, 90, ins, h); err != nil {
		t.Fatal(err)
	}
	if !near(ins.Coord, []float64{0, 10, 0}) || ins.Rotation != 90 || !near(ins.Attributes[0].Coord1, []float64{0, 11, 0}) {
		t.Errorf("insert: %v %v attribute %v", ins.Coord, ins.Rotation, ins.Attributes[0].Coord1)
	}
	if !near(h.Boundaries[0].Vertices[1], []float64{0, 10}) || h.PatternAngle != 90 || h.PatternLines[0].Angle != 135 ||
		!near(h.PatternLines[0].Offset, []float64{-2.2450640302672884, -2.245064030267288}) {
		t.Errorf("hatch: %v angle %v line %+v", h.Boundaries[0].Vertices, h.PatternAngle, h.PatternLines[0])
	}

	if err := d.Move(5, 5, 0, dim); err != nil {
		t.Fatal(err)
	}
	if !near(dim.DefPoint2, []float64{5, 5, 0}) || !near(ext.Start, []float64{before[0] + 5, before[1] + 5, 0}) || dim.Measurement != 100 {
		t.Errorf("dimension move: %v, block line %v", dim.DefPoint2, ext.Start)
	}
	d.Scale([]float64{0, 0, 0}, 2, dim)
	if dim.Measurement != 200 {
		t.Errorf("dimension scale: measurement %v", dim.Measurement)
	}
	checkStructure(t, d)
}

func TestRemoveAndQuery(t *testing.T) {
	d := NewDrawing()
	d.AddLayer("A", color.Red, DefaultLineType, false)
	l1, _ := d.Line(0, 0, 0, 1, 0, 0)
	d.ChangeLayer("A")
	l2, _ := d.Line(0, 0, 0, 0, 1, 0)
	c, _ := d.Circle(0, 0, 0, 1)
	d.ChangeLayer("0")
	d.BeginBlock("B", 0, 0, 0)
	inBlock, _ := d.Line(0, 0, 0, 1, 1, 0)
	d.EndBlock()
	g, _ := d.Group("G", "", l1, l2)
	ld, mt, _ := d.LeaderText("note", 0, []float64{0, 0, 0}, []float64{5, 5, 0})

	if es := d.EntitiesOnLayer("a"); len(es) != 2 || es[0] != l2 || es[1] != c {
		t.Errorf("EntitiesOnLayer: %v", es)
	}
	circles := d.Query(func(e entity.Entity) bool { _, ok := e.(*entity.Circle); return ok })
	if len(circles) != 1 {
		t.Errorf("Query: %d", len(circles))
	}

	if n := d.RemoveEntity(l1, inBlock, mt, entity.NewLine()); n != 3 {
		t.Errorf("removed %d", n)
	}
	b, _ := d.Block("B")
	if len(b.Entities) != 0 || len(g.Entities()) != 1 || g.Entities()[0] != l2 || ld.Annotation != nil {
		t.Errorf("block %d, group %v, leader annotation %v", len(b.Entities), g.Entities(), ld.Annotation)
	}
	for _, e := range d.Entities() {
		if e == l1 || e == mt {
			t.Error("removed entity is still in ENTITIES")
		}
	}
	checkStructure(t, d)
}

func TestLayerAndBlockOperations(t *testing.T) {
	d := NewDrawing()
	d.AddLayer("A", color.Red, DefaultLineType, false)
	d.AddLayer("UNUSED", color.Red, DefaultLineType, false)
	d.ChangeLayer("A")
	l, _ := d.Line(0, 0, 0, 1, 0, 0)
	d.ChangeLayer("0")
	if err := d.RenameLayer("A", "壁"); err != nil {
		t.Fatal(err)
	}
	if l.Layer().Name() != "壁" || d.Layers["壁"] == nil || d.Layers["A"] != nil {
		t.Error("RenameLayer")
	}
	for _, name := range []string{"0", "壁", "NONE"} {
		if err := d.RemoveLayer(name); err == nil {
			t.Errorf("RemoveLayer(%s) should be an error", name)
		}
	}
	if err := d.RenameLayer("0", "X"); err == nil {
		t.Error("layer 0 cannot be renamed")
	}
	if err := d.RemoveLayer("UNUSED"); err != nil {
		t.Fatal(err)
	}
	out := writeString(t, d)
	// 壁 is written as \U+58C1 in the default code page
	if strings.Contains(out, "\n2\nUNUSED\n") || !strings.Contains(out, "\n8\n"+`\U+58C1`+"\n") {
		t.Error("layers in output")
	}

	d.BeginBlock("B", 0, 0, 0)
	d.Line(0, 0, 0, 1, 0, 0)
	d.EndBlock()
	d.BeginBlock("UNREF", 0, 0, 0)
	d.EndBlock()
	ins, _ := d.Insert("B", 0, 0, 0)
	if err := d.RenameBlock("B", "DOOR"); err != nil {
		t.Fatal(err)
	}
	if ins.BlockName != "DOOR" {
		t.Error("insert should follow the renamed block")
	}
	if err := d.RemoveBlock("DOOR"); err == nil {
		t.Error("referenced block cannot be removed")
	}
	if err := d.RemoveBlock("*Model_Space"); err == nil {
		t.Error("layout block cannot be removed")
	}
	if err := d.RenameBlock("DOOR", "*X"); err == nil {
		t.Error("anonymous name should be an error")
	}
	if err := d.RemoveBlock("UNREF"); err != nil {
		t.Fatal(err)
	}
	out = writeString(t, d)
	if strings.Contains(out, "UNREF") || !strings.Contains(out, "\n2\nDOOR\n") {
		t.Error("blocks in output")
	}
	checkStructure(t, d)
}

func TestCopy(t *testing.T) {
	d := NewDrawing()
	l, _ := d.Line(0.1, 0.2, 0, 1e-20, 1, 0)
	l.SetXData([][2]string{{"1001", "APP"}, {"1000", "x"}})
	l.SetColor(color.Red)
	c, err := Copy(d, l)
	if err != nil {
		t.Fatal(err)
	}
	cl := c.(*entity.Line)
	if cl == l || cl.Start[0] != 0.1 || cl.End[0] != 1e-20 || cl.ColorIndex() != 1 || len(cl.XData()) != 2 || cl.Layer() != l.Layer() {
		t.Errorf("copy: %+v", cl)
	}
	cl.Start[0] = 5
	if l.Start[0] != 0.1 {
		t.Error("copy should be independent")
	}

	pl, _ := d.Polyline(true, []float64{0, 0, 0}, []float64{1, 1, 1})
	cp, _ := Copy(d, pl)
	if p := cp.(*entity.Polyline); len(p.Vertices) != 2 || p.Vertices[1].Coord[2] != 1 || p.Flag != pl.Flag {
		t.Errorf("polyline copy: %+v", p)
	}
	d.BeginBlock("B", 0, 0, 0)
	d.AttributeDefinition("TAG", "", "v", 0, 0, 0, 1)
	d.EndBlock()
	ins, _ := d.Insert("B", 0, 0, 0)
	ins.SetAttribute("TAG", "値")
	ci, _ := Copy(d, ins)
	if a := ci.(*entity.Insert).Attribute("TAG"); a == nil || a.Value != "値" || a == ins.Attribute("TAG") {
		t.Error("insert copy with attributes")
	}
	dim, _ := d.LinearDimension([]float64{0, 0, 0}, []float64{10, 0, 0}, []float64{5, 5, 0}, 0)
	cd, _ := Copy(d, dim)
	if cdim := cd.(*entity.Dimension); cdim.BlockName == dim.BlockName || cdim.BlockName == "" {
		t.Errorf("dimension copy should have its own block: %s", cdim.BlockName)
	}

	copies, err := CopyEntities(d, entity.Move(100, 0, 0), l, dim)
	if err != nil {
		t.Fatal(err)
	}
	es := d.Entities()
	if len(copies) != 2 || es[1] != copies[0] || !near(copies[0].(*entity.Line).Start, []float64{100.1, 0.2, 0}) {
		t.Errorf("CopyEntities: copy should follow the original and be moved")
	}
	b1, _ := d.Block(dim.BlockName)
	b2, _ := d.Block(copies[1].(*entity.Dimension).BlockName)
	if b1.Entities[0].(*entity.Line).Start[0]+100 != b2.Entities[0].(*entity.Line).Start[0] {
		t.Error("copied dimension graphics should be moved")
	}
	checkStructure(t, d)
}

func TestExplode(t *testing.T) {
	d := NewDrawing()
	d.AddLayer("A", color.Green, DefaultLineType, false)
	d.AddLayer("INS", color.Blue, DefaultLineType, false)
	d.BeginBlock("B", 1, 0, 0)
	l, _ := d.Line(1, 0, 0, 2, 0, 0) // layer 0
	l.SetColorByBlock()
	l.SetLineWeight(table.LineWeightByBlock)
	d.ChangeLayer("A")
	d.Circle(1, 1, 0, 0.5)
	d.ChangeLayer("0")
	d.AttributeDefinition("TAG", "", "default", 1, 2, 0, 1)
	d.EndBlock()
	d.ChangeLayer("INS")
	ins, _ := d.Insert("B", 10, 10, 0)
	d.ChangeLayer("0")
	ins.SetColor(color.Magenta)
	ins.SetLineWeight(50)
	ins.Scale = []float64{2, 2, 2}
	ins.Rotation = 90
	ins.Columns, ins.ColumnSpacing = 2, 5
	d.SyncAttributes(ins)
	after, _ := d.Point(0, 0, 0)

	es, err := Explode(d, ins)
	if err != nil {
		t.Fatal(err)
	}
	if len(es) != 5 { // (line + circle) x 2 columns + attribute text
		t.Fatalf("exploded %d entities", len(es))
	}
	line := es[0].(*entity.Line)
	// (1,0)-(2,0) relative to the base (1,0): scaled 2, rotated 90, at (10,10)
	if !near(line.Start, []float64{10, 10, 0}) || !near(line.End, []float64{10, 12, 0}) {
		t.Errorf("line: %v %v", line.Start, line.End)
	}
	if line.Layer().Name() != "INS" || line.ColorIndex() != int(color.Magenta) || line.LineWeight() != 50 {
		t.Errorf("line properties: %s %d %d", line.Layer().Name(), line.ColorIndex(), line.LineWeight())
	}
	circle := es[1].(*entity.Circle)
	if circle.Layer().Name() != "A" || circle.Radius != 1 || !near(circle.Center, []float64{8, 10, 0}) {
		t.Errorf("circle: %v r %v layer %s", circle.Center, circle.Radius, circle.Layer().Name())
	}
	// second column: 5 along the rotated X axis
	if l2 := es[2].(*entity.Line); !near(l2.Start, []float64{10, 15, 0}) {
		t.Errorf("second column: %v", l2.Start)
	}
	if tx, ok := es[4].(*entity.Text); !ok || tx.Value != "default" || !near(tx.Coord1, []float64{6, 10, 0}) {
		t.Errorf("attribute text: %+v", es[4])
	}
	all := d.Entities()
	if all[0] != es[0] || all[len(all)-1] != after {
		t.Error("exploded entities should replace the insert")
	}
	for _, e := range all {
		if e == ins {
			t.Error("insert should be removed")
		}
	}
	checkStructure(t, d)

	// non-uniform scale: the circle becomes an ellipse
	ins2, _ := d.Insert("B", 0, 0, 0)
	ins2.Scale = []float64{1, 2, 1}
	es2, err := Explode(d, ins2)
	if err != nil {
		t.Fatal(err)
	}
	el, ok := es2[1].(*entity.Ellipse)
	if !ok || !near(el.Center, []float64{0, 2, 0}) || !near(el.MajorAxis, []float64{0, 1, 0}) || !cmpF64(el.Ratio, 0.5) || el.Layer().Name() != "A" {
		t.Errorf("ellipse: %+v", es2[1])
	}
	// mirrored block reference
	ins3, _ := d.Insert("B", 0, 0, 0)
	ins3.Scale = []float64{-1, 1, 1}
	es3, err := Explode(d, ins3)
	if err != nil {
		t.Fatal(err)
	}
	if l3 := es3[0].(*entity.Line); !near(l3.Start, []float64{0, 0, 0}) || !near(l3.End, []float64{-1, 0, 0}) {
		t.Errorf("mirrored line: %v %v", l3.Start, l3.End)
	}
	checkStructure(t, d)
}
