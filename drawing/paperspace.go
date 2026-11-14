package drawing

import (
	"errors"
	"fmt"

	"github.com/yofu/dxf/entity"
)

// PaperSpace makes Drawing methods (Line, Text, ...) create entities in
// paper space (the layout "Layout1", code 67 = 1) until ModelSpace is called.
func (d *Drawing) PaperSpace() {
	d.leaveLayoutBlock()
	d.paperSpace = true
}

// leaveLayoutBlock ends adding entities to a non-active layout.
func (d *Drawing) leaveLayoutBlock() {
	if d.inLayoutBlock {
		d.currentBlock = nil
		d.inLayoutBlock = false
	}
}

// ModelSpace makes Drawing methods create entities in model space (default).
func (d *Drawing) ModelSpace() {
	d.leaveLayoutBlock()
	d.paperSpace = false
}

// IsPaperSpace reports whether new entities are created in paper space.
func (d *Drawing) IsPaperSpace() bool {
	return d.paperSpace
}

// PaperSpaceEntities returns the entities in paper space.
func (d *Drawing) PaperSpaceEntities() entity.Entities {
	es := entity.New()
	for _, e := range d.Entities() {
		if sp, ok := e.(entity.SpaceSelector); ok && sp.InPaperSpace() {
			es = es.Add(e)
		}
	}
	return es
}

// ModelSpaceEntities returns the entities in model space.
func (d *Drawing) ModelSpaceEntities() entity.Entities {
	es := entity.New()
	for _, e := range d.Entities() {
		if sp, ok := e.(entity.SpaceSelector); !ok || !sp.InPaperSpace() {
			es = es.Add(e)
		}
	}
	return es
}

// allViewports returns the VIEWPORT entities of all layouts.
func (d *Drawing) allViewports() []*entity.Viewport {
	var vs []*entity.Viewport
	add := func(es entity.Entities) {
		for _, e := range es {
			if v, ok := e.(*entity.Viewport); ok {
				vs = append(vs, v)
			}
		}
	}
	add(d.Entities())
	for _, b := range d.Blocks() {
		add(b.Entities)
	}
	return vs
}

// Viewport creates a VIEWPORT showing model space in the paper space
// layout in use (the active one after PaperSpace, or another one after
// UseLayout). center, width and height are in paper space; viewCenter
// (model space) is shown at the center, and viewHeight model space units
// fit the height (the scale is height / viewHeight).
// The paper space viewport of the layout, covering its limits, is created
// first if the layout has no viewport.
func (d *Drawing) Viewport(center []float64, width, height float64, viewCenter []float64, viewHeight float64) (*entity.Viewport, error) {
	if d.currentBlock != nil && !d.inLayoutBlock {
		return nil, errors.New("viewport cannot be created in a block")
	}
	if width <= 0 || height <= 0 || viewHeight <= 0 {
		return nil, errors.New("viewport size must be positive")
	}
	block := "*Paper_Space"
	var existing entity.Entities // entities of the layout
	if d.inLayoutBlock {
		block = d.currentBlock.Name
		existing = d.currentBlock.Entities
	} else {
		existing = d.PaperSpaceEntities()
	}
	n := 0
	for _, e := range existing {
		if _, ok := e.(*entity.Viewport); ok {
			n++
		}
	}
	maxID := 0
	for _, v := range d.allViewports() {
		if v.ID > maxID {
			maxID = v.ID
		}
	}
	add := func(v *entity.Viewport) {
		v.SetPaperSpace(true)
		if d.inLayoutBlock {
			d.currentBlock.AddEntity(v)
		} else {
			d.Sections[ENTITIES] = d.Entities().Add(v)
		}
	}
	if n == 0 {
		l := d.layoutOf(block)
		if l == nil {
			return nil, fmt.Errorf("no layout for block %s", block)
		}
		w := l.LimMax[0] - l.LimMin[0]
		h := l.LimMax[1] - l.LimMin[1]
		main := entity.NewViewport()
		main.Center = []float64{l.LimMin[0] + w/2, l.LimMin[1] + h/2, 0}
		main.Width, main.Height = w, h
		main.ViewCenter = []float64{main.Center[0], main.Center[1]}
		main.ViewHeight = h
		maxID++
		main.ID, main.Status = maxID, 1
		main.SetLayer(d.Layers["0"])
		add(main)
		n = 1
	}
	v := entity.NewViewport()
	v.Center = point3(center)
	v.Width, v.Height = width, height
	v.ViewCenter = []float64{viewCenter[0], viewCenter[1]}
	v.ViewHeight = viewHeight
	v.ID = maxID + 1
	v.Status = n + 1
	paper := d.paperSpace
	d.paperSpace = true
	d.applyCurrent(v)
	d.paperSpace = paper
	add(v)
	return v, nil
}
