package dxf

import (
	"strings"
	"testing"

	"github.com/yofu/dxf/entity"
)

func TestPdfUnderlay(t *testing.T) {
	d := NewDrawing()
	def, err := d.AddUnderlayDefinition(entity.UNDERLAY_PDF, "plans/site.pdf", "2")
	if err != nil {
		t.Fatal(err)
	}
	def2, _ := d.AddUnderlayDefinition("pdf", "other/site.pdf", "2") // same key
	if _, err := d.AddUnderlayDefinition("PNG", "a.png", ""); err == nil {
		t.Error("unknown format should be an error")
	}
	u, _ := d.Underlay(def, 10, 20, 0, 2)
	u.Rotation = 30
	u.SetClipping([]float64{0, 0}, []float64{100, 50})
	u.Fade = 25
	d.Underlay(def2, 0, 0, 0, 1)

	out := writeString(t, d)
	checkStructure(t, d)
	for _, w := range []string{
		"\n0\nPDFUNDERLAY\n",
		"\n100\nAcDbUnderlayReference\n340\n" + hexHandle(def.Handle()) + "\n",
		"\n280\n3\n281\n100\n282\n25\n",
		"\n3\nACAD_PDFDEFINITIONS\n",
		"\n3\nsite - 2\n", "\n3\nsite - 2 (1)\n",
		"\n0\nPDFDEFINITION\n5\n" + hexHandle(def.Handle()) + "\n102\n{ACAD_REACTORS\n",
		"\n330\n" + hexHandle(u.Handle()) + "\n102\n}\n",
		"\n100\nAcDbUnderlayDefinition\n1\nplans/site.pdf\n2\n2\n",
		"\n1\nPDFDEFINITION\n2\nAcDbPdfDefinition\n", "\n1\nPDFUNDERLAY\n2\nAcDbPdfReference\n",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("%q not found", w)
		}
	}

	d2 := readPreserve(t, out)
	defs := d2.UnderlayDefinitions()
	if len(defs) != 2 || defs[0].FileName != "plans/site.pdf" || defs[0].Name != "2" || defs[0].FileFormat != "PDF" {
		t.Fatalf("definitions: %+v", defs)
	}
	r := d2.Entities()[0].(*entity.Underlay)
	if r.Definition != defs[0] || r.Insert[0] != 10 || r.Scale[1] != 2 || r.Rotation != 30 || len(r.Boundary) != 2 || r.Boundary[1][0] != 100 || r.Fade != 25 || r.Flags != 3 {
		t.Errorf("underlay: %+v", r)
	}
	if n := len(d2.RawObjects()); n != 0 {
		t.Errorf("no object should be kept raw: %d", n)
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}

	// copy shares the definition; transform
	c, err := Copy(d2, r)
	if err != nil {
		t.Fatal(err)
	}
	cu := c.(*entity.Underlay)
	cu.ApplyTransform(entity.Transform{Scale: 2, Angle: 15})
	if cu.Definition != defs[0] || cu.Scale[0] != 4 || cu.Rotation != 45 {
		t.Errorf("copy: %+v", cu)
	}
	min, max := r.BBox()
	if min[0] > 10 || max[0] < 10+200*0.86 {
		t.Errorf("bbox: %v %v", min, max)
	}
}

// an underlay whose definition is not found keeps its handle
func TestUnderlayUnresolved(t *testing.T) {
	src := "0\nSECTION\n2\nENTITIES\n0\nPDFREFERENCE\n5\n50\n8\n0\n100\nAcDbUnderlayReference\n340\n99\n10\n1\n20\n2\n30\n0\n41\n1\n42\n1\n43\n1\n50\n0\n280\n10\n281\n50\n282\n0\n0\nENDSEC\n0\nEOF\n"
	d := readPreserve(t, src)
	u := d.Entities()[0].(*entity.Underlay)
	if u.Definition != nil || u.DefinitionHandle() != "99" || u.TypeName() != "PDFREFERENCE" || u.Contrast != 50 || u.Flags != 10 {
		t.Errorf("underlay: %+v", u)
	}
	if out := writeString(t, d); !strings.Contains(out, "\n0\nPDFREFERENCE\n") || !strings.Contains(out, "\n340\n99\n") {
		t.Error("type and definition handle should be written as read")
	}
}
