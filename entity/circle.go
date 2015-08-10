package entity

import (
	"bytes"
	"fmt"
)

type Circle struct {
	*entity
	Center []float64 // 10, 20, 30
	Radius float64   // 40
	Direction []float64 // 210, 220, 230
}

func (c *Circle) IsEntity() bool {
	return true
}

func NewCircle() *Circle {
	c := &Circle{
		NewEntity(CIRCLE),
		[]float64{0.0, 0.0, 0.0},
		0.0,
		[]float64{0.0, 0.0, 1.0},
	}
	return c
}

func (c *Circle) String() string {
	var otp bytes.Buffer
	otp.WriteString(c.entity.String())
	otp.WriteString("100\nAcDbCircle\n")
	for i := 0; i < 3; i++ {
		otp.WriteString(fmt.Sprintf("%d\n%f\n", (i+1)*10, c.Center[i]))
	}
	otp.WriteString(fmt.Sprintf("40\n%f\n", c.Radius))
	for i := 0; i < 3; i++ {
		otp.WriteString(fmt.Sprintf("%d\n%f\n", 200+(i+1)*10, c.Direction[i]))
	}
	return otp.String()
}

func (c *Circle) SetDirection(d []float64) {
	c.Direction = d
}
