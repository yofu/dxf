package object

import (
	"bytes"
)

type Object interface {
	IsObject() bool
	String() string
	Handle() int
	SetHandle(*int)
}

type Objects []Object

func New() Objects {
	o := make([]Object, 0)
	return o
}

func (os Objects) WriteTo(b *bytes.Buffer) error {
	b.WriteString("0\nSECTION\n2\nOBJECTS\n")
	for _, o := range os {
		b.WriteString(o.String())
	}
	b.WriteString("0\nENDSEC\n")
	return nil
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
