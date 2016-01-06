package table

import (
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

var (
	LY_0 = NewLayer("0", color.White, LT_CONTINUOUS)
)

type Layer struct {
	handle    int
	owner     handle.Handler
	name      string
	flag      int
	Color     color.ColorNumber
	LineType  *LineType
	lineWidth int
	PlotStyle handle.Handler
}

func NewLayer(name string, color color.ColorNumber, lt *LineType) *Layer {
	l := new(Layer)
	l.name = name
	l.Color = color
	l.LineType = lt
	l.lineWidth = -3
	return l
}

func (l *Layer) IsSymbolTable() bool {
	return true
}

func (l *Layer) Format(f *format.Formatter) {
	f.WriteString(0, "LAYER")
	f.WriteHex(5, l.handle)
	if l.owner != nil {
		f.WriteHex(330, l.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbLayerTableRecord")
	f.WriteString(2, l.name)
	f.WriteInt(70, l.flag)
	f.WriteInt(62, int(l.Color))
	f.WriteString(6, l.LineType.Name())
	f.WriteInt(370, l.lineWidth)
	f.WriteHex(390, l.PlotStyle.Handle())
}

func (l *Layer) String() string {
	f := format.New()
	return l.FormatString(f)
}

func (l *Layer) FormatString(f *format.Formatter) string {
	l.Format(f)
	return f.Output()
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

func (l *Layer) Name() string {
	return l.name
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

func (l *Layer) SetFlag(val int) {
	l.flag = val
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
