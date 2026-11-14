package object

import (
	"sort"
	"strconv"
	"strings"

	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Wrap returns the typed object of a raw object read (XRECORD, DICTIONARY,
// DICTIONARYVAR, SPATIAL_FILTER, IDBUFFER, LAYER_FILTER, FIELDLIST, FIELD,
// GEODATA, SUN, MATERIAL, VISUALSTYLE, PLOTSETTINGS, RASTERVARIABLES,
// WIPEOUTVARIABLES), or r itself.
func Wrap(r *Raw) Object {
	switch strings.ToUpper(r.Type) {
	case "XRECORD":
		return &XRecord{r}
	case "DICTIONARY":
		return &DictionaryObject{r}
	case "DICTIONARYVAR":
		return &DictionaryVar{r}
	case "SPATIAL_FILTER":
		return &SpatialFilter{r}
	case "IDBUFFER", "LAYER_FILTER", "FIELDLIST":
		return &IDBuffer{r}
	case "FIELD":
		return &Field{r}
	case "GEODATA":
		return &GeoData{r}
	case "SUN":
		return &Sun{r}
	case "MATERIAL", "VISUALSTYLE", "PLOTSETTINGS":
		return &Named{r}
	case "RASTERVARIABLES", "WIPEOUTVARIABLES":
		return &ImageVariables{r}
	}
	return r
}

// DictItem is an item of a dictionary.
type DictItem struct {
	Key    string
	Handle string         // as read ("" for objects added)
	Object handle.Handler // object added (nil for items read)
}

// DictionaryObject is a DICTIONARY kept as its group codes, with its items.
type DictionaryObject struct{ *Raw }

// NewDictionaryObject creates a dictionary owned by owner; hardOwner
// makes its items owned by it (280 = 1, items written with 360).
func NewDictionaryObject(owner handle.Handler, hardOwner bool) *DictionaryObject {
	h := "0"
	if hardOwner {
		h = "1"
	}
	return &DictionaryObject{newObject("DICTIONARY", owner, [2]string{"100", "AcDbDictionary"}, [2]string{"280", h}, [2]string{"281", "1"})}
}

// Items returns the items of the dictionary.
func (d *DictionaryObject) Items() []DictItem {
	var items []DictItem
	key := ""
	for i, p := range d.Data {
		switch p[0] {
		case "3":
			key = p[1]
		case "350", "360":
			if key != "" {
				items = append(items, DictItem{Key: key, Handle: strings.TrimSpace(p[1]), Object: d.refs[i]})
				key = ""
			}
		}
	}
	return items
}

// Item returns the item key (ok false if there is none).
func (d *DictionaryObject) Item(key string) (DictItem, bool) {
	for _, it := range d.Items() {
		if it.Key == key {
			return it, true
		}
	}
	return DictItem{}, false
}

// SetItem sets the item key to o (replacing an item with the same key).
func (d *DictionaryObject) SetItem(key string, o handle.Handler) {
	d.RemoveItem(key)
	code := "350"
	if d.Int("280", 0) == 1 {
		code = "360"
	}
	d.Data = append(d.Data, [2]string{"3", key})
	d.AppendRef(code, o)
}

// RemoveItem removes the item key.
func (d *DictionaryObject) RemoveItem(key string) {
	var data [][2]string
	refs := map[int]handle.Handler{}
	skip := false
	for i, p := range d.Data {
		if p[0] == "3" && p[1] == key {
			skip = true
			continue
		}
		if skip && (p[0] == "350" || p[0] == "360") {
			skip = false
			continue
		}
		if h, ok := d.refs[i]; ok {
			refs[len(data)] = h
		}
		data = append(data, p)
	}
	d.Data, d.refs = data, refs
}

// XRecord is an XRECORD: arbitrary group codes (Tags) stored in a
// dictionary (e.g. an extension dictionary).
type XRecord struct{ *Raw }

// NewXRecord creates an XRECORD owned by owner with tags.
func NewXRecord(owner handle.Handler, tags [][2]string) *XRecord {
	x := &XRecord{newObject("XRECORD", owner, [2]string{"100", "AcDbXrecord"}, [2]string{"280", "1"})}
	x.Data = append(x.Data, tags...)
	return x
}

func (x *XRecord) tagStart() int {
	i := x.bodyStart()
	if i < len(x.Data) && x.Data[i][0] == "280" {
		i++ // duplicate record cloning flag
	}
	return i
}

// Tags returns the group codes of the record.
func (x *XRecord) Tags() [][2]string {
	return append([][2]string{}, x.Data[x.tagStart():]...)
}

// SetTags replaces the group codes of the record.
func (x *XRecord) SetTags(tags [][2]string) {
	x.Data = append(x.Data[:x.tagStart():x.tagStart()], tags...)
	for i := range x.refs {
		if i >= x.tagStart() {
			delete(x.refs, i)
		}
	}
}

// DictionaryVar is a DICTIONARYVAR: a named value of the dictionary
// AcDbVariableDictionary (e.g. XCLIPFRAME).
type DictionaryVar struct{ *Raw }

// NewDictionaryVar creates a DICTIONARYVAR owned by owner.
func NewDictionaryVar(owner handle.Handler, value string) *DictionaryVar {
	return &DictionaryVar{newObject("DICTIONARYVAR", owner, [2]string{"100", "DictionaryVariables"}, [2]string{"280", "0"}, [2]string{"1", value})}
}

// Value returns the value (code 1).
func (v *DictionaryVar) Value() string { s, _ := v.Raw.Value("1"); return s }

// SetValue sets the value.
func (v *DictionaryVar) SetValue(s string) { v.Set("1", s) }

// SpatialFilter is a SPATIAL_FILTER: the clipping boundary of a block
// reference (XCLIP), in the dictionary ACAD_FILTER of its extension
// dictionary.
type SpatialFilter struct{ *Raw }

// identity3x4 is the identity matrix as stored in SPATIAL_FILTER: 3
// columns of 4 values (the 4th row is the translation).
var identity3x4 = []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0}

// NewSpatialFilter creates a SPATIAL_FILTER with boundary vertices (block
// coordinates; two vertices: rectangle) and identity matrices.
func NewSpatialFilter(owner handle.Handler, vertices [][]float64) *SpatialFilter {
	data := [][2]string{{"100", "AcDbFilter"}, {"100", "AcDbSpatialFilter"}, {"70", strconv.Itoa(len(vertices))}}
	for _, v := range vertices {
		data = append(data, [2]string{"10", ftoa(v[0])}, [2]string{"20", ftoa(v[1])})
	}
	data = append(data, [2]string{"210", "0"}, [2]string{"220", "0"}, [2]string{"230", "1"},
		[2]string{"11", "0"}, [2]string{"21", "0"}, [2]string{"31", "0"}, [2]string{"71", "1"}, [2]string{"72", "0"}, [2]string{"73", "0"})
	for k := 0; k < 2; k++ {
		for _, v := range identity3x4 {
			data = append(data, [2]string{"40", ftoa(v)})
		}
	}
	return &SpatialFilter{newObject("SPATIAL_FILTER", owner, data...)}
}

func ftoa(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// Boundary returns the boundary vertices (block coordinates, before the
// inverse insert matrix).
func (s *SpatialFilter) Boundary() [][]float64 {
	var pts [][]float64
	for _, p := range s.body() {
		v, _ := strconv.ParseFloat(strings.TrimSpace(p[1]), 64)
		switch p[0] {
		case "10":
			pts = append(pts, []float64{v, 0})
		case "20":
			if len(pts) > 0 {
				pts[len(pts)-1][1] = v
			}
		}
	}
	return pts
}

// Enabled reports whether clipping is enabled (71).
func (s *SpatialFilter) Enabled() bool { return s.Int("71", 1) != 0 }

// SetEnabled enables or disables clipping.
func (s *SpatialFilter) SetEnabled(on bool) {
	v := 0
	if on {
		v = 1
	}
	s.SetInt("71", v)
}

// IDBuffer is an IDBUFFER, LAYER_FILTER or FIELDLIST: a list of handles
// (330) or, for LAYER_FILTER, layer names (8).
type IDBuffer struct{ *Raw }

// Handles returns the handles (330).
func (b *IDBuffer) Handles() []string { return b.values("330") }

// LayerNames returns the layer names of a LAYER_FILTER (8).
func (b *IDBuffer) LayerNames() []string { return b.values("8") }

func (r *Raw) values(code string) []string {
	var vs []string
	for _, p := range r.body() {
		if p[0] == code {
			vs = append(vs, strings.TrimSpace(p[1]))
		}
	}
	return vs
}

// Field is a FIELD object (field code and evaluated value of a text).
type Field struct{ *Raw }

// Code returns the field code (1 and 2 chunks... as stored in code 2).
func (f *Field) Code() string { s, _ := f.Value("2"); return s }

// EvaluatorID returns the evaluator ID (1).
func (f *Field) EvaluatorID() string { s, _ := f.Value("1"); return s }

// GeoData is a GEODATA object (geographic location of the drawing).
type GeoData struct{ *Raw }

// DesignPoint returns the design point in WCS (10, 20, 30).
func (g *GeoData) DesignPoint() []float64 {
	return []float64{g.Float("10", 0), g.Float("20", 0), g.Float("30", 0)}
}

// ReferencePoint returns the reference point in geographic coordinates
// (11, 21, 31: longitude, latitude, elevation).
func (g *GeoData) ReferencePoint() []float64 {
	return []float64{g.Float("11", 0), g.Float("21", 0), g.Float("31", 0)}
}

// CoordinateSystem returns the coordinate system definition (303 chunks
// and 301).
func (g *GeoData) CoordinateSystem() string {
	var sb strings.Builder
	for _, p := range g.body() {
		if p[0] == "303" || p[0] == "301" {
			sb.WriteString(p[1])
		}
	}
	return sb.String()
}

// Sun is a SUN object (sun light of a viewport or the model space).
type Sun struct{ *Raw }

// On reports whether the sun is on (290).
func (s *Sun) On() bool { return s.Int("290", 0) != 0 }

// Intensity returns the intensity (40).
func (s *Sun) Intensity() float64 { return s.Float("40", 1) }

// Julian returns the date (91, julian day) and the time (92, seconds past midnight).
func (s *Sun) Julian() (int, int) { return s.Int("91", 0), s.Int("92", 0) }

// Named is a MATERIAL, VISUALSTYLE or PLOTSETTINGS object: its name and
// its group codes (Value, Float, Int, Set).
type Named struct{ *Raw }

// Name returns the name (1 for MATERIAL, 2 for VISUALSTYLE, 1 for PLOTSETTINGS).
func (n *Named) Name() string {
	code := "1"
	if strings.EqualFold(n.Type, "VISUALSTYLE") {
		code = "2"
	}
	s, _ := n.Value(code)
	return s
}

// ImageVariables is a RASTERVARIABLES (ACAD_IMAGE_VARS) or
// WIPEOUTVARIABLES (ACAD_WIPEOUT_VARS) object.
type ImageVariables struct{ *Raw }

// NewRasterVariables creates RASTERVARIABLES: frame (0 no frame, 1 show
// frame), quality (0 draft, 1 high) and units (0 none, 1 mm, 2 cm, 3 m,
// 4 km, 5 inch, 6 foot, 7 yard, 8 mile).
func NewRasterVariables(owner handle.Handler, frame, quality, units int) *ImageVariables {
	return &ImageVariables{newObject("RASTERVARIABLES", owner, [2]string{"100", "AcDbRasterVariables"},
		[2]string{"90", "0"}, [2]string{"70", strconv.Itoa(frame)}, [2]string{"71", strconv.Itoa(quality)}, [2]string{"72", strconv.Itoa(units)})}
}

// NewWipeoutVariables creates WIPEOUTVARIABLES: frame (0 not displayed,
// 1 displayed and plotted, 2 displayed, not plotted).
func NewWipeoutVariables(owner handle.Handler, frame int) *ImageVariables {
	return &ImageVariables{newObject("WIPEOUTVARIABLES", owner, [2]string{"100", "AcDbWipeoutVariables"},
		[2]string{"70", strconv.Itoa(frame)})}
}

// Frame returns the frame setting (70).
func (v *ImageVariables) Frame() int { return v.Int("70", 0) }

// SortEntry is an entry of a SORTENTSTABLE: an entity and its sort handle.
type SortEntry struct {
	Entity     handle.Handler // entity (nil for entries read)
	Handle     string         // 331 as read
	SortHandle string         // 5 as read
}

// SortEntsTable is a SORTENTSTABLE: the draw order of the entities of a
// block (ACAD_SORTENTS in the extension dictionary of its BLOCK_RECORD).
// Entities drawn later are on top.
type SortEntsTable struct {
	Block   handle.Handler // 330: BLOCK_RECORD
	Entries []SortEntry
	// Order is the draw order set by SetOrder (entities later in the list
	// are drawn later); the sort handles are computed when writing.
	order  []handle.Handler
	owner  handle.Handler
	handle int
	src    int
	raw    *Raw // as read
}

// NewSortEntsTable creates a SORTENTSTABLE of block owned by owner.
func NewSortEntsTable(owner, block handle.Handler) *SortEntsTable {
	return &SortEntsTable{owner: owner, Block: block}
}

// SortEntsTableFrom returns the SORTENTSTABLE read as r.
func SortEntsTableFrom(r *Raw) *SortEntsTable {
	s := &SortEntsTable{src: r.src, raw: r}
	cur := -1
	for _, p := range r.body() {
		switch p[0] {
		case "331":
			s.Entries = append(s.Entries, SortEntry{Handle: strings.TrimSpace(p[1])})
			cur = len(s.Entries) - 1
		case "5":
			if cur >= 0 {
				s.Entries[cur].SortHandle = strings.TrimSpace(p[1])
			}
		}
	}
	return s
}

// SetOrder sets the draw order: entities drawn in the order given.
func (s *SortEntsTable) SetOrder(es ...handle.Handler) {
	s.order = es
	s.Entries = nil
}

// Order returns the order set by SetOrder.
func (s *SortEntsTable) Order() []handle.Handler { return s.order }

// IsObject is for Object interface.
func (s *SortEntsTable) IsObject() bool { return true }

// Handle returns the handle.
func (s *SortEntsTable) Handle() int { return s.handle }

// SetHandle sets the handle (the handle read is kept).
func (s *SortEntsTable) SetHandle(v *int) { s.handle = handle.Next(s.src, v) }

// SetOwner sets the owner (the extension dictionary).
func (s *SortEntsTable) SetOwner(h handle.Handler) { s.owner = h }

// SourceHandle returns the handle in the file it was read from ("" if new).
func (s *SortEntsTable) SourceHandle() string { return handle.Format(s.src) }

// Format writes data to formatter: entries read as read; an order set by
// SetOrder gets the sorted handles of its entities as sort handles.
func (s *SortEntsTable) Format(f format.Formatter) {
	if s.raw != nil && s.order == nil {
		s.raw.handle = s.handle
		s.raw.Format(f)
		return
	}
	f.WriteString(0, "SORTENTSTABLE")
	f.WriteHex(5, s.handle)
	f.WriteString(102, "{ACAD_REACTORS")
	writeRef(f, 330, s.owner)
	f.WriteString(102, "}")
	writeRef(f, 330, s.owner)
	f.WriteString(100, "AcDbSortentsTable")
	writeRef(f, 330, s.Block)
	hs := make([]int, len(s.order))
	for i, e := range s.order {
		hs[i] = e.Handle()
	}
	sorted := append([]int{}, hs...)
	sort.Ints(sorted)
	for i, h := range hs {
		f.WriteHex(331, h)
		f.WriteHex(5, sorted[i])
	}
}

func writeRef(f format.Formatter, code int, h handle.Handler) {
	if h == nil {
		f.WriteString(code, "0")
		return
	}
	f.WriteHex(code, h.Handle())
}
