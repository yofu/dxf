package drawing

import (
	"errors"
	"fmt"
	"strings"

	"github.com/yofu/dxf/block"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/table"
)

// Query returns the entities of the ENTITIES section (model space and the
// active paper space layout) for which match returns true.
func (d *Drawing) Query(match func(entity.Entity) bool) entity.Entities {
	es := entity.New()
	for _, e := range d.Entities() {
		if match(e) {
			es = es.Add(e)
		}
	}
	return es
}

// EntitiesOnLayer returns the entities of the ENTITIES section on the
// layer named name (case-insensitive).
func (d *Drawing) EntitiesOnLayer(name string) entity.Entities {
	return d.Query(func(e entity.Entity) bool {
		return e.Layer() != nil && strings.EqualFold(e.Layer().Name(), name)
	})
}

// RemoveEntity removes entities from the drawing (the ENTITIES section and
// blocks) and from groups; leaders lose them as annotations.
// It returns the number of entities removed.
func (d *Drawing) RemoveEntity(es ...entity.Entity) int {
	targets := map[entity.Entity]bool{}
	for _, e := range es {
		targets[e] = true
	}
	removed := map[entity.Entity]bool{}
	filter := func(list entity.Entities) entity.Entities {
		kept := entity.New()
		for _, e := range list {
			if targets[e] {
				removed[e] = true
			} else {
				kept = kept.Add(e)
			}
		}
		return kept
	}
	d.Sections[ENTITIES] = filter(d.Entities())
	for _, b := range d.Blocks() {
		b.Entities = filter(b.Entities)
	}
	if len(removed) == 0 {
		return 0
	}
	list := make([]entity.Entity, 0, len(removed))
	for _, e := range es {
		if removed[e] {
			list = append(list, e)
		}
	}
	for _, g := range d.Groups {
		g.RemoveEntity(list...)
	}
	d.forEachEntity(func(e entity.Entity) {
		if l, ok := e.(*entity.Leader); ok && l.Annotation != nil && removed[l.Annotation] {
			l.Annotation = nil
		}
	})
	return len(removed)
}

// forEachEntity calls f for the entities of the ENTITIES section and blocks.
func (d *Drawing) forEachEntity(f func(entity.Entity)) {
	for _, e := range d.Entities() {
		f(e)
	}
	for _, b := range d.Blocks() {
		for _, e := range b.Entities {
			f(e)
		}
	}
}

// TransformEntities applies t to entities (see entity.Transform). The
// graphics of dimensions are transformed with them. Under non-uniform
// scaling, circles and arcs are replaced by ellipses (in the drawing,
// blocks and groups; see entity.TransformEntity). It stops at the first
// entity that cannot be transformed (entity.ErrTransform: an extrusion
// direction other than the Z axis, an unsupported entity type, or
// entity.ErrNonUniform).
func (d *Drawing) TransformEntities(t entity.Transform, es ...entity.Entity) error {
	_, err := d.TransformEntitiesReplacing(t, es...)
	return err
}

// TransformEntitiesReplacing is TransformEntities returning the
// transformed entities: es, except for circles and arcs replaced by
// ellipses.
func (d *Drawing) TransformEntitiesReplacing(t entity.Transform, es ...entity.Entity) (entity.Entities, error) {
	out := make(entity.Entities, 0, len(es))
	for _, e := range es {
		n, err := entity.TransformEntity(e, t)
		if err != nil {
			return out, fmt.Errorf("%T: %w", e, err)
		}
		if n != e {
			d.ReplaceEntity(e, n)
		}
		out = append(out, n)
		if dim, ok := e.(*entity.Dimension); ok && dim.Flags&entity.DIM_BLOCK_OWNED != 0 {
			if b, err := d.Block(dim.BlockName); err == nil {
				for _, be := range b.Entities {
					if bt, ok := be.(entity.Transformable); ok {
						if err := bt.ApplyTransform(t); err != nil {
							return out, fmt.Errorf("dimension block %s: %w", b.Name, err)
						}
					}
				}
			}
		}
	}
	return out, nil
}

// ReplaceEntity replaces old by n where old is: the ENTITIES section,
// blocks, groups and leader annotations. It reports whether old was found.
func (d *Drawing) ReplaceEntity(old, n entity.Entity) bool {
	found := false
	replace := func(list entity.Entities) {
		for i, e := range list {
			if e == old {
				list[i] = n
				found = true
			}
		}
	}
	replace(d.Entities())
	for _, b := range d.Blocks() {
		replace(b.Entities)
	}
	for _, g := range d.Groups {
		g.ReplaceEntity(old, n)
	}
	d.forEachEntity(func(e entity.Entity) {
		if l, ok := e.(*entity.Leader); ok && l.Annotation == old {
			l.Annotation = n
		}
	})
	return found
}

// Mirror mirrors entities about the line through p1 and p2 (in the XY
// plane). Text stays readable (see entity.Transform.MirrorText).
func (d *Drawing) Mirror(p1, p2 []float64, es ...entity.Entity) error {
	return d.TransformEntities(entity.MirrorBy(p1, p2), es...)
}

// ScaleXYZ scales entities by sx, sy and sz (non-zero; negative values
// mirror) around base. Circles and arcs become ellipses when sx and sy
// differ (see TransformEntities).
func (d *Drawing) ScaleXYZ(base []float64, sx, sy, sz float64, es ...entity.Entity) error {
	if sx == 0 || sy == 0 || sz == 0 {
		return errors.New("scale factors must not be 0")
	}
	return d.TransformEntities(entity.ScaleXYZ(base, sx, sy, sz), es...)
}

// Move moves entities by (dx, dy, dz).
func (d *Drawing) Move(dx, dy, dz float64, es ...entity.Entity) error {
	return d.TransformEntities(entity.Move(dx, dy, dz), es...)
}

// Rotate rotates entities by angle (degrees, counterclockwise) about the
// Z axis through base.
func (d *Drawing) Rotate(base []float64, angle float64, es ...entity.Entity) error {
	return d.TransformEntities(entity.Rotate(base, angle), es...)
}

// Scale scales entities uniformly by factor (> 0) around base.
func (d *Drawing) Scale(base []float64, factor float64, es ...entity.Entity) error {
	if factor <= 0 {
		return errors.New("scale factor must be positive")
	}
	return d.TransformEntities(entity.ScaleBy(base, factor), es...)
}

// RenameLayer renames a layer. Entities refer to the layer itself, so
// they follow the new name.
func (d *Drawing) RenameLayer(name, newName string) error {
	l, err := d.Layer(name, false)
	if err != nil {
		return err
	}
	if newName == "" {
		return errors.New("layer name is blank")
	}
	if strings.EqualFold(name, "0") {
		return errors.New("layer 0 cannot be renamed")
	}
	if other, err := d.Layer(newName, false); err == nil && other != l {
		return fmt.Errorf("layer %s already exists", newName)
	}
	delete(d.Layers, l.Name())
	l.SetName(newName)
	d.Layers[newName] = l
	return nil
}

// layerUsed reports whether the layer is used by an entity or a block.
func (d *Drawing) layerUsed(l *table.Layer) bool {
	used := false
	d.forEachEntity(func(e entity.Entity) {
		if e.Layer() == l {
			used = true
		}
		switch c := e.(type) {
		case *entity.Polyline:
			for _, v := range c.Vertices {
				used = used || v.Layer() == l
			}
		case *entity.Insert:
			for _, a := range c.Attributes {
				used = used || a.Layer() == l
			}
		}
	})
	for _, b := range d.Blocks() {
		used = used || b.Layer() == l
	}
	return used
}

// RemoveLayer removes an unused layer. Layer 0, the current layer and
// layers used by entities or blocks cannot be removed.
func (d *Drawing) RemoveLayer(name string) error {
	l, err := d.Layer(name, false)
	if err != nil {
		return err
	}
	switch {
	case strings.EqualFold(l.Name(), "0"):
		return errors.New("layer 0 cannot be removed")
	case l == d.CurrentLayer:
		return fmt.Errorf("layer %s is the current layer", l.Name())
	case d.layerUsed(l):
		return fmt.Errorf("layer %s is used", l.Name())
	}
	d.Sections[TABLES].(table.Tables)[table.LAYER].Remove(l.Name())
	delete(d.Layers, l.Name())
	return nil
}

// blockUsed reports whether the block is referenced by an INSERT or a DIMENSION.
func (d *Drawing) blockUsed(name string) bool {
	used := false
	d.forEachEntity(func(e entity.Entity) {
		switch c := e.(type) {
		case *entity.Insert:
			used = used || strings.EqualFold(c.BlockName, name)
		case *entity.Dimension:
			used = used || strings.EqualFold(c.BlockName, name)
		}
	})
	return used
}

func layoutBlock(name string) bool {
	n := strings.ToUpper(name)
	return n == "*MODEL_SPACE" || strings.HasPrefix(n, "*PAPER_SPACE")
}

// RenameBlock renames a block; block references and its BLOCK_RECORD
// follow the new name.
func (d *Drawing) RenameBlock(name, newName string) error {
	b, err := d.Block(name)
	if err != nil {
		return err
	}
	if newName == "" || strings.HasPrefix(newName, "*") || layoutBlock(b.Name) {
		return fmt.Errorf("block %s cannot be renamed to %q", b.Name, newName)
	}
	if other, err := d.Block(newName); err == nil && other != b {
		return fmt.Errorf("block %s already exists", newName)
	}
	old := b.Name
	d.forEachEntity(func(e entity.Entity) {
		if ins, ok := e.(*entity.Insert); ok && strings.EqualFold(ins.BlockName, old) {
			ins.BlockName = newName
		}
	})
	if r, err := d.Sections[TABLES].(table.Tables)[table.BLOCK_RECORD].Contains(old); err == nil {
		r.(*table.BlockRecord).SetName(newName)
	}
	b.Name = newName
	return nil
}

// RemoveBlock removes a block definition that is not referenced.
// Layout blocks cannot be removed.
func (d *Drawing) RemoveBlock(name string) error {
	b, err := d.Block(name)
	if err != nil {
		return err
	}
	switch {
	case layoutBlock(b.Name):
		return fmt.Errorf("layout block %s cannot be removed", b.Name)
	case d.blockUsed(b.Name):
		return fmt.Errorf("block %s is referenced", b.Name)
	case d.currentBlock == b:
		return fmt.Errorf("block %s is being defined", b.Name)
	}
	kept := make(block.Blocks, 0, len(d.Blocks()))
	for _, ob := range d.Blocks() {
		if ob != b {
			kept = append(kept, ob)
		}
	}
	d.Sections[BLOCKS] = kept
	d.Sections[TABLES].(table.Tables)[table.BLOCK_RECORD].Remove(b.Name)
	return nil
}
