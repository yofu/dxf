package entity

import (
	"bytes"
	"fmt"
)

type Circle struct {
	*entity
	Center []float64 // 10, 20, 30
	Radius float64   // 40
}

func (c *Circle) IsEntity() bool {
	return true
}

func NewCircle() *Circle {
	c := &Circle{
		NewEntity(CIRCLE),
		[]float64{0.0, 0.0, 0.0},
		0.0,
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
	return otp.String()
}
