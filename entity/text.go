package entity

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/table"
)

// Text Anchor
const (
	LEFT_BASE = iota
	CENTER_BASE
	RIGHT_BASE
	LEFT_BOTTOM
	CENTER_BOTTOM
	RIGHT_BOTTOM
	LEFT_CENTER
	CENTER_CENTER
	RIGHT_CENTER
	LEFT_TOP
	CENTER_TOP
	RIGHT_TOP
)

// Text represents TEXT Entity.
type Text struct {
	*entity
	Coord1         []float64    // 10, 20, 30
	Coord2         []float64    // 11, 21, 31
	Height         float64      // 40
	Rotation       float64      // 50
	WidthFactor    float64      // 41
	ObliqueAngle   float64      // 51
	Value          string       // 1
	Style          *table.Style // 7
	GenFlag        int          // 71
	HorizontalFlag int          // 72
	VerticalFlag   int          // 73
}

// IsEntity is for Entity interface.
func (t *Text) IsEntity() bool {
	return true
}

// NewText creates a new Text.
func NewText() *Text {
	t := &Text{
		entity:         NewEntity(TEXT),
		Coord1:         []float64{0.0, 0.0, 0.0},
		Coord2:         []float64{0.0, 0.0, 0.0},
		Height:         1.0,
		Value:          "",
		Style:          table.ST_STANDARD,
		GenFlag:        0,
		HorizontalFlag: 0,
		VerticalFlag:   0,
	}
	return t
}

// Format writes data to formatter.
func (t *Text) Format(f format.Formatter) {
	t.entity.Format(f)
	f.WriteString(100, "AcDbText")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, t.Coord1[i])
	}
	f.WriteFloat(40, t.Height)
	f.WriteFloat(50, t.Rotation)
	f.WriteFloat(41, t.WidthFactor)
	f.WriteFloat(51, t.ObliqueAngle)
	f.WriteString(1, t.Value)
	f.WriteString(7, t.Style.Name())
	if t.GenFlag != 0 {
		f.WriteInt(71, t.GenFlag)
	}
	if t.HorizontalFlag != 0 {
		f.WriteInt(72, t.HorizontalFlag)
		if t.VerticalFlag != 0 {
			for i := 0; i < 3; i++ {
				f.WriteFloat((i+1)*10+1, t.Coord1[i])
			}
		}
	}
	f.WriteString(100, "AcDbText")
	if t.VerticalFlag != 0 {
		f.WriteInt(73, t.VerticalFlag)
	}
}

// String outputs data using default formatter.
func (t *Text) String() string {
	f := format.NewASCII()
	return t.FormatString(f)
}

// FormatString outputs data using given formatter.
func (t *Text) FormatString(f format.Formatter) string {
	t.Format(f)
	return f.Output()
}

func (t *Text) togglegenflag(val int) {
	if t.GenFlag&val != 0 {
		t.GenFlag &= ^val
	} else {
		t.GenFlag |= val
	}
}

// FlipHorizontal flips Text horizontally.
func (t *Text) FlipHorizontal() {
	t.togglegenflag(2)
}

// FlipVertical flips Text vertically.
func (t *Text) FlipVertical() {
	t.togglegenflag(4)
}

// Anchor sets anchor point flags.
func (t *Text) Anchor(pos int) {
	switch pos {
	case LEFT_BASE:
		t.HorizontalFlag = 0
		t.VerticalFlag = 0
	case CENTER_BASE:
		t.HorizontalFlag = 1
		t.VerticalFlag = 0
	case RIGHT_BASE:
		t.HorizontalFlag = 2
		t.VerticalFlag = 0
	case LEFT_BOTTOM:
		t.HorizontalFlag = 0
		t.VerticalFlag = 1
	case CENTER_BOTTOM:
		t.HorizontalFlag = 1
		t.VerticalFlag = 1
	case RIGHT_BOTTOM:
		t.HorizontalFlag = 2
		t.VerticalFlag = 1
	case LEFT_CENTER:
		t.HorizontalFlag = 0
		t.VerticalFlag = 2
	case CENTER_CENTER:
		t.HorizontalFlag = 1
		t.VerticalFlag = 2
	case RIGHT_CENTER:
		t.HorizontalFlag = 2
		t.VerticalFlag = 2
	case LEFT_TOP:
		t.HorizontalFlag = 0
		t.VerticalFlag = 3
	case CENTER_TOP:
		t.HorizontalFlag = 1
		t.VerticalFlag = 3
	case RIGHT_TOP:
		t.HorizontalFlag = 2
		t.VerticalFlag = 3
	}
}

func (t *Text) BBox() ([]float64, []float64) {
	// TODO: text length, anchor point
	mins := []float64{t.Coord1[0], t.Coord1[1], t.Coord1[2]}
	maxs := []float64{t.Coord1[0], t.Coord1[1] + t.Height, t.Coord1[2]}
	return mins, maxs
}
