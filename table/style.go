package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

var (
	ST_STANDARD = NewStyle("Standard")
)

type Style struct {
	handle      int
	owner       handle.Handler
	Name        string // 2
	FontName    string // 3
	BigFontName string // 4
}

func NewStyle(name string) *Style {
	st := new(Style)
	st.Name = name
	st.FontName = "arial.ttf"
	return st
}

func (st *Style) IsSymbolTable() bool {
	return true
}

func (st *Style) Format(f *format.Formatter) {
	f.WriteString(0, "STYLE")
	f.WriteHex(5, st.handle)
	if st.owner != nil {
		f.WriteHex(330, st.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbTextStyleTableRecord")
	f.WriteString(2, st.Name)
	f.WriteInt(70, 0)
	f.WriteString(40, "0.0")
	f.WriteString(41, "1.0")
	f.WriteString(50, "0.0")
	f.WriteInt(71, 0)
	f.WriteString(42, "0.2")
	f.WriteString(3, st.FontName)
	f.WriteString(4, st.BigFontName)
}

func (st *Style) String() string {
	f := format.New()
	return st.FormatString(f)
}

func (st *Style) FormatString(f *format.Formatter) string {
	st.Format(f)
	return f.Output()
}

func (st *Style) Handle() int {
	return st.handle
}
func (st *Style) SetHandle(v *int) {
	st.handle = *v
	(*v)++
}

func (st *Style) SetOwner(h handle.Handler) {
	st.owner = h
}
