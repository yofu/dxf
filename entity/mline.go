package entity

import (
	"errors"
	"math"

	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Justifications of MLINE (code 70).
const (
	MLINE_TOP    = 0
	MLINE_ZERO   = 1
	MLINE_BOTTOM = 2
)

// Flags of MLINE (code 71).
const (
	MLINE_HAS_VERTICES        = 1
	MLINE_CLOSED              = 2
	MLINE_SUPPRESS_START_CAPS = 4
	MLINE_SUPPRESS_END_CAPS   = 8
)

// Flags of MLINESTYLE (code 70).
const (
	MLINESTYLE_FILL            = 1
	MLINESTYLE_MITER           = 2
	MLINESTYLE_START_SQUARE    = 16
	MLINESTYLE_START_INNER_ARC = 32
	MLINESTYLE_START_ROUND     = 64
	MLINESTYLE_END_SQUARE      = 256
	MLINESTYLE_END_INNER_ARC   = 512
	MLINESTYLE_END_ROUND       = 1024
)

// MLineStyleElement is a line element of an MLINESTYLE.
type MLineStyleElement struct {
	Offset   float64 // 49: offset from the reference line (positive: left)
	Color    int     // 62: color number
	LineType string  // 6: linetype name
}

// MLineStyle represents MLINESTYLE objects (dictionary ACAD_MLINESTYLE).
// Use Drawing.AddMLineStyle to create one.
type MLineStyle struct {
	Flags       int                 // 70: MLINESTYLE_*
	Description string              // 3
	FillColor   int                 // 62: fill color number
	StartAngle  float64             // 51 (degrees)
	EndAngle    float64             // 52 (degrees)
	Elements    []MLineStyleElement // 71, 49, 62, 6

	name   string
	handle int
	src    int
	owner  handle.Handler
}

// NewMLineStyle creates an MLINESTYLE with elements.
func NewMLineStyle(name string, elements ...MLineStyleElement) *MLineStyle {
	return &MLineStyle{name: name, FillColor: 256, StartAngle: 90, EndAngle: 90, Elements: elements}
}

// Name returns the name of the style.
func (s *MLineStyle) Name() string { return s.name }

// IsObject is for object.Object interface.
func (s *MLineStyle) IsObject() bool { return true }

// Format writes data to formatter.
func (s *MLineStyle) Format(f format.Formatter) {
	f.WriteString(0, "MLINESTYLE")
	f.WriteHex(5, s.handle)
	writeOwner(f, s.owner, true)
	f.WriteString(100, "AcDbMlineStyle")
	f.WriteString(2, s.name)
	f.WriteInt(70, s.Flags)
	f.WriteString(3, s.Description)
	f.WriteInt(62, s.FillColor)
	f.WriteFloat(51, s.StartAngle)
	f.WriteFloat(52, s.EndAngle)
	f.WriteInt(71, len(s.Elements))
	for _, e := range s.Elements {
		f.WriteFloat(49, e.Offset)
		f.WriteInt(62, e.Color)
		lt := e.LineType
		if lt == "" {
			lt = "BYLAYER"
		}
		f.WriteString(6, lt)
	}
}

// writeOwner writes the owner of an object: 102 ACAD_REACTORS group (if
// reactors) and 330.
func writeOwner(f format.Formatter, owner handle.Handler, reactors bool) {
	if owner == nil {
		f.WriteString(330, "0")
		return
	}
	if reactors {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteHex(330, owner.Handle())
		f.WriteString(102, "}")
	}
	f.WriteHex(330, owner.Handle())
}

// Handle returns the handle.
func (s *MLineStyle) Handle() int { return s.handle }

// SetHandle sets the handle (the handle read is kept).
func (s *MLineStyle) SetHandle(v *int) { s.handle = handle.Next(s.src, v) }

// SetOwner sets the owner dictionary.
func (s *MLineStyle) SetOwner(h handle.Handler) { s.owner = h }

// SourceHandle returns the handle in the file it was read from ("" if new).
func (s *MLineStyle) SourceHandle() string { return handle.Format(s.src) }

// SetSourceHandle sets the handle in the file it was read from.
func (s *MLineStyle) SetSourceHandle(h string) { s.src = handle.Parse(h) }

// MLineVertex is a vertex of an MLINE with the geometry of the elements
// of the segment starting at it.
type MLineVertex struct {
	Location       []float64   // 11, 21, 31
	LineDirection  []float64   // 12, 22, 32: direction of the segment
	MiterDirection []float64   // 13, 23, 33: direction of the miter
	LineParams     [][]float64 // 74, 41: per element: offset along the miter, start along the line, dash, gap, ...
	FillParams     [][]float64 // 75, 42: per element
}

// MLine represents MLINE entities (multilines). The geometry of the
// elements is stored in the vertices (see Generate).
type MLine struct {
	*entity
	Style         *MLineStyle    // 2, 340
	Scale         float64        // 40
	Justification int            // 70: MLINE_TOP, MLINE_ZERO, MLINE_BOTTOM
	Flags         int            // 71: MLINE_CLOSED, ...
	Extrusion     []float64      // 210, 220, 230
	Vertices      []*MLineVertex // 72, 11, ...

	styleName    string // 2 as read, until resolved
	styleHandle  string // 340 as read, until resolved
	elementCount int    // 73 as read (used when the style is not resolved)
}

// IsEntity is for Entity interface.
func (m *MLine) IsEntity() bool { return true }

// NewMLine creates an MLINE of style.
func NewMLine(style *MLineStyle) *MLine {
	return &MLine{entity: NewEntity(MLINE), Style: style, Scale: 1, Justification: MLINE_TOP, Flags: MLINE_HAS_VERTICES,
		Extrusion: []float64{0, 0, 1}}
}

// SetStyleReference sets the style name (2) and handle (340) as read;
// Style takes precedence.
func (m *MLine) SetStyleReference(name, h string, elements int) {
	m.styleName, m.styleHandle, m.elementCount = name, h, elements
}

// StyleReference returns the style name and handle as read.
func (m *MLine) StyleReference() (string, string) { return m.styleName, m.styleHandle }

// IsClosed reports whether the multiline is closed.
func (m *MLine) IsClosed() bool { return m.Flags&MLINE_CLOSED != 0 }

// Generate computes the vertices of the multiline through points (XY
// plane at the z of the first point) from its style, scale and
// justification, as AutoCAD and ezdxf do.
func (m *MLine) Generate(points [][]float64) error {
	if m.Style == nil || len(m.Style.Elements) == 0 {
		return errors.New("mline: style with elements required")
	}
	var pts [][]float64
	for _, p := range points {
		q := []float64{coord(p, 0), coord(p, 1), coord(points[0], 2)}
		if len(pts) > 0 {
			l := pts[len(pts)-1]
			if math.Hypot(q[0]-l[0], q[1]-l[1]) < 1e-6 {
				continue
			}
		}
		pts = append(pts, q)
	}
	m.Vertices = nil
	if len(pts) == 0 {
		return nil
	}
	if len(pts) == 1 {
		m.Vertices = []*MLineVertex{{Location: pts[0], LineDirection: []float64{1, 0, 0}, MiterDirection: []float64{0, 1, 0}}}
		return nil
	}
	dir := func(a, b []float64) []float64 { return normalize3([]float64{b[0] - a[0], b[1] - a[1], 0}) }
	rotate := func(v []float64, deg float64) []float64 {
		r := deg * math.Pi / 180
		return []float64{v[0]*math.Cos(r) - v[1]*math.Sin(r), v[0]*math.Sin(r) + v[1]*math.Cos(r), 0}
	}
	miter := func(d1, d2 []float64) []float64 {
		s := normalize3([]float64{d1[0] + d2[0], d1[1] + d2[1], 0})
		return []float64{-s[1], s[0], 0}
	}
	var lineDirs [][]float64
	for i := 0; i+1 < len(pts); i++ {
		lineDirs = append(lineDirs, dir(pts[i], pts[i+1]))
	}
	var miters [][]float64
	if m.IsClosed() {
		lineDirs = append(lineDirs, dir(pts[len(pts)-1], pts[0]))
		miters = append(miters, miter(lineDirs[0], lineDirs[len(lineDirs)-1]))
		for i := 1; i < len(pts); i++ {
			miters = append(miters, miter(lineDirs[i-1], lineDirs[i]))
		}
	} else {
		lineDirs = append(lineDirs, lineDirs[len(lineDirs)-1])
		miters = append(miters, rotate(lineDirs[0], m.Style.StartAngle))
		for i := 1; i+1 < len(pts); i++ {
			miters = append(miters, miter(lineDirs[i-1], lineDirs[i]))
		}
		miters = append(miters, rotate(lineDirs[len(lineDirs)-1], m.Style.EndAngle))
	}
	minOff, maxOff := math.Inf(1), math.Inf(-1)
	for _, e := range m.Style.Elements {
		minOff = math.Min(minOff, e.Offset)
		maxOff = math.Max(maxOff, e.Offset)
	}
	shift := 0.0
	switch m.Justification {
	case MLINE_TOP:
		shift = -maxOff
	case MLINE_BOTTOM:
		shift = -minOff
	}
	for i, p := range pts {
		d, mt := lineDirs[i], miters[i]
		stretch := 1.0
		if s := math.Abs(d[0]*mt[1] - d[1]*mt[0]); s > 1e-12 {
			stretch = m.Scale / s
		}
		v := &MLineVertex{Location: p, LineDirection: d, MiterDirection: mt}
		for _, e := range m.Style.Elements {
			v.LineParams = append(v.LineParams, []float64{(e.Offset + shift) * stretch, 0})
			v.FillParams = append(v.FillParams, []float64{})
		}
		m.Vertices = append(m.Vertices, v)
	}
	return nil
}

// Format writes data to formatter.
func (m *MLine) Format(f format.Formatter) {
	m.entity.Format(f)
	f.WriteString(100, "AcDbMline")
	name, count := m.styleName, m.elementCount
	if m.Style != nil {
		name, count = m.Style.name, len(m.Style.Elements)
	}
	if name == "" {
		name = "Standard"
	}
	f.WriteString(2, name)
	switch {
	case m.Style != nil:
		f.WriteHex(340, m.Style.Handle())
	case m.styleHandle != "":
		f.WriteString(340, m.styleHandle)
	default:
		f.WriteString(340, "0")
	}
	f.WriteFloat(40, m.Scale)
	f.WriteInt(70, m.Justification)
	flags := m.Flags &^ MLINE_HAS_VERTICES
	if len(m.Vertices) > 0 {
		flags |= MLINE_HAS_VERTICES
	}
	f.WriteInt(71, flags)
	f.WriteInt(72, len(m.Vertices))
	f.WriteInt(73, count)
	var start []float64
	if len(m.Vertices) > 0 {
		start = m.Vertices[0].Location
	}
	writePoint3(f, 10, start)
	writePoint3(f, 210, m.Extrusion)
	for _, v := range m.Vertices {
		writePoint3(f, 11, v.Location)
		writePoint3(f, 12, v.LineDirection)
		writePoint3(f, 13, v.MiterDirection)
		for i, lp := range v.LineParams {
			f.WriteInt(74, len(lp))
			for _, p := range lp {
				f.WriteFloat(41, p)
			}
			var fp []float64
			if i < len(v.FillParams) {
				fp = v.FillParams[i]
			}
			f.WriteInt(75, len(fp))
			for _, p := range fp {
				f.WriteFloat(42, p)
			}
		}
	}
}

// String outputs data using default formatter.
func (m *MLine) String() string { return m.FormatString(format.NewASCII()) }

// FormatString outputs data using given formatter.
func (m *MLine) FormatString(f format.Formatter) string {
	m.Format(f)
	return f.Output()
}

// ElementPoints returns the points of element i at each vertex.
func (m *MLine) ElementPoints(i int) [][]float64 {
	var pts [][]float64
	for _, v := range m.Vertices {
		if i >= len(v.LineParams) || len(v.LineParams[i]) == 0 {
			continue
		}
		o := v.LineParams[i][0]
		pts = append(pts, []float64{
			coord(v.Location, 0) + o*coord(v.MiterDirection, 0),
			coord(v.Location, 1) + o*coord(v.MiterDirection, 1),
			coord(v.Location, 2) + o*coord(v.MiterDirection, 2),
		})
	}
	return pts
}

// BBox returns the bounding box of the element points.
func (m *MLine) BBox() ([]float64, []float64) {
	var pts [][]float64
	for _, v := range m.Vertices {
		pts = append(pts, v.Location)
	}
	n := 0
	if len(m.Vertices) > 0 {
		n = len(m.Vertices[0].LineParams)
	}
	for i := 0; i < n; i++ {
		pts = append(pts, m.ElementPoints(i)...)
	}
	return pointsBBox(pts...)
}

// ApplyTransform transforms the vertices and the element geometry (any
// transformation: the offsets along the miter and line directions are
// scaled with the transformed directions).
func (m *MLine) ApplyTransform(t Transform) error {
	if err := checkZ(t, m.Extrusion); err != nil {
		return err
	}
	for _, v := range m.Vertices {
		transformPoint(t, v.Location)
		ld, md := t.Vector(v.LineDirection), t.Vector(v.MiterDirection)
		kl, km := norm3(ld), norm3(md)
		v.LineDirection, v.MiterDirection = normalize3(ld), normalize3(md)
		for _, lp := range v.LineParams {
			for k := range lp {
				if k == 0 {
					lp[k] *= km
				} else {
					lp[k] *= kl
				}
			}
		}
		for _, fp := range v.FillParams {
			for k := range fp {
				fp[k] *= kl
			}
		}
	}
	if t.Conformal() {
		m.Scale = t.Length(m.Scale)
	}
	return nil
}

func norm3(v []float64) float64 {
	return math.Sqrt(coord(v, 0)*coord(v, 0) + coord(v, 1)*coord(v, 1) + coord(v, 2)*coord(v, 2))
}
