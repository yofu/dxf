package entity

import (
	"math"

	"github.com/yofu/dxf/format"
)

// Annotation types of LEADER (code 73).
const (
	LEADER_MTEXT     = 0
	LEADER_TOLERANCE = 1
	LEADER_BLOCK     = 2
	LEADER_NONE      = 3
)

// Leader represents LEADER Entity. The leader line is drawn through
// Vertices, with an arrowhead at the first vertex.
type Leader struct {
	*entity
	Style               string      // 3: dimension style name
	Arrow               bool        // 71: arrowhead
	Spline              bool        // 72: path type (false: straight segments)
	AnnotationType      int         // 73: LEADER_MTEXT, ...
	HooklineDirection   int         // 74: 1 if the hookline is in the same direction as HorizontalDirection
	Hookline            bool        // 75
	TextHeight          float64     // 40
	TextWidth           float64     // 41
	Vertices            [][]float64 // 76, 10, 20, 30
	Annotation          Entity      // 340: annotation (MTEXT, INSERT, ...), nil if none
	Extrusion           []float64   // 210, 220, 230
	HorizontalDirection []float64   // 211, 221, 231
}

// IsEntity is for Entity interface.
func (l *Leader) IsEntity() bool {
	return true
}

// NewLeader creates a new Leader without annotation.
func NewLeader() *Leader {
	return &Leader{
		entity:              NewEntity(LEADER),
		Style:               "Standard",
		Arrow:               true,
		AnnotationType:      LEADER_NONE,
		Extrusion:           []float64{0, 0, 1},
		HorizontalDirection: []float64{1, 0, 0},
	}
}

// Format writes data to formatter.
func (l *Leader) Format(f format.Formatter) {
	l.entity.Format(f)
	f.WriteString(100, "AcDbLeader")
	f.WriteString(3, l.Style)
	f.WriteInt(71, boolInt(l.Arrow))
	f.WriteInt(72, boolInt(l.Spline))
	f.WriteInt(73, l.AnnotationType)
	f.WriteInt(74, l.HooklineDirection)
	f.WriteInt(75, boolInt(l.Hookline))
	f.WriteFloat(40, l.TextHeight)
	f.WriteFloat(41, l.TextWidth)
	f.WriteInt(76, len(l.Vertices))
	for _, v := range l.Vertices {
		p := []float64{0, 0, 0}
		copy(p, v)
		writePoint(f, 10, p)
	}
	if l.Annotation != nil {
		f.WriteHex(340, l.Annotation.Handle())
	}
	writePoint(f, 210, l.Extrusion)
	writePoint(f, 211, l.HorizontalDirection)
}

// String outputs data using default formatter.
func (l *Leader) String() string {
	f := format.NewASCII()
	return l.FormatString(f)
}

// FormatString outputs data using given formatter.
func (l *Leader) FormatString(f format.Formatter) string {
	l.Format(f)
	return f.Output()
}

// BBox returns the bounding box of the vertices.
func (l *Leader) BBox() ([]float64, []float64) {
	mins := []float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxs := []float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, v := range l.Vertices {
		for i := 0; i < 3 && i < len(v); i++ {
			mins[i] = math.Min(mins[i], v[i])
			maxs[i] = math.Max(maxs[i], v[i])
		}
	}
	if len(l.Vertices) == 0 {
		return []float64{0, 0, 0}, []float64{0, 0, 0}
	}
	return mins, maxs
}
