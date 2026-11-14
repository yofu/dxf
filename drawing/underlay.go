package drawing

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/handle"
	"github.com/yofu/dxf/object"
)

// underlay classes (CLASSES section): definition and reference of each format
var underlayClasses = map[string][2][]string{
	entity.UNDERLAY_PDF: {{"PDFDEFINITION", "AcDbPdfDefinition", "1153", "0"}, {"PDFUNDERLAY", "AcDbPdfReference", "4095", "1"}},
	entity.UNDERLAY_DWF: {{"DWFDEFINITION", "AcDbDwfDefinition", "1153", "0"}, {"DWFUNDERLAY", "AcDbDwfReference", "1153", "1"}},
	entity.UNDERLAY_DGN: {{"DGNDEFINITION", "AcDbDgnDefinition", "1153", "0"}, {"DGNUNDERLAY", "AcDbDgnReference", "1153", "1"}},
}

// AddUnderlayDefinition adds the definition of an underlay file to the
// OBJECTS section (dictionary ACAD_PDFDEFINITIONS, ...): fileFormat is
// entity.UNDERLAY_PDF, UNDERLAY_DWF or UNDERLAY_DGN, and name is the page
// number for PDF ("1" if empty), the sheet for DWF or the model for DGN
// ("default" if empty). Use Underlay to place it.
func (d *Drawing) AddUnderlayDefinition(fileFormat, filename, name string) (*entity.UnderlayDefinition, error) {
	fileFormat = strings.ToUpper(fileFormat)
	if _, ok := underlayClasses[fileFormat]; !ok {
		return nil, fmt.Errorf("unknown underlay format: %s", fileFormat)
	}
	if name == "" {
		name = "1"
		if fileFormat == entity.UNDERLAY_DGN {
			name = "default"
		}
	}
	def := entity.NewUnderlayDefinition(fileFormat, filename, name)
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)) + " - " + name
	key := base
	dict := d.underlayDictionary(fileFormat)
	for i := 1; dict.Item(key) != nil; i++ {
		key = fmt.Sprintf("%s (%d)", base, i)
	}
	if err := d.AddUnderlayDefinitionWithKey(def, key); err != nil {
		return nil, err
	}
	return def, nil
}

// AddUnderlayDefinitionWithKey adds an underlay definition with its key in
// the dictionary of its format (used when reading).
func (d *Drawing) AddUnderlayDefinitionWithKey(def *entity.UnderlayDefinition, key string) error {
	if _, ok := underlayClasses[def.FileFormat]; !ok {
		return fmt.Errorf("unknown underlay format: %s", def.FileFormat)
	}
	dict := d.underlayDictionary(def.FileFormat)
	if err := dict.AddItem(key, def); err != nil {
		return err
	}
	def.SetOwner(dict)
	d.addObject(def)
	d.underlayDefs = append(d.underlayDefs, def)
	return nil
}

// UnderlayDefinitions returns the underlay definitions.
func (d *Drawing) UnderlayDefinitions() []*entity.UnderlayDefinition {
	return d.underlayDefs
}

// underlayDictionary returns the dictionary of the definitions of
// fileFormat (e.g. ACAD_PDFDEFINITIONS), adding it if it does not exist.
func (d *Drawing) underlayDictionary(fileFormat string) *object.Dictionary {
	if dict, ok := d.underlayDicts[fileFormat]; ok {
		return dict
	}
	if d.underlayDicts == nil {
		d.underlayDicts = map[string]*object.Dictionary{}
	}
	dict := object.NewDictionary()
	dict.SetOwner(d.dictionary)
	d.addObject(dict)
	d.dictionary.AddItem("ACAD_"+fileFormat+"DEFINITIONS", dict)
	d.underlayDicts[fileFormat] = dict
	return dict
}

// Underlay places the underlay of definition def at (x, y, z) with scale.
func (d *Drawing) Underlay(def *entity.UnderlayDefinition, x, y, z, scale float64) (*entity.Underlay, error) {
	if def == nil {
		return nil, fmt.Errorf("underlay without definition")
	}
	u := entity.NewUnderlay(def)
	u.Insert = []float64{x, y, z}
	u.Scale = []float64{scale, scale, scale}
	d.applyCurrent(u)
	d.AddEntity(u)
	return u, nil
}

// linkUnderlays sets the underlays referring to each definition (reactors)
// and adds the classes of the underlay formats used.
func (d *Drawing) linkUnderlays() {
	if len(d.underlayDefs) == 0 {
		return
	}
	refs := map[*entity.UnderlayDefinition][]handle.Handler{}
	add := func(es entity.Entities) {
		for _, e := range es {
			if u, ok := e.(*entity.Underlay); ok && u.Definition != nil {
				refs[u.Definition] = append(refs[u.Definition], u)
			}
		}
	}
	add(d.Entities())
	for _, b := range d.Blocks() {
		add(b.Entities)
	}
	for _, def := range d.underlayDefs {
		def.SetReactors(refs[def]...)
		for _, c := range underlayClasses[def.FileFormat] {
			d.ensureClass([]string{c[0], c[1], "ObjectDBX Classes", c[2], c[3]})
		}
	}
}
