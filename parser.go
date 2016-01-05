package dxf

import (
	"errors"
	"fmt"
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/header"
	"strconv"
)

func SetFloat(data [2]string, f func(float64)) error {
	val, err := strconv.ParseFloat(data[1], 64)
	if err != nil {
		return fmt.Errorf("code %s: %s", data[0], err.Error())
	}
	f(val)
	return nil
}

// HEADER
func ParseHeader(d *drawing.Drawing, line int, data [][2]string) error {
	h := d.Sections[drawing.HEADER].(*header.Header)
	var name string
	var err error
	for _, dt := range data {
		switch dt[0] {
		case "9":
			name = dt[1]
		case "1":
			switch name {
			case "$ACADVER":
				h.Version = dt[1]
			}
		case "10":
			switch name {
			case "$INSBASE":
				err = SetFloat(dt, func(val float64) { h.InsBase[0] = val })
			case "$EXTMIN":
				err = SetFloat(dt, func(val float64) { h.ExtMin[0] = val })
			case "$EXTMAX":
				err = SetFloat(dt, func(val float64) { h.ExtMax[0] = val })
			}
		case "20":
			switch name {
			case "$INSBASE":
				err = SetFloat(dt, func(val float64) { h.InsBase[1] = val })
			case "$EXTMIN":
				err = SetFloat(dt, func(val float64) { h.ExtMin[1] = val })
			case "$EXTMAX":
				err = SetFloat(dt, func(val float64) { h.ExtMax[1] = val })
			}
		case "30":
			switch name {
			case "$INSBASE":
				err = SetFloat(dt, func(val float64) { h.InsBase[2] = val })
			case "$EXTMIN":
				err = SetFloat(dt, func(val float64) { h.ExtMin[2] = val })
			case "$EXTMAX":
				err = SetFloat(dt, func(val float64) { h.ExtMax[2] = val })
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// CLASSES
func ParseClasses(d *drawing.Drawing, line int, data [][2]string) error {
	return nil
}

// TABLES
func ParseTables(d *drawing.Drawing, line int, data [][2]string) error {
	return nil
}

// BLOCKS
func ParseBlocks(d *drawing.Drawing, line int, data [][2]string) error {
	return nil
}

// ENTITIES
func ParseEntities(d *drawing.Drawing, line int, data [][2]string) error {
	tmpdata := make([][2]string, 0)
	for i, dt := range data {
		if dt[0] == "0" {
			if len(tmpdata) > 0 {
				e, err := ParseEntity(d, tmpdata)
				if err != nil {
					return fmt.Errorf("line %d: %s", line + 2*i, err.Error())
				}
				d.AddEntity(e)
				tmpdata = make([][2]string, 0)
			}
		}
		tmpdata = append(tmpdata, dt)
	}
	if len(tmpdata) > 0 {
		e, err := ParseEntity(d, tmpdata)
		if err != nil {
			return fmt.Errorf("line %d: %s", line + 2*len(data), err.Error())
		}
		d.AddEntity(e)
		tmpdata = make([][2]string, 0)
	}
	return nil
}

func ParseEntity(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("no data")
	}
	if data[0][0] != "0" {
		return nil, fmt.Errorf("invalid group code: %d", data[0][0])
	}
	f, err := ParseEntityFunc(data[0][1])
	if err != nil {
		return nil, err
	}
	return f(d, data)
}

func ParseEntityFunc(t string) (func(*drawing.Drawing, [][2]string)(entity.Entity, error), error) {
	switch t {
	case "LINE":
		return ParseLine, nil
	// case "3DFACE":
	// 	return Parse3DFace, nil
	// case "LWPOLYLINE":
	// 	return ParseLwPolyline, nil
	// case "CIRCLE":
	// 	return ParseCircle, nil
	// case "POLYLINE":
	// 	return ParsePolyline, nil
	// case "VERTEX":
	// 	return ParseVertex, nil
	// case "POINT":
	// 	return ParsePoint, nil
	// case "TEXT":
	// 	return ParseText, nil
	default:
		return nil, errors.New("unknown entity type")
	}
}

func ParseLine(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	l := entity.NewLine()
	var err error
	for _, dt := range data {
		switch dt[0] {
		default:
			continue
		case "8":
			if layer, exists := d.Layers[dt[1]]; exists {
				l.SetLayer(layer)
			} else {
				err = fmt.Errorf("unknown layer: %s", dt[1])
			}
		case "10":
			err = SetFloat(dt, func(val float64) { l.Start[0] = val })
		case "20":
			err = SetFloat(dt, func(val float64) { l.Start[1] = val })
		case "30":
			err = SetFloat(dt, func(val float64) { l.Start[2] = val })
		case "11":
			err = SetFloat(dt, func(val float64) { l.End[0] = val })
		case "21":
			err = SetFloat(dt, func(val float64) { l.End[1] = val })
		case "31":
			err = SetFloat(dt, func(val float64) { l.End[2] = val })
		}
		if err != nil {
			return l, err
		}
	}
	return l, nil
}

// OBJECTS
func ParseObjects(d *drawing.Drawing, line int, data [][2]string) error {
	return nil
}
