package table

import (
	"bytes"
	"fmt"
)

var (
	LT_CONTINUOUS = NewLineType("Continuous", "Solid Line")
	LT_BYLAYER    = NewLineType("ByLayer", "")
	LT_BYBLOCK    = NewLineType("ByBlock", "")
)

type LineType struct {
	handle      int
	Name        string // 2
	Description string // 3
}

func NewLineType(name, desc string) *LineType {
	lt := new(LineType)
	lt.Name = name
	lt.Description = desc
	return lt
}

func (lt *LineType) IsSymbolTable() bool {
	return true
}

func (lt *LineType) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nLTYPE\n")
	otp.WriteString(fmt.Sprintf("5\n%x\n", lt.handle))
	otp.WriteString("100\nAcDbSymbolTableRecord\n100\nAcDbLinetypeTableRecord\n")
	otp.WriteString(fmt.Sprintf("2\n%s\n", lt.Name))
	otp.WriteString("70\n0\n")
	otp.WriteString(fmt.Sprintf("3\n%s\n", lt.Description))
	otp.WriteString("72\n65\n")
	otp.WriteString("73\n0\n")
	otp.WriteString("40\n0.0\n")
	return otp.String()
}

func (lt *LineType) Handle() int {
	return lt.handle
}
func (lt *LineType) SetHandle(v *int) {
	lt.handle = *v
	(*v)++
}
