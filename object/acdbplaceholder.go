package object

import (
	"bytes"
	"fmt"
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

func (p *AcDbPlaceHolder) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nACDBPLACEHOLDER\n")
	otp.WriteString(fmt.Sprintf("5\n%x\n", p.handle))
	if p.owner != nil {
		otp.WriteString(fmt.Sprintf("102\n{ACAD_REACTORS\n330\n%x\n102\n}\n", p.owner.Handle()))
		otp.WriteString(fmt.Sprintf("330\n%x\n", p.owner.Handle()))
	}
	return otp.String()
}

func (p *AcDbPlaceHolder) Handle() int {
	return p.handle
}
func (p *AcDbPlaceHolder) SetHandle(v *int) {
	p.handle = *v
	(*v)++
}
