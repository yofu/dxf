package dxf

import (
	"strings"
	"testing"

	"github.com/yofu/dxf/entity"
)

func TestPaperSpaceEntities(t *testing.T) {
	d := NewDrawing()
	m, _ := d.Line(0, 0, 0, 1000, 1000, 0)
	d.PaperSpace()
	if !d.IsPaperSpace() {
		t.Fatal("IsPaperSpace")
	}
	frame, _ := d.LwPolyline(true, []float64{10, 10}, []float64{410, 10}, []float64{410, 287}, []float64{10, 287})
	title, _ := d.Text("タイトル", 20, 20, 0, 5)
	d.BeginBlock("B", 0, 0, 0) // entities in blocks are not marked as paper space
	inBlock, _ := d.Line(0, 0, 0, 1, 0, 0)
	d.EndBlock()
	d.ModelSpace()
	m2, _ := d.Circle(0, 0, 0, 10)

	if m.InPaperSpace() || !frame.InPaperSpace() || !title.InPaperSpace() || inBlock.InPaperSpace() || m2.InPaperSpace() {
		t.Error("paper space flags")
	}
	if len(d.PaperSpaceEntities()) != 2 || len(d.ModelSpaceEntities()) != 2 {
		t.Errorf("paper %d, model %d", len(d.PaperSpaceEntities()), len(d.ModelSpaceEntities()))
	}
	if !strings.Contains(frame.String(), "AcDbEntity\n67\n1\n8\n") {
		t.Error("67 should be written after AcDbEntity")
	}
	checkStructure(t, d)

	d.SetExt()
	if h := d.Header(); h.ExtMax[0] != 1000 || h.ExtMin[0] != -10 {
		t.Errorf("$EXTMIN/$EXTMAX are model space extents: %v - %v", h.ExtMin, h.ExtMax)
	}

	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	checkStructure(t, d2)
	if len(d2.PaperSpaceEntities()) != 2 || !d2.Entities()[1].(*entity.LwPolyline).InPaperSpace() {
		t.Error("paper space flag round trip")
	}
}

func TestViewport(t *testing.T) {
	d := NewDrawing()
	d.Line(0, 0, 0, 10000, 5000, 0)
	v, err := d.Viewport([]float64{210, 148.5, 0}, 380, 250, []float64{5000, 2500}, 6000)
	if err != nil {
		t.Fatal(err)
	}
	vs := d.PaperSpaceEntities()
	if len(vs) != 2 {
		t.Fatalf("want the paper space viewport and the new viewport, got %d", len(vs))
	}
	main := vs[0].(*entity.Viewport)
	if main.ID != 1 || main.Status != 1 || main.Width != 420 || main.Height != 297 || main.Center[0] != 210 {
		t.Errorf("main viewport: %+v", main)
	}
	if v.ID != 2 || v.Status != 2 || !v.InPaperSpace() || !cmpF64(v.Scale(), 250.0/6000) {
		t.Errorf("viewport: %+v", v)
	}
	v2, _ := d.Viewport([]float64{100, 100, 0}, 50, 50, []float64{0, 0}, 100)
	if v2.ID != 3 || len(d.PaperSpaceEntities()) != 3 {
		t.Errorf("second viewport ID %d", v2.ID)
	}
	if _, err := d.Viewport([]float64{0, 0, 0}, 0, 10, []float64{0, 0}, 1); err == nil {
		t.Error("zero width should be an error")
	}
	d.BeginBlock("B", 0, 0, 0)
	if _, err := d.Viewport([]float64{0, 0, 0}, 10, 10, []float64{0, 0}, 1); err == nil {
		t.Error("viewport in a block should be an error")
	}
	d.EndBlock()
	checkStructure(t, d)

	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	checkStructure(t, d2)
	g := d2.Entities()[2].(*entity.Viewport)
	if g.ID != 2 || g.Width != 380 || g.Height != 250 || g.ViewCenter[0] != 5000 || g.ViewHeight != 6000 ||
		g.ViewDirection[2] != 1 || g.UCSXAxis[0] != 1 || !g.InPaperSpace() || g.Flags != v.Flags {
		t.Errorf("viewport round trip: %+v", g)
	}
}
