package object

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

type AcDbPlaceHolder struct {
	handle int
	owner  handle.Handler
}

func (p *AcDbPlaceHolder) IsObject() bool {
	return true
}

func NewAcDbPlaceHolder() *AcDbPlaceHolder {
	p := &AcDbPlaceHolder{
		handle: 0,
		owner:  nil,
	}
	return p
}

func (p *AcDbPlaceHolder) Format(f *format.Formatter) {
	f.WriteString(0, "ACDBPLACEHOLDER")
	f.WriteHex(5, p.handle)
	if p.owner != nil {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteHex(330, p.owner.Handle())
		f.WriteString(102, "}")
		f.WriteHex(330, p.owner.Handle())
	}
}

func (p *AcDbPlaceHolder) String() string {
	f := format.New()
	return p.FormatString(f)
}

func (p *AcDbPlaceHolder) FormatString(f *format.Formatter) string {
	p.Format(f)
	return f.Output()
}

func (p *AcDbPlaceHolder) Handle() int {
	return p.handle
}
func (p *AcDbPlaceHolder) SetHandle(v *int) {
	p.handle = *v
	(*v)++
}
