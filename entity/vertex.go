package entity

import (
	"github.com/yofu/dxf/format"
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
		entity: NewEntity(VERTEX),
		Flag:   32,
		Coord:  []float64{x, y, z},
	}
	return v
}

func ParseVertex(data [][2]string) (Entity, error) {
	v := NewVertex(0.0, 0.0, 0.0)
	return v, nil
}

func (v *Vertex) Format(f *format.Formatter) {
	v.entity.Format(f)
	f.WriteString(100, "AcDbVertex")
	f.WriteString(100, "AcDb3dPolylineVertex")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, v.Coord[i])
	}
	f.WriteInt(70, v.Flag)
}

func (v *Vertex) String() string {
	f := format.New()
	return v.FormatString(f)
}

func (v *Vertex) FormatString(f *format.Formatter) string {
	v.Format(f)
	return f.Output()
}
