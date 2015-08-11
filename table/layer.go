package table

import (
	"bytes"
	"fmt"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/handle"
)

var (
	LY_0 = NewLayer("0", color.White, LT_CONTINUOUS)
)

type Layer struct {
	handle    int
	Name      string
	flag      int
	Color     color.ColorNumber
	LineType  *LineType
	LineWidth int
	PlotStyle handle.Handler
}

func NewLayer(name string, color color.ColorNumber, lt *LineType) *Layer {
	l := new(Layer)
	l.Name = name
	l.Color = color
	l.LineType = lt
	l.LineWidth = -3
	return l
}

func (l *Layer) IsSymbolTable() bool {
	return true
}

func (l *Layer) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nLAYER\n")
	otp.WriteString(fmt.Sprintf("5\n%X\n", l.handle))
	otp.WriteString("100\nAcDbSymbolTableRecord\n100\nAcDbLayerTableRecord\n")
	otp.WriteString(fmt.Sprintf("2\n%s\n", l.Name))
	otp.WriteString(fmt.Sprintf("70\n%d\n", l.flag))
	otp.WriteString(fmt.Sprintf("62\n%d\n", l.Color))
	otp.WriteString(fmt.Sprintf("6\n%s\n", l.LineType.Name))
	otp.WriteString(fmt.Sprintf("370\n%d\n", l.LineWidth))
	otp.WriteString(fmt.Sprintf("390\n%X\n", l.PlotStyle.Handle()))
	return otp.String()
}

func (l *Layer) Handle() int {
	return l.handle
}
func (l *Layer) SetHandle(v *int) {
	l.handle = *v
	(*v)++
}

func (l *Layer) SetPlotStyle(ps handle.Handler) {
	l.PlotStyle = ps
}
