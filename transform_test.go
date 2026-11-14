package dxf

import (
	"errors"
	"math"
	"testing"

	"github.com/yofu/dxf/entity"
)

func ellipsePoint(e *entity.Ellipse, p float64) []float64 {
	ux, uy := e.MajorAxis[0], e.MajorAxis[1]
	vx, vy := -uy*e.Ratio, ux*e.Ratio
	return []float64{e.Center[0] + ux*math.Cos(p) + vx*math.Sin(p), e.Center[1] + uy*math.Cos(p) + vy*math.Sin(p)}
}

func TestMirrorArc(t *testing.T) {
	d := NewDrawing()
	a, _ := d.Arc(2, 0, 0, 1, 0, 90)
	if err := d.Mirror([]float64{0, 0}, []float64{0, 1}, a); err != nil {
		t.Fatal(err)
	}
	if !near(a.Center, []float64{-2, 0, 0}) || !cmpF64(a.Angle[0], 90) || !cmpF64(a.Angle[1], 180) {
		t.Errorf("mirrored arc: %v %v", a.Center, a.Angle)
	}
}

func TestEllipseAffine(t *testing.T) {
	for _, tr := range []entity.Transform{
		{Angle: 30, ScaleX: 2, ScaleY: -0.5, Offset: []float64{1, 2, 0}},
		{Angle: -75, ScaleX: 0.3, ScaleY: 3},
		entity.MirrorBy([]float64{1, 1}, []float64{3, 2}),
	} {
		e := entity.NewEllipse()
		e.Center = []float64{1, 1, 0}
		e.MajorAxis = []float64{2, 1, 0}
		e.Ratio = 0.4
		e.Start, e.End = 0.5, 4
		mid := (e.Start + e.End) / 2
		want := [][]float64{tr.Point(ellipsePoint(e, e.Start)), tr.Point(ellipsePoint(e, mid)), tr.Point(ellipsePoint(e, e.End))}
		if err := e.ApplyTransform(tr); err != nil {
			t.Fatal(err)
		}
		if e.Ratio > 1 || e.Ratio <= 0 {
			t.Errorf("ratio %v", e.Ratio)
		}
		end := e.End
		if end < e.Start {
			end += 2 * math.Pi
		}
		got := [][]float64{ellipsePoint(e, e.Start), ellipsePoint(e, (e.Start+end)/2), ellipsePoint(e, end)}
		if tr.Mirrored() {
			got[0], got[2] = got[2], got[0]
		}
		for i := range want {
			if !near(got[i], want[i]) {
				t.Errorf("%+v: point %d: %v, want %v", tr, i, got[i], want[i])
			}
		}
	}
}

func TestScaleXYZ(t *testing.T) {
	d := NewDrawing()
	c, _ := d.Circle(1, 1, 0, 1)
	l, _ := d.LwPolyline(false, []float64{0, 0}, []float64{1, 0})
	l.Bulges[0] = 0.5
	if err := d.ScaleXYZ([]float64{0, 0, 0}, 2, 1, 1, c); err != nil {
		t.Fatal(err)
	}
	el, ok := d.Entities()[0].(*entity.Ellipse)
	if !ok || !near(el.Center, []float64{2, 1, 0}) || !near(el.MajorAxis, []float64{2, 0, 0}) || !cmpF64(el.Ratio, 0.5) {
		t.Fatalf("circle -> ellipse: %+v", d.Entities()[0])
	}
	err := d.ScaleXYZ(nil, 2, 1, 1, l)
	if !errors.Is(err, entity.ErrNonUniform) || !errors.Is(err, entity.ErrTransform) {
		t.Errorf("polyline with bulges: %v", err)
	}
	if err := d.Mirror([]float64{0, 0}, []float64{1, 0}, l); err != nil || l.Bulges[0] != -0.5 {
		t.Errorf("mirrored bulge: %v %v", l.Bulges, err)
	}
	checkStructure(t, d)
}

func TestMirrorText(t *testing.T) {
	d := NewDrawing()
	tx, _ := d.Text("A", 1, 1, 0, 1)
	// about the Y axis: readable, right aligned at the mirrored point
	d.Mirror([]float64{0, 0}, []float64{0, 1}, tx)
	if !cmpF64(tx.Rotation, 0) || tx.HorizontalFlag != 2 || !near(tx.Coord2, []float64{-1, 1, 0}) || tx.GenFlag != 0 {
		t.Errorf("mirrored about Y: rot %v h %d %v", tx.Rotation, tx.HorizontalFlag, tx.Coord2)
	}
	tx2, _ := d.Text("B", 1, 1, 0, 1)
	d.Mirror([]float64{0, 0}, []float64{1, 0}, tx2)
	if !cmpF64(tx2.Rotation, 0) || tx2.VerticalFlag != 3 || !near(tx2.Coord2, []float64{1, -1, 0}) {
		t.Errorf("mirrored about X: rot %v v %d %v", tx2.Rotation, tx2.VerticalFlag, tx2.Coord2)
	}
	// exact mirror: upside down
	tx3, _ := d.Text("C", 1, 1, 0, 1)
	tr := entity.MirrorBy([]float64{0, 0}, []float64{0, 1})
	tr.MirrorText = true
	d.TransformEntities(tr, tx3)
	if !cmpF64(tx3.Rotation, 180) || tx3.GenFlag != 4 || !near(tx3.Coord1, []float64{-1, 1, 0}) {
		t.Errorf("exact mirror: rot %v gen %d %v", tx3.Rotation, tx3.GenFlag, tx3.Coord1)
	}
	// non-uniform along the text: width factor
	tx4, _ := d.Text("D", 0, 0, 0, 1)
	if err := d.ScaleXYZ(nil, 2, 1, 1, tx4); err != nil || !cmpF64(tx4.WidthFactor, 2) || !cmpF64(tx4.Height, 1) {
		t.Errorf("text scaled along X: %v %v %v", err, tx4.WidthFactor, tx4.Height)
	}
	tx5, _ := d.Text("E", 0, 0, 0, 1)
	tx5.Rotation = 30
	if err := d.ScaleXYZ(nil, 2, 1, 1, tx5); !errors.Is(err, entity.ErrNonUniform) {
		t.Errorf("rotated text: %v", err)
	}
	m, _ := d.MText("M", 1, 1, 0, 1, 10)
	d.Mirror([]float64{0, 0}, []float64{0, 1}, m)
	if m.Attachment != entity.MTEXT_TOP_RIGHT || !cmpF64(m.Rotation, 0) {
		t.Errorf("mirrored mtext: %d %v", m.Attachment, m.Rotation)
	}
}

func TestMirrorInsert(t *testing.T) {
	d := NewDrawing()
	d.BeginBlock("B", 0, 0, 0)
	d.Line(0, 0, 0, 1, 0, 0)
	d.EndBlock()
	check := func(ins *entity.Insert, tr entity.Transform) {
		t.Helper()
		var want [][]float64
		for _, p := range [][]float64{{1, 0, 0}, {0, 1, 0}, {2, 3, 0}} {
			want = append(want, tr.Point(ins.Transform(p, []float64{0, 0, 0}, 0, 0)))
		}
		if err := d.TransformEntities(tr, ins); err != nil {
			t.Fatal(err)
		}
		for i, p := range [][]float64{{1, 0, 0}, {0, 1, 0}, {2, 3, 0}} {
			if got := ins.Transform(p, []float64{0, 0, 0}, 0, 0); !near(got, want[i]) {
				t.Errorf("%+v: %v -> %v, want %v (scale %v rot %v)", tr, p, got, want[i], ins.Scale, ins.Rotation)
			}
		}
	}
	ins, _ := d.Insert("B", 1, 2, 0)
	ins.Rotation = 30
	check(ins, entity.MirrorBy([]float64{0, 0}, []float64{1, 0}))
	if ins.Scale[1] >= 0 {
		t.Errorf("mirrored insert: %v", ins.Scale)
	}
	ins2, _ := d.Insert("B", 1, 2, 0)
	ins2.Rotation = 90
	check(ins2, entity.ScaleXYZ(nil, 2, 1, 1))
	ins3, _ := d.Insert("B", 1, 2, 0)
	ins3.Rotation = 30
	if err := d.ScaleXYZ(nil, 2, 1, 1, ins3); !errors.Is(err, entity.ErrNonUniform) {
		t.Errorf("rotated insert: %v", err)
	}
}

// hatch arc edges: clockwise edges store 360 - angle
func TestMirrorHatchArcEdge(t *testing.T) {
	edgePoint := func(a *entity.HatchArc, angle float64) []float64 {
		if !a.CCW {
			angle = -angle
		}
		r := angle * math.Pi / 180
		return []float64{a.Center[0] + a.Radius*math.Cos(r), a.Center[1] + a.Radius*math.Sin(r)}
	}
	for _, ccw := range []bool{true, false} {
		for _, tr := range []entity.Transform{entity.Rotate(nil, 40), entity.MirrorBy([]float64{0, 0}, []float64{1, 2})} {
			a := &entity.HatchArc{Center: []float64{1, 2}, Radius: 2, Start: 10, End: 100, CCW: ccw}
			h := entity.NewHatch()
			h.AddEdgePath(a)
			start, end := tr.Point(edgePoint(a, a.Start)), tr.Point(edgePoint(a, a.End))
			if err := h.ApplyTransform(tr); err != nil {
				t.Fatal(err)
			}
			if !near(edgePoint(a, a.Start), start) || !near(edgePoint(a, a.End), end) {
				t.Errorf("ccw %v %+v: %v %v, want %v %v", ccw, tr, edgePoint(a, a.Start), edgePoint(a, a.End), start, end)
			}
			if a.CCW == ccw == tr.Mirrored() {
				t.Errorf("direction: ccw %v -> %v", ccw, a.CCW)
			}
		}
	}
}

func TestRotateViewport(t *testing.T) {
	d := NewDrawing()
	d.PaperSpace()
	v, _ := d.Viewport([]float64{10, 0, 0}, 40, 20, []float64{0, 0}, 100)
	if err := d.Rotate([]float64{0, 0, 0}, 90, v); err != nil {
		t.Fatal(err)
	}
	if !near(v.Center, []float64{0, 10, 0}) || !cmpF64(v.Width, 20) || !cmpF64(v.Height, 40) {
		t.Errorf("viewport: %v %v %v", v.Center, v.Width, v.Height)
	}
}
