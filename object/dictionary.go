package object

import (
	"errors"
	"fmt"
	"github.com/yofu/dxf/format"
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

func (d *Dictionary) Format(f *format.Formatter) {
	f.WriteString(0, "DICTIONARY")
	f.WriteHex(5, d.handle)
	f.WriteString(100, "AcDbDictionary")
	f.WriteInt(281, 1)
	for k, v := range d.item {
		f.WriteString(3, k)
		f.WriteHex(350, v.Handle())
	}
}

func (d *Dictionary) String() string {
	f := format.New()
	return d.FormatString(f)
}

func (d *Dictionary) FormatString(f *format.Formatter) string {
	d.Format(f)
	return f.Output()
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
