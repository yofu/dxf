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

// Format writes data to formatter. The instance count (91) is written
// for AC1018 and later only.
func (c *Class) Format(f format.Formatter) {
	f.WriteString(0, "CLASS")
	has91 := false
	for _, p := range c.Data {
		has91 = has91 || p[0] == "91"
	}
	r2004 := format.AtLeast(f, format.R2004)
	for _, p := range c.Data {
		code, err := strconv.Atoi(p[0])
		if err != nil || (code == 91 && !r2004) {
			continue
		}
		f.WriteString(code, p[1])
		if code == 90 && r2004 && !has91 {
			f.WriteInt(91, 0)
		}
	}
}

// NewObjectClass returns the class of an object type defined in
// "ObjectDBX Classes" (e.g. "LAYOUT", "AcDbLayout").
func NewObjectClass(name, cppName string) *Class {
	return &Class{Data: [][2]string{
		{"1", name}, {"2", cppName}, {"3", "ObjectDBX Classes"},
		{"90", "0"}, {"280", "0"}, {"281", "0"},
	}}
}

// Defaults returns the classes of the object types this package writes:
// ACDBDICTIONARYWDFLT, ACDBPLACEHOLDER and LAYOUT.
func Defaults() Classes {
	return Classes{
		NewObjectClass("ACDBDICTIONARYWDFLT", "AcDbDictionaryWithDefault"),
		NewObjectClass("ACDBPLACEHOLDER", "AcDbPlaceHolder"),
		NewObjectClass("LAYOUT", "AcDbLayout"),
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
	return Defaults()
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
