package object

import (
	"github.com/yofu/dxf/format"
)

type Object interface {
	IsObject() bool
	Format(f *format.Formatter)
	Handle() int
	SetHandle(*int)
}

type Objects []Object

func New() Objects {
	o := make([]Object, 0)
	return o
}

func (os Objects) WriteTo(f *format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "OBJECTS")
	for _, o := range os {
		o.Format(f)
	}
	f.WriteString(0, "ENDSEC")
}

func (os Objects) Add(o Object) Objects {
	os = append(os, o)
	return os
}

func (os Objects) SetHandle(v *int) {
	for _, o := range os {
		o.SetHandle(v)
	}
}

func (os Objects) Read(line int, data [][2]string) error {
	return nil
}
