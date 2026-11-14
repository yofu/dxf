package dxf

import (
	"math"
	"strings"
	"testing"

	"github.com/yofu/dxf/entity"
)

// regressions found by validating with ezdxf

func TestOldVersionIsWrittenAsAC1015(t *testing.T) {
	// an R12 file: written with the R2000 structure, so $ACADVER must be AC1015
	src := "0\nSECTION\n2\nHEADER\n9\n$ACADVER\n1\nAC1009\n0\nENDSEC\n" +
		"0\nSECTION\n2\nBLOCKS\n0\nBLOCK\n8\n0\n2\n$MODEL_SPACE\n70\n0\n10\n0\n20\n0\n0\nENDBLK\n8\n0\n" +
		"0\nBLOCK\n8\n0\n2\n$PAPER_SPACE\n70\n0\n10\n0\n20\n0\n0\nENDBLK\n8\n0\n0\nENDSEC\n" +
		entitiesSection("0\nLINE\n8\n0\n10\n0\n20\n0\n11\n1\n21\n1\n")
	d := readPreserve(t, src)
	out := writeString(t, d)
	if !strings.Contains(out, "$ACADVER\n1\nAC1015\n") {
		t.Error("$ACADVER should be AC1015")
	}
	if strings.Contains(out, "$MODEL_SPACE") || strings.Contains(out, "$Model_Space") {
		t.Error("R12 model space block should be *Model_Space")
	}
	if len(d.Blocks()) != 2 {
		t.Errorf("want *Model_Space and *Paper_Space, got %d blocks", len(d.Blocks()))
	}
	checkStructure(t, d)

	d = NewDrawing()
	d.Header().Version = "AC1018"
	if out := writeString(t, d); !strings.Contains(out, "$ACADVER\n1\nAC1018\n") {
		t.Error("newer versions are kept")
	}
}

func TestArcBBox(t *testing.T) {
	a := entity.NewArc(nil)
	a.Center = []float64{0, 0, 0}
	a.Radius = 10
	cases := []struct {
		start, end float64
		mins, maxs []float64
	}{
		{0, 90, []float64{0, 0}, []float64{10, 10}},
		{45, 135, []float64{-7.0710678, 7.0710678}, []float64{7.0710678, 10}},
		{270, 90, []float64{0, -10}, []float64{10, 10}}, // across 0
		{0, 360, []float64{-10, -10}, []float64{10, 10}},
	}
	for _, c := range cases {
		a.Angle = []float64{c.start, c.end}
		mins, maxs := a.BBox()
		for i := 0; i < 2; i++ {
			if !cmpF64(mins[i], c.mins[i]) || !cmpF64(maxs[i], c.maxs[i]) {
				t.Errorf("arc %v-%v: %v - %v, want %v - %v", c.start, c.end, mins, maxs, c.mins, c.maxs)
				break
			}
		}
	}
}

func TestTextBBox(t *testing.T) {
	tx := entity.NewText()
	tx.Value = "ABCD"
	tx.Height = 2
	tx.Coord1 = []float64{10, 10, 0}
	if mins, maxs := tx.BBox(); mins[0] != 10 || maxs[0] != 18 || maxs[1] != 12 {
		t.Errorf("left: %v - %v", mins, maxs)
	}
	tx.Coord2 = []float64{20, 20, 0}
	tx.Anchor(entity.CENTER_CENTER)
	if mins, maxs := tx.BBox(); mins[0] != 16 || maxs[0] != 24 || mins[1] != 19 || maxs[1] != 21 {
		t.Errorf("center: %v - %v", mins, maxs)
	}
	tx.Anchor(entity.LEFT_BASE)
	tx.Rotation = 90
	if mins, maxs := tx.BBox(); !cmpF64(mins[0], 8) || !cmpF64(maxs[0], 10) || !cmpF64(maxs[1], 18) {
		t.Errorf("rotated: %v - %v", mins, maxs)
	}
}

func TestRotatedInsertExtents(t *testing.T) {
	d := NewDrawing()
	d.BeginBlock("B", 0, 0, 0)
	d.Line(0, 0, 0, 100, 0, 0)
	d.Circle(50, 10, 0, 5)
	d.EndBlock()
	ins, _ := d.Insert("B", 0, 0, 0)
	ins.Rotation = 45
	d.SetExt()
	// the line from the origin is the leftmost part; the box of the whole
	// block rotated would give x < 0
	if h := d.Header(); h.ExtMin[0] < -1e-9 || math.Abs(h.ExtMax[0]-70.71) > 0.01 {
		t.Errorf("extents: %v - %v", h.ExtMin, h.ExtMax)
	}
}
