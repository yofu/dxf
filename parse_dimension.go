package dxf

import (
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// pointTarget returns the coordinate addressed by a point group code
// (10/20/30 style, and 210/220/230) in points keyed by the X code.
func pointTarget(code string, points map[string][]float64) (*float64, bool) {
	if len(code) < 2 {
		return nil, false
	}
	var base string
	var idx int
	switch code[0] {
	case '1':
		base, idx = code, 0
	case '2':
		base, idx = "1"+code[1:], 1
	case '3':
		base, idx = "1"+code[1:], 2
	default:
		return nil, false
	}
	if code == "210" || code == "220" || code == "230" {
		base, idx = "210", int(code[1]-'1')
	}
	p, ok := points[base]
	if !ok {
		return nil, false
	}
	return &p[idx], true
}

// ParseDimension parses DIMENSION entities.
func ParseDimension(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	dim := entity.NewDimension(entity.DIM_LINEAR)
	dim.Flags = 0
	clone := []float64{0, 0, 0}
	points := map[string][]float64{
		"10": dim.DefPoint, "11": dim.TextMid, "12": clone, "13": dim.DefPoint2, "14": dim.DefPoint3,
		"15": dim.DefPoint4, "16": dim.DefPoint5, "210": dim.Extrusion,
	}
	var err error
	for _, dt := range data {
		switch dt[0] {
		case "8":
			if layer, lerr := d.Layer(dt[1], false); lerr == nil {
				dim.SetLayer(layer)
			}
		case "48":
			err = setFloat(dt, func(val float64) { dim.SetLtscale(val) })
		case "2":
			dim.BlockName = dt[1]
		case "1":
			dim.Text = dt[1]
		case "3":
			dim.Style = dt[1]
		case "70":
			err = setInt(dt, func(val int) {
				dim.Type = val & 7
				dim.Flags = val &^ 7
			})
		case "71":
			err = setInt(dt, func(val int) { dim.Attachment = val })
		case "42":
			err = setFloat(dt, func(val float64) { dim.Measurement = val })
		case "53":
			err = setFloat(dt, func(val float64) { dim.TextRotation = val })
		case "50":
			err = setFloat(dt, func(val float64) { dim.Angle = val })
		case "52":
			err = setFloat(dt, func(val float64) { dim.Oblique = val })
		case "40":
			err = setFloat(dt, func(val float64) { dim.LeaderLength = val })
		case "51":
			err = setFloat(dt, func(val float64) { dim.HorizontalDirection = val })
		case "12", "22", "32":
			dim.ClonePoint = clone
			if target, ok := pointTarget(dt[0], points); ok {
				err = setFloat(dt, func(val float64) { *target = val })
			}
		default:
			if target, ok := pointTarget(dt[0], points); ok {
				err = setFloat(dt, func(val float64) { *target = val })
			}
		}
		if err != nil {
			return dim, err
		}
	}
	return dim, nil
}
