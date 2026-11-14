// TABLES section
package table

import (
	"github.com/yofu/dxf/format"
)

// Tables represents TABLES section.
type Tables []*Table

// New creates a new Tables.
func New() Tables {
	// Default entries are created per Tables (i.e. per Drawing) so that
	// handles, owners and plot styles are not shared between drawings.
	t := make([]*Table, 9)
	t[0] = NewTable("VPORT")
	t[0].Add(NewViewport("*Active")) // the active viewport of model space
	t[1] = NewTable("LTYPE")
	var continuous *LineType
	for _, lt := range []*LineType{LT_BYLAYER, LT_BYBLOCK, LT_CONTINUOUS, LT_HIDDEN, LT_DASHDOT} {
		c := lt.clone()
		if lt == LT_CONTINUOUS {
			continuous = c
		}
		t[1].Add(c)
	}
	t[2] = NewTable("LAYER")
	t[2].Add(NewLayer(LY_0.name, LY_0.Color, continuous))
	t[3] = NewTable("STYLE")
	t[3].Add(ST_STANDARD.clone())
	t[4] = NewTable("VIEW")
	t[5] = NewTable("UCS")
	t[6] = NewTable("APPID")
	t[6].Add(NewAppID("ACAD"))
	t[7] = NewTable("DIMSTYLE")
	t[7].Add(NewDimStyle("Standard"))
	t[8] = NewTable("BLOCK_RECORD")
	t[8].Add(NewBlockRecord("*Model_Space"))
	t[8].Add(NewBlockRecord("*Paper_Space"))
	return t
}

// Format writes TABLES data to formatter.
func (ts Tables) Format(f format.Formatter) {
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
