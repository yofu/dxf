package object

import (
	"bytes"
)

type Objects struct {
}

func New() *Objects {
	o := new(Objects)
	return o
}

func (o *Objects) WriteTo(b *bytes.Buffer) error {
	b.WriteString("0\nSECTION\n2\nOBJECTS\n")
	b.WriteString("0\nENDSEC\n")
	return nil
}

func (os *Objects) SetHandle(v *int) {
	return
}
