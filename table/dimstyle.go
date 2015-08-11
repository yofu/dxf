package table

import (
	"bytes"
	"fmt"
)

type DimStyle struct {
	handle   int
	Name     string
}

func NewDimStyle(name string) *DimStyle {
	d := new(DimStyle)
	d.Name = name
	return d
}

func (d *DimStyle) IsSymbolTable() bool {
	return true
}

func (d *DimStyle) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nDIMSTYLE\n")
	otp.WriteString(fmt.Sprintf("105\n%x\n", d.handle))
	otp.WriteString("100\nAcDbSymbolTableRecord\n100\nAcDbDimStyleTableRecord\n")
	otp.WriteString(fmt.Sprintf("2\n%s\n", d.Name))
	otp.WriteString("70\n0\n")
	return otp.String()
}

func (d *DimStyle) Handle() int {
	return d.handle
}
func (d *DimStyle) SetHandle(v *int) {
	d.handle = *v
	(*v)++
}
