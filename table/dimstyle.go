package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// DimStyle represents DIMSTYLE SymbolTable.
type DimStyle struct {
	handle int
	owner  handle.Handler
	name   string

	Scale                 float64 // 40 DIMSCALE: overall scale factor
	ArrowSize             float64 // 41 DIMASZ
	ExtOffset             float64 // 42 DIMEXO: offset of extension lines from origin points
	ExtExtend             float64 // 44 DIMEXE: extension beyond the dimension line
	TextInsideHorizontal  bool    // 73 DIMTIH
	TextOutsideHorizontal bool    // 74 DIMTOH
	TextAbove             bool    // 77 DIMTAD: text above the dimension line
	TextHeight            float64 // 140 DIMTXT
	TextGap               float64 // 147 DIMGAP: gap between text and dimension line
	Decimals              int     // 271 DIMDEC: decimal places of measurement
	// Extra holds group codes read from a file that are not interpreted by
	// this package (and XDATA); they are written after the other values.
	Extra [][2]string
}

// NewDimStyle creates a new DimStyle with metric (ISO-25 like) values.
func NewDimStyle(name string) *DimStyle {
	return &DimStyle{
		name:       name,
		Scale:      1.0,
		ArrowSize:  2.5,
		ExtOffset:  0.625,
		ExtExtend:  1.25,
		TextAbove:  true,
		TextHeight: 2.5,
		TextGap:    0.625,
		Decimals:   2,
	}
}

// IsSymbolTable is for SymbolTable interface.
func (d *DimStyle) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (d *DimStyle) Format(f format.Formatter) {
	f.WriteString(0, "DIMSTYLE")
	f.WriteHex(105, d.handle)
	if d.owner != nil {
		f.WriteHex(330, d.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbDimStyleTableRecord")
	f.WriteString(2, d.name)
	f.WriteInt(70, 0)
	f.WriteFloat(40, d.Scale)
	f.WriteFloat(41, d.ArrowSize)
	f.WriteFloat(42, d.ExtOffset)
	f.WriteFloat(44, d.ExtExtend)
	f.WriteInt(73, boolInt(d.TextInsideHorizontal))
	f.WriteInt(74, boolInt(d.TextOutsideHorizontal))
	f.WriteInt(77, boolInt(d.TextAbove))
	f.WriteFloat(140, d.TextHeight)
	f.WriteFloat(147, d.TextGap)
	f.WriteInt(271, d.Decimals)
	writeRaw(f, d.Extra)
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// String outputs data using default formatter.
func (d *DimStyle) String() string {
	f := format.NewASCII()
	return d.FormatString(f)
}

// FormatString outputs data using given formatter.
func (d *DimStyle) FormatString(f format.Formatter) string {
	d.Format(f)
	return f.Output()
}

// Handle returns a handle value.
func (d *DimStyle) Handle() int {
	return d.handle
}

// SetHandle sets a handle.
func (d *DimStyle) SetHandle(v *int) {
	d.handle = *v
	(*v)++
}

// SetOwner sets an owner.
func (d *DimStyle) SetOwner(h handle.Handler) {
	d.owner = h
}

// Name returns a name of DIMSTYLE (code 2).
func (d *DimStyle) Name() string {
	return d.name
}

// SetName sets the name of the dimension style.
func (d *DimStyle) SetName(name string) {
	d.name = name
}
