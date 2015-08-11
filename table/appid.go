package table

import (
	"bytes"
	"fmt"
)

type AppID struct {
	handle   int
	Name     string
}

func NewAppID(name string) *AppID {
	a := new(AppID)
	a.Name = name
	return a
}

func (a *AppID) IsSymbolTable() bool {
	return true
}

func (a *AppID) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nAPPID\n")
	otp.WriteString(fmt.Sprintf("5\n%X\n", a.handle))
	otp.WriteString("100\nAcDbSymbolTableRecord\n100\nAcDbRegAppTableRecord\n")
	otp.WriteString(fmt.Sprintf("2\n%s\n", a.Name))
	otp.WriteString("70\n0\n")
	return otp.String()
}

func (a *AppID) Handle() int {
	return a.handle
}
func (a *AppID) SetHandle(v *int) {
	a.handle = *v
	(*v)++
}
