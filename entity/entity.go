package entity

import (
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
	Type        EntityType      // 0
	handle      int             // 5
	blockRecord handle.Handler  // 102 330
	owner       handle.Handler  // 330
	layer       *table.Layer    // 8
	lineType    *table.LineType // 6: nil = BYLAYER
	color       int             // 62: 0 = BYBLOCK, 1-255, 256 = BYLAYER
	lineWeight  int             // 370: -1 = BYLAYER, -2 = BYBLOCK, -3 = DEFAULT, 0-211 (1/100 mm)
	ltscale     float64         // 48
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
		lineWeight:  table.LineWeightByLayer,
		ltscale:     1.0,
	}
	return e
}

// Format writes data to formatter.
func (e *entity) Format(f format.Formatter) {
	f.WriteString(0, EntityTypeString(e.Type))
	f.WriteHex(5, e.handle)
	if e.blockRecord != nil {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteHex(330, e.blockRecord.Handle())
		f.WriteString(102, "}")
	}
	if e.owner != nil {
		f.WriteHex(330, e.owner.Handle())
	}
	f.WriteString(100, "AcDbEntity")
	f.WriteString(8, e.layer.Name())
	// 62 is omitted if BYLAYER (DXF reference, common group codes for entities)
	// 6, 62 and 370 are omitted if BYLAYER
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
	e.handle = *v
	(*v)++
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
