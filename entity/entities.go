package entity

import (
	"github.com/yofu/dxf/format"
)

type Entities []Entity

func New() Entities {
	e := make([]Entity, 0)
	return e
}

func (es Entities) WriteTo(f *format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "ENTITIES")
	for _, e := range es {
		e.Format(f)
	}
	f.WriteString(0, "ENDSEC")
}

func (es Entities) Add(e Entity) Entities {
	es = append(es, e)
	return es
}

func (es Entities) SetHandle(v *int) {
	for _, e := range es {
		e.SetHandle(v)
	}
}

func (es Entities) Read(line int, data [][2]string) error {
	return nil
}
