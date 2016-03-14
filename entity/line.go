package entity

import (
	"github.com/yofu/dxf/format"
)

// Line represents LINE Entity.
type Line struct {
	*entity
	Start []float64 // 10, 20, 30
	End   []float64 // 11, 21, 31
}

// IsEntity is for Entity interface.
func (l *Line) IsEntity() bool {
	return true
}

// NewLine creates a new Line.
func NewLine() *Line {
	l := &Line{
		entity: NewEntity(LINE),
		Start:  []float64{0.0, 0.0, 0.0},
		End:    []float64{0.0, 0.0, 0.0},
	}
	return l
}

// Format writes data to formatter.
func (l *Line) Format(f format.Formatter) {
	l.entity.Format(f)
	f.WriteString(100, "AcDbLine")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, l.Start[i])
	}
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10+1, l.End[i])
	}
}

// String outputs data using default formatter.
func (l *Line) String() string {
	f := format.NewASCII()
	return l.FormatString(f)
}

// FormatString outputs data using given formatter.
func (l *Line) FormatString(f format.Formatter) string {
	l.Format(f)
	return f.Output()
}

func (l *Line) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)
	for i := 0; i < 3; i++ {
		if l.Start[i] <= l.End[i] {
			mins[i] = l.Start[i]
			maxs[i] = l.End[i]
		} else {
			mins[i] = l.End[i]
			maxs[i] = l.Start[i]
		}
	}
	return mins, maxs
}
