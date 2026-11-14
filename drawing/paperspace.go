package drawing

import (
	"errors"

	"github.com/yofu/dxf/entity"
)

// PaperSpace makes Drawing methods (Line, Text, ...) create entities in
// paper space (the layout "Layout1", code 67 = 1) until ModelSpace is called.
func (d *Drawing) PaperSpace() {
	d.paperSpace = true
}

// ModelSpace makes Drawing methods create entities in model space (default).
func (d *Drawing) ModelSpace() {
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

// viewports returns the VIEWPORT entities in paper space.
func (d *Drawing) viewports() []*entity.Viewport {
	var vs []*entity.Viewport
	for _, e := range d.Entities() {
		if v, ok := e.(*entity.Viewport); ok && v.InPaperSpace() {
			vs = append(vs, v)
		}
	}
	return vs
}

// Viewport creates a VIEWPORT in paper space showing model space.
// center, width and height are in paper space; viewCenter (model space)
// is shown at the center, and viewHeight model space units fit the height
// (the scale is height / viewHeight).
// The paper space viewport (ID 1) covering the paper is created first if
// it does not exist.
func (d *Drawing) Viewport(center []float64, width, height float64, viewCenter []float64, viewHeight float64) (*entity.Viewport, error) {
	if d.currentBlock != nil {
		return nil, errors.New("viewport cannot be created in a block")
	}
	if width <= 0 || height <= 0 || viewHeight <= 0 {
		return nil, errors.New("viewport size must be positive")
	}
	vs := d.viewports()
	maxID, hasMain := 0, false
	for _, v := range vs {
		if v.ID > maxID {
			maxID = v.ID
		}
		hasMain = hasMain || v.ID == 1
	}
	if !hasMain {
		l := d.layouts[1]
		w := l.LimMax[0] - l.LimMin[0]
		h := l.LimMax[1] - l.LimMin[1]
		main := entity.NewViewport()
		main.Center = []float64{l.LimMin[0] + w/2, l.LimMin[1] + h/2, 0}
		main.Width, main.Height = w, h
		main.ViewCenter = []float64{main.Center[0], main.Center[1]}
		main.ViewHeight = h
		main.ID, main.Status = 1, 1
		main.SetLayer(d.Layers["0"])
		d.Sections[ENTITIES] = d.Entities().Add(main)
		if maxID < 1 {
			maxID = 1
		}
		vs = append(vs, main)
	}
	v := entity.NewViewport()
	v.Center = point3(center)
	v.Width, v.Height = width, height
	v.ViewCenter = []float64{viewCenter[0], viewCenter[1]}
	v.ViewHeight = viewHeight
	v.ID = maxID + 1
	v.Status = len(vs) + 1
	paper := d.paperSpace
	d.paperSpace = true
	d.applyCurrent(v)
	d.paperSpace = paper
	d.Sections[ENTITIES] = d.Entities().Add(v)
	return v, nil
}
