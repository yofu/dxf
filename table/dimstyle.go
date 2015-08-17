package table

import (
	"github.com/yofu/dxf/format"
)

type DimStyle struct {
	handle   int
	Name     string
}

func NewDimStyle(name string) *DimStyle {
	d := new(DimStyle)
	d.Name = name
	return d
}

func (d *DimStyle) IsSymbolTable() bool {
	return true
}

func (d *DimStyle) Format(f *format.Formatter) {
	f.WriteString(0, "DIMSTYLE")
	f.WriteHex(105, d.handle)
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbDimStyleTableRecord")
	f.WriteString(2, d.Name)
	f.WriteInt(70, 0)
}

func (d *DimStyle) String() string {
	f := format.New()
	return d.FormatString(f)
}

func (d *DimStyle) FormatString(f *format.Formatter) string {
	d.Format(f)
	return f.Output()
}

func (d *DimStyle) Handle() int {
	return d.handle
}
func (d *DimStyle) SetHandle(v *int) {
	d.handle = *v
	(*v)++
}
