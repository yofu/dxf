package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Default Styles.
var (
	ST_STANDARD = NewStyle("Standard")
)

// Style represents STYLE SymbolTable.
type Style struct {
	handle      int
	owner       handle.Handler
	name        string // 2
	FontName    string // 3
	BigFontName string // 4
}

// NewStyle create a new Style.
func NewStyle(name string) *Style {
	st := new(Style)
	st.name = name
	st.FontName = "arial.ttf"
	return st
}

// IsSymbolTable is for SymbolTable interface.
func (st *Style) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (st *Style) Format(f *format.Formatter) {
	f.WriteString(0, "STYLE")
	f.WriteHex(5, st.handle)
	if st.owner != nil {
		f.WriteHex(330, st.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbTextStyleTableRecord")
	f.WriteString(2, st.name)
	f.WriteInt(70, 0)
	f.WriteString(40, "0.0")
	f.WriteString(41, "1.0")
	f.WriteString(50, "0.0")
	f.WriteInt(71, 0)
	f.WriteString(42, "0.2")
	f.WriteString(3, st.FontName)
	f.WriteString(4, st.BigFontName)
}

// String outputs data using default formatter.
func (st *Style) String() string {
	f := format.New()
	return st.FormatString(f)
}

// FormatString outputs data using given formatter.
func (st *Style) FormatString(f *format.Formatter) string {
	st.Format(f)
	return f.Output()
}

// Handle returns a handle value.
func (st *Style) Handle() int {
	return st.handle
}

// SetHandle sets a handle.
func (st *Style) SetHandle(v *int) {
	st.handle = *v
	(*v)++
}

// SetOwner sets an owner.
func (st *Style) SetOwner(h handle.Handler) {
	st.owner = h
}

// Name returns a name of STYLE (code 2).
func (st *Style) Name() string {
	return st.name
}
