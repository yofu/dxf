package entity

import (
	"github.com/yofu/dxf/format"
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
		entity:   NewEntity(LWPOLYLINE),
		Num:      size,
		Closed:   false,
		Vertices: vs,
	}
	return l
}

func (l *LwPolyline) Format(f *format.Formatter) {
	l.entity.Format(f)
	f.WriteString(100, "AcDbPolyline")
	f.WriteInt(90, l.Num)
	if l.Closed {
		f.WriteInt(70, 1)
	} else {
		f.WriteInt(70, 0)
	}
	for i := 0; i < l.Num; i++ {
		for j := 0; j < 2; j++ {
			f.WriteFloat((j+1)*10, l.Vertices[i][j])
		}
	}
}

func (l *LwPolyline) String() string {
	f := format.New()
	return l.FormatString(f)
}

func (l *LwPolyline) FormatString(f *format.Formatter) string {
	l.Format(f)
	return f.Output()
}

func (l *LwPolyline) Close() {
	l.Closed = true
}
