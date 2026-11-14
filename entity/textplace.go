package entity

import (
	"fmt"
	"strings"
)

// Text alignments for Text.SetPlacement (horizontal 72, vertical 73).
var textAlignments = map[string][2]int{
	"LEFT": {0, 0}, "CENTER": {1, 0}, "RIGHT": {2, 0}, "ALIGNED": {3, 0}, "MIDDLE": {4, 0}, "FIT": {5, 0},
	"BOTTOM_LEFT": {0, 1}, "BOTTOM_CENTER": {1, 1}, "BOTTOM_RIGHT": {2, 1},
	"MIDDLE_LEFT": {0, 2}, "MIDDLE_CENTER": {1, 2}, "MIDDLE_RIGHT": {2, 2},
	"TOP_LEFT": {0, 3}, "TOP_CENTER": {1, 3}, "TOP_RIGHT": {2, 3},
}

// SetPlacement places the text at p1 with alignment (as ezdxf's
// set_placement): "LEFT" (baseline, default), "CENTER", "RIGHT", "MIDDLE",
// "BOTTOM_LEFT" ... "TOP_RIGHT", or "ALIGNED" and "FIT" between p1 and p2.
// For alignments other than LEFT the alignment point (11) is p1 and the
// insertion point (10) is computed by CAD applications (it is set to p1).
func (t *Text) SetPlacement(p1, p2 []float64, alignment string) error {
	a, ok := textAlignments[strings.ToUpper(alignment)]
	if !ok {
		return fmt.Errorf("unknown text alignment: %s", alignment)
	}
	t.HorizontalFlag, t.VerticalFlag = a[0], a[1]
	p := []float64{coord(p1, 0), coord(p1, 1), coord(p1, 2)}
	t.Coord1 = p
	switch {
	case a == [2]int{0, 0}:
		t.Coord2 = nil
	case a[0] == 3 || a[0] == 5:
		if p2 == nil {
			return fmt.Errorf("alignment %s needs two points", alignment)
		}
		t.Coord2 = []float64{coord(p2, 0), coord(p2, 1), coord(p2, 2)}
	default:
		t.Coord2 = append([]float64{}, p...)
	}
	return nil
}
