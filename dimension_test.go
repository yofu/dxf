package dxf

import (
	"strings"
	"testing"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

func dimBlockEntities(t *testing.T, d *drawing.Drawing, dim *entity.Dimension) entity.Entities {
	t.Helper()
	b, err := d.Block(dim.BlockName)
	if err != nil {
		t.Fatalf("dimension block %q: %v", dim.BlockName, err)
	}
	if !strings.HasPrefix(b.Name, "*D") || b.Flag&1 == 0 {
		t.Errorf("dimension block should be anonymous *Dn: %s flag %d", b.Name, b.Flag)
	}
	return b.Entities
}

func countTypes(es entity.Entities) map[string]int {
	m := map[string]int{}
	for _, e := range es {
		switch e.(type) {
		case *entity.Line:
			m["LINE"]++
		case *entity.Solid:
			m["SOLID"]++
		case *entity.MText:
			m["MTEXT"]++
		case *entity.Arc:
			m["ARC"]++
		}
	}
	return m
}

func blockText(t *testing.T, es entity.Entities) *entity.MText {
	t.Helper()
	for _, e := range es {
		if mt, ok := e.(*entity.MText); ok {
			return mt
		}
	}
	t.Fatal("no text in dimension block")
	return nil
}

func TestLinearDimension(t *testing.T) {
	d := NewDrawing()
	dim, err := d.LinearDimension([]float64{0, 0, 0}, []float64{100, 30, 0}, []float64{50, 50, 0}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if dim.Measurement != 100 || dim.DefPoint[0] != 100 || dim.DefPoint[1] != 50 {
		t.Errorf("horizontal: measurement %v defpoint %v", dim.Measurement, dim.DefPoint)
	}
	es := dimBlockEntities(t, d, dim)
	if c := countTypes(es); c["LINE"] != 3 || c["SOLID"] != 2 || c["MTEXT"] != 1 {
		t.Errorf("block entities: %v", c)
	}
	// extension line from (0, 0.625) to (0, 51.25)
	ext := es[0].(*entity.Line)
	if !cmpF64(ext.Start[1], 0.625) || !cmpF64(ext.End[1], 51.25) || ext.Start[0] != 0 {
		t.Errorf("extension line: %v - %v", ext.Start, ext.End)
	}
	mt := blockText(t, es)
	if mt.Value != "100.00" || !cmpF64(mt.Coord[0], 50) || !cmpF64(mt.Coord[1], 51.875) || mt.Rotation != 0 {
		t.Errorf("text %q at %v rotation %v", mt.Value, mt.Coord, mt.Rotation)
	}
	if !cmpF64(dim.TextMid[1], 51.875) {
		t.Errorf("TextMid: %v", dim.TextMid)
	}
	if es[0].(entity.Colored).ColorIndex() != entity.ColorByBlock {
		t.Error("block entities should be BYBLOCK")
	}

	v, _ := d.LinearDimension([]float64{0, 0, 0}, []float64{30, 50, 0}, []float64{-10, 0, 0}, 90)
	if v.Measurement != 50 || v.DefPoint[0] != -10 || v.DefPoint[1] != 50 {
		t.Errorf("vertical: measurement %v defpoint %v", v.Measurement, v.DefPoint)
	}
	if mt := blockText(t, dimBlockEntities(t, d, v)); mt.Rotation != 90 || !cmpF64(mt.Coord[0], -11.875) {
		t.Errorf("vertical text at %v rotation %v", mt.Coord, mt.Rotation)
	}
	if v.BlockName == dim.BlockName {
		t.Error("each dimension should have its own block")
	}
}

func TestAlignedRadialAngularDimensions(t *testing.T) {
	d := NewDrawing()
	a, err := d.AlignedDimension([]float64{0, 0, 0}, []float64{30, 40, 0}, []float64{-8, 6, 0})
	if err != nil {
		t.Fatal(err)
	}
	if !cmpF64(a.Measurement, 50) || !cmpF64(a.DefPoint[0], 22) || !cmpF64(a.DefPoint[1], 46) {
		t.Errorf("aligned: %v %v", a.Measurement, a.DefPoint)
	}
	if _, err := d.AlignedDimension([]float64{1, 1, 0}, []float64{1, 1, 0}, []float64{0, 0, 0}); err == nil {
		t.Error("aligned dimension of the same points should be an error")
	}

	r, _ := d.RadiusDimension([]float64{0, 0, 0}, []float64{6, 8, 0})
	if r.Measurement != 10 || blockText(t, dimBlockEntities(t, d, r)).Value != "R10.00" {
		t.Errorf("radius: %v", r.Measurement)
	}
	if c := countTypes(dimBlockEntities(t, d, r)); c["SOLID"] != 1 || c["LINE"] != 1 {
		t.Errorf("radius block: %v", c)
	}
	dia, _ := d.DiameterDimension([]float64{0, 0, 0}, []float64{10, 0, 0})
	if dia.Measurement != 20 || dia.DefPoint[0] != -10 || blockText(t, dimBlockEntities(t, d, dia)).Value != "∅20.00" {
		t.Errorf("diameter: %v %v", dia.Measurement, dia.DefPoint)
	}
	if _, err := d.RadiusDimension([]float64{1, 1, 0}, []float64{1, 1, 0}); err == nil {
		t.Error("zero radius should be an error")
	}

	ang, _ := d.AngularDimension([]float64{0, 0, 0}, []float64{10, 0, 0}, []float64{0, 10, 0}, []float64{5, 5, 0})
	if !cmpF64(ang.Measurement, 90) || blockText(t, dimBlockEntities(t, d, ang)).Value != "90.00°" {
		t.Errorf("angular: %v", ang.Measurement)
	}
	if c := countTypes(dimBlockEntities(t, d, ang)); c["ARC"] != 1 || c["SOLID"] != 2 {
		t.Errorf("angular block: %v", c)
	}
	reflex, _ := d.AngularDimension([]float64{0, 0, 0}, []float64{10, 0, 0}, []float64{0, 10, 0}, []float64{-5, -5, 0})
	if !cmpF64(reflex.Measurement, 270) {
		t.Errorf("reflex angle: %v", reflex.Measurement)
	}
}

func TestDimensionTextAndStyle(t *testing.T) {
	d := NewDrawing()
	st, err := d.DimStyle("standard")
	if err != nil {
		t.Fatal(err)
	}
	st.Decimals = 0
	st.Scale = 10
	dim := entity.NewDimension(entity.DIM_LINEAR)
	dim.DefPoint2, dim.DefPoint3, dim.DefPoint = []float64{0, 0, 0}, []float64{1500, 0, 0}, []float64{0, 500, 0}
	dim.Text = "<> mm"
	if err := d.RegenerateDimension(dim); err != nil {
		t.Fatal(err)
	}
	d.AddEntity(dim)
	mt := blockText(t, dimBlockEntities(t, d, dim))
	if mt.Value != "1500 mm" || mt.Height != 25 {
		t.Errorf("text %q height %v", mt.Value, mt.Height)
	}

	// regenerate after moving a point: the same block is reused
	name := dim.BlockName
	dim.DefPoint3 = []float64{2000, 0, 0}
	if err := d.RegenerateDimension(dim); err != nil {
		t.Fatal(err)
	}
	if dim.BlockName != name || dim.Measurement != 2000 || blockText(t, dimBlockEntities(t, d, dim)).Value != "2000 mm" {
		t.Errorf("regenerate: %s %v", dim.BlockName, dim.Measurement)
	}
	ord := entity.NewDimension(entity.DIM_ORDINATE)
	if err := d.RegenerateDimension(ord); err == nil {
		t.Error("ordinate dimension generation is not supported")
	}

	// DIMSTYLE values round trip
	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	st2, _ := d2.DimStyle("Standard")
	if st2.Decimals != 0 || st2.Scale != 10 || st2.ArrowSize != 2.5 || !st2.TextAbove {
		t.Errorf("dimstyle round trip: %+v", st2)
	}
}

func TestDimensionRoundTrip(t *testing.T) {
	d := NewDrawing()
	d.LinearDimension([]float64{0, 0, 0}, []float64{100, 0, 0}, []float64{50, 20, 0}, 0)
	d.AlignedDimension([]float64{0, 0, 0}, []float64{30, 40, 0}, []float64{-8, 6, 0})
	d.RadiusDimension([]float64{0, 0, 0}, []float64{6, 8, 0})
	d.DiameterDimension([]float64{0, 0, 0}, []float64{10, 0, 0})
	d.AngularDimension([]float64{0, 0, 0}, []float64{10, 0, 0}, []float64{0, 10, 0}, []float64{5, 5, 0})
	checkStructure(t, d)

	out := writeString(t, d)
	d2, err := FromStringData(out)
	if err != nil {
		t.Fatal(err)
	}
	checkStructure(t, d2)
	es, es2 := d.Entities(), d2.Entities()
	if len(es2) != len(es) {
		t.Fatalf("want %d dimensions, got %d", len(es), len(es2))
	}
	for i := range es {
		a, b := es[i].(*entity.Dimension), es2[i].(*entity.Dimension)
		if a.Type != b.Type || a.Flags != b.Flags || a.BlockName != b.BlockName || !cmpF64(a.Measurement, b.Measurement) || a.Style != b.Style {
			t.Errorf("dimension %d: %+v, want %+v", i, b, a)
		}
		for j, pts := range [][2][]float64{{a.DefPoint, b.DefPoint}, {a.TextMid, b.TextMid}, {a.DefPoint2, b.DefPoint2}, {a.DefPoint3, b.DefPoint3}, {a.DefPoint4, b.DefPoint4}} {
			for k := 0; k < 3; k++ {
				if a.Type == entity.DIM_RADIUS && j > 1 && j < 4 {
					continue // 13/14 are not written for radius
				}
				if !cmpF64(pts[0][k], pts[1][k]) {
					t.Errorf("dimension %d point %d: %v, want %v", i, j, pts[1], pts[0])
				}
			}
		}
		if len(dimBlockEntities(t, d2, b)) != len(dimBlockEntities(t, d, a)) {
			t.Errorf("dimension %d: block entities differ", i)
		}
	}
	if !strings.Contains(out, "AcDbRotatedDimension") || !strings.Contains(out, "AcDb3PointAngularDimension") ||
		!strings.Contains(out, "AcDbRadialDimension") || !strings.Contains(out, "AcDbDiametricDimension") {
		t.Error("subclass markers missing")
	}

	d.SetExt()
	if h := d.Header(); h.ExtMax[1] < 20 || h.ExtMin[0] > -8 {
		t.Errorf("extents should include dimension graphics: %v - %v", h.ExtMin, h.ExtMax)
	}
}

func TestReadDimensionAndSolid(t *testing.T) {
	src := entitiesSection("0\nDIMENSION\n5\n30\n8\n0\n100\nAcDbEntity\n100\nAcDbDimension\n2\n*D5\n10\n100\n20\n20\n30\n0\n11\n50\n21\n21.875\n31\n0\n70\n32\n71\n5\n42\n100\n3\nISO-25\n" +
		"100\nAcDbAlignedDimension\n13\n0\n23\n0\n33\n0\n14\n100\n24\n0\n34\n0\n50\n0\n100\nAcDbRotatedDimension\n" +
		"0\nSOLID\n8\n0\n10\n0\n20\n0\n11\n1\n21\n0\n12\n0\n22\n1\n")
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	dim := d.Entities()[0].(*entity.Dimension)
	if dim.Type != entity.DIM_LINEAR || dim.Flags != entity.DIM_BLOCK_OWNED || dim.BlockName != "*D5" || dim.Style != "ISO-25" ||
		dim.DefPoint[0] != 100 || dim.TextMid[1] != 21.875 || dim.DefPoint3[0] != 100 || dim.Measurement != 100 || dim.Attachment != 5 {
		t.Errorf("dimension: %+v", dim)
	}
	s := d.Entities()[1].(*entity.Solid)
	if s.Points[1][0] != 1 || s.Points[2][1] != 1 || s.Points[3][1] != 1 {
		t.Errorf("solid (triangle: 4th = 3rd): %v", s.Points)
	}
}
