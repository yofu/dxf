package entity

import (
	"bytes"
	"fmt"
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
		NewEntity(POINT),
		[]float64{0.0, 0.0, 0.0},
	}
	return p
}

func (p *Point) String() string {
	var otp bytes.Buffer
	otp.WriteString(p.entity.String())
	otp.WriteString("100\nAcDbPoint\n")
	for i := 0; i < 3; i++ {
		otp.WriteString(fmt.Sprintf("%d\n%f\n", (i+1)*10, p.Coord[i]))
	}
	return otp.String()
}

