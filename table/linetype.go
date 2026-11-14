package table

import (
	"math"

	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Default LineTypes. They are templates: each Drawing has its own copies
// in its LTYPE table (use Drawing.LineType), so changes to these variables
// are not reflected in drawings.
var (
	LT_CONTINUOUS = NewLineType("Continuous", "Solid Line")
	LT_BYLAYER    = NewLineType("ByLayer", "")
	LT_BYBLOCK    = NewLineType("ByBlock", "")
	LT_HIDDEN     = NewLineType("HIDDEN", "Hidden __ __ __ __ __ __ __ __ __ __ __ __ __ _", 0.25, -0.125)
	LT_DASHDOT    = NewLineType("DASHDOT", "Dash dot __ . __ . __ . __ . __ . __ . __ . __", 0.5, -0.25, 0.0, -0.25)
)

// LineType represents LTYPE SymbolTable.
type LineType struct {
	handle      int
	src         int // handle in the file it was read from (kept when writing)
	owner       handle.Handler
	name        string // 2
	Description string // 3
	elements    []*LineTypeElement
	total       float64 // 40 as read (< 0: computed from the elements)
	// Extra holds group codes read from a file that are not interpreted by
	// this package (and XDATA); they are written after the other values.
	Extra [][2]string
}

// Complex element flags of linetypes (code 74).
const (
	LT_ABSOLUTE_ROTATION = 1 // rotation is absolute (otherwise relative to the line)
	LT_TEXT              = 2 // the element is a text string
	LT_SHAPE             = 4 // the element is a shape
)

// LineTypeElement is an element of a linetype pattern: a dash (positive
// length), a space (negative length) or a dot (0), optionally with a text
// string or a shape drawn at its start (complex linetypes).
type LineTypeElement struct {
	Length   float64   // 49
	Flags    int       // 74: LT_ABSOLUTE_ROTATION, LT_TEXT, LT_SHAPE (0: simple element)
	Shape    int       // 75: shape number (LT_SHAPE)
	Style    *Style    // 340: text style (LT_TEXT) or shape file (LT_SHAPE)
	Scale    float64   // 46
	Rotation float64   // 50: radians
	Offset   []float64 // 44, 45: shift along and normal to the line
	Text     string    // 9 (LT_TEXT)

	styleHandle string // 340 as read, until it is resolved to Style
}

// IsComplex reports whether the element has a text string or a shape.
func (e *LineTypeElement) IsComplex() bool {
	return e.Flags&(LT_TEXT|LT_SHAPE) != 0
}

// StyleHandle returns the handle of the text style or shape file as read
// when it is not resolved to Style ("" otherwise).
func (e *LineTypeElement) StyleHandle() string {
	return e.styleHandle
}

// SetStyleHandle sets the handle of the text style or shape file
// (used when the style is not resolved; Style takes precedence).
func (e *LineTypeElement) SetStyleHandle(h string) {
	e.styleHandle = h
}

// TextElement returns an element with a text string drawn in style at the
// start of a space of length (negative) or a dash. rotation is in radians
// (relative to the line); offset is along and normal to the line.
func TextElement(length float64, text string, style *Style, scale, rotation, offsetX, offsetY float64) *LineTypeElement {
	return &LineTypeElement{Length: length, Flags: LT_TEXT, Style: style, Scale: scale,
		Rotation: rotation, Offset: []float64{offsetX, offsetY}, Text: text}
}

// ShapeElement returns an element with shape number shape of the shape file
// (a Style with flag 1, see Drawing.AddShapeFile) drawn at the start of a
// space of length (negative) or a dash.
func ShapeElement(length float64, shape int, shapeFile *Style, scale, rotation, offsetX, offsetY float64) *LineTypeElement {
	return &LineTypeElement{Length: length, Flags: LT_SHAPE, Shape: shape, Style: shapeFile, Scale: scale,
		Rotation: rotation, Offset: []float64{offsetX, offsetY}}
}

func (e *LineTypeElement) format(f format.Formatter) {
	f.WriteFloat(49, e.Length)
	f.WriteInt(74, e.Flags)
	if !e.IsComplex() {
		return
	}
	f.WriteInt(75, e.Shape)
	switch {
	case e.Style != nil:
		f.WriteHex(340, e.Style.Handle())
	case e.styleHandle != "":
		f.WriteString(340, e.styleHandle)
	default:
		f.WriteString(340, "0")
	}
	f.WriteFloat(46, e.Scale)
	f.WriteFloat(50, e.Rotation)
	offset := e.Offset
	if len(offset) < 2 {
		offset = []float64{0, 0}
	}
	f.WriteFloat(44, offset[0])
	f.WriteFloat(45, offset[1])
	if e.Flags&LT_TEXT != 0 {
		f.WriteString(9, e.Text)
	}
}

func (e *LineTypeElement) clone() *LineTypeElement {
	c := *e
	c.Offset = append([]float64(nil), e.Offset...)
	return &c
}

// NewLineType creates a new LineType with pattern lengths (see SetLength).
func NewLineType(name, desc string, ls ...float64) *LineType {
	lt := new(LineType)
	lt.name = name
	lt.Description = desc
	lt.SetLength(ls)
	return lt
}

// IsSymbolTable is for SymbolTable interface.
func (lt *LineType) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (lt *LineType) Format(f format.Formatter) {
	f.WriteString(0, "LTYPE")
	f.WriteHex(5, lt.handle)
	if lt.owner != nil {
		f.WriteHex(330, lt.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbLinetypeTableRecord")
	f.WriteString(2, lt.name)
	f.WriteInt(70, 0)
	f.WriteString(3, lt.Description)
	f.WriteInt(72, 65)
	f.WriteInt(73, len(lt.elements))
	f.WriteFloat(40, lt.TotalLength())
	for _, e := range lt.elements {
		e.format(f)
	}
	writeRaw(f, lt.Extra)
}

// String outputs data using default formatter.
func (lt *LineType) String() string {
	f := format.NewASCII()
	return lt.FormatString(f)
}

// FormatString outputs data using given formatter.
func (lt *LineType) FormatString(f format.Formatter) string {
	lt.Format(f)
	return f.Output()
}

// Handle returns a handle value.
func (lt *LineType) Handle() int {
	return lt.handle
}

// SetHandle sets a handle.
func (lt *LineType) SetHandle(v *int) {
	lt.handle = handle.Next(lt.src, v)
}

// SetOwner sets an owner.
func (lt *LineType) SetOwner(h handle.Handler) {
	lt.owner = h
}

// Name returns a name of LAYER (code 2).
func (lt *LineType) Name() string {
	return lt.name
}

// TotalLength returns total pattern length (code 40): the value read or
// set by SetTotalLength, or the sum of the absolute element lengths.
func (lt *LineType) TotalLength() float64 {
	if lt.total >= 0 {
		return lt.total
	}
	sum := 0.0
	for _, e := range lt.elements {
		sum += math.Abs(e.Length)
	}
	return sum
}

// SetLength sets pattern length (code 49).
//
//	positive value: Dash
//	0.0: Dot
//	negative value: Space
func (lt *LineType) SetLength(ls []float64) {
	lt.total = -1
	lt.elements = make([]*LineTypeElement, len(ls))
	for i, l := range ls {
		lt.elements[i] = &LineTypeElement{Length: l}
	}
}

// Lengths returns the pattern lengths (code 49).
func (lt *LineType) Lengths() []float64 {
	ls := make([]float64, len(lt.elements))
	for i, e := range lt.elements {
		ls[i] = e.Length
	}
	return ls
}

// Elements returns the pattern elements.
func (lt *LineType) Elements() []*LineTypeElement {
	return lt.elements
}

// SetElements sets the pattern elements (complex linetypes).
func (lt *LineType) SetElements(es ...*LineTypeElement) {
	lt.total = -1
	lt.elements = es
}

// SetTotalLength sets total pattern length (code 40); a negative value
// means the sum of the absolute element lengths. SetLength and SetElements
// reset it.
func (lt *LineType) SetTotalLength(l float64) {
	lt.total = l
}

// IsComplex reports whether the linetype has text strings or shapes.
func (lt *LineType) IsComplex() bool {
	for _, e := range lt.elements {
		if e.IsComplex() {
			return true
		}
	}
	return false
}

// clone returns a copy of lt without handle and owner.
func (lt *LineType) clone() *LineType {
	c := NewLineType(lt.name, lt.Description)
	for _, e := range lt.elements {
		c.elements = append(c.elements, e.clone())
	}
	c.total = lt.total
	c.Extra = append([][2]string(nil), lt.Extra...)
	return c
}

// SourceHandle returns the handle in the file it was read from ("" if new).
func (lt *LineType) SourceHandle() string {
	return handle.Format(lt.src)
}

// SetSourceHandle sets the handle in the file it was read from;
// it is kept when the drawing is written.
func (lt *LineType) SetSourceHandle(h string) {
	lt.src = handle.Parse(h)
}
