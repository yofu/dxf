package entity

import (
	"errors"
	"math"
)

// ErrTransform is returned by ApplyTransform for entities that cannot be
// transformed: entities whose extrusion direction is not the Z axis
// (their coordinates are in an OCS), and Unknown entities.
var ErrTransform = errors.New("entity cannot be transformed")

// Transform is a similarity transformation: scaling by Scale and rotation
// by Angle (degrees, counterclockwise about the Z axis) around Base,
// followed by the translation Offset.
//
//	p' = Base + Rz(Angle) * Scale * (p - Base) + Offset
type Transform struct {
	Base   []float64 // center of rotation and scaling (nil: origin)
	Angle  float64   // degrees
	Scale  float64   // uniform scale factor (0 is treated as 1; must not be negative)
	Offset []float64 // translation (nil: none)
}

// Move returns a translation.
func Move(dx, dy, dz float64) Transform {
	return Transform{Offset: []float64{dx, dy, dz}}
}

// Rotate returns a rotation by angle (degrees) about the Z axis through base.
func Rotate(base []float64, angle float64) Transform {
	return Transform{Base: base, Angle: angle}
}

// ScaleBy returns a uniform scaling by factor around base.
func ScaleBy(base []float64, factor float64) Transform {
	return Transform{Base: base, Scale: factor}
}

// Transformable is implemented by entities that can be transformed.
type Transformable interface {
	ApplyTransform(Transform) error
}

func (t Transform) s() float64 {
	if t.Scale == 0 {
		return 1
	}
	return t.Scale
}

func coord(v []float64, i int) float64 {
	if i < len(v) {
		return v[i]
	}
	return 0
}

func (t Transform) valid() error {
	if t.Scale < 0 {
		return errors.New("negative scale is not supported")
	}
	return nil
}

// Point transforms a point (2D points keep 2 coordinates; their z is
// taken as the z of Base).
func (t Transform) Point(p []float64) []float64 {
	s := t.s()
	rad := t.Angle * math.Pi / 180
	sin, cos := math.Sin(rad), math.Cos(rad)
	x := (coord(p, 0) - coord(t.Base, 0)) * s
	y := (coord(p, 1) - coord(t.Base, 1)) * s
	q := []float64{
		coord(t.Base, 0) + x*cos - y*sin + coord(t.Offset, 0),
		coord(t.Base, 1) + x*sin + y*cos + coord(t.Offset, 1),
	}
	if len(p) >= 3 {
		q = append(q, coord(t.Base, 2)+(p[2]-coord(t.Base, 2))*s+coord(t.Offset, 2))
	}
	for i := 3; i < len(p); i++ { // keep extra values (e.g. widths)
		q = append(q, p[i])
	}
	return q
}

// Vector transforms a direction vector (rotation and scaling, no translation).
func (t Transform) Vector(v []float64) []float64 {
	s := t.s()
	rad := t.Angle * math.Pi / 180
	sin, cos := math.Sin(rad), math.Cos(rad)
	q := []float64{(coord(v, 0)*cos - coord(v, 1)*sin) * s, (coord(v, 0)*sin + coord(v, 1)*cos) * s}
	if len(v) >= 3 {
		q = append(q, v[2]*s)
	}
	return q
}

// Z transforms a z coordinate (elevation).
func (t Transform) Z(z float64) float64 {
	return coord(t.Base, 2) + (z-coord(t.Base, 2))*t.s() + coord(t.Offset, 2)
}

// Length transforms a length.
func (t Transform) Length(l float64) float64 {
	return l * t.s()
}

func transformPoint(t Transform, p []float64) {
	copy(p, t.Point(p))
}

func transformVector(t Transform, v []float64) {
	copy(v, t.Vector(v))
}

// zAxis reports whether the extrusion direction is the Z axis (or not given).
func zAxis(e []float64) bool {
	return len(e) < 3 || (math.Abs(e[0]) < 1e-12 && math.Abs(e[1]) < 1e-12 && e[2] > 0)
}

func checkZ(t Transform, e []float64) error {
	if err := t.valid(); err != nil {
		return err
	}
	if !zAxis(e) {
		return ErrTransform
	}
	return nil
}

// ApplyTransform transforms the line.
func (l *Line) ApplyTransform(t Transform) error {
	if err := t.valid(); err != nil {
		return err
	}
	transformPoint(t, l.Start)
	transformPoint(t, l.End)
	l.Thickness = t.Length(l.Thickness)
	return nil
}

// ApplyTransform transforms the circle.
func (c *Circle) ApplyTransform(t Transform) error {
	if err := checkZ(t, c.Direction); err != nil {
		return err
	}
	transformPoint(t, c.Center)
	c.Radius = t.Length(c.Radius)
	c.Thickness = t.Length(c.Thickness)
	return nil
}

// ApplyTransform transforms the arc.
func (a *Arc) ApplyTransform(t Transform) error {
	if err := a.Circle.ApplyTransform(t); err != nil {
		return err
	}
	a.Angle[0] += t.Angle
	a.Angle[1] += t.Angle
	return nil
}

// ApplyTransform transforms the ellipse.
func (e *Ellipse) ApplyTransform(t Transform) error {
	if err := checkZ(t, e.Extrusion); err != nil {
		return err
	}
	transformPoint(t, e.Center)
	transformVector(t, e.MajorAxis)
	return nil
}

// ApplyTransform transforms the point.
func (p *Point) ApplyTransform(t Transform) error {
	if err := checkZ(t, p.Extrusion); err != nil {
		return err
	}
	transformPoint(t, p.Coord)
	p.Thickness = t.Length(p.Thickness)
	p.Angle += t.Angle
	return nil
}

// ApplyTransform transforms the text (also ATTDEF and ATTRIB).
func (tx *Text) ApplyTransform(t Transform) error {
	if err := checkZ(t, tx.Extrusion); err != nil {
		return err
	}
	transformPoint(t, tx.Coord1)
	if len(tx.Coord2) > 0 {
		transformPoint(t, tx.Coord2)
	}
	tx.Height = t.Length(tx.Height)
	tx.Thickness = t.Length(tx.Thickness)
	tx.Rotation += t.Angle
	return nil
}

// ApplyTransform transforms the multiline text.
func (m *MText) ApplyTransform(t Transform) error {
	if err := checkZ(t, m.Extrusion); err != nil {
		return err
	}
	transformPoint(t, m.Coord)
	m.Height = t.Length(m.Height)
	m.Width = t.Length(m.Width)
	m.Rotation += t.Angle
	return nil
}

// ApplyTransform transforms the 3D face.
func (f *ThreeDFace) ApplyTransform(t Transform) error {
	if err := t.valid(); err != nil {
		return err
	}
	for _, p := range f.Points {
		transformPoint(t, p)
	}
	return nil
}

// ApplyTransform transforms the solid.
func (s *Solid) ApplyTransform(t Transform) error {
	if err := checkZ(t, s.Extrusion); err != nil {
		return err
	}
	for _, p := range s.Points {
		transformPoint(t, p)
	}
	s.Thickness = t.Length(s.Thickness)
	return nil
}

// ApplyTransform transforms the lightweight polyline.
func (l *LwPolyline) ApplyTransform(t Transform) error {
	if err := checkZ(t, l.Extrusion); err != nil {
		return err
	}
	for _, v := range l.Vertices {
		copy(v, t.Point(v[:2]))
	}
	for _, w := range l.Widths {
		for i := range w {
			w[i] = t.Length(w[i])
		}
	}
	l.ConstantWidth = t.Length(l.ConstantWidth)
	l.Elevation = t.Z(l.Elevation)
	l.Thickness = t.Length(l.Thickness)
	return nil
}

// ApplyTransform transforms the polyline and its vertices.
func (p *Polyline) ApplyTransform(t Transform) error {
	if err := checkZ(t, p.Extrusion); err != nil {
		return err
	}
	for _, v := range p.Vertices {
		v.ApplyTransform(t)
	}
	if p.Flag&(8|16|64) == 0 { // 2D polyline: z of vertices is the elevation
		p.Elevation = t.Z(p.Elevation)
	}
	p.StartWidth = t.Length(p.StartWidth)
	p.EndWidth = t.Length(p.EndWidth)
	p.Thickness = t.Length(p.Thickness)
	return nil
}

// ApplyTransform transforms the vertex.
func (v *Vertex) ApplyTransform(t Transform) error {
	if err := t.valid(); err != nil {
		return err
	}
	if v.Flag&128 != 0 && v.Flag&64 == 0 { // polyface face record: no location
		return nil
	}
	transformPoint(t, v.Coord)
	v.StartWidth = t.Length(v.StartWidth)
	v.EndWidth = t.Length(v.EndWidth)
	v.Tangent += t.Angle
	return nil
}

// ApplyTransform transforms the spline.
func (s *Spline) ApplyTransform(t Transform) error {
	if err := checkZ(t, s.Normal); err != nil {
		return err
	}
	for _, p := range s.Controls {
		transformPoint(t, p)
	}
	for _, p := range s.Fits {
		transformPoint(t, p)
	}
	if len(s.StartTangent) > 0 {
		transformVector(t, s.StartTangent)
	}
	if len(s.EndTangent) > 0 {
		transformVector(t, s.EndTangent)
	}
	return nil
}

// ApplyTransform transforms the block reference and its attributes
// (the block definition is not changed).
func (i *Insert) ApplyTransform(t Transform) error {
	if err := checkZ(t, i.Direction); err != nil {
		return err
	}
	transformPoint(t, i.Coord)
	for k := range i.Scale {
		i.Scale[k] = t.Length(i.Scale[k])
	}
	i.Rotation += t.Angle
	i.ColumnSpacing = t.Length(i.ColumnSpacing)
	i.RowSpacing = t.Length(i.RowSpacing)
	for _, a := range i.Attributes {
		if err := a.ApplyTransform(t); err != nil {
			return err
		}
	}
	return nil
}

// ApplyTransform transforms the hatch: boundaries, pattern and seed points.
func (h *Hatch) ApplyTransform(t Transform) error {
	if err := checkZ(t, h.Extrusion); err != nil {
		return err
	}
	p2 := func(p []float64) {
		if len(p) >= 2 {
			copy(p, t.Point(p[:2]))
		}
	}
	for _, b := range h.Boundaries {
		for _, v := range b.Vertices {
			p2(v)
		}
		for _, e := range b.Edges {
			switch c := e.(type) {
			case *HatchLine:
				p2(c.Start)
				p2(c.End)
			case *HatchArc:
				p2(c.Center)
				c.Radius = t.Length(c.Radius)
				c.Start += t.Angle
				c.End += t.Angle
			case *HatchEllipse:
				p2(c.Center)
				copy(c.MajorAxis, t.Vector(c.MajorAxis[:2]))
			case *HatchSpline:
				for _, p := range c.Controls {
					p2(p)
				}
				for _, p := range c.Fits {
					p2(p)
				}
			}
		}
	}
	for _, l := range h.PatternLines {
		l.Angle += t.Angle
		p2(l.Base)
		copy(l.Offset, t.Vector(l.Offset[:2]))
		for k := range l.Dashes {
			l.Dashes[k] = t.Length(l.Dashes[k])
		}
	}
	if h.Gradient != nil {
		h.Gradient.Rotation += t.Angle
	}
	if !h.Solid {
		h.PatternAngle += t.Angle
		h.PatternScale = t.Length(h.PatternScale)
	}
	for _, s := range h.Seeds {
		p2(s)
	}
	h.Elevation = t.Z(h.Elevation)
	return nil
}

// ApplyTransform transforms the definition points of the dimension.
// The graphics block is not changed: transform its entities as well,
// or regenerate it (Drawing methods do this).
func (d *Dimension) ApplyTransform(t Transform) error {
	if err := checkZ(t, d.Extrusion); err != nil {
		return err
	}
	for _, p := range [][]float64{d.DefPoint, d.TextMid, d.DefPoint2, d.DefPoint3, d.DefPoint4, d.DefPoint5, d.ClonePoint} {
		if len(p) > 0 {
			transformPoint(t, p)
		}
	}
	if d.Type == DIM_LINEAR {
		d.Angle += t.Angle
	}
	if d.Type != DIM_ANGULAR && d.Type != DIM_ANGULAR3P {
		d.Measurement = t.Length(d.Measurement)
	}
	d.LeaderLength = t.Length(d.LeaderLength)
	return nil
}

// ApplyTransform transforms the leader.
func (l *Leader) ApplyTransform(t Transform) error {
	if err := checkZ(t, l.Extrusion); err != nil {
		return err
	}
	for _, v := range l.Vertices {
		transformPoint(t, v)
	}
	transformVector(t, l.HorizontalDirection)
	l.TextHeight = t.Length(l.TextHeight)
	l.TextWidth = t.Length(l.TextWidth)
	for _, v := range [][]float64{l.BlockOffset, l.AnnotationOffset} {
		if len(v) > 0 {
			transformVector(t, v)
		}
	}
	return nil
}

// ApplyTransform transforms the viewport in paper space (the view of
// model space is not changed).
func (v *Viewport) ApplyTransform(t Transform) error {
	if err := t.valid(); err != nil {
		return err
	}
	transformPoint(t, v.Center)
	v.Width = t.Length(v.Width)
	v.Height = t.Length(v.Height)
	return nil
}

// ApplyTransform returns ErrTransform: the data of unknown entities is
// not interpreted.
func (u *Unknown) ApplyTransform(t Transform) error {
	return ErrTransform
}

// ApplyTransform transforms the underlay: insertion point, scale and
// rotation (the clipping boundary is in underlay coordinates).
func (u *Underlay) ApplyTransform(t Transform) error {
	if err := checkZ(t, u.Extrusion); err != nil {
		return err
	}
	transformPoint(t, u.Insert)
	for k := range u.Scale {
		u.Scale[k] = t.Length(u.Scale[k])
	}
	u.Rotation += t.Angle
	return nil
}
