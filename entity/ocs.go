package entity

import (
	"math"

	"github.com/yofu/dxf/geometry"
)

// OCS (object coordinate system) support of transformations: entities
// with an extrusion direction other than +Z are transformed in their OCS,
// as ezdxf does. For extrusions +Z and -Z every Transform can be applied;
// for other directions the scaling must be uniform in 3D.

// ocsAxes returns the axes of the OCS of extrusion n (arbitrary axis
// algorithm).
func ocsAxes(n []float64) (ax, ay, az []float64) {
	az = normalize3(n)
	ax, ay, _ = geometry.ArbitraryAxis(az)
	return ax, ay, az
}

// extrusionField returns the extrusion vector of e (nil if e has none or
// it is not used).
func extrusionField(e Entity) []float64 {
	switch x := e.(type) {
	case *Arc:
		return x.Direction
	case *Circle:
		return x.Direction
	case *LwPolyline:
		return x.Extrusion
	case *Polyline:
		if x.Flag&(8|16|64) == 0 {
			return x.Extrusion
		}
	case *Text:
		return x.Extrusion
	case *Attribute:
		return x.Extrusion
	case *AttributeDefinition:
		return x.Extrusion
	case *Insert:
		return x.Direction
	case *Solid:
		return x.Extrusion
	case *Hatch:
		return x.Extrusion
	case *MPolygon:
		return x.Extrusion
	case *Shape:
		return x.Extrusion
	case *Point:
		return x.Extrusion
	case *Ellipse:
		return x.Extrusion
	case *Spline:
		return x.Normal
	case *Helix:
		return x.Normal
	case *MText:
		return x.Extrusion
	case *Line:
		return x.Extrusion
	}
	return nil
}

// isOCSEntity reports whether the coordinates of e are in its OCS (the
// others are in WCS with the extrusion as normal).
func isOCSEntity(e Entity) bool {
	switch e.(type) {
	case *Arc, *Circle, *LwPolyline, *Polyline, *Text, *Attribute, *AttributeDefinition, *Insert, *Solid, *Hatch, *MPolygon, *Shape:
		return true
	}
	return false
}

// linear3 returns the 3x3 matrix of the linear part of t.
func (t Transform) linear3() [3][3]float64 {
	m := t.linear()
	_, _, sz := t.Factors()
	return [3][3]float64{{m[0][0], m[0][1], 0}, {m[1][0], m[1][1], 0}, {0, 0, sz}}
}

func mulMV(m [3][3]float64, v []float64) []float64 {
	return []float64{
		m[0][0]*coord(v, 0) + m[0][1]*coord(v, 1) + m[0][2]*coord(v, 2),
		m[1][0]*coord(v, 0) + m[1][1]*coord(v, 1) + m[1][2]*coord(v, 2),
		m[2][0]*coord(v, 0) + m[2][1]*coord(v, 1) + m[2][2]*coord(v, 2),
	}
}

func dot3(a, b []float64) float64 {
	return coord(a, 0)*coord(b, 0) + coord(a, 1)*coord(b, 1) + coord(a, 2)*coord(b, 2)
}

func cross3(a, b []float64) []float64 {
	return []float64{
		coord(a, 1)*coord(b, 2) - coord(a, 2)*coord(b, 1),
		coord(a, 2)*coord(b, 0) - coord(a, 0)*coord(b, 2),
		coord(a, 0)*coord(b, 1) - coord(a, 1)*coord(b, 0),
	}
}

// transformNormal returns the normal of a plane with normal n after the
// linear map m (cofactor matrix: det(m) m^-T n).
func transformNormal(m [3][3]float64, n []float64) []float64 {
	c0 := []float64{m[0][0], m[1][0], m[2][0]}
	c1 := []float64{m[0][1], m[1][1], m[2][1]}
	c2 := []float64{m[0][2], m[1][2], m[2][2]}
	// cofactor matrix columns: c1 x c2, c2 x c0, c0 x c1
	r0, r1, r2 := cross3(c1, c2), cross3(c2, c0), cross3(c0, c1)
	return normalize3([]float64{
		r0[0]*coord(n, 0) + r1[0]*coord(n, 1) + r2[0]*coord(n, 2),
		r0[1]*coord(n, 0) + r1[1]*coord(n, 1) + r2[1]*coord(n, 2),
		r0[2]*coord(n, 0) + r1[2]*coord(n, 1) + r2[2]*coord(n, 2),
	})
}

// uniform3D reports whether t scales uniformly in 3D (|sx| = |sy| = |sz|).
func (t Transform) uniform3D() bool {
	sx, sy, sz := t.Factors()
	a := math.Abs(sx)
	return math.Abs(math.Abs(sy)-a) <= 1e-12*a && math.Abs(math.Abs(sz)-a) <= 1e-12*a
}

// ocsTransform returns the transformation in the OCS of n (to the OCS of
// the new normal) equivalent to t, and the new normal.
func ocsTransform(t Transform, n []float64) (Transform, []float64, bool) {
	m := t.linear3()
	n2 := transformNormal(m, n)
	ax, ay, az := ocsAxes(n)
	bx, by, bz := ocsAxes(n2)
	// K = A'^T M A
	ma := [][]float64{mulMV(m, ax), mulMV(m, ay), mulMV(m, az)}
	k := [3][3]float64{}
	for j, col := range ma {
		k[0][j], k[1][j], k[2][j] = dot3(bx, col), dot3(by, col), dot3(bz, col)
	}
	small := 1e-9 * (math.Abs(k[0][0]) + math.Abs(k[1][1]) + math.Abs(k[2][2]))
	if math.Abs(k[0][2]) > small || math.Abs(k[1][2]) > small || math.Abs(k[2][0]) > small || math.Abs(k[2][1]) > small {
		return Transform{}, nil, false
	}
	// K_xy = R(a) diag(sx, sy): the columns must be orthogonal
	c1 := []float64{k[0][0], k[1][0]}
	c2 := []float64{k[0][1], k[1][1]}
	sx, sy := math.Hypot(c1[0], c1[1]), math.Hypot(c2[0], c2[1])
	if math.Abs(c1[0]*c2[0]+c1[1]*c2[1]) > 1e-9*sx*sy {
		return Transform{}, nil, false
	}
	if c1[0]*c2[1]-c1[1]*c2[0] < 0 {
		sy = -sy
	}
	// translation: the image of the origin, in the new OCS
	o := t.Point([]float64{0, 0, 0})
	off := []float64{dot3(bx, o), dot3(by, o), dot3(bz, o)}
	ot := Transform{Angle: math.Atan2(c1[1], c1[0]) * 180 / math.Pi, ScaleX: sx, ScaleY: sy, ScaleZ: k[2][2], Offset: off, MirrorText: t.MirrorText}
	return ot, n2, true
}

// transformOCS transforms an entity with extrusion n (not +Z).
func transformOCS(e Entity, t Transform, n []float64) (Entity, error) {
	minusZ := math.Abs(coord(n, 0)) < 1e-12 && math.Abs(coord(n, 1)) < 1e-12
	if !minusZ && !t.uniform3D() {
		return e, ErrNonUniform
	}
	if !isOCSEntity(e) {
		return transformWCSWithNormal(e, t, n)
	}
	ot, n2, ok := ocsTransform(t, n)
	if !ok {
		return e, ErrNonUniform
	}
	// the planar logic of the entity in its OCS (extrusion +Z meanwhile)
	setExtrusionAll(e, []float64{0, 0, 1})
	ne, err := TransformEntity(e, ot)
	if err != nil {
		setExtrusionAll(e, n)
		return e, err
	}
	if el, ok := ne.(*Ellipse); ok {
		// a circle became an ellipse: ellipses have WCS coordinates
		bx, by, bz := ocsAxes(n2)
		toWCS := func(p []float64) []float64 {
			return []float64{
				bx[0]*coord(p, 0) + by[0]*coord(p, 1) + bz[0]*coord(p, 2),
				bx[1]*coord(p, 0) + by[1]*coord(p, 1) + bz[1]*coord(p, 2),
				bx[2]*coord(p, 0) + by[2]*coord(p, 1) + bz[2]*coord(p, 2),
			}
		}
		el.Center = toWCS(el.Center)
		el.MajorAxis = toWCS(el.MajorAxis)
		el.Extrusion = n2
		return el, nil
	}
	setExtrusionAll(ne, n2)
	return ne, nil
}

// setExtrusionAll sets the extrusion of e and of the attributes of a
// block reference.
func setExtrusionAll(e Entity, n []float64) {
	setExtrusion(e, n)
	if ins, ok := e.(*Insert); ok {
		for _, a := range ins.Attributes {
			if a.Extrusion == nil {
				a.Extrusion = []float64{0, 0, 1}
			}
			copy(a.Extrusion, n)
		}
	}
}

func setExtrusion(e Entity, n []float64) {
	if v := extrusionField(e); v != nil {
		copy(v, n)
		return
	}
	switch x := e.(type) {
	case *Ellipse:
		x.Extrusion = append([]float64{}, n...)
	}
}

// transformWCSWithNormal transforms entities with WCS coordinates and a
// normal (uniform 3D scaling or -Z normal).
func transformWCSWithNormal(e Entity, t Transform, n []float64) (Entity, error) {
	m := t.linear3()
	n2 := transformNormal(m, n)
	switch x := e.(type) {
	case *Ellipse:
		transformPoint(t, x.Center)
		x.MajorAxis = mulMV(m, x.MajorAxis)
		x.Extrusion = n2
		return x, nil
	case *Spline:
		for _, p := range x.Controls {
			transformPoint(t, p)
		}
		for _, p := range x.Fits {
			transformPoint(t, p)
		}
		if len(x.StartTangent) > 0 {
			transformVector(t, x.StartTangent)
		}
		if len(x.EndTangent) > 0 {
			transformVector(t, x.EndTangent)
		}
		copy(x.Normal, n2)
		return x, nil
	case *Line:
		transformPoint(t, x.Start)
		transformPoint(t, x.End)
		x.Thickness *= t.Length(1)
		copy(x.Extrusion, n2)
		return x, nil
	case *Point:
		transformPoint(t, x.Coord)
		copy(x.Extrusion, n2)
		return x, nil
	case *MText:
		// rotation: direction of the OCS X axis rotated by Rotation
		ax, ay, _ := ocsAxes(n)
		r := x.Rotation * math.Pi / 180
		d := []float64{ax[0]*math.Cos(r) + ay[0]*math.Sin(r), ax[1]*math.Cos(r) + ay[1]*math.Sin(r), ax[2]*math.Cos(r) + ay[2]*math.Sin(r)}
		d2 := mulMV(m, d)
		bx, by, _ := ocsAxes(n2)
		transformPoint(t, x.Coord)
		k := t.Length(1)
		x.Height *= k
		x.Width *= k
		x.Rotation = normDeg(math.Atan2(dot3(by, d2), dot3(bx, d2)) * 180 / math.Pi)
		copy(x.Extrusion, n2)
		return x, nil
	}
	return e, ErrTransform
}

// Upright converts an entity with extrusion -Z (often created by mirroring
// in CAD applications) into the same geometry with extrusion +Z, as ezdxf's
// upright does. Other entities are not changed.
func Upright(e Entity) (Entity, error) {
	n := extrusionField(e)
	if n == nil || math.Abs(coord(n, 0)) > 1e-12 || math.Abs(coord(n, 1)) > 1e-12 || coord(n, 2) >= 0 {
		return e, nil
	}
	if !isOCSEntity(e) {
		if s, ok := e.(*Spline); ok {
			s.Normal = []float64{0, 0, 1}
			return s, nil
		}
		if el, ok := e.(*Ellipse); ok {
			// the parameter direction is reversed with the normal
			el.Extrusion = []float64{0, 0, 1}
			s, en := el.Start, el.End
			el.Start, el.End = normRad(-en), normRad(-s)
			if en-s >= 2*math.Pi-1e-12 {
				el.Start, el.End = 0, 2*math.Pi
			}
			return el, nil
		}
		copy(n, []float64{0, 0, 1})
		return e, nil
	}
	// OCS of -Z: X axis is -X, so the WCS coordinates are (-x, y, -z)
	setExtrusionAll(e, []float64{0, 0, 1})
	ne, err := TransformEntity(e, Transform{ScaleX: -1, ScaleY: 1, ScaleZ: -1, MirrorText: true})
	if err != nil {
		setExtrusionAll(e, []float64{0, 0, -1})
		return e, err
	}
	setExtrusionAll(ne, []float64{0, 0, 1})
	return ne, nil
}
