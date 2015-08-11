package object

import (
	"bytes"
	"fmt"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/handle"
)

type Group struct {
	Name        string
	Description string
	handle      int
	owner       handle.Handler
	entities    []entity.Entity
	selectable  bool
}

func (g *Group) IsObject() bool {
	return true
}

func NewGroup(name, desc string, es ...entity.Entity) *Group {
	g := &Group{
		Name:        name,
		Description: desc,
		handle:      0,
		owner:       nil,
		entities:    es,
		selectable:  true,
	}
	return g
}

func (g *Group) SetOwner(d *Dictionary) {
	g.owner = d
	d.AddItem(g.Name, g)
}

func (g *Group) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nGROUP\n")
	otp.WriteString(fmt.Sprintf("5\n%X\n", g.handle))
	otp.WriteString(fmt.Sprintf("102\n{ACAD_REACTORS\n330\n%X\n102\n}\n", g.owner.Handle()))
	otp.WriteString(fmt.Sprintf("330\n%X\n", g.owner.Handle()))
	otp.WriteString("100\nAcDbGroup\n")
	otp.WriteString(fmt.Sprintf("300\n%s\n", g.Description))
	otp.WriteString("70\n0\n")
	if g.selectable {
		otp.WriteString("71\n1\n")
	} else {
		otp.WriteString("71\n0\n")
	}
	for _, e := range g.entities {
		otp.WriteString(fmt.Sprintf("340\n%X\n", e.Handle()))
	}
	return otp.String()
}

func (g *Group) Handle() int {
	return g.handle
}
func (g *Group) SetHandle(v *int) {
	g.handle = *v
	(*v)++
}

func (g *Group) AddEntity(es ...entity.Entity) {
	for _, e := range es {
		e.SetBlockRecord(g)
	}
	g.entities = append(g.entities, es...)
}
