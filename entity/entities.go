package entity

import (
	"bytes"
)

type Entities []Entity

func New() Entities {
	e := make([]Entity, 0)
	return e
}

func (es Entities) WriteTo(b *bytes.Buffer) error {
	b.WriteString("0\nSECTION\n2\nENTITIES\n")
	for _, e := range es {
		b.WriteString(e.String())
	}
	b.WriteString("0\nENDSEC\n")
	return nil
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
