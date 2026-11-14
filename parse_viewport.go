package dxf

import (
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// ParseViewportEntity parses VIEWPORT entities
// (ParseViewport parses VPORT table entries).
func ParseViewportEntity(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	v := entity.NewViewport()
	v.SetPaperSpace(false) // set by 67
	floats := map[string]*float64{
		"10": &v.Center[0], "20": &v.Center[1], "30": &v.Center[2],
		"40": &v.Width, "41": &v.Height,
		"12": &v.ViewCenter[0], "22": &v.ViewCenter[1],
		"13": &v.SnapBase[0], "23": &v.SnapBase[1],
		"14": &v.SnapSpacing[0], "24": &v.SnapSpacing[1],
		"15": &v.GridSpacing[0], "25": &v.GridSpacing[1],
		"16": &v.ViewDirection[0], "26": &v.ViewDirection[1], "36": &v.ViewDirection[2],
		"17": &v.ViewTarget[0], "27": &v.ViewTarget[1], "37": &v.ViewTarget[2],
		"42": &v.LensLength, "43": &v.FrontClip, "44": &v.BackClip, "45": &v.ViewHeight,
		"50": &v.SnapAngle, "51": &v.TwistAngle,
		"110": &v.UCSOrigin[0], "120": &v.UCSOrigin[1], "130": &v.UCSOrigin[2],
		"111": &v.UCSXAxis[0], "121": &v.UCSXAxis[1], "131": &v.UCSXAxis[2],
		"112": &v.UCSYAxis[0], "122": &v.UCSYAxis[1], "132": &v.UCSYAxis[2],
		"146": &v.Elevation,
	}
	ints := map[string]*int{
		"68": &v.Status, "69": &v.ID, "72": &v.CircleZoom, "90": &v.Flags,
		"281": &v.RenderMode, "71": &v.UCSPerView, "74": &v.UCSIcon, "79": &v.OrthoType,
	}
	for _, dt := range data {
		var err error
		if p, ok := floats[dt[0]]; ok {
			err = setFloat(dt, func(val float64) { *p = val })
		} else if p, ok := ints[dt[0]]; ok {
			err = setInt(dt, func(val int) { *p = val })
		} else {
			switch dt[0] {
			case "8":
				if layer, lerr := d.Layer(dt[1], false); lerr == nil {
					v.SetLayer(layer)
				}
			case "1":
				v.PlotStyle = dt[1]
			case "331", "340", "345", "346":
				v.References = append(v.References, [2]string{dt[0], dt[1]})
			}
		}
		if err != nil {
			return v, err
		}
	}
	return v, nil
}
