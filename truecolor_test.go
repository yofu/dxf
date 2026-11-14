package dxf

import (
	"strconv"
	"strings"
	"testing"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/table"
)

func TestEntityTrueColorAndTransparency(t *testing.T) {
	d := NewDrawing()
	l, _ := d.Line(0, 0, 0, 1, 1, 0)
	l.SetTrueColor(255, 128, 0)
	l.SetColorName("RAL CLASSIC$RAL 2003")
	l.SetTransparency(0.5)
	if l.ColorIndex() < 1 || l.ColorIndex() > 255 {
		t.Errorf("color number should be the nearest one: %d", l.ColorIndex())
	}
	c, _ := d.Circle(0, 0, 0, 1)
	c.SetTransparencyByBlock()

	// AC1015: not written
	if out := writeString(t, d); strings.Contains(out, "\n420\n") || strings.Contains(out, "\n440\n") {
		t.Error("420 and 440 should not be written for AC1015")
	}
	// AC1018: 420, 430; not 440
	d.SetVersion(format.R2004)
	out := writeString(t, d)
	if !strings.Contains(out, "\n420\n16744448\n430\nRAL CLASSIC$RAL 2003\n") || strings.Contains(out, "\n440\n") {
		t.Error("AC1018: true color and color name should be written, transparency not")
	}
	// AC1024: 440
	if err := d.SetVersion(format.R2010); err != nil {
		t.Fatal(err)
	}
	out = writeString(t, d)
	if !strings.Contains(out, "\n440\n33554560\n") || !strings.Contains(out, "\n440\n16777216\n") {
		t.Error("AC1024: transparency should be written (0.5 -> alpha 128, BYBLOCK)")
	}
	if err := d.SetVersion("AC1009"); err == nil {
		t.Error("unsupported version should be an error")
	}

	d2 := readPreserve(t, out)
	if d2.Version() != format.R2010 {
		t.Errorf("version: %s", d2.Version())
	}
	g := d2.Entities()[0].(*entity.Line)
	r, gg, b, ok := g.TrueColor()
	tr, tok := g.Transparency()
	if !ok || r != 255 || gg != 128 || b != 0 || g.ColorName() != "RAL CLASSIC$RAL 2003" || !tok || tr < 0.49 || tr > 0.51 || g.ColorIndex() != l.ColorIndex() {
		t.Errorf("round trip: %v %v %v %v %q %v %v", r, gg, b, ok, g.ColorName(), tr, tok)
	}
	if _, ok := d2.Entities()[1].(entity.TrueColorer).Transparency(); ok {
		t.Error("BYBLOCK transparency has no value")
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
}

func TestLayerTrueColorAndTransparency(t *testing.T) {
	d := NewDrawing()
	d.SetVersion(format.R2010)
	l, _ := d.AddLayer("GLASS", color.Cyan, DefaultLineType, false)
	l.SetTrueColor(10, 200, 250)
	l.SetTransparency(0.3)
	out := writeString(t, d)
	if !strings.Contains(out, "\n420\n"+strconv.Itoa(color.RGB(10, 200, 250))+"\n") ||
		!strings.Contains(out, "\n1001\nAcCmTransparency\n1071\n") || !strings.Contains(out, "\n2\nAcCmTransparency\n70\n0\n") {
		t.Error("layer true color, transparency and its APPID should be written")
	}
	checkStructure(t, d)

	d2 := readPreserve(t, out)
	g := d2.Layers["GLASS"]
	r, gg, b, ok := g.TrueColor()
	tr, tok := g.Transparency()
	if !ok || r != 10 || gg != 200 || b != 250 || !tok || tr < 0.29 || tr > 0.31 {
		t.Errorf("layer round trip: %v %v %v %v %v", r, gg, b, tr, tok)
	}
	for _, p := range g.Extra {
		if p[1] == table.TransparencyApp || p[0] == "420" {
			t.Errorf("Extra should not duplicate: %v", g.Extra)
		}
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}

	d.SetVersion(format.R2004)
	if out := writeString(t, d); strings.Contains(out, "AcCmTransparency\n1071") {
		t.Error("layer transparency is not written before AC1024")
	}
}

func TestExplodeTrueColorByBlock(t *testing.T) {
	d := NewDrawing()
	d.BeginBlock("B", 0, 0, 0)
	l, _ := d.Line(0, 0, 0, 1, 0, 0)
	l.SetColorByBlock()
	l.SetTransparencyByBlock()
	d.EndBlock()
	ins, _ := d.Insert("B", 0, 0, 0)
	ins.SetTrueColor(1, 2, 3)
	ins.SetTransparency(0.2)
	es, err := Explode(d, ins)
	if err != nil {
		t.Fatal(err)
	}
	x := es[0].(*entity.Line)
	r, g, b, ok := x.TrueColor()
	tr, tok := x.Transparency()
	if !ok || r != 1 || g != 2 || b != 3 || x.ColorIndex() != ins.ColorIndex() || !tok || tr < 0.19 || tr > 0.21 {
		t.Errorf("exploded: %v %v %v %v color %d transparency %v", r, g, b, ok, x.ColorIndex(), tr)
	}
}
