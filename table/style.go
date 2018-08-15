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
	handle          int
	owner           handle.Handler
	name            string  // 2
	FontName        string  // 3
	BigFontName     string  // 4
	FixedTextHeight float64 // 40
	WidthFactor     float64 // 41
	LastHeightUsed  float64 // 42
	ObliqueAngle    float64 // 50
}

// NewStyle create a new Style.
func NewStyle(name string) *Style {
	return &Style{
		name:            name,
		FontName:        "arial.ttf",
		FixedTextHeight: 0.0,
		WidthFactor:     1.0,
		LastHeightUsed:  100.0,
		ObliqueAngle:    0.0,
	}
}

// IsSymbolTable is for SymbolTable interface.
func (st *Style) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (st *Style) Format(f format.Formatter) {
	f.WriteString(0, "STYLE")
	f.WriteHex(5, st.handle)
	if st.owner != nil {
		f.WriteHex(330, st.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbTextStyleTableRecord")
	f.WriteString(2, st.name)
	f.WriteInt(70, 0)
	f.WriteFloat(40, st.FixedTextHeight)
	f.WriteFloat(41, st.WidthFactor)
	f.WriteFloat(50, st.ObliqueAngle)
	f.WriteInt(71, 0)
	f.WriteFloat(42, st.LastHeightUsed)
	f.WriteString(3, st.FontName)
	f.WriteString(4, st.BigFontName)
}

// String outputs data using default formatter.
func (st *Style) String() string {
	f := format.NewASCII()
	return st.FormatString(f)
}

// FormatString outputs data using given formatter.
func (st *Style) FormatString(f format.Formatter) string {
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
