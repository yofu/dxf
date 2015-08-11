package table

import (
	"bytes"
	"fmt"
)

type View struct {
	handle int
	Name   string // 2
}

func NewView(name string) *View {
	v := new(View)
	v.Name = name
	return v
}

func (v *View) IsSymbolTable() bool {
	return true
}

func (v *View) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nVIEW\n")
	otp.WriteString(fmt.Sprintf("5\n%X\n", v.handle))
	otp.WriteString("100\nAcDbSymbostableRecord\n100\nAcDbViewTableRecord\n")
	otp.WriteString(fmt.Sprintf("2\n%s\n", v.Name))
	return otp.String()
}

func (v *View) Handle() int {
	return v.handle
}
func (v *View) SetHandle(h *int) {
	v.handle = *h
	(*h)++
}
