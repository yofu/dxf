package class

import (
	"bytes"
)

type Class struct {
}

func (c *Class) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nCLASS\n")
	return otp.String()
}

type Classes struct {
	values []*Class
}

func New() *Classes {
	c := new(Classes)
	c.values = make([]*Class, 0)
	return c
}

func (cs *Classes) WriteTo(b *bytes.Buffer) error {
	b.WriteString("0\nSECTION\n2\nCLASSES\n")
	for _, c := range cs.values {
		b.WriteString(c.String())
	}
	b.WriteString("0\nENDSEC\n")
	return nil
}

func (cs *Classes) SetHandle(v *int) {
	return
}
