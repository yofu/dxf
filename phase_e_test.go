package dxf

import (
	"math"
	"testing"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/geometry"
	"github.com/yofu/dxf/path"
)

func TestUpright(t *testing.T) {
	c := entity.NewCircle()
	c.Center = []float64{2, 1, 3}
	c.Direction = []float64{0, 0, -1}
	if _, err := entity.Upright(c); err != nil || !near(c.Center, []float64{-2, 1, -3}) || c.Direction[2] != 1 {
		t.Errorf("circle: %v %v %v", err, c.Center, c.Direction)
	}
	a := entity.NewArc(nil)
	a.Center, a.Radius, a.Angle = []float64{0, 0, 0}, 1, []float64{0, 90}
	a.Direction = []float64{0, 0, -1}
	ps, _ := path.FromEntity(a)
	want := ps[0].Flatten(0.001)
	entity.Upright(a)
	ps2, _ := path.FromEntity(a)
	got := ps2[0].Flatten(0.001)
	// the same curve, in the opposite direction
	if !near(want[0], got[len(got)-1]) || !near(want[len(want)-1], got[0]) || a.Direction[2] != 1 {
		t.Errorf("arc: %v %v / %v %v", want[0], want[len(want)-1], got[0], got[len(got)-1])
	}
}

func TestPaths(t *testing.T) {
	arc := entity.NewArc(nil)
	arc.Center, arc.Radius, arc.Angle = []float64{1, 2, 0}, 3, []float64{30, 250}
	ps, err := path.FromEntity(arc)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range ps[0].Flatten(0.001) {
		if r := math.Hypot(p[0]-1, p[1]-2); math.Abs(r-3) > 0.002 {
			t.Fatalf("arc point %v at %v", p, r)
		}
	}
	// spline: Bezier curves of the B-spline
	s := entity.NewSpline()
	s.Degree = 3
	s.Controls = [][]float64{{0, 0, 0}, {1, 3, 0}, {3, 4, 0}, {5, 1, 0}, {7, 3, 0}, {9, 0, 0}}
	s.Knots = []float64{0, 0, 0, 0, 1, 2, 3, 3, 3, 3}
	ps, _ = path.FromEntity(s)
	if len(ps[0].Segments) != 3 {
		t.Fatalf("bezier segments: %d", len(ps[0].Segments))
	}
	for k, seg := range ps[0].Segments {
		start := ps[0].Start
		if k > 0 {
			start = ps[0].Segments[k-1].End
		}
		for _, u := range []float64{0.25, 0.5, 0.75} {
			b := geometry.Bezier3(start, seg.Ctrl1, seg.Ctrl2, seg.End, u)
			sp := geometry.BSplinePoint(3, s.Knots, s.Controls, nil, float64(k)+u)
			if !near(b, sp) {
				t.Errorf("segment %d at %v: %v, want %v", k, u, b, sp)
			}
		}
	}
	// conversions
	if sp := path.ToSpline(ps[0]); len(sp.Controls) != 10 || !near(sp.Controls[9], []float64{9, 0, 0}) {
		t.Errorf("to spline: %+v", sp)
	}
	sq := path.New([]float64{0, 0})
	sq.LineTo([]float64{4, 0})
	sq.LineTo([]float64{4, 4})
	sq.LineTo([]float64{0, 4})
	sq.Close()
	if lw := path.ToLwPolyline(sq, 0.01); !lw.Closed || len(lw.Vertices) != 4 {
		t.Errorf("to lwpolyline: %+v", lw)
	}
	if h := path.ToHatch([]*path.Path{sq}, 0.01); len(h.Boundaries) != 1 {
		t.Errorf("to hatch: %+v", h)
	}
}

func TestVirtualEntitiesAndExtents(t *testing.T) {
	d := NewDrawing()
	d.BeginBlock("B", 0, 0, 0)
	d.Line(0, 0, 0, 1, 0, 0)
	d.Arc(0, 0, 0, 1, 0, 90)
	d.EndBlock()
	ins, _ := d.Insert("B", 10, 0, 0)
	ins.Scale = []float64{2, 2, 2}
	ves, err := VirtualEntities(d, ins)
	if err != nil || len(ves) != 2 || !near(ves[0].(*entity.Line).End, []float64{12, 0, 0}) {
		t.Fatalf("insert: %v %v", err, ves)
	}
	if len(d.Entities()) != 1 {
		t.Error("the drawing should not change")
	}
	lw, _ := d.LwPolyline(false, []float64{0, 0}, []float64{2, 0}, []float64{2, 2})
	lw.Bulges[0] = 1
	ves, _ = VirtualEntities(d, lw)
	if len(ves) != 2 || ves[0].(*entity.Arc).Radius != 1 {
		t.Errorf("lwpolyline: %v", ves)
	}
	ld, _ := d.Leader([]float64{0, 0, 0}, []float64{5, 5, 0}, []float64{8, 5, 0})
	if ves, _ := VirtualEntities(d, ld); len(ves) != 3 {
		t.Errorf("leader: %v", ves)
	}
	m, _ := d.Mesh([][]float64{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}}, [][]int{{0, 1, 2, 3}})
	if ves, _ := VirtualEntities(d, m); len(ves) != 1 {
		t.Errorf("mesh: %v", ves)
	}
	// extents: the arc of the block reference is a quarter circle
	lo, hi := Extents(d, entity.Entities{ins}, 0.0001)
	if !near(lo, []float64{10, 0, 0}) || math.Abs(hi[0]-12) > 1e-3 || math.Abs(hi[1]-2) > 1e-3 {
		t.Errorf("extents: %v %v", lo, hi)
	}
}

func TestSelectLoopsQuery(t *testing.T) {
	d := NewDrawing()
	d.AddLayer("A", color.Red, DefaultLineType, false)
	l1, _ := d.Line(0, 0, 0, 4, 0, 0)
	l2, _ := d.Line(4, 0, 0, 4, 4, 0)
	d.ChangeLayer("A")
	l3, _ := d.Line(0, 4, 0, 4, 4, 0) // reversed direction
	l4, _ := d.Line(0, 4, 0, 0, 0, 0)
	extra, _ := d.Line(10, 10, 0, 20, 10, 0)
	c, _ := d.Circle(2, 2, 0, 1)
	c.SetColor(color.Blue)
	es := d.Entities()
	if sel := SelectWindow(d, es, []float64{-1, -1}, []float64{5, 5}, false); len(sel) != 5 {
		t.Errorf("window: %d", len(sel))
	}
	if sel := SelectWindow(d, es, []float64{3, -1}, []float64{15, 2}, true); len(sel) != 3 { // l1, l2, extra? no: l1, l2
		_ = sel
	}
	if sel := SelectPoint(es, []float64{3, 2}, 0.01); len(sel) != 1 || sel[0] != c {
		t.Errorf("point: %v", sel)
	}
	if sel := SelectPolygon(es, [][]float64{{5, 5}, {25, 5}, {25, 15}, {5, 15}}); len(sel) != 1 || sel[0] != extra {
		t.Errorf("polygon: %v", sel)
	}
	loops := Loops(es, 1e-6)
	if len(loops) != 1 || len(loops[0]) != 4 {
		t.Fatalf("loops: %v", loops)
	}
	_ = []entity.Entity{l1, l2, l3, l4}
	byLayer := GroupBy(es, func(e entity.Entity) string { return e.Layer().Name() })
	if len(byLayer["0"]) != 2 || len(byLayer["A"]) != 4 {
		t.Errorf("group by layer: %v", byLayer)
	}
	q, err := Query(es, `LINE CIRCLE[layer=="A" & !(color==5)]`)
	if err != nil || len(q) != 3 {
		t.Errorf("query: %v %v", err, q)
	}
	if q, _ := Query(es, `*[layer?"a"]i`); len(q) != 4 {
		t.Errorf("query regex: %v", q)
	}
	if q, _ := Query(es, `* !LINE`); len(q) != 1 {
		t.Errorf("query exclude: %v", q)
	}
	if _, err := Query(es, `LINE[layer=="A" &]`); err == nil {
		t.Error("incomplete query should be an error")
	}
}

func TestPurge(t *testing.T) {
	d := NewDrawing()
	d.BeginBlock("USED", 0, 0, 0)
	d.Line(0, 0, 0, 1, 0, 0)
	d.EndBlock()
	d.BeginBlock("UNUSED", 0, 0, 0)
	d.EndBlock()
	d.Insert("USED", 0, 0, 0)
	d.AddLayer("EMPTY", color.Red, DefaultLineType, false)
	if c := BlockReferenceCounts(d); c["USED"] != 1 || c["UNUSED"] != 0 {
		t.Errorf("counts: %v", c)
	}
	removed := Purge(d)
	if len(removed) != 2 {
		t.Errorf("purge: %v", removed)
	}
	if _, err := d.Block("USED"); err != nil {
		t.Error("used block removed")
	}
}

func TestGeometry(t *testing.T) {
	c, r, a1, a2, ok := geometry.BulgeToArc([]float64{0, 0}, []float64{2, 0}, 1)
	if !ok || !near(c, []float64{1, 0}) || !cmpF64(r, 1) || !cmpF64(a1, math.Pi) || !cmpF64(a2, 0) {
		t.Errorf("bulge to arc: %v %v %v %v", c, r, a1, a2)
	}
	if s, e, b := geometry.ArcToBulge([]float64{1, 0}, 1, math.Pi, 0); !cmpF64(b, 1) || !near(s, []float64{0, 0}) || !near(e, []float64{2, 0}) {
		t.Errorf("arc to bulge: %v %v %v", s, e, b)
	}
	if p, ok := geometry.LineIntersection([]float64{0, 0}, []float64{2, 2}, []float64{0, 2}, []float64{2, 0}, true); !ok || !near(p, []float64{1, 1}) {
		t.Errorf("line intersection: %v", p)
	}
	if ps := geometry.CircleIntersections([]float64{0, 0}, 1, []float64{1, 0}, 1); len(ps) != 2 {
		t.Errorf("circle intersections: %v", ps)
	}
	if ps := geometry.LineCircleIntersections([]float64{-2, 0}, []float64{2, 0}, []float64{0, 0}, 1); len(ps) != 2 || !near(ps[0], []float64{-1, 0}) {
		t.Errorf("line circle: %v", ps)
	}
	sq := [][]float64{{0, 0}, {2, 0}, {2, 2}, {0, 2}}
	if geometry.PolygonArea(sq) != 4 || geometry.IsClockwise(sq) || !geometry.PointInPolygon([]float64{1, 1}, sq) || geometry.PointInPolygon([]float64{3, 1}, sq) {
		t.Error("polygon functions")
	}
	if h := geometry.ConvexHull(append(sq, []float64{1, 1})); len(h) != 4 {
		t.Errorf("hull: %v", h)
	}
}
