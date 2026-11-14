package entity

import (
	"math"

	"github.com/yofu/dxf/format"
)

// writePoint3 writes a 3D point; missing coordinates are 0.
func writePoint3(f format.Formatter, code int, p []float64) {
	for i := 0; i < 3; i++ {
		f.WriteFloat(code+i*10, coord(p, i))
	}
}

// pointsBBox returns the bounding box of points (3D).
func pointsBBox(pts ...[]float64) ([]float64, []float64) {
	mins := []float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxs := []float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, p := range pts {
		for i := 0; i < 3; i++ {
			mins[i] = math.Min(mins[i], coord(p, i))
			maxs[i] = math.Max(maxs[i], coord(p, i))
		}
	}
	if len(pts) == 0 {
		return []float64{0, 0, 0}, []float64{0, 0, 0}
	}
	return mins, maxs
}

// XLine represents XLINE (infinite line) and RAY (semi-infinite line)
// entities.
type XLine struct {
	*entity
	Point     []float64 // 10, 20, 30: a point on the line (start point of a ray)
	Direction []float64 // 11, 21, 31: unit direction vector
}

// IsEntity is for Entity interface.
func (x *XLine) IsEntity() bool { return true }

// NewXLine creates an XLINE through point in direction (normalized).
func NewXLine(point, direction []float64) *XLine {
	x := &XLine{entity: NewEntity(XLINE), Point: []float64{coord(point, 0), coord(point, 1), coord(point, 2)}}
	x.Direction = normalize3(direction)
	return x
}

// NewRay creates a RAY from point in direction (normalized).
func NewRay(point, direction []float64) *XLine {
	x := NewXLine(point, direction)
	x.Type = RAY
	return x
}

func normalize3(v []float64) []float64 {
	x, y, z := coord(v, 0), coord(v, 1), coord(v, 2)
	n := math.Sqrt(x*x + y*y + z*z)
	if n == 0 {
		return []float64{1, 0, 0}
	}
	return []float64{x / n, y / n, z / n}
}

// IsRay reports whether the entity is a RAY.
func (x *XLine) IsRay() bool { return x.Type == RAY }

// Format writes data to formatter.
func (x *XLine) Format(f format.Formatter) {
	x.entity.Format(f)
	if x.IsRay() {
		f.WriteString(100, "AcDbRay")
	} else {
		f.WriteString(100, "AcDbXline")
	}
	writePoint3(f, 10, x.Point)
	writePoint3(f, 11, x.Direction)
}

// String outputs data using default formatter.
func (x *XLine) String() string { return x.FormatString(format.NewASCII()) }

// FormatString outputs data using given formatter.
func (x *XLine) FormatString(f format.Formatter) string {
	x.Format(f)
	return f.Output()
}

// BBox returns the point (the line is infinite).
func (x *XLine) BBox() ([]float64, []float64) { return pointsBBox(x.Point) }

// ApplyTransform transforms the line (any transformation).
func (x *XLine) ApplyTransform(t Transform) error {
	transformPoint(t, x.Point)
	x.Direction = normalize3(t.Vector(x.Direction))
	return nil
}

// Shape represents SHAPE entities: a shape of a shape file (a STYLE with
// flag 1) by name.
type Shape struct {
	*entity
	Insert    []float64 // 10, 20, 30
	Size      float64   // 40
	Name      string    // 2: shape name
	Thickness float64   // 39
	Rotation  float64   // 50 (degrees)
	XScale    float64   // 41: relative X scale factor
	Oblique   float64   // 51 (degrees)
	Extrusion []float64 // 210, 220, 230
}

// IsEntity is for Entity interface.
func (s *Shape) IsEntity() bool { return true }

// NewShape creates a shape of name at insert with size.
func NewShape(name string, insert []float64, size float64) *Shape {
	return &Shape{entity: NewEntity(SHAPE), Name: name, Insert: []float64{coord(insert, 0), coord(insert, 1), coord(insert, 2)},
		Size: size, XScale: 1, Extrusion: []float64{0, 0, 1}}
}

// Format writes data to formatter.
func (s *Shape) Format(f format.Formatter) {
	s.entity.Format(f)
	f.WriteString(100, "AcDbShape")
	writePoint3(f, 10, s.Insert)
	f.WriteFloat(40, s.Size)
	f.WriteString(2, s.Name)
	if s.Thickness != 0 {
		f.WriteFloat(39, s.Thickness)
	}
	if s.Rotation != 0 {
		f.WriteFloat(50, s.Rotation)
	}
	if s.XScale != 1 {
		f.WriteFloat(41, s.XScale)
	}
	if s.Oblique != 0 {
		f.WriteFloat(51, s.Oblique)
	}
	if coord(s.Extrusion, 0) != 0 || coord(s.Extrusion, 1) != 0 || coord(s.Extrusion, 2) != 1 {
		writePoint3(f, 210, s.Extrusion)
	}
}

// String outputs data using default formatter.
func (s *Shape) String() string { return s.FormatString(format.NewASCII()) }

// FormatString outputs data using given formatter.
func (s *Shape) FormatString(f format.Formatter) string {
	s.Format(f)
	return f.Output()
}

// BBox returns the insertion point (the shape geometry is in the shape file).
func (s *Shape) BBox() ([]float64, []float64) { return pointsBBox(s.Insert) }

// ApplyTransform transforms the shape like text (see Text.ApplyTransform;
// a mirrored shape gets a negative X scale).
func (s *Shape) ApplyTransform(t Transform) error {
	if err := checkZ(t, s.Extrusion); err != nil {
		return err
	}
	rot, kx, ky, ok := t.frame(s.Rotation)
	if !ok {
		return ErrNonUniform
	}
	transformPoint(t, s.Insert)
	s.Size *= math.Abs(ky)
	s.XScale *= kx / math.Abs(ky)
	if ky < 0 { // mirrored: flip along the baseline
		s.XScale = -s.XScale
		rot += 180
	}
	s.Rotation = normDeg(rot)
	s.Thickness = t.ZLength(s.Thickness)
	return nil
}

// Tolerance represents TOLERANCE entities (feature control frames). Content
// uses the format codes of the DXF reference (e.g. "{\Fgdt;j}%%v0.1").
type Tolerance struct {
	*entity
	Style     string    // 3: dimension style name
	Insert    []float64 // 10, 20, 30
	Content   string    // 1
	Extrusion []float64 // 210, 220, 230
	XAxis     []float64 // 11, 21, 31: X axis direction
}

// IsEntity is for Entity interface.
func (tl *Tolerance) IsEntity() bool { return true }

// NewTolerance creates a feature control frame with content at insert.
func NewTolerance(content string, insert []float64) *Tolerance {
	return &Tolerance{entity: NewEntity(TOLERANCE), Style: "Standard", Content: content,
		Insert:    []float64{coord(insert, 0), coord(insert, 1), coord(insert, 2)},
		Extrusion: []float64{0, 0, 1}, XAxis: []float64{1, 0, 0}}
}

// Format writes data to formatter.
func (tl *Tolerance) Format(f format.Formatter) {
	tl.entity.Format(f)
	f.WriteString(100, "AcDbFcf")
	f.WriteString(3, tl.Style)
	writePoint3(f, 10, tl.Insert)
	f.WriteString(1, tl.Content)
	if coord(tl.Extrusion, 0) != 0 || coord(tl.Extrusion, 1) != 0 || coord(tl.Extrusion, 2) != 1 {
		writePoint3(f, 210, tl.Extrusion)
	}
	writePoint3(f, 11, tl.XAxis)
}

// String outputs data using default formatter.
func (tl *Tolerance) String() string { return tl.FormatString(format.NewASCII()) }

// FormatString outputs data using given formatter.
func (tl *Tolerance) FormatString(f format.Formatter) string {
	tl.Format(f)
	return f.Output()
}

// BBox returns the insertion point.
func (tl *Tolerance) BBox() ([]float64, []float64) { return pointsBBox(tl.Insert) }

// ApplyTransform transforms the insertion point and the X axis
// (conformal transformations only; the frame stays readable).
func (tl *Tolerance) ApplyTransform(t Transform) error {
	if err := checkZ(t, tl.Extrusion); err != nil {
		return err
	}
	if !t.Conformal() {
		return ErrNonUniform
	}
	transformPoint(t, tl.Insert)
	x := normalize3(t.Vector(tl.XAxis))
	if t.Mirrored() { // readable: reversed X axis
		x[0], x[1] = -x[0], -x[1]
	}
	tl.XAxis = x
	return nil
}

// Light types (code 70).
const (
	LIGHT_DISTANT = 1
	LIGHT_POINT   = 2
	LIGHT_SPOT    = 3
)

// Light represents LIGHT entities (AC1021 or later).
type Light struct {
	*entity
	Version           int       // 90
	Name              string    // 1
	LightType         int       // 70: LIGHT_DISTANT, LIGHT_POINT, LIGHT_SPOT
	On                bool      // 290: status
	PlotGlyph         bool      // 291
	Intensity         float64   // 40
	Location          []float64 // 10, 20, 30
	Target            []float64 // 11, 21, 31
	AttenuationType   int       // 72: 0 none, 1 inverse linear, 2 inverse square
	UseLimits         bool      // 292
	AttenuationStart  float64   // 41
	AttenuationEnd    float64   // 42
	HotspotAngle      float64   // 50 (degrees)
	FalloffAngle      float64   // 51 (degrees)
	CastShadows       bool      // 293
	ShadowType        int       // 73: 0 ray traced, 1 shadow maps
	ShadowMapSize     int       // 91
	ShadowMapSoftness int       // 280
}

// IsEntity is for Entity interface.
func (l *Light) IsEntity() bool { return true }

// NewLight creates a light of type lightType at location.
func NewLight(name string, lightType int, location []float64) *Light {
	return &Light{entity: NewEntity(LIGHT), Name: name, LightType: lightType, On: true, Intensity: 1,
		Location:     []float64{coord(location, 0), coord(location, 1), coord(location, 2)},
		Target:       []float64{coord(location, 0), coord(location, 1), coord(location, 2) - 1},
		HotspotAngle: 45, FalloffAngle: 50, CastShadows: true, ShadowMapSize: 256, ShadowMapSoftness: 1}
}

// Format writes data to formatter.
func (l *Light) Format(f format.Formatter) {
	l.entity.Format(f)
	f.WriteString(100, "AcDbLight")
	f.WriteInt(90, l.Version)
	f.WriteString(1, l.Name)
	f.WriteInt(70, l.LightType)
	f.WriteInt(290, boolInt(l.On))
	f.WriteInt(291, boolInt(l.PlotGlyph))
	f.WriteFloat(40, l.Intensity)
	writePoint3(f, 10, l.Location)
	writePoint3(f, 11, l.Target)
	f.WriteInt(72, l.AttenuationType)
	f.WriteInt(292, boolInt(l.UseLimits))
	f.WriteFloat(41, l.AttenuationStart)
	f.WriteFloat(42, l.AttenuationEnd)
	f.WriteFloat(50, l.HotspotAngle)
	f.WriteFloat(51, l.FalloffAngle)
	f.WriteInt(293, boolInt(l.CastShadows))
	f.WriteInt(73, l.ShadowType)
	f.WriteInt(91, l.ShadowMapSize)
	f.WriteInt(280, l.ShadowMapSoftness)
}

// String outputs data using default formatter.
func (l *Light) String() string { return l.FormatString(format.NewASCII()) }

// FormatString outputs data using given formatter.
func (l *Light) FormatString(f format.Formatter) string {
	l.Format(f)
	return f.Output()
}

// BBox returns the bounding box of location and target.
func (l *Light) BBox() ([]float64, []float64) { return pointsBBox(l.Location, l.Target) }

// ApplyTransform transforms location and target.
func (l *Light) ApplyTransform(t Transform) error {
	transformPoint(t, l.Location)
	transformPoint(t, l.Target)
	return nil
}
