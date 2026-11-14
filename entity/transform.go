package entity

import (
	"errors"
	"fmt"
	"math"
)

// ErrTransform is returned by ApplyTransform for entities that cannot be
// transformed: entities whose extrusion direction is not the Z axis
// (their coordinates are in an OCS), and Unknown entities.
var ErrTransform = errors.New("entity cannot be transformed")

// ErrNonUniform is returned by ApplyTransform for entities that cannot be
// scaled non-uniformly (e.g. arcs of polylines, rotated text). It wraps
// ErrTransform. Circles and arcs are converted to ellipses by
// TransformEntity.
var ErrNonUniform = fmt.Errorf("%w: non-uniform scaling", ErrTransform)

// Transform is an affine transformation in the XY plane (scaling, mirror
// and rotation) with scaling of Z, around Base, followed by the
// translation Offset:
//
//	p' = Base + Rz(Angle) * diag(ScaleX, ScaleY, ScaleZ) * (p - Base) + Offset
//
// ScaleX, ScaleY and ScaleZ default to Scale, and Scale to 1 (0 means the
// default). A negative factor mirrors: ScaleY = -1 mirrors about the X
// axis through Base (see MirrorBy). Entities keep the Z axis as their
// extrusion direction: a mirrored entity is mirrored in the XY plane.
type Transform struct {
	Base   []float64 // center of rotation and scaling (nil: origin)
	Angle  float64   // degrees
	Scale  float64   // uniform scale factor (0 is treated as 1)
	ScaleX float64   // scale factor along X before rotation (0: Scale)
	ScaleY float64   // scale factor along Y before rotation (0: Scale)
	ScaleZ float64   // scale factor along Z (0: Scale)
	Offset []float64 // translation (nil: none)
	// MirrorText mirrors text (TEXT, ATTDEF, ATTRIB) exactly, as upside
	// down text. By default mirrored text stays readable: its position and
	// alignment are mirrored, like AutoCAD with MIRRTEXT = 0. MTEXT is
	// always readable.
	MirrorText bool
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

// ScaleXYZ returns a scaling by sx, sy and sz (non-zero; negative values
// mirror) around base.
func ScaleXYZ(base []float64, sx, sy, sz float64) Transform {
	return Transform{Base: base, ScaleX: sx, ScaleY: sy, ScaleZ: sz}
}

// MirrorBy returns the mirror about the line through p1 and p2 (in the
// XY plane).
func MirrorBy(p1, p2 []float64) Transform {
	a := math.Atan2(coord(p2, 1)-coord(p1, 1), coord(p2, 0)-coord(p1, 0)) * 180 / math.Pi
	return Transform{Base: []float64{coord(p1, 0), coord(p1, 1), 0}, Angle: 2 * a, ScaleY: -1}
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

// Factors returns the scale factors along X, Y and Z.
func (t Transform) Factors() (sx, sy, sz float64) {
	sx, sy, sz = t.ScaleX, t.ScaleY, t.ScaleZ
	if sx == 0 {
		sx = t.s()
	}
	if sy == 0 {
		sy = t.s()
	}
	if sz == 0 {
		sz = t.s()
	}
	return
}

// linear returns the 2x2 matrix of the transformation in the XY plane.
func (t Transform) linear() [2][2]float64 {
	sx, sy, _ := t.Factors()
	rad := t.Angle * math.Pi / 180
	sin, cos := math.Sin(rad), math.Cos(rad)
	return [2][2]float64{{cos * sx, -sin * sy}, {sin * sx, cos * sy}}
}

func (t Transform) apply2(x, y float64) (float64, float64) {
	m := t.linear()
	return m[0][0]*x + m[0][1]*y, m[1][0]*x + m[1][1]*y
}

// Conformal reports whether the transformation keeps shapes (uniform
// scaling in the XY plane, possibly mirrored).
func (t Transform) Conformal() bool {
	sx, sy, _ := t.Factors()
	return math.Abs(math.Abs(sx)-math.Abs(sy)) <= 1e-12*math.Max(math.Abs(sx), math.Abs(sy))
}

// Mirrored reports whether the transformation mirrors in the XY plane.
func (t Transform) Mirrored() bool {
	sx, sy, _ := t.Factors()
	return sx*sy < 0
}

func coord(v []float64, i int) float64 {
	if i < len(v) {
		return v[i]
	}
	return 0
}

func (t Transform) valid() error {
	return nil
}

// Point transforms a point (2D points keep 2 coordinates; their z is
// taken as the z of Base).
func (t Transform) Point(p []float64) []float64 {
	_, _, sz := t.Factors()
	x, y := t.apply2(coord(p, 0)-coord(t.Base, 0), coord(p, 1)-coord(t.Base, 1))
	q := []float64{
		coord(t.Base, 0) + x + coord(t.Offset, 0),
		coord(t.Base, 1) + y + coord(t.Offset, 1),
	}
	if len(p) >= 3 {
		q = append(q, coord(t.Base, 2)+(p[2]-coord(t.Base, 2))*sz+coord(t.Offset, 2))
	}
	for i := 3; i < len(p); i++ { // keep extra values (e.g. widths)
		q = append(q, p[i])
	}
	return q
}

// Vector transforms a direction vector (no translation).
func (t Transform) Vector(v []float64) []float64 {
	_, _, sz := t.Factors()
	x, y := t.apply2(coord(v, 0), coord(v, 1))
	q := []float64{x, y}
	if len(v) >= 3 {
		q = append(q, v[2]*sz)
	}
	return q
}

// Z transforms a z coordinate (elevation).
func (t Transform) Z(z float64) float64 {
	_, _, sz := t.Factors()
	return coord(t.Base, 2) + (z-coord(t.Base, 2))*sz + coord(t.Offset, 2)
}

// Length transforms a length in the XY plane (for conformal
// transformations).
func (t Transform) Length(l float64) float64 {
	sx, _, _ := t.Factors()
	return l * math.Abs(sx)
}

// ZLength transforms a length along Z (thickness).
func (t Transform) ZLength(l float64) float64 {
	_, _, sz := t.Factors()
	return l * sz
}

// Direction transforms a direction angle (degrees) in the XY plane.
func (t Transform) Direction(deg float64) float64 {
	rad := deg * math.Pi / 180
	x, y := t.apply2(math.Cos(rad), math.Sin(rad))
	return math.Atan2(y, x) * 180 / math.Pi
}

// frame transforms the frame of an entity rotated by rot (degrees), e.g.
// a block reference: the images of its X and Y axes must be orthogonal.
// It returns the new rotation and the scale factors along the new X and Y
// axes (ky < 0: mirrored).
func (t Transform) frame(rot float64) (newRot, kx, ky float64, ok bool) {
	rad := rot * math.Pi / 180
	x1, y1 := t.apply2(math.Cos(rad), math.Sin(rad))
	x2, y2 := t.apply2(-math.Sin(rad), math.Cos(rad))
	kx, ky = math.Hypot(x1, y1), math.Hypot(x2, y2)
	if math.Abs(x1*x2+y1*y2) > 1e-9*kx*ky {
		return 0, 0, 0, false
	}
	if x1*y2-y1*x2 < 0 {
		ky = -ky
	}
	return math.Atan2(y1, x1) * 180 / math.Pi, kx, ky, true
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

// TransformEntity transforms e by t. It returns e, or a new entity when e
// cannot keep its type: circles and arcs become ellipses under
// non-uniform scaling (the new entity keeps the properties and the handle
// of e).
func TransformEntity(e Entity, t Transform) (Entity, error) {
	if !t.Conformal() {
		switch c := e.(type) {
		case *Arc:
			if err := checkZ(t, c.Direction); err != nil {
				return e, err
			}
			el := circleToEllipse(c.Circle, c.Angle[0], c.Angle[1])
			return el, el.ApplyTransform(t)
		case *Circle:
			if err := checkZ(t, c.Direction); err != nil {
				return e, err
			}
			el := circleToEllipse(c, 0, 360)
			return el, el.ApplyTransform(t)
		}
	}
	tr, ok := e.(Transformable)
	if !ok {
		return e, ErrTransform
	}
	return e, tr.ApplyTransform(t)
}

// circleToEllipse returns an ellipse with the entity data of c (start and
// end angles in degrees).
func circleToEllipse(c *Circle, start, end float64) *Ellipse {
	c.entity.Type = ELLIPSE
	el := &Ellipse{
		entity:    c.entity,
		Center:    append([]float64{}, c.Center...),
		MajorAxis: []float64{c.Radius, 0, 0},
		Extrusion: []float64{0, 0, 1},
		Ratio:     1,
		Start:     start * math.Pi / 180,
		End:       end * math.Pi / 180,
	}
	if end-start >= 360 {
		el.Start, el.End = 0, 2*math.Pi
	}
	return el
}

// ApplyTransform transforms the line.
func (l *Line) ApplyTransform(t Transform) error {
	transformPoint(t, l.Start)
	transformPoint(t, l.End)
	l.Thickness = t.ZLength(l.Thickness)
	return nil
}

// ApplyTransform transforms the circle (see TransformEntity for
// non-uniform scaling).
func (c *Circle) ApplyTransform(t Transform) error {
	if err := checkZ(t, c.Direction); err != nil {
		return err
	}
	if !t.Conformal() {
		return ErrNonUniform
	}
	transformPoint(t, c.Center)
	c.Radius = t.Length(c.Radius)
	c.Thickness = t.ZLength(c.Thickness)
	return nil
}

// ApplyTransform transforms the arc (see TransformEntity for non-uniform
// scaling). A mirrored arc goes from the image of its end angle to the
// image of its start angle.
func (a *Arc) ApplyTransform(t Transform) error {
	if err := a.Circle.ApplyTransform(t); err != nil {
		return err
	}
	s, e := t.Direction(a.Angle[0]), t.Direction(a.Angle[1])
	if t.Mirrored() {
		s, e = e, s
	}
	a.Angle[0], a.Angle[1] = normDeg(s), normDeg(e)
	return nil
}

func normDeg(a float64) float64 {
	a = math.Mod(a, 360)
	if a < 0 {
		a += 360
	}
	if a > 360-1e-10 {
		a = 0
	}
	return a
}

func normRad(a float64) float64 {
	a = math.Mod(a, 2*math.Pi)
	if a < 0 {
		a += 2 * math.Pi
	}
	return a
}

// ApplyTransform transforms the ellipse (any transformation: the axes and
// parameters are recomputed).
func (e *Ellipse) ApplyTransform(t Transform) error {
	if err := checkZ(t, e.Extrusion); err != nil {
		return err
	}
	transformPoint(t, e.Center)
	ux, uy := coord(e.MajorAxis, 0), coord(e.MajorAxis, 1)
	vx, vy := -uy*e.Ratio, ux*e.Ratio
	ax, ay := t.apply2(ux, uy)
	bx, by := t.apply2(vx, vy)
	// principal axes of the conjugate semi-diameters a and b
	t0 := 0.5 * math.Atan2(2*(ax*bx+ay*by), ax*ax+ay*ay-bx*bx-by*by)
	axes := func(t0 float64) (float64, float64, float64, float64) {
		c, s := math.Cos(t0), math.Sin(t0)
		return ax*c + bx*s, ay*c + by*s, -ax*s + bx*c, -ay*s + by*c
	}
	mx, my, nx, ny := axes(t0)
	if math.Hypot(nx, ny) > math.Hypot(mx, my) {
		t0 += math.Pi / 2
		mx, my, nx, ny = axes(t0)
	}
	full := e.End-e.Start >= 2*math.Pi-1e-12 || (e.Start == 0 && e.End == 0)
	sweep := e.End - e.Start
	if sweep < 0 {
		sweep += 2 * math.Pi
	}
	// P'(p) = C' + m cos(p - t0) + n sin(p - t0); n is +90 degrees from m
	// unless the transformation mirrors (then the parameter is reversed)
	if mx*ny-my*nx >= 0 {
		e.Start = e.Start - t0
	} else {
		e.Start = t0 - e.Start - sweep
	}
	if full {
		e.Start, e.End = 0, 2*math.Pi
	} else {
		e.Start = normRad(e.Start)
		e.End = e.Start + sweep
		if e.End > 2*math.Pi {
			e.End -= 2 * math.Pi
		}
	}
	e.Ratio = math.Hypot(nx, ny) / math.Hypot(mx, my)
	_, _, sz := t.Factors()
	e.MajorAxis = []float64{mx, my, coord(e.MajorAxis, 2) * sz}
	return nil
}

// ApplyTransform transforms the point.
func (p *Point) ApplyTransform(t Transform) error {
	if err := checkZ(t, p.Extrusion); err != nil {
		return err
	}
	transformPoint(t, p.Coord)
	p.Thickness = t.ZLength(p.Thickness)
	p.Angle = t.Direction(p.Angle)
	return nil
}

// Ways text is mirrored (textFrame).
const (
	textNotMirrored = iota
	textUpsideDown  // mirrored text (Transform.MirrorText)
	textFlipAcross  // readable: alignment flipped across the baseline (bottom <-> top)
	textFlipAlong   // readable: alignment flipped along the baseline (left <-> right)
)

// textFrame transforms the frame of text rotated by rot: it returns the
// new rotation, the factors of height and width, and how the text is
// mirrored.
func textFrame(t Transform, rot float64) (newRot, kh, kw float64, mirror int, err error) {
	r, kx, ky, ok := t.frame(rot)
	if !ok {
		return 0, 0, 0, 0, ErrNonUniform
	}
	kh, kw = math.Abs(ky), kx/math.Abs(ky)
	if ky > 0 {
		return normDeg(r), kh, kw, textNotMirrored, nil
	}
	if t.MirrorText {
		return normDeg(r), kh, kw, textUpsideDown, nil
	}
	// readable: flipped across the baseline (rotation r) or along it
	// (rotation r + 180), whichever is closer to the original rotation
	diff := func(a float64) float64 {
		d := math.Abs(math.Mod(a-rot, 360))
		return math.Min(d, 360-d)
	}
	if diff(r) <= diff(r+180) {
		return normDeg(r), kh, kw, textFlipAcross, nil
	}
	return normDeg(r + 180), kh, kw, textFlipAlong, nil
}

// ApplyTransform transforms the text (also ATTDEF and ATTRIB). Mirrored
// text stays readable unless Transform.MirrorText is set. Under
// non-uniform scaling the axes of the text must stay orthogonal (e.g.
// scaling along the text); the width factor is adjusted.
func (tx *Text) ApplyTransform(t Transform) error {
	if err := checkZ(t, tx.Extrusion); err != nil {
		return err
	}
	rot, kh, kw, mirror, err := textFrame(t, tx.Rotation)
	if err != nil {
		return err
	}
	p1 := t.Point(tx.Coord1)
	var p2 []float64
	if len(tx.Coord2) > 0 {
		p2 = t.Point(tx.Coord2)
	}
	h, v := tx.HorizontalFlag, tx.VerticalFlag
	switch mirror {
	case textUpsideDown:
		tx.GenFlag ^= 4
	case textFlipAlong: // left <-> right
		switch h {
		case 0:
			if v == 0 {
				p2 = append([]float64{}, p1...)
			}
			tx.HorizontalFlag = 2
		case 2:
			if v == 0 && p2 != nil {
				p1 = append([]float64{}, p2...)
			}
			tx.HorizontalFlag = 0
		case 3, 5: // aligned, fit: the points are swapped
			if p2 != nil {
				p1, p2 = p2, p1
			}
		}
	case textFlipAcross: // bottom <-> top (the baseline becomes the top)
		switch {
		case h == 3 || h == 5:
			// aligned and fit text stays on the baseline
		case v == 0 || v == 1:
			if v == 0 && h == 0 {
				p2 = append([]float64{}, p1...)
			}
			tx.VerticalFlag = 3
		case v == 3:
			tx.VerticalFlag = 1
		}
	}
	tx.Coord1 = p1
	if p2 != nil {
		tx.Coord2 = p2
	}
	tx.Rotation = rot
	tx.Height *= kh
	if kw != 1 {
		w := tx.WidthFactor
		if w == 0 {
			w = 1
		}
		tx.WidthFactor = w * kw
	}
	tx.Thickness = t.ZLength(tx.Thickness)
	return nil
}

// ApplyTransform transforms the multiline text. Mirrored text stays
// readable: its attachment point is mirrored.
func (m *MText) ApplyTransform(t Transform) error {
	if err := checkZ(t, m.Extrusion); err != nil {
		return err
	}
	mt := t
	mt.MirrorText = false
	rot, kh, kw, mirror, err := textFrame(mt, m.Rotation)
	if err != nil {
		return err
	}
	transformPoint(t, m.Coord)
	if a := m.Attachment; a >= 1 && a <= 9 {
		row, col := (a-1)/3, (a-1)%3
		switch mirror {
		case textFlipAlong:
			col = 2 - col
		case textFlipAcross:
			row = 2 - row
		}
		m.Attachment = row*3 + col + 1
	}
	m.Rotation = rot
	m.Height *= kh
	m.Width *= kh * kw
	if c := m.Columns; c != nil {
		c.Width *= kh * kw
		c.Gutter *= kh * kw
		c.TotalWidth *= kh * kw
		c.DefinedHeight *= kh
		c.TotalHeight *= kh
		for i := range c.Heights {
			c.Heights[i] *= kh
		}
	}
	return nil
}

// ApplyTransform transforms the 3D face.
func (f *ThreeDFace) ApplyTransform(t Transform) error {
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
	s.Thickness = t.ZLength(s.Thickness)
	return nil
}

// nonZero reports whether any value is not 0.
func nonZero(vs ...float64) bool {
	for _, v := range vs {
		if v != 0 {
			return true
		}
	}
	return false
}

// ApplyTransform transforms the lightweight polyline. Arc segments
// (bulges) and widths need a conformal transformation; mirrored bulges
// change their sign.
func (l *LwPolyline) ApplyTransform(t Transform) error {
	if err := checkZ(t, l.Extrusion); err != nil {
		return err
	}
	if !t.Conformal() {
		if nonZero(l.Bulges...) || nonZero(l.ConstantWidth) {
			return ErrNonUniform
		}
		for _, w := range l.Widths {
			if nonZero(w...) {
				return ErrNonUniform
			}
		}
	}
	for _, v := range l.Vertices {
		copy(v, t.Point(v[:2]))
	}
	if t.Mirrored() {
		for i := range l.Bulges {
			l.Bulges[i] = -l.Bulges[i]
		}
	}
	for _, w := range l.Widths {
		for i := range w {
			w[i] = t.Length(w[i])
		}
	}
	l.ConstantWidth = t.Length(l.ConstantWidth)
	l.Elevation = t.Z(l.Elevation)
	l.Thickness = t.ZLength(l.Thickness)
	return nil
}

// ApplyTransform transforms the polyline and its vertices.
func (p *Polyline) ApplyTransform(t Transform) error {
	if err := checkZ(t, p.Extrusion); err != nil {
		return err
	}
	if !t.Conformal() {
		if nonZero(p.StartWidth, p.EndWidth) {
			return ErrNonUniform
		}
		for _, v := range p.Vertices {
			if nonZero(v.Bulge, v.StartWidth, v.EndWidth) {
				return ErrNonUniform
			}
		}
	}
	for _, v := range p.Vertices {
		if err := v.ApplyTransform(t); err != nil {
			return err
		}
	}
	if p.Flag&(8|16|64) == 0 { // 2D polyline: z of vertices is the elevation
		p.Elevation = t.Z(p.Elevation)
	}
	p.StartWidth = t.Length(p.StartWidth)
	p.EndWidth = t.Length(p.EndWidth)
	p.Thickness = t.ZLength(p.Thickness)
	return nil
}

// ApplyTransform transforms the vertex.
func (v *Vertex) ApplyTransform(t Transform) error {
	if v.Flag&128 != 0 && v.Flag&64 == 0 { // polyface face record: no location
		return nil
	}
	if !t.Conformal() && nonZero(v.Bulge, v.StartWidth, v.EndWidth) {
		return ErrNonUniform
	}
	transformPoint(t, v.Coord)
	if t.Mirrored() {
		v.Bulge = -v.Bulge
	}
	v.StartWidth = t.Length(v.StartWidth)
	v.EndWidth = t.Length(v.EndWidth)
	if v.Tangent != 0 || v.Flag&2 != 0 { // curve fit tangent
		v.Tangent = normDeg(t.Direction(v.Tangent))
	}
	return nil
}

// ApplyTransform transforms the spline (any transformation: splines are
// affine invariant).
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

// ApplyTransform transforms the block reference and its attributes (the
// block definition is not changed). Mirroring makes the Y scale negative.
// Under non-uniform scaling the axes of the reference must stay
// orthogonal (e.g. scaling along its axes).
func (i *Insert) ApplyTransform(t Transform) error {
	if err := checkZ(t, i.Direction); err != nil {
		return err
	}
	rot, kx, ky, ok := t.frame(i.Rotation)
	if !ok {
		return ErrNonUniform
	}
	for _, a := range i.Attributes {
		if err := a.ApplyTransform(t); err != nil {
			return err
		}
	}
	_, _, sz := t.Factors()
	transformPoint(t, i.Coord)
	for len(i.Scale) < 3 {
		i.Scale = append(i.Scale, 1)
	}
	i.Scale[0] *= kx
	i.Scale[1] *= ky
	i.Scale[2] *= sz
	i.Rotation = normDeg(rot)
	i.ColumnSpacing *= kx
	i.RowSpacing *= ky
	return nil
}

// transformEdgeAngle transforms a start or end angle (degrees) of a hatch
// arc edge. Clockwise edges store 360 - angle (DXF reference); a mirrored
// edge changes its direction.
func transformEdgeAngle(t Transform, a float64, ccw bool) (float64, bool) {
	real := a
	if !ccw {
		real = -a
	}
	real = t.Direction(real)
	if t.Mirrored() {
		ccw = !ccw
	}
	if !ccw {
		real = -real
	}
	return normDeg(real), ccw
}

// ApplyTransform transforms the hatch: boundaries, pattern and seed points.
// Arcs (bulges, arc and ellipse edges) need a conformal transformation.
func (h *Hatch) ApplyTransform(t Transform) error {
	if err := checkZ(t, h.Extrusion); err != nil {
		return err
	}
	if !t.Conformal() {
		for _, b := range h.Boundaries {
			if nonZero(b.Bulges...) {
				return ErrNonUniform
			}
			for _, e := range b.Edges {
				switch e.(type) {
				case *HatchArc, *HatchEllipse:
					return ErrNonUniform
				}
			}
		}
	}
	p2 := func(p []float64) {
		if len(p) >= 2 {
			copy(p, t.Point(p[:2]))
		}
	}
	mirrored := t.Mirrored()
	for _, b := range h.Boundaries {
		for _, v := range b.Vertices {
			p2(v)
		}
		if mirrored {
			for i := range b.Bulges {
				b.Bulges[i] = -b.Bulges[i]
			}
		}
		for _, e := range b.Edges {
			switch c := e.(type) {
			case *HatchLine:
				p2(c.Start)
				p2(c.End)
			case *HatchArc:
				p2(c.Center)
				c.Radius = t.Length(c.Radius)
				ccw := c.CCW
				c.Start, _ = transformEdgeAngle(t, c.Start, ccw)
				c.End, c.CCW = transformEdgeAngle(t, c.End, ccw)
			case *HatchEllipse:
				// the angles are relative to the major axis: a mirror
				// changes only the direction
				p2(c.Center)
				copy(c.MajorAxis, t.Vector(c.MajorAxis[:2]))
				if mirrored {
					c.CCW = !c.CCW
				}
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
	// pattern lines are affine invariant: direction, base point, offset
	// and dashes are transformed
	for _, l := range h.PatternLines {
		rad := l.Angle * math.Pi / 180
		dx, dy := t.apply2(math.Cos(rad), math.Sin(rad))
		k := math.Hypot(dx, dy)
		l.Angle = normDeg(math.Atan2(dy, dx) * 180 / math.Pi)
		p2(l.Base)
		copy(l.Offset, t.Vector(l.Offset[:2]))
		for i := range l.Dashes {
			l.Dashes[i] *= k
		}
	}
	if h.Gradient != nil {
		h.Gradient.Rotation = normDeg(t.Direction(h.Gradient.Rotation))
	}
	if !h.Solid {
		h.PatternAngle = normDeg(t.Direction(h.PatternAngle))
		if t.Conformal() {
			h.PatternScale = t.Length(h.PatternScale)
		}
	}
	for _, s := range h.Seeds {
		p2(s)
	}
	h.Elevation = t.Z(h.Elevation)
	return nil
}

// ApplyTransform transforms the definition points of the dimension
// (conformal transformations only). The graphics block is not changed:
// transform its entities as well, or regenerate it (Drawing methods do
// this).
func (d *Dimension) ApplyTransform(t Transform) error {
	if err := checkZ(t, d.Extrusion); err != nil {
		return err
	}
	if !t.Conformal() {
		return ErrNonUniform
	}
	for _, p := range [][]float64{d.DefPoint, d.TextMid, d.DefPoint2, d.DefPoint3, d.DefPoint4, d.DefPoint5, d.ClonePoint, d.ArcLeader1, d.ArcLeader2} {
		if len(p) > 0 {
			transformPoint(t, p)
		}
	}
	if d.Type == DIM_LINEAR {
		d.Angle = normDeg(t.Direction(d.Angle))
	}
	if d.Type == DIM_ARC && t.Mirrored() {
		// the arc goes the other way: its ends are swapped
		d.DefPoint2, d.DefPoint3 = d.DefPoint3, d.DefPoint2
	}
	if d.Type != DIM_ANGULAR && d.Type != DIM_ANGULAR3P {
		d.Measurement = t.Length(d.Measurement)
	}
	d.LeaderLength = t.Length(d.LeaderLength)
	return nil
}

// ApplyTransform transforms the leader (conformal transformations only).
func (l *Leader) ApplyTransform(t Transform) error {
	if err := checkZ(t, l.Extrusion); err != nil {
		return err
	}
	if !t.Conformal() {
		return ErrNonUniform
	}
	for _, v := range l.Vertices {
		transformPoint(t, v)
	}
	if len(l.HorizontalDirection) >= 2 {
		hd := t.Vector(l.HorizontalDirection)
		if n := math.Hypot(hd[0], hd[1]); n > 0 {
			hd[0], hd[1] = hd[0]/n, hd[1]/n
		}
		copy(l.HorizontalDirection, hd)
	}
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
// model space is not changed). Its sides stay parallel to the axes: they
// are scaled when the transformation keeps them parallel (rotations by
// multiples of 90 degrees swap width and height).
func (v *Viewport) ApplyTransform(t Transform) error {
	ax, ay := t.apply2(v.Width, 0)
	bx, by := t.apply2(0, v.Height)
	small := 1e-9 * (math.Abs(v.Width) + math.Abs(v.Height))
	switch {
	case math.Abs(ay) <= small && math.Abs(bx) <= small:
		v.Width, v.Height = math.Abs(ax), math.Abs(by)
	case math.Abs(ax) <= small && math.Abs(by) <= small:
		v.Width, v.Height = math.Abs(bx), math.Abs(ay)
	case t.Conformal():
		v.Width = t.Length(v.Width)
		v.Height = t.Length(v.Height)
	default:
		return ErrNonUniform
	}
	transformPoint(t, v.Center)
	return nil
}

// ApplyTransform returns ErrTransform: the data of unknown entities is
// not interpreted.
func (u *Unknown) ApplyTransform(t Transform) error {
	return ErrTransform
}

// ApplyTransform transforms the underlay: insertion point, scale and
// rotation (the clipping boundary is in underlay coordinates). Mirroring
// makes the Y scale negative.
func (u *Underlay) ApplyTransform(t Transform) error {
	if err := checkZ(t, u.Extrusion); err != nil {
		return err
	}
	rot, kx, ky, ok := t.frame(u.Rotation)
	if !ok {
		return ErrNonUniform
	}
	_, _, sz := t.Factors()
	transformPoint(t, u.Insert)
	u.Scale[0] *= kx
	u.Scale[1] *= ky
	u.Scale[2] *= sz
	u.Rotation = normDeg(rot)
	return nil
}
