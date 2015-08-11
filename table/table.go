package table

import (
	"bytes"
	"fmt"
)

type Table struct {
	name   string // 2
	handle int    // 5
	size   int    // 70
	tables []SymbolTable
}

func NewTable(name string) *Table {
	t := new(Table)
	t.name = name
	return t
}

func (t *Table) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nTABLE\n")
	otp.WriteString(fmt.Sprintf("2\n%s\n", t.name))
	otp.WriteString(fmt.Sprintf("5\n%X\n", t.handle))
	otp.WriteString("100\nAcDbSymbolTable\n")
	otp.WriteString(fmt.Sprintf("70\n%d\n", t.size))
	if t.name == "DIMSTYLE" {
		otp.WriteString("100\nAcDbDimStyleTable\n")
		otp.WriteString(fmt.Sprintf("71\n%d\n", t.size))
		for i := 0; i < t.size; i++ {
			otp.WriteString(fmt.Sprintf("340\n%X\n", t.tables[i].Handle()))
		}
	}
	for i := 0; i < t.size; i++ {
		otp.WriteString(t.tables[i].String())
	}
	otp.WriteString("0\nENDTAB\n")
	return otp.String()
}

func (t *Table) Handle() int {
	return t.handle
}
func (t *Table) SetHandle(v *int) {
	t.handle = *v
	(*v)++
	for i := 0; i < t.size; i++ {
		t.tables[i].SetHandle(v)
	}
}

func (t *Table) Add(st SymbolTable) {
	t.tables = append(t.tables, st)
	t.size++
}
