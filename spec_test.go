package dxf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yofu/dxf/block"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

func TestLwPolylineBulgeBelongsToPrecedingVertex(t *testing.T) {
	// In LWPOLYLINE, 42 follows the 10/20 of the vertex it belongs to.
	src := entitiesSection("0\nLWPOLYLINE\n8\n0\n90\n3\n70\n0\n" +
		"10\n0\n20\n0\n42\n1\n" +
		"10\n2\n20\n0\n" +
		"10\n2\n20\n2\n42\n0.5\n")
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	lw := d.Entities()[0].(*entity.LwPolyline)
	want := []float64{1, 0, 0.5}
	for i := range want {
		if !cmpF64(lw.Bulges[i], want[i]) {
			t.Fatalf("bulges: want %v got %v", want, lw.Bulges)
		}
	}

	// round trip
	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	lw2 := d2.Entities()[0].(*entity.LwPolyline)
	for i := range want {
		if !cmpF64(lw2.Bulges[i], want[i]) {
			t.Fatalf("round trip bulges: want %v got %v", want, lw2.Bulges)
		}
	}
}

func TestLwPolylineClosedFlagIsBitCoded(t *testing.T) {
	d, err := FromStringData(entitiesSection("0\nLWPOLYLINE\n8\n0\n90\n2\n70\n129\n10\n0\n20\n0\n10\n1\n20\n0\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !d.Entities()[0].(*entity.LwPolyline).Closed {
		t.Error("70 = 129 (closed + plinegen) should be closed")
	}
}

func TestBulgeCenter(t *testing.T) {
	cases := []struct {
		name   string
		bulge  float64
		center []float64
		radius float64
	}{
		{"ccw 90deg", 0.41421356237, []float64{1, 1}, 1.41421356237},
		{"cw 90deg", -0.41421356237, []float64{1, -1}, 1.41421356237},
		{"ccw 270deg", 2.41421356237, []float64{1, -1}, 1.41421356237},
		{"semicircle", 1, []float64{1, 0}, 1},
	}
	for _, c := range cases {
		lw := entity.NewLwPolyline(2)
		lw.Vertices = [][]float64{{0, 0}, {2, 0}}
		lw.Bulges = []float64{c.bulge, 0}
		center, r := lw.Bulge(1)
		if center == nil || !cmpF64(center[0], c.center[0]) || !cmpF64(center[1], c.center[1]) || !cmpF64(r, c.radius) {
			t.Errorf("%s: want center %v r %v, got %v r %v", c.name, c.center, c.radius, center, r)
		}
	}
}

func TestUnsupportedEntitiesArePreserved(t *testing.T) {
	src := entitiesSection("0\nLINE\n8\n0\n10\n0\n20\n0\n11\n1\n21\n1\n" +
		"0\nACAD_PROXY_ENTITY\n8\n0\n90\n498\n91\n0\n" +
		"0\nPOINT\n8\n0\n10\n3\n20\n4\n")
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(d.Entities()); n != 3 {
		t.Fatalf("want 3 entities (LINE, ACAD_PROXY_ENTITY, POINT), got %d", n)
	}
	if u, ok := d.Entities()[1].(*entity.Unknown); !ok || u.TypeName() != "ACAD_PROXY_ENTITY" || len(u.Data) != 2 {
		t.Errorf("unknown entity: %+v", d.Entities()[1])
	}
}

func TestPolylineWithVertices(t *testing.T) {
	src := entitiesSection("0\nPOLYLINE\n8\n0\n66\n1\n70\n1\n" +
		"0\nVERTEX\n8\n0\n10\n0\n20\n0\n42\n1\n" +
		"0\nVERTEX\n8\n0\n10\n2\n20\n0\n" +
		"0\nSEQEND\n8\n0\n" +
		"0\nPOINT\n8\n0\n10\n3\n20\n4\n")
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	es := d.Entities()
	if len(es) != 2 {
		t.Fatalf("want 2 entities (POLYLINE, POINT), got %d", len(es))
	}
	p, ok := es[0].(*entity.Polyline)
	if !ok {
		t.Fatalf("want *entity.Polyline, got %T", es[0])
	}
	if len(p.Vertices) != 2 || p.Flag != 1 || p.Vertices[0].Bulge != 1 {
		t.Errorf("unexpected polyline: flag %d, %d vertices", p.Flag, len(p.Vertices))
	}
	out := writeString(t, d)
	if !strings.Contains(out, "AcDb2dPolyline") || !strings.Contains(out, "AcDb2dVertex") {
		t.Error("2D polyline should be written with AcDb2dPolyline / AcDb2dVertex")
	}

	// polylines written by this package can be read back
	d = NewDrawing()
	d.Polyline(false, []float64{0, 0, 0}, []float64{1, 1, 1})
	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	if p := d2.Entities()[0].(*entity.Polyline); len(p.Vertices) != 2 {
		t.Errorf("want 2 vertices, got %d", len(p.Vertices))
	}
}

func TestCommentsAndCase(t *testing.T) {
	src := "999\nfile comment\n" + entitiesSection("999\ncomment\n0\npoint\n8\n0\n10\n1\n20\n2\n")
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Entities()) != 1 {
		t.Errorf("want 1 entity, got %d", len(d.Entities()))
	}
	if _, err := FromStringData("0\nENDSEC\n0\nEOF\n"); err == nil {
		t.Error("ENDSEC without SECTION should be an error")
	}
}

func TestBlocksAreKept(t *testing.T) {
	src := "0\nSECTION\n2\nBLOCKS\n" +
		"0\nBLOCK\n8\n0\n2\nMYBLK\n70\n0\n10\n5\n20\n6\n30\n0\n3\nMYBLK\n1\n\n4\ndesc\n" +
		"0\nLINE\n8\n0\n10\n100\n20\n100\n11\n200\n21\n200\n" +
		"0\nENDBLK\n8\n0\n" +
		"0\nENDSEC\n0\nEOF\n"
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	var b *block.Block
	for _, bl := range d.Sections[drawing.BLOCKS].(block.Blocks) {
		if bl.Name == "MYBLK" {
			b = bl
		}
	}
	if b == nil {
		t.Fatal("block MYBLK not found")
	}
	if b.Coord[0] != 5 || b.Coord[1] != 6 {
		t.Errorf("base point: want [5 6 0], got %v", b.Coord)
	}
	if b.Description != "desc" || b.XrefPath != "" {
		t.Errorf("description %q xref %q", b.Description, b.XrefPath)
	}
	if len(b.Entities) != 1 {
		t.Fatalf("want 1 entity in block, got %d", len(b.Entities))
	}
	out := writeString(t, d)
	if !strings.Contains(out, "\n4\ndesc\n") {
		t.Error("block description should be written with code 4")
	}
	d2, err := FromStringData(out)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(d2.Sections[drawing.BLOCKS].(block.Blocks)); n != 3 { // *Model_Space, *Paper_Space, MYBLK
		t.Errorf("round trip: want 3 blocks, got %d", n)
	}
}

func TestTextAlignmentPoint(t *testing.T) {
	for _, pos := range []int{entity.CENTER_BASE, entity.LEFT_TOP, entity.CENTER_CENTER} {
		tx := entity.NewText()
		tx.Coord1 = []float64{10, 20, 0}
		tx.Coord2 = []float64{30, 40, 0}
		tx.Anchor(pos)
		v, ok := groupValue(tx.String(), "11")
		if !ok || !strings.HasPrefix(v, "30") {
			t.Errorf("anchor %d: 11 should be Coord2.x (30), got %q", pos, v)
		}
	}
	tx := entity.NewText()
	if _, ok := groupValue(tx.String(), "11"); ok {
		t.Error("baseline/left text should not have code 11")
	}

	// Drawing.Text keeps the given point as alignment point
	d := NewDrawing()
	txt, _ := d.Text("abc", 10, 20, 0, 1)
	txt.Anchor(entity.CENTER_CENTER)
	if v, _ := groupValue(txt.String(), "11"); !strings.HasPrefix(v, "10") {
		t.Errorf("Drawing.Text: 11 should be 10, got %q", v)
	}
}

func TestTextWidthFactorAndStyle(t *testing.T) {
	if v, _ := groupValue(entity.NewText().String(), "41"); !strings.HasPrefix(v, "1.") {
		t.Errorf("default width factor should be 1, got %s", v)
	}
	src := "0\nSECTION\n2\nTABLES\n0\nTABLE\n2\nSTYLE\n70\n1\n0\nSTYLE\n2\nMYSTYLE\n70\n0\n3\nromans.shx\n0\nENDTAB\n0\nENDSEC\n" +
		entitiesSection("0\nTEXT\n8\n0\n10\n0\n20\n0\n40\n1\n1\nhi\n41\n0.8\n51\n15\n7\nmystyle\n")
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	tx := d.Entities()[0].(*entity.Text)
	if tx.Style.Name() != "MYSTYLE" || tx.WidthFactor != 0.8 || tx.ObliqueAngle != 15 {
		t.Errorf("style %s width %v oblique %v", tx.Style.Name(), tx.WidthFactor, tx.ObliqueAngle)
	}
}

func TestSplineKnots(t *testing.T) {
	d := NewDrawing()
	for n := 2; n <= 6; n++ {
		p := make([][]float64, n)
		for i := range p {
			p[i] = []float64{float64(i), 0, 0}
		}
		s, err := d.Spline(p)
		if err != nil {
			t.Fatal(err)
		}
		if len(s.Knots) != n+s.Degree+1 {
			t.Errorf("%d control points, degree %d: want %d knots, got %d", n, s.Degree, n+s.Degree+1, len(s.Knots))
		}
	}
}

func TestExtentsDoNotIncludeOrigin(t *testing.T) {
	d := NewDrawing()
	d.LwPolyline(false, []float64{10, 10}, []float64{20, 20})
	d.Polyline(false, []float64{10, 10, 10}, []float64{20, 20, 20})
	d.ThreeDFace([][]float64{{10, 10, 10}, {20, 10, 10}, {20, 20, 10}})
	d.SetExt()
	if h := d.Header(); h.ExtMin[0] != 10 || h.ExtMin[1] != 10 {
		t.Errorf("EXTMIN: want [10 10 ...], got %v", h.ExtMin)
	}
}

func TestTablesRobustness(t *testing.T) {
	// leading spaces in 49, layer referencing an undefined linetype
	src := "0\nSECTION\n2\nTABLES\n" +
		"0\nTABLE\n2\nLTYPE\n70\n1\n0\nLTYPE\n2\nDASH\n70\n0\n3\n__ __\n72\n65\n73\n2\n40\n0.75\n49\n 0.5\n49\n-0.25\n0\nENDTAB\n" +
		"0\nTABLE\n2\nLAYER\n70\n1\n0\nLAYER\n2\nA\n70\n0\n62\n1\n6\nUNDEFINED\n0\nENDTAB\n" +
		"0\nENDSEC\n0\nEOF\n"
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	if d.Layers["A"].LineType.Name() != "UNDEFINED" {
		t.Error("linetype reference should be kept")
	}
}

func TestAddToGroup(t *testing.T) {
	d := NewDrawing()
	l, _ := d.Line(0, 0, 0, 1, 1, 0)
	d.Group("G", "")
	if err := d.AddToGroup("G", l); err != nil {
		t.Error(err)
	}
	if err := d.AddToGroup("NONE", l); err == nil {
		t.Error("want error for unknown group")
	}
}

func TestDrawingsDoNotShareTableEntries(t *testing.T) {
	d1 := NewDrawing()
	d2 := NewDrawing()
	if d1.Layers["0"] == d2.Layers["0"] || d1.CurrentStyle == d2.CurrentStyle {
		t.Fatal("layer 0 / style Standard must not be shared between drawings")
	}
	lt1, _ := d1.LineType("Continuous")
	lt2, _ := d2.LineType("Continuous")
	if lt1 == lt2 {
		t.Fatal("linetypes must not be shared between drawings")
	}
	// layer 0 of the first drawing must point to its own plot style placeholder
	s := writeString(t, d1)
	ph, _ := groupValue(s, "0\nACDBPLACEHOLDER\n5")
	ps, _ := groupValue(s, "390")
	if ph == "" || ph != ps {
		t.Errorf("layer 0 390 should be %q (placeholder), got %q", ph, ps)
	}
	d2.Text("x", 0, 0, 0, 42)
	if d1.CurrentStyle.LastHeightUsed == 42 {
		t.Error("style state leaked between drawings")
	}
}

func TestLayerOff(t *testing.T) {
	src := "0\nSECTION\n2\nTABLES\n0\nTABLE\n2\nLAYER\n70\n1\n" +
		"0\nLAYER\n2\nOFF\n70\n0\n62\n-7\n6\nCONTINUOUS\n0\nENDTAB\n0\nENDSEC\n0\nEOF\n"
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	l := d.Layers["OFF"]
	if l.Color != 7 || l.IsOn() {
		t.Fatalf("want color 7 and off, got color %d on=%v", l.Color, l.IsOn())
	}
	if !strings.Contains(writeString(t, d), "\n2\nOFF\n70\n0\n62\n-7\n") {
		t.Error("off layer should be written with negative color")
	}
	l.On()
	if !strings.Contains(writeString(t, d), "\n2\nOFF\n70\n0\n62\n7\n") {
		t.Error("on layer should be written with positive color")
	}
}

func TestConcurrentDrawings(t *testing.T) {
	done := make(chan struct{})
	for i := 0; i < 4; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			d := NewDrawing()
			d.Line(0, 0, 0, 1, 1, 0)
			d.Text("x", 0, 0, 0, 1)
			var b bytes.Buffer
			d.WriteTo(&b)
		}()
	}
	for i := 0; i < 4; i++ {
		<-done
	}
}

func TestEntityColor(t *testing.T) {
	d := NewDrawing()
	l, _ := d.Line(0, 0, 0, 1, 1, 0)
	c, _ := d.Circle(0, 0, 0, 1)
	a, _ := d.Arc(0, 0, 0, 1, 0, 90)
	tx, _ := d.Text("x", 0, 0, 0, 1)
	p, _ := d.Point(0, 0, 0)
	f, _ := d.ThreeDFace([][]float64{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}})
	lw, _ := d.LwPolyline(false, []float64{0, 0}, []float64{1, 0})
	pl, _ := d.Polyline(false, []float64{0, 0, 0}, []float64{1, 1, 1})
	sp, _ := d.Spline([][]float64{{0, 0, 0}, {1, 1, 0}, {2, 0, 0}, {3, 1, 0}})
	want := []int{1, 2, 3, 4, 5, 6, 7, 0, 256}
	ents := []entity.Colored{l, c, a, tx, p, f, lw, pl, sp}
	for i, e := range ents {
		switch want[i] {
		case entity.ColorByBlock:
			e.SetColorByBlock()
		case entity.ColorByLayer:
			e.SetColorByLayer()
		default:
			e.SetColor(color.ColorNumber(want[i]))
		}
	}
	pl.Vertices[1].SetColor(color.Red)

	out := writeString(t, d)
	// 8 entities + 1 vertex + layers + $CECOLOR in the header
	if strings.Count(out, "\n62\n") != 8+1+len(d.Layers)+1 {
		t.Errorf("62 should be written for non-BYLAYER entities only: got %d", strings.Count(out, "\n62\n"))
	}
	d2, err := FromStringData(out)
	if err != nil {
		t.Fatal(err)
	}
	es := d2.Entities()
	if len(es) != len(want) {
		t.Fatalf("want %d entities, got %d", len(want), len(es))
	}
	for i, e := range es {
		if got := e.(entity.Colored).ColorIndex(); got != want[i] {
			t.Errorf("%T: want color %d, got %d", e, want[i], got)
		}
	}
	vs := es[7].(*entity.Polyline).Vertices
	if vs[0].ColorIndex() != entity.ColorByLayer || vs[1].ColorIndex() != int(color.Red) {
		t.Errorf("vertex colors: %d %d", vs[0].ColorIndex(), vs[1].ColorIndex())
	}
}

func TestEntityColorFromFile(t *testing.T) {
	src := entitiesSection("0\nLINE\n8\n0\n62\n-3\n10\n0\n20\n0\n11\n1\n21\n1\n" +
		"0\nPOINT\n8\n0\n10\n0\n20\n0\n") // negative: layer off
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	if c := d.Entities()[0].(entity.Colored).ColorIndex(); c != 3 {
		t.Errorf("want 3, got %d", c)
	}
	if !d.Entities()[1].(entity.Colored).IsColorByLayer() {
		t.Error("default color should be BYLAYER")
	}
}

func TestBlockAndInsert(t *testing.T) {
	d := NewDrawing()
	if _, err := d.BeginBlock("BOLT", 1, 1, 0); err != nil {
		t.Fatal(err)
	}
	d.Line(1, 1, 0, 2, 1, 0)
	d.Circle(1, 1, 0, 0.5)
	if err := d.EndBlock(); err != nil {
		t.Fatal(err)
	}
	if len(d.Entities()) != 0 {
		t.Fatalf("entities between BeginBlock/EndBlock must go to the block, got %d in ENTITIES", len(d.Entities()))
	}
	if _, err := d.AddBlock("bolt", 0, 0, 0); err == nil {
		t.Error("duplicate block name (case-insensitive) should be an error")
	}
	if _, err := d.Insert("NUT", 0, 0, 0); err == nil {
		t.Error("insert of undefined block should be an error")
	}
	if d.EndBlock() == nil {
		t.Error("EndBlock without BeginBlock should be an error")
	}

	ins, err := d.Insert("BOLT", 10, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	ins.Scale = []float64{2, 3, 1}
	ins.Rotation = 90
	ins.Columns, ins.Rows = 2, 3
	ins.ColumnSpacing, ins.RowSpacing = 5, 6
	ins.SetColor(color.Red)

	out := writeString(t, d)
	if !strings.Contains(out, "\n0\nBLOCK_RECORD\n") || !strings.Contains(out, "AcDbBlockTableRecord\n2\nBOLT\n") {
		t.Error("BLOCK_RECORD for BOLT should be written")
	}
	if !strings.Contains(out, "AcDbBlockReference\n2\nBOLT\n") {
		t.Error("INSERT should be written with AcDbBlockReference")
	}

	d2, err := FromStringData(out)
	if err != nil {
		t.Fatal(err)
	}
	b, err := d2.Block("BOLT")
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Entities) != 2 || b.Coord[0] != 1 || b.Coord[1] != 1 {
		t.Errorf("block: %d entities, base %v", len(b.Entities), b.Coord)
	}
	got := d2.Entities()[0].(*entity.Insert)
	if got.BlockName != "BOLT" || got.Coord[0] != 10 || got.Coord[1] != 20 ||
		got.Scale[0] != 2 || got.Scale[1] != 3 || got.Scale[2] != 1 || got.Rotation != 90 ||
		got.Columns != 2 || got.Rows != 3 || got.ColumnSpacing != 5 || got.RowSpacing != 6 ||
		got.ColorIndex() != int(color.Red) {
		t.Errorf("insert round trip: %+v", got)
	}
}

func TestInsertExtents(t *testing.T) {
	d := NewDrawing()
	d.BeginBlock("L", 0, 0, 0)
	d.Line(0, 0, 0, 1, 0, 0)
	d.EndBlock()
	d.BeginBlock("NESTED", 0, 0, 0)
	ins, _ := d.Insert("L", 0, 0, 0)
	ins.Scale[0] = 3
	d.EndBlock()

	// L: scale 2, rotate 90 deg at (10, 10) => (10,10)-(10,12)
	ins, _ = d.Insert("L", 10, 10, 0)
	ins.Scale[0] = 2
	ins.Rotation = 90
	d.SetExt()
	h := d.Header()
	if !cmpF64(h.ExtMin[0], 10) || !cmpF64(h.ExtMin[1], 10) || !cmpF64(h.ExtMax[0], 10) || !cmpF64(h.ExtMax[1], 12) {
		t.Errorf("rotated insert: got %v - %v", h.ExtMin, h.ExtMax)
	}

	// NESTED (L scaled by 3) at (-5, 0) => (-5,0)-(-2,0)
	d.Insert("NESTED", -5, 0, 0)
	d.SetExt()
	if !cmpF64(h.ExtMin[0], -5) || !cmpF64(h.ExtMax[1], 12) {
		t.Errorf("nested insert: got %v - %v", h.ExtMin, h.ExtMax)
	}

	// array: 3 columns with spacing 10 => x up to 20+1
	d = NewDrawing()
	d.BeginBlock("L", 0, 0, 0)
	d.Line(0, 0, 0, 1, 0, 0)
	d.EndBlock()
	ins, _ = d.Insert("L", 0, 0, 0)
	ins.Columns, ins.ColumnSpacing = 3, 10
	d.SetExt()
	if h := d.Header(); !cmpF64(h.ExtMax[0], 21) {
		t.Errorf("array insert: ExtMax %v", h.ExtMax)
	}
}

func TestInsertWithAttributesFromFile(t *testing.T) {
	// ATTRIB entities are attached to the INSERT; SEQEND ends them.
	src := entitiesSection("0\nINSERT\n8\n0\n66\n1\n2\nB\n10\n1\n20\n2\n" +
		"0\nATTRIB\n8\n0\n10\n1\n20\n2\n40\n1\n1\nVALUE\n2\nTAG\n70\n0\n" +
		"0\nSEQEND\n8\n0\n" +
		"0\nPOINT\n8\n0\n10\n3\n20\n4\n")
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Entities()) != 2 {
		t.Fatalf("want INSERT and POINT, got %d entities", len(d.Entities()))
	}
	if ins := d.Entities()[0].(*entity.Insert); ins.BlockName != "B" || ins.Coord[1] != 2 ||
		len(ins.Attributes) != 1 || ins.Attribute("TAG").Value != "VALUE" {
		t.Errorf("insert: %+v", ins)
	}
}
