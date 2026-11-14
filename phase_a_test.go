package dxf

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
)

func TestMoreEntities(t *testing.T) {
	d := NewDrawing()
	d.SetVersion(format.R2010)
	x, _ := d.XLine([]float64{1, 2, 0}, []float64{3, 4, 0})
	d.Ray([]float64{0, 0, 0}, []float64{0, 2, 0})
	d.Trace([][]float64{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}, {1, 1, 0}})
	shx := d.AddShapeFile("ltypeshp.shx")
	_ = shx
	sh, _ := d.Shape("TRACK1", []float64{5, 5, 0}, 2)
	sh.Rotation = 30
	tl, _ := d.Tolerance(`{\Fgdt;j}%%v0.1`, []float64{0, 10, 0})
	lt, _ := d.Light("L1", entity.LIGHT_SPOT, []float64{0, 0, 10})
	h, _ := d.Helix([]float64{10, 0, 0}, 2, 1, 3, true)
	m, _ := d.Mesh([][]float64{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 1}}, [][]int{{0, 1, 2}, {0, 2, 3}})
	m.Edges = [][2]int{{0, 1}}
	m.Creases = []float64{0.5}
	mp, _ := d.MPolygon([][]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}, [][]float64{{2, 2}, {4, 2}, {4, 4}})
	mp.FillColor = 3
	sat := []string{"400 0 1 0", "16 ASM-Body 7.0 24 0 ", "body $-1 -1 $-1 $1 $-1 $2 #", "End-of-ACIS-data"}
	b, _ := d.Body("3DSOLID", sat)
	arc := entity.NewDimension(entity.DIM_ARC)
	arc.DefPoint = []float64{0, 6, 0}
	arc.DefPoint2, arc.DefPoint3, arc.DefPoint4 = []float64{5, 0, 0}, []float64{0, 5, 0}, []float64{0, 0, 0}
	arc.ArcStart, arc.ArcEnd = 0, math.Pi/2
	arc.ArcLeader1, arc.ArcLeader2 = []float64{0, 0, 0}, []float64{0, 0, 0}
	arc.Measurement = 5 * math.Pi / 2
	d.AddEntity(arc)

	// helix curve: a point on the helix at a quarter turn
	if p := h.Controls[3]; math.Abs(p[0]-10) > 1e-9 || math.Abs(p[1]-2) > 1e-9 || math.Abs(p[2]-0.25) > 1e-9 {
		t.Errorf("helix control point: %v", p)
	}
	out := writeString(t, d)
	checkStructure(t, d)
	for _, s := range []string{
		"\n0\nXLINE\n", "\n100\nAcDbXline\n", "\n0\nRAY\n", "\n100\nAcDbRay\n", "\n0\nTRACE\n", "\n0\nSHAPE\n", "\n2\nTRACK1\n",
		"\n0\nTOLERANCE\n", "\n100\nAcDbFcf\n", "\n0\nLIGHT\n", "\n0\nHELIX\n", "\n100\nAcDbHelix\n",
		"\n0\nMESH\n", "\n100\nAcDbSubDMesh\n", "\n0\nMPOLYGON\n", "\n63\n3\n", "\n0\n3DSOLID\n", "\n100\nAcDbModelerGeometry\n",
		"\n0\nARC_DIMENSION\n", "\n100\nAcDbArcDimension\n",
		"\n1\nHELIX\n2\nAcDbHelix\n", "\n1\nMPOLYGON\n2\nAcDbMPolygon\n", "\n1\nARC_DIMENSION\n",
	} {
		if !strings.Contains(out, s) {
			t.Errorf("%q not found", s)
		}
	}
	d2 := readPreserve(t, out)
	es := d2.Entities()
	if rx := es[0].(*entity.XLine); !near(rx.Point, x.Point) || !near(rx.Direction, []float64{0.6, 0.8, 0}) || rx.IsRay() || !es[1].(*entity.XLine).IsRay() {
		t.Errorf("xline: %+v", rx)
	}
	if es[2].(*entity.Solid).Type != entity.TRACE {
		t.Error("trace")
	}
	if rs := es[3].(*entity.Shape); rs.Name != "TRACK1" || rs.Rotation != 30 || rs.Size != 2 {
		t.Errorf("shape: %+v", rs)
	}
	if rt := es[4].(*entity.Tolerance); rt.Content != tl.Content || !near(rt.Insert, tl.Insert) {
		t.Errorf("tolerance: %+v", rt)
	}
	if rl := es[5].(*entity.Light); rl.Name != "L1" || rl.LightType != lt.LightType || !near(rl.Location, lt.Location) {
		t.Errorf("light: %+v", rl)
	}
	if rh := es[6].(*entity.Helix); rh.Turns != 3 || rh.Radius != 2 || !near(rh.AxisBase, []float64{10, 0, 0}) || len(rh.Controls) != len(h.Controls) {
		t.Errorf("helix: %+v", rh)
	}
	if rm := es[7].(*entity.Mesh); !reflect.DeepEqual(rm.Faces, m.Faces) || len(rm.Vertices) != 4 || !reflect.DeepEqual(rm.Edges, m.Edges) || rm.Creases[0] != 0.5 {
		t.Errorf("mesh: %+v", rm)
	}
	if rp := es[8].(*entity.MPolygon); len(rp.Boundaries) != 2 || len(rp.Boundaries[1].Vertices) != 3 || rp.FillColor != 3 || !rp.Solid {
		t.Errorf("mpolygon: %+v", rp)
	}
	if rb := es[9].(*entity.Body); rb.TypeName() != "3DSOLID" || !reflect.DeepEqual(rb.SAT, b.SAT) {
		t.Errorf("body: %q", rb.SAT)
	}
	if rd := es[10].(*entity.Dimension); rd.Type != entity.DIM_ARC || !cmpF64(rd.ArcEnd, math.Pi/2) || !near(rd.DefPoint3, []float64{0, 5, 0}) {
		t.Errorf("arc dimension: %+v", rd)
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
	// transformations
	if err := d2.Mirror([]float64{0, 0}, []float64{0, 1}, es[0], es[3], es[4], es[5], es[7], es[8], es[10]); err != nil {
		t.Fatal(err)
	}
	if err := d2.Move(1, 0, 0, es[6]); err != nil {
		t.Fatal(err)
	}
	if err := d2.Move(1, 0, 0, es[9]); err == nil {
		t.Error("ACIS bodies cannot be transformed")
	}
}

func TestSATCoding(t *testing.T) {
	lines := []string{"400 0 1 0", "@7 unknown-ACIS_body Abc @_ xyz ~"}
	enc := entity.EncodeSAT(lines)
	if enc[0] == lines[0] || !reflect.DeepEqual(entity.DecodeSAT(enc), lines) {
		t.Errorf("SAT coding: %q -> %q", lines, enc)
	}
	if enc := entity.EncodeSAT([]string{"A"}); enc[0] != "^ " {
		t.Errorf("A is encoded with a space: %q", enc[0])
	}
}

func TestRawSections(t *testing.T) {
	src := "0\nSECTION\n2\nHEADER\n9\n$ACADVER\n1\nAC1027\n0\nENDSEC\n" +
		"0\nSECTION\n2\nENTITIES\n0\nENDSEC\n" +
		"0\nSECTION\n2\nACDSDATA\n70\n2\n71\n6\n0\nACDSSCHEMA\n90\n0\n1\nAcDb3DSolid_ASM_Data\n0\nENDSEC\n" +
		"0\nSECTION\n2\nTHUMBNAILIMAGE\n90\n3\n310\nABCDEF\n0\nENDSEC\n0\nEOF\n"
	d := readPreserve(t, src)
	rs := d.RawSections()
	if len(rs) != 2 || rs[0].Name != "ACDSDATA" || rs[1].Name != "THUMBNAILIMAGE" || len(rs[0].Data) != 5 {
		t.Fatalf("raw sections: %+v", rs)
	}
	out := writeString(t, d)
	if i, j := strings.Index(out, "\n2\nOBJECTS\n"), strings.Index(out, "\n2\nACDSDATA\n70\n2\n71\n6\n0\nACDSSCHEMA\n"); i < 0 || j < i ||
		!strings.Contains(out, "\n2\nTHUMBNAILIMAGE\n90\n3\n310\nABCDEF\n0\nENDSEC\n") {
		t.Error("raw sections should be written after OBJECTS")
	}
	d.SetVersion(format.R2010)
	if strings.Contains(writeString(t, d), "ACDSDATA") {
		t.Error("raw sections are not written for earlier versions")
	}
}
