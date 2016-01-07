// CLASSES section
package class

import (
	"github.com/yofu/dxf/format"
)

// Class represents each CLASS.
type Class struct {
}

// Format writes data to formatter.
func (c *Class) Format(f format.Formatter) {
	f.WriteString(0, "CLASS")
}

// String outputs data using default formatter.
func (c *Class) String() string {
	f := format.NewASCII()
	return c.FormatString(f)
}

// FormatString outputs data using given formatter.
func (c *Class) FormatString(f format.Formatter) string {
	c.Format(f)
	return f.Output()
}

// Classes represents CLASSES section.
type Classes []*Class

// New creates a new Classes.
func New() Classes {
	c := make([]*Class, 0)
	return c
}

// WriteTo writes CLASSES data to formatter.
func (cs Classes) WriteTo(f format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "CLASSES")
	for _, c := range cs {
		c.Format(f)
	}
	f.WriteString(0, "ENDSEC")
}

// SetHandle sets handles to each class.
func (cs Classes) SetHandle(v *int) {
	return
}
