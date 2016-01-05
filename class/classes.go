package class

import (
	"github.com/yofu/dxf/format"
)

type Class struct {
}

func (c *Class) Format(f *format.Formatter) {
	f.WriteString(0, "CLASS")
}

func (c *Class) String() string {
	f := format.New()
	return c.FormatString(f)
}

func (c *Class) FormatString(f *format.Formatter) string {
	c.Format(f)
	return f.Output()
}

type Classes []*Class

func New() Classes {
	c := make([]*Class, 0)
	return c
}

func (cs Classes) WriteTo(f *format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "CLASSES")
	for _, c := range cs {
		c.Format(f)
	}
	f.WriteString(0, "ENDSEC")
}

func (cs Classes) SetHandle(v *int) {
	return
}
