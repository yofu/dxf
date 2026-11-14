package geometry

import (
	"math"
	"sort"
)

// Vec returns the vector b - a (3D; missing coordinates are 0).
func Vec(a, b []float64) []float64 {
	return []float64{c(b, 0) - c(a, 0), c(b, 1) - c(a, 1), c(b, 2) - c(a, 2)}
}

func c(v []float64, i int) float64 {
	if i < len(v) {
		return v[i]
	}
	return 0
}

// Distance returns the distance between a and b (3D).
func Distance(a, b []float64) float64 {
	d := Vec(a, b)
	return math.Sqrt(d[0]*d[0] + d[1]*d[1] + d[2]*d[2])
}

// BulgeToArc returns the arc of a bulge between start and end (XY): its
// center, radius and start and end angles (radians, counterclockwise).
// ok is false for a bulge 0 or equal points.
func BulgeToArc(start, end []float64, bulge float64) (center []float64, radius, a1, a2 float64, ok bool) {
	dx, dy := c(end, 0)-c(start, 0), c(end, 1)-c(start, 1)
	chord := math.Hypot(dx, dy)
	if bulge == 0 || chord == 0 {
		return nil, 0, 0, 0, false
	}
	offset := chord / 2 * (1 - bulge*bulge) / (2 * bulge)
	radius = chord / 4 * (1 + bulge*bulge) / math.Abs(bulge)
	center = []float64{(c(start, 0)+c(end, 0))/2 - offset*dy/chord, (c(start, 1)+c(end, 1))/2 + offset*dx/chord}
	a1 = math.Atan2(c(start, 1)-center[1], c(start, 0)-center[0])
	a2 = math.Atan2(c(end, 1)-center[1], c(end, 0)-center[0])
	if bulge < 0 {
		a1, a2 = a2, a1
	}
	return center, radius, a1, a2, true
}

// ArcToBulge returns the bulge of the arc with center and radius from
// angle a1 to a2 (radians, counterclockwise), and its start and end points.
func ArcToBulge(center []float64, radius, a1, a2 float64) (start, end []float64, bulge float64) {
	sweep := math.Mod(a2-a1+4*math.Pi, 2*math.Pi)
	start = []float64{c(center, 0) + radius*math.Cos(a1), c(center, 1) + radius*math.Sin(a1)}
	end = []float64{c(center, 0) + radius*math.Cos(a2), c(center, 1) + radius*math.Sin(a2)}
	return start, end, math.Tan(sweep / 4)
}

// LineIntersection returns the intersection of the lines a1-a2 and b1-b2
// (XY); segments limits the result to both segments.
func LineIntersection(a1, a2, b1, b2 []float64, segments bool) ([]float64, bool) {
	dx1, dy1 := c(a2, 0)-c(a1, 0), c(a2, 1)-c(a1, 1)
	dx2, dy2 := c(b2, 0)-c(b1, 0), c(b2, 1)-c(b1, 1)
	den := dx1*dy2 - dy1*dx2
	if math.Abs(den) < 1e-12*math.Hypot(dx1, dy1)*math.Hypot(dx2, dy2) {
		return nil, false
	}
	t := ((c(b1, 0)-c(a1, 0))*dy2 - (c(b1, 1)-c(a1, 1))*dx2) / den
	u := ((c(b1, 0)-c(a1, 0))*dy1 - (c(b1, 1)-c(a1, 1))*dx1) / den
	if segments && (t < -1e-12 || t > 1+1e-12 || u < -1e-12 || u > 1+1e-12) {
		return nil, false
	}
	return []float64{c(a1, 0) + t*dx1, c(a1, 1) + t*dy1}, true
}

// LineCircleIntersections returns the intersections of the line p1-p2
// (infinite) and a circle (XY).
func LineCircleIntersections(p1, p2, center []float64, radius float64) [][]float64 {
	dx, dy := c(p2, 0)-c(p1, 0), c(p2, 1)-c(p1, 1)
	fx, fy := c(p1, 0)-c(center, 0), c(p1, 1)-c(center, 1)
	a := dx*dx + dy*dy
	b := 2 * (fx*dx + fy*dy)
	cc := fx*fx + fy*fy - radius*radius
	disc := b*b - 4*a*cc
	if a == 0 || disc < -1e-12 {
		return nil
	}
	if disc < 0 {
		disc = 0
	}
	sq := math.Sqrt(disc)
	var out [][]float64
	for _, t := range []float64{(-b - sq) / (2 * a), (-b + sq) / (2 * a)} {
		out = append(out, []float64{c(p1, 0) + t*dx, c(p1, 1) + t*dy})
		if sq == 0 {
			break
		}
	}
	return out
}

// CircleIntersections returns the intersections of two circles (XY).
func CircleIntersections(c1 []float64, r1 float64, c2 []float64, r2 float64) [][]float64 {
	d := math.Hypot(c(c2, 0)-c(c1, 0), c(c2, 1)-c(c1, 1))
	if d == 0 || d > r1+r2+1e-12 || d < math.Abs(r1-r2)-1e-12 {
		return nil
	}
	a := (r1*r1 - r2*r2 + d*d) / (2 * d)
	h2 := r1*r1 - a*a
	if h2 < 0 {
		h2 = 0
	}
	h := math.Sqrt(h2)
	ux, uy := (c(c2, 0)-c(c1, 0))/d, (c(c2, 1)-c(c1, 1))/d
	px, py := c(c1, 0)+a*ux, c(c1, 1)+a*uy
	if h == 0 {
		return [][]float64{{px, py}}
	}
	return [][]float64{{px - h*uy, py + h*ux}, {px + h*uy, py - h*ux}}
}

// DistancePointLine returns the distance of p to the segment a-b (XY).
func DistancePointLine(p, a, b []float64) float64 {
	dx, dy := c(b, 0)-c(a, 0), c(b, 1)-c(a, 1)
	l2 := dx*dx + dy*dy
	t := 0.0
	if l2 > 0 {
		t = math.Max(0, math.Min(1, ((c(p, 0)-c(a, 0))*dx+(c(p, 1)-c(a, 1))*dy)/l2))
	}
	return math.Hypot(c(p, 0)-c(a, 0)-t*dx, c(p, 1)-c(a, 1)-t*dy)
}

// PolygonArea returns the signed area of a polygon (XY; positive:
// counterclockwise).
func PolygonArea(pts [][]float64) float64 {
	area := 0.0
	for i := range pts {
		j := (i + 1) % len(pts)
		area += c(pts[i], 0)*c(pts[j], 1) - c(pts[j], 0)*c(pts[i], 1)
	}
	return area / 2
}

// IsClockwise reports whether a polygon is clockwise (XY).
func IsClockwise(pts [][]float64) bool { return PolygonArea(pts) < 0 }

// PointInPolygon reports whether p is inside the polygon (XY, even-odd
// rule; points on the boundary are inside).
func PointInPolygon(p []float64, pts [][]float64) bool {
	x, y := c(p, 0), c(p, 1)
	in := false
	for i := range pts {
		j := (i + len(pts) - 1) % len(pts)
		xi, yi, xj, yj := c(pts[i], 0), c(pts[i], 1), c(pts[j], 0), c(pts[j], 1)
		if DistancePointLine(p, pts[i], pts[j]) < 1e-12 {
			return true
		}
		if (yi > y) != (yj > y) && x < (xj-xi)*(y-yi)/(yj-yi)+xi {
			in = !in
		}
	}
	return in
}

// ConvexHull returns the convex hull of points (XY, counterclockwise).
func ConvexHull(pts [][]float64) [][]float64 {
	p := append([][]float64{}, pts...)
	sort.Slice(p, func(i, j int) bool {
		return c(p[i], 0) < c(p[j], 0) || (c(p[i], 0) == c(p[j], 0) && c(p[i], 1) < c(p[j], 1))
	})
	if len(p) < 3 {
		return p
	}
	cross := func(o, a, b []float64) float64 {
		return (c(a, 0)-c(o, 0))*(c(b, 1)-c(o, 1)) - (c(a, 1)-c(o, 1))*(c(b, 0)-c(o, 0))
	}
	var hull [][]float64
	for pass := 0; pass < 2; pass++ {
		start := len(hull)
		for _, q := range p {
			for len(hull) >= start+2 && cross(hull[len(hull)-2], hull[len(hull)-1], q) <= 0 {
				hull = hull[:len(hull)-1]
			}
			hull = append(hull, q)
		}
		hull = hull[:len(hull)-1]
		for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
			p[i], p[j] = p[j], p[i]
		}
	}
	return hull
}

// Bezier3 returns the point of the cubic Bezier curve p0..p3 at t (3D).
func Bezier3(p0, p1, p2, p3 []float64, t float64) []float64 {
	u := 1 - t
	a, b, cc, d := u*u*u, 3*u*u*t, 3*u*t*t, t*t*t
	out := make([]float64, 3)
	for i := 0; i < 3; i++ {
		out[i] = a*c(p0, i) + b*c(p1, i) + cc*c(p2, i) + d*c(p3, i)
	}
	return out
}

// SplitBezier3 splits a cubic Bezier curve at t (de Casteljau).
func SplitBezier3(p0, p1, p2, p3 []float64, t float64) ([4][]float64, [4][]float64) {
	lerp := func(a, b []float64) []float64 {
		return []float64{c(a, 0) + (c(b, 0)-c(a, 0))*t, c(a, 1) + (c(b, 1)-c(a, 1))*t, c(a, 2) + (c(b, 2)-c(a, 2))*t}
	}
	a, b, cc := lerp(p0, p1), lerp(p1, p2), lerp(p2, p3)
	d, e := lerp(a, b), lerp(b, cc)
	f := lerp(d, e)
	return [4][]float64{p0, a, d, f}, [4][]float64{f, e, cc, p3}
}

// ArcBeziers returns cubic Bezier curves (4 points each) approximating the
// arc of center and radius from a1 to a2 (radians, counterclockwise), at
// most 90 degrees each (z of center).
func ArcBeziers(center []float64, radius, a1, a2 float64) [][4][]float64 {
	sweep := a2 - a1
	n := int(math.Ceil(math.Abs(sweep) / (math.Pi / 2)))
	if n < 1 {
		n = 1
	}
	step := sweep / float64(n)
	k := 4.0 / 3.0 * math.Tan(step/4) * radius
	z := c(center, 2)
	var out [][4][]float64
	for i := 0; i < n; i++ {
		s, e := a1+float64(i)*step, a1+float64(i+1)*step
		p0 := []float64{c(center, 0) + radius*math.Cos(s), c(center, 1) + radius*math.Sin(s), z}
		p3 := []float64{c(center, 0) + radius*math.Cos(e), c(center, 1) + radius*math.Sin(e), z}
		p1 := []float64{p0[0] - k*math.Sin(s), p0[1] + k*math.Cos(s), z}
		p2 := []float64{p3[0] + k*math.Sin(e), p3[1] - k*math.Cos(e), z}
		out = append(out, [4][]float64{p0, p1, p2, p3})
	}
	return out
}

// BSplinePoint returns the point of a B-spline at u (de Boor): degree,
// knots, control points and weights (nil: non-rational).
func BSplinePoint(degree int, knots []float64, ctrl [][]float64, weights []float64, u float64) []float64 {
	n := len(ctrl)
	if n == 0 {
		return nil
	}
	p := degree
	if p > n-1 {
		p = n - 1
	}
	k := p
	for k+1 < n && knots[k+1] <= u {
		k++
	}
	d := make([][]float64, p+1)
	for j := 0; j <= p; j++ {
		w := 1.0
		if weights != nil && j+k-p < len(weights) {
			w = weights[j+k-p]
		}
		q := ctrl[j+k-p]
		d[j] = []float64{c(q, 0) * w, c(q, 1) * w, c(q, 2) * w, w}
	}
	for r := 1; r <= p; r++ {
		for j := p; j >= r; j-- {
			den := knots[j+1+k-r] - knots[j+k-p]
			a := 0.0
			if den != 0 {
				a = (u - knots[j+k-p]) / den
			}
			for i := 0; i < 4; i++ {
				d[j][i] = (1-a)*d[j-1][i] + a*d[j][i]
			}
		}
	}
	w := d[p][3]
	if w == 0 {
		w = 1
	}
	return []float64{d[p][0] / w, d[p][1] / w, d[p][2] / w}
}

// DistancePointLine3 returns the distance of p to the segment a-b (3D).
func DistancePointLine3(p, a, b []float64) float64 {
	d := Vec(a, b)
	l2 := d[0]*d[0] + d[1]*d[1] + d[2]*d[2]
	t := 0.0
	if l2 > 0 {
		ap := Vec(a, p)
		t = math.Max(0, math.Min(1, (ap[0]*d[0]+ap[1]*d[1]+ap[2]*d[2])/l2))
	}
	q := []float64{c(a, 0) + t*d[0], c(a, 1) + t*d[1], c(a, 2) + t*d[2]}
	return Distance(p, q)
}
