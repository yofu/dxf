// Package path converts entities to paths of lines and cubic Bezier
// curves (WCS), flattens and transforms them, and converts them back to
// entities (as ezdxf's path module).
package path

import (
	"errors"
	"fmt"
	"math"

	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/geometry"
)

// ErrUnsupported is returned for entities without a path.
var ErrUnsupported = errors.New("entity has no path")

// Segment is a line (Ctrl1 and Ctrl2 nil) or a cubic Bezier curve to End.
type Segment struct {
	Ctrl1, Ctrl2, End []float64
}

// IsCurve reports whether the segment is a cubic Bezier curve.
func (s Segment) IsCurve() bool { return s.Ctrl1 != nil }

// Path is a sequence of connected segments from Start (WCS).
type Path struct {
	Start    []float64
	Segments []Segment
}

// New returns a path from start.
func New(start []float64) *Path { return &Path{Start: p3(start)} }

func p3(p []float64) []float64 {
	q := []float64{0, 0, 0}
	copy(q, p)
	return q
}

// End returns the end point of the path.
func (p *Path) End() []float64 {
	if len(p.Segments) == 0 {
		return p.Start
	}
	return p.Segments[len(p.Segments)-1].End
}

// LineTo adds a line to q.
func (p *Path) LineTo(q []float64) { p.Segments = append(p.Segments, Segment{End: p3(q)}) }

// CurveTo adds a cubic Bezier curve to q with control points c1 and c2.
func (p *Path) CurveTo(c1, c2, q []float64) {
	p.Segments = append(p.Segments, Segment{Ctrl1: p3(c1), Ctrl2: p3(c2), End: p3(q)})
}

// IsClosed reports whether the path ends at its start.
func (p *Path) IsClosed() bool { return geometry.Distance(p.Start, p.End()) < 1e-9 }

// Close adds a line to the start if the path is not closed.
func (p *Path) Close() {
	if !p.IsClosed() {
		p.LineTo(p.Start)
	}
}

// Flatten returns the points of the path: curves are approximated by lines
// whose distance to the curve is at most tolerance.
func (p *Path) Flatten(tolerance float64) [][]float64 {
	if tolerance <= 0 {
		tolerance = 0.01
	}
	pts := [][]float64{p.Start}
	cur := p.Start
	for _, s := range p.Segments {
		if !s.IsCurve() {
			pts = append(pts, s.End)
		} else {
			pts = append(pts, flattenBezier(cur, s.Ctrl1, s.Ctrl2, s.End, tolerance, 0)...)
		}
		cur = s.End
	}
	return pts
}

// flattenBezier subdivides a cubic Bezier curve until it is flat (the
// control points are within tolerance of the chord); returns the points
// after p0.
func flattenBezier(p0, p1, p2, p3 []float64, tol float64, depth int) [][]float64 {
	if depth > 16 || (geometry.DistancePointLine3(p1, p0, p3) <= tol && geometry.DistancePointLine3(p2, p0, p3) <= tol) {
		return [][]float64{p3}
	}
	a, b := geometry.SplitBezier3(p0, p1, p2, p3, 0.5)
	return append(flattenBezier(a[0], a[1], a[2], a[3], tol, depth+1), flattenBezier(b[0], b[1], b[2], b[3], tol, depth+1)...)
}

// Transform transforms the path (points and control points: Bezier curves
// are affine invariant).
func (p *Path) Transform(t entity.Transform) {
	p.Start = t.Point(p.Start)
	for i := range p.Segments {
		s := &p.Segments[i]
		s.End = t.Point(s.End)
		if s.IsCurve() {
			s.Ctrl1, s.Ctrl2 = t.Point(s.Ctrl1), t.Point(s.Ctrl2)
		}
	}
}

// BBox returns the bounding box of the path (curves flattened with
// tolerance).
func (p *Path) BBox(tolerance float64) ([]float64, []float64) {
	mins := []float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxs := []float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, q := range p.Flatten(tolerance) {
		for i := 0; i < 3; i++ {
			mins[i] = math.Min(mins[i], q[i])
			maxs[i] = math.Max(maxs[i], q[i])
		}
	}
	return mins, maxs
}

// Reversed returns the path in the opposite direction.
func (p *Path) Reversed() *Path {
	r := New(p.End())
	for i := len(p.Segments) - 1; i >= 0; i-- {
		s := p.Segments[i]
		start := p.Start
		if i > 0 {
			start = p.Segments[i-1].End
		}
		if s.IsCurve() {
			r.CurveTo(s.Ctrl2, s.Ctrl1, start)
		} else {
			r.LineTo(start)
		}
	}
	return r
}

// ocs returns the function mapping OCS coordinates of extrusion n to WCS.
func ocs(n []float64) func([]float64) []float64 {
	if len(n) < 3 || (n[0] == 0 && n[1] == 0 && n[2] > 0) {
		return func(p []float64) []float64 { return p3(p) }
	}
	l := math.Sqrt(n[0]*n[0] + n[1]*n[1] + n[2]*n[2])
	az := []float64{n[0] / l, n[1] / l, n[2] / l}
	ax, ay, _ := geometry.ArbitraryAxis(az)
	return func(p []float64) []float64 {
		x, y, z := c(p, 0), c(p, 1), c(p, 2)
		return []float64{ax[0]*x + ay[0]*y + az[0]*z, ax[1]*x + ay[1]*y + az[1]*z, ax[2]*x + ay[2]*y + az[2]*z}
	}
}

func c(v []float64, i int) float64 {
	if i < len(v) {
		return v[i]
	}
	return 0
}

// addArc adds the counterclockwise arc from a1 to a2 (OCS center with z,
// radians) mapped by toWCS: from a1 to a2 if ccw, otherwise from a2 back
// to a1.
func addArc(p *Path, toWCS func([]float64) []float64, center []float64, radius, a1, a2 float64, ccw bool) {
	for a2 < a1 {
		a2 += 2 * math.Pi
	}
	bs := geometry.ArcBeziers(center, radius, a1, a2)
	if !ccw {
		for i := len(bs) - 1; i >= 0; i-- {
			b := bs[i]
			p.CurveTo(toWCS(b[2]), toWCS(b[1]), toWCS(b[0]))
		}
		return
	}
	for _, b := range bs {
		p.CurveTo(toWCS(b[1]), toWCS(b[2]), toWCS(b[3]))
	}
}

// bulgePath adds the vertices with bulges (OCS) to p.
func bulgePath(toWCS func([]float64) []float64, vertices [][]float64, bulges []float64, closed bool, z float64) *Path {
	at := func(i int) []float64 { return []float64{c(vertices[i], 0), c(vertices[i], 1), z} }
	p := New(toWCS(at(0)))
	n := len(vertices)
	segs := n - 1
	if closed {
		segs = n
	}
	for i := 0; i < segs; i++ {
		s, e := at(i), at((i+1)%n)
		b := 0.0
		if i < len(bulges) {
			b = bulges[i]
		}
		center, r, a1, a2, ok := geometry.BulgeToArc(s, e, b)
		if !ok {
			p.LineTo(toWCS(e))
			continue
		}
		center = []float64{center[0], center[1], z}
		if b > 0 {
			addArc(p, toWCS, center, r, a1, a2, true)
		} else {
			addArc(p, toWCS, center, r, a1, a2, false)
		}
	}
	return p
}

// ellipsePath returns the path of an ellipse arc (WCS center, major axis,
// normal, ratio, parameters).
func ellipsePath(center, major, normal []float64, ratio, start, end float64) *Path {
	n := normal
	if len(n) < 3 {
		n = []float64{0, 0, 1}
	}
	l := math.Sqrt(n[0]*n[0] + n[1]*n[1] + n[2]*n[2])
	nz := []float64{n[0] / l, n[1] / l, n[2] / l}
	minor := []float64{
		ratio * (nz[1]*c(major, 2) - nz[2]*c(major, 1)),
		ratio * (nz[2]*c(major, 0) - nz[0]*c(major, 2)),
		ratio * (nz[0]*c(major, 1) - nz[1]*c(major, 0)),
	}
	m := func(q []float64) []float64 { // unit circle -> ellipse (affine)
		return []float64{
			c(center, 0) + q[0]*c(major, 0) + q[1]*minor[0],
			c(center, 1) + q[0]*c(major, 1) + q[1]*minor[1],
			c(center, 2) + q[0]*c(major, 2) + q[1]*minor[2],
		}
	}
	if end <= start {
		end += 2 * math.Pi
	}
	bs := geometry.ArcBeziers([]float64{0, 0, 0}, 1, start, end)
	p := New(m(bs[0][0]))
	for _, b := range bs {
		p.CurveTo(m(b[1]), m(b[2]), m(b[3]))
	}
	return p
}

// FromEntity returns the paths of an entity (WCS): LINE, ARC, CIRCLE,
// ELLIPSE, LWPOLYLINE, POLYLINE (2D, 3D; meshes as their edges), SPLINE,
// HELIX, SOLID, TRACE, 3DFACE, HATCH and MPOLYGON (boundaries), XLINE and
// RAY (a segment of length 1), MLINE (its elements), IMAGE and WIPEOUT
// (boundary) and MESH (face edges).
func FromEntity(e entity.Entity) ([]*Path, error) {
	switch x := e.(type) {
	case *entity.Line:
		p := New(x.Start)
		p.LineTo(x.End)
		return []*Path{p}, nil
	case *entity.Arc:
		center := []float64{x.Center[0], x.Center[1], c(x.Center, 2)}
		a1, a2 := x.Angle[0]*math.Pi/180, x.Angle[1]*math.Pi/180
		toWCS := ocs(x.Direction)
		p := New(toWCS([]float64{center[0] + x.Radius*math.Cos(a1), center[1] + x.Radius*math.Sin(a1), center[2]}))
		addArc(p, toWCS, center, x.Radius, a1, a2, true)
		return []*Path{p}, nil
	case *entity.Circle:
		center := []float64{x.Center[0], x.Center[1], c(x.Center, 2)}
		toWCS := ocs(x.Direction)
		p := New(toWCS([]float64{center[0] + x.Radius, center[1], center[2]}))
		addArc(p, toWCS, center, x.Radius, 0, 2*math.Pi, true)
		return []*Path{p}, nil
	case *entity.Ellipse:
		return []*Path{ellipsePath(x.Center, x.MajorAxis, x.Extrusion, x.Ratio, x.Start, x.End)}, nil
	case *entity.LwPolyline:
		if len(x.Vertices) == 0 {
			return nil, nil
		}
		return []*Path{bulgePath(ocs(x.Extrusion), x.Vertices, x.Bulges, x.Closed, x.Elevation)}, nil
	case *entity.Polyline:
		return polylinePaths(x), nil
	case *entity.Helix:
		return []*Path{splinePath(x.Spline)}, nil
	case *entity.Spline:
		if len(x.Controls) == 0 {
			if len(x.Fits) == 0 {
				return nil, nil
			}
			p := New(x.Fits[0])
			for _, q := range x.Fits[1:] {
				p.LineTo(q)
			}
			return []*Path{p}, nil
		}
		return []*Path{splinePath(x)}, nil
	case *entity.Solid:
		toWCS := ocs(x.Extrusion)
		p := New(toWCS(x.Points[0]))
		for _, i := range []int{1, 3, 2, 0} {
			p.LineTo(toWCS(x.Points[i]))
		}
		return []*Path{p}, nil
	case *entity.ThreeDFace:
		p := New(x.Points[0])
		for _, q := range x.Points[1:] {
			p.LineTo(q)
		}
		p.Close()
		return []*Path{p}, nil
	case *entity.Hatch:
		return boundaryPaths(x.Boundaries, x.Extrusion, x.Elevation), nil
	case *entity.MPolygon:
		return boundaryPaths(x.Boundaries, x.Extrusion, x.Elevation), nil
	case *entity.XLine:
		p := New(x.Point)
		p.LineTo([]float64{x.Point[0] + x.Direction[0], x.Point[1] + x.Direction[1], c(x.Point, 2) + c(x.Direction, 2)})
		return []*Path{p}, nil
	case *entity.MLine:
		var ps []*Path
		if len(x.Vertices) == 0 {
			return nil, nil
		}
		for i := range x.Vertices[0].LineParams {
			pts := x.ElementPoints(i)
			if len(pts) < 2 {
				continue
			}
			p := New(pts[0])
			for _, q := range pts[1:] {
				p.LineTo(q)
			}
			if x.IsClosed() {
				p.Close()
			}
			ps = append(ps, p)
		}
		return ps, nil
	case *entity.Image:
		cs := x.Corners()
		p := New(cs[0])
		for _, q := range cs[1:] {
			p.LineTo(q)
		}
		p.Close()
		return []*Path{p}, nil
	case *entity.Mesh:
		var ps []*Path
		for _, f := range x.Faces {
			if len(f) < 2 {
				continue
			}
			p := New(x.Vertices[f[0]])
			for _, i := range f[1:] {
				p.LineTo(x.Vertices[i])
			}
			p.Close()
			ps = append(ps, p)
		}
		return ps, nil
	}
	return nil, fmt.Errorf("%T: %w", e, ErrUnsupported)
}

func polylinePaths(x *entity.Polyline) []*Path {
	closed := x.Flag&1 != 0
	switch {
	case x.Flag&64 != 0: // polyface: faces
		var verts [][]float64
		var ps []*Path
		for _, v := range x.Vertices {
			if v.Flag&128 != 0 && v.Flag&64 == 0 {
				var pts [][]float64
				for _, i := range v.Indices {
					if i == 0 {
						continue
					}
					if i < 0 {
						i = -i
					}
					if i-1 < len(verts) {
						pts = append(pts, verts[i-1])
					}
				}
				if len(pts) > 1 {
					p := New(pts[0])
					for _, q := range pts[1:] {
						p.LineTo(q)
					}
					p.Close()
					ps = append(ps, p)
				}
			} else {
				verts = append(verts, v.Coord)
			}
		}
		return ps
	case x.Flag&16 != 0: // polygon mesh: M and N lines
		m, n := x.MeshM, x.MeshN
		if m*n > len(x.Vertices) || m < 1 || n < 1 {
			return nil
		}
		at := func(i, j int) []float64 { return x.Vertices[i*n+j].Coord }
		var ps []*Path
		for i := 0; i < m; i++ {
			p := New(at(i, 0))
			for j := 1; j < n; j++ {
				p.LineTo(at(i, j))
			}
			if x.Flag&32 != 0 {
				p.Close()
			}
			ps = append(ps, p)
		}
		for j := 0; j < n; j++ {
			p := New(at(0, j))
			for i := 1; i < m; i++ {
				p.LineTo(at(i, j))
			}
			if closed {
				p.Close()
			}
			ps = append(ps, p)
		}
		return ps
	case x.Flag&8 != 0: // 3D polyline
		if len(x.Vertices) == 0 {
			return nil
		}
		p := New(x.Vertices[0].Coord)
		for _, v := range x.Vertices[1:] {
			p.LineTo(v.Coord)
		}
		if closed {
			p.Close()
		}
		return []*Path{p}
	}
	if len(x.Vertices) == 0 {
		return nil
	}
	var vs [][]float64
	var bs []float64
	for _, v := range x.Vertices {
		vs = append(vs, v.Coord)
		bs = append(bs, v.Bulge)
	}
	return []*Path{bulgePath(ocs(x.Extrusion), vs, bs, closed, x.Elevation)}
}

// splinePath converts a spline: non-rational cubic splines exactly into
// Bezier curves (knot insertion), others by sampling 32 points per span.
func splinePath(s *entity.Spline) *Path {
	ctrl := s.Controls
	rational := false
	for _, w := range s.Weights {
		rational = rational || w != 1
	}
	knots := append([]float64{}, s.Knots...)
	if s.Degree == 3 && !rational && len(knots) == len(ctrl)+4 {
		return bezierFromBSpline(knots, ctrl)
	}
	lo, hi := knots[s.Degree], knots[len(knots)-s.Degree-1]
	var w []float64
	if rational {
		w = s.Weights
	}
	n := 32 * (len(ctrl) - s.Degree)
	if n < 32 {
		n = 32
	}
	p := New(geometry.BSplinePoint(s.Degree, knots, ctrl, w, lo))
	for i := 1; i <= n; i++ {
		p.LineTo(geometry.BSplinePoint(s.Degree, knots, ctrl, w, lo+(hi-lo)*float64(i)/float64(n)))
	}
	return p
}

// bezierFromBSpline converts a cubic B-spline (clamped or not) into Bezier
// curves by inserting its interior knots until their multiplicity is 3.
func bezierFromBSpline(knots []float64, ctrl [][]float64) *Path {
	const p = 3
	cp := make([][]float64, len(ctrl))
	for i, q := range ctrl {
		cp[i] = p3(q)
	}
	insert := func(u float64) {
		k := p
		for k+1 < len(knots)-p-1 && knots[k+1] <= u {
			k++
		}
		var nc [][]float64
		for i := 0; i <= len(cp); i++ {
			switch {
			case i <= k-p:
				nc = append(nc, cp[i])
			case i > k:
				nc = append(nc, cp[i-1])
			default:
				a := (u - knots[i]) / (knots[i+p] - knots[i])
				q := make([]float64, 3)
				for j := 0; j < 3; j++ {
					q[j] = (1-a)*cp[i-1][j] + a*cp[i][j]
				}
				nc = append(nc, q)
			}
		}
		nk := append(append(append([]float64{}, knots[:k+1]...), u), knots[k+1:]...)
		cp, knots = nc, nk
	}
	lo, hi := knots[p], knots[len(knots)-p-1]
	// clamp the ends, then make every interior knot of multiplicity 3
	for count(knots, lo) < p+1 {
		insert(lo)
	}
	for count(knots, hi) < p+1 {
		insert(hi)
	}
	var interior []float64
	for _, k := range knots {
		if k > lo && k < hi && (len(interior) == 0 || interior[len(interior)-1] != k) {
			interior = append(interior, k)
		}
	}
	for _, u := range interior {
		for count(knots, u) < p {
			insert(u)
		}
	}
	// the curve at lo is the control point at the first knot lo; then one
	// Bezier curve per span
	first := 0
	for knots[first] != lo {
		first++
	}
	path := New(cp[first])
	for sp := 0; sp <= len(interior); sp++ {
		i := first + 3*sp
		if i+3 >= len(cp) {
			break
		}
		path.CurveTo(cp[i+1], cp[i+2], cp[i+3])
	}
	return path
}

func count(knots []float64, u float64) int {
	n := 0
	for _, k := range knots {
		if k == u {
			n++
		}
	}
	return n
}

// boundaryPaths returns the boundary paths of a hatch (OCS with elevation).
func boundaryPaths(bs []*entity.HatchBoundary, extrusion []float64, z float64) []*Path {
	toWCS := ocs(extrusion)
	var ps []*Path
	for _, b := range bs {
		if b.IsPolyline() {
			if len(b.Vertices) > 0 {
				ps = append(ps, bulgePath(toWCS, b.Vertices, b.Bulges, true, z))
			}
			continue
		}
		var p *Path
		for _, e := range b.Edges {
			switch x := e.(type) {
			case *entity.HatchLine:
				s, en := []float64{x.Start[0], x.Start[1], z}, []float64{x.End[0], x.End[1], z}
				if p == nil {
					p = New(toWCS(s))
				}
				p.LineTo(toWCS(en))
			case *entity.HatchArc:
				a1, a2 := x.Start*math.Pi/180, x.End*math.Pi/180
				center := []float64{x.Center[0], x.Center[1], z}
				if !x.CCW { // stored as 360 - angle
					a1, a2 = -a1, -a2
				}
				s := []float64{center[0] + x.Radius*math.Cos(a1), center[1] + x.Radius*math.Sin(a1), z}
				if p == nil {
					p = New(toWCS(s))
				}
				if x.CCW {
					addArc(p, toWCS, center, x.Radius, a1, a2, true)
				} else {
					addArc(p, toWCS, center, x.Radius, a2, a1, false)
				}
			case *entity.HatchEllipse:
				param := func(deg float64) float64 {
					a := deg * math.Pi / 180
					return math.Atan2(math.Sin(a)/x.Ratio, math.Cos(a))
				}
				s, en := param(x.Start), param(x.End)
				if !x.CCW {
					s, en = -s, -en
				}
				major := []float64{x.MajorAxis[0], x.MajorAxis[1], 0}
				var ep *Path
				if x.CCW {
					ep = ellipsePath([]float64{x.Center[0], x.Center[1], z}, major, nil, x.Ratio, s, en)
				} else {
					ep = ellipsePath([]float64{x.Center[0], x.Center[1], z}, major, nil, x.Ratio, en, s).Reversed()
				}
				ep.Transform2(toWCS)
				if p == nil {
					p = New(ep.Start)
				}
				p.Segments = append(p.Segments, ep.Segments...)
			case *entity.HatchSpline:
				var ctrl [][]float64
				for _, q := range x.Controls {
					ctrl = append(ctrl, []float64{q[0], q[1], z})
				}
				sp := entity.NewSpline()
				sp.Degree, sp.Knots, sp.Controls = x.Degree, x.Knots, ctrl
				if x.Rational {
					sp.Weights = x.Weights
				}
				if len(ctrl) == 0 {
					continue
				}
				spp := splinePath(sp)
				spp.Transform2(toWCS)
				if p == nil {
					p = New(spp.Start)
				}
				p.Segments = append(p.Segments, spp.Segments...)
			}
		}
		if p != nil {
			ps = append(ps, p)
		}
	}
	return ps
}

// Transform2 maps all points of the path by f.
func (p *Path) Transform2(f func([]float64) []float64) {
	p.Start = f(p.Start)
	for i := range p.Segments {
		s := &p.Segments[i]
		s.End = f(s.End)
		if s.IsCurve() {
			s.Ctrl1, s.Ctrl2 = f(s.Ctrl1), f(s.Ctrl2)
		}
	}
}

// ToLwPolyline returns a LWPOLYLINE of the flattened path (in the XY plane
// at the z of its start).
func ToLwPolyline(p *Path, tolerance float64) *entity.LwPolyline {
	pts := p.Flatten(tolerance)
	closed := p.IsClosed() && len(pts) > 2
	if closed {
		pts = pts[:len(pts)-1]
	}
	l := entity.NewLwPolyline(len(pts))
	for i, q := range pts {
		l.Vertices[i] = []float64{q[0], q[1]}
	}
	l.Closed = closed
	l.Elevation = c(p.Start, 2)
	return l
}

// ToSpline returns a cubic SPLINE of the path: Bezier curves exactly, lines
// as straight Bezier curves.
func ToSpline(p *Path) *entity.Spline {
	s := entity.NewSpline()
	s.Degree = 3
	s.Controls = [][]float64{p.Start}
	s.Knots = []float64{0, 0, 0, 0}
	cur := p.Start
	for i, seg := range p.Segments {
		c1, c2 := seg.Ctrl1, seg.Ctrl2
		if !seg.IsCurve() {
			c1 = lerp(cur, seg.End, 1.0/3)
			c2 = lerp(cur, seg.End, 2.0/3)
		}
		s.Controls = append(s.Controls, c1, c2, seg.End)
		k := float64(i + 1)
		if i < len(p.Segments)-1 {
			s.Knots = append(s.Knots, k, k, k)
		} else {
			s.Knots = append(s.Knots, k, k, k, k)
		}
		cur = seg.End
	}
	if p.IsClosed() {
		s.Flag |= 1
	}
	return s
}

func lerp(a, b []float64, t float64) []float64 {
	return []float64{c(a, 0) + (c(b, 0)-c(a, 0))*t, c(a, 1) + (c(b, 1)-c(a, 1))*t, c(a, 2) + (c(b, 2)-c(a, 2))*t}
}

// ToHatch returns a solid HATCH with the closed paths as polyline
// boundaries (flattened; the first is external).
func ToHatch(paths []*Path, tolerance float64) *entity.Hatch {
	h := entity.NewHatch()
	for _, p := range paths {
		pts := p.Flatten(tolerance)
		if p.IsClosed() {
			pts = pts[:len(pts)-1]
		}
		var vs [][]float64
		for _, q := range pts {
			vs = append(vs, []float64{q[0], q[1]})
		}
		if len(vs) > 2 {
			h.AddPolylinePath(vs, nil)
			h.Elevation = c(p.Start, 2)
		}
	}
	return h
}
