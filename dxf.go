// Package dxf reads and writes DXF (Drawing Exchange Format) files.
//
// Drawings are written in the structure of AutoCAD 2000 (AC1015), as ASCII
// or binary DXF. Files of AutoCAD R12 and later, ASCII or binary, can be
// read; data not interpreted by this package is kept when a drawing read
// is written again. See the README for an overview.
//
// The drawing model is in package drawing (sections, tables, blocks,
// entities and objects), entities in package entity. This package has the
// readers (FromFile, FromReader, ...) and editing functions that need the
// reader (Copy, CopyEntities, Explode).
package dxf

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/yofu/dxf/codepage"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/handle"
	"github.com/yofu/dxf/table"
)

// Default values. DefaultLineType is a template: drawings have their own
// linetypes (Drawing.LineType); layers refer to linetypes by name when
// written.
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
	d, _, err := readDXF(r, readOptions{codePage: codePage})
	return d, err
}

// readOptions are the options of readDXF.
type readOptions struct {
	codePage string
	// onEntity receives the entities of the ENTITIES section one by one
	// instead of adding them to the drawing (streaming).
	onEntity func(*drawing.Drawing, entity.Entity) error
	// lenient reads damaged files: errors are collected and reading goes on.
	lenient bool
}

// readDXF reads a drawing; with lenient, it returns the problems found.
func readDXF(r io.Reader, o readOptions) (*drawing.Drawing, []error, error) {
	codePage := o.codePage
	var problems []error
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

	// entities one by one (streaming, lenient reading)
	chunked := o.onEntity != nil
	flushEntities := func() error {
		if len(data) == 0 {
			return nil
		}
		chunk := data
		data = make([][2]string, 0)
		es, err := parseEntityList(d, chunk)
		if err != nil {
			if !o.lenient {
				return fmt.Errorf("line %d: %w", startline, err)
			}
			problems = append(problems, fmt.Errorf("line %d: %w", startline, err))
			if u, ok := parseUnknown(d, chunk); ok {
				es = entity.Entities{u}
			}
		}
		for _, e := range es {
			if o.onEntity != nil {
				if err := o.onEntity(d, e); err != nil {
					return err
				}
			} else {
				d.AddEntity(e)
			}
		}
		return nil
	}
	subEntity := map[string]bool{"VERTEX": true, "SEQEND": true, "ATTRIB": true}
	// flushOne parses the chunk of one entity (lenient reading)
	flushOne := func(chunk [][2]string) {
		es, err := parseEntityList(d, chunk)
		if err != nil {
			problems = append(problems, err)
			if u, ok := parseUnknown(d, chunk); ok {
				es = entity.Entities{u}
			}
		}
		for _, e := range es {
			d.AddEntity(e)
		}
	}

	for {
		c, raw, err := src.next()
		if err == io.EOF {
			break
		}
		if err != nil {
			if o.lenient {
				problems = append(problems, err)
				break
			}
			return d, problems, err
		}
		code = c
		line = src.line()
		if code == "999" { // comment
			d.Comments = append(d.Comments, decode(raw))
			continue
		}
		if o.lenient {
			if _, err := strconv.Atoi(strings.TrimSpace(code)); err != nil {
				problems = append(problems, fmt.Errorf("line %d: invalid group code %q", line, code))
				continue
			}
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
				if o.lenient {
					problems = append(problems, fmt.Errorf("line %d: section without name", line))
					setparser = false
					continue
				}
				return d, problems, fmt.Errorf("line %d: invalid group code: %s", line, code)
			}
			section = strings.ToUpper(strings.TrimSpace(value))
			ind := drawing.SectionTypeValue(section)
			if ind < 0 {
				// kept as read (e.g. ACDSDATA, THUMBNAILIMAGE)
				name := strings.TrimSpace(value)
				parser = func(d *drawing.Drawing, _ int, data [][2]string) error {
					d.AddRawSection(name, data)
					return nil
				}
			} else {
				parser = parsers[ind]
			}
			startline = line + 1
			setparser = false
			continue
		}
		if code == "0" {
			switch strings.ToUpper(strings.TrimSpace(value)) {
			case "EOF":
				return d, problems, nil
			case "SECTION":
				setparser = true
				continue
			case "ENDSEC":
				if parser == nil {
					if o.lenient {
						problems = append(problems, fmt.Errorf("line %d: ENDSEC without SECTION", line))
						continue
					}
					return d, problems, fmt.Errorf("line %d: ENDSEC without SECTION", line)
				}
				var err error
				if section == "ENTITIES" && chunked {
					err = flushEntities()
				} else {
					err = parser(d, startline, data)
				}
				if err != nil && o.lenient && section == "ENTITIES" {
					// entity by entity: those that cannot be parsed are kept as read
					err = lenientEntities(data, flushOne)
				}
				if err != nil {
					if !o.lenient {
						return d, problems, err
					}
					problems = append(problems, err)
				}
				parser = nil
				data = make([][2]string, 0)
				startline = line + 1
				section = ""
				continue
			}
		}
		if section == "ENTITIES" && chunked && code == "0" && !subEntity[strings.ToUpper(strings.TrimSpace(value))] {
			if err := flushEntities(); err != nil {
				return d, problems, err
			}
			startline = line
		}
		data = append(data, [2]string{code, value})
	}
	if len(data) > 0 && parser != nil {
		var err error
		if section == "ENTITIES" && chunked {
			err = flushEntities()
		} else {
			err = parser(d, startline, data)
		}
		if err != nil && o.lenient && section == "ENTITIES" {
			err = lenientEntities(data, flushOne)
		}
		if err != nil {
			if !o.lenient {
				return d, problems, err
			}
			problems = append(problems, err)
		}
	}
	if o.lenient && section != "" {
		problems = append(problems, fmt.Errorf("section %s not closed (no ENDSEC)", section))
	}
	return d, problems, nil
}

// lenientEntities parses the ENTITIES section entity by entity.
func lenientEntities(data [][2]string, one func([][2]string)) error {
	for _, chunk := range lenientEntitiesSplit(data) {
		one(chunk)
	}
	return nil
}

// lenientEntitiesSplit splits the ENTITIES section into the chunks of the
// entities (with their vertices and attributes).
func lenientEntitiesSplit(data [][2]string) [][][2]string {
	sub := map[string]bool{"VERTEX": true, "SEQEND": true, "ATTRIB": true}
	var chunks [][][2]string
	var cur [][2]string
	for _, dt := range data {
		if dt[0] == "0" && !sub[strings.ToUpper(strings.TrimSpace(dt[1]))] && len(cur) > 0 {
			chunks = append(chunks, cur)
			cur = nil
		}
		cur = append(cur, dt)
	}
	if len(cur) > 0 {
		chunks = append(chunks, cur)
	}
	return chunks
}

// IterEntities reads a DXF file and calls f with each entity of the
// ENTITIES section in turn, without keeping them in the drawing (as
// ezdxf's iterdxf): for large files. The drawing passed to f has the
// header, tables and blocks read before the entities.
func IterEntities(r io.Reader, f func(*drawing.Drawing, entity.Entity) error) error {
	_, _, err := readDXF(r, readOptions{onEntity: f})
	return err
}

// Recover reads a damaged DXF file as far as possible (as ezdxf's recover):
// invalid group codes are skipped, entities that cannot be parsed are kept
// as read (entity.Unknown), sections with errors keep what was read, and a
// missing ENDSEC or EOF is accepted. It returns the problems found.
func Recover(r io.Reader) (*drawing.Drawing, []error, error) {
	return readDXF(r, readOptions{lenient: true})
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
