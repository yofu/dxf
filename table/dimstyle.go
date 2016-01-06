package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

type DimStyle struct {
	handle   int
	owner    handle.Handler
	name     string
}

func NewDimStyle(name string) *DimStyle {
	d := new(DimStyle)
	d.name = name
	return d
}

func (d *DimStyle) IsSymbolTable() bool {
	return true
}

func (d *DimStyle) Format(f *format.Formatter) {
	f.WriteString(0, "DIMSTYLE")
	f.WriteHex(105, d.handle)
	if d.owner != nil {
		f.WriteHex(330, d.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbDimStyleTableRecord")
	f.WriteString(2, d.name)
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

func (d *DimStyle) SetOwner(h handle.Handler) {
	d.owner = h
}

func (d *DimStyle) Name() string {
	return d.name
}
