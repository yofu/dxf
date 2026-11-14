package dxf

import (
	"strings"
	"testing"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/header"
)

func TestHeaderVariablesRoundTrip(t *testing.T) {
	d := NewDrawing()
	h := d.Header()
	if h.Measurement != header.Metric || h.CEColor != 256 || h.CELType != "ByLayer" || h.CELWeight != -1 || h.CLayer != "0" {
		t.Errorf("defaults: %+v", h)
	}
	h.LimMin = []float64{-10, -20}
	h.LimMax = []float64{841, 594}
	h.TextSize = 3.5
	h.CELTScale = 0.5
	h.DimStyle = "Standard"
	h.Measurement = header.Imperial
	h.LUPrec = 2
	h.AUnits = 1
	h.AUPrec = 3
	h.AngBase = 90
	h.AngDir = 1
	h.PdMode = 3
	h.PdSize = 1.5
	h.CEColor = int(color.Blue)
	h.CELType = "HIDDEN"
	h.CELWeight = 25

	out := writeString(t, d)
	for _, v := range []string{"$CLAYER\n8\n0\n", "$CECOLOR\n62\n5\n", "$CELTYPE\n6\nHIDDEN\n", "$CELWEIGHT\n370\n25\n", "$MEASUREMENT\n70\n0\n"} {
		if !strings.Contains(out, v) {
			t.Errorf("header should contain %q", v)
		}
	}
	d2, err := FromStringData(out)
	if err != nil {
		t.Fatal(err)
	}
	g := d2.Header()
	if g.LimMin[0] != -10 || g.LimMax[1] != 594 || g.TextSize != 3.5 || g.CELTScale != 0.5 || g.Measurement != header.Imperial ||
		g.LUPrec != 2 || g.AUnits != 1 || g.AUPrec != 3 || g.AngBase != 90 || g.AngDir != 1 || g.PdMode != 3 || g.PdSize != 1.5 ||
		g.CEColor != 5 || g.CELType != "HIDDEN" || g.CELWeight != 25 || g.DimStyle != "Standard" {
		t.Errorf("round trip: %+v", g)
	}
}

func TestCurrentLayerAndStyle(t *testing.T) {
	d := NewDrawing()
	d.AddLayer("壁", color.Red, DefaultLineType, false)
	d.AddStyle("MINCHO", "msmincho.ttc", "", false)
	d.ChangeLayer("壁")
	d.Style("MINCHO", true)
	out := writeString(t, d)
	if !strings.Contains(out, "$TEXTSTYLE\n7\nMINCHO\n") {
		t.Error("$TEXTSTYLE should be the current style")
	}
	d2, err := FromStringData(out)
	if err != nil {
		t.Fatal(err)
	}
	if d2.CurrentLayer.Name() != "壁" || d2.Header().CLayer != "壁" || d2.CurrentStyle.Name() != "MINCHO" {
		t.Errorf("current layer %s / style %s", d2.CurrentLayer.Name(), d2.CurrentStyle.Name())
	}
	// new entities use the current layer read from the file
	if l, _ := d2.Line(0, 0, 0, 1, 1, 0); l.Layer().Name() != "壁" {
		t.Errorf("new line on layer %s", l.Layer().Name())
	}
}

func TestCurrentEntityProperties(t *testing.T) {
	d := NewDrawing()
	before, _ := d.Line(0, 0, 0, 1, 0, 0)
	h := d.Header()
	h.CEColor = int(color.Red)
	h.CELType = "hidden"
	h.CELWeight = 50
	h.CELTScale = 2
	l, _ := d.Line(0, 0, 0, 1, 1, 0)
	c, _ := d.Circle(0, 0, 0, 1)
	for _, e := range []interface {
		ColorIndex() int
		LineTypeName() string
		LineWeight() int
	}{l, c} {
		if e.ColorIndex() != 1 || e.LineTypeName() != "HIDDEN" || e.LineWeight() != 50 {
			t.Errorf("%T: color %d linetype %s lineweight %d", e, e.ColorIndex(), e.LineTypeName(), e.LineWeight())
		}
	}
	if !strings.Contains(l.String(), "\n48\n2.0") {
		t.Error("linetype scale should be $CELTSCALE")
	}
	if !before.IsColorByLayer() || before.LineType() != nil {
		t.Error("entities created before are not changed")
	}
}
