package entity

import (
	"github.com/yofu/dxf/format"
)

// Polyline represents POLYLINE Entity.
type Polyline struct {
	*entity
	Flag      int
	size      int
	Vertices  []*Vertex
	endhandle int
}

// IsEntity is for Entity interface.
func (p *Polyline) IsEntity() bool {
	return true
}

// NewPolyline creates a new Polyline.
func NewPolyline() *Polyline {
	vs := make([]*Vertex, 0)
	p := &Polyline{
		entity:    NewEntity(POLYLINE),
		Flag:      8,
		size:      0,
		Vertices:  vs,
		endhandle: 0,
	}
	return p
}

// Format writes data to formatter.
func (p *Polyline) Format(f format.Formatter) {
	p.entity.Format(f)
	f.WriteString(100, "AcDb3dPolyline")
	f.WriteInt(66, 1)
	f.WriteString(10, "0.0")
	f.WriteString(20, "0.0")
	f.WriteString(30, "0.0")
	f.WriteInt(70, p.Flag)
	for _, v := range p.Vertices {
		v.Format(f)
	}
	f.WriteString(0, "SEQEND")
	f.WriteHex(5, p.endhandle)
	f.WriteString(100, "AcDbEntity")
	f.WriteString(8, p.Layer().Name())
}

// String outputs data using default formatter.
func (p *Polyline) String() string {
	f := format.NewASCII()
	return p.FormatString(f)
}

// FormatString outputs data using given formatter.
func (p *Polyline) FormatString(f format.Formatter) string {
	p.Format(f)
	return f.Output()
}

// Close closes Polyline.
func (p *Polyline) Close() {
	p.Flag |= 1
}

// AddVertex adds a new vertex to Polyline.
func (p *Polyline) AddVertex(x, y, z float64) *Vertex {
	v := NewVertex(x, y, z)
	p.Vertices = append(p.Vertices, v)
	p.size++
	v.SetLayer(p.Layer())
	v.SetOwner(p)
	return v
}

// SetHandle sets handles to itself and its vertices.
func (p *Polyline) SetHandle(h *int) {
	p.entity.SetHandle(h)
	for _, v := range p.Vertices {
		v.SetHandle(h)
	}
	p.endhandle = *h
	(*h)++
}

func (p *Polyline) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)
	for _, v := range p.Vertices {
		for i := 0; i < 3; i++ {
			if v.Coord[i] < mins[i] {
				mins[i] = v.Coord[i]
			}
			if v.Coord[i] > maxs[i] {
				maxs[i] = v.Coord[i]
			}
		}
	}
	return mins, maxs
}
