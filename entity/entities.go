package entity

import (
	"fmt"
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
	tmpdata := make([][2]string, 0)
	for i, d := range data {
		if d[0] == "0" {
			if len(tmpdata) > 0 {
				e, err := Parse(tmpdata)
				if err != nil {
					return fmt.Errorf("line %d: %s", line + 2*i, err.Error())
				}
				es.Add(e)
				tmpdata = make([][2]string, 0)
			}
		}
		tmpdata = append(tmpdata, d)
	}
	if len(tmpdata) > 0 {
		e, err := Parse(tmpdata)
		if err != nil {
			return fmt.Errorf("line %d: %s", line + 2*len(data), err.Error())
		}
		es.Add(e)
		tmpdata = make([][2]string, 0)
	}
	return nil
}
