package entity

import (
	"github.com/yofu/dxf/format"
)

type Polyline struct {
	*entity
	Flag      int
	size      int
	Vertices  []*Vertex
	endhandle int
}

func (p *Polyline) IsEntity() bool {
	return true
}

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

func (p *Polyline) Format(f *format.Formatter) {
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
	f.WriteString(8, p.Layer().Name)
}

func (p *Polyline) String() string {
	f := format.New()
	return p.FormatString(f)
}

func (p *Polyline) FormatString(f *format.Formatter) string {
	p.Format(f)
	return f.Output()
}

func (p *Polyline) Close() {
	p.Flag |= 1
}

func (p *Polyline) AddVertex(x, y, z float64) *Vertex {
	v := NewVertex(x, y, z)
	p.Vertices = append(p.Vertices, v)
	p.size++
	v.SetLayer(p.Layer())
	v.SetOwner(p)
	return v
}

func (p *Polyline) SetHandle(h *int) {
	p.entity.SetHandle(h)
	for _, v := range p.Vertices {
		v.SetHandle(h)
	}
	p.endhandle = *h
	(*h)++
}
