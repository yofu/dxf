package entity

import (
	"math"

	"github.com/yofu/dxf/format"
)

// Ellipse represents ELLIPSE Entity.
// A point of the ellipse at parameter t is
// Center + cos(t)*MajorAxis + sin(t)*MinorAxis(), t from Start to End.
type Ellipse struct {
	*entity
	Center    []float64 // 10, 20, 30 (WCS)
	MajorAxis []float64 // 11, 21, 31: endpoint of major axis, relative to Center (WCS)
	Extrusion []float64 // 210, 220, 230
	Ratio     float64   // 40: ratio of minor axis to major axis
	Start     float64   // 41: start parameter (radians, 0 for a full ellipse)
	End       float64   // 42: end parameter (radians, 2*pi for a full ellipse)
}

// IsEntity is for Entity interface.
func (e *Ellipse) IsEntity() bool {
	return true
}

// NewEllipse creates a new full Ellipse with unit major axis along X.
func NewEllipse() *Ellipse {
	return &Ellipse{
		entity:    NewEntity(ELLIPSE),
		Center:    []float64{0.0, 0.0, 0.0},
		MajorAxis: []float64{1.0, 0.0, 0.0},
		Extrusion: []float64{0.0, 0.0, 1.0},
		Ratio:     1.0,
		Start:     0.0,
		End:       2 * math.Pi,
	}
}

// Format writes data to formatter.
func (e *Ellipse) Format(f format.Formatter) {
	e.entity.Format(f)
	f.WriteString(100, "AcDbEllipse")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, e.Center[i])
	}
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10+1, e.MajorAxis[i])
	}
	for i := 0; i < 3; i++ {
		f.WriteFloat(200+(i+1)*10, e.Extrusion[i])
	}
	f.WriteFloat(40, e.Ratio)
	f.WriteFloat(41, e.Start)
	f.WriteFloat(42, e.End)
}

// String outputs data using default formatter.
func (e *Ellipse) String() string {
	f := format.NewASCII()
	return e.FormatString(f)
}

// FormatString outputs data using given formatter.
func (e *Ellipse) FormatString(f format.Formatter) string {
	e.Format(f)
	return f.Output()
}

// MinorAxis returns the minor axis vector: Ratio * (Extrusion x MajorAxis) / |Extrusion|.
func (e *Ellipse) MinorAxis() []float64 {
	n, m := e.Extrusion, e.MajorAxis
	c := []float64{
		n[1]*m[2] - n[2]*m[1],
		n[2]*m[0] - n[0]*m[2],
		n[0]*m[1] - n[1]*m[0],
	}
	l := math.Sqrt(n[0]*n[0] + n[1]*n[1] + n[2]*n[2])
	if l == 0 {
		l = 1
	}
	for i := range c {
		c[i] *= e.Ratio / l
	}
	return c
}

// PointAt returns the point at parameter t (radians).
func (e *Ellipse) PointAt(t float64) []float64 {
	m := e.MinorAxis()
	p := make([]float64, 3)
	for i := 0; i < 3; i++ {
		p[i] = e.Center[i] + math.Cos(t)*e.MajorAxis[i] + math.Sin(t)*m[i]
	}
	return p
}

// BBox returns the bounding box of the ellipse (or elliptical arc).
func (e *Ellipse) BBox() ([]float64, []float64) {
	start, end := e.Start, e.End
	for end <= start {
		end += 2 * math.Pi
	}
	m := e.MinorAxis()
	ts := []float64{start, end}
	// extremes of each coordinate: d/dt (cos t M_i + sin t m_i) = 0
	for i := 0; i < 3; i++ {
		t0 := math.Atan2(m[i], e.MajorAxis[i])
		for _, t := range []float64{t0, t0 + math.Pi} {
			for t < start {
				t += 2 * math.Pi
			}
			for t-2*math.Pi >= start {
				t -= 2 * math.Pi
			}
			if t <= end {
				ts = append(ts, t)
			}
		}
	}
	mins := []float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxs := []float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, t := range ts {
		p := e.PointAt(t)
		for i := 0; i < 3; i++ {
			mins[i] = math.Min(mins[i], p[i])
			maxs[i] = math.Max(maxs[i], p[i])
		}
	}
	return mins, maxs
}
