package object

import (
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
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

func (g *Group) Format(f *format.Formatter) {
	f.WriteString(0, "GROUP")
	f.WriteHex(5, g.handle)
	f.WriteString(102, "{ACAD_REACTORS")
	f.WriteHex(330, g.owner.Handle())
	f.WriteString(102, "}")
	f.WriteHex(330, g.owner.Handle())
	f.WriteString(100, "AcDbGroup")
	f.WriteString(300, g.Description)
	f.WriteInt(70, 0)
	if g.selectable {
		f.WriteInt(71, 1)
	} else {
		f.WriteInt(71, 0)
	}
	for _, e := range g.entities {
		f.WriteHex(340, e.Handle())
	}
}

func (g *Group) String() string {
	f := format.New()
	return g.FormatString(f)
}

func (g *Group) FormatString(f *format.Formatter) string {
	g.Format(f)
	return f.Output()
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
