package entity

import (
	"bytes"
)

type Entities struct {
	values []Entity
	size   int
}

func New() *Entities {
	e := new(Entities)
	e.values = make([]Entity, 0)
	return e
}

func (es *Entities) WriteTo(b *bytes.Buffer) error {
	b.WriteString("0\nSECTION\n2\nENTITIES\n")
	for _, e := range es.values {
		b.WriteString(e.String())
	}
	b.WriteString("0\nENDSEC\n")
	return nil
}

func (es *Entities) Add(e Entity) error {
	es.values = append(es.values, e)
	es.size++
	return nil
}

func (es *Entities) SetHandle(v *int) {
	for i := 0; i < es.size; i++ {
		es.values[i].SetHandle(v)
	}
}
