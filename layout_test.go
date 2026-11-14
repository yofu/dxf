package dxf

import (
	"strings"
	"testing"

	"github.com/yofu/dxf/entity"
)

func TestGroupsRoundTrip(t *testing.T) {
	d := NewDrawing()
	l1, _ := d.Line(0, 0, 0, 1, 0, 0)
	l2, _ := d.Line(0, 0, 0, 0, 1, 0)
	c, _ := d.Circle(0, 0, 0, 1)
	d.Group("柱", "columns", l1, l2)
	g2, _ := d.Group("*A1", "", c)
	g2.Unnamed = true
	g2.SetSelectable(false)
	checkStructure(t, d)

	out := writeString(t, d)
	d2 := readPreserve(t, out)
	checkStructure(t, d2)
	if len(d2.Groups) != 2 {
		t.Fatalf("want 2 groups, got %d", len(d2.Groups))
	}
	g := d2.Groups["柱"]
	if g == nil || g.Description != "columns" || len(g.Entities()) != 2 || g.Unnamed || !g.Selectable() {
		t.Fatalf("group: %+v", g)
	}
	if g.Entities()[0] != d2.Entities()[0] || g.Entities()[1] != d2.Entities()[1] {
		t.Error("group entities should be the entities read")
	}
	a := d2.Groups["*A1"]
	if a == nil || !a.Unnamed || a.Selectable() || a.Entities()[0] != d2.Entities()[2] {
		t.Errorf("unnamed group: %+v", a)
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
}

func TestLayouts(t *testing.T) {
	d := NewDrawing()
	if err := d.RenameLayout("Layout1", "配置1"); err != nil {
		t.Fatal(err)
	}
	l, err := d.AddLayout("A3縦")
	if err != nil {
		t.Fatal(err)
	}
	if l.Block != "*Paper_Space0" || l.TabOrder != 2 {
		t.Errorf("layout: %+v", l)
	}
	if _, err := d.AddLayout("a3縦"); err == nil {
		t.Error("duplicate layout name should be an error")
	}
	if err := d.UseLayout("NONE"); err == nil {
		t.Error("unknown layout should be an error")
	}

	d.Line(0, 0, 0, 1000, 0, 0) // model space
	if err := d.UseLayout("A3縦"); err != nil {
		t.Fatal(err)
	}
	frame, _ := d.LwPolyline(true, []float64{10, 10}, []float64{287, 10}, []float64{287, 410}, []float64{10, 410})
	if _, err := d.Viewport([]float64{148.5, 210, 0}, 250, 380, []float64{500, 0}, 2000); err != nil {
		t.Fatal(err)
	}
	if err := d.EndBlock(); err == nil {
		t.Error("EndBlock while a layout is in use should be an error")
	}
	if _, err := d.BeginBlock("X", 0, 0, 0); err == nil {
		t.Error("BeginBlock while a layout is in use should be an error")
	}
	d.UseLayout("配置1")
	d.Text("active layout", 0, 0, 0, 5)
	d.UseLayout("Model")
	d.Circle(0, 0, 0, 10)

	b, _ := d.Block("*Paper_Space0")
	if len(b.Entities) != 3 || !frame.InPaperSpace() {
		t.Fatalf("layout block: %d entities (frame, paper space viewport, viewport)", len(b.Entities))
	}
	if v := b.Entities[1].(*entity.Viewport); v.Width != 420 || v.Height != 297 || v.Status != 1 {
		t.Errorf("paper space viewport of the layout: %+v", v)
	}
	if len(d.Entities()) != 3 || len(d.PaperSpaceEntities()) != 1 {
		t.Errorf("ENTITIES: %d, paper space %d", len(d.Entities()), len(d.PaperSpaceEntities()))
	}
	checkStructure(t, d)

	out := writeString(t, d)
	if !strings.Contains(out, "\n3\nA3縦\n") && !strings.Contains(out, "A3") {
		t.Error("layout name should be in the ACAD_LAYOUT dictionary")
	}
	d2 := readPreserve(t, out)
	checkStructure(t, d2)
	names := []string{}
	for _, l := range d2.Layouts() {
		names = append(names, l.Name+":"+l.Block)
	}
	if strings.Join(names, ",") != "Model:*Model_Space,配置1:*Paper_Space,A3縦:*Paper_Space0" {
		t.Errorf("layouts: %v", names)
	}
	b2, _ := d2.Block("*Paper_Space0")
	if len(b2.Entities) != 3 {
		t.Errorf("layout block round trip: %d entities", len(b2.Entities))
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
}

func TestReadObjects(t *testing.T) {
	// handles in the file differ from the ones this package writes
	src := "0\nSECTION\n2\nTABLES\n0\nTABLE\n2\nBLOCK_RECORD\n70\n3\n" +
		"0\nBLOCK_RECORD\n5\n1F\n2\n*Model_Space\n340\n22\n" +
		"0\nBLOCK_RECORD\n5\n58\n2\n*Paper_Space\n340\n59\n" +
		"0\nBLOCK_RECORD\n5\n5D\n2\n*Paper_Space0\n340\n5E\n0\nENDTAB\n0\nENDSEC\n" +
		"0\nSECTION\n2\nBLOCKS\n0\nBLOCK\n8\n0\n2\n*Paper_Space0\n70\n0\n10\n0\n20\n0\n0\nLINE\n5\n70\n67\n1\n8\n0\n10\n0\n20\n0\n11\n1\n21\n1\n0\nENDBLK\n8\n0\n0\nENDSEC\n" +
		"0\nSECTION\n2\nENTITIES\n0\nLINE\n5\n80\n8\n0\n10\n0\n20\n0\n11\n1\n21\n1\n0\nCIRCLE\n5\n81\n8\n0\n10\n0\n20\n0\n40\n1\n0\nENDSEC\n" +
		"0\nSECTION\n2\nOBJECTS\n" +
		"0\nDICTIONARY\n5\nC\n330\n0\n100\nAcDbDictionary\n3\nACAD_GROUP\n350\nD\n3\nACAD_LAYOUT\n350\n1A\n3\nACAD_MLINESTYLE\n350\n17\n" +
		"0\nDICTIONARY\n5\nD\n330\nC\n100\nAcDbDictionary\n3\nWALLS\n350\n90\n" +
		"0\nDICTIONARY\n5\n1A\n330\nC\n100\nAcDbDictionary\n3\nLayout1\n350\n59\n3\nLayout2\n350\n5E\n3\nModel\n350\n22\n" +
		"0\nLAYOUT\n5\n22\n330\n1A\n100\nAcDbPlotSettings\n1\n\n100\nAcDbLayout\n1\nModel\n70\n1\n71\n0\n10\n0\n20\n0\n11\n12\n21\n9\n330\n1F\n" +
		"0\nLAYOUT\n5\n59\n330\n1A\n100\nAcDbPlotSettings\n1\n\n100\nAcDbLayout\n1\nLayout1\n70\n1\n71\n1\n10\n0\n20\n0\n11\n420\n21\n297\n330\n58\n" +
		"0\nLAYOUT\n5\n5E\n330\n1A\n100\nAcDbPlotSettings\n1\n\n100\nAcDbLayout\n1\nLayout2\n70\n1\n71\n2\n10\n0\n20\n0\n11\n841\n21\n594\n330\n5D\n" +
		"0\nGROUP\n5\n90\n330\nD\n100\nAcDbGroup\n300\nall walls\n70\n0\n71\n1\n340\n81\n340\n80\n340\nFF\n" +
		"0\nENDSEC\n0\nEOF\n"
	d := readPreserve(t, src)
	ls := d.Layouts()
	if len(ls) != 3 || ls[2].Name != "Layout2" || ls[2].Block != "*Paper_Space0" || ls[2].TabOrder != 2 || ls[2].LimMax[0] != 841 || ls[0].LimMax[0] != 12 {
		t.Errorf("layouts: %+v %+v %+v", ls[0], ls[1], ls[2])
	}
	g := d.Groups["WALLS"]
	if g == nil || g.Description != "all walls" || len(g.Entities()) != 2 {
		t.Fatalf("group: %+v", g)
	}
	if _, ok := g.Entities()[0].(*entity.Circle); !ok {
		t.Error("group entities should keep their order")
	}
	checkStructure(t, d)
}
