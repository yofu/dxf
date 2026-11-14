package entity

import (
	"math"

	"github.com/yofu/dxf/format"
)

// Helix represents HELIX entities: a spline (the curve) with the helix
// parameters.
type Helix struct {
	*Spline
	MajorRelease int       // 90
	Maintenance  int       // 91
	AxisBase     []float64 // 10, 20, 30: base point of the axis
	StartPoint   []float64 // 11, 21, 31: start point (base radius = distance from the axis)
	Axis         []float64 // 12, 22, 32: axis direction
	Radius       float64   // 40: top radius
	Turns        float64   // 41
	TurnHeight   float64   // 42
	RightHanded  bool      // 290: handedness (counterclockwise)
	Constrain    int       // 280: 0 turn height, 1 turns, 2 height
}

// NewHelix creates a helix around the Z axis through the origin with
// radius, pitch (height of one turn; negative: downwards) and turns,
// counterclockwise if ccw. The curve is a cubic B-spline made of four
// Bezier segments per turn (as ezdxf creates it).
func NewHelix(radius, pitch, turns float64, ccw bool) *Helix {
	s := NewSpline()
	s.Type = HELIX
	h := &Helix{
		Spline:       s,
		MajorRelease: 29,
		Maintenance:  63,
		AxisBase:     []float64{0, 0, 0},
		StartPoint:   []float64{radius, 0, 0},
		Axis:         []float64{0, 0, 1},
		Radius:       radius,
		Turns:        turns,
		TurnHeight:   pitch,
		RightHanded:  ccw,
		Constrain:    1,
	}
	if pitch < 0 {
		h.Axis[2] = -1
	}
	h.Degree = 3
	h.Flag = 0
	sign := 1.0
	if !ccw {
		sign = -1
	}
	segments := int(math.Ceil(turns * 4))
	if segments < 1 {
		segments = 1
	}
	total := turns * 2 * math.Pi
	step := total / float64(segments)
	k := 4.0 / 3.0 * math.Tan(step/4) * radius
	point := func(a float64) []float64 {
		return []float64{radius * math.Cos(a), sign * radius * math.Sin(a), pitch * a / (2 * math.Pi)}
	}
	tangent := func(a float64) []float64 {
		return []float64{-k * math.Sin(a), sign * k * math.Cos(a), pitch * step / (2 * math.Pi) / 3}
	}
	h.Controls = [][]float64{point(0)}
	h.Knots = []float64{0, 0, 0, 0}
	for i := 0; i < segments; i++ {
		a0, a1 := float64(i)*step, float64(i+1)*step
		p0, p3 := point(a0), point(a1)
		t0, t1 := tangent(a0), tangent(a1)
		h.Controls = append(h.Controls,
			[]float64{p0[0] + t0[0], p0[1] + t0[1], p0[2] + t0[2]},
			[]float64{p3[0] - t1[0], p3[1] - t1[1], p3[2] - t1[2]},
			p3)
		if i < segments-1 {
			h.Knots = append(h.Knots, float64(i+1), float64(i+1), float64(i+1))
		}
	}
	n := float64(segments)
	h.Knots = append(h.Knots, n, n, n, n)
	return h
}

// IsEntity is for Entity interface.
func (h *Helix) IsEntity() bool { return true }

// Format writes data to formatter.
func (h *Helix) Format(f format.Formatter) {
	h.Spline.Format(f)
	f.WriteString(100, "AcDbHelix")
	f.WriteInt(90, h.MajorRelease)
	f.WriteInt(91, h.Maintenance)
	writePoint3(f, 10, h.AxisBase)
	writePoint3(f, 11, h.StartPoint)
	writePoint3(f, 12, h.Axis)
	f.WriteFloat(40, h.Radius)
	f.WriteFloat(41, h.Turns)
	f.WriteFloat(42, h.TurnHeight)
	f.WriteInt(290, boolInt(h.RightHanded))
	f.WriteInt(280, h.Constrain)
}

// String outputs data using default formatter.
func (h *Helix) String() string { return h.FormatString(format.NewASCII()) }

// FormatString outputs data using given formatter.
func (h *Helix) FormatString(f format.Formatter) string {
	h.Format(f)
	return f.Output()
}

// ApplyTransform transforms the curve and the helix parameters
// (conformal transformations only; a mirror changes the handedness).
func (h *Helix) ApplyTransform(t Transform) error {
	if !t.Conformal() {
		return ErrNonUniform
	}
	if err := h.Spline.ApplyTransform(t); err != nil {
		return err
	}
	transformPoint(t, h.AxisBase)
	transformPoint(t, h.StartPoint)
	h.Axis = normalize3(t.Vector(h.Axis))
	h.Radius = t.Length(h.Radius)
	h.TurnHeight = t.ZLength(h.TurnHeight)
	if t.Mirrored() {
		h.RightHanded = !h.RightHanded
	}
	return nil
}

// Mesh represents MESH entities (subdivision meshes).
type Mesh struct {
	*entity
	Version     int         // 71
	BlendCrease bool        // 72
	Subdivision int         // 91: subdivision levels
	Vertices    [][]float64 // 92, 10, 20, 30
	Faces       [][]int     // 93, 90: vertex indices of each face
	Edges       [][2]int    // 94, 90
	Creases     []float64   // 95, 140: crease of each edge
}

// IsEntity is for Entity interface.
func (m *Mesh) IsEntity() bool { return true }

// NewMesh creates a mesh of vertices and faces (vertex indices).
func NewMesh(vertices [][]float64, faces [][]int) *Mesh {
	return &Mesh{entity: NewEntity(MESH), Version: 2, Vertices: vertices, Faces: faces}
}

// Format writes data to formatter.
func (m *Mesh) Format(f format.Formatter) {
	m.entity.Format(f)
	f.WriteString(100, "AcDbSubDMesh")
	f.WriteInt(71, m.Version)
	f.WriteInt(72, boolInt(m.BlendCrease))
	f.WriteInt(91, m.Subdivision)
	f.WriteInt(92, len(m.Vertices))
	for _, v := range m.Vertices {
		writePoint3(f, 10, v)
	}
	n := 0
	for _, face := range m.Faces {
		n += len(face) + 1
	}
	f.WriteInt(93, n)
	for _, face := range m.Faces {
		f.WriteInt(90, len(face))
		for _, i := range face {
			f.WriteInt(90, i)
		}
	}
	f.WriteInt(94, len(m.Edges))
	for _, e := range m.Edges {
		f.WriteInt(90, e[0])
		f.WriteInt(90, e[1])
	}
	// the number of creases must be the number of edges (AutoCAD)
	f.WriteInt(95, len(m.Edges))
	for i := range m.Edges {
		c := 0.0
		if i < len(m.Creases) {
			c = m.Creases[i]
		}
		f.WriteFloat(140, c)
	}
	f.WriteInt(90, 0) // no overridden properties
}

// String outputs data using default formatter.
func (m *Mesh) String() string { return m.FormatString(format.NewASCII()) }

// FormatString outputs data using given formatter.
func (m *Mesh) FormatString(f format.Formatter) string {
	m.Format(f)
	return f.Output()
}

// BBox returns the bounding box of the vertices.
func (m *Mesh) BBox() ([]float64, []float64) { return pointsBBox(m.Vertices...) }

// ApplyTransform transforms the vertices (any transformation).
func (m *Mesh) ApplyTransform(t Transform) error {
	for _, v := range m.Vertices {
		transformPoint(t, v)
	}
	return nil
}
