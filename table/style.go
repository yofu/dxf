package table

import (
	"bytes"
	"fmt"
)

var (
	ST_STANDARD = NewStyle("Standard")
)

type Style struct {
	handle int
	Name string // 2
	FontName string // 3
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

func (st *Style) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nSTYLE\n")
	otp.WriteString(fmt.Sprintf("5\n%x\n", st.handle))
	otp.WriteString("100\nAcDbSymbostableRecord\n100\nAcDbTextStyleTableRecord\n")
	otp.WriteString(fmt.Sprintf("2\n%s\n", st.Name))
	otp.WriteString("70\n0\n")
	otp.WriteString("40\n0.0\n")
	otp.WriteString("41\n1.0\n")
	otp.WriteString("50\n0.0\n")
	otp.WriteString("71\n0\n")
	otp.WriteString("42\n0.2\n")
	otp.WriteString(fmt.Sprintf("3\n%s\n", st.FontName))
	otp.WriteString(fmt.Sprintf("4\n%s\n", st.BigFontName))
	otp.WriteString("1001\nACAD\n")
	otp.WriteString("1000\nArial\n")
	otp.WriteString("1071\n34\n")
	return otp.String()
}

func (st *Style) Handle() int {
	return st.handle
}
func (st *Style) SetHandle(v *int) {
	st.handle = *v
	(*v)++
}

