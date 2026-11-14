package dxf

import (
	"testing"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/table"
)

// values of supported entities and table entries that were not kept before
// (AC1021: BLOCK_RECORD 70, 280, 281 are written for AC1021 and later)
const keepSrc = "0\nSECTION\n2\nHEADER\n9\n$ACADVER\n1\nAC1021\n0\nENDSEC\n" +
	"0\nSECTION\n2\nTABLES\n" +
	"0\nTABLE\n2\nSTYLE\n70\n1\n0\nSTYLE\n2\nTATE\n70\n4\n40\n0\n41\n1\n50\n0\n71\n6\n42\n2.5\n3\nmsgothic.ttc\n0\nENDTAB\n" +
	"0\nTABLE\n2\nBLOCK_RECORD\n70\n1\n0\nBLOCK_RECORD\n2\nB\n70\n4\n280\n0\n281\n1\n1001\nACAD\n1000\nDesignCenter Data\n0\nENDTAB\n" +
	"0\nENDSEC\n" +
	"0\nSECTION\n2\nBLOCKS\n0\nBLOCK\n8\n0\n2\nB\n70\n0\n10\n0\n20\n0\n0\nENDBLK\n8\n0\n0\nENDSEC\n" +
	"0\nSECTION\n2\nENTITIES\n" +
	"0\nLINE\n8\n0\n48\n2.0\n60\n1\n10\n0\n20\n0\n30\n0\n11\n1\n21\n0\n31\n0\n210\n0\n220\n0\n230\n-1\n" +
	"0\nPOINT\n8\n0\n10\n1\n20\n2\n30\n0\n210\n0\n220\n1\n230\n0\n50\n30\n" +
	"0\nTEXT\n8\n0\n10\n0\n20\n0\n30\n0\n40\n2\n1\nT\n210\n0\n220\n0\n230\n-1\n" +
	"0\nSPLINE\n8\n0\n48\n0.5\n70\n12\n71\n2\n72\n6\n73\n3\n74\n0\n42\n0.0000001\n43\n0.0000001\n12\n1\n22\n0\n32\n0\n13\n0\n23\n1\n33\n0\n" +
	"40\n0\n40\n0\n40\n0\n40\n1\n40\n1\n40\n1\n41\n1\n41\n0.5\n41\n1\n10\n0\n20\n0\n30\n0\n10\n1\n20\n1\n30\n0\n10\n2\n20\n0\n30\n0\n" +
	"0\nPOLYLINE\n8\n0\n66\n1\n10\n0\n20\n0\n30\n5\n70\n0\n40\n0.1\n41\n0.2\n" +
	"0\nVERTEX\n8\n0\n10\n0\n20\n0\n30\n0\n40\n0.3\n41\n0.4\n42\n0.5\n70\n0\n50\n45\n" +
	"0\nVERTEX\n8\n0\n10\n1\n20\n0\n30\n0\n70\n0\n0\nSEQEND\n8\n0\n" +
	"0\nPOLYLINE\n8\n0\n66\n1\n10\n0\n20\n0\n30\n0\n70\n64\n71\n3\n72\n1\n" +
	"0\nVERTEX\n8\n0\n10\n0\n20\n0\n30\n0\n70\n192\n0\nVERTEX\n8\n0\n10\n1\n20\n0\n30\n0\n70\n192\n0\nVERTEX\n8\n0\n10\n0\n20\n1\n30\n0\n70\n192\n" +
	"0\nVERTEX\n8\n0\n10\n0\n20\n0\n30\n0\n70\n128\n71\n1\n72\n2\n73\n-3\n0\nSEQEND\n8\n0\n" +
	"0\nDIMENSION\n8\n0\n2\n*D1\n10\n10\n20\n5\n30\n0\n11\n5\n21\n6\n31\n0\n70\n33\n1\n\n51\n15\n3\nStandard\n" +
	"100\nAcDbAlignedDimension\n12\n1\n22\n2\n32\n0\n13\n0\n23\n0\n33\n0\n14\n10\n24\n0\n34\n0\n" +
	"0\nLEADER\n8\n0\n3\nStandard\n71\n1\n72\n0\n73\n3\n76\n2\n10\n0\n20\n0\n30\n0\n10\n5\n20\n5\n30\n0\n77\n1\n212\n1\n222\n2\n232\n0\n213\n3\n223\n4\n233\n0\n" +
	"0\nENDSEC\n0\nEOF\n"

func checkKept(t *testing.T, d *drawing.Drawing) {
	t.Helper()
	st, _ := d.Style("TATE", false)
	if st.Flags != 4 || st.GenFlag != 6 || st.LastHeightUsed != 2.5 {
		t.Errorf("style: %+v", st)
	}
	r, _ := d.Sections[drawing.TABLES].(table.Tables)[table.BLOCK_RECORD].Contains("B")
	br := r.(*table.BlockRecord)
	if br.InsUnits != 4 || br.Explodable || !br.Scalable || len(br.Extra) != 2 {
		t.Errorf("block record: %+v", br)
	}

	es := d.Entities()
	l := es[0].(*entity.Line)
	if l.IsVisible() || l.Extrusion[2] != -1 || !cmpF64(l.End[0], 1) {
		t.Errorf("line: visible %v extrusion %v", l.IsVisible(), l.Extrusion)
	}
	p := es[1].(*entity.Point)
	if p.Extrusion[1] != 1 || p.Angle != 30 {
		t.Errorf("point: %+v", p)
	}
	if tx := es[2].(*entity.Text); tx.Extrusion[2] != -1 {
		t.Errorf("text extrusion: %v", tx.Extrusion)
	}
	s := es[3].(*entity.Spline)
	if len(s.Weights) != 3 || s.Weights[1] != 0.5 || s.StartTangent[0] != 1 || s.EndTangent[1] != 1 || len(s.Controls) != 3 {
		t.Errorf("spline: %+v", s)
	}
	pl := es[4].(*entity.Polyline)
	v := pl.Vertices[0]
	if pl.Elevation != 5 || pl.StartWidth != 0.1 || pl.EndWidth != 0.2 || v.StartWidth != 0.3 || v.EndWidth != 0.4 || v.Bulge != 0.5 || v.Tangent != 45 {
		t.Errorf("polyline: %+v vertex %+v", pl, v)
	}
	mesh := es[5].(*entity.Polyline)
	face := mesh.Vertices[3]
	if mesh.MeshM != 3 || mesh.MeshN != 1 || len(face.Indices) != 3 || face.Indices[2] != -3 {
		t.Errorf("polyface: M %d N %d face %v", mesh.MeshM, mesh.MeshN, face.Indices)
	}
	dim := es[6].(*entity.Dimension)
	if dim.HorizontalDirection != 15 || len(dim.ClonePoint) != 3 || dim.ClonePoint[1] != 2 {
		t.Errorf("dimension: %+v", dim)
	}
	ld := es[7].(*entity.Leader)
	if ld.DimLineColor != 1 || ld.BlockOffset[1] != 2 || ld.AnnotationOffset[0] != 3 {
		t.Errorf("leader: %+v", ld)
	}
}

func TestKeepDataOfSupportedEntities(t *testing.T) {
	d := readPreserve(t, keepSrc)
	checkKept(t, d)
	if s := d.Entities()[3].(*entity.Spline); !cmpF64(s.Ltscale(), 0.5) {
		t.Errorf("spline linetype scale: %v", s.Ltscale())
	}
	out := writeString(t, d)
	checkStructure(t, d)
	d2 := readPreserve(t, out)
	checkKept(t, d2)
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
}

func TestVisibility(t *testing.T) {
	d := NewDrawing()
	c, _ := d.Circle(0, 0, 0, 1)
	if !c.IsVisible() {
		t.Error("entities are visible by default")
	}
	c.SetVisible(false)
	d2 := readPreserve(t, writeString(t, d))
	if d2.Entities()[0].(entity.Visibility).IsVisible() {
		t.Error("visibility round trip")
	}
}
