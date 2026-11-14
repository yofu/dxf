package entity

import (
	"math"
	"strings"

	"github.com/yofu/dxf/format"
)

// Insert represents INSERT Entity (block reference).
type Insert struct {
	*entity
	BlockName     string       // 2
	Coord         []float64    // 10, 20, 30: insertion point
	Scale         []float64    // 41, 42, 43: X, Y, Z scale factors (default 1)
	Rotation      float64      // 50: rotation angle in degrees
	Columns       int          // 70: column count (default 1)
	Rows          int          // 71: row count (default 1)
	ColumnSpacing float64      // 44
	RowSpacing    float64      // 45
	Direction     []float64    // 210, 220, 230: extrusion direction
	Attributes    []*Attribute // ATTRIB entities following the INSERT (66 = 1), ended by SEQEND
	endhandle     int
}

// IsEntity is for Entity interface.
func (i *Insert) IsEntity() bool {
	return true
}

// NewInsert creates a new Insert referring to the block named name.
func NewInsert(name string) *Insert {
	return &Insert{
		entity:    NewEntity(INSERT),
		BlockName: name,
		Coord:     []float64{0.0, 0.0, 0.0},
		Scale:     []float64{1.0, 1.0, 1.0},
		Columns:   1,
		Rows:      1,
		Direction: []float64{0.0, 0.0, 1.0},
	}
}

// Format writes data to formatter.
// Optional group codes are written only when they differ from the defaults.
func (i *Insert) Format(f format.Formatter) {
	i.entity.Format(f)
	f.WriteString(100, "AcDbBlockReference")
	if len(i.Attributes) > 0 {
		f.WriteInt(66, 1)
	}
	f.WriteString(2, i.BlockName)
	for j := 0; j < 3; j++ {
		f.WriteFloat((j+1)*10, i.Coord[j])
	}
	for j := 0; j < 3; j++ {
		if i.Scale[j] != 1.0 {
			f.WriteFloat(41+j, i.Scale[j])
		}
	}
	if i.Rotation != 0.0 {
		f.WriteFloat(50, i.Rotation)
	}
	if i.Columns != 1 {
		f.WriteInt(70, i.Columns)
	}
	if i.Rows != 1 {
		f.WriteInt(71, i.Rows)
	}
	if i.ColumnSpacing != 0.0 {
		f.WriteFloat(44, i.ColumnSpacing)
	}
	if i.RowSpacing != 0.0 {
		f.WriteFloat(45, i.RowSpacing)
	}
	if i.Direction[0] != 0.0 || i.Direction[1] != 0.0 || i.Direction[2] != 1.0 {
		for j := 0; j < 3; j++ {
			f.WriteFloat(200+(j+1)*10, i.Direction[j])
		}
	}
	if len(i.Attributes) > 0 {
		for _, a := range i.Attributes {
			a.Format(f)
		}
		f.WriteString(0, "SEQEND")
		f.WriteHex(5, i.endhandle)
		f.WriteHex(330, i.Handle()) // owner: INSERT
		f.WriteString(100, "AcDbEntity")
		f.WriteString(8, i.Layer().Name())
	}
}

// String outputs data using default formatter.
func (i *Insert) String() string {
	f := format.NewASCII()
	return i.FormatString(f)
}

// FormatString outputs data using given formatter.
func (i *Insert) FormatString(f format.Formatter) string {
	i.Format(f)
	return f.Output()
}

// BBox returns the insertion point as bounding box,
// since the block definition is not known by the entity.
// Drawing.SetExt computes the extents including the block contents.
func (i *Insert) BBox() ([]float64, []float64) {
	return []float64{i.Coord[0], i.Coord[1], i.Coord[2]}, []float64{i.Coord[0], i.Coord[1], i.Coord[2]}
}

// Transform maps a point p in block coordinates (relative to the block base
// point base) to WCS for the array element at column c and row r:
// scale, rotate around Z, then translate to the insertion point.
// The extrusion direction is not taken into account.
func (i *Insert) Transform(p, base []float64, c, r int) []float64 {
	x := (p[0] - base[0]) * i.Scale[0]
	y := (p[1] - base[1]) * i.Scale[1]
	z := (p[2] - base[2]) * i.Scale[2]
	x += float64(c) * i.ColumnSpacing
	y += float64(r) * i.RowSpacing
	rad := i.Rotation * math.Pi / 180.0
	sin, cos := math.Sin(rad), math.Cos(rad)
	return []float64{
		i.Coord[0] + x*cos - y*sin,
		i.Coord[1] + x*sin + y*cos,
		i.Coord[2] + z,
	}
}

// SetHandle sets handles to the INSERT, its attributes and SEQEND.
func (i *Insert) SetHandle(v *int) {
	i.entity.SetHandle(v)
	if len(i.Attributes) == 0 {
		return
	}
	for _, a := range i.Attributes {
		a.SetOwner(i)
		a.SetHandle(v)
	}
	i.endhandle = *v
	(*v)++
}

// AddAttribute appends an attribute (ATTRIB) to the INSERT.
func (i *Insert) AddAttribute(a *Attribute) {
	a.SetOwner(i)
	i.Attributes = append(i.Attributes, a)
}

// Attribute returns the attribute with tag (case-insensitive), or nil.
func (i *Insert) Attribute(tag string) *Attribute {
	for _, a := range i.Attributes {
		if strings.EqualFold(a.Tag, tag) {
			return a
		}
	}
	return nil
}

// SetAttribute sets the value of the attribute with tag.
// It returns false if there is no such attribute.
func (i *Insert) SetAttribute(tag, value string) bool {
	a := i.Attribute(tag)
	if a == nil {
		return false
	}
	a.Value = value
	return true
}
