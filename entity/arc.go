package entity

import (
	"math"

	"github.com/yofu/dxf/format"
)

type Arc struct {
	*Circle
	Angle []float64 // 50, 51 (Degree)
}

// IsEntity is for Entity interface.
func (a *Arc) IsEntity() bool {
	return true
}

// NewArc creates a new Arc.
func NewArc(c *Circle) *Arc {
	if c == nil {
		c = NewCircle()
	}
	a := &Arc{
		Circle: c,
		Angle:  []float64{0.0, 180.0},
	}
	a.SetEntityType(ARC)
	return a
}

// Format writes data to formatter.
func (a *Arc) Format(f format.Formatter) {
	a.Circle.Format(f)
	f.WriteString(100, "AcDbArc")
	for i := 0; i < 2; i++ {
		f.WriteFloat(50+i, a.Angle[i])
	}
}

// String write out the String representation
func (a *Arc) String() string {
	f := format.NewASCII()
	a.Format(f)
	return f.Output()
}

// BBox returns the bounding box of the arc (in the plane of the arc,
// assuming the extrusion direction is the Z axis).
func (a *Arc) BBox() ([]float64, []float64) {
	start := a.Angle[0]
	end := a.Angle[1]
	for end <= start {
		end += 360
	}
	angles := []float64{start, end}
	for q := math.Ceil(start/90) * 90; q < end; q += 90 {
		angles = append(angles, q) // the arc passes the quadrant points
	}
	mins := []float64{math.Inf(1), math.Inf(1), a.Center[2]}
	maxs := []float64{math.Inf(-1), math.Inf(-1), a.Center[2]}
	for _, deg := range angles {
		rad := deg * math.Pi / 180
		x := a.Center[0] + a.Radius*math.Cos(rad)
		y := a.Center[1] + a.Radius*math.Sin(rad)
		mins[0], maxs[0] = math.Min(mins[0], x), math.Max(maxs[0], x)
		mins[1], maxs[1] = math.Min(mins[1], y), math.Max(maxs[1], y)
	}
	return extrude(mins, maxs, a.Thickness)
}
