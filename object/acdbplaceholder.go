package object

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// AcDbPlaceHolder represents ACDBPLACEHOLDER Object.
type AcDbPlaceHolder struct {
	handle int
	owner  handle.Handler
}

// IsObject is for Object interface.
func (p *AcDbPlaceHolder) IsObject() bool {
	return true
}

// NewAcDbPlaceHolder creates a new AcDbPlaceHolder.
func NewAcDbPlaceHolder() *AcDbPlaceHolder {
	p := &AcDbPlaceHolder{
		handle: 0,
		owner:  nil,
	}
	return p
}

// Format writes data to formatter.
func (p *AcDbPlaceHolder) Format(f format.Formatter) {
	f.WriteString(0, "ACDBPLACEHOLDER")
	f.WriteHex(5, p.handle)
	if p.owner != nil {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteHex(330, p.owner.Handle())
		f.WriteString(102, "}")
		f.WriteHex(330, p.owner.Handle())
	}
}

// String outputs data using default formatter.
func (p *AcDbPlaceHolder) String() string {
	f := format.NewASCII()
	return p.FormatString(f)
}

// FormatString outputs data using given formatter.
func (p *AcDbPlaceHolder) FormatString(f format.Formatter) string {
	p.Format(f)
	return f.Output()
}

// Handle returns a handle value.
func (p *AcDbPlaceHolder) Handle() int {
	return p.handle
}

// SetHandle sets a handle.
func (p *AcDbPlaceHolder) SetHandle(v *int) {
	p.handle = *v
	(*v)++
}
