package entity

import (
	"github.com/yofu/dxf/format"
)

type Circle struct {
	*entity
	Center    []float64 // 10, 20, 30
	Radius    float64   // 40
	Direction []float64 // 210, 220, 230
}

func (c *Circle) IsEntity() bool {
	return true
}

func NewCircle() *Circle {
	c := &Circle{
		entity:    NewEntity(CIRCLE),
		Center:    []float64{0.0, 0.0, 0.0},
		Radius:    0.0,
		Direction: []float64{0.0, 0.0, 1.0},
	}
	return c
}

func ParseCircle(data [][2]string) (Entity, error) {
	c := NewCircle()
	return c, nil
}

func (c *Circle) Format(f *format.Formatter) {
	c.entity.Format(f)
	f.WriteString(100, "AcDbCircle")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, c.Center[i])
	}
	f.WriteFloat(40, c.Radius)
	for i := 0; i < 3; i++ {
		f.WriteFloat(200+(i+1)*10, c.Direction[i])
	}
}

func (c *Circle) String() string {
	f := format.New()
	return c.FormatString(f)
}

func (c *Circle) FormatString(f *format.Formatter) string {
	c.Format(f)
	return f.Output()
}

func (c *Circle) CurrentDirection() []float64 {
	return c.Direction
}
func (c *Circle) SetDirection(d []float64) {
	c.Direction = d
}
func (c *Circle) CurrentCoord() []float64 {
	return c.Center
}
func (c *Circle) SetCoord(co []float64) {
	c.Center = co
}
