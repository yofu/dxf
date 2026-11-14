package dxf

import (
	"strings"
	"testing"

	"github.com/yofu/dxf/entity"
)

func TestLwPolylineWidthsAndElevation(t *testing.T) {
	d := NewDrawing()
	lw, _ := d.LwPolyline(true, []float64{0, 0}, []float64{10, 0}, []float64{10, 10})
	lw.SetWidth(0, 1, 2)
	lw.SetWidth(2, 0.5, 0.5)
	lw.Elevation = 3
	lw.Plinegen = true
	lw.Bulges[1] = 0.5

	s := lw.String()
	if !strings.Contains(s, "\n70\n129\n38\n3.0") {
		t.Error("70 should be 129 (closed + plinegen) followed by elevation 38")
	}
	if !strings.Contains(s, "\n10\n0.000000\n20\n0.000000\n40\n1.000000\n41\n2.000000\n") {
		t.Error("widths 40/41 should follow the vertex")
	}

	cw, _ := d.LwPolyline(false, []float64{0, 0}, []float64{5, 5})
	cw.ConstantWidth = 0.7
	cw.Extrusion = []float64{0, 0, -1}
	if s := cw.String(); !strings.Contains(s, "\n43\n0.7") || strings.Contains(s, "\n40\n") || !strings.Contains(s, "\n230\n-1.0") {
		t.Error("constant width 43 and extrusion 210 should be written, and no 40/41")
	}

	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	g := d2.Entities()[0].(*entity.LwPolyline)
	if !g.Closed || !g.Plinegen || g.Elevation != 3 || g.Bulges[1] != 0.5 ||
		g.Widths[0][0] != 1 || g.Widths[0][1] != 2 || g.Widths[1][0] != 0 || g.Widths[2][1] != 0.5 {
		t.Errorf("round trip: %+v", g)
	}
	g2 := d2.Entities()[1].(*entity.LwPolyline)
	if g2.ConstantWidth != 0.7 || g2.Extrusion[2] != -1 || g2.Closed || g2.Plinegen {
		t.Errorf("round trip constant width: %+v", g2)
	}
}

func TestReadLwPolylineWidths(t *testing.T) {
	// AutoCAD order: 90, 70, 43, 38, then 10, 20, 40, 41, 42 per vertex
	src := entitiesSection("0\nLWPOLYLINE\n8\n0\n100\nAcDbPolyline\n90\n2\n70\n0\n43\n0\n38\n-1.5\n" +
		"10\n0\n20\n0\n40\n0.2\n41\n0.4\n42\n1\n10\n4\n20\n0\n40\n0.4\n41\n0.2\n")
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	lw := d.Entities()[0].(*entity.LwPolyline)
	if lw.Elevation != -1.5 || lw.Widths[0][1] != 0.4 || lw.Widths[1][0] != 0.4 || lw.Bulges[0] != 1 {
		t.Errorf("lwpolyline: %+v", lw)
	}
	mins, maxs := lw.BBox()
	// half of the largest width (0.4) around the vertices, z = elevation
	if !cmpF64(mins[0], -0.2) || !cmpF64(maxs[0], 4.2) || !cmpF64(mins[1], -0.2) || mins[2] != -1.5 || maxs[2] != -1.5 {
		t.Errorf("bbox: %v - %v", mins, maxs)
	}
}
