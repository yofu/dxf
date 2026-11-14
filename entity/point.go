package entity

import (
	"github.com/yofu/dxf/format"
)

// Point represents POINT Entity.
type Point struct {
	*entity
	Coord     []float64 // 10, 20, 30
	Thickness float64   // 39
	Extrusion []float64 // 210, 220, 230 (nil: 0, 0, 1)
	Angle     float64   // 50: angle of the X axis of the UCS for the point display (degrees)
}

// IsEntity is for Entity interface.
func (p *Point) IsEntity() bool {
	return true
}

// NewPoint creates a new Point.
func NewPoint(coord ...float64) *Point {
	p := &Point{
		entity:    NewEntity(POINT),
		Coord:     []float64{0.0, 0.0, 0.0},
		Extrusion: []float64{0.0, 0.0, 1.0},
	}
	ln := len(coord)
	if ln == 0 {
		return p
	}
	if ln < 3 {
		for i := ln; i < 3; i++ {
			coord = append(coord, 0.0)
		}
		p.Coord = coord
		return p
	}
	p.Coord = coord[:3]
	return p
}

// Format writes data to formatter.
func (p *Point) Format(f format.Formatter) {
	p.entity.Format(f)
	f.WriteString(100, "AcDbPoint")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, p.Coord[i])
	}
	writeThickness(f, p.Thickness)
	writeExtrusion(f, p.Extrusion)
	if p.Angle != 0 {
		f.WriteFloat(50, p.Angle)
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
	return extrude([]float64{p.Coord[0], p.Coord[1], p.Coord[2]}, []float64{p.Coord[0], p.Coord[1], p.Coord[2]}, p.Thickness)
}
