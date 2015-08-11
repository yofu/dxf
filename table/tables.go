package table

import (
	"bytes"
)

type Tables []*Table

func New() Tables {
	t := make([]*Table, 4)
	t[0] = NewTable("LTYPE")
	t[0].Add(LT_BYLAYER)
	t[0].Add(LT_BYBLOCK)
	t[0].Add(LT_CONTINUOUS)
	t[1] = NewTable("LAYER")
	t[1].Add(LY_0)
	t[2] = NewTable("STYLE")
	t[2].Add(ST_STANDARD)
	t[3] = NewTable("VIEW")
	return t
}

func (ts Tables) WriteTo(b *bytes.Buffer) error {
	b.WriteString("0\nSECTION\n2\nTABLES\n")
	for _, t := range ts {
		b.WriteString(t.String())
	}
	b.WriteString("0\nENDSEC\n")
	return nil
}

func (ts Tables) Add(t *Table) Tables {
	ts = append(ts, t)
	return ts
}

func (ts Tables) SetHandle(h *int) {
	for _, t := range ts {
		t.SetHandle(h)
	}
}

func (ts Tables) AddLayer(l *Layer) {
	ts[1].Add(l)
}
