package table

import (
	"bytes"
	"fmt"
	"github.com/yofu/dxf/color"
)

var (
	LY_0 = NewLayer("0", color.White, LT_CONTINUOUS)
)

type Layer struct {
	handle int
	Name string
	Color color.ColorNumber
	LineType *LineType
}

func NewLayer(name string, color color.ColorNumber, lt *LineType) *Layer {
	l := new(Layer)
	l.Name = name
	l.Color = color
	l.LineType = lt
	return l
}

func (l *Layer) IsSymbolTable() bool {
	return true
}

func (l *Layer) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nLAYER\n")
	otp.WriteString(fmt.Sprintf("5\n%x\n", l.handle))
	otp.WriteString("100\nAcDbSymbolableRecord\n100\nAcDbLayerTableRecord\n")
	otp.WriteString(fmt.Sprintf("2\n%s\n", l.Name))
	otp.WriteString("70\n0\n")
	otp.WriteString(fmt.Sprintf("62\n%d\n", l.Color))
	otp.WriteString(fmt.Sprintf("6\n%s\n", l.LineType.Name))
	// otp.WriteString("370\n-3\n")
	return otp.String()
}

func (l *Layer) Handle() int {
	return l.handle
}
func (l *Layer) SetHandle(v *int) {
	l.handle = *v
	(*v)++
}
