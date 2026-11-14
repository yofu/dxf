package entity

import (
	"strconv"
	"strings"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
	"github.com/yofu/dxf/table"
)

// Entity is interface for DXF Entities.
type Entity interface {
	IsEntity() bool
	Format(format.Formatter)
	Handle() int
	SetHandle(*int)
	SetBlockRecord(handle.Handler)
	Layer() *table.Layer
	SetLayer(*table.Layer)
	SetLtscale(float64)
	BBox() ([]float64, []float64)
}

// Color number values with special meaning (code 62).
const (
	ColorByBlock = 0   // BYBLOCK
	ColorByLayer = 256 // BYLAYER (default)
)

// Colored is implemented by all entities in this package.
// It gives access to the entity color (code 62).
type Colored interface {
	// ColorIndex returns the color number: 0 (BYBLOCK), 1-255 or 256 (BYLAYER).
	ColorIndex() int
	// SetColor sets an explicit color number (0 means BYBLOCK).
	SetColor(color.ColorNumber)
	// SetColorByLayer sets the color to BYLAYER.
	SetColorByLayer()
	// SetColorByBlock sets the color to BYBLOCK.
	SetColorByBlock()
	// IsColorByLayer reports whether the color is BYLAYER.
	IsColorByLayer() bool
}

// LineProperties is implemented by all entities in this package.
// It gives access to the linetype (code 6) and the lineweight (code 370).
type LineProperties interface {
	// LineType returns the linetype; nil means BYLAYER.
	LineType() *table.LineType
	// LineTypeName returns the linetype name ("ByLayer" if BYLAYER).
	LineTypeName() string
	// SetLineType sets the linetype; nil means BYLAYER.
	SetLineType(*table.LineType)
	// LineWeight returns the lineweight in 1/100 mm, or
	// table.LineWeightByLayer, table.LineWeightByBlock, table.LineWeightDefault.
	LineWeight() int
	// SetLineWeight sets the lineweight (see table.NearestLineWeight) and returns the value set.
	SetLineWeight(int) int
}

// entity is common part of Entities.
// It is embedded in each entities to implement Entity interface.
type entity struct {
	Type         EntityType      // 0
	handle       int             // 5
	blockRecord  handle.Handler  // 102 330
	owner        handle.Handler  // 330
	layer        *table.Layer    // 8
	typeName     string          // 0: entity type name of Unknown entities
	srcHandle    string          // 5: handle in the file the entity was read from
	xdata        [][2]string     // 1001-: extended data, kept as read
	groups       [][2]string     // 102 groups (reactors, extension dictionary), kept as read
	paperSpace   bool            // 67: entity is in paper space
	invisible    bool            // 60: object visibility (1 = invisible)
	lineType     *table.LineType // 6: nil = BYLAYER
	color        int             // 62: 0 = BYBLOCK, 1-255, 256 = BYLAYER
	trueColor    int             // 420: 0x00RRGGBB, -1 = none
	colorName    string          // 430: color name (color book)
	transparency int             // 440: 0 = BYLAYER (none), 0x01000000 = BYBLOCK, 0x020000AA = alpha AA
	lineWeight   int             // 370: -1 = BYLAYER, -2 = BYBLOCK, -3 = DEFAULT, 0-211 (1/100 mm)
	ltscale      float64         // 48
}

// NewEntity creates a new entity.
func NewEntity(t EntityType) *entity {
	e := &entity{
		Type:        t,
		handle:      0,
		blockRecord: nil,
		owner:       nil,
		layer:       table.LY_0,
		color:       ColorByLayer,
		trueColor:   -1,
		lineWeight:  table.LineWeightByLayer,
		ltscale:     1.0,
	}
	return e
}

// Format writes data to formatter.
func (e *entity) Format(f format.Formatter) {
	name := EntityTypeString(e.Type)
	if e.typeName != "" {
		name = e.typeName
	}
	f.WriteString(0, name)
	f.WriteHex(5, e.handle)
	e.writeGroups102(f)
	if e.owner != nil {
		f.WriteHex(330, e.owner.Handle())
	}
	f.WriteString(100, "AcDbEntity")
	if e.paperSpace {
		f.WriteInt(67, 1)
	}
	f.WriteString(8, e.layer.Name())
	// 6, 62 and 370 are omitted if BYLAYER (DXF reference, common group codes for entities)
	if e.lineType != nil && !strings.EqualFold(e.lineType.Name(), "BYLAYER") {
		f.WriteString(6, e.lineType.Name())
	}
	if e.color != ColorByLayer {
		f.WriteInt(62, e.color)
	}
	if e.lineWeight != table.LineWeightByLayer {
		f.WriteInt(370, e.lineWeight)
	}
	if e.ltscale != 1.0 {
		f.WriteFloat(48, e.ltscale)
	}
	if e.invisible {
		f.WriteInt(60, 1)
	}
	if format.AtLeast(f, format.R2004) {
		if e.trueColor >= 0 {
			f.WriteInt(420, e.trueColor)
		}
		if e.colorName != "" {
			f.WriteString(430, e.colorName)
		}
	}
	if e.transparency != 0 && format.AtLeast(f, format.R2004) {
		f.WriteInt(440, e.transparency)
	}
}

// String outputs data using default formatter.
func (e *entity) String() string {
	f := format.NewASCII()
	return e.FormatString(f)
}

// FormatString outputs data using given formatter.
func (e *entity) FormatString(f format.Formatter) string {
	e.Format(f)
	return f.Output()
}

// Handle returns a handle value of TABLE.
func (e *entity) Handle() int {
	return e.handle
}

// SetHandle sets handles to TABLE itself and each SymbolTable.
func (e *entity) SetHandle(v *int) {
	e.handle = handle.Next(handle.Parse(e.srcHandle), v)
}

// SetBlockRecord sets BLOCK_RECORD to entity (code 330).
func (e *entity) SetBlockRecord(h handle.Handler) {
	e.blockRecord = h
}

// SetOwner sets an owner.
func (e *entity) SetOwner(h handle.Handler) {
	e.owner = h
}

// Layer returns entity's Layer.
func (e *entity) Layer() *table.Layer {
	return e.layer
}

// SetLayer sets Layer to entity.
func (e *entity) SetLayer(l *table.Layer) {
	e.layer = l
}

// SetLtscale sets Layer to entity.
func (e *entity) SetLtscale(v float64) {
	e.ltscale = v
}

// SetEntityType sets entity type.
func (e *entity) SetEntityType(t EntityType) {
	e.Type = t
}

// ColorIndex returns the color number: 0 (BYBLOCK), 1-255 or 256 (BYLAYER).
func (e *entity) ColorIndex() int {
	return e.color
}

// SetColor sets an explicit color number (0 means BYBLOCK).
func (e *entity) SetColor(c color.ColorNumber) {
	e.color = int(c)
}

// SetColorByLayer sets the color to BYLAYER.
func (e *entity) SetColorByLayer() {
	e.color = ColorByLayer
}

// SetColorByBlock sets the color to BYBLOCK.
func (e *entity) SetColorByBlock() {
	e.color = ColorByBlock
}

// IsColorByLayer reports whether the color is BYLAYER.
func (e *entity) IsColorByLayer() bool {
	return e.color == ColorByLayer
}

// setColorIndex sets a raw color number read from a file.
// Negative values (layer off) are stored as their absolute value,
// and values out of range fall back to BYLAYER.
func (e *entity) setColorIndex(c int) {
	if c < 0 {
		c = -c
	}
	if c > ColorByLayer {
		c = ColorByLayer
	}
	e.color = c
}

// SetColorIndex sets a raw color number (0-256) to an entity implementing Colored.
// It is used by the parser; negative values are treated as their absolute value.
func SetColorIndex(e Entity, c int) bool {
	type colorIndexSetter interface{ setColorIndex(int) }
	if s, ok := e.(colorIndexSetter); ok {
		s.setColorIndex(c)
		return true
	}
	return false
}

var _ Colored = &entity{}

// LineType returns the linetype; nil means BYLAYER.
func (e *entity) LineType() *table.LineType {
	return e.lineType
}

// LineTypeName returns the linetype name ("ByLayer" if BYLAYER).
func (e *entity) LineTypeName() string {
	if e.lineType == nil {
		return table.LT_BYLAYER.Name()
	}
	return e.lineType.Name()
}

// SetLineType sets the linetype; nil means BYLAYER.
// Use the linetypes of the drawing (Drawing.LineType) so that they are
// defined in the LTYPE table.
func (e *entity) SetLineType(lt *table.LineType) {
	if lt != nil && strings.EqualFold(lt.Name(), "BYLAYER") {
		lt = nil
	}
	e.lineType = lt
}

// LineWeight returns the lineweight in 1/100 mm, or
// table.LineWeightByLayer (default), table.LineWeightByBlock, table.LineWeightDefault.
func (e *entity) LineWeight() int {
	return e.lineWeight
}

// SetLineWeight sets the lineweight in 1/100 mm. Values that are not valid
// lineweights are rounded up to the next valid value (see table.NearestLineWeight).
// It returns the value set.
func (e *entity) SetLineWeight(w int) int {
	e.lineWeight = table.NearestLineWeight(w)
	return e.lineWeight
}

var _ LineProperties = &entity{}

// Thick is implemented by entities with a thickness (code 39):
// LINE, CIRCLE, ARC, POINT, TEXT, ATTDEF, ATTRIB, LWPOLYLINE, POLYLINE and SOLID.
// The thickness is the extrusion along the extrusion direction.
type Thick interface {
	SetThickness(float64)
}

func writeThickness(f format.Formatter, t float64) {
	if t != 0 {
		f.WriteFloat(39, t)
	}
}

// extrude extends a bounding box by thickness t along Z
// (the extrusion direction is assumed to be the Z axis).
func extrude(mins, maxs []float64, t float64) ([]float64, []float64) {
	if t > 0 {
		maxs[2] += t
	} else {
		mins[2] += t
	}
	return mins, maxs
}

// SetThickness sets the thickness.
func (l *Line) SetThickness(t float64) { l.Thickness = t }

// SetThickness sets the thickness.
func (c *Circle) SetThickness(t float64) { c.Thickness = t }

// SetThickness sets the thickness.
func (p *Point) SetThickness(t float64) { p.Thickness = t }

// SetThickness sets the thickness.
func (t *Text) SetThickness(v float64) { t.Thickness = v }

// SetThickness sets the thickness.
func (l *LwPolyline) SetThickness(t float64) { l.Thickness = t }

// SetThickness sets the thickness.
func (p *Polyline) SetThickness(t float64) { p.Thickness = t }

// SetThickness sets the thickness.
func (s *Solid) SetThickness(t float64) { s.Thickness = t }

// SpaceSelector is implemented by all entities in this package.
// Entities in the ENTITIES section with InPaperSpace belong to the paper
// space layout (*Paper_Space, code 67 = 1).
type SpaceSelector interface {
	InPaperSpace() bool
	SetPaperSpace(bool)
}

// InPaperSpace reports whether the entity is in paper space (code 67).
func (e *entity) InPaperSpace() bool {
	return e.paperSpace
}

// SetPaperSpace sets whether the entity is in paper space (code 67).
func (e *entity) SetPaperSpace(p bool) {
	e.paperSpace = p
}

// XData returns the extended data (XDATA, group codes 1001 and after) of
// the entity as group code/value pairs.
func (e *entity) XData() [][2]string {
	return e.xdata
}

// SetXData sets the extended data (XDATA). It must start with 1001
// (registered application name); the application is added to the APPID
// table when the drawing is written.
func (e *entity) SetXData(x [][2]string) {
	e.xdata = x
}

func (e *entity) writeXData(f format.Formatter) {
	writeRaw(f, e.xdata)
}

// XDataHolder is implemented by all entities in this package.
type XDataHolder interface {
	XData() [][2]string
	SetXData([][2]string)
}

// FormatEntity writes an entity followed by its extended data.
// POLYLINE and INSERT write their extended data themselves, before their
// VERTEX / ATTRIB entities.
func FormatEntity(f format.Formatter, e Entity) {
	e.Format(f)
	switch e.(type) {
	case *Polyline, *Insert:
		return
	}
	if x, ok := e.(interface{ writeXData(format.Formatter) }); ok {
		x.writeXData(f)
	}
}

// writeRaw writes group code/value pairs kept as read.
func writeRaw(f format.Formatter, pairs [][2]string) {
	for _, p := range pairs {
		if c, err := strconv.Atoi(p[0]); err == nil {
			f.WriteString(c, p[1])
		}
	}
}

// Visibility is implemented by all entities in this package (code 60).
type Visibility interface {
	IsVisible() bool
	SetVisible(bool)
}

// IsVisible reports whether the entity is visible (code 60 is 0 or omitted).
func (e *entity) IsVisible() bool {
	return !e.invisible
}

// SetVisible sets the visibility of the entity (code 60).
func (e *entity) SetVisible(v bool) {
	e.invisible = !v
}

// writeExtrusion writes the extrusion direction (210, 220, 230) if it is
// not the default (0, 0, 1).
func writeExtrusion(f format.Formatter, e []float64) {
	if len(e) < 3 || (e[0] == 0 && e[1] == 0 && e[2] == 1) {
		return
	}
	for i := 0; i < 3; i++ {
		f.WriteFloat(200+(i+1)*10, e[i])
	}
}

// Ltscale returns the linetype scale (code 48).
func (e *entity) Ltscale() float64 {
	return e.ltscale
}

// SourceHandle returns the handle (code 5) of the entity in the file it
// was read from, or "" for new entities. Handles are renumbered when a
// drawing is written; this is for resolving references while reading.
func (e *entity) SourceHandle() string {
	return e.srcHandle
}

// SetSourceHandle sets the handle of the entity in the file it was read from.
func (e *entity) SetSourceHandle(h string) {
	e.srcHandle = h
}

// ExtensionGroups returns the 102 groups ({ACAD_REACTORS ...},
// {ACAD_XDICTIONARY ...}) of the entity, as read.
func (e *entity) ExtensionGroups() [][2]string {
	return e.groups
}

// SetExtensionGroups sets the 102 groups of the entity (as read). The
// reactor to the group of the entity is added when writing.
func (e *entity) SetExtensionGroups(g [][2]string) {
	e.groups = g
}

// writeGroups102 writes the 102 groups: those read, and the reactor to
// the group (GROUP) of the entity.
func (e *entity) writeGroups102(f format.Formatter) {
	groupHandle := ""
	if e.blockRecord != nil {
		groupHandle = handle.Format(e.blockRecord.Handle())
	}
	wroteReactors := false
	inReactors := false
	for _, p := range e.groups {
		if p[0] == "102" && strings.HasPrefix(p[1], "{") {
			inReactors = strings.EqualFold(p[1], "{ACAD_REACTORS")
			wroteReactors = wroteReactors || inReactors
		} else if p[0] == "102" && inReactors {
			if groupHandle != "" {
				f.WriteString(330, groupHandle)
			}
			inReactors = false
		} else if inReactors && p[0] == "330" && strings.EqualFold(strings.TrimSpace(p[1]), groupHandle) {
			groupHandle = "" // already a reactor
		}
		writeRaw(f, [][2]string{p})
	}
	if !wroteReactors && groupHandle != "" {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteString(330, groupHandle)
		f.WriteString(102, "}")
	}
}

// TransparencyByBlock is the value of code 440 for BYBLOCK.
const TransparencyByBlock = color.TransparencyByBlock

// TrueColorer is implemented by all entities in this package: true color
// (code 420, AC1018 or later), color name (430, AC1018 or later) and
// transparency (440, AC1018 or later). They are not written for earlier
// versions.
type TrueColorer interface {
	TrueColor() (r, g, b uint8, ok bool)
	SetTrueColor(r, g, b uint8)
	ClearTrueColor()
	ColorName() string
	SetColorName(string)
	Transparency() (t float64, ok bool)
	SetTransparency(t float64)
	SetTransparencyByBlock()
	ClearTransparency()
}

// TrueColor returns the true color (code 420); ok is false if not set.
func (e *entity) TrueColor() (r, g, b uint8, ok bool) {
	if e.trueColor < 0 {
		return 0, 0, 0, false
	}
	r, g, b = color.SplitRGB(e.trueColor)
	return r, g, b, true
}

// SetTrueColor sets the true color (code 420). The color number (62) is
// set to the nearest one, for applications without true color.
func (e *entity) SetTrueColor(r, g, b uint8) {
	e.trueColor = color.RGB(r, g, b)
	e.color = int(color.Nearest(r, g, b))
}

// ClearTrueColor removes the true color and the color name.
func (e *entity) ClearTrueColor() {
	e.trueColor = -1
	e.colorName = ""
}

// ColorName returns the color name (code 430, "book$name").
func (e *entity) ColorName() string {
	return e.colorName
}

// SetColorName sets the color name (code 430, "book$name") of a color
// book color; set the true color as well.
func (e *entity) SetColorName(n string) {
	e.colorName = n
}

// Transparency returns the transparency (0: opaque, 1: fully transparent);
// ok is false if it is BYLAYER (not set) or BYBLOCK.
func (e *entity) Transparency() (float64, bool) {
	return DecodeTransparency(e.transparency)
}

// SetTransparency sets the transparency (0: opaque, 1: fully transparent).
func (e *entity) SetTransparency(t float64) {
	e.transparency = EncodeTransparency(t)
}

// SetTransparencyByBlock sets the transparency to BYBLOCK.
func (e *entity) SetTransparencyByBlock() {
	e.transparency = TransparencyByBlock
}

// ClearTransparency sets the transparency to BYLAYER.
func (e *entity) ClearTransparency() {
	e.transparency = 0
}

func (e *entity) setTransparencyRaw(v int) {
	e.transparency = v
}

func (e *entity) setTrueColorRaw(v int) {
	e.trueColor = v & 0xFFFFFF
}

// EncodeTransparency returns the value of code 440 for a transparency
// (see color.EncodeTransparency).
func EncodeTransparency(t float64) int { return color.EncodeTransparency(t) }

// DecodeTransparency returns the transparency of a value of code 440
// (see color.DecodeTransparency).
func DecodeTransparency(v int) (float64, bool) { return color.DecodeTransparency(v) }

// SetRawColorValues sets the true color (420), color name (430) and
// transparency (440) values read from a file; c < 0, name "" and t == 0
// are ignored. The color number (62) is not changed.
func SetRawColorValues(e Entity, c int, name string, t int) {
	type raw interface {
		setTrueColorRaw(int)
		setTransparencyRaw(int)
		SetColorName(string)
	}
	if r, ok := e.(raw); ok {
		if c >= 0 {
			r.setTrueColorRaw(c)
		}
		if name != "" {
			r.SetColorName(name)
		}
		if t != 0 {
			r.setTransparencyRaw(t)
		}
	}
}

var _ TrueColorer = &entity{}

// TransparencyValue returns the raw value of code 440 (0: BYLAYER,
// TransparencyByBlock, or 0x020000AA with alpha AA).
func (e *entity) TransparencyValue() int {
	return e.transparency
}
