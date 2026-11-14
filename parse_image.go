package dxf

import (
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// ParseImage parses IMAGE and WIPEOUT entities. The definition (340) and
// the IMAGEDEF_REACTOR (360) are resolved by ParseObjects.
func ParseImage(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	im := entity.NewImage(nil)
	if len(data) > 0 && strings.EqualFold(strings.TrimSpace(data[0][1]), "WIPEOUT") {
		im.Type = entity.WIPEOUT
	}
	im.Boundary = nil
	im.Flags, im.Clipping = 0, false
	var err error
	for _, dt := range data {
		switch dt[0] {
		case "8":
			if layer, lerr := d.Layer(dt[1], false); lerr == nil {
				im.SetLayer(layer)
			}
		case "48":
			err = setFloat(dt, func(val float64) { im.SetLtscale(val) })
		case "10", "20", "30":
			k := int(dt[0][0] - '1')
			err = setFloat(dt, func(val float64) { im.Insert[k] = val })
		case "11", "21", "31":
			k := int(dt[0][0] - '1')
			err = setFloat(dt, func(val float64) { im.U[k] = val })
		case "12", "22", "32":
			k := int(dt[0][0] - '1')
			err = setFloat(dt, func(val float64) { im.V[k] = val })
		case "13", "23":
			k := int(dt[0][0] - '1')
			err = setFloat(dt, func(val float64) { im.Size[k] = val })
		case "340":
			if h := strings.TrimSpace(dt[1]); h != "0" {
				im.SetDefinitionHandle(h)
			}
		case "360":
			if h := strings.TrimSpace(dt[1]); h != "0" {
				im.SetReactorHandle(h)
			}
		case "70":
			err = setInt(dt, func(val int) { im.Flags = val })
		case "280":
			err = setInt(dt, func(val int) { im.Clipping = val != 0 })
		case "281":
			err = setInt(dt, func(val int) { im.Brightness = val })
		case "282":
			err = setInt(dt, func(val int) { im.Contrast = val })
		case "283":
			err = setInt(dt, func(val int) { im.Fade = val })
		case "71":
			err = setInt(dt, func(val int) { im.BoundaryType = val })
		case "14":
			im.Boundary = append(im.Boundary, []float64{0, 0})
			err = setFloat(dt, func(val float64) { im.Boundary[len(im.Boundary)-1][0] = val })
		case "24":
			if len(im.Boundary) > 0 {
				err = setFloat(dt, func(val float64) { im.Boundary[len(im.Boundary)-1][1] = val })
			}
		case "290":
			err = setInt(dt, func(val int) { im.ClipInside = val != 0 })
		}
		if err != nil {
			return im, err
		}
	}
	return im, nil
}
