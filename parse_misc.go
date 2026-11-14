package dxf

import (
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// splitAt splits data at the subclass marker (100) marker: the codes
// before it and after it (nil if not found).
func splitAt(data [][2]string, marker string) ([][2]string, [][2]string) {
	for i, dt := range data {
		if dt[0] == "100" && strings.EqualFold(strings.TrimSpace(dt[1]), marker) {
			return data[:i], data[i+1:]
		}
	}
	return data, nil
}

// setLayer sets the layer (8) and the linetype scale (48) of e from data.
func setLayerAndLtscale(d *drawing.Drawing, e entity.Entity, dt [2]string) error {
	switch dt[0] {
	case "8":
		if layer, err := d.Layer(dt[1], false); err == nil {
			e.SetLayer(layer)
		}
	case "48":
		return setFloat(dt, func(val float64) { e.SetLtscale(val) })
	}
	return nil
}

// setPoint sets coordinate k of p (k from the code: 1x -> 0, 2x -> 1, 3x -> 2).
func setPoint(dt [2]string, p []float64) error {
	k := int(dt[0][len(dt[0])-2] - '1')
	if len(dt[0]) == 2 {
		k = int(dt[0][0] - '1')
	}
	return setFloat(dt, func(val float64) { p[k] = val })
}

// ParseXLine parses XLINE and RAY entities.
func ParseXLine(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	x := entity.NewXLine(nil, nil)
	if len(data) > 0 && strings.EqualFold(strings.TrimSpace(data[0][1]), "RAY") {
		x.Type = entity.RAY
	}
	for _, dt := range data {
		var err error
		switch dt[0] {
		case "8", "48":
			err = setLayerAndLtscale(d, x, dt)
		case "10", "20", "30":
			err = setPoint(dt, x.Point)
		case "11", "21", "31":
			err = setPoint(dt, x.Direction)
		}
		if err != nil {
			return x, err
		}
	}
	return x, nil
}

// ParseTrace parses TRACE entities (like SOLID).
func ParseTrace(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	e, err := ParseSolid(d, data)
	if s, ok := e.(*entity.Solid); ok {
		s.Type = entity.TRACE
	}
	return e, err
}

// ParseShape parses SHAPE entities.
func ParseShape(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	s := entity.NewShape("", nil, 1)
	for _, dt := range data {
		var err error
		switch dt[0] {
		case "8", "48":
			err = setLayerAndLtscale(d, s, dt)
		case "10", "20", "30":
			err = setPoint(dt, s.Insert)
		case "210", "220", "230":
			err = setPoint(dt, s.Extrusion)
		case "40":
			err = setFloat(dt, func(v float64) { s.Size = v })
		case "2":
			s.Name = dt[1]
		case "39":
			err = setFloat(dt, func(v float64) { s.Thickness = v })
		case "50":
			err = setFloat(dt, func(v float64) { s.Rotation = v })
		case "41":
			err = setFloat(dt, func(v float64) { s.XScale = v })
		case "51":
			err = setFloat(dt, func(v float64) { s.Oblique = v })
		}
		if err != nil {
			return s, err
		}
	}
	return s, nil
}

// ParseTolerance parses TOLERANCE entities.
func ParseTolerance(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	tl := entity.NewTolerance("", nil)
	for _, dt := range data {
		var err error
		switch dt[0] {
		case "8", "48":
			err = setLayerAndLtscale(d, tl, dt)
		case "3":
			tl.Style = dt[1]
		case "10", "20", "30":
			err = setPoint(dt, tl.Insert)
		case "1":
			tl.Content = dt[1]
		case "210", "220", "230":
			err = setPoint(dt, tl.Extrusion)
		case "11", "21", "31":
			err = setPoint(dt, tl.XAxis)
		}
		if err != nil {
			return tl, err
		}
	}
	return tl, nil
}

// ParseLight parses LIGHT entities.
func ParseLight(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	l := entity.NewLight("", entity.LIGHT_POINT, nil)
	_, body := splitAt(data, "AcDbLight")
	for _, dt := range data[:len(data)-len(body)] {
		if err := setLayerAndLtscale(d, l, dt); err != nil {
			return l, err
		}
	}
	for _, dt := range body {
		var err error
		switch dt[0] {
		case "90":
			err = setInt(dt, func(v int) { l.Version = v })
		case "1":
			l.Name = dt[1]
		case "70":
			err = setInt(dt, func(v int) { l.LightType = v })
		case "290":
			err = setInt(dt, func(v int) { l.On = v != 0 })
		case "291":
			err = setInt(dt, func(v int) { l.PlotGlyph = v != 0 })
		case "40":
			err = setFloat(dt, func(v float64) { l.Intensity = v })
		case "10", "20", "30":
			err = setPoint(dt, l.Location)
		case "11", "21", "31":
			err = setPoint(dt, l.Target)
		case "72":
			err = setInt(dt, func(v int) { l.AttenuationType = v })
		case "292":
			err = setInt(dt, func(v int) { l.UseLimits = v != 0 })
		case "41":
			err = setFloat(dt, func(v float64) { l.AttenuationStart = v })
		case "42":
			err = setFloat(dt, func(v float64) { l.AttenuationEnd = v })
		case "50":
			err = setFloat(dt, func(v float64) { l.HotspotAngle = v })
		case "51":
			err = setFloat(dt, func(v float64) { l.FalloffAngle = v })
		case "293":
			err = setInt(dt, func(v int) { l.CastShadows = v != 0 })
		case "73":
			err = setInt(dt, func(v int) { l.ShadowType = v })
		case "91":
			err = setInt(dt, func(v int) { l.ShadowMapSize = v })
		case "280":
			err = setInt(dt, func(v int) { l.ShadowMapSoftness = v })
		}
		if err != nil {
			return l, err
		}
	}
	return l, nil
}

// ParseHelix parses HELIX entities: the spline and the helix data.
func ParseHelix(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	splineData, helixData := splitAt(data, "AcDbHelix")
	e, err := ParseSpline(d, splineData)
	if err != nil {
		return e, err
	}
	s := e.(*entity.Spline)
	s.Type = entity.HELIX
	h := &entity.Helix{Spline: s, AxisBase: []float64{0, 0, 0}, StartPoint: []float64{1, 0, 0}, Axis: []float64{0, 0, 1}, RightHanded: true}
	for _, dt := range helixData {
		switch dt[0] {
		case "90":
			err = setInt(dt, func(v int) { h.MajorRelease = v })
		case "91":
			err = setInt(dt, func(v int) { h.Maintenance = v })
		case "10", "20", "30":
			err = setPoint(dt, h.AxisBase)
		case "11", "21", "31":
			err = setPoint(dt, h.StartPoint)
		case "12", "22", "32":
			err = setPoint(dt, h.Axis)
		case "40":
			err = setFloat(dt, func(v float64) { h.Radius = v })
		case "41":
			err = setFloat(dt, func(v float64) { h.Turns = v })
		case "42":
			err = setFloat(dt, func(v float64) { h.TurnHeight = v })
		case "290":
			err = setInt(dt, func(v int) { h.RightHanded = v != 0 })
		case "280":
			err = setInt(dt, func(v int) { h.Constrain = v })
		}
		if err != nil {
			return h, err
		}
	}
	return h, nil
}

// ParseMesh parses MESH entities.
func ParseMesh(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	m := entity.NewMesh(nil, nil)
	head, body := splitAt(data, "AcDbSubDMesh")
	for _, dt := range head {
		if err := setLayerAndLtscale(d, m, dt); err != nil {
			return m, err
		}
	}
	r := &groupReader{data: body}
	var err error
	for !r.done() && err == nil {
		switch r.peek() {
		case "71":
			m.Version, err = r.int("71", 2)
		case "72":
			var v int
			v, err = r.int("72", 0)
			m.BlendCrease = v != 0
		case "91":
			m.Subdivision, err = r.int("91", 0)
		case "92":
			var n int
			if n, err = r.int("92", 0); err != nil {
				break
			}
			for i := 0; i < n && err == nil; i++ {
				var x, y, z float64
				if x, err = r.float("10", 0); err != nil {
					break
				}
				if y, err = r.float("20", 0); err != nil {
					break
				}
				z, err = r.float("30", 0)
				m.Vertices = append(m.Vertices, []float64{x, y, z})
			}
		case "93":
			var n int
			if n, err = r.int("93", 0); err != nil {
				break
			}
			for read := 0; read < n && err == nil; {
				var size int
				if size, err = r.int("90", 0); err != nil {
					break
				}
				read++
				face := make([]int, 0, size)
				for j := 0; j < size && err == nil; j++ {
					var v int
					v, err = r.int("90", 0)
					face = append(face, v)
					read++
				}
				m.Faces = append(m.Faces, face)
			}
		case "94":
			var n int
			if n, err = r.int("94", 0); err != nil {
				break
			}
			for i := 0; i < n && err == nil; i++ {
				var a, b int
				if a, err = r.int("90", 0); err != nil {
					break
				}
				b, err = r.int("90", 0)
				m.Edges = append(m.Edges, [2]int{a, b})
			}
		case "95":
			var n int
			if n, err = r.int("95", 0); err != nil {
				break
			}
			for i := 0; i < n && err == nil; i++ {
				var c float64
				c, err = r.float("140", 0)
				m.Creases = append(m.Creases, c)
			}
		default:
			r.next() // overridden properties (90 ...) are not kept
		}
	}
	return m, err
}

// ParseMPolygon parses MPOLYGON entities.
func ParseMPolygon(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	m := entity.NewMPolygon()
	m.Solid, m.PatternName = false, ""
	head, body := splitAt(data, "AcDbMPolygon")
	for _, dt := range head {
		if err := setLayerAndLtscale(d, m, dt); err != nil {
			return m, err
		}
	}
	r := &groupReader{data: body}
	var err error
	afterPaths := false
	for !r.done() && err == nil {
		switch r.peek() {
		case "70":
			m.Version, err = r.int("70", 1)
		case "30":
			if afterPaths {
				r.next()
			} else {
				m.Elevation, err = r.float("30", 0)
			}
		case "210", "220", "230":
			err = setPoint(r.next(), m.Extrusion)
		case "2":
			m.PatternName = r.next()[1]
		case "71":
			var v int
			v, err = r.int("71", 0)
			m.Solid = v == 1
		case "91":
			var n int
			if n, err = r.int("91", 0); err != nil {
				break
			}
			for i := 0; i < n && err == nil; i++ {
				var b *entity.HatchBoundary
				if b, err = parseMPolygonPath(r); err == nil {
					m.Boundaries = append(m.Boundaries, b)
				}
			}
			afterPaths = true
		case "76":
			m.PatternType, err = r.int("76", 0)
		case "52":
			m.PatternAngle, err = r.float("52", 0)
		case "41":
			m.PatternScale, err = r.float("41", 1)
		case "77":
			var v int
			v, err = r.int("77", 0)
			m.Double = v == 1
		case "78":
			var n int
			if n, err = r.int("78", 0); err != nil {
				break
			}
			for i := 0; i < n && err == nil; i++ {
				var l *entity.HatchPatternLine
				if l, err = parseHatchPatternLine(r); err == nil {
					m.PatternLines = append(m.PatternLines, l)
				}
			}
		case "63":
			m.FillColor, err = r.int("63", -1)
		case "11", "21":
			err = setPoint(r.next(), m.Offset)
		case "99":
			m.Degenerated, err = r.int("99", 0)
		case "450":
			m.Gradient, err = parseHatchGradient(r)
		default:
			r.next()
		}
	}
	return m, err
}

// parseMPolygonPath parses a polyline path of MPOLYGON (73 and 72 in any
// order, no source boundary objects).
func parseMPolygonPath(r *groupReader) (*entity.HatchBoundary, error) {
	flag, err := r.int("92", 0)
	if err != nil {
		return nil, err
	}
	b := &entity.HatchBoundary{Flag: flag | entity.HATCH_PATH_POLYLINE}
	hasBulge := 0
	for r.peek() == "72" || r.peek() == "73" {
		if r.peek() == "72" {
			hasBulge, err = r.int("72", 0)
		} else {
			var c int
			c, err = r.int("73", 0)
			b.Closed = c == 1
		}
		if err != nil {
			return nil, err
		}
	}
	n, err := r.int("93", 0)
	if err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		p, err := r.point2("10", "20")
		if err != nil {
			return nil, err
		}
		bulge := 0.0
		if hasBulge == 1 || r.peek() == "42" {
			if bulge, err = r.float("42", 0); err != nil {
				return nil, err
			}
		}
		b.Vertices = append(b.Vertices, p)
		b.Bulges = append(b.Bulges, bulge)
	}
	return b, nil
}

// ParseBody parses the entities with ACIS data (BODY, REGION, 3DSOLID and
// surfaces).
func ParseBody(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	name := "BODY"
	if len(data) > 0 && data[0][0] == entityTypeCode {
		name = strings.TrimSpace(data[0][1])
	}
	b := entity.NewBody(name, nil)
	b.SetVersionGiven(false)
	head, body := splitAt(data, "AcDbModelerGeometry")
	for _, dt := range head {
		if err := setLayerAndLtscale(d, b, dt); err != nil {
			return b, err
		}
	}
	var lines []string
	for i, dt := range body {
		var err error
		switch dt[0] {
		case "70":
			b.SetVersionGiven(true)
			err = setInt(dt, func(v int) { b.Version = v })
		case "1":
			lines = append(lines, dt[1])
		case "3":
			if len(lines) > 0 {
				lines[len(lines)-1] += dt[1]
			}
		case "290":
			err = setInt(dt, func(v int) { b.Flags = v })
		case "2":
			b.UID = dt[1]
		case "100":
			// subclasses after the modeler geometry are kept as read
			rest, _ := splitXData(body[i:])
			b.Extra = append([][2]string{}, rest...)
			b.SAT = entity.DecodeSAT(lines)
			return b, nil
		}
		if err != nil {
			return b, err
		}
		if dt[0] == "1001" {
			break
		}
	}
	b.SAT = entity.DecodeSAT(lines)
	return b, nil
}

// ParseSpecialDimension parses ARC_DIMENSION and LARGE_RADIAL_DIMENSION
// entities: the common dimension data and their subclass.
func ParseSpecialDimension(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	arc := len(data) > 0 && strings.EqualFold(strings.TrimSpace(data[0][1]), "ARC_DIMENSION")
	marker := "AcDbRadialDimensionLarge"
	if arc {
		marker = "AcDbArcDimension"
	}
	common, sub := splitAt(data, marker)
	e, err := ParseDimension(d, common)
	if err != nil {
		return e, err
	}
	dim := e.(*entity.Dimension)
	if arc {
		dim.Type = entity.DIM_ARC
	} else {
		dim.Type = entity.DIM_LARGE_RADIAL
	}
	dim.ArcLeader1, dim.ArcLeader2 = []float64{0, 0, 0}, []float64{0, 0, 0}
	for _, dt := range sub {
		switch dt[0] {
		case "13", "23", "33":
			err = setPoint(dt, dim.DefPoint2)
		case "14", "24", "34":
			err = setPoint(dt, dim.DefPoint3)
		case "15", "25", "35":
			err = setPoint(dt, dim.DefPoint4)
		case "16", "26", "36":
			err = setPoint(dt, dim.ArcLeader1)
		case "17", "27", "37":
			err = setPoint(dt, dim.ArcLeader2)
		case "40":
			if arc {
				err = setFloat(dt, func(v float64) { dim.ArcStart = v })
			} else {
				err = setFloat(dt, func(v float64) { dim.JogValue = v })
			}
		case "41":
			err = setFloat(dt, func(v float64) { dim.ArcEnd = v })
		case "70":
			err = setInt(dt, func(v int) { dim.ArcPartial = v != 0 })
		case "71":
			err = setInt(dt, func(v int) { dim.ArcHasLeader = v != 0 })
		}
		if err != nil {
			return dim, err
		}
	}
	return dim, nil
}
