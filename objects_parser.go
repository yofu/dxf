package dxf

import (
	"sort"
	"strconv"
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/table"
)

// object is a record of the OBJECTS section.
type object struct {
	typ  string
	data [][2]string
}

func (o object) handle() string {
	h, _ := value(o.data, "5")
	return strings.TrimSpace(h)
}

// dictionaryItems returns the items (3: name, 350/360: handle) of a DICTIONARY.
func (o object) dictionaryItems() map[string]string {
	items := map[string]string{}
	name := ""
	for _, dt := range o.data {
		switch dt[0] {
		case "3":
			name = dt[1]
		case "350", "360":
			if name != "" {
				items[name] = strings.TrimSpace(dt[1])
				name = ""
			}
		}
	}
	return items
}

// after returns the group codes after the subclass marker (100) marker.
func (o object) after(marker string) [][2]string {
	for i, dt := range o.data {
		if dt[0] == "100" && strings.TrimSpace(dt[1]) == marker {
			return o.data[i+1:]
		}
	}
	return nil
}

// ParseObjects parses OBJECTS section: groups (GROUP in the ACAD_GROUP
// dictionary) and layouts (LAYOUT in the ACAD_LAYOUT dictionary).
// Other objects are not kept.
func ParseObjects(d *drawing.Drawing, line int, data [][2]string) error {
	var objects []object
	for _, dt := range data {
		if dt[0] == "0" {
			objects = append(objects, object{typ: strings.ToUpper(strings.TrimSpace(dt[1]))})
			continue
		}
		if len(objects) > 0 {
			objects[len(objects)-1].data = append(objects[len(objects)-1].data, dt)
		}
	}
	byHandle := map[string]object{}
	var root *object
	for i, o := range objects {
		byHandle[o.handle()] = o
		if root == nil && o.typ == "DICTIONARY" {
			root = &objects[i]
		}
	}
	if root == nil {
		return nil
	}
	items := root.dictionaryItems()
	if h, ok := items["ACAD_LAYOUT"]; ok {
		if err := parseLayouts(d, byHandle[h].dictionaryItems(), byHandle); err != nil {
			return err
		}
	}
	if h, ok := items["ACAD_GROUP"]; ok {
		if err := parseGroups(d, byHandle[h].dictionaryItems(), objects); err != nil {
			return err
		}
	}
	return nil
}

func parseLayouts(d *drawing.Drawing, items map[string]string, byHandle map[string]object) error {
	// BLOCK_RECORD handles in the file -> block names
	blocks := map[string]string{}
	for _, st := range d.Sections[drawing.TABLES].(table.Tables)[table.BLOCK_RECORD].Entries() {
		if r, ok := st.(*table.BlockRecord); ok && r.SourceHandle != "" {
			blocks[r.SourceHandle] = r.Name()
		}
	}
	type layout struct {
		name, block    string
		tab            int
		limMin, limMax []float64
	}
	var layouts []layout
	for _, name := range sortedKeys(items) {
		h := items[name]
		o, ok := byHandle[h]
		if !ok || o.typ != "LAYOUT" {
			continue
		}
		l := layout{name: name, limMin: []float64{0, 0}, limMax: []float64{420, 297}}
		for _, dt := range o.after("AcDbLayout") {
			v := strings.TrimSpace(dt[1])
			switch dt[0] {
			case "1":
				l.name = dt[1]
			case "71":
				l.tab, _ = strconv.Atoi(v)
			case "10":
				l.limMin[0], _ = strconv.ParseFloat(v, 64)
			case "20":
				l.limMin[1], _ = strconv.ParseFloat(v, 64)
			case "11":
				l.limMax[0], _ = strconv.ParseFloat(v, 64)
			case "21":
				l.limMax[1], _ = strconv.ParseFloat(v, 64)
			case "330":
				l.block = blocks[v]
			}
		}
		if l.block != "" {
			layouts = append(layouts, l)
		}
	}
	// model space and the active layout first, then by tab order
	sort.SliceStable(layouts, func(i, j int) bool { return layouts[i].tab < layouts[j].tab })
	for pass := 0; pass < 2; pass++ {
		for _, l := range layouts {
			active := strings.EqualFold(l.block, "*Model_Space") || strings.EqualFold(l.block, "*Paper_Space")
			if (pass == 0) != active {
				continue
			}
			ol, err := d.DefineLayout(l.name, l.block, l.tab)
			if err != nil {
				return err
			}
			ol.LimMin, ol.LimMax = l.limMin, l.limMax
		}
	}
	return nil
}

// parseGroups creates the groups in the order of the GROUP objects in the file.
func parseGroups(d *drawing.Drawing, items map[string]string, objects []object) error {
	// entities by their handles in the file
	entities := map[string]entity.Entity{}
	add := func(es entity.Entities) {
		for _, e := range es {
			if s, ok := e.(interface{ SourceHandle() string }); ok && s.SourceHandle() != "" {
				entities[s.SourceHandle()] = e
			}
		}
	}
	add(d.Entities())
	for _, b := range d.Blocks() {
		add(b.Entities)
	}
	names := map[string]string{} // handle -> group name
	for name, h := range items {
		names[h] = name
	}
	for _, o := range objects {
		name, ok := names[o.handle()]
		if !ok || o.typ != "GROUP" {
			continue
		}
		desc := ""
		unnamed, selectable := false, true
		var es []entity.Entity
		for _, dt := range o.after("AcDbGroup") {
			v := strings.TrimSpace(dt[1])
			switch dt[0] {
			case "300":
				desc = dt[1]
			case "70":
				unnamed = v == "1"
			case "71":
				selectable = v != "0"
			case "340":
				if e, ok := entities[v]; ok {
					es = append(es, e)
				}
			}
		}
		g, err := d.Group(name, desc, es...)
		if err != nil {
			return err
		}
		g.Unnamed = unnamed
		g.SetSelectable(selectable)
	}
	return nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
