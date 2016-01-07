// OBJECTS section
package object

import (
	"github.com/yofu/dxf/format"
)

// Object is interface for OBJECT.
type Object interface {
	IsObject() bool
	Format(f format.Formatter)
	Handle() int
	SetHandle(*int)
}

// Objects represents OBJECTS section.
type Objects []Object

// New create a new Objects.
func New() Objects {
	o := make([]Object, 0)
	return o
}

// WriteTo writes OBJECTS data to formatter.
func (os Objects) WriteTo(f format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "OBJECTS")
	for _, o := range os {
		o.Format(f)
	}
	f.WriteString(0, "ENDSEC")
}

// Add adds a new object to OBJECTS section.
func (os Objects) Add(o Object) Objects {
	os = append(os, o)
	return os
}

// SetHandle sets handles to each object.
func (os Objects) SetHandle(v *int) {
	for _, o := range os {
		o.SetHandle(v)
	}
}
