package dxf

import (
	"strings"
	"testing"

	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
)

func TestImage(t *testing.T) {
	d := NewDrawing()
	def, err := d.AddImageDefinition("images/plan.png", 640, 480)
	if err != nil {
		t.Fatal(err)
	}
	im, _ := d.Image(def, 10, 20, 0, 64, 48, 90)
	im.SetBoundary([]float64{-0.5, -0.5}, []float64{319.5, 239.5})
	im.Fade = 30
	d.Image(def, 0, 0, 0, 6.4, 4.8, 0)
	w, _ := d.Wipeout([]float64{0, 0}, []float64{10, 0}, []float64{10, 5}, []float64{0, 5})
	if _, err := d.AddImageDefinition("x.png", 0, 1); err == nil {
		t.Error("zero size should be an error")
	}
	if mins, maxs := im.BBox(); !near(mins, []float64{-38, 20, 0}) || !near(maxs, []float64{-14, 52, 0}) {
		t.Errorf("image bbox: %v %v", mins, maxs)
	}
	if mins, maxs := w.BBox(); !near(mins, []float64{0, 0, 0}) || !near(maxs, []float64{10, 5, 0}) {
		t.Errorf("wipeout bbox: %v %v", mins, maxs)
	}

	d.SetVersion(format.R2010)
	out := writeString(t, d)
	checkStructure(t, d)
	for _, s := range []string{
		"\n0\nIMAGE\n", "\n100\nAcDbRasterImage\n90\n0\n",
		"\n340\n" + hexHandle(def.Handle()) + "\n70\n7\n280\n1\n281\n50\n282\n50\n283\n30\n360\n" + hexHandle(im.Reactor().Handle()) + "\n71\n1\n91\n2\n",
		"\n0\nIMAGEDEF_REACTOR\n5\n" + hexHandle(im.Reactor().Handle()) + "\n330\n" + hexHandle(im.Handle()) + "\n",
		"\n3\nACAD_IMAGE_DICT\n", "\n3\nplan\n",
		"\n100\nAcDbRasterImageDef\n90\n0\n1\nimages/plan.png\n",
		"\n0\nWIPEOUT\n", "\n100\nAcDbWipeout\n", "\n71\n2\n91\n5\n", "\n290\n0\n",
		"\n1\nIMAGEDEF_REACTOR\n", "\n1\nWIPEOUT\n2\nAcDbWipeout\n3\nWipeOut\n",
	} {
		if !strings.Contains(out, s) {
			t.Errorf("%q not found", s)
		}
	}

	d2 := readPreserve(t, out)
	defs := d2.ImageDefinitions()
	if len(defs) != 1 || defs[0].FileName != "images/plan.png" || defs[0].Size[1] != 480 {
		t.Fatalf("definitions: %+v", defs)
	}
	r := d2.Entities()[0].(*entity.Image)
	if r.Definition != defs[0] || r.Reactor() == nil || r.Fade != 30 || len(r.Boundary) != 2 || !near(r.V, []float64{-0.1, 0, 0}) {
		t.Errorf("image: %+v", r)
	}
	if rw := d2.Entities()[2].(*entity.Image); !rw.IsWipeout() || rw.Definition != nil || len(rw.Boundary) != 5 {
		t.Errorf("wipeout: %+v", rw)
	}
	if raws := d2.RawObjects(); len(raws) != 0 {
		t.Errorf("raw objects: %v", raws)
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}

	// copy: shared definition, its own reactor; mirrored
	c, err := CopyEntities(d2, entity.MirrorBy([]float64{0, 0}, []float64{0, 1}), r)
	if err != nil {
		t.Fatal(err)
	}
	ci := c[0].(*entity.Image)
	out2 := writeString(t, d2)
	if ci.Definition != defs[0] || ci.Reactor() == nil || ci.Reactor() == r.Reactor() || !near(ci.Insert, []float64{-10, 20, 0}) {
		t.Errorf("copy: %+v", ci)
	}
	if strings.Count(out2, "\n0\nIMAGEDEF_REACTOR\n") != 3 {
		t.Error("each image should have a reactor")
	}
	checkStructure(t, d2)
	// removed images lose their reactors
	d2.RemoveEntity(ci)
	if strings.Count(writeString(t, d2), "\n0\nIMAGEDEF_REACTOR\n") != 2 {
		t.Error("the reactor of a removed image should be removed")
	}
}

// LTYPE, STYLE, APPID and VPORT keep the group codes and XDATA they do
// not interpret
func TestTableExtra(t *testing.T) {
	src := "0\nSECTION\n2\nTABLES\n" +
		"0\nTABLE\n2\nVPORT\n70\n1\n0\nVPORT\n5\n20\n2\n*ACTIVE\n70\n0\n10\n0\n20\n0\n11\n1\n21\n1\n40\n300\n41\n1.5\n71\n16\n76\n1\n1001\nMYAPP\n1000\nvport\n0\nENDTAB\n" +
		"0\nTABLE\n2\nLTYPE\n70\n1\n0\nLTYPE\n5\n21\n2\nDASH\n70\n0\n3\n__\n72\n65\n73\n2\n40\n1.5\n49\n1\n74\n0\n49\n-0.5\n74\n0\n1001\nMYAPP\n1000\nltype\n0\nENDTAB\n" +
		"0\nTABLE\n2\nSTYLE\n70\n1\n0\nSTYLE\n5\n22\n2\nARIAL\n70\n0\n40\n0\n41\n1\n50\n0\n71\n0\n42\n2.5\n3\narial.ttf\n4\n\n1001\nACAD\n1000\nArial\n1071\n34\n0\nENDTAB\n" +
		"0\nTABLE\n2\nAPPID\n70\n1\n0\nAPPID\n5\n23\n2\nMYAPP\n70\n0\n1001\nMYAPP\n1000\nappid\n0\nENDTAB\n" +
		"0\nENDSEC\n0\nEOF\n"
	d := readPreserve(t, src)
	out := writeString(t, d)
	for _, s := range []string{
		"\n71\n16\n76\n1\n1001\nMYAPP\n1000\nvport\n",
		"\n49\n-0.5000000000000000\n74\n0\n1001\nMYAPP\n1000\nltype\n",
		"\n4\n\n1001\nACAD\n1000\nArial\n1071\n34\n",
		"\n2\nMYAPP\n70\n0\n1001\nMYAPP\n1000\nappid\n",
	} {
		if !strings.Contains(out, s) {
			t.Errorf("%q not found", s)
		}
	}
	if writeString(t, readPreserve(t, out)) != out {
		t.Error("writing a read drawing again should give the same output")
	}
	checkStructure(t, d)
}
