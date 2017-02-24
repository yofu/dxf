package table

import (
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Default layers.
var (
	LY_0 = NewLayer("0", color.White, LT_CONTINUOUS)
)

// Layer represents LAYER SymbolTable.
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

// NewLayer creates a new Layer.
func NewLayer(name string, color color.ColorNumber, lt *LineType) *Layer {
	l := new(Layer)
	l.name = name
	l.Color = color
	l.LineType = lt
	l.lineWidth = -3
	return l
}

// IsSymbolTable is for SymbolTable interface.
func (l *Layer) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (l *Layer) Format(f format.Formatter) {
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

// String outputs data using default formatter.
func (l *Layer) String() string {
	f := format.NewASCII()
	return l.FormatString(f)
}

// FormatString outputs data using given formatter.
func (l *Layer) FormatString(f format.Formatter) string {
	l.Format(f)
	return f.Output()
}

// Handle returns a handle value.
func (l *Layer) Handle() int {
	return l.handle
}

// SetHandle sets a handle.
func (l *Layer) SetHandle(v *int) {
	l.handle = *v
	(*v)++
}

// SetOwner sets an owner.
func (l *Layer) SetOwner(h handle.Handler) {
	l.owner = h
}

// Name returns a name of LAYER (code 2).
func (l *Layer) Name() string {
	return l.name
}

// SetLineWidth sets line width.
// As DXF has limitation in line width,
// it returns the actual value set to Layer.
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
	for k := range LineWidth {
		tmp := k - w
		if tmp > 0 && tmp < minval {
			minkey = k
			minval = tmp
		}
	}
	l.lineWidth = minkey
	return minkey
}

// SetPlotStyle sets plot style by a handle.
func (l *Layer) SetPlotStyle(ps handle.Handler) {
	l.PlotStyle = ps
}

// SetFlag sets standard flags.
//     1  = Layer is frozen; otherwise layer is thawed.
//     2  = Layer is frozen by default in new viewports.
//     4  = Layer is locked.
//     16 = If set, table entry is externally dependent on an xref.
//     32 = If this bit and bit 16 are both set, the externally dependent xref has been successfully resolved.
//     64 = If set, the table entry was referenced by at least one entity in the drawing the last time the drawing was edited. (This flag is for the benefit of AutoCAD commands. It can be ignored by most programs that read DXF files and need not be set by programs that write DXF files.)
func (l *Layer) SetFlag(val int) {
	l.flag = val
}

// Freeze freezes Layer.
func (l *Layer) Freeze() {
	l.flag |= 1
}

// UnFreeze unfreezes Layer.
func (l *Layer) UnFreeze() {
	l.flag &= ^1
}

// Lock locks Layer.
func (l *Layer) Lock() {
	l.flag |= 4
}

// UnLock unlocks Layer.
func (l *Layer) UnLock() {
	l.flag &= ^4
}
