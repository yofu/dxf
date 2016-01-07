// TABLES section
package table

import (
	"github.com/yofu/dxf/format"
)

// Tables represents TABLES section.
type Tables []*Table

// New creates a new Tables.
func New() Tables {
	t := make([]*Table, 9)
	t[0] = NewTable("VPORT")
	t[1] = NewTable("LTYPE")
	t[1].Add(LT_BYLAYER)
	t[1].Add(LT_BYBLOCK)
	t[1].Add(LT_CONTINUOUS)
	t[1].Add(LT_HIDDEN)
	t[1].Add(LT_DASHDOT)
	t[2] = NewTable("LAYER")
	t[2].Add(LY_0)
	t[3] = NewTable("STYLE")
	t[3].Add(ST_STANDARD)
	t[4] = NewTable("VIEW")
	t[5] = NewTable("UCS")
	t[6] = NewTable("APPID")
	t[6].Add(NewAppID("ACAD"))
	t[7] = NewTable("DIMSTYLE")
	t[8] = NewTable("BLOCK_RECORD")
	t[8].Add(NewBlockRecord("*Model_Space"))
	t[8].Add(NewBlockRecord("*Paper_Space"))
	t[8].Add(NewBlockRecord("*Paper_Space0"))
	return t
}

// WriteTo writes TABLES data to formatter.
func (ts Tables) WriteTo(f format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "TABLES")
	for _, t := range ts {
		t.Format(f)
	}
	f.WriteString(0, "ENDSEC")
}

// Add adds a new table to TABLES section.
func (ts Tables) Add(t *Table) Tables {
	ts = append(ts, t)
	return ts
}

// SetHandle sets handles to each table.
func (ts Tables) SetHandle(h *int) {
	for _, t := range ts {
		t.SetHandle(h)
	}
}

// AddLayer adds a new layer to LAYER table.
func (ts Tables) AddLayer(l *Layer) {
	ts[2].Add(l)
}
