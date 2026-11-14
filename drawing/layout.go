package drawing

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/yofu/dxf/object"
)

// Layouts returns the layouts: "Model" (model space), the active paper
// space layout (block *Paper_Space) and the other layouts.
func (d *Drawing) Layouts() []*object.Layout {
	return d.layouts
}

// Layout returns the layout named name (case-insensitive).
func (d *Drawing) Layout(name string) (*object.Layout, error) {
	for _, l := range d.layouts {
		if strings.EqualFold(l.Name, name) {
			return l, nil
		}
	}
	return nil, fmt.Errorf("layout %s doesn't exist", name)
}

// AddLayout adds a paper space layout. Its entities are in the block
// "*Paper_Spacen"; draw them after UseLayout(name).
func (d *Drawing) AddLayout(name string) (*object.Layout, error) {
	if name == "" {
		return nil, errors.New("layout name is blank")
	}
	if l, err := d.Layout(name); err == nil {
		return l, fmt.Errorf("layout %s already exists", name)
	}
	var block string
	for n := 0; ; n++ {
		block = "*Paper_Space" + strconv.Itoa(n)
		if _, err := d.Block(block); err != nil {
			break
		}
	}
	tab := 0
	for _, l := range d.layouts {
		if l.TabOrder > tab {
			tab = l.TabOrder
		}
	}
	return d.DefineLayout(name, block, tab+1)
}

// DefineLayout sets the name and tab order of the layout of block
// (e.g. "*Paper_Space" or "*Paper_Space0"), or creates the layout and its
// block if it does not exist. It is used when reading LAYOUT objects;
// use AddLayout to add a layout.
func (d *Drawing) DefineLayout(name, block string, tabOrder int) (*object.Layout, error) {
	if name == "" || block == "" {
		return nil, errors.New("layout name and block are required")
	}
	var l *object.Layout
	for _, ol := range d.layouts {
		if strings.EqualFold(ol.Block, block) {
			l = ol
			break
		}
	}
	if other, err := d.Layout(name); err == nil && other != l {
		return nil, fmt.Errorf("layout %s already exists", name)
	}
	if l == nil {
		if _, err := d.Block(block); err != nil {
			if _, err := d.AddBlock(block, 0, 0, 0); err != nil {
				return nil, err
			}
		}
		l = object.NewLayout(name, tabOrder)
		l.Block = block
		l.SetOwner(d.layoutdict)
		d.layoutdict.AddItem(name, l)
		d.addObject(l)
		d.layouts = append(d.layouts, l)
		return l, nil
	}
	if l.Name != name {
		d.layoutdict.RemoveItem(l.Name)
		l.Name = name
		d.layoutdict.AddItem(name, l)
	}
	l.TabOrder = tabOrder
	return l, nil
}

// RenameLayout renames a layout.
func (d *Drawing) RenameLayout(name, newName string) error {
	l, err := d.Layout(name)
	if err != nil {
		return err
	}
	_, err = d.DefineLayout(newName, l.Block, l.TabOrder)
	return err
}

// UseLayout makes Drawing methods create entities in the layout named name:
// model space ("Model"), the active paper space layout, or another layout
// (its block). ModelSpace and PaperSpace also end it.
func (d *Drawing) UseLayout(name string) error {
	l, err := d.Layout(name)
	if err != nil {
		return err
	}
	if d.currentBlock != nil && !d.inLayoutBlock {
		return fmt.Errorf("block %s is being defined", d.currentBlock.Name)
	}
	switch {
	case strings.EqualFold(l.Block, "*Model_Space"):
		d.ModelSpace()
	case strings.EqualFold(l.Block, "*Paper_Space"):
		d.PaperSpace()
	default:
		b, err := d.Block(l.Block)
		if err != nil {
			return err
		}
		d.currentBlock = b
		d.inLayoutBlock = true
		d.paperSpace = true
	}
	return nil
}

// layoutOf returns the layout whose block is block.
func (d *Drawing) layoutOf(block string) *object.Layout {
	for _, l := range d.layouts {
		if strings.EqualFold(l.Block, block) {
			return l
		}
	}
	return nil
}
