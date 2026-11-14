package entity

import (
	"math"
	"strings"

	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Underlay file formats.
const (
	UNDERLAY_PDF = "PDF"
	UNDERLAY_DWF = "DWF"
	UNDERLAY_DGN = "DGN"
)

// Display flags of underlays (code 280).
const (
	UNDERLAY_CLIPPING          = 1
	UNDERLAY_ON                = 2
	UNDERLAY_MONOCHROME        = 4
	UNDERLAY_ADJUST_BACKGROUND = 8
)

// UnderlayDefinition represents PDFDEFINITION, DWFDEFINITION and
// DGNDEFINITION objects (OBJECTS section): the file of underlays.
// Use Drawing.AddUnderlayDefinition to create one.
type UnderlayDefinition struct {
	FileFormat string // PDF, DWF or DGN
	FileName   string // 1: file name (relative or absolute path)
	Name       string // 2: page number (PDF), sheet (DWF) or model (DGN)

	handle   int
	src      int
	owner    handle.Handler   // ACAD_PDFDEFINITIONS dictionary, ...
	reactors []handle.Handler // underlays referring to it (set when writing)
}

// NewUnderlayDefinition creates an underlay definition of fileFormat
// (UNDERLAY_PDF, ...), file name and name (page number for PDF).
func NewUnderlayDefinition(fileFormat, filename, name string) *UnderlayDefinition {
	return &UnderlayDefinition{FileFormat: strings.ToUpper(fileFormat), FileName: filename, Name: name}
}

// IsObject is for object.Object interface.
func (u *UnderlayDefinition) IsObject() bool { return true }

// TypeName returns the object type (e.g. "PDFDEFINITION").
func (u *UnderlayDefinition) TypeName() string { return u.FileFormat + "DEFINITION" }

// EntityTypeName returns the type of its underlays (e.g. "PDFUNDERLAY").
func (u *UnderlayDefinition) EntityTypeName() string { return u.FileFormat + "UNDERLAY" }

// DictionaryName returns the name of the dictionary of the definitions of
// the format (e.g. "ACAD_PDFDEFINITIONS").
func (u *UnderlayDefinition) DictionaryName() string { return "ACAD_" + u.FileFormat + "DEFINITIONS" }

// Format writes data to formatter.
func (u *UnderlayDefinition) Format(f format.Formatter) {
	f.WriteString(0, u.TypeName())
	f.WriteHex(5, u.handle)
	f.WriteString(102, "{ACAD_REACTORS")
	if u.owner != nil {
		f.WriteHex(330, u.owner.Handle())
	}
	for _, r := range u.reactors {
		f.WriteHex(330, r.Handle())
	}
	f.WriteString(102, "}")
	if u.owner != nil {
		f.WriteHex(330, u.owner.Handle())
	} else {
		f.WriteString(330, "0")
	}
	f.WriteString(100, "AcDbUnderlayDefinition")
	f.WriteString(1, u.FileName)
	f.WriteString(2, u.Name)
}

// Handle returns the handle.
func (u *UnderlayDefinition) Handle() int { return u.handle }

// SetHandle sets the handle (the handle read is kept).
func (u *UnderlayDefinition) SetHandle(v *int) { u.handle = handle.Next(u.src, v) }

// SetOwner sets the owner dictionary.
func (u *UnderlayDefinition) SetOwner(h handle.Handler) { u.owner = h }

// SetReactors sets the underlays referring to the definition.
func (u *UnderlayDefinition) SetReactors(hs ...handle.Handler) { u.reactors = hs }

// SourceHandle returns the handle in the file it was read from ("" if new).
func (u *UnderlayDefinition) SourceHandle() string { return handle.Format(u.src) }

// SetSourceHandle sets the handle in the file it was read from;
// it is kept when the drawing is written.
func (u *UnderlayDefinition) SetSourceHandle(h string) { u.src = handle.Parse(h) }

// Underlay represents PDFUNDERLAY, DWFUNDERLAY and DGNUNDERLAY entities:
// a reference to an UnderlayDefinition placed in the drawing.
type Underlay struct {
	*entity
	Definition *UnderlayDefinition // 340
	Insert     []float64           // 10, 20, 30
	Scale      []float64           // 41, 42, 43
	Rotation   float64             // 50 (degrees)
	Extrusion  []float64           // 210, 220, 230
	Flags      int                 // 280: UNDERLAY_* display flags
	Contrast   int                 // 281: 20 to 100
	Fade       int                 // 282: 0 to 80
	Boundary   [][]float64         // 11, 21: clipping boundary in underlay coordinates (2 points: rectangle)

	defHandle string // 340 as read, until it is resolved to Definition
}

// IsEntity is for Entity interface.
func (u *Underlay) IsEntity() bool { return true }

// NewUnderlay creates an underlay of definition def (nil: set it later)
// at the origin with scale 1.
func NewUnderlay(def *UnderlayDefinition) *Underlay {
	u := &Underlay{
		entity:     NewEntity(UNDERLAY),
		Definition: def,
		Insert:     []float64{0, 0, 0},
		Scale:      []float64{1, 1, 1},
		Extrusion:  []float64{0, 0, 1},
		Flags:      UNDERLAY_ON,
		Contrast:   100,
	}
	u.typeName = UNDERLAY_PDF + "UNDERLAY"
	if def != nil {
		u.typeName = def.EntityTypeName()
	}
	return u
}

// TypeName returns the entity type (e.g. "PDFUNDERLAY").
func (u *Underlay) TypeName() string { return u.typeName }

// SetTypeName sets the entity type ("PDFUNDERLAY", "DWFUNDERLAY" or
// "DGNUNDERLAY"); it is set from the definition when writing.
func (u *Underlay) SetTypeName(name string) { u.typeName = strings.ToUpper(name) }

// DefinitionHandle returns the handle of the definition as read when it is
// not resolved to Definition ("" otherwise).
func (u *Underlay) DefinitionHandle() string { return u.defHandle }

// SetDefinitionHandle sets the handle of the definition (used when the
// definition is not resolved; Definition takes precedence).
func (u *Underlay) SetDefinitionHandle(h string) { u.defHandle = h }

// SetClipping sets a rectangular clipping boundary (in underlay
// coordinates) and turns clipping on; nil corners remove it.
func (u *Underlay) SetClipping(min, max []float64) {
	if min == nil || max == nil {
		u.Boundary = nil
		u.Flags &^= UNDERLAY_CLIPPING
		return
	}
	u.Boundary = [][]float64{{min[0], min[1]}, {max[0], max[1]}}
	u.Flags |= UNDERLAY_CLIPPING
}

// Format writes data to formatter.
func (u *Underlay) Format(f format.Formatter) {
	if u.Definition != nil {
		u.typeName = u.Definition.EntityTypeName()
	}
	u.entity.Format(f)
	f.WriteString(100, "AcDbUnderlayReference")
	switch {
	case u.Definition != nil:
		f.WriteHex(340, u.Definition.Handle())
	case u.defHandle != "":
		f.WriteString(340, u.defHandle)
	default:
		f.WriteString(340, "0")
	}
	writePoint(f, 10, u.Insert)
	f.WriteFloat(41, u.Scale[0])
	f.WriteFloat(42, u.Scale[1])
	f.WriteFloat(43, u.Scale[2])
	f.WriteFloat(50, u.Rotation)
	if u.Extrusion[0] != 0 || u.Extrusion[1] != 0 || u.Extrusion[2] != 1 {
		writePoint(f, 210, u.Extrusion)
	}
	f.WriteInt(280, u.Flags)
	f.WriteInt(281, u.Contrast)
	f.WriteInt(282, u.Fade)
	for _, p := range u.Boundary {
		f.WriteFloat(11, p[0])
		f.WriteFloat(21, p[1])
	}
}

// String outputs data using default formatter.
func (u *Underlay) String() string {
	f := format.NewASCII()
	return u.FormatString(f)
}

// FormatString outputs data using given formatter.
func (u *Underlay) FormatString(f format.Formatter) string {
	u.Format(f)
	return f.Output()
}

// BBox returns the bounding box of the clipping boundary (transformed by
// insertion point, scale and rotation), or the insertion point if there is
// no boundary: the size of the underlay file is not known.
func (u *Underlay) BBox() ([]float64, []float64) {
	if len(u.Boundary) == 0 {
		return append([]float64{}, u.Insert...), append([]float64{}, u.Insert...)
	}
	pts := u.Boundary
	if len(pts) == 2 {
		pts = [][]float64{pts[0], {pts[1][0], pts[0][1]}, pts[1], {pts[0][0], pts[1][1]}}
	}
	rad := u.Rotation * math.Pi / 180
	s, c := math.Sin(rad), math.Cos(rad)
	mins := []float64{math.Inf(1), math.Inf(1), u.Insert[2]}
	maxs := []float64{math.Inf(-1), math.Inf(-1), u.Insert[2]}
	for _, p := range pts {
		x, y := p[0]*u.Scale[0], p[1]*u.Scale[1]
		q := []float64{u.Insert[0] + x*c - y*s, u.Insert[1] + x*s + y*c}
		for i := 0; i < 2; i++ {
			mins[i] = math.Min(mins[i], q[i])
			maxs[i] = math.Max(maxs[i], q[i])
		}
	}
	return mins, maxs
}
