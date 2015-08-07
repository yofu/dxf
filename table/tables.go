package table

import (
	"bytes"
)

type Tables struct {
	values []*Table
	size int
}

func New() *Tables {
	t := new(Tables)
	t.values = make([]*Table, 4)
	t.size = 4
	t.values[0] = NewTable("LTYPE")
	t.values[0].Add(LT_BYLAYER)
	t.values[0].Add(LT_BYBLOCK)
	t.values[0].Add(LT_CONTINUOUS)
	t.values[1] = NewTable("LAYER")
	t.values[1].Add(LY_0)
	t.values[2] = NewTable("STYLE")
	t.values[2].Add(ST_STANDARD)
	t.values[3] = NewTable("VIEW")
	return t
}

func (ts *Tables) WriteTo(b *bytes.Buffer) error {
	b.WriteString("0\nSECTION\n2\nTABLES\n")
	for i:=0; i<ts.size; i++ {
		b.WriteString(ts.values[i].String())
	}
	b.WriteString("0\nENDSEC\n")
	return nil
}

func (ts *Tables) Add(t *Table) {
	ts.values = append(ts.values, t)
	ts.size++
}

func (ts *Tables) SetHandle(h *int) {
	for _, t := range ts.values {
		t.SetHandle(h)
	}
}
