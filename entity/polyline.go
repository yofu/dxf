package entity

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Polyline represents POLYLINE Entity.
type Polyline struct {
	*entity
	Flag       int       // 70
	Thickness  float64   // 39
	Elevation  float64   // 30: elevation of 2D polylines (10 and 20 are 0)
	StartWidth float64   // 40: default starting width
	EndWidth   float64   // 41: default ending width
	MeshM      int       // 71: polygon mesh M vertex count / polyface mesh vertex count
	MeshN      int       // 72: polygon mesh N vertex count / polyface mesh face count
	SmoothM    int       // 73: smooth surface M density
	SmoothN    int       // 74: smooth surface N density
	CurveType  int       // 75: curves and smooth surface type
	Extrusion  []float64 // 210, 220, 230 (nil: 0, 0, 1)
	size       int
	Vertices   []*Vertex
	endhandle  int
	endSrc     int // handle of SEQEND in the file read from
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
	if p.Elevation != 0 {
		f.WriteFloat(30, p.Elevation)
	} else {
		f.WriteString(30, "0.0")
	}
	writeThickness(f, p.Thickness)
	f.WriteInt(70, p.Flag)
	if p.StartWidth != 0 {
		f.WriteFloat(40, p.StartWidth)
	}
	if p.EndWidth != 0 {
		f.WriteFloat(41, p.EndWidth)
	}
	for i, v := range []int{p.MeshM, p.MeshN, p.SmoothM, p.SmoothN, p.CurveType} {
		if v != 0 {
			f.WriteInt(71+i, v)
		}
	}
	writeExtrusion(f, p.Extrusion)
	p.writeXData(f)
	for _, v := range p.Vertices {
		v.Format(f)
		v.writeXData(f)
	}
	f.WriteString(0, "SEQEND")
	f.WriteHex(5, p.endhandle)
	f.WriteHex(330, p.Handle()) // owner: POLYLINE
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
	p.endhandle = handle.Next(p.endSrc, h)
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
	return extrude(mins, maxs, p.Thickness)
}

// Bulge computes the center and the radius of the arc from the vertex at index
// to the next vertex.
func (p Polyline) Bulge(index int) ([]float64, float64) {
	if index > len(p.Vertices)-2 {
		return nil, 0
	}

	return bulge(p.Vertices[index].Bulge, p.Vertices[index].Coord, p.Vertices[index+1].Coord)
}

// SetEndSourceHandle sets the handle of SEQEND in the file it was read from.
func (p *Polyline) SetEndSourceHandle(h string) {
	p.endSrc = handle.Parse(h)
}
