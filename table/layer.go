package table

import (
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Default layers.
//
// LY_0 is a template only: each Drawing has its own layer "0" in its
// LAYER table, so changes to LY_0 are not reflected in drawings.
var (
	LY_0 = NewLayer("0", color.White, LT_CONTINUOUS)
)

// Layer represents LAYER SymbolTable.
type Layer struct {
	ExtGroups
	handle       int
	src          int // handle in the file it was read from (kept when writing)
	owner        handle.Handler
	name         string
	flag         int
	off          bool              // written as negative color number (62)
	Color        color.ColorNumber // 62 (absolute value)
	trueColor    int               // 420: 0x00RRGGBB, -1 = none
	colorName    string            // 430
	transparency int               // XDATA AcCmTransparency (1071): 0 = none
	LineType     *LineType
	lineWidth    int
	PlotStyle    handle.Handler
	// Extra holds group codes read from a file that are not interpreted by
	// this package (and XDATA); they are written after the other values.
	Extra [][2]string
}

// NewLayer creates a new Layer.
func NewLayer(name string, color color.ColorNumber, lt *LineType) *Layer {
	l := new(Layer)
	l.name = name
	l.Color = color
	l.LineType = lt
	l.lineWidth = -3
	l.trueColor = -1
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
	l.writeGroups(f)
	if l.owner != nil {
		f.WriteHex(330, l.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbLayerTableRecord")
	f.WriteString(2, l.name)
	f.WriteInt(70, l.flag)
	// color number is negative if the layer is off (DXF reference, LAYER).
	if l.off {
		f.WriteInt(62, -int(l.Color))
	} else {
		f.WriteInt(62, int(l.Color))
	}
	if format.AtLeast(f, format.R2004) {
		if l.trueColor >= 0 {
			f.WriteInt(420, l.trueColor)
		}
		if l.colorName != "" {
			f.WriteString(430, l.colorName)
		}
	}
	if l.LineType != nil {
		f.WriteString(6, l.LineType.Name())
	} else {
		f.WriteString(6, LT_CONTINUOUS.Name())
	}
	f.WriteInt(370, l.lineWidth)
	if l.PlotStyle != nil {
		f.WriteHex(390, l.PlotStyle.Handle())
	}
	writeRaw(f, l.Extra)
	// the transparency of layers is in the extended data (AC1018 or later)
	if l.transparency != 0 && format.AtLeast(f, format.R2004) {
		f.WriteString(1001, TransparencyApp)
		f.WriteInt(1071, l.transparency)
	}
}

// On turns the layer on.
func (l *Layer) On() {
	l.off = false
}

// Off turns the layer off.
func (l *Layer) Off() {
	l.off = true
}

// IsOn reports whether the layer is on.
func (l *Layer) IsOn() bool {
	return !l.off
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
	l.handle = handle.Next(l.src, v)
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
// LineWidth returns the lineweight of the layer (370; -3 DEFAULT).
func (l *Layer) LineWidth() int { return l.lineWidth }

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
//
//	1  = Layer is frozen; otherwise layer is thawed.
//	2  = Layer is frozen by default in new viewports.
//	4  = Layer is locked.
//	16 = If set, table entry is externally dependent on an xref.
//	32 = If this bit and bit 16 are both set, the externally dependent xref has been successfully resolved.
//	64 = If set, the table entry was referenced by at least one entity in the drawing the last time the drawing was edited. (This flag is for the benefit of AutoCAD commands. It can be ignored by most programs that read DXF files and need not be set by programs that write DXF files.)
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

// SetName renames the layer. Use Drawing.RenameLayer for a layer of a drawing.
func (l *Layer) SetName(name string) {
	l.name = name
}

// SourceHandle returns the handle in the file it was read from ("" if new).
func (l *Layer) SourceHandle() string {
	return handle.Format(l.src)
}

// SetSourceHandle sets the handle in the file it was read from;
// it is kept when the drawing is written.
func (l *Layer) SetSourceHandle(h string) {
	l.src = handle.Parse(h)
}

// LineWeight returns the lineweight of the layer (code 370, 1/100 mm;
// -3 is DEFAULT).
func (l *Layer) LineWeight() int {
	return l.lineWidth
}

// SetLineWeight sets the lineweight of the layer (see SetLineWidth).
func (l *Layer) SetLineWeight(w int) int {
	return l.SetLineWidth(w)
}

// TransparencyApp is the application name of the extended data holding the
// transparency of layers.
const TransparencyApp = "AcCmTransparency"

// TrueColor returns the true color of the layer (code 420); ok is false if not set.
func (l *Layer) TrueColor() (r, g, b uint8, ok bool) {
	if l.trueColor < 0 {
		return 0, 0, 0, false
	}
	r, g, b = color.SplitRGB(l.trueColor)
	return r, g, b, true
}

// SetTrueColor sets the true color of the layer (code 420, AC1018 or
// later); the color number is set to the nearest one.
func (l *Layer) SetTrueColor(r, g, b uint8) {
	l.trueColor = color.RGB(r, g, b)
	l.Color = color.Nearest(r, g, b)
}

// SetTrueColorValue sets a true color value (0x00RRGGBB) read from a file.
func (l *Layer) SetTrueColorValue(v int) {
	l.trueColor = v & 0xFFFFFF
}

// ClearTrueColor removes the true color and the color name of the layer.
func (l *Layer) ClearTrueColor() {
	l.trueColor = -1
	l.colorName = ""
}

// ColorName returns the color name of the layer (code 430).
func (l *Layer) ColorName() string {
	return l.colorName
}

// SetColorName sets the color name of the layer (code 430, "book$name").
func (l *Layer) SetColorName(n string) {
	l.colorName = n
}

// Transparency returns the transparency of the layer (0: opaque, 1: fully
// transparent); ok is false if not set.
func (l *Layer) Transparency() (float64, bool) {
	return color.DecodeTransparency(l.transparency)
}

// HasTransparency reports whether the layer has a transparency.
func (l *Layer) HasTransparency() bool {
	return l.transparency != 0
}

// SetTransparency sets the transparency of the layer (0: opaque, 1: fully
// transparent; written for AC1018 or later).
func (l *Layer) SetTransparency(t float64) {
	l.transparency = color.EncodeTransparency(t)
}

// SetTransparencyValue sets a transparency value (code 1071 of
// AcCmTransparency) read from a file.
func (l *Layer) SetTransparencyValue(v int) {
	l.transparency = v
}

// ClearTransparency removes the transparency of the layer.
func (l *Layer) ClearTransparency() {
	l.transparency = 0
}
