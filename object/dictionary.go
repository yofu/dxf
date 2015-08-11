package object

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/yofu/dxf/handle"
)

type Dictionary struct {
	handle int
	item   map[string]handle.Handler
}

func (d *Dictionary) IsObject() bool {
	return true
}

func NewDictionary() *Dictionary {
	ds := make(map[string]handle.Handler)
	d := &Dictionary{
		handle: 0,
		item:   ds,
	}
	return d
}

func (d *Dictionary) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nDICTIONARY\n")
	otp.WriteString(fmt.Sprintf("5\n%X\n", d.handle))
	otp.WriteString("100\nAcDbDictionary\n")
	otp.WriteString("281\n1\n")
	for k, v := range d.item {
		otp.WriteString(fmt.Sprintf("3\n%s\n", k))
		otp.WriteString(fmt.Sprintf("350\n%X\n", v.Handle()))
	}
	return otp.String()
}

func (d *Dictionary) Handle() int {
	return d.handle
}
func (d *Dictionary) SetHandle(v *int) {
	d.handle = *v
	(*v)++
}

func (d *Dictionary) AddItem(key string, value handle.Handler) error {
	if _, exist := d.item[key]; exist {
		return errors.New(fmt.Sprintf("key %s already exists"))
	}
	d.item[key] = value
	return nil
}
