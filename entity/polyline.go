package entity

import (
	"bytes"
	"fmt"
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
		NewEntity(POLYLINE),
		8,
		0,
		vs,
		0,
	}
	return p
}

func (p *Polyline) String() string {
	var otp bytes.Buffer
	otp.WriteString(p.entity.String())
	otp.WriteString("100\nAcDb3dPolyline\n")
	otp.WriteString("66\n1\n10\n0.0\n20\n0.0\n30\n0.0\n")
	otp.WriteString(fmt.Sprintf("70\n%d\n", p.Flag))
	for _, v := range p.Vertices {
		otp.WriteString(v.String())
	}
	otp.WriteString(fmt.Sprintf("0\nSEQEND\n5\n%X\n100\nAcDbEntity\n8\n%s\n", p.endhandle, p.Layer().Name))
	return otp.String()
}

func (p *Polyline) Close() {
	p.Flag |= 1
}

func (p *Polyline) AddVertex(x, y, z float64) *Vertex {
	v := NewVertex(x, y, z)
	p.Vertices = append(p.Vertices, v)
	p.size++
	v.SetLayer(p.Layer())
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
