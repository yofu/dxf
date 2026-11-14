package dxf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yofu/dxf/block"
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// Regression tests checked against the DXF reference.

func entitiesSection(ents string) string {
	return "0\nSECTION\n2\nENTITIES\n" + ents + "0\nENDSEC\n0\nEOF\n"
}

func writeString(t *testing.T, d *drawing.Drawing) string {
	t.Helper()
	var b bytes.Buffer
	if _, err := d.WriteTo(&b); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// groupValue returns the value following the first "\n<code>\n" in s.
func groupValue(s, code string) (string, bool) {
	key := "\n" + code + "\n"
	i := strings.Index(s, key)
	if i < 0 {
		return "", false
	}
	v := s[i+len(key):]
	return v[:strings.Index(v, "\n")], true
}

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

func TestUnsupportedEntitiesAreSkipped(t *testing.T) {
	src := entitiesSection("0\nLINE\n8\n0\n10\n0\n20\n0\n11\n1\n21\n1\n" +
		"0\nINSERT\n8\n0\n2\nB\n10\n0\n20\n0\n" +
		"0\nPOINT\n8\n0\n10\n3\n20\n4\n")
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(d.Entities()); n != 2 {
		t.Errorf("want 2 entities (LINE, POINT), got %d", n)
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
	if len(p.Vertices) != 2 || p.Flag != 1 || p.Vertices[0].Buldge != 1 {
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
	if n := len(d2.Sections[drawing.BLOCKS].(block.Blocks)); n != 4 {
		t.Errorf("round trip: want 4 blocks, got %d", n)
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
