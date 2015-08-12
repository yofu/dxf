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
	owner     handle.Handler
	Name      string
	flag      int
	Color     color.ColorNumber
	LineType  *LineType
	lineWidth int
	PlotStyle handle.Handler
}

func NewLayer(name string, color color.ColorNumber, lt *LineType) *Layer {
	l := new(Layer)
	l.Name = name
	l.Color = color
	l.LineType = lt
	l.lineWidth = -3
	return l
}

func (l *Layer) IsSymbolTable() bool {
	return true
}

func (l *Layer) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nLAYER\n")
	otp.WriteString(fmt.Sprintf("5\n%X\n", l.handle))
	if l.owner != nil {
		otp.WriteString(fmt.Sprintf("330\n%X\n", l.owner.Handle()))
	}
	otp.WriteString("100\nAcDbSymbolTableRecord\n100\nAcDbLayerTableRecord\n")
	otp.WriteString(fmt.Sprintf("2\n%s\n", l.Name))
	otp.WriteString(fmt.Sprintf("70\n%d\n", l.flag))
	otp.WriteString(fmt.Sprintf("62\n%d\n", l.Color))
	otp.WriteString(fmt.Sprintf("6\n%s\n", l.LineType.Name))
	otp.WriteString(fmt.Sprintf("370\n%d\n", l.lineWidth))
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

func (l *Layer) SetOwner(h handle.Handler) {
	l.owner = h
}

func (l *Layer) SetLineWidth(w int) int {
	if _, ok := LineWidth[w]; ok {
		l.lineWidth = w
		return w
	}
	if w > 211 {
		l.lineWidth = 211
		return 211
	}
	if w < 0 {
		l.lineWidth = -3
		return -3
	}
	minkey := -3
	minval := 211
	for k, _ := range LineWidth {
		tmp := k-w
		if tmp > 0 && tmp < minval {
			minkey = k
			minval = tmp
		}
	}
	l.lineWidth = minkey
	return minkey
}

func (l *Layer) SetPlotStyle(ps handle.Handler) {
	l.PlotStyle = ps
}

func (l *Layer) Freeze() {
	l.flag |= 1
}
func (l *Layer) UnFreeze() {
	l.flag &= ^1
}

func (l *Layer) Lock() {
	l.flag |= 4
}
func (l *Layer) UnLock() {
	l.flag &= ^4
}
