// Package dxf is a DXF(Drawing Exchange Format) library for golang.
// ACAD2000(AC1015), ASCII format is only supported.
// http://www.autodesk.com/techpubs/autocad/acad2000/dxf/index.htm
package dxf

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yofu/dxf/codepage"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/handle"
	"github.com/yofu/dxf/table"
)

// Default values.
var (
	DefaultColor    = color.White
	DefaultLineType = table.LT_CONTINUOUS
)

// NewDrawing creates a drawing.
func NewDrawing() *drawing.Drawing {
	return drawing.New()
}

// Create drawing from file
func FromFile(fn string) (*drawing.Drawing, error) {
	var err error
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return FromReader(f)
}

// Create drawing from string
func FromStringData(d string) (*drawing.Drawing, error) {
	sr := strings.NewReader(d)
	return FromReader(sr)
}

// FromFileWithCodePage creates a drawing from file, using codePage
// (e.g. "ANSI_932" for Shift_JIS) for text if the file has no $DWGCODEPAGE.
func FromFileWithCodePage(fn string, codePage string) (*drawing.Drawing, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return FromReaderWithCodePage(f, codePage)
}

// Main logic to create a drawing.
//
// Text is decoded according to $DWGCODEPAGE (AC1018 or earlier) or as UTF-8
// (AC1021 or later), and \U+XXXX / \M+nXXXX escapes are converted.
// If the file has no $DWGCODEPAGE, text is read as UTF-8 if valid,
// otherwise as ANSI_1252. Use FromReaderWithCodePage for files that
// omit $DWGCODEPAGE but are written in another code page (e.g. Shift_JIS).
func FromReader(r io.Reader) (*drawing.Drawing, error) {
	return FromReaderWithCodePage(r, "")
}

// FromReaderWithCodePage creates a drawing from r, using codePage
// (e.g. "ANSI_932" for Shift_JIS) for text if the file has no $DWGCODEPAGE.
// The code page is kept in Header().CodePage and used when the drawing is written.
func FromReaderWithCodePage(r io.Reader, codePage string) (*drawing.Drawing, error) {
	src := newPairSource(r)
	d := NewDrawing()
	var code, value string
	parsers := []func(*drawing.Drawing, int, [][2]string) error{
		ParseHeader,
		ParseClasses,
		ParseTables,
		ParseBlocks,
		ParseEntities,
		ParseObjects,
	}
	data := make([][2]string, 0)
	setparser := false
	var parser func(*drawing.Drawing, int, [][2]string) error
	line := 0
	startline := 0

	// text decoding state, taken from the HEADER section
	section := ""
	variable := ""
	unicode := false
	cp := codePage
	declared := false
	decode := func(raw string) string {
		if !unicode {
			raw = codepage.Decode(raw, cp)
		}
		return codepage.Unescape(raw)
	}
	maxHandle := 0
	finish := func() {
		// new objects get handles above the handles of the file
		d.SetHandleBase(maxHandle + 1)
		h := d.Header()
		if !declared && codePage != "" {
			h.CodePage = codePage
		}
		// current layer and text style from $CLAYER and $TEXTSTYLE
		if l, ok := d.Layers[h.CLayer]; ok {
			d.CurrentLayer = l
		}
		if st, ok := findStyle(d, h.TextStyle); ok {
			d.CurrentStyle = st
		}
	}
	defer finish()

	for {
		c, raw, err := src.next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return d, err
		}
		code = c
		line = src.line()
		if code == "999" { // comment
			continue
		}
		value = decode(raw)
		if (code == "5" || code == "105") && section != "HEADER" { // not $HANDSEED
			if h := handle.Parse(value); h > maxHandle {
				maxHandle = h
			}
		}
		if section == "HEADER" {
			switch code {
			case "9":
				variable = strings.TrimSpace(value)
			case "1":
				if variable == "$ACADVER" {
					unicode = codepage.IsUnicodeVersion(value)
				}
			case "3":
				if variable == "$DWGCODEPAGE" {
					cp = strings.TrimSpace(value)
					declared = true
				}
			}
		}
		if setparser {
			if code != "2" {
				return d, fmt.Errorf("line %d: invalid group code: %s", line, code)
			}
			section = strings.ToUpper(strings.TrimSpace(value))
			ind := drawing.SectionTypeValue(section)
			if ind < 0 {
				return d, fmt.Errorf("line %d: unknown section name: %s", line, value)
			}
			parser = parsers[ind]
			startline = line + 1
			setparser = false
			continue
		}
		if code == "0" {
			switch strings.ToUpper(strings.TrimSpace(value)) {
			case "EOF":
				return d, nil
			case "SECTION":
				setparser = true
				continue
			case "ENDSEC":
				if parser == nil {
					return d, fmt.Errorf("line %d: ENDSEC without SECTION", line)
				}
				err := parser(d, startline, data)
				if err != nil {
					return d, err
				}
				data = make([][2]string, 0)
				startline = line + 1
				section = ""
				continue
			}
		}
		data = append(data, [2]string{code, value})
	}
	if len(data) > 0 && parser != nil {
		err := parser(d, startline, data)
		if err != nil {
			return d, err
		}
	}
	return d, nil
}

// Open is deprecated, please use FromFile
func Open(filename string) (*drawing.Drawing, error) {
	return FromFile(filename)
}

// ColorIndex converts RGB value to corresponding color number.
func ColorIndex(cl []int) color.ColorNumber {
	minind := 0
	minval := 1000000
	for i, c := range color.ColorRGB {
		tmpval := 0
		for j := 0; j < 3; j++ {
			tmpval += (cl[j] - int(c[j])) * (cl[j] - int(c[j]))
		}
		if tmpval < minval {
			minind = i
			minval = tmpval
			if minval == 0 {
				break
			}
		}
	}
	return color.ColorNumber(minind)
}

// IndexColor converts color number to RGB value.
func IndexColor(index uint8) []uint8 {
	return color.ColorRGB[index]
}
