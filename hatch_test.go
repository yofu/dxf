package dxf

import (
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

func TestEllipseRoundTrip(t *testing.T) {
	d := NewDrawing()
	e, err := d.Ellipse([]float64{1, 2, 0}, []float64{3, 0, 0}, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	e.Start, e.End = 0.5, 2.5
	e.SetColor(color.Cyan)
	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	g := d2.Entities()[0].(*entity.Ellipse)
	if g.Center[1] != 2 || g.MajorAxis[0] != 3 || g.Ratio != 0.5 || g.Start != 0.5 || g.End != 2.5 ||
		g.Extrusion[2] != 1 || g.ColorIndex() != int(color.Cyan) {
		t.Errorf("ellipse round trip: %+v", g)
	}
	for _, c := range []struct {
		center, major []float64
		ratio         float64
	}{{[]float64{0, 0}, []float64{1, 0, 0}, 1}, {[]float64{0, 0, 0}, []float64{1, 0, 0}, 0}, {[]float64{0, 0, 0}, []float64{1, 0, 0}, 1.5}, {[]float64{0, 0, 0}, []float64{0, 0, 0}, 1}} {
		if _, err := d.Ellipse(c.center, c.major, c.ratio); err == nil {
			t.Errorf("invalid ellipse %v %v %v should be an error", c.center, c.major, c.ratio)
		}
	}
}

func TestEllipseBBox(t *testing.T) {
	check := func(name string, e *entity.Ellipse, wmin, wmax []float64) {
		t.Helper()
		mins, maxs := e.BBox()
		for i := 0; i < 2; i++ {
			if !cmpF64(mins[i], wmin[i]) || !cmpF64(maxs[i], wmax[i]) {
				t.Errorf("%s: got %v - %v, want %v - %v", name, mins, maxs, wmin, wmax)
				return
			}
		}
	}
	e := entity.NewEllipse()
	e.Center = []float64{1, 1, 0}
	e.MajorAxis = []float64{2, 0, 0}
	e.Ratio = 0.5
	check("full", e, []float64{-1, 0}, []float64{3, 2})
	e.MajorAxis = []float64{0, 2, 0}
	check("rotated", e, []float64{0, -1}, []float64{2, 3})
	e.MajorAxis = []float64{2, 0, 0}
	e.Start, e.End = 0, math.Pi/2
	check("quarter", e, []float64{1, 1}, []float64{3, 2})
	e.Start, e.End = 3*math.Pi/2, math.Pi/2 // across 0
	check("wrapped", e, []float64{1, 0}, []float64{3, 2})
}

func TestHatchSolid(t *testing.T) {
	d := NewDrawing()
	h, err := d.Hatch([]float64{0, 0}, []float64{10, 0}, []float64{10, 5}, []float64{0, 5})
	if err != nil {
		t.Fatal(err)
	}
	h.SetColor(color.Yellow)
	// a hole with a bulged edge
	hole := h.AddPolylinePath([][]float64{{2, 2}, {4, 2}, {4, 3}}, []float64{0, 0.5, 0})
	if h.Boundaries[0].Flag != entity.HATCH_PATH_EXTERNAL|entity.HATCH_PATH_POLYLINE || hole.Flag != entity.HATCH_PATH_POLYLINE {
		t.Errorf("path flags: outer %d, hole %d", h.Boundaries[0].Flag, hole.Flag)
	}
	d.Ellipse([]float64{20, 0, 0}, []float64{1, 0, 0}, 0.5)

	s := h.String()
	if !strings.Contains(s, "\n2\nSOLID\n70\n1\n") || strings.Contains(s, "\n78\n") {
		t.Error("solid hatch: 70 should be 1 without pattern lines")
	}
	checkStructure(t, d)

	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	g := d2.Entities()[0].(*entity.Hatch)
	if !g.Solid || g.PatternName != "SOLID" || len(g.Boundaries) != 2 || g.ColorIndex() != int(color.Yellow) {
		t.Fatalf("hatch round trip: %+v", g)
	}
	b0, b1 := g.Boundaries[0], g.Boundaries[1]
	if !b0.IsPolyline() || !b0.Closed || len(b0.Vertices) != 4 || b0.Vertices[2][0] != 10 || b0.Vertices[2][1] != 5 {
		t.Errorf("outer path: %+v", b0)
	}
	if b1.Flag != entity.HATCH_PATH_POLYLINE || b1.Bulges[1] != 0.5 {
		t.Errorf("hole path: %+v", b1)
	}
	if _, err := d.Hatch([]float64{0, 0}, []float64{1, 0}); err == nil {
		t.Error("hatch with 2 vertices should be an error")
	}
}

func TestHatchPatterns(t *testing.T) {
	d := NewDrawing()
	h, _ := d.Hatch([]float64{0, 0}, []float64{10, 0}, []float64{10, 10})
	if err := h.SetPattern("ansi31", 0, 1); err != nil {
		t.Fatal(err)
	}
	if err := h.SetPattern("NOSUCH", 0, 1); err == nil {
		t.Error("unknown pattern should be an error")
	}
	l := h.PatternLines[0]
	// ANSI31: 45 degrees, spacing 3.175 => offset (-2.245, 2.245)
	if h.PatternName != "ANSI31" || h.Solid || l.Angle != 45 || !cmpF64(l.Offset[0], -2.2450640) || !cmpF64(l.Offset[1], 2.2450640) {
		t.Errorf("ANSI31: %+v %+v", h, l)
	}
	h.SetPattern("ANSI31", 15, 2)
	if l := h.PatternLines[0]; l.Angle != 60 || !cmpF64(math.Hypot(l.Offset[0], l.Offset[1]), 6.35) {
		t.Errorf("ANSI31 angle 15 scale 2: %+v", l)
	}

	h2, _ := d.Hatch([]float64{0, 0}, []float64{10, 0}, []float64{10, 10})
	h2.SetUserPattern(30, 5, true)
	if h2.PatternType != entity.HATCH_PATTERN_USER || len(h2.PatternLines) != 2 ||
		h2.PatternLines[1].Angle != 120 || !cmpF64(math.Hypot(h2.PatternLines[1].Offset[0], h2.PatternLines[1].Offset[1]), 5) {
		t.Errorf("user pattern: %+v", h2.PatternLines)
	}

	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	for i, e := range d2.Entities() {
		g, want := e.(*entity.Hatch), d.Entities()[i].(*entity.Hatch)
		if g.PatternName != want.PatternName || g.Solid || g.PatternAngle != want.PatternAngle || g.PatternScale != want.PatternScale ||
			g.Double != want.Double || g.PatternType != want.PatternType || len(g.PatternLines) != len(want.PatternLines) {
			t.Fatalf("pattern round trip: %+v", g)
		}
		for j, l := range g.PatternLines {
			w := want.PatternLines[j]
			if !cmpF64(l.Angle, w.Angle) || !cmpF64(l.Offset[0], w.Offset[0]) || !cmpF64(l.Offset[1], w.Offset[1]) || len(l.Dashes) != len(w.Dashes) {
				t.Errorf("pattern line %d: %+v, want %+v", j, l, w)
			}
		}
	}
}

func TestHatchEdgePath(t *testing.T) {
	d := NewDrawing()
	h := entity.NewHatch()
	h.AddEdgePath(
		&entity.HatchLine{Start: []float64{0, 0}, End: []float64{10, 0}},
		&entity.HatchArc{Center: []float64{10, 5}, Radius: 5, Start: 270, End: 90, CCW: true},
		&entity.HatchEllipse{Center: []float64{5, 10}, MajorAxis: []float64{5, 0}, Ratio: 0.5, Start: 0, End: 180, CCW: true},
		&entity.HatchSpline{Degree: 3, Rational: true, Knots: []float64{0, 0, 0, 0, 1, 1, 1, 1},
			Controls: [][]float64{{0, 10}, {0, 7}, {0, 3}, {0, 0}}, Weights: []float64{1, 0.5, 0.5, 1}},
	)
	h.Seeds = [][]float64{{5, 5}}
	d.AddEntity(h)
	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	g := d2.Entities()[0].(*entity.Hatch)
	if len(g.Boundaries) != 1 || len(g.Boundaries[0].Edges) != 4 || len(g.Seeds) != 1 || g.Seeds[0][0] != 5 {
		t.Fatalf("edge path round trip: %+v", g)
	}
	es := g.Boundaries[0].Edges
	if l := es[0].(*entity.HatchLine); l.End[0] != 10 {
		t.Errorf("line: %+v", l)
	}
	if a := es[1].(*entity.HatchArc); a.Radius != 5 || a.Start != 270 || a.End != 90 || !a.CCW {
		t.Errorf("arc: %+v", a)
	}
	if e := es[2].(*entity.HatchEllipse); e.MajorAxis[0] != 5 || e.Ratio != 0.5 || e.End != 180 || !e.CCW {
		t.Errorf("ellipse: %+v", e)
	}
	if s := es[3].(*entity.HatchSpline); s.Degree != 3 || !s.Rational || len(s.Knots) != 8 || len(s.Controls) != 4 || len(s.Weights) != 4 || s.Weights[1] != 0.5 {
		t.Errorf("spline: %+v", s)
	}
	mins, maxs := g.BBox()
	if mins[0] != 0 || mins[1] != 0 || maxs[0] != 15 || maxs[1] != 15 {
		t.Errorf("bbox: %v - %v", mins, maxs)
	}
}

func TestReadHatchFromFile(t *testing.T) {
	// associative hatch with source objects, pixel size, seed points and a
	// spline edge with fit data (AC1024 style), followed by another entity
	src := entitiesSection("0\nHATCH\n5\n2F\n8\n0\n100\nAcDbHatch\n10\n0\n20\n0\n30\n1.5\n210\n0\n220\n0\n230\n1\n2\nANSI31\n70\n0\n71\n1\n91\n2\n" +
		"92\n7\n72\n0\n73\n1\n93\n3\n10\n0\n20\n0\n10\n4\n20\n0\n10\n4\n20\n3\n97\n1\n330\n2A\n" +
		"92\n1\n93\n1\n72\n4\n94\n3\n73\n0\n74\n0\n95\n8\n96\n4\n40\n0\n40\n0\n40\n0\n40\n0\n40\n1\n40\n1\n40\n1\n40\n1\n" +
		"10\n0\n20\n0\n10\n1\n20\n1\n10\n2\n20\n1\n10\n3\n20\n0\n97\n2\n11\n0\n21\n0\n11\n3\n21\n0\n12\n1\n22\n0\n13\n1\n23\n0\n97\n0\n" +
		"75\n1\n76\n1\n52\n0\n41\n1\n77\n0\n78\n1\n53\n45\n43\n0\n44\n0\n45\n-2.245064973\n46\n2.245064973\n79\n0\n47\n0.05\n98\n1\n10\n1\n20\n1\n" +
		"0\nPOINT\n8\n0\n10\n3\n20\n4\n")
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Entities()) != 2 {
		t.Fatalf("want HATCH and POINT, got %d entities", len(d.Entities()))
	}
	h := d.Entities()[0].(*entity.Hatch)
	if h.Elevation != 1.5 || h.PatternName != "ANSI31" || h.Solid || !h.Associative || len(h.Boundaries) != 2 ||
		h.Style != 1 || len(h.PatternLines) != 1 || h.PatternLines[0].Angle != 45 || len(h.Seeds) != 1 {
		t.Fatalf("hatch: %+v", h)
	}
	if b := h.Boundaries[0]; !b.IsPolyline() || len(b.Vertices) != 3 || !b.Closed {
		t.Errorf("polyline path: %+v", b)
	}
	s := h.Boundaries[1].Edges[0].(*entity.HatchSpline)
	if len(s.Controls) != 4 || len(s.Fits) != 2 {
		t.Errorf("spline edge: %+v", s)
	}
	// the source boundary objects are kept (their handles are kept)
	if !strings.Contains(h.String(), "\n97\n1\n330\n2A\n") {
		t.Error("source boundary handles should be written")
	}
}

func TestHatchExtents(t *testing.T) {
	d := NewDrawing()
	d.Hatch([]float64{10, 10}, []float64{20, 10}, []float64{20, 15})
	d.SetExt()
	h := d.Header()
	if h.ExtMin[0] != 10 || h.ExtMin[1] != 10 || h.ExtMax[0] != 20 || h.ExtMax[1] != 15 {
		t.Errorf("extents: %v - %v", h.ExtMin, h.ExtMax)
	}
}
