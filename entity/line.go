package entity

import (
	"bytes"
	"fmt"
)

type Line struct {
	entity
	Start     []float64 // 10, 20, 30
	End       []float64 // 11, 21, 31
}

func (l *Line) IsEntity() bool {
	return true
}

func NewLine() *Line {
	l := &Line{
		NewEntity(),
		[]float64{0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0},
	}
	return l
}

func (l *Line) String() string {
	var otp bytes.Buffer
	otp.WriteString(l.entity.String())
	otp.WriteString("100\nAcDbLine\n")
	for i := 0; i < 3; i++ {
		otp.WriteString(fmt.Sprintf("%d\n%f\n", (i+1)*10, l.Start[i]))
	}
	for i := 0; i < 3; i++ {
		otp.WriteString(fmt.Sprintf("%d\n%f\n", (i+1)*10+1, l.End[i]))
	}
	return otp.String()
}
