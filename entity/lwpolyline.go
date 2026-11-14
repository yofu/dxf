package entity

import (
	"github.com/yofu/dxf/format"
)

// LwPolyline represents LWPOLYLINE Entity.
type LwPolyline struct {
	*entity
	Num      int // 90
	Closed   bool
	Vertices [][]float64
	Bulges   []float64 // 42
}

// IsEntity is for Entity interface.
func (l *LwPolyline) IsEntity() bool {
	return true
}

// NewLwPolyline creates a new LwPolyline.
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
		Bulges:   make([]float64, size),
	}
	return l
}

// Format writes data to formatter.
func (l *LwPolyline) Format(f format.Formatter) {
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

		if l.Bulges[i] != 0 {
			f.WriteFloat(42, l.Bulges[i])
		}
	}
}

// String outputs data using default formatter.
func (l *LwPolyline) String() string {
	f := format.NewASCII()
	return l.FormatString(f)
}

// FormatString outputs data using given formatter.
func (l *LwPolyline) FormatString(f format.Formatter) string {
	l.Format(f)
	return f.Output()
}

// Close closes LwPolyline.
func (l *LwPolyline) Close() {
	l.Closed = true
}

func (l *LwPolyline) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)
	for i := 0; i < len(mins) && len(l.Vertices) > 0; i++ {
		if i < len(l.Vertices[0]) {
			mins[i] = l.Vertices[0][i]
			maxs[i] = l.Vertices[0][i]
		}
	}
	for _, p := range l.Vertices {
		for i := 0; i < len(p) && i < 3; i++ {
			if p[i] < mins[i] {
				mins[i] = p[i]
			}
			if p[i] > maxs[i] {
				maxs[i] = p[i]
			}
		}
	}
	return mins, maxs
}

// Bulge computes the center and the radius of the arc to reach the vertex.
// The bulge of a segment is stored on its start vertex (Bulges[index-1]).
func (l LwPolyline) Bulge(index int) ([]float64, float64) {
	if index > len(l.Vertices)-1 || index < 1 {
		return nil, 0
	}

	return bulge(l.Bulges[index-1], l.Vertices[index-1], l.Vertices[index])
}
