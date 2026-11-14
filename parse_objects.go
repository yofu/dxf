package dxf

import (
	"sort"
	"strconv"
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/handle"
	"github.com/yofu/dxf/object"
	"github.com/yofu/dxf/table"
)

// objRecord is a record of the OBJECTS section.
type objRecord struct {
	typ  string // upper case
	name string // as read
	data [][2]string
}

func (o objRecord) handle() string {
	h, _ := value(o.data, "5")
	return strings.TrimSpace(h)
}

// dictionaryItems returns the items (3: name, 350/360: handle) of a DICTIONARY.
func (o objRecord) dictionaryItems() map[string]string {
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
func (o objRecord) after(marker string) [][2]string {
	for i, dt := range o.data {
		if dt[0] == "100" && strings.TrimSpace(dt[1]) == marker {
			return o.data[i+1:]
		}
	}
	return nil
}

// ParseObjects parses OBJECTS section: groups (GROUP in the ACAD_GROUP
// dictionary), layouts (LAYOUT in the ACAD_LAYOUT dictionary), underlay
// definitions (PDFDEFINITION in the ACAD_PDFDEFINITIONS dictionary, ...)
// and image definitions (IMAGEDEF in the ACAD_IMAGE_DICT dictionary, and
// IMAGEDEF_REACTOR).
// Other objects are kept as read (object.Raw) and written again.
func ParseObjects(d *drawing.Drawing, line int, data [][2]string) error {
	var objects []objRecord
	for _, dt := range data {
		if dt[0] == "0" {
			objects = append(objects, objRecord{typ: strings.ToUpper(strings.TrimSpace(dt[1])), name: strings.TrimSpace(dt[1])})
			continue
		}
		if len(objects) > 0 {
			objects[len(objects)-1].data = append(objects[len(objects)-1].data, dt)
		}
	}
	byHandle := map[string]objRecord{}
	var root *objRecord
	for i, o := range objects {
		byHandle[o.handle()] = o
		if root == nil && o.typ == "DICTIONARY" {
			root = &objects[i]
		}
	}
	if root == nil {
		return nil
	}
	// objects this package creates keep the handles of the file
	modeled := map[string]bool{root.handle(): true}
	d.SetObjectSourceHandle("", root.handle())
	items := root.dictionaryItems()
	for _, key := range []string{"ACAD_GROUP", "ACAD_LAYOUT", "ACAD_PLOTSTYLENAME"} {
		if h, ok := items[key]; ok {
			modeled[h] = true
			d.SetObjectSourceHandle(key, h)
		}
	}
	if h, ok := items["ACAD_PLOTSTYLENAME"]; ok {
		if n, ok := byHandle[h].dictionaryItems()["Normal"]; ok {
			modeled[n] = true
			d.SetObjectSourceHandle("ACAD_PLOTSTYLENAME/Normal", n)
		}
	}
	if h, ok := items["ACAD_LAYOUT"]; ok {
		if err := parseLayouts(d, byHandle[h].dictionaryItems(), byHandle, modeled); err != nil {
			return err
		}
	}
	if h, ok := items["ACAD_GROUP"]; ok {
		if err := parseGroups(d, byHandle[h].dictionaryItems(), objects, modeled); err != nil {
			return err
		}
	}
	parseUnderlayDefinitions(d, items, byHandle, modeled)
	parseImageDefinitions(d, items, objects, byHandle, modeled)
	// the other objects are kept as read, with their handles
	rootKeys := map[string]string{}
	for name, h := range items {
		rootKeys[h] = name
	}
	for _, o := range objects {
		h := o.handle()
		if modeled[h] || h == "" {
			continue
		}
		data := make([][2]string, 0, len(o.data))
		skipped := false
		for _, dt := range o.data {
			if dt[0] == "5" && !skipped {
				skipped = true
				continue
			}
			data = append(data, dt)
		}
		d.AddRawObject(object.NewRaw(o.name, h, data), rootKeys[h])
	}
	return nil
}

func parseLayouts(d *drawing.Drawing, items map[string]string, byHandle map[string]objRecord, modeled map[string]bool) error {
	// BLOCK_RECORD handles in the file -> block names
	blocks := map[string]string{}
	for _, st := range d.Sections[drawing.TABLES].(table.Tables)[table.BLOCK_RECORD].Entries() {
		if r, ok := st.(*table.BlockRecord); ok && r.SourceHandle() != "" {
			blocks[r.SourceHandle()] = r.Name()
		}
	}
	type layout struct {
		name, block    string
		handle         string
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
		l := layout{name: name, handle: h, limMin: []float64{0, 0}, limMax: []float64{420, 297}}
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
			ol.SetSourceHandle(l.handle)
			modeled[l.handle] = true
		}
	}
	return nil
}

// parseGroups creates the groups in the order of the GROUP objects in the file.
func parseGroups(d *drawing.Drawing, items map[string]string, objects []objRecord, modeled map[string]bool) error {
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
		g.SetSourceHandle(o.handle())
		modeled[o.handle()] = true
	}
	return nil
}

// parseUnderlayDefinitions creates the underlay definitions of the
// dictionaries ACAD_PDFDEFINITIONS, ACAD_DWFDEFINITIONS and
// ACAD_DGNDEFINITIONS, and links the underlays read to them.
func parseUnderlayDefinitions(d *drawing.Drawing, items map[string]string, byHandle map[string]objRecord, modeled map[string]bool) {
	defs := map[int]*entity.UnderlayDefinition{} // by handle in the file
	for _, ff := range []string{entity.UNDERLAY_PDF, entity.UNDERLAY_DWF, entity.UNDERLAY_DGN} {
		key := "ACAD_" + ff + "DEFINITIONS"
		h, ok := items[key]
		if !ok || byHandle[h].typ != "DICTIONARY" {
			continue
		}
		modeled[h] = true
		d.SetObjectSourceHandle(key, h)
		dictItems := byHandle[h].dictionaryItems()
		for _, name := range sortedKeys(dictItems) {
			dh := dictItems[name]
			o, ok := byHandle[dh]
			if !ok || o.typ != ff+"DEFINITION" {
				continue
			}
			def := entity.NewUnderlayDefinition(ff, "", "")
			for _, dt := range o.after("AcDbUnderlayDefinition") {
				switch dt[0] {
				case "1":
					def.FileName = dt[1]
				case "2":
					def.Name = dt[1]
				}
			}
			def.SetSourceHandle(dh)
			if err := d.AddUnderlayDefinitionWithKey(def, name); err != nil {
				continue
			}
			modeled[dh] = true
			defs[handle.Parse(dh)] = def
		}
	}
	if len(defs) == 0 {
		return
	}
	link := func(es entity.Entities) {
		for _, e := range es {
			u, ok := e.(*entity.Underlay)
			if !ok || u.DefinitionHandle() == "" {
				continue
			}
			if def, ok := defs[handle.Parse(u.DefinitionHandle())]; ok {
				u.Definition = def
				u.SetDefinitionHandle("")
			}
		}
	}
	link(d.Entities())
	for _, b := range d.Blocks() {
		link(b.Entities)
	}
}

// parseImageDefinitions creates the image definitions of the dictionary
// ACAD_IMAGE_DICT and the IMAGEDEF_REACTOR objects, and links the images
// read to them.
func parseImageDefinitions(d *drawing.Drawing, items map[string]string, objects []objRecord, byHandle map[string]objRecord, modeled map[string]bool) {
	defs := map[int]*entity.ImageDefinition{} // by handle in the file
	if h, ok := items["ACAD_IMAGE_DICT"]; ok && byHandle[h].typ == "DICTIONARY" {
		modeled[h] = true
		d.SetObjectSourceHandle("ACAD_IMAGE_DICT", h)
		dictItems := byHandle[h].dictionaryItems()
		for _, name := range sortedKeys(dictItems) {
			dh := dictItems[name]
			o, ok := byHandle[dh]
			if !ok || o.typ != "IMAGEDEF" {
				continue
			}
			def := entity.NewImageDefinition("", 0, 0)
			for _, dt := range o.after("AcDbRasterImageDef") {
				switch dt[0] {
				case "1":
					def.FileName = dt[1]
				case "10", "20":
					setFloat(dt, func(v float64) { def.Size[int(dt[0][0]-'1')] = v })
				case "11", "21":
					setFloat(dt, func(v float64) { def.PixelSize[int(dt[0][0]-'1')] = v })
				case "280":
					setInt(dt, func(v int) { def.Loaded = v != 0 })
				case "281":
					setInt(dt, func(v int) { def.ResolutionUnits = v })
				}
			}
			def.SetSourceHandle(dh)
			if err := d.AddImageDefinitionWithKey(def, name); err != nil {
				continue
			}
			modeled[dh] = true
			defs[handle.Parse(dh)] = def
		}
	}
	reactors := map[int]*entity.ImageDefReactor{}
	reactorHandles := map[*entity.ImageDefReactor]string{} // as read
	for _, o := range objects {
		if o.typ == "IMAGEDEF_REACTOR" {
			r := &entity.ImageDefReactor{}
			r.SetSourceHandle(o.handle())
			reactors[handle.Parse(o.handle())] = r
			reactorHandles[r] = o.handle()
		}
	}
	if len(defs) == 0 && len(reactors) == 0 {
		return
	}
	link := func(es entity.Entities) {
		for _, e := range es {
			im, ok := e.(*entity.Image)
			if !ok || im.IsWipeout() {
				continue
			}
			if def, ok := defs[handle.Parse(im.DefinitionHandle())]; ok && im.DefinitionHandle() != "" {
				im.Definition = def
				im.SetDefinitionHandle("")
			}
			if im.Definition == nil {
				continue // the reactor is kept as read
			}
			if r, ok := reactors[handle.Parse(im.ReactorHandle())]; ok && im.ReactorHandle() != "" {
				im.SetReactor(r)
				modeled[reactorHandles[r]] = true
			}
		}
	}
	link(d.Entities())
	for _, b := range d.Blocks() {
		link(b.Entities)
	}
}
