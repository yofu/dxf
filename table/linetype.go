package table

import (
	"bytes"
	"fmt"
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
	Name        string // 2
	Description string // 3
	lengths     []float64
}

func NewLineType(name, desc string, ls ...float64) *LineType {
	lt := new(LineType)
	lt.Name = name
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

func (lt *LineType) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nLTYPE\n")
	otp.WriteString(fmt.Sprintf("5\n%X\n", lt.handle))
	if lt.owner != nil {
		otp.WriteString(fmt.Sprintf("330\n%X\n", lt.owner.Handle()))
	}
	otp.WriteString("100\nAcDbSymbolTableRecord\n100\nAcDbLinetypeTableRecord\n")
	otp.WriteString(fmt.Sprintf("2\n%s\n", lt.Name))
	otp.WriteString("70\n0\n")
	otp.WriteString(fmt.Sprintf("3\n%s\n", lt.Description))
	otp.WriteString("72\n65\n")
	otp.WriteString(fmt.Sprintf("73\n%d\n", len(lt.lengths)))
	otp.WriteString(fmt.Sprintf("40\n%f\n", lt.TotalLength()))
	for _, l := range lt.lengths {
		otp.WriteString(fmt.Sprintf("49\n%f\n", l))
		otp.WriteString("74\n0\n")
	}
	return otp.String()
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
