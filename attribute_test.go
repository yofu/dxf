package dxf

import (
	"strings"
	"testing"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// titleBlock defines a block with attributes and returns the drawing.
func titleBlock(t *testing.T) *drawing.Drawing {
	t.Helper()
	d := NewDrawing()
	d.AddLayer("注記", color.Green, DefaultLineType, false)
	if _, err := d.BeginBlock("TITLE", 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	d.Line(0, 0, 0, 100, 0, 0)
	name, err := d.AttributeDefinition("NAME", "図面名", "未定", 10, 5, 0, 3)
	if err != nil {
		t.Fatal(err)
	}
	name.Anchor(entity.LEFT_BOTTOM)
	date, _ := d.AttributeDefinition("DATE", "日付", "2026-01-01", 60, 5, 0, 2.5)
	date.FieldLength = 10
	scale, _ := d.AttributeDefinition("SCALE", "", "1:100", 90, 5, 0, 2.5)
	scale.Flags = entity.ATTRIBUTE_CONSTANT
	d.ChangeLayer("注記")
	d.AttributeDefinition("CHECK", "検図", "", 90, 1, 0, 2)
	d.ChangeLayer("0")
	d.EndBlock()
	return d
}

func TestAttributeDefinitionAndInsert(t *testing.T) {
	d := titleBlock(t)
	if _, err := d.AttributeDefinition("BAD TAG", "", "", 0, 0, 0, 1); err == nil {
		t.Error("tag with a space should be an error")
	}

	ins, err := d.Insert("TITLE", 1000, 500, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(ins.Attributes) != 3 {
		t.Fatalf("want 3 attributes (constant SCALE excluded), got %d", len(ins.Attributes))
	}
	a := ins.Attribute("name")
	if a == nil || a.Value != "未定" || a.Coord1[0] != 1010 || a.Coord1[1] != 505 || a.Height != 3 || a.VerticalFlag != 1 {
		t.Fatalf("NAME attribute: %+v", a)
	}
	if !ins.SetAttribute("NAME", "1階平面図") || ins.SetAttribute("NONE", "x") {
		t.Error("SetAttribute")
	}
	if ins.Attribute("CHECK").Layer().Name() != "注記" || ins.Attribute("DATE").Layer().Name() != "0" {
		t.Error("attribute layers: layer 0 definitions take the insert's layer")
	}

	// scale and rotate, then sync: the value is kept
	ins.Scale = []float64{2, 2, 1}
	ins.Rotation = 90
	if err := d.SyncAttributes(ins); err != nil {
		t.Fatal(err)
	}
	a = ins.Attribute("NAME")
	if a.Value != "1階平面図" || !cmpF64(a.Coord1[0], 990) || !cmpF64(a.Coord1[1], 520) || a.Height != 6 || a.Rotation != 90 {
		t.Errorf("synced NAME attribute: %v height %v rotation %v value %q", a.Coord1, a.Height, a.Rotation, a.Value)
	}

	s := ins.String()
	if !strings.Contains(s, "AcDbBlockReference\n66\n1\n2\nTITLE\n") {
		t.Error("INSERT with attributes should have 66 = 1")
	}
	if strings.Count(s, "\n0\nATTRIB\n") != 3 || !strings.Contains(s, "\n0\nSEQEND\n") {
		t.Error("ATTRIBs and SEQEND should follow the INSERT")
	}
}

func TestAttributeRoundTrip(t *testing.T) {
	d := titleBlock(t)
	ins, _ := d.Insert("TITLE", 1000, 500, 0)
	ins.SetAttribute("NAME", "配置図")
	ins.SetColor(color.Red)
	d.Insert("TITLE", 0, 0, 0)
	d.Line(0, 0, 0, 1, 1, 0) // an entity after INSERT/SEQEND
	checkStructure(t, d)

	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	checkStructure(t, d2)
	es := d2.Entities()
	if len(es) != 3 {
		t.Fatalf("want 2 INSERTs and a LINE, got %d entities", len(es))
	}
	g := es[0].(*entity.Insert)
	if len(g.Attributes) != 3 || g.Attribute("NAME").Value != "配置図" || g.Attribute("DATE").FieldLength != 10 ||
		g.Attribute("NAME").VerticalFlag != 1 || g.ColorIndex() != int(color.Red) {
		t.Errorf("insert round trip: %+v", g.Attributes)
	}
	if v := es[1].(*entity.Insert).Attribute("NAME").Value; v != "未定" {
		t.Errorf("second insert default value: %q", v)
	}
	if _, ok := es[2].(*entity.Line); !ok {
		t.Errorf("third entity should be LINE: %T", es[2])
	}

	b, _ := d2.Block("TITLE")
	defs := 0
	for _, e := range b.Entities {
		if def, ok := e.(*entity.AttributeDefinition); ok {
			defs++
			if def.Tag == "NAME" && (def.Prompt != "図面名" || def.Value != "未定" || def.VerticalFlag != 1) {
				t.Errorf("ATTDEF round trip: %+v", def)
			}
			if def.Tag == "SCALE" && !def.IsConstant() {
				t.Error("SCALE should be constant")
			}
		}
	}
	if defs != 4 {
		t.Errorf("want 4 ATTDEFs in the block, got %d", defs)
	}
}

func TestReadAttributesFromFile(t *testing.T) {
	// AutoCAD style: INSERT (66=1), ATTRIBs, SEQEND
	src := entitiesSection("0\nINSERT\n8\n0\n66\n1\n2\nB\n10\n1\n20\n2\n" +
		"0\nATTRIB\n8\n0\n100\nAcDbText\n10\n1\n20\n2\n40\n2.5\n1\nVALUE\n100\nAcDbAttribute\n2\nTAG\n70\n0\n72\n1\n11\n3\n21\n4\n74\n2\n" +
		"0\nATTRIB\n8\n0\n10\n1\n20\n2\n40\n2.5\n1\n値\n2\nTAG2\n70\n1\n" +
		"0\nSEQEND\n8\n0\n" +
		"0\nATTRIB\n8\n0\n2\nORPHAN\n" + // without INSERT: skipped
		"0\nPOINT\n8\n0\n10\n3\n20\n4\n")
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Entities()) != 2 {
		t.Fatalf("want INSERT and POINT, got %d entities", len(d.Entities()))
	}
	ins := d.Entities()[0].(*entity.Insert)
	a := ins.Attribute("TAG")
	if len(ins.Attributes) != 2 || a == nil || a.Value != "VALUE" || a.HorizontalFlag != 1 || a.VerticalFlag != 2 || a.Coord2[0] != 3 {
		t.Fatalf("attributes: %+v", ins.Attributes)
	}
	if b := ins.Attribute("TAG2"); b.Value != "値" || b.Flags != entity.ATTRIBUTE_INVISIBLE {
		t.Errorf("TAG2: %+v", b)
	}
}

func TestInsertExtentsWithAttributes(t *testing.T) {
	d := titleBlock(t)
	d.Insert("TITLE", 0, 0, 0)
	d.SetExt()
	h := d.Header()
	// the line (0..100, 0) and attributes (text above y = 1..5)
	if h.ExtMin[0] != 0 || h.ExtMax[0] < 100 || h.ExtMin[1] != 0 || h.ExtMax[1] < 5 {
		t.Errorf("extents: %v - %v", h.ExtMin, h.ExtMax)
	}
}
