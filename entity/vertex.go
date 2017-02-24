package entity

import (
	"github.com/yofu/dxf/format"
)

// Vertex represents VERTEX Entity.
type Vertex struct {
	*entity
	Flag  int
	Coord []float64
}

// IsEntity is for Entity interface.
func (v *Vertex) IsEntity() bool {
	return true
}

// NewVertex creates a new Vertex.
func NewVertex(x, y, z float64) *Vertex {
	v := &Vertex{
		entity: NewEntity(VERTEX),
		Flag:   32,
		Coord:  []float64{x, y, z},
	}
	return v
}

// Format writes data to formatter.
func (v *Vertex) Format(f format.Formatter) {
	v.entity.Format(f)
	f.WriteString(100, "AcDbVertex")
	f.WriteString(100, "AcDb3dPolylineVertex")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, v.Coord[i])
	}
	f.WriteInt(70, v.Flag)
}

// String outputs data using default formatter.
func (v *Vertex) String() string {
	f := format.NewASCII()
	return v.FormatString(f)
}

// FormatString outputs data using given formatter.
func (v *Vertex) FormatString(f format.Formatter) string {
	v.Format(f)
	return f.Output()
}

func (v *Vertex) BBox() ([]float64, []float64) {
	mins := []float64{v.Coord[0], v.Coord[1], v.Coord[2]}
	maxs := []float64{v.Coord[0], v.Coord[1], v.Coord[2]}
	return mins, maxs
}
