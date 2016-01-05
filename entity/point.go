package entity

import (
	"github.com/yofu/dxf/format"
)

type Point struct {
	*entity
	Coord []float64 // 10, 20, 30
}

func (p *Point) IsEntity() bool {
	return true
}

func NewPoint() *Point {
	p := &Point{
		entity: NewEntity(POINT),
		Coord:  []float64{0.0, 0.0, 0.0},
	}
	return p
}

func ParsePoint(data [][2]string) (Entity, error) {
	p := NewPoint()
	return p, nil
}

func (p *Point) Format(f *format.Formatter) {
	p.entity.Format(f)
	f.WriteString(100, "AcDbPoint")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, p.Coord[i])
	}
}

func (p *Point) String() string {
	f := format.New()
	return p.FormatString(f)
}

func (p *Point) FormatString(f *format.Formatter) string {
	p.Format(f)
	return f.Output()
}
