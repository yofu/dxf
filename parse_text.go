package dxf

import (
	"math"
	"strconv"
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// ParseText parses TEXT entities.
func ParseText(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	t := entity.NewText()
	for _, dt := range data {
		var err error
		if dt[0] == "73" {
			err = setInt(dt, func(val int) { t.VerticalFlag = val })
		} else {
			err = parseTextCode(d, t, dt)
		}
		if err != nil {
			return t, err
		}
	}
	return t, nil
}

// parseTextCode parses a group code of TEXT shared with ATTDEF and ATTRIB
// (all except the vertical justification, 73 in TEXT and 74 in attributes).
func parseTextCode(d *drawing.Drawing, t *entity.Text, dt [2]string) error {
	var err error
	switch dt[0] {
	case "8":
		if layer, lerr := d.Layer(dt[1], false); lerr == nil {
			t.SetLayer(layer)
		}
	case "48":
		err = setFloat(dt, func(val float64) { t.SetLtscale(val) })
	case "10":
		err = setFloat(dt, func(val float64) { t.Coord1[0] = val })
	case "20":
		err = setFloat(dt, func(val float64) { t.Coord1[1] = val })
	case "30":
		err = setFloat(dt, func(val float64) { t.Coord1[2] = val })
	case "11":
		err = setFloat(dt, func(val float64) { t.Coord2[0] = val })
	case "21":
		err = setFloat(dt, func(val float64) { t.Coord2[1] = val })
	case "31":
		err = setFloat(dt, func(val float64) { t.Coord2[2] = val })
	case "40":
		err = setFloat(dt, func(val float64) { t.Height = val })
	case "50":
		err = setFloat(dt, func(val float64) { t.Rotation = val })
	case "41":
		err = setFloat(dt, func(val float64) { t.WidthFactor = val })
	case "51":
		err = setFloat(dt, func(val float64) { t.ObliqueAngle = val })
	case "1":
		t.Value = dt[1]
	case "7":
		if s, ok := findStyle(d, dt[1]); ok {
			t.Style = s
		}
	case "71":
		err = setInt(dt, func(val int) { t.GenFlag = val })
	case "72":
		err = setInt(dt, func(val int) { t.HorizontalFlag = val })
	case "210", "220", "230":
		err = setExtrusion(dt, t.Extrusion)
	}
	return err
}

// ParseAttdef parses ATTDEF entities.
func ParseAttdef(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	a := entity.NewAttributeDefinition("", "", "")
	if s, ok := findStyle(d, "Standard"); ok {
		a.Style = s
	}
	for _, dt := range data {
		var err error
		switch dt[0] {
		case "2":
			a.Tag = dt[1]
		case "3":
			a.Prompt = dt[1]
		case "70":
			err = setInt(dt, func(val int) { a.Flags = val })
		case "73":
			err = setInt(dt, func(val int) { a.FieldLength = val })
		case "74":
			err = setInt(dt, func(val int) { a.VerticalFlag = val })
		default:
			err = parseTextCode(d, a.Text, dt)
		}
		if err != nil {
			return a, err
		}
	}
	return a, nil
}

// ParseAttrib parses ATTRIB entities.
// They are attached to the preceding INSERT by ParseEntities.
func ParseAttrib(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	a := entity.NewAttribute("", "")
	if s, ok := findStyle(d, "Standard"); ok {
		a.Style = s
	}
	for _, dt := range data {
		var err error
		switch dt[0] {
		case "2":
			a.Tag = dt[1]
		case "70":
			err = setInt(dt, func(val int) { a.Flags = val })
		case "73":
			err = setInt(dt, func(val int) { a.FieldLength = val })
		case "74":
			err = setInt(dt, func(val int) { a.VerticalFlag = val })
		default:
			err = parseTextCode(d, a.Text, dt)
		}
		if err != nil {
			return a, err
		}
	}
	return a, nil
}

// OBJECTS

// ParseMText parses MTEXT entities.
// Text in group 3 (250-character chunks) and group 1 is concatenated.
// The X-axis direction (11, 21, 31) takes precedence over the rotation angle (50).
func ParseMText(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	t := entity.NewMText()
	if s, ok := findStyle(d, "Standard"); ok {
		t.Style = s
	}
	var text strings.Builder
	xaxis := []float64{0, 0, 0}
	hasXAxis := false
	var err error
	// the embedded object (101) of AC1032 has the column layout
	for i, dt := range data {
		if dt[0] == "101" {
			cols, _ := splitXData(data[i+1:])
			data = data[:i]
			t.Columns = parseMTextColumns(cols)
			break
		}
	}
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
			err = setFloat(dt, func(val float64) { t.Coord[0] = val })
		case "20":
			err = setFloat(dt, func(val float64) { t.Coord[1] = val })
		case "30":
			err = setFloat(dt, func(val float64) { t.Coord[2] = val })
		case "40":
			err = setFloat(dt, func(val float64) { t.Height = val })
		case "41":
			err = setFloat(dt, func(val float64) { t.Width = val })
		case "71":
			err = setInt(dt, func(val int) { t.Attachment = val })
		case "72":
			err = setInt(dt, func(val int) { t.Direction = val })
		case "1", "3":
			text.WriteString(dt[1])
		case "7":
			if s, ok := findStyle(d, dt[1]); ok {
				t.Style = s
			}
		case "11":
			err = setFloat(dt, func(val float64) { xaxis[0] = val; hasXAxis = true })
		case "21":
			err = setFloat(dt, func(val float64) { xaxis[1] = val; hasXAxis = true })
		case "31":
			err = setFloat(dt, func(val float64) { xaxis[2] = val })
		case "50":
			// degrees as written by AutoCAD (the DXF reference says radians)
			err = setFloat(dt, func(val float64) {
				if !hasXAxis {
					t.Rotation = val
				}
			})
		case "73":
			err = setInt(dt, func(val int) { t.LineSpacingStyle = val })
		case "44":
			err = setFloat(dt, func(val float64) { t.LineSpacingFactor = val })
		case "90":
			err = setInt(dt, func(val int) { t.BackgroundFill = val })
		case "45":
			err = setFloat(dt, func(val float64) { t.BackgroundScale = val })
		case "63":
			err = setInt(dt, func(val int) { t.BackgroundColor = val })
		case "421":
			err = setInt(dt, func(val int) { t.BackgroundTrueColor = val })
		case "431":
			t.BackgroundColorName = dt[1]
		case "441":
			err = setInt(dt, func(val int) { t.BackgroundTransparency = val })
		case "210":
			err = setFloat(dt, func(val float64) { t.Extrusion[0] = val })
		case "220":
			err = setFloat(dt, func(val float64) { t.Extrusion[1] = val })
		case "230":
			err = setFloat(dt, func(val float64) { t.Extrusion[2] = val })
		}
		if err != nil {
			return t, err
		}
	}
	t.Value = text.String()
	if hasXAxis && (xaxis[0] != 0 || xaxis[1] != 0) {
		t.Rotation = math.Atan2(xaxis[1], xaxis[0]) * 180.0 / math.Pi
	}
	return t, nil
}

// parseMTextColumns parses the embedded object of MTEXT (columns).
func parseMTextColumns(data [][2]string) *entity.MTextColumns {
	c := &entity.MTextColumns{}
	for _, dt := range data {
		v := strings.TrimSpace(dt[1])
		f, _ := strconv.ParseFloat(v, 64)
		i, _ := strconv.Atoi(v)
		switch dt[0] {
		case "41":
			c.DefinedHeight = f
		case "42":
			c.TotalWidth = f
		case "43":
			c.TotalHeight = f
		case "71":
			c.Type = i
		case "72":
			c.Count = i
		case "44":
			c.Width = f
		case "45":
			c.Gutter = f
		case "73":
			c.AutoHeight = i != 0
		case "74":
			c.Reversed = i != 0
		case "46":
			c.Heights = append(c.Heights, f)
		}
	}
	if c.Type == entity.MTEXT_COLUMNS_DYNAMIC && c.AutoHeight && c.Count == 0 {
		c.Count = 1
	}
	return c
}
