package entity

import (
	"math"
	"unicode/utf8"

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
	Thickness      float64      // 39
	Extrusion      []float64    // 210, 220, 230 (nil: 0, 0, 1)
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
		Extrusion:      []float64{0.0, 0.0, 1.0},
		Height:         1.0,
		WidthFactor:    1.0,
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
	writeThickness(f, t.Thickness)
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
	}
	// The alignment point (11, 21, 31) is required unless the justification
	// is baseline/left (both 72 and 73 are 0).
	if t.HorizontalFlag != 0 || t.VerticalFlag != 0 {
		p := t.Coord2
		if len(p) < 3 {
			p = t.Coord1
		}
		for i := 0; i < 3; i++ {
			f.WriteFloat((i+1)*10+1, p[i])
		}
	}
	writeExtrusion(f, t.Extrusion)
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

// The width is estimated as the number of characters times the height and
// the width factor; the justification (72, 73) and rotation are applied.
func (t *Text) BBox() ([]float64, []float64) {
	h := t.Height
	wf := t.WidthFactor
	if wf <= 0 {
		wf = 1
	}
	w := float64(utf8.RuneCountInString(t.Value)) * h * wf
	ref := t.Coord1
	justified := t.HorizontalFlag != 0 || t.VerticalFlag != 0
	if justified && len(t.Coord2) >= 3 {
		ref = t.Coord2
	}
	x0 := 0.0
	switch t.HorizontalFlag {
	case 1, 4: // center, middle
		x0 = -w / 2
	case 2: // right
		x0 = -w
	case 3, 5: // aligned, fit: between the two points
		x0 = 0
		ref = t.Coord1
		w = math.Hypot(t.Coord2[0]-t.Coord1[0], t.Coord2[1]-t.Coord1[1])
	}
	y0 := 0.0
	switch {
	case t.HorizontalFlag == 4 || t.VerticalFlag == 2: // middle
		y0 = -h / 2
	case t.VerticalFlag == 3: // top
		y0 = -h
	}
	rad := t.Rotation * math.Pi / 180
	sin, cos := math.Sin(rad), math.Cos(rad)
	mins := []float64{math.Inf(1), math.Inf(1), ref[2]}
	maxs := []float64{math.Inf(-1), math.Inf(-1), ref[2]}
	for _, p := range [][2]float64{{x0, y0}, {x0 + w, y0}, {x0, y0 + h}, {x0 + w, y0 + h}} {
		x := ref[0] + p[0]*cos - p[1]*sin
		y := ref[1] + p[0]*sin + p[1]*cos
		mins[0], maxs[0] = math.Min(mins[0], x), math.Max(maxs[0], x)
		mins[1], maxs[1] = math.Min(mins[1], y), math.Max(maxs[1], y)
	}
	return extrude(mins, maxs, t.Thickness)
}
