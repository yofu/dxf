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

type Classes []*Class

func New() Classes {
	c := make([]*Class, 0)
	return c
}

func (cs Classes) WriteTo(b *bytes.Buffer) error {
	b.WriteString("0\nSECTION\n2\nCLASSES\n")
	for _, c := range cs {
		b.WriteString(c.String())
	}
	b.WriteString("0\nENDSEC\n")
	return nil
}

func (cs Classes) SetHandle(v *int) {
	return
}
