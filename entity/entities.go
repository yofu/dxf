// ENTITIES section
package entity

import (
	"math"

	"github.com/yofu/dxf/format"
)

// Entities represents ENTITIES section.
type Entities []Entity

// New creates a new Entities.
func New() Entities {
	e := make([]Entity, 0)
	return e
}

// Format writes ENTITIES data to formatter.
func (es Entities) Format(f format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "ENTITIES")
	for _, e := range es {
		FormatEntity(f, e)
	}
	f.WriteString(0, "ENDSEC")
}

// Add adds a new entity to ENTITIES section.
func (es Entities) Add(e Entity) Entities {
	es = append(es, e)
	return es
}

// SetHandle sets handles to each entity.
func (es Entities) SetHandle(v *int) {
	for _, e := range es {
		e.SetHandle(v)
	}
}

// bulge computes the center and the radius of the arc between 2 vertices.
//
// bulge is the tangent of 1/4 of the included angle of the arc, negative
// if the arc goes clockwise from start to end (DXF reference, VERTEX/LWPOLYLINE).
func bulge(bulge float64, start []float64, end []float64) ([]float64, float64) {
	if bulge == 0 {
		return nil, 0
	}
	dx := end[0] - start[0]
	dy := end[1] - start[1]
	chord := math.Sqrt(dx*dx + dy*dy)
	if chord == 0 {
		return nil, 0
	}
	// signed distance from the chord midpoint to the center, measured along
	// the left normal of the chord (start -> end).
	offset := chord / 2 * (1 - bulge*bulge) / (2 * bulge)
	radius := chord / 4 * (1 + bulge*bulge) / math.Abs(bulge)
	center := []float64{
		(start[0]+end[0])/2 - offset*dy/chord,
		(start[1]+end[1])/2 + offset*dx/chord,
	}
	return center, radius
}
