// CLASSES section
package class

import (
	"strconv"

	"github.com/yofu/dxf/format"
)

// Class represents each CLASS. Its group codes are kept as read.
type Class struct {
	Data [][2]string // group codes after "0 CLASS"
}

// Name returns the DXF record name of the class (code 1).
func (c *Class) Name() string {
	for _, p := range c.Data {
		if p[0] == "1" {
			return p[1]
		}
	}
	return ""
}

// Format writes data to formatter.
func (c *Class) Format(f format.Formatter) {
	f.WriteString(0, "CLASS")
	for _, p := range c.Data {
		if code, err := strconv.Atoi(p[0]); err == nil {
			f.WriteString(code, p[1])
		}
	}
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

// Format writes CLASSES data to formatter.
func (cs Classes) Format(f format.Formatter) {
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
