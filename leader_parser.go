package dxf

import (
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// ParseLeader parses LEADER entities.
// The annotation (340) is resolved by ParseEntities from the handles of
// the entities in the same section or block.
func ParseLeader(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	l := entity.NewLeader()
	l.Arrow = false
	l.AnnotationType = entity.LEADER_MTEXT
	var err error
	for _, dt := range data {
		switch dt[0] {
		case "8":
			if layer, lerr := d.Layer(dt[1], false); lerr == nil {
				l.SetLayer(layer)
			}
		case "48":
			err = setFloat(dt, func(val float64) { l.SetLtscale(val) })
		case "3":
			l.Style = dt[1]
		case "71":
			err = setInt(dt, func(val int) { l.Arrow = val == 1 })
		case "72":
			err = setInt(dt, func(val int) { l.Spline = val == 1 })
		case "73":
			err = setInt(dt, func(val int) { l.AnnotationType = val })
		case "74":
			err = setInt(dt, func(val int) { l.HooklineDirection = val })
		case "75":
			err = setInt(dt, func(val int) { l.Hookline = val == 1 })
		case "40":
			err = setFloat(dt, func(val float64) { l.TextHeight = val })
		case "41":
			err = setFloat(dt, func(val float64) { l.TextWidth = val })
		case "10":
			l.Vertices = append(l.Vertices, []float64{0, 0, 0})
			err = setFloat(dt, func(val float64) { l.Vertices[len(l.Vertices)-1][0] = val })
		case "20", "30":
			if len(l.Vertices) > 0 {
				k := 1
				if dt[0] == "30" {
					k = 2
				}
				err = setFloat(dt, func(val float64) { l.Vertices[len(l.Vertices)-1][k] = val })
			}
		case "210", "220", "230":
			k := int(dt[0][1] - '1')
			err = setFloat(dt, func(val float64) { l.Extrusion[k] = val })
		case "211", "221", "231":
			k := int(dt[0][1] - '1')
			err = setFloat(dt, func(val float64) { l.HorizontalDirection[k] = val })
		}
		if err != nil {
			return l, err
		}
	}
	return l, nil
}

// value returns the value of the first group code in data.
func value(data [][2]string, code string) (string, bool) {
	for _, dt := range data {
		if dt[0] == code {
			return dt[1], true
		}
	}
	return "", false
}
