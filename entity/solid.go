package entity

import (
	"math"

	"github.com/yofu/dxf/format"
)

// Solid represents SOLID Entity (filled triangle or quadrilateral).
//
// Note the vertex order of SOLID: a quadrilateral p0-p1-p3-p2 is filled
// (the third and fourth points are crossed). For a triangle, the fourth
// point is the same as the third.
type Solid struct {
	*entity
	Points    [][]float64 // 10, 11, 12, 13 (OCS)
	Thickness float64     // 39
	Extrusion []float64   // 210, 220, 230
}

// IsEntity is for Entity interface.
func (s *Solid) IsEntity() bool {
	return true
}

// NewSolid creates a new Solid.
func NewSolid() *Solid {
	return &Solid{
		entity: NewEntity(SOLID),
		Points: [][]float64{
			{0.0, 0.0, 0.0},
			{0.0, 0.0, 0.0},
			{0.0, 0.0, 0.0},
			{0.0, 0.0, 0.0},
		},
		Extrusion: []float64{0.0, 0.0, 1.0},
	}
}

// Format writes data to formatter.
func (s *Solid) Format(f format.Formatter) {
	s.entity.Format(f)
	f.WriteString(100, "AcDbTrace")
	for i := 0; i < 4; i++ {
		for j := 0; j < 3; j++ {
			f.WriteFloat((j+1)*10+i, s.Points[i][j])
		}
	}
	if s.Thickness != 0 {
		f.WriteFloat(39, s.Thickness)
	}
	if s.Extrusion[0] != 0 || s.Extrusion[1] != 0 || s.Extrusion[2] != 1 {
		for i := 0; i < 3; i++ {
			f.WriteFloat(200+(i+1)*10, s.Extrusion[i])
		}
	}
}

// String outputs data using default formatter.
func (s *Solid) String() string {
	f := format.NewASCII()
	return s.FormatString(f)
}

// FormatString outputs data using given formatter.
func (s *Solid) FormatString(f format.Formatter) string {
	s.Format(f)
	return f.Output()
}

// BBox returns the bounding box of the points.
func (s *Solid) BBox() ([]float64, []float64) {
	mins := []float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxs := []float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, p := range s.Points {
		for i := 0; i < 3; i++ {
			mins[i] = math.Min(mins[i], p[i])
			maxs[i] = math.Max(maxs[i], p[i])
		}
	}
	return mins, maxs
}
