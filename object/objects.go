package object

import (
	"bytes"
)

type Object struct {
}

type Objects []Object

func New() Objects {
	o := make([]Object, 0)
	return o
}

func (o Objects) WriteTo(b *bytes.Buffer) error {
	b.WriteString("0\nSECTION\n2\nOBJECTS\n")
	b.WriteString("0\nENDSEC\n")
	return nil
}

func (os Objects) SetHandle(v *int) {
	return
}
