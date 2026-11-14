package dxf

import (
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// ParseUnderlay parses PDFUNDERLAY (PDFREFERENCE), DWFUNDERLAY and
// DGNUNDERLAY entities. The definition (340) is resolved by ParseObjects.
func ParseUnderlay(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	u := entity.NewUnderlay(nil)
	u.Flags, u.Contrast = 10, 100 // defaults of the DXF reference
	if len(data) > 0 && data[0][0] == entityTypeCode {
		u.SetTypeName(strings.TrimSpace(data[0][1]))
	}
	var err error
	for _, dt := range data {
		switch dt[0] {
		case "8":
			if layer, lerr := d.Layer(dt[1], false); lerr == nil {
				u.SetLayer(layer)
			}
		case "48":
			err = setFloat(dt, func(val float64) { u.SetLtscale(val) })
		case "340":
			u.SetDefinitionHandle(strings.TrimSpace(dt[1]))
		case "10", "20", "30":
			k := int(dt[0][0] - '1')
			err = setFloat(dt, func(val float64) { u.Insert[k] = val })
		case "41", "42", "43":
			k := int(dt[0][1] - '1')
			err = setFloat(dt, func(val float64) { u.Scale[k] = val })
		case "50":
			err = setFloat(dt, func(val float64) { u.Rotation = val })
		case "210", "220", "230":
			k := int(dt[0][1] - '1')
			err = setFloat(dt, func(val float64) { u.Extrusion[k] = val })
		case "280":
			err = setInt(dt, func(val int) { u.Flags = val })
		case "281":
			err = setInt(dt, func(val int) { u.Contrast = val })
		case "282":
			err = setInt(dt, func(val int) { u.Fade = val })
		case "11":
			u.Boundary = append(u.Boundary, []float64{0, 0})
			err = setFloat(dt, func(val float64) { u.Boundary[len(u.Boundary)-1][0] = val })
		case "21":
			if len(u.Boundary) > 0 {
				err = setFloat(dt, func(val float64) { u.Boundary[len(u.Boundary)-1][1] = val })
			}
		}
		if err != nil {
			return u, err
		}
	}
	return u, nil
}
