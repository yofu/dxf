package entity

import (
	"github.com/yofu/dxf/format"
)

// Point represents POINT Entity.
type Point struct {
	*entity
	Coord []float64 // 10, 20, 30
}

// IsEntity is for Entity interface.
func (p *Point) IsEntity() bool {
	return true
}

// NewPoint creates a new Point.
func NewPoint() *Point {
	p := &Point{
		entity: NewEntity(POINT),
		Coord:  []float64{0.0, 0.0, 0.0},
	}
	return p
}

// Format writes data to formatter.
func (p *Point) Format(f format.Formatter) {
	p.entity.Format(f)
	f.WriteString(100, "AcDbPoint")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, p.Coord[i])
	}
}

// String outputs data using default formatter.
func (p *Point) String() string {
	f := format.NewASCII()
	return p.FormatString(f)
}

// FormatString outputs data using given formatter.
func (p *Point) FormatString(f format.Formatter) string {
	p.Format(f)
	return f.Output()
}

func (p *Point) BBox() ([]float64, []float64) {
	return []float64{p.Coord[0], p.Coord[1], p.Coord[2]}, []float64{p.Coord[0], p.Coord[1], p.Coord[2]}
}
