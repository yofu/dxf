package table

import (
	"math"

	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Default LineTypes.
var (
	LT_CONTINUOUS = NewLineType("Continuous", "Solid Line")
	LT_BYLAYER    = NewLineType("ByLayer", "")
	LT_BYBLOCK    = NewLineType("ByBlock", "")
	LT_HIDDEN     = NewLineType("HIDDEN", "Hidden __ __ __ __ __ __ __ __ __ __ __ __ __ _", 0.25, -0.125)
	LT_DASHDOT    = NewLineType("DASHDOT", "Dash dot __ . __ . __ . __ . __ . __ . __ . __", 0.5, -0.25, 0.0, -0.25)
)

// LineType represents LTYPE SymbolTable.
type LineType struct {
	handle      int
	owner       handle.Handler
	name        string // 2
	Description string // 3
	lengths     []float64
}

// NewLineType creates a new LineType.
func NewLineType(name, desc string, ls ...float64) *LineType {
	lt := new(LineType)
	lt.name = name
	lt.Description = desc
	if len(ls) > 0 {
		lt.lengths = ls
	} else {
		lt.lengths = make([]float64, 0)
	}
	return lt
}

// IsSymbolTable is for SymbolTable interface.
func (lt *LineType) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (lt *LineType) Format(f format.Formatter) {
	f.WriteString(0, "LTYPE")
	f.WriteHex(5, lt.handle)
	if lt.owner != nil {
		f.WriteHex(330, lt.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbLinetypeTableRecord")
	f.WriteString(2, lt.name)
	f.WriteInt(70, 0)
	f.WriteString(3, lt.Description)
	f.WriteInt(72, 65)
	f.WriteInt(73, len(lt.lengths))
	f.WriteFloat(40, lt.TotalLength())
	for _, l := range lt.lengths {
		f.WriteFloat(49, l)
		f.WriteInt(74, 0)
	}
}

// String outputs data using default formatter.
func (lt *LineType) String() string {
	f := format.NewASCII()
	return lt.FormatString(f)
}

// FormatString outputs data using given formatter.
func (lt *LineType) FormatString(f format.Formatter) string {
	lt.Format(f)
	return f.Output()
}

// Handle returns a handle value.
func (lt *LineType) Handle() int {
	return lt.handle
}

// SetHandle sets a handle.
func (lt *LineType) SetHandle(v *int) {
	lt.handle = *v
	(*v)++
}

// SetOwner sets an owner.
func (lt *LineType) SetOwner(h handle.Handler) {
	lt.owner = h
}

// Name returns a name of LAYER (code 2).
func (lt *LineType) Name() string {
	return lt.name
}

// TotalLength returns total pattern length (code 40).
func (lt *LineType) TotalLength() float64 {
	sum := 0.0
	for _, l := range lt.lengths {
		sum += math.Abs(l)
	}
	return sum
}

// SetLength sets pattern length (code 49).
//     positive value: Dash
//     0.0: Dot
//     negative value: Space
func (lt *LineType) SetLength(ls []float64) {
	lt.lengths = ls
}
