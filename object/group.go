package object

import (
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Group represents GROUP Object.
type Group struct {
	Name        string
	Description string
	handle      int
	owner       handle.Handler
	entities    []entity.Entity
	selectable  bool
	Unnamed     bool // 70: unnamed (anonymous) group, named *An
}

// IsObject is for Object interface.
func (g *Group) IsObject() bool {
	return true
}

// NewGroup creates a new Group.
func NewGroup(name, desc string, es ...entity.Entity) *Group {
	g := &Group{
		Name:        name,
		Description: desc,
		handle:      0,
		owner:       nil,
		entities:    es,
		selectable:  true,
	}
	for _, e := range es {
		e.SetBlockRecord(g)
	}
	return g
}

// SetOwner sets an owner(Dictionary).
func (g *Group) SetOwner(d *Dictionary) {
	g.owner = d
	d.AddItem(g.Name, g)
}

// Format writes data to formatter.
func (g *Group) Format(f format.Formatter) {
	f.WriteString(0, "GROUP")
	f.WriteHex(5, g.handle)
	f.WriteString(102, "{ACAD_REACTORS")
	f.WriteHex(330, g.owner.Handle())
	f.WriteString(102, "}")
	f.WriteHex(330, g.owner.Handle())
	f.WriteString(100, "AcDbGroup")
	f.WriteString(300, g.Description)
	if g.Unnamed {
		f.WriteInt(70, 1)
	} else {
		f.WriteInt(70, 0)
	}
	if g.selectable {
		f.WriteInt(71, 1)
	} else {
		f.WriteInt(71, 0)
	}
	for _, e := range g.entities {
		f.WriteHex(340, e.Handle())
	}
}

// String outputs data using default formatter.
func (g *Group) String() string {
	f := format.NewASCII()
	return g.FormatString(f)
}

// FormatString outputs data using given formatter.
func (g *Group) FormatString(f format.Formatter) string {
	g.Format(f)
	return f.Output()
}

// Handle returns a handle value.
func (g *Group) Handle() int {
	return g.handle
}

// SetHandle sets a handle.
func (g *Group) SetHandle(v *int) {
	g.handle = *v
	(*v)++
}

// AddEntity adds entities to Group.
func (g *Group) AddEntity(es ...entity.Entity) {
	for _, e := range es {
		e.SetBlockRecord(g)
	}
	g.entities = append(g.entities, es...)
}

// Selectable reports whether the group is selectable (code 71).
func (g *Group) Selectable() bool {
	return g.selectable
}

// SetSelectable sets whether the group is selectable (code 71).
func (g *Group) SetSelectable(s bool) {
	g.selectable = s
}

// Entities returns the entities of the group.
func (g *Group) Entities() []entity.Entity {
	return g.entities
}

// RemoveEntity removes entities from the group.
func (g *Group) RemoveEntity(es ...entity.Entity) {
	kept := make([]entity.Entity, 0, len(g.entities))
	for _, e := range g.entities {
		remove := false
		for _, r := range es {
			if e == r {
				remove = true
				break
			}
		}
		if !remove {
			kept = append(kept, e)
		}
	}
	g.entities = kept
}
