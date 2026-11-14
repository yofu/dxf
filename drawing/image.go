package drawing

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/yofu/dxf/class"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/handle"
	"github.com/yofu/dxf/object"
)

// image classes (CLASSES section): name, C++ class, application, proxy flags, entity
var (
	imageClasses = [][]string{
		{"IMAGE", "AcDbRasterImage", "ISM", "2175", "1"},
		{"IMAGEDEF", "AcDbRasterImageDef", "ISM", "0", "0"},
		{"IMAGEDEF_REACTOR", "AcDbRasterImageDefReactor", "ISM", "1", "0"},
	}
	wipeoutClass = []string{"WIPEOUT", "AcDbWipeout", "WipeOut", "127", "1"}
)

// ensureClass adds the class c (name, C++ class, application, proxy
// flags, entity) to the CLASSES section if it does not exist.
func (d *Drawing) ensureClass(c []string) {
	cs := d.Sections[CLASSES].(class.Classes)
	for _, x := range cs {
		if strings.EqualFold(x.Name(), c[0]) {
			return
		}
	}
	d.Sections[CLASSES] = append(cs, &class.Class{Data: [][2]string{
		{"1", c[0]}, {"2", c[1]}, {"3", c[2]},
		{"90", c[3]}, {"280", "0"}, {"281", c[4]},
	}})
}

// AddImageDefinition adds the definition of an image file of width x
// height pixels to the OBJECTS section (dictionary ACAD_IMAGE_DICT). Use
// Image to place it.
func (d *Drawing) AddImageDefinition(filename string, width, height int) (*entity.ImageDefinition, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("image size must be positive: %d x %d", width, height)
	}
	def := entity.NewImageDefinition(filename, width, height)
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	key := base
	dict := d.imageDictionary()
	for i := 1; dict.Item(key) != nil; i++ {
		key = fmt.Sprintf("%s (%d)", base, i)
	}
	if err := d.AddImageDefinitionWithKey(def, key); err != nil {
		return nil, err
	}
	return def, nil
}

// AddImageDefinitionWithKey adds an image definition with its key in the
// dictionary ACAD_IMAGE_DICT (used when reading).
func (d *Drawing) AddImageDefinitionWithKey(def *entity.ImageDefinition, key string) error {
	dict := d.imageDictionary()
	if err := dict.AddItem(key, def); err != nil {
		return err
	}
	def.SetOwner(dict)
	d.addObject(def)
	d.imageDefs = append(d.imageDefs, def)
	return nil
}

// ImageDefinitions returns the image definitions.
func (d *Drawing) ImageDefinitions() []*entity.ImageDefinition {
	return d.imageDefs
}

// imageDictionary returns the dictionary ACAD_IMAGE_DICT, adding it if it
// does not exist.
func (d *Drawing) imageDictionary() *object.Dictionary {
	if d.imageDict != nil {
		return d.imageDict
	}
	d.imageDict = object.NewDictionary()
	d.imageDict.SetOwner(d.dictionary)
	d.addObject(d.imageDict)
	d.dictionary.AddItem("ACAD_IMAGE_DICT", d.imageDict)
	return d.imageDict
}

// Image places the image of definition def with its lower left corner at
// (x, y, z), width and height in drawing units and rotation in degrees.
func (d *Drawing) Image(def *entity.ImageDefinition, x, y, z, width, height, rotation float64) (*entity.Image, error) {
	if def == nil {
		return nil, fmt.Errorf("image without definition")
	}
	im := entity.NewImage(def)
	im.Insert = []float64{x, y, z}
	im.SetSize(width, height, rotation)
	d.applyCurrent(im)
	d.AddEntity(im)
	return im, nil
}

// Wipeout creates a WIPEOUT masking the polygon of vertices (in the XY
// plane).
func (d *Drawing) Wipeout(vertices ...[]float64) (*entity.Image, error) {
	if len(vertices) < 3 {
		return nil, fmt.Errorf("wipeout needs 3 or more vertices")
	}
	w := entity.NewWipeout(vertices...)
	d.applyCurrent(w)
	d.AddEntity(w)
	return w, nil
}

// linkImages gives each image of a definition its IMAGEDEF_REACTOR (in the
// OBJECTS section), sets the reactors of the definitions, and adds the
// classes of images and wipeouts used.
func (d *Drawing) linkImages() {
	reactors := map[*entity.ImageDefinition][]handle.Handler{}
	var all []object.Object
	wipeout := false
	check := func(es entity.Entities) {
		for _, e := range es {
			im, ok := e.(*entity.Image)
			if !ok {
				continue
			}
			if im.IsWipeout() {
				wipeout = true
				continue
			}
			if im.Definition == nil {
				continue
			}
			if im.Reactor() == nil {
				im.SetReactor(&entity.ImageDefReactor{})
			}
			r := im.Reactor()
			im.SetReactor(r) // owner
			reactors[im.Definition] = append(reactors[im.Definition], r)
			all = append(all, r)
		}
	}
	check(d.Entities())
	for _, b := range d.Blocks() {
		check(b.Entities)
	}
	// the reactors of the images in the drawing replace those of the OBJECTS section
	objs := object.New()
	for _, o := range d.Sections[OBJECTS].(object.Objects) {
		if _, ok := o.(*entity.ImageDefReactor); !ok {
			objs = objs.Add(o)
		}
	}
	for _, r := range all {
		objs = objs.Add(r)
	}
	d.Sections[OBJECTS] = objs
	for _, def := range d.imageDefs {
		def.SetReactors(reactors[def]...)
	}
	if len(d.imageDefs) > 0 {
		for _, c := range imageClasses {
			d.ensureClass(c)
		}
	}
	if wipeout {
		d.ensureClass(wipeoutClass)
	}
}
