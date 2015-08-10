package entity

import (
	"bytes"
	"fmt"
)

type Vertex struct {
	*entity
	Flag  int
	Coord []float64
}

func (p *Vertex) IsEntity() bool {
	return true
}

func NewVertex(x, y, z float64) *Vertex {
	v := &Vertex{
		NewEntity(VERTEX),
		32,
		[]float64{x, y, z},
	}
	return v
}

func (v *Vertex) String() string {
	var otp bytes.Buffer
	otp.WriteString(v.entity.String())
	otp.WriteString("100\nAcDbVertex\n100\nAcDb3DPolylineVertex\n")
	otp.WriteString(fmt.Sprintf("70\n%d\n", v.Flag))
	for i := 0; i < 3; i++ {
		otp.WriteString(fmt.Sprintf("%d\n%f\n", (i+1)*10, v.Coord[i]))
	}
	return otp.String()
}
