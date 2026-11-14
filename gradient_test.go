package dxf

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
)

func TestHatchGradient(t *testing.T) {
	d := NewDrawing()
	h, _ := d.Hatch([]float64{0, 0}, []float64{10, 0}, []float64{10, 5}, []float64{0, 5})
	h.SetPattern("ANSI31", 0, 1)
	h.SetGradient("cylinder", color.RGB(255, 0, 0), color.RGB(0, 0, 255), 30)
	if !h.Solid || h.PatternName != "SOLID" || h.PatternLines != nil {
		t.Errorf("a gradient hatch should be solid: %q %v", h.PatternName, h.Solid)
	}
	h2, _ := d.Hatch([]float64{20, 0}, []float64{30, 0}, []float64{30, 5})
	h2.SetOneColorGradient(entity.GRADIENT_SPHERICAL, color.RGB(0, 128, 0), 0.75, 0)
	h2.Gradient.Centered = 1

	if out := writeString(t, d); strings.Contains(out, "\n450\n") {
		t.Error("gradient should not be written for AC1015")
	}
	d.SetVersion(format.R2004)
	deg30 := 30.0
	out := writeString(t, d)
	want := groups(98, 0, 450, 1, 451, 0, 460, deg30*math.Pi/180, 461, 0.0, 452, 0, 462, 0.0, 453, 2,
		463, 0.0, 421, 16711680, 463, 1.0, 421, 255, 470, "CYLINDER")
	if !strings.Contains(out, want) {
		t.Errorf("gradient codes not found:\n%s", out)
	}
	checkStructure(t, d)

	d2 := readPreserve(t, out)
	g := d2.Entities()[0].(*entity.Hatch).Gradient
	if g == nil || g.Name != "CYLINDER" || math.Abs(g.Rotation-30) > 1e-9 || len(g.Colors) != 2 || g.Colors[0].RGB != color.RGB(255, 0, 0) || g.Colors[1].Value != 1 {
		t.Fatalf("read gradient: %+v", g)
	}
	g2 := d2.Entities()[1].(*entity.Hatch).Gradient
	if g2 == nil || !g2.OneColor || g2.Tint != 0.75 || g2.Centered != 1 || g2.Colors[1].RGB != color.RGB(191, 191, 191) {
		t.Errorf("read one-color gradient: %+v", g2)
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}

	// copy and transform
	c, err := Copy(d2, d2.Entities()[0])
	if err != nil {
		t.Fatal(err)
	}
	ch := c.(*entity.Hatch)
	ch.ApplyTransform(entity.Transform{Scale: 1, Angle: 15})
	if ch.Gradient == nil || math.Abs(ch.Gradient.Rotation-45) > 1e-9 || g.Rotation == ch.Gradient.Rotation {
		t.Errorf("copied gradient: %+v", ch.Gradient)
	}
	h.SetSolid()
	if h.Gradient != nil {
		t.Error("SetSolid should clear the gradient")
	}
}

// groups returns the group codes and values as written with precision 16.
func groups(cv ...interface{}) string {
	var b strings.Builder
	for i := 0; i+1 < len(cv); i += 2 {
		b.WriteString(fmt.Sprintf("\n%d\n", cv[i]))
		switch v := cv[i+1].(type) {
		case float64:
			b.WriteString(fmt.Sprintf("%.16f", v))
		default:
			b.WriteString(fmt.Sprint(v))
		}
	}
	return b.String() + "\n"
}

func TestHatchNotGradient(t *testing.T) {
	// 450 = 0: gradient data of a pattern hatch is ignored
	src := "0\nSECTION\n2\nENTITIES\n0\nHATCH\n8\n0\n100\nAcDbHatch\n10\n0\n20\n0\n30\n0\n210\n0\n220\n0\n230\n1\n2\nSOLID\n70\n1\n71\n0\n91\n0\n75\n0\n76\n1\n98\n0\n" +
		"450\n0\n451\n0\n460\n0\n461\n0\n452\n0\n462\n1\n453\n0\n470\nLINEAR\n1001\nACAD\n1000\nX\n0\nENDSEC\n0\nEOF\n"
	d := readPreserve(t, src)
	h := d.Entities()[0].(*entity.Hatch)
	if h.Gradient != nil || !h.Solid {
		t.Errorf("hatch: %+v", h.Gradient)
	}
}

func TestMTextBackground(t *testing.T) {
	d := NewDrawing()
	m1, _ := d.MText("color", 0, 0, 0, 1, 10)
	m1.SetBackgroundColor(color.Yellow, 1.2)
	m2, _ := d.MText("true color", 0, 5, 0, 1, 10)
	m2.SetBackgroundTrueColor(10, 20, 30, 2)
	m2.SetTextFrame(true)
	m3, _ := d.MText("canvas", 0, 10, 0, 1, 10)
	m3.SetBackgroundCanvas(1.5)
	m4, _ := d.MText("frame", 0, 15, 0, 1, 10)
	m4.SetTextFrame(true)

	if out := writeString(t, d); strings.Contains(out, groups(45, 1.2)) || strings.Contains(out, groups(90, 16, 0, "ENDSEC")) {
		t.Error("background fill should not be written for AC1015")
	}
	d.SetVersion(format.R2004)
	out := writeString(t, d)
	for _, w := range []string{
		groups(44, 1.0, 45, 1.2, 90, 1, 63, 2, 0, "MTEXT"),
		groups(44, 1.0, 45, 2.0, 90, 17, 63, int(color.Nearest(10, 20, 30)), 421, 660510, 0, "MTEXT"),
		groups(44, 1.0, 45, 1.5, 90, 3, 63, 0, 0, "MTEXT"),
		groups(44, 1.0, 90, 16, 0, "ENDSEC"),
	} {
		if !strings.Contains(out, w) {
			t.Errorf("%q not found", w)
		}
	}
	checkStructure(t, d)

	d2 := readPreserve(t, out)
	es := d2.Entities()
	r1, r2, r3, r4 := es[0].(*entity.MText), es[1].(*entity.MText), es[2].(*entity.MText), es[3].(*entity.MText)
	if r1.BackgroundFill != 1 || r1.BackgroundColor != 2 || r1.BackgroundScale != 1.2 || r1.BackgroundTrueColor != -1 {
		t.Errorf("color: %+v", r1)
	}
	if r2.BackgroundFill != 17 || r2.BackgroundTrueColor != color.RGB(10, 20, 30) || r2.BackgroundColor == 0 {
		t.Errorf("true color: %+v", r2)
	}
	if r3.BackgroundFill != 3 || r4.BackgroundFill != 16 {
		t.Errorf("canvas %d, frame %d", r3.BackgroundFill, r4.BackgroundFill)
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
	r2.ClearBackground()
	if r2.BackgroundFill != 16 || r2.BackgroundTrueColor != -1 {
		t.Errorf("clear: %+v", r2)
	}
}
