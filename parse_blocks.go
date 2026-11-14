package dxf

import (
	"fmt"
	"strings"

	"github.com/yofu/dxf/block"
	"github.com/yofu/dxf/drawing"
)

// ParseBlocks parses BLOCKS section.
func ParseBlocks(d *drawing.Drawing, line int, data [][2]string) error {
	tmpdata := make([][2]string, 0)
	start := 0
	add := false           // skip ENDBLK and anything outside BLOCK ... ENDBLK
	var ended *block.Block // block whose ENDBLK is being read
	for i, dt := range data {
		if !add && ended != nil && dt[0] == "5" {
			ended.SetEndSourceHandle(dt[1])
			ended = nil
		}
		if dt[0] == entityTypeCode {
			switch strings.ToUpper(dt[1]) {
			case "BLOCK":
				add = true
				start = i + 1
				tmpdata = make([][2]string, 0)
				continue
			case "ENDBLK":
				ended = nil
				if add {
					if err := ParseBlock(d, tmpdata); err != nil {
						return fmt.Errorf("line %d: %w", line+2*start, err)
					}
					if name, ok := value(tmpdata, "2"); ok {
						ended, _ = d.Block(blockName(name))
					}
				}
				tmpdata = make([][2]string, 0)
				add = false
				continue
			}
		}
		if add {
			tmpdata = append(tmpdata, dt)
		}
	}
	if add && len(tmpdata) > 0 {
		if err := ParseBlock(d, tmpdata); err != nil {
			return fmt.Errorf("line %d: %w", line+2*start, err)
		}
	}
	return nil
}

// blockName converts the R12 names of the model space and paper space
// blocks ($MODEL_SPACE, $PAPER_SPACE).
func blockName(name string) string {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "$MODEL_SPACE":
		return "*Model_Space"
	case "$PAPER_SPACE":
		return "*Paper_Space"
	}
	return name
}

// ParseBlock parses each BLOCK, which starts with "0\nBLOCK\n" and ends with "0\nENDBLK\n".
// data contains the group codes after "0\nBLOCK\n": the BLOCK header followed by its entities.
// The parsed block replaces the block of the same name in the drawing, or is appended.
func ParseBlock(d *drawing.Drawing, data [][2]string) error {
	b := block.NewBlock("", "")
	var err error
	header := data
	for i, dt := range data {
		if dt[0] == entityTypeCode {
			header = data[:i]
			es, err := parseEntityList(d, data[i:])
			if err != nil {
				return err
			}
			b.Entities = es
			break
		}
	}
	setSourceHandle(b, header)
	for _, dt := range header {
		switch dt[0] {
		case "2":
			b.Name = blockName(dt[1])
		case "1":
			b.XrefPath = dt[1]
		case "4":
			b.Description = dt[1]
		case "8":
			layer, err := d.Layer(dt[1], false)
			if err == nil {
				b.SetLayer(layer)
			}
		case "10":
			err = setFloat(dt, func(val float64) { b.Coord[0] = val })
		case "20":
			err = setFloat(dt, func(val float64) { b.Coord[1] = val })
		case "30":
			err = setFloat(dt, func(val float64) { b.Coord[2] = val })
		case "70":
			err = setInt(dt, func(val int) { b.Flag = val })
		}
		if err != nil {
			return err
		}
	}
	bs := d.Sections[drawing.BLOCKS].(block.Blocks)
	for i, ob := range bs {
		if strings.EqualFold(ob.Name, b.Name) {
			bs[i] = b
			return nil
		}
	}
	d.Sections[drawing.BLOCKS] = bs.Add(b)
	return nil
}

// ENTITIES
