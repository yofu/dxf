package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Default Styles. ST_STANDARD is a template: each Drawing has its own
// "Standard" style in its STYLE table (Drawing.Style("Standard", ...)).
var (
	ST_STANDARD = NewStyle("Standard")
)

// Style represents STYLE SymbolTable.
type Style struct {
	handle          int
	src             int // handle in the file it was read from (kept when writing)
	owner           handle.Handler
	name            string  // 2
	FontName        string  // 3
	BigFontName     string  // 4
	FixedTextHeight float64 // 40
	WidthFactor     float64 // 41
	LastHeightUsed  float64 // 42
	ObliqueAngle    float64 // 50
	Flags           int     // 70: 1 shape file, 4 vertical text
	GenFlag         int     // 71: text generation flags (2 backward, 4 upside down)
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
	f.WriteInt(70, st.Flags)
	f.WriteFloat(40, st.FixedTextHeight)
	f.WriteFloat(41, st.WidthFactor)
	f.WriteFloat(50, st.ObliqueAngle)
	f.WriteInt(71, st.GenFlag)
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
	st.handle = handle.Next(st.src, v)
}

// SetOwner sets an owner.
func (st *Style) SetOwner(h handle.Handler) {
	st.owner = h
}

// Name returns a name of STYLE (code 2).
func (st *Style) Name() string {
	return st.name
}

// clone returns a copy of st without handle and owner.
func (st *Style) clone() *Style {
	c := *st
	c.handle = 0
	c.owner = nil
	return &c
}

// SetName sets the name of the style.
func (st *Style) SetName(name string) {
	st.name = name
}

// SourceHandle returns the handle in the file it was read from ("" if new).
func (st *Style) SourceHandle() string {
	return handle.Format(st.src)
}

// SetSourceHandle sets the handle in the file it was read from;
// it is kept when the drawing is written.
func (st *Style) SetSourceHandle(h string) {
	st.src = handle.Parse(h)
}
