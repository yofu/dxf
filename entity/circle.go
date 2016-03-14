package entity

import (
	"github.com/yofu/dxf/format"
)

// Circle represents CIRCLE Entity.
type Circle struct {
	*entity
	Center    []float64 // 10, 20, 30
	Radius    float64   // 40
	Direction []float64 // 210, 220, 230
}

// IsEntity is for Entity interface.
func (c *Circle) IsEntity() bool {
	return true
}

// NewCircle creates a new Circle.
func NewCircle() *Circle {
	c := &Circle{
		entity:    NewEntity(CIRCLE),
		Center:    []float64{0.0, 0.0, 0.0},
		Radius:    0.0,
		Direction: []float64{0.0, 0.0, 1.0},
	}
	return c
}

// Format writes data to formatter.
func (c *Circle) Format(f format.Formatter) {
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

// String outputs data using default formatter.
func (c *Circle) String() string {
	f := format.NewASCII()
	return c.FormatString(f)
}

// FormatString outputs data using given formatter.
func (c *Circle) FormatString(f format.Formatter) string {
	c.Format(f)
	return f.Output()
}

// CurrentDirection returns extrusion direction.
func (c *Circle) CurrentDirection() []float64 {
	return c.Direction
}

// SetDirection sets new extrusion direction.
func (c *Circle) SetDirection(d []float64) {
	c.Direction = d
}

// CurrentCoord returns center point coord.
func (c *Circle) CurrentCoord() []float64 {
	return c.Center
}

// SetCoord sets new center point coord.
func (c *Circle) SetCoord(co []float64) {
	c.Center = co
}

func (c *Circle) BBox() ([]float64, []float64) {
	// TODO: extrusion
	return []float64{c.Center[0] - c.Radius, c.Center[1] - c.Radius, c.Center[2]}, []float64{c.Center[0] + c.Radius, c.Center[1] + c.Radius, c.Center[2]}
}
