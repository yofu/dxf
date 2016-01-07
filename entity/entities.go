// ENTITIES section
package entity

import (
	"github.com/yofu/dxf/format"
)

// Entities represents ENTITIES section.
type Entities []Entity

// New creates a new Entities.
func New() Entities {
	e := make([]Entity, 0)
	return e
}

// WriteTo writes ENTITIES data to formatter.
func (es Entities) WriteTo(f format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "ENTITIES")
	for _, e := range es {
		e.Format(f)
	}
	f.WriteString(0, "ENDSEC")
}

// Add adds a new entity to ENTITIES section.
func (es Entities) Add(e Entity) Entities {
	es = append(es, e)
	return es
}

// SetHandle sets handles to each entity.
func (es Entities) SetHandle(v *int) {
	for _, e := range es {
		e.SetHandle(v)
	}
}
