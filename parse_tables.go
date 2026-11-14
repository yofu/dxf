package dxf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/handle"
	"github.com/yofu/dxf/table"
)

// ParseTables parses TABLES section.
func ParseTables(d *drawing.Drawing, line int, data [][2]string) error {
	parsers := []func(*drawing.Drawing, [][2]string) (table.SymbolTable, error){
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
				return fmt.Errorf("line %d: invalid group code: %s", line+2*i, dt[0])
			}
			ind = int(table.TableTypeValue(strings.ToUpper(dt[1])))
			if ind < 0 {
				return fmt.Errorf("line %d: unknown table type: %s", line+2*i+1, dt[1])
			}
			parser = parsers[ind]
			setparser = false
		} else {
			if dt[0] == entityTypeCode {
				switch strings.ToUpper(dt[1]) {
				case "TABLE":
					setparser = true
				case "ENDTAB":
					if len(tmpdata) > 0 {
						err := ParseTable(d, tmpdata, ind, parser)
						if err != nil {
							return fmt.Errorf("line %d: %w", line+2*i-len(tmpdata)*2, err)
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
			return fmt.Errorf("line %d: %w", line+2*len(data)-len(tmpdata)*2, err)
		}
	}
	resolveLinetypeStyles(d)
	return nil
}

// ParseTable parses each TABLE, which starts with "0\nTABLE\n" and ends with "0\nENDTAB\n".
func ParseTable(d *drawing.Drawing, data [][2]string, index int, parser func(*drawing.Drawing, [][2]string) (table.SymbolTable, error)) error {
	t := d.Sections[drawing.TABLES].(table.Tables)[index]
	t.Clear()
	tmpdata := make([][2]string, 0)
	add := false // skip before first 0-code
	for _, dt := range data {
		switch dt[0] {
		case "0":
			if len(tmpdata) > 0 {
				st, err := parser(d, tmpdata)
				if err != nil {
					return err
				}
				setSourceHandle(st, tmpdata)
				t.Add(st)
				register(d, st)
				tmpdata = make([][2]string, 0)
			}
			add = true
		default:
			if add {
				tmpdata = append(tmpdata, dt)
			} else if dt[0] == "5" { // handle of the TABLE
				t.SetSourceHandle(dt[1])
			}
		}
	}
	if len(tmpdata) > 0 {
		st, err := parser(d, tmpdata)
		if err != nil {
			return err
		}
		setSourceHandle(st, tmpdata)
		t.Add(st)
		register(d, st)
	}
	return nil
}

// register adds layers and styles read from TABLES to the drawing's lookup maps.
func register(d *drawing.Drawing, st table.SymbolTable) {
	switch s := st.(type) {
	case *table.Layer:
		d.Layers[s.Name()] = s
	case *table.Style:
		for k := range d.Styles {
			if strings.EqualFold(k, s.Name()) {
				if d.CurrentStyle == d.Styles[k] {
					d.CurrentStyle = s
				}
				delete(d.Styles, k)
			}
		}
		d.Styles[s.Name()] = s
	}
}

// findStyle looks up a text style by name (case-insensitive).
func findStyle(d *drawing.Drawing, name string) (*table.Style, bool) {
	if s, ok := d.Styles[name]; ok {
		return s, true
	}
	for k, s := range d.Styles {
		if strings.EqualFold(k, name) {
			return s, true
		}
	}
	return nil, false
}

// ParseViewport parses VPORT tables.
func ParseViewport(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	v := table.NewViewport("")
	var err error
	for _, dt := range data {
		switch dt[0] {
		case "2":
			v.SetName(dt[1])
		case "10":
			err = setFloat(dt, func(val float64) { v.LowerLeft[0] = val })
		case "20":
			err = setFloat(dt, func(val float64) { v.LowerLeft[1] = val })
		case "11":
			err = setFloat(dt, func(val float64) { v.UpperRight[0] = val })
		case "21":
			err = setFloat(dt, func(val float64) { v.UpperRight[1] = val })
		case "12":
			err = setFloat(dt, func(val float64) { v.ViewCenter[0] = val })
		case "22":
			err = setFloat(dt, func(val float64) { v.ViewCenter[1] = val })
		case "13":
			err = setFloat(dt, func(val float64) { v.SnapBase[0] = val })
		case "23":
			err = setFloat(dt, func(val float64) { v.SnapBase[1] = val })
		case "14":
			err = setFloat(dt, func(val float64) {
				v.SnapSpacing[0] = val
				v.SnapSpacing[1] = val
			})
		case "24":
			err = setFloat(dt, func(val float64) { v.SnapSpacing[1] = val })
		case "15":
			err = setFloat(dt, func(val float64) {
				v.GridSpacing[0] = val
				v.GridSpacing[1] = val
			})
		case "25":
			err = setFloat(dt, func(val float64) { v.GridSpacing[1] = val })
		case "16":
			err = setFloat(dt, func(val float64) { v.ViewDirection[0] = val })
		case "26":
			err = setFloat(dt, func(val float64) { v.ViewDirection[1] = val })
		case "36":
			err = setFloat(dt, func(val float64) { v.ViewDirection[2] = val })
		case "17":
			err = setFloat(dt, func(val float64) { v.ViewTarget[0] = val })
		case "27":
			err = setFloat(dt, func(val float64) { v.ViewTarget[1] = val })
		case "37":
			err = setFloat(dt, func(val float64) { v.ViewTarget[2] = val })
		case "40":
			err = setFloat(dt, func(val float64) { v.Height = val })
		case "41":
			err = setFloat(dt, func(val float64) { v.AspectRatio = val })
		case "42":
			err = setFloat(dt, func(val float64) { v.LensLength = val })
		case "43":
			err = setFloat(dt, func(val float64) { v.FrontClip = val })
		case "44":
			err = setFloat(dt, func(val float64) { v.BackClip = val })
		case "50":
			err = setFloat(dt, func(val float64) { v.SnapAngle = val })
		case "51":
			err = setFloat(dt, func(val float64) { v.TwistAngle = val })
		}
		if err != nil {
			return v, err
		}
	}
	return v, nil
}

// ParseLtype parses LTYPE tables, including the text strings and shapes of
// complex linetypes. Their styles (340) are resolved by ParseTables.
func ParseLtype(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	lt := table.NewLineType("", "")
	var name string
	total := -1.0
	var es []*table.LineTypeElement
	var cur *table.LineTypeElement
	for _, dt := range data {
		var err error
		if cur == nil && dt[0] != "49" {
			switch dt[0] {
			case "2":
				name = dt[1]
			case "3":
				lt.Description = dt[1]
			case "40":
				if err := setFloat(dt, func(val float64) { total = val }); err != nil {
					return nil, err
				}
			}
			continue
		}
		switch dt[0] {
		case "49":
			cur = &table.LineTypeElement{}
			err = setFloat(dt, func(val float64) { cur.Length = val })
			es = append(es, cur)
		case "74":
			err = setInt(dt, func(val int) { cur.Flags = val })
		case "75":
			err = setInt(dt, func(val int) { cur.Shape = val })
		case "340":
			cur.SetStyleHandle(strings.TrimSpace(dt[1]))
		case "46":
			err = setFloat(dt, func(val float64) { cur.Scale = val })
		case "50":
			err = setFloat(dt, func(val float64) { cur.Rotation = val })
		case "44", "45":
			if cur.Offset == nil {
				cur.Offset = []float64{0, 0}
			}
			k := 0
			if dt[0] == "45" {
				k = 1
			}
			err = setFloat(dt, func(val float64) { cur.Offset[k] = val })
		case "9":
			cur.Text = dt[1]
		}
		if err != nil {
			return nil, err
		}
	}
	nlt := table.NewLineType(name, lt.Description)
	nlt.SetElements(es...)
	nlt.SetTotalLength(total)
	return nlt, nil
}

// resolveLinetypeStyles links the text styles and shape files of complex
// linetype elements (340) to the STYLE table entries read.
func resolveLinetypeStyles(d *drawing.Drawing) {
	tables := d.Sections[drawing.TABLES].(table.Tables)
	styles := map[int]*table.Style{}
	for _, st := range tables[table.STYLE].Entries() {
		if s, ok := st.(*table.Style); ok && s.SourceHandle() != "" {
			styles[handle.Parse(s.SourceHandle())] = s
		}
	}
	for _, st := range tables[table.LTYPE].Entries() {
		lt, ok := st.(*table.LineType)
		if !ok {
			continue
		}
		for _, e := range lt.Elements() {
			if e.StyleHandle() == "" {
				continue
			}
			if s, ok := styles[handle.Parse(e.StyleHandle())]; ok {
				e.Style = s
				e.SetStyleHandle("")
			}
		}
	}
}

// ParseLayer parses LAYER tables.
func ParseLayer(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	var name string
	var flag int
	var col color.ColorNumber
	var lt *table.LineType
	lw := -3 // default lineweight
	off := false
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
			// negative color number means the layer is off
			if val < 0 {
				off = true
				val = -val
			}
			col = color.ColorNumber(val)
		case "6":
			l, err := d.LineType(dt[1])
			if err != nil {
				// keep the reference even if the linetype is not defined
				l = table.NewLineType(dt[1], "")
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
	if lt == nil {
		if c, err := d.LineType(table.LT_CONTINUOUS.Name()); err == nil {
			lt = c
		} else {
			lt = table.LT_CONTINUOUS
		}
	}
	l := table.NewLayer(name, col, lt)
	if off {
		l.Off()
	}
	l.SetFlag(flag)
	l.SetLineWidth(lw)
	l.SetPlotStyle(d.PlotStyle)
	l.Extra = rawExtra(data, codes("2", "70", "62", "6", "370", "390", "420", "430"), false)
	if v, ok := value(data, "420"); ok {
		if c, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			l.SetTrueColorValue(c)
		}
	}
	if v, ok := value(data, "430"); ok {
		l.SetColorName(v)
	}
	// transparency: extended data AcCmTransparency (1071)
	l.Extra = takeXData(l.Extra, table.TransparencyApp, func(x [][2]string) {
		if v, ok := value(x, "1071"); ok {
			if t, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
				l.SetTransparencyValue(t)
			}
		}
	})
	return l, nil
}

// ParseStyle parses STYLE tables.
func ParseStyle(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	s := table.NewStyle("")
	name := ""
	for _, dt := range data {
		var err error
		switch dt[0] {
		case "2":
			name = dt[1]
		case "3":
			s.FontName = dt[1]
		case "4":
			s.BigFontName = dt[1]
		case "40":
			err = setFloat(dt, func(val float64) { s.FixedTextHeight = val })
		case "41":
			err = setFloat(dt, func(val float64) { s.WidthFactor = val })
		case "42":
			err = setFloat(dt, func(val float64) { s.LastHeightUsed = val })
		case "50":
			err = setFloat(dt, func(val float64) { s.ObliqueAngle = val })
		case "70":
			err = setInt(dt, func(val int) { s.Flags = val })
		case "71":
			err = setInt(dt, func(val int) { s.GenFlag = val })
		}
		if err != nil {
			return s, err
		}
	}
	s.SetName(name)
	return s, nil
}

// ParseView parses VIEW tables.
func ParseView(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	var name string
	for _, dt := range data {
		switch dt[0] {
		case "2":
			name = dt[1]
		}
	}
	v := table.NewView(name)
	v.Extra = rawExtra(data, codes("2"), false)
	return v, nil
}

// ParseUCS parses UCS tables.
func ParseUCS(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	var name string
	for _, dt := range data {
		switch dt[0] {
		case "2":
			name = dt[1]
		}
	}
	u := table.NewUCS(name)
	u.Extra = rawExtra(data, codes("2"), false)
	return u, nil
}

// ParseAppID parses APPID tables.
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

// ParseBlockRecord parses BLOCK_RECORD tables.
func ParseBlockRecord(d *drawing.Drawing, data [][2]string) (table.SymbolTable, error) {
	var name string
	b := table.NewBlockRecord("")
	for _, dt := range data {
		var err error
		switch dt[0] {
		case "2":
			name = dt[1]
		case "70":
			err = setInt(dt, func(val int) { b.InsUnits = val })
		case "280":
			err = setInt(dt, func(val int) { b.Explodable = val != 0 })
		case "281":
			err = setInt(dt, func(val int) { b.Scalable = val != 0 })
		}
		if err != nil {
			return b, err
		}
	}
	b.SetName(name)
	b.Extra = rawExtra(data, codes("2", "70", "280", "281", "340"), false)
	return b, nil
}

// BLOCKS

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
