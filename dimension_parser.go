package dxf

import (
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/table"
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
	points := map[string][]float64{
		"10": dim.DefPoint, "11": dim.TextMid, "13": dim.DefPoint2, "14": dim.DefPoint3,
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

// ParseSolid parses SOLID entities.
func ParseSolid(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	s := entity.NewSolid()
	points := map[string][]float64{
		"10": s.Points[0], "11": s.Points[1], "12": s.Points[2], "13": s.Points[3], "210": s.Extrusion,
	}
	fourth := false
	var err error
	for _, dt := range data {
		switch dt[0] {
		case "8":
			if layer, lerr := d.Layer(dt[1], false); lerr == nil {
				s.SetLayer(layer)
			}
		case "48":
			err = setFloat(dt, func(val float64) { s.SetLtscale(val) })
		case "39":
			err = setFloat(dt, func(val float64) { s.Thickness = val })
		default:
			if target, ok := pointTarget(dt[0], points); ok {
				if dt[0] == "13" || dt[0] == "23" {
					fourth = true
				}
				err = setFloat(dt, func(val float64) { *target = val })
			}
		}
		if err != nil {
			return s, err
		}
	}
	if !fourth { // triangle: the fourth point is the same as the third
		copy(s.Points[3], s.Points[2])
	}
	return s, nil
}

// ParseDimStyle parses DIMSTYLE tables.
func ParseDimStyle(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	ds := table.NewDimStyle("")
	name := ""
	var err error
	for _, dt := range data {
		switch dt[0] {
		case "2":
			name = dt[1]
		case "40":
			err = setFloat(dt, func(val float64) { ds.Scale = val })
		case "41":
			err = setFloat(dt, func(val float64) { ds.ArrowSize = val })
		case "42":
			err = setFloat(dt, func(val float64) { ds.ExtOffset = val })
		case "44":
			err = setFloat(dt, func(val float64) { ds.ExtExtend = val })
		case "73":
			err = setInt(dt, func(val int) { ds.TextInsideHorizontal = val != 0 })
		case "74":
			err = setInt(dt, func(val int) { ds.TextOutsideHorizontal = val != 0 })
		case "77":
			err = setInt(dt, func(val int) { ds.TextAbove = val != 0 })
		case "140":
			err = setFloat(dt, func(val float64) { ds.TextHeight = val })
		case "147":
			err = setFloat(dt, func(val float64) { ds.TextGap = val })
		case "271":
			err = setInt(dt, func(val int) { ds.Decimals = val })
		}
		if err != nil {
			return ds, err
		}
	}
	ds.SetName(name)
	ds.Extra = rawExtra(data, codes("2", "70", "40", "41", "42", "44", "73", "74", "77", "140", "147", "271"), false)
	return ds, nil
}
