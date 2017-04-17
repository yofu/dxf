package table

import (
	"fmt"
	"strings"

	"github.com/yofu/dxf/format"
)

// Table represents each TABLE.
type Table struct {
	name   string // 2
	handle int    // 5
	size   int    // 70
	tables []SymbolTable
}

// NewTable creates a new Table.
func NewTable(name string) *Table {
	t := new(Table)
	t.name = name
	return t
}

// Format writes data to formatter.
func (t *Table) Format(f format.Formatter) {
	f.WriteString(0, "TABLE")
	f.WriteString(2, t.name)
	f.WriteHex(5, t.handle)
	f.WriteString(100, "AcDbSymbolTable")
	f.WriteInt(70, t.size)
	if t.name == "DIMSTYLE" {
		f.WriteString(100, "AcDbDimStyleTable")
		f.WriteInt(71, t.size)
		for i := 0; i < t.size; i++ {
			f.WriteHex(340, t.tables[i].Handle())
		}
	}
	for i := 0; i < t.size; i++ {
		t.tables[i].Format(f)
	}
	f.WriteString(0, "ENDTAB")
}

// String outputs data using default formatter.
func (t *Table) String() string {
	f := format.NewASCII()
	return t.FormatString(f)
}

// FormatString outputs data using given formatter.
func (t *Table) FormatString(f format.Formatter) string {
	t.Format(f)
	return f.Output()
}

// Handle returns a handle value of TABLE.
func (t *Table) Handle() int {
	return t.handle
}

// SetHandle sets handles to TABLE itself and each SymbolTable.
func (t *Table) SetHandle(v *int) {
	t.handle = *v
	(*v)++
	for i := 0; i < t.size; i++ {
		t.tables[i].SetHandle(v)
	}
}

// Add adds a new SymbolTable to TABLE.
func (t *Table) Add(st SymbolTable) {
	t.tables = append(t.tables, st)
	st.SetOwner(t)
	t.size++
}

// Clear deletes existing SymbolTables.
func (t *Table) Clear() {
	t.tables = make([]SymbolTable, 0)
	t.size = 0
}

// Contains reports if TABLE has the named SymbolTable.
func (t *Table) Contains(name string) (SymbolTable, error) {
	for _, st := range t.tables {
		if strings.EqualFold(st.Name(), name) {
			return st, nil
		}
	}
	return nil, fmt.Errorf("%s doesn't exist", name)
}
