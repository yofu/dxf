package dxf

import (
	"errors"
	"fmt"
	"github.com/yofu/dxf/block"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/header"
	"github.com/yofu/dxf/table"
	"strconv"
	"strings"
)

func SetFloat(data [2]string, f func(float64)) error {
	val, err := strconv.ParseFloat(strings.TrimSpace(data[1]), 64)
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
	parsers := []func(*drawing.Drawing, [][2]string) (table.SymbolTable, error) {
		ParseViewport,
		ParseLtype,
		ParseLayer,
		ParseStyle,
		ParseView,
		ParseUCS,
		ParseAppID,
		ParseDimStyle,
		ParseBlockRecord,
	}
	tmpdata := make([][2]string, 0)
	setparser := false
	var parser func(*drawing.Drawing, [][2]string) (table.SymbolTable, error)
	var ind int
	for i, dt := range data {
		if setparser {
			if dt[0] != "2" {
				return fmt.Errorf("line %d: invalid group code: %s", line + 2*i, dt[0])
			}
			ind = int(table.TableTypeValue(strings.ToUpper(dt[1])))
			if ind < 0 {
				return fmt.Errorf("line %d: unknown table type: %s", line + 2*i, dt[1])
			}
			parser = parsers[ind]
			setparser = false
		} else {
			if dt[0] == "0" {
				switch strings.ToUpper(dt[1]) {
				case "TABLE":
					setparser = true
				case "ENDTAB":
					if len(tmpdata) > 0 {
						err := ParseTable(d, tmpdata, ind, parser)
						if err != nil {
							return err
						}
						tmpdata = make([][2]string, 0)
					}
				default:
					tmpdata = append(tmpdata, dt)
				}
			} else {
				tmpdata = append(tmpdata, dt)
			}
		}
	}
	if len(tmpdata) > 0 {
		err := ParseTable(d, tmpdata, ind, parser)
		if err != nil {
			return fmt.Errorf("line %d: %s", line + 2*len(data), err.Error())
		}
		tmpdata = make([][2]string, 0)
	}
	return nil
}

func ParseTable(d *drawing.Drawing, data [][2]string, index int, parser func(*drawing.Drawing, [][2]string)(table.SymbolTable, error)) error {
	t := d.Sections[drawing.TABLES].(table.Tables)[index]
	t.Clear()
	tmpdata := make([][2]string, 0)
	for _, dt := range data {
		switch dt[0] {
		case "0":
			if len(tmpdata) > 0 {
				st, err := parser(d, tmpdata)
				if err != nil {
					return err
				}
				if st != nil {
					t.Add(st)
				}
				tmpdata = make([][2]string, 0)
			}
		default:
			tmpdata = append(tmpdata, dt)
		}
	}
	if len(tmpdata) > 0 {
		st, err := parser(d, tmpdata)
		if err != nil {
			return err
		}
		if st != nil {
			t.Add(st)
		}
		tmpdata = make([][2]string, 0)
	}
	return nil
}

func ParseViewport(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	v := table.NewViewport("")
	var err error
	for _, dt := range data {
		switch dt[0] {
		case "2":
			v.SetName(dt[1])
		case "10":
			err = SetFloat(dt, func(val float64) { v.LowerLeft[0] = val })
		case "20":
			err = SetFloat(dt, func(val float64) { v.LowerLeft[1] = val })
		case "11":
			err = SetFloat(dt, func(val float64) { v.UpperRight[0] = val })
		case "21":
			err = SetFloat(dt, func(val float64) { v.UpperRight[1] = val })
		case "12":
			err = SetFloat(dt, func(val float64) { v.ViewCenter[0] = val })
		case "22":
			err = SetFloat(dt, func(val float64) { v.ViewCenter[1] = val })
		case "13":
			err = SetFloat(dt, func(val float64) { v.SnapBase[0] = val })
		case "23":
			err = SetFloat(dt, func(val float64) { v.SnapBase[1] = val })
		case "14":
			err = SetFloat(dt, func(val float64) {
				v.SnapSpacing[0] = val
				v.SnapSpacing[1] = val
			})
		case "24":
			err = SetFloat(dt, func(val float64) { v.SnapSpacing[1] = val })
		case "15":
			err = SetFloat(dt, func(val float64) {
				v.GridSpacing[0] = val
				v.GridSpacing[1] = val
			})
		case "25":
			err = SetFloat(dt, func(val float64) { v.GridSpacing[1] = val })
		case "16":
			err = SetFloat(dt, func(val float64) { v.ViewDirection[0] = val })
		case "26":
			err = SetFloat(dt, func(val float64) { v.ViewDirection[1] = val })
		case "36":
			err = SetFloat(dt, func(val float64) { v.ViewDirection[2] = val })
		case "17":
			err = SetFloat(dt, func(val float64) { v.ViewTarget[0] = val })
		case "27":
			err = SetFloat(dt, func(val float64) { v.ViewTarget[1] = val })
		case "37":
			err = SetFloat(dt, func(val float64) { v.ViewTarget[2] = val })
		case "40":
			err = SetFloat(dt, func(val float64) { v.Height = val })
		case "41":
			err = SetFloat(dt, func(val float64) { v.AspectRatio = val })
		case "42":
			err = SetFloat(dt, func(val float64) { v.LensLength = val })
		case "43":
			err = SetFloat(dt, func(val float64) { v.FrontClip = val })
		case "44":
			err = SetFloat(dt, func(val float64) { v.BackClip = val })
		case "50":
			err = SetFloat(dt, func(val float64) { v.SnapAngle = val })
		case "51":
			err = SetFloat(dt, func(val float64) { v.TwistAngle = val })
		}
		if err != nil {
			return v, err
		}
	}
	return v, nil
}

func ParseLtype(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	var name, desc string
	var lengths []float64
	ind := 0
	for _, dt := range data {
		switch dt[0] {
		case "2":
			name = dt[1]
		case "3":
			desc = dt[1]
		case "73":
			l, err := strconv.ParseInt(strings.TrimSpace(dt[1]), 10, 64)
			if err != nil {
				return nil, err
			}
			lengths = make([]float64, int(l))
		case "49":
			if ind >= len(lengths) {
				return nil, fmt.Errorf("ltype too long")
			}
			val, err := strconv.ParseFloat(dt[1], 64)
			if err != nil {
				return nil, err
			}
			lengths[ind] = val
			ind++
		}
	}
	return table.NewLineType(name, desc, lengths...), nil
}

func ParseLayer(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	var name string
	var flag int
	var col color.ColorNumber
	var lt *table.LineType
	var lw int
	for _, dt := range data {
		switch dt[0] {
		case "2":
			name = dt[1]
		case "70":
			val, err := strconv.ParseInt(strings.TrimSpace(dt[1]), 10, 64)
			if err != nil {
				return nil, err
			}
			flag = int(val)
		case "62":
			val, err := strconv.ParseInt(strings.TrimSpace(dt[1]), 10, 64)
			if err != nil {
				return nil, err
			}
			col = color.ColorNumber(val)
		case "6":
			l, err := d.LineType(dt[1])
			if err != nil {
				return nil, err
			}
			lt = l
		case "370":
			val, err := strconv.ParseInt(strings.TrimSpace(dt[1]), 10, 64)
			if err != nil {
				return nil, err
			}
			lw = int(val)
		case "390":
			// plotstyle
		}
	}
	l := table.NewLayer(name, col, lt)
	l.SetFlag(flag)
	l.SetLineWidth(lw)
	return l, nil
}

func ParseStyle(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	var name, font, bigfont string
	for _, dt := range data {
		switch dt[0] {
		case "2":
			name = dt[1]
		case "3":
			font = dt[1]
		case "4":
			bigfont = dt[1]
		}
	}
	s := table.NewStyle(name)
	s.FontName = font
	s.BigFontName = bigfont
	return s, nil
}

func ParseView(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	var name string
	for _, dt := range data {
		switch dt[0] {
		case "2":
			name = dt[1]
		}
	}
	v := table.NewView(name)
	return v, nil
}

func ParseUCS(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	var name string
	for _, dt := range data {
		switch dt[0] {
		case "2":
			name = dt[1]
		}
	}
	u := table.NewUCS(name)
	return u, nil
}

func ParseAppID(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	var name string
	for _, dt := range data {
		switch dt[0] {
		case "2":
			name = dt[1]
		}
	}
	a := table.NewAppID(name)
	return a, nil
}

func ParseDimStyle(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	var name string
	for _, dt := range data {
		switch dt[0] {
		case "2":
			name = dt[1]
		}
	}
	ds := table.NewDimStyle(name)
	return ds, nil
}

func ParseBlockRecord(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	var name string
	for _, dt := range data {
		switch dt[0] {
		case "2":
			name = dt[1]
		}
	}
	b := table.NewBlockRecord(name)
	return b, nil
}

// BLOCKS
func ParseBlocks(d *drawing.Drawing, line int, data [][2]string) error {
	tmpdata := make([][2]string, 0)
	add := true // skip ENDBLK
	for i, dt := range data {
		if dt[0] == "0" {
			switch strings.ToUpper(dt[1]) {
			case "BLOCK":
				add = true
			case "ENDBLK":
				if len(tmpdata) > 0 {
					err := ParseBlock(d, tmpdata)
					if err != nil {
						return fmt.Errorf("line %d: %s", line + 2*i, err.Error())
					}
					tmpdata = make([][2]string, 0)
				}
				add = false
			default:
				if add {
					tmpdata = append(tmpdata, dt)
				}
			}
		} else {
			if add {
				tmpdata = append(tmpdata, dt)
			}
		}
	}
	if len(tmpdata) > 0 {
		err := ParseBlock(d, tmpdata)
		if err != nil {
			return fmt.Errorf("line %d: %s", line + 2*len(data), err.Error())
		}
		tmpdata = make([][2]string, 0)
	}
	return nil
}

func ParseBlock(d *drawing.Drawing, data [][2]string) error {
	b := block.NewBlock("", "")
	var err error
	for _, dt := range data {
		switch dt[0] {
		case "2":
			b.Name = dt[1]
		case "1": // 4?
			b.Description = dt[1]
		case "8":
			layer, err := d.Layer(dt[1], false)
			if err == nil {
				b.SetLayer(layer)
			}
		case "10":
			err = SetFloat(dt, func(val float64) { b.Coord[0] = val })
		case "20":
			err = SetFloat(dt, func(val float64) { b.Coord[1] = val })
		case "30":
			err = SetFloat(dt, func(val float64) { b.Coord[2] = val })
		case "70":
			val, err := strconv.ParseInt(strings.TrimSpace(dt[1]), 10, 64)
			if err != nil {
				return err
			}
			b.Flag = int(val)
		}
		if err != nil {
			return err
		}
	}
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
