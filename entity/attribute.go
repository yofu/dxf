package entity

import (
	"github.com/yofu/dxf/format"
)

// Attribute flags (code 70) of ATTDEF and ATTRIB.
const (
	ATTRIBUTE_INVISIBLE = 1
	ATTRIBUTE_CONSTANT  = 2 // ATTDEF only: the value is fixed, no ATTRIB is created
	ATTRIBUTE_VERIFY    = 4
	ATTRIBUTE_PRESET    = 8
)

// AttributeDefinition represents ATTDEF Entity (attribute definition in a block).
// The text properties (position, height, style, justification) are those of
// the embedded Text; Text.Value is the default value.
type AttributeDefinition struct {
	*Text
	Tag         string // 2
	Prompt      string // 3
	Flags       int    // 70: ATTRIBUTE_*
	FieldLength int    // 73
}

// NewAttributeDefinition creates a new AttributeDefinition.
func NewAttributeDefinition(tag, prompt, defaultValue string) *AttributeDefinition {
	t := NewText()
	t.SetEntityType(ATTDEF)
	t.Value = defaultValue
	return &AttributeDefinition{Text: t, Tag: tag, Prompt: prompt}
}

// IsConstant reports whether the attribute is constant.
func (a *AttributeDefinition) IsConstant() bool {
	return a.Flags&ATTRIBUTE_CONSTANT != 0
}

// Format writes data to formatter.
func (a *AttributeDefinition) Format(f format.Formatter) {
	t := a.Text
	t.entity.Format(f)
	f.WriteString(100, "AcDbText")
	writeThickness(f, t.Thickness)
	writePoint(f, 10, t.Coord1)
	f.WriteFloat(40, t.Height)
	f.WriteString(1, t.Value)
	writeTextProperties(f, t)
	f.WriteString(100, "AcDbAttributeDefinition")
	f.WriteString(3, a.Prompt)
	f.WriteString(2, a.Tag)
	f.WriteInt(70, a.Flags)
	if a.FieldLength != 0 {
		f.WriteInt(73, a.FieldLength)
	}
	if t.VerticalFlag != 0 {
		f.WriteInt(74, t.VerticalFlag)
	}
}

// String outputs data using default formatter.
func (a *AttributeDefinition) String() string {
	f := format.NewASCII()
	a.Format(f)
	return f.Output()
}

// Attribute represents ATTRIB Entity (attribute value of an INSERT).
// The text properties are those of the embedded Text; Text.Value is the value.
type Attribute struct {
	*Text
	Tag         string // 2
	Flags       int    // 70: ATTRIBUTE_*
	FieldLength int    // 73
}

// NewAttribute creates a new Attribute.
func NewAttribute(tag, value string) *Attribute {
	t := NewText()
	t.SetEntityType(ATTRIB)
	t.Value = value
	return &Attribute{Text: t, Tag: tag}
}

// Format writes data to formatter.
func (a *Attribute) Format(f format.Formatter) {
	t := a.Text
	t.entity.Format(f)
	f.WriteString(100, "AcDbText")
	writeThickness(f, t.Thickness)
	writePoint(f, 10, t.Coord1)
	f.WriteFloat(40, t.Height)
	f.WriteString(1, t.Value)
	f.WriteString(100, "AcDbAttribute")
	f.WriteString(2, a.Tag)
	f.WriteInt(70, a.Flags)
	if a.FieldLength != 0 {
		f.WriteInt(73, a.FieldLength)
	}
	writeTextProperties(f, t)
	if t.VerticalFlag != 0 {
		f.WriteInt(74, t.VerticalFlag)
	}
}

// String outputs data using default formatter.
func (a *Attribute) String() string {
	f := format.NewASCII()
	a.Format(f)
	return f.Output()
}

// writeTextProperties writes 50, 41, 51, 7, 71, 72 and the alignment point
// (11, 21, 31) of attributes. The vertical justification is written as 74.
func writeTextProperties(f format.Formatter, t *Text) {
	if t.Rotation != 0 {
		f.WriteFloat(50, t.Rotation)
	}
	f.WriteFloat(41, t.WidthFactor)
	if t.ObliqueAngle != 0 {
		f.WriteFloat(51, t.ObliqueAngle)
	}
	f.WriteString(7, t.Style.Name())
	if t.GenFlag != 0 {
		f.WriteInt(71, t.GenFlag)
	}
	if t.HorizontalFlag != 0 {
		f.WriteInt(72, t.HorizontalFlag)
	}
	if t.HorizontalFlag != 0 || t.VerticalFlag != 0 {
		p := t.Coord2
		if len(p) < 3 {
			p = t.Coord1
		}
		writePoint(f, 11, p)
	}
}
