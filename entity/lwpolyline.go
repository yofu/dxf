package entity

import (
	"bytes"
	"fmt"
)

type LwPolyline struct {
	*entity
	Num      int // 90
	Closed   bool
	Vertices [][]float64
}

func (l *LwPolyline) IsEntity() bool {
	return true
}

func NewLwPolyline(size int) *LwPolyline {
	vs := make([][]float64, size)
	for i := 0; i < size; i++ {
		vs[i] = make([]float64, 2)
	}
	l := &LwPolyline{
		NewEntity(LWPOLYLINE),
		size,
		false,
		vs,
	}
	return l
}

func (l *LwPolyline) String() string {
	var otp bytes.Buffer
	otp.WriteString(l.entity.String())
	otp.WriteString("100\nAcDbPolyline\n")
	otp.WriteString(fmt.Sprintf("90\n%d\n", l.Num))
	if l.Closed {
		otp.WriteString("70\n1\n")
	} else {
		otp.WriteString("70\n0\n")
	}
	for i := 0; i < l.Num; i++ {
		for j := 0; j < 2; j++ {
			otp.WriteString(fmt.Sprintf("%d\n%f\n", (j+1)*10, l.Vertices[i][j]))
		}
	}
	return otp.String()
}

func (l *LwPolyline) Close() {
	l.Closed = true
}
