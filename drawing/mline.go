package drawing

import (
	"fmt"
	"strings"

	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/object"
)

// namedDictionary returns the dictionary key of the root dictionary (e.g.
// "ACAD_MLINESTYLE"), adding it if it does not exist.
func (d *Drawing) namedDictionary(key string) *object.Dictionary {
	if dict, ok := d.namedDicts[key]; ok {
		return dict
	}
	if d.namedDicts == nil {
		d.namedDicts = map[string]*object.Dictionary{}
	}
	dict := object.NewDictionary()
	dict.SetOwner(d.dictionary)
	d.addObject(dict)
	d.dictionary.AddItem(key, dict)
	d.namedDicts[key] = dict
	return dict
}

// HasNamedDictionary reports whether the root dictionary has the
// dictionary key created by this package.
func (d *Drawing) HasNamedDictionary(key string) bool {
	_, ok := d.namedDicts[key]
	return ok
}

// AddMLineStyle adds an MLINESTYLE with elements (dictionary
// ACAD_MLINESTYLE).
func (d *Drawing) AddMLineStyle(name string, elements ...entity.MLineStyleElement) (*entity.MLineStyle, error) {
	if s, err := d.MLineStyle(name); err == nil {
		return s, fmt.Errorf("mline style %s already exists", name)
	}
	s := entity.NewMLineStyle(name, elements...)
	return s, d.AddMLineStyleObject(s)
}

// AddMLineStyleObject adds an MLINESTYLE (used when reading).
func (d *Drawing) AddMLineStyleObject(s *entity.MLineStyle) error {
	dict := d.namedDictionary("ACAD_MLINESTYLE")
	if err := dict.AddItem(s.Name(), s); err != nil {
		return err
	}
	s.SetOwner(dict)
	d.addObject(s)
	d.mlineStyles = append(d.mlineStyles, s)
	return nil
}

// MLineStyle returns the MLINESTYLE name (case-insensitive).
func (d *Drawing) MLineStyle(name string) (*entity.MLineStyle, error) {
	for _, s := range d.mlineStyles {
		if strings.EqualFold(s.Name(), name) {
			return s, nil
		}
	}
	return nil, fmt.Errorf("mline style %s doesn't exist", name)
}

// MLineStyles returns the MLINESTYLE objects.
func (d *Drawing) MLineStyles() []*entity.MLineStyle {
	return d.mlineStyles
}

// standardMLineStyle returns the style "Standard", adding it (two elements
// at +-0.5) if it does not exist.
func (d *Drawing) standardMLineStyle() *entity.MLineStyle {
	if s, err := d.MLineStyle("Standard"); err == nil {
		return s
	}
	s, _ := d.AddMLineStyle("Standard",
		entity.MLineStyleElement{Offset: 0.5, Color: 256, LineType: "BYLAYER"},
		entity.MLineStyleElement{Offset: -0.5, Color: 256, LineType: "BYLAYER"})
	return s
}

// MLine creates an MLINE through points with the MLINESTYLE style
// ("" for "Standard", added if needed), scale and justification
// (entity.MLINE_TOP, ...).
func (d *Drawing) MLine(style string, points [][]float64, closed bool, scale float64, justification int) (*entity.MLine, error) {
	var s *entity.MLineStyle
	if style == "" || strings.EqualFold(style, "Standard") {
		s = d.standardMLineStyle()
	} else {
		var err error
		if s, err = d.MLineStyle(style); err != nil {
			return nil, err
		}
	}
	m := entity.NewMLine(s)
	if scale != 0 {
		m.Scale = scale
	}
	m.Justification = justification
	if closed {
		m.Flags |= entity.MLINE_CLOSED
	}
	if err := m.Generate(points); err != nil {
		return nil, err
	}
	d.add(m)
	return m, nil
}

// linkMLines gives MLINE entities without style the style of their name
// (or "Standard").
func (d *Drawing) linkMLines() {
	d.forEachEntity(func(e entity.Entity) {
		m, ok := e.(*entity.MLine)
		if !ok || m.Style != nil {
			return
		}
		name, h := m.StyleReference()
		if h != "" && h != "0" {
			return // kept as read
		}
		if s, err := d.MLineStyle(name); err == nil {
			m.Style = s
		} else {
			m.Style = d.standardMLineStyle()
		}
	})
}
