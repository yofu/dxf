// Package dxf is a DXF(Drawing Exchange Format) library for golang.
// ACAD2000(AC1015), ASCII format is only supported.
// http://www.autodesk.com/techpubs/autocad/acad2000/dxf/index.htm
package dxf

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/drawing"
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

// Main logic to create a drawing
func FromReader(r io.Reader) (*drawing.Drawing, error) {
	var err error
	scanner := bufio.NewScanner(r)
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
	for scanner.Scan() {
		line++
		if line%2 == 1 {
			code = strings.TrimSpace(scanner.Text())
			if err != nil {
				return d, err
			}
		} else {
			value = scanner.Text()
			if setparser {
				if code != "2" {
					return d, fmt.Errorf("line %d: invalid group code: %s", line, code)
				}
				ind := drawing.SectionTypeValue(strings.ToUpper(value))
				if ind < 0 {
					return d, fmt.Errorf("line %d: unknown section name: %s", line, value)
				}
				parser = parsers[ind]
				startline = line + 1
				setparser = false
			} else {
				if code == "0" {
					switch strings.ToUpper(value) {
					case "EOF":
						return d, nil
					case "SECTION":
						setparser = true
					case "ENDSEC":
						err := parser(d, startline, data)
						if err != nil {
							return d, err
						}
						data = make([][2]string, 0)
						startline = line + 1
					default:
						data = append(data, [2]string{code, scanner.Text()})
					}
				} else {
					data = append(data, [2]string{code, scanner.Text()})
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return d, err
	}
	if len(data) > 0 {
		err := parser(d, startline, data)
		if err != nil {
			return d, err
		}
	}
	return d, nil
}

// Deprecated in favor of FromFile
func Open(filename string) (*drawing.Drawing, error) {
	var err error
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return FromReader(f)
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
