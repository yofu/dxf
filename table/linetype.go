package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
	"math"
)

var (
	LT_CONTINUOUS = NewLineType("Continuous", "Solid Line")
	LT_BYLAYER    = NewLineType("ByLayer", "")
	LT_BYBLOCK    = NewLineType("ByBlock", "")
	LT_HIDDEN     = NewLineType("HIDDEN", "Hidden __ __ __ __ __ __ __ __ __ __ __ __ __ _", 0.25, -0.125)
	LT_DASHDOT    = NewLineType("DASHDOT", "Dash dot __ . __ . __ . __ . __ . __ . __ . __", 0.5, -0.25, 0.0, -0.25)
)

type LineType struct {
	handle      int
	owner       handle.Handler
	name        string // 2
	Description string // 3
	lengths     []float64
}

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

func (lt *LineType) IsSymbolTable() bool {
	return true
}

func (lt *LineType) Format(f *format.Formatter) {
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

func (lt *LineType) String() string {
	f := format.New()
	return lt.FormatString(f)
}

func (lt *LineType) FormatString(f *format.Formatter) string {
	lt.Format(f)
	return f.Output()
}

func (lt *LineType) Handle() int {
	return lt.handle
}
func (lt *LineType) SetHandle(v *int) {
	lt.handle = *v
	(*v)++
}

func (lt *LineType) SetOwner(h handle.Handler) {
	lt.owner = h
}

func (lt *LineType) Name() string {
	return lt.name
}

func (lt *LineType) TotalLength() float64 {
	sum := 0.0
	for _, l := range lt.lengths {
		sum += math.Abs(l)
	}
	return sum
}

func (lt *LineType) SetLength(ls []float64) {
	lt.lengths = ls
}
