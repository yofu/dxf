package table

import (
	"github.com/yofu/dxf/format"
)

type View struct {
	handle int
	Name   string // 2
}

func NewView(name string) *View {
	v := new(View)
	v.Name = name
	return v
}

func (v *View) IsSymbolTable() bool {
	return true
}

func (v *View) Format(f *format.Formatter) {
	f.WriteString(0, "VIEW")
	f.WriteHex(5, v.handle)
	f.WriteString(100, "AcDbSymbostableRecord")
	f.WriteString(100, "AcDbViewTableRecord")
	f.WriteString(2, v.Name)
}

func (v *View) String() string {
	f := format.New()
	return v.FormatString(f)
}

func (v *View) FormatString(f *format.Formatter) string {
	v.Format(f)
	return f.Output()
}

func (v *View) Handle() int {
	return v.handle
}
func (v *View) SetHandle(h *int) {
	v.handle = *h
	(*h)++
}
