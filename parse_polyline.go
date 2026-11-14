package dxf

import (
	"fmt"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// ParsePolyline parses the POLYLINE entity header.
// Its VERTEX entities are attached by ParseEntities.
func ParsePolyline(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	p := entity.NewPolyline()
	p.Flag = 0
	var err error
	for _, dt := range data {
		switch dt[0] {
		case "8":
			layer, err := d.Layer(dt[1], false)
			if err == nil {
				p.SetLayer(layer)
			}
		case "48":
			err = setFloat(dt, func(val float64) { p.SetLtscale(val) })
		case "70":
			err = setInt(dt, func(val int) { p.Flag = val })
		case "30":
			err = setFloat(dt, func(val float64) { p.Elevation = val })
		case "40":
			err = setFloat(dt, func(val float64) { p.StartWidth = val })
		case "41":
			err = setFloat(dt, func(val float64) { p.EndWidth = val })
		case "71":
			err = setInt(dt, func(val int) { p.MeshM = val })
		case "72":
			err = setInt(dt, func(val int) { p.MeshN = val })
		case "73":
			err = setInt(dt, func(val int) { p.SmoothM = val })
		case "74":
			err = setInt(dt, func(val int) { p.SmoothN = val })
		case "75":
			err = setInt(dt, func(val int) { p.CurveType = val })
		case "210", "220", "230":
			if p.Extrusion == nil {
				p.Extrusion = []float64{0, 0, 1}
			}
			err = setExtrusion(dt, p.Extrusion)
		}
		if err != nil {
			return p, err
		}
	}
	return p, nil
}

// ParseLwPolyline parses LWPOLYLINE entities.
func ParseLwPolyline(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	lw := entity.NewLwPolyline(0)
	// ind is the index of the current vertex. Each vertex starts with code 10,
	// and the following 20, 40, 41, 42 belong to it (DXF reference, LWPOLYLINE).
	ind := -1
	var err error
	for _, dt := range data {
		switch dt[0] {
		default:
			continue
		case "8":
			layer, err := d.Layer(dt[1], false)
			if err == nil {
				lw.SetLayer(layer)
			}
		case "48":
			err = setFloat(dt, func(val float64) { lw.SetLtscale(val) })
		case "90":
			err = setInt(dt, func(val int) {
				lw.Num = val
				lw.Vertices = make([][]float64, val)
				for i := 0; i < val; i++ {
					lw.Vertices[i] = make([]float64, 2)
				}
				lw.Bulges = make([]float64, val)
				lw.Widths = make([][]float64, val)
				for i := 0; i < val; i++ {
					lw.Widths[i] = []float64{0, 0}
				}
			})
		case "10":
			ind++
			if ind < lw.Num {
				err = setFloat(dt, func(val float64) { lw.Vertices[ind][0] = val })
			} else {
				err = fmt.Errorf("LWPOLYLINE extra vertices")
			}
		case "20":
			if ind >= 0 && ind < lw.Num {
				err = setFloat(dt, func(val float64) { lw.Vertices[ind][1] = val })
			} else {
				err = fmt.Errorf("LWPOLYLINE code 20 without code 10")
			}
		case "42":
			if ind >= 0 && ind < lw.Num {
				err = setFloat(dt, func(val float64) { lw.Bulges[ind] = val })
			} else {
				err = fmt.Errorf("LWPOLYLINE bulge without vertex")
			}
		case "40", "41":
			if ind >= 0 && ind < lw.Num {
				k := 0
				if dt[0] == "41" {
					k = 1
				}
				err = setFloat(dt, func(val float64) { lw.Widths[ind][k] = val })
			} else {
				err = fmt.Errorf("LWPOLYLINE width without vertex")
			}
		case "43":
			err = setFloat(dt, func(val float64) { lw.ConstantWidth = val })
		case "38":
			err = setFloat(dt, func(val float64) { lw.Elevation = val })
		case "210":
			err = setFloat(dt, func(val float64) { lw.Extrusion[0] = val })
		case "220":
			err = setFloat(dt, func(val float64) { lw.Extrusion[1] = val })
		case "230":
			err = setFloat(dt, func(val float64) { lw.Extrusion[2] = val })
		case "70":
			// bit-coded: 1 = closed, 128 = plinegen
			err = setInt(dt, func(val int) {
				if val&1 != 0 {
					lw.Close()
				}
				lw.Plinegen = val&128 != 0
			})
		}
		if err != nil {
			return lw, err
		}
	}
	if ind+1 != lw.Num {
		return lw, fmt.Errorf("LWPOLYLINE not enough vertices")
	}
	return lw, nil
}

func ParseVertex(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	v := entity.NewDefaultVertex()
	v.Flag = 0 // default of code 70 is 0
	var err error
	for _, dt := range data {
		switch dt[0] {
		default:
			continue
		case "8":
			layer, err := d.Layer(dt[1], false)
			if err == nil {
				v.SetLayer(layer)
			}
		case "48":
			err = setFloat(dt, func(val float64) { v.SetLtscale(val) })
		case "10":
			err = setFloat(dt, func(val float64) { v.Coord[0] = val })
		case "20":
			err = setFloat(dt, func(val float64) { v.Coord[1] = val })
		case "30":
			err = setFloat(dt, func(val float64) { v.Coord[2] = val })
		case "42":
			err = setFloat(dt, func(val float64) { v.Bulge = val })
		case "70":
			err = setInt(dt, func(val int) { v.Flag = val })
		case "40":
			err = setFloat(dt, func(val float64) { v.StartWidth = val })
		case "41":
			err = setFloat(dt, func(val float64) { v.EndWidth = val })
		case "50":
			err = setFloat(dt, func(val float64) { v.Tangent = val })
		case "71", "72", "73", "74":
			k := int(dt[0][1] - '1')
			err = setInt(dt, func(val int) {
				for len(v.Indices) <= k {
					v.Indices = append(v.Indices, 0)
				}
				v.Indices[k] = val
			})
		}
		if err != nil {
			return v, err
		}
	}
	return v, nil
}
