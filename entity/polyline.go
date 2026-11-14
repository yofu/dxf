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
	f.WriteString(100, p.subclass())
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

// subclass returns the subclass marker according to the polyline flag (code 70).
func (p *Polyline) subclass() string {
	switch {
	case p.Flag&8 != 0:
		return "AcDb3dPolyline"
	case p.Flag&16 != 0:
		return "AcDbPolygonMesh"
	case p.Flag&64 != 0:
		return "AcDbPolyFaceMesh"
	default:
		return "AcDb2dPolyline"
	}
}

// AppendVertex appends an existing Vertex (e.g. read from a file) to Polyline.
func (p *Polyline) AppendVertex(v *Vertex) {
	p.Vertices = append(p.Vertices, v)
	p.size++
	v.SetOwner(p)
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
	if len(p.Vertices) > 0 {
		copy(mins, p.Vertices[0].Coord)
		copy(maxs, p.Vertices[0].Coord)
	}
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

// Bulge computes the center and the radius of the arc from the vertex at index
// to the next vertex.
func (p Polyline) Bulge(index int) ([]float64, float64) {
	if index > len(p.Vertices)-2 {
		return nil, 0
	}

	return bulge(p.Vertices[index].Buldge, p.Vertices[index].Coord, p.Vertices[index+1].Coord)
}
