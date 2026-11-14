package entity

import (
	"math"
	"strconv"

	"github.com/yofu/dxf/format"
)

// Unknown represents an entity of a type not supported by this package,
// read from a file. Its data (except the common entity properties, handles
// and references) is kept as group code/value pairs and written back.
type Unknown struct {
	*entity
	Data [][2]string // group codes after the common entity properties
}

// IsEntity is for Entity interface.
func (u *Unknown) IsEntity() bool {
	return true
}

// NewUnknown creates a new Unknown entity of type name (e.g. "XLINE").
func NewUnknown(name string) *Unknown {
	e := NewEntity(UNKNOWN)
	e.typeName = name
	return &Unknown{entity: e}
}

// TypeName returns the entity type name.
func (u *Unknown) TypeName() string {
	return u.typeName
}

// Format writes data to formatter.
func (u *Unknown) Format(f format.Formatter) {
	u.entity.Format(f)
	writeRaw(f, u.Data)
}

// String outputs data using default formatter.
func (u *Unknown) String() string {
	f := format.NewASCII()
	u.Format(f)
	return f.Output()
}

// BBox returns the bounding box of the points (10-18, 20-28, 30-38) in
// Data. Without points, it returns an empty box (+Inf, -Inf), which does
// not change extents.
func (u *Unknown) BBox() ([]float64, []float64) {
	mins := []float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxs := []float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, p := range u.Data {
		c, err := strconv.Atoi(p[0])
		if err != nil || c < 10 || c > 38 || c%10 > 8 {
			continue
		}
		v, err := strconv.ParseFloat(p[1], 64)
		if err != nil {
			continue
		}
		i := c/10 - 1
		mins[i] = math.Min(mins[i], v)
		maxs[i] = math.Max(maxs[i], v)
	}
	for i := 0; i < 3; i++ { // 2D points
		if math.IsInf(mins[i], 1) && !math.IsInf(mins[0], 1) {
			mins[i], maxs[i] = 0, 0
		}
	}
	return mins, maxs
}
