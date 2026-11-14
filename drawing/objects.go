package drawing

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/handle"
	"github.com/yofu/dxf/object"
	"github.com/yofu/dxf/table"
)

// AddObject adds an object to the OBJECTS section; with a key, it is also
// an item of the root dictionary.
func (d *Drawing) AddObject(o object.Object, key string) {
	d.addObject(o)
	if key != "" {
		d.dictionary.AddItem(key, o)
	}
}

// Objects returns the objects of the OBJECTS section.
func (d *Drawing) Objects() []object.Object {
	return d.Sections[OBJECTS].(object.Objects)
}

// RootObject returns the item key of the root dictionary (nil if none).
func (d *Drawing) RootObject(key string) object.Object {
	if o, ok := d.dictionary.Item(key).(object.Object); ok {
		return o
	}
	return nil
}

// ObjectByHandle returns the object with handle h: the handle read, or
// the handle written last for new objects.
func (d *Drawing) ObjectByHandle(h string) object.Object {
	v := handle.Parse(h)
	if v == 0 {
		return nil
	}
	for _, o := range d.Objects() {
		if s, ok := o.(interface{ SourceHandle() string }); ok && s.SourceHandle() != "" {
			if handle.Parse(s.SourceHandle()) == v {
				return o
			}
		} else if o.Handle() == v {
			return o
		}
	}
	return nil
}

// itemObject returns the object of a dictionary item.
func (d *Drawing) itemObject(it object.DictItem) handle.Handler {
	if it.Object != nil {
		return it.Object
	}
	return d.ObjectByHandle(it.Handle)
}

// extended is implemented by entities and table entries with an extension
// dictionary.
type extended interface {
	handle.Handler
	ExtensionDictionaryHandle() string
	ExtensionDictionary() handle.Handler
	SetExtensionDictionary(handle.Handler)
}

// ExtensionDictionary returns the extension dictionary of an entity or a
// table entry, creating it if create is true (nil without error if it does
// not exist).
func (d *Drawing) ExtensionDictionary(owner interface{}, create bool) (*object.DictionaryObject, error) {
	x, ok := owner.(extended)
	if !ok {
		return nil, fmt.Errorf("%T has no extension dictionary", owner)
	}
	if dict, ok := x.ExtensionDictionary().(*object.DictionaryObject); ok {
		return dict, nil
	}
	if h := x.ExtensionDictionaryHandle(); h != "" {
		if dict, ok := d.ObjectByHandle(h).(*object.DictionaryObject); ok {
			return dict, nil
		}
	}
	if !create {
		return nil, nil
	}
	dict := object.NewDictionaryObject(x, true)
	d.addObject(dict)
	x.SetExtensionDictionary(dict)
	return dict, nil
}

// subDictionary returns the dictionary key of dict, creating it.
func (d *Drawing) subDictionary(dict *object.DictionaryObject, key string, hardOwner bool) *object.DictionaryObject {
	if it, ok := dict.Item(key); ok {
		if sub, ok := d.itemObject(it).(*object.DictionaryObject); ok {
			return sub
		}
	}
	sub := object.NewDictionaryObject(dict, hardOwner)
	d.addObject(sub)
	dict.SetItem(key, sub)
	return sub
}

// SetXRecord stores tags as the XRECORD key of the extension dictionary of
// owner (an entity or a table entry), replacing an XRECORD with the same key.
func (d *Drawing) SetXRecord(owner interface{}, key string, tags [][2]string) (*object.XRecord, error) {
	dict, err := d.ExtensionDictionary(owner, true)
	if err != nil {
		return nil, err
	}
	if x, err := d.XRecord(owner, key); err == nil && x != nil {
		x.SetTags(tags)
		return x, nil
	}
	x := object.NewXRecord(dict, tags)
	d.addObject(x)
	dict.SetItem(key, x)
	return x, nil
}

// XRecord returns the XRECORD key of the extension dictionary of owner
// (nil if there is none).
func (d *Drawing) XRecord(owner interface{}, key string) (*object.XRecord, error) {
	dict, err := d.ExtensionDictionary(owner, false)
	if err != nil || dict == nil {
		return nil, err
	}
	it, ok := dict.Item(key)
	if !ok {
		return nil, nil
	}
	x, _ := d.itemObject(it).(*object.XRecord)
	return x, nil
}

// variables returns the root dictionary AcDbVariableDictionary.
func (d *Drawing) variables(create bool) *object.DictionaryObject {
	if dict, ok := d.RootObject("AcDbVariableDictionary").(*object.DictionaryObject); ok {
		return dict
	}
	if !create {
		return nil
	}
	dict := object.NewDictionaryObject(d.dictionary, false)
	d.AddObject(dict, "AcDbVariableDictionary")
	return dict
}

// Variable returns the value of the DICTIONARYVAR name (e.g. XCLIPFRAME)
// of the dictionary AcDbVariableDictionary.
func (d *Drawing) Variable(name string) (string, bool) {
	dict := d.variables(false)
	if dict == nil {
		return "", false
	}
	it, ok := dict.Item(name)
	if !ok {
		return "", false
	}
	if v, ok := d.itemObject(it).(*object.DictionaryVar); ok {
		return v.Value(), true
	}
	return "", false
}

// SetVariable sets the DICTIONARYVAR name of the dictionary
// AcDbVariableDictionary.
func (d *Drawing) SetVariable(name, value string) {
	dict := d.variables(true)
	if it, ok := dict.Item(name); ok {
		if v, ok := d.itemObject(it).(*object.DictionaryVar); ok {
			v.SetValue(value)
			return
		}
	}
	v := object.NewDictionaryVar(dict, value)
	d.addObject(v)
	dict.SetItem(name, v)
}

// blockRecord returns the BLOCK_RECORD of block ("" for model space).
func (d *Drawing) blockRecord(block string) (*table.BlockRecord, error) {
	if block == "" {
		block = "*Model_Space"
	}
	records := d.Sections[TABLES].(table.Tables)[table.BLOCK_RECORD]
	r, err := records.Contains(block)
	if err != nil {
		// the records of model and paper space are created when writing
		if strings.EqualFold(block, "*Model_Space") || strings.EqualFold(block, "*Paper_Space") {
			br := table.NewBlockRecord(block)
			records.Add(br)
			return br, nil
		}
		return nil, err
	}
	return r.(*table.BlockRecord), nil
}

// sortEntsTable returns the SORTENTSTABLE of the block record (nil if none).
func (d *Drawing) sortEntsTable(rec *table.BlockRecord, create bool) (*object.SortEntsTable, error) {
	dict, err := d.ExtensionDictionary(rec, create)
	if err != nil || dict == nil {
		return nil, err
	}
	if it, ok := dict.Item("ACAD_SORTENTS"); ok {
		if s, ok := d.itemObject(it).(*object.SortEntsTable); ok {
			return s, nil
		}
	}
	if !create {
		return nil, nil
	}
	s := object.NewSortEntsTable(dict, rec)
	d.addObject(s)
	dict.SetItem("ACAD_SORTENTS", s)
	d.ensureClass([]string{"SORTENTSTABLE", "AcDbSortentsTable", "ObjectDBX Classes", "0", "0"})
	return s, nil
}

// SetDrawOrder sets the draw order of the entities of block ("" for the
// model space): es are drawn in the order given, after the other entities
// (SORTENTSTABLE, as AutoCAD's DRAWORDER).
func (d *Drawing) SetDrawOrder(block string, es ...entity.Entity) error {
	rec, err := d.blockRecord(block)
	if err != nil {
		return err
	}
	current := d.DrawOrder(block)
	s, err := d.sortEntsTable(rec, true)
	if err != nil {
		return err
	}
	s.Block = rec
	in := map[entity.Entity]bool{}
	for _, e := range es {
		in[e] = true
	}
	var order []handle.Handler
	for _, e := range current {
		if !in[e] {
			order = append(order, e)
		}
	}
	for _, e := range es {
		order = append(order, e)
	}
	s.SetOrder(order...)
	return nil
}

// DrawOrder returns the entities of block ("" for the model space) in
// draw order: by the sort handles of its SORTENTSTABLE, or as stored.
func (d *Drawing) DrawOrder(block string) entity.Entities {
	var es entity.Entities
	if block == "" || strings.EqualFold(block, "*Model_Space") {
		for _, e := range d.Entities() {
			if sp, ok := e.(entity.SpaceSelector); !ok || !sp.InPaperSpace() {
				es = append(es, e)
			}
		}
	} else if b, err := d.Block(block); err == nil {
		es = append(es, b.Entities...)
	}
	rec, err := d.blockRecord(block)
	if err != nil {
		return es
	}
	s, _ := d.sortEntsTable(rec, false)
	if s == nil {
		return es
	}
	if order := s.Order(); order != nil {
		var out entity.Entities
		for _, h := range order {
			if e, ok := h.(entity.Entity); ok {
				out = append(out, e)
			}
		}
		return out
	}
	// sort handles read: entities without one sort by their handle
	sortHandles := map[int]int{}
	for _, en := range s.Entries {
		sortHandles[handle.Parse(en.Handle)] = handle.Parse(en.SortHandle)
	}
	key := func(e entity.Entity) int {
		h := 0
		if sh, ok := e.(interface{ SourceHandle() string }); ok {
			h = handle.Parse(sh.SourceHandle())
		}
		if v, ok := sortHandles[h]; ok {
			return v
		}
		return h
	}
	out := append(entity.Entities{}, es...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && key(out[j-1]) > key(out[j]); j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}

// ClipBlockReference sets the clipping boundary (XCLIP) of a block
// reference: vertices in WCS (XY plane; two vertices: rectangle). It is
// stored in block coordinates (SPATIAL_FILTER), as ezdxf does.
func (d *Drawing) ClipBlockReference(ins *entity.Insert, vertices [][]float64) (*object.SpatialFilter, error) {
	if len(vertices) < 2 {
		return nil, errors.New("clipping boundary needs 2 or more vertices")
	}
	if len(vertices) == 2 {
		a, b := vertices[0], vertices[1]
		vertices = [][]float64{{a[0], a[1]}, {b[0], a[1]}, {b[0], b[1]}, {a[0], b[1]}}
	}
	base := []float64{0, 0, 0}
	if blk, err := d.Block(ins.BlockName); err == nil {
		base = blk.Coord
	}
	sx, sy := 1.0, 1.0
	if len(ins.Scale) >= 2 {
		sx, sy = ins.Scale[0], ins.Scale[1]
	}
	if sx == 0 || sy == 0 {
		return nil, errors.New("block reference with scale 0")
	}
	rad := ins.Rotation * math.Pi / 180
	var bv [][]float64
	for _, v := range vertices {
		x, y := v[0]-ins.Coord[0], v[1]-ins.Coord[1]
		x, y = x*math.Cos(rad)+y*math.Sin(rad), -x*math.Sin(rad)+y*math.Cos(rad)
		bv = append(bv, []float64{x/sx + base[0], y/sy + base[1]})
	}
	xdict, err := d.ExtensionDictionary(ins, true)
	if err != nil {
		return nil, err
	}
	filters := d.subDictionary(xdict, "ACAD_FILTER", true)
	f := object.NewSpatialFilter(filters, bv)
	d.addObject(f)
	filters.SetItem("SPATIAL", f)
	return f, nil
}

// BlockReferenceClipping returns the clipping boundary of a block
// reference (nil if it is not clipped).
func (d *Drawing) BlockReferenceClipping(ins *entity.Insert) *object.SpatialFilter {
	xdict, _ := d.ExtensionDictionary(ins, false)
	if xdict == nil {
		return nil
	}
	it, ok := xdict.Item("ACAD_FILTER")
	if !ok {
		return nil
	}
	filters, ok := d.itemObject(it).(*object.DictionaryObject)
	if !ok {
		return nil
	}
	if it, ok = filters.Item("SPATIAL"); !ok {
		return nil
	}
	f, _ := d.itemObject(it).(*object.SpatialFilter)
	return f
}

// SetRasterVariables sets the display of images (ACAD_IMAGE_VARS): frame
// (0 no, 1 yes), quality (0 draft, 1 high) and units (0 none, 1 mm, ...).
func (d *Drawing) SetRasterVariables(frame, quality, units int) {
	d.dictionary.RemoveItem("ACAD_IMAGE_VARS")
	v := object.NewRasterVariables(d.dictionary, frame, quality, units)
	d.AddObject(v, "ACAD_IMAGE_VARS")
	d.ensureClass([]string{"RASTERVARIABLES", "AcDbRasterVariables", "ISM", "0", "0"})
}

// SetWipeoutVariables sets the frame of wipeouts (ACAD_WIPEOUT_VARS): 0 not
// displayed, 1 displayed and plotted, 2 displayed, not plotted.
func (d *Drawing) SetWipeoutVariables(frame int) {
	d.dictionary.RemoveItem("ACAD_WIPEOUT_VARS")
	v := object.NewWipeoutVariables(d.dictionary, frame)
	d.AddObject(v, "ACAD_WIPEOUT_VARS")
	d.ensureClass([]string{"WIPEOUTVARIABLES", "AcDbWipeoutVariables", "WipeOut", "0", "0"})
}
