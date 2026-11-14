package dxf

import (
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// ParseLine parses LINE entities.
func ParseLine(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	l := entity.NewLine()
	var err error
	for _, dt := range data {
		switch dt[0] {
		default:
			continue
		case "8":
			layer, err := d.Layer(dt[1], false)
			if err == nil {
				l.SetLayer(layer)
			}
		case "48":
			err = setFloat(dt, func(val float64) { l.SetLtscale(val) })
		case "10":
			err = setFloat(dt, func(val float64) { l.Start[0] = val })
		case "20":
			err = setFloat(dt, func(val float64) { l.Start[1] = val })
		case "30":
			err = setFloat(dt, func(val float64) { l.Start[2] = val })
		case "11":
			err = setFloat(dt, func(val float64) { l.End[0] = val })
		case "21":
			err = setFloat(dt, func(val float64) { l.End[1] = val })
		case "31":
			err = setFloat(dt, func(val float64) { l.End[2] = val })
		case "210", "220", "230":
			err = setExtrusion(dt, l.Extrusion)
		}
		if err != nil {
			return l, err
		}
	}
	return l, nil
}

// Parse3DFace parses 3DFACE entities.
func Parse3DFace(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	t := entity.New3DFace()
	fourth := 0
	var err error
	for _, dt := range data {
		switch dt[0] {
		default:
			continue
		case "8":
			layer, err := d.Layer(dt[1], false)
			if err == nil {
				t.SetLayer(layer)
			}
		case "48":
			err = setFloat(dt, func(val float64) { t.SetLtscale(val) })
		case "10":
			err = setFloat(dt, func(val float64) { t.Points[0][0] = val })
		case "20":
			err = setFloat(dt, func(val float64) { t.Points[0][1] = val })
		case "30":
			err = setFloat(dt, func(val float64) { t.Points[0][2] = val })
		case "11":
			err = setFloat(dt, func(val float64) { t.Points[1][0] = val })
		case "21":
			err = setFloat(dt, func(val float64) { t.Points[1][1] = val })
		case "31":
			err = setFloat(dt, func(val float64) { t.Points[1][2] = val })
		case "12":
			err = setFloat(dt, func(val float64) { t.Points[2][0] = val })
		case "22":
			err = setFloat(dt, func(val float64) { t.Points[2][1] = val })
		case "32":
			err = setFloat(dt, func(val float64) { t.Points[2][2] = val })
		case "13":
			err = setFloat(dt, func(val float64) {
				t.Points[3][0] = val
				fourth |= 1
			})
		case "23":
			err = setFloat(dt, func(val float64) {
				t.Points[3][1] = val
				fourth |= 2
			})
		case "33":
			err = setFloat(dt, func(val float64) {
				t.Points[3][2] = val
				fourth |= 4
			})
		case "70":
			err = setInt(dt, func(val int) { t.Flag = val })
		}
		if err != nil {
			return t, err
		}
	}
	if fourth != 7 {
		for i := 0; i < 3; i++ {
			t.Points[3][i] = t.Points[2][i]
		}
	}
	return t, nil
}

// ParseCircle parses CIRCLE entities.
func ParseCircle(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	c := entity.NewCircle()
	var err error
	for _, dt := range data {
		switch dt[0] {
		default:
			continue
		case "8":
			layer, err := d.Layer(dt[1], false)
			if err == nil {
				c.SetLayer(layer)
			}
		case "48":
			err = setFloat(dt, func(val float64) { c.SetLtscale(val) })
		case "10":
			err = setFloat(dt, func(val float64) { c.Center[0] = val })
		case "20":
			err = setFloat(dt, func(val float64) { c.Center[1] = val })
		case "30":
			err = setFloat(dt, func(val float64) { c.Center[2] = val })
		case "40":
			err = setFloat(dt, func(val float64) { c.Radius = val })
		case "210":
			err = setFloat(dt, func(val float64) { c.Direction[0] = val })
		case "220":
			err = setFloat(dt, func(val float64) { c.Direction[1] = val })
		case "230":
			err = setFloat(dt, func(val float64) { c.Direction[2] = val })
		}
		if err != nil {
			return c, err
		}
	}
	return c, nil
}

// ParseArc parses ARC entities.
func ParseArc(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	c := entity.NewCircle()
	angle := make([]float64, 2)
	var err error
	for _, dt := range data {
		switch dt[0] {
		default:
			continue
		case "8":
			layer, err := d.Layer(dt[1], false)
			if err == nil {
				c.SetLayer(layer)
			}
		case "48":
			err = setFloat(dt, func(val float64) { c.SetLtscale(val) })
		case "10":
			err = setFloat(dt, func(val float64) { c.Center[0] = val })
		case "20":
			err = setFloat(dt, func(val float64) { c.Center[1] = val })
		case "30":
			err = setFloat(dt, func(val float64) { c.Center[2] = val })
		case "40":
			err = setFloat(dt, func(val float64) { c.Radius = val })
		case "50":
			err = setFloat(dt, func(val float64) { angle[0] = val })
		case "51":
			err = setFloat(dt, func(val float64) { angle[1] = val })
		case "210":
			err = setFloat(dt, func(val float64) { c.Direction[0] = val })
		case "220":
			err = setFloat(dt, func(val float64) { c.Direction[1] = val })
		case "230":
			err = setFloat(dt, func(val float64) { c.Direction[2] = val })
		}
		if err != nil {
			a := entity.NewArc(c)
			for i := 0; i < 2; i++ {
				a.Angle[i] = angle[i]
			}
			return a, err
		}
	}
	a := entity.NewArc(c)
	for i := 0; i < 2; i++ {
		a.Angle[i] = angle[i]
	}
	return a, nil
}

// ParsePoint parses POINT entities.
func ParsePoint(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	p := entity.NewPoint()
	var err error
	for _, dt := range data {
		switch dt[0] {
		default:
			continue
		case "8":
			layer, err := d.Layer(dt[1], false)
			if err == nil {
				p.SetLayer(layer)
			}
		case "48":
			err = setFloat(dt, func(val float64) { p.SetLtscale(val) })
		case "10":
			err = setFloat(dt, func(val float64) { p.Coord[0] = val })
		case "20":
			err = setFloat(dt, func(val float64) { p.Coord[1] = val })
		case "30":
			err = setFloat(dt, func(val float64) { p.Coord[2] = val })
		case "210", "220", "230":
			err = setExtrusion(dt, p.Extrusion)
		case "50":
			err = setFloat(dt, func(val float64) { p.Angle = val })
		}
		if err != nil {
			return p, err
		}
	}
	return p, nil
}

// ParseSpline parses SPLINE entities.
func ParseSpline(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	s := entity.NewSpline()
	var err error
	var controlPoint []float64
	for _, dt := range data {
		switch dt[0] {
		default:
			continue
		case "8":
			layer, err := d.Layer(dt[1], false)
			if err == nil {
				s.SetLayer(layer)
			}
		case "210":
			err = setFloat(dt, func(val float64) { s.Normal[0] = val })
		case "220":
			err = setFloat(dt, func(val float64) { s.Normal[1] = val })
		case "230":
			err = setFloat(dt, func(val float64) { s.Normal[2] = val })
		case "70":
			err = setInt(dt, func(val int) { s.Flag = val })
		case "71":
			err = setInt(dt, func(val int) { s.Degree = val })
		case "72":
			// Node values
		case "73":
			// Control point coordinates
		case "74":
			// Fitting point coordinates
		case "42":
			err = setFloat(dt, func(val float64) { s.Tolerance[0] = val })
		case "43":
			err = setFloat(dt, func(val float64) { s.Tolerance[1] = val })
		case "44":
			err = setFloat(dt, func(val float64) { s.Tolerance[2] = val })
		case "40":
			err = setFloat(dt, func(val float64) { s.Knots = append(s.Knots, val) })
		case "41":
			err = setFloat(dt, func(val float64) { s.Weights = append(s.Weights, val) })
		case "12", "22", "32":
			if s.StartTangent == nil {
				s.StartTangent = []float64{0, 0, 0}
			}
			k := int(dt[0][0] - '1')
			err = setFloat(dt, func(val float64) { s.StartTangent[k] = val })
		case "13", "23", "33":
			if s.EndTangent == nil {
				s.EndTangent = []float64{0, 0, 0}
			}
			k := int(dt[0][0] - '1')
			err = setFloat(dt, func(val float64) { s.EndTangent[k] = val })
		case "10", "20", "30":
			err = setFloat(dt, func(val float64) {
				controlPoint = append(controlPoint, val)
				if len(controlPoint) == 3 {
					s.Controls = append(s.Controls, controlPoint)
					controlPoint = nil
				}
			})
		case "11", "21", "31":
			err = setFloat(dt, func(val float64) {
				controlPoint = append(controlPoint, val)
				if len(controlPoint) == 3 {
					s.Fits = append(s.Fits, controlPoint)
					controlPoint = nil
				}
			})
		}
		if err != nil {
			return s, err
		}
	}
	return s, nil
}

// ParseInsert parses INSERT entities.
// ATTRIB entities following an INSERT are attached to it by parseEntityList.
func ParseInsert(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	ins := entity.NewInsert("")
	var err error
	for _, dt := range data {
		switch dt[0] {
		default:
			continue
		case "2":
			ins.BlockName = dt[1]
		case "8":
			layer, err := d.Layer(dt[1], false)
			if err == nil {
				ins.SetLayer(layer)
			}
		case "48":
			err = setFloat(dt, func(val float64) { ins.SetLtscale(val) })
		case "10":
			err = setFloat(dt, func(val float64) { ins.Coord[0] = val })
		case "20":
			err = setFloat(dt, func(val float64) { ins.Coord[1] = val })
		case "30":
			err = setFloat(dt, func(val float64) { ins.Coord[2] = val })
		case "41":
			err = setFloat(dt, func(val float64) { ins.Scale[0] = val })
		case "42":
			err = setFloat(dt, func(val float64) { ins.Scale[1] = val })
		case "43":
			err = setFloat(dt, func(val float64) { ins.Scale[2] = val })
		case "50":
			err = setFloat(dt, func(val float64) { ins.Rotation = val })
		case "70":
			err = setInt(dt, func(val int) { ins.Columns = val })
		case "71":
			err = setInt(dt, func(val int) { ins.Rows = val })
		case "44":
			err = setFloat(dt, func(val float64) { ins.ColumnSpacing = val })
		case "45":
			err = setFloat(dt, func(val float64) { ins.RowSpacing = val })
		case "210":
			err = setFloat(dt, func(val float64) { ins.Direction[0] = val })
		case "220":
			err = setFloat(dt, func(val float64) { ins.Direction[1] = val })
		case "230":
			err = setFloat(dt, func(val float64) { ins.Direction[2] = val })
		}
		if err != nil {
			return ins, err
		}
	}
	return ins, nil
}

// ParseEllipse parses ELLIPSE entities.
func ParseEllipse(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	e := entity.NewEllipse()
	var err error
	for _, dt := range data {
		var target *float64
		switch dt[0] {
		default:
			continue
		case "8":
			layer, err := d.Layer(dt[1], false)
			if err == nil {
				e.SetLayer(layer)
			}
			continue
		case "48":
			err = setFloat(dt, func(val float64) { e.SetLtscale(val) })
			if err != nil {
				return e, err
			}
			continue
		case "10":
			target = &e.Center[0]
		case "20":
			target = &e.Center[1]
		case "30":
			target = &e.Center[2]
		case "11":
			target = &e.MajorAxis[0]
		case "21":
			target = &e.MajorAxis[1]
		case "31":
			target = &e.MajorAxis[2]
		case "210":
			target = &e.Extrusion[0]
		case "220":
			target = &e.Extrusion[1]
		case "230":
			target = &e.Extrusion[2]
		case "40":
			target = &e.Ratio
		case "41":
			target = &e.Start
		case "42":
			target = &e.End
		}
		if err = setFloat(dt, func(val float64) { *target = val }); err != nil {
			return e, err
		}
	}
	return e, nil
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
