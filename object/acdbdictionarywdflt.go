package object

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/yofu/dxf/handle"
)

type AcDbDictionaryWDFLT struct {
	handle        int
	item          map[string]handle.Handler
	owner         handle.Handler
	defaulthandle handle.Handler
}

func (d *AcDbDictionaryWDFLT) IsObject() bool {
	return true
}

func NewAcDbDictionaryWDFLT(owner handle.Handler) (*AcDbDictionaryWDFLT, *AcDbPlaceHolder) {
	ds := make(map[string]handle.Handler)
	p := NewAcDbPlaceHolder()
	ds["Normal"] = p
	d := &AcDbDictionaryWDFLT{
		handle:        0,
		item:          ds,
		owner:         owner,
		defaulthandle: p,
	}
	p.owner = d
	return d, p
}

func (d *AcDbDictionaryWDFLT) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nACDBDICTIONARYWDFLT\n")
	otp.WriteString(fmt.Sprintf("5\n%x\n", d.handle))
	if d.owner != nil {
		otp.WriteString(fmt.Sprintf("102\n{ACAD_REACTORS\n330\n%x\n102\n}\n", d.owner.Handle()))
		otp.WriteString(fmt.Sprintf("330\n%x\n", d.owner.Handle()))
	}
	otp.WriteString("100\nAcDbDictionary\n")
	otp.WriteString("281\n1\n")
	for k, v := range d.item {
		otp.WriteString(fmt.Sprintf("3\n%s\n", k))
		otp.WriteString(fmt.Sprintf("350\n%x\n", v.Handle()))
	}
	otp.WriteString("100\nAcDbDictionaryWithDefault\n")
	otp.WriteString(fmt.Sprintf("340\n%x\n", d.defaulthandle))
	return otp.String()
}

func (d *AcDbDictionaryWDFLT) Handle() int {
	return d.handle
}
func (d *AcDbDictionaryWDFLT) SetHandle(v *int) {
	d.handle = *v
	(*v)++
}

func (d *AcDbDictionaryWDFLT) AddItem(key string, value handle.Handler) error {
	if _, exist := d.item[key]; exist {
		return errors.New(fmt.Sprintf("key %s already exists"))
	}
	d.item[key] = value
	return nil
}
