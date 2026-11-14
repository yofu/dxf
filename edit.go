package dxf

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/geometry"
	"github.com/yofu/dxf/table"
)

// Extruder represents an entity with code 210, 220, 230 like Circle.
//
//	210: Extrusion direction (optional; default = 0, 0, 1)
//	X value
//	220, 230: Y and Z values of extrusion direction (optional)
type Extruder interface { // 210 220 230
	CurrentDirection() []float64
	SetDirection([]float64)
	CurrentCoord() []float64
	SetCoord([]float64)
}

// SetExtrusion sets new coord according to given direction.
func SetExtrusion(e Extruder, d []float64) {
	dx, dy, err := geometry.ArbitraryAxis(d)
	if err != nil {
		return
	}
	b := e.CurrentDirection()
	e.SetDirection(d)
	n := e.CurrentCoord()
	bx, by, _ := geometry.ArbitraryAxis(b)
	before := [][]float64{bx, by, b}
	after := [][]float64{dx, dy, d}
	newcoord := make([]float64, 3)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				newcoord[i] += n[j] * before[j][k] * after[i][k]
			}
		}
	}
	e.SetCoord(newcoord)
}

// Copy returns a copy of entity e of drawing d. The copy is not added to
// the drawing (use d.AddEntity). Layers, linetypes and styles are shared;
// a dimension gets a copy of its graphics block; the annotation of a
// leader and group membership are not copied.
func Copy(d *drawing.Drawing, e entity.Entity) (entity.Entity, error) {
	// written and read as binary DXF: all values are copied exactly
	f := format.NewBinary()
	entity.FormatEntity(f, e)
	if err := f.Err(); err != nil {
		return nil, err
	}
	data := f.Output()[len(format.BinarySentinel):]
	src := &binarySource{r: bufio.NewReader(strings.NewReader(data)), twoByte: true}
	var pairs [][2]string
	for {
		c, v, err := src.next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, [2]string{c, v})
	}
	es, err := parseEntityList(d, pairs)
	if err != nil {
		return nil, err
	}
	if len(es) != 1 {
		return nil, fmt.Errorf("copy of %T: %d entities", e, len(es))
	}
	c := es[0]
	if sh, ok := c.(interface{ SetSourceHandle(string) }); ok {
		sh.SetSourceHandle("")
	}
	if u, ok := c.(*entity.Underlay); ok {
		// the definition is shared
		src := e.(*entity.Underlay)
		u.Definition = src.Definition
		u.SetDefinitionHandle(src.DefinitionHandle())
	}
	if dim, ok := c.(*entity.Dimension); ok && dim.BlockName != "" {
		if err := copyDimensionBlock(d, dim); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// copyDimensionBlock gives dim a copy of its graphics block.
func copyDimensionBlock(d *drawing.Drawing, dim *entity.Dimension) error {
	b, err := d.Block(dim.BlockName)
	if err != nil {
		return nil // no graphics
	}
	nb, err := d.AddAnonymousBlock("*D")
	if err != nil {
		return err
	}
	for _, e := range b.Entities {
		c, err := Copy(d, e)
		if err != nil {
			return err
		}
		nb.AddEntity(c)
	}
	dim.BlockName = nb.Name
	return nil
}

// CopyEntities copies entities of drawing d (see Copy), transforms the
// copies by t, and adds them where the original entities are (the ENTITIES
// section or a block). It returns the copies.
func CopyEntities(d *drawing.Drawing, t entity.Transform, es ...entity.Entity) (entity.Entities, error) {
	copies := entity.New()
	for _, e := range es {
		c, err := Copy(d, e)
		if err != nil {
			return copies, err
		}
		if err := d.TransformEntities(t, c); err != nil {
			return copies, err
		}
		if !insertAfter(d, e, c) {
			d.AddEntity(c)
		}
		copies = copies.Add(c)
	}
	return copies, nil
}

// insertAfter inserts n after e in the list containing e.
func insertAfter(d *drawing.Drawing, e entity.Entity, n ...entity.Entity) bool {
	insert := func(list entity.Entities) (entity.Entities, bool) {
		for i, x := range list {
			if x == e {
				out := append(entity.Entities{}, list[:i+1]...)
				out = append(out, n...)
				return append(out, list[i+1:]...), true
			}
		}
		return list, false
	}
	if list, ok := insert(d.Entities()); ok {
		d.Sections[drawing.ENTITIES] = list
		return true
	}
	for _, b := range d.Blocks() {
		if list, ok := insert(b.Entities); ok {
			b.Entities = list
			return true
		}
	}
	return false
}

// Explode replaces a block reference by copies of the entities of its
// block, transformed by the insertion point, scale, rotation and array of
// the reference. Entities on layer 0 take the layer of the reference, and
// BYBLOCK color, linetype and lineweight are resolved. Visible attributes
// become TEXT entities; attribute definitions are dropped. Nested block
// references are exploded by one level only. The scale must be uniform.
// It returns the new entities.
func Explode(d *drawing.Drawing, ins *entity.Insert) (entity.Entities, error) {
	b, err := d.Block(ins.BlockName)
	if err != nil {
		return nil, err
	}
	s := ins.Scale[0]
	if s <= 0 || math.Abs(ins.Scale[1]-s) > 1e-12 || math.Abs(ins.Scale[2]-s) > 1e-12 {
		return nil, errors.New("explode: the scale must be uniform and positive")
	}
	if len(ins.Direction) >= 3 && (ins.Direction[0] != 0 || ins.Direction[1] != 0 || ins.Direction[2] <= 0) {
		return nil, entity.ErrTransform
	}
	rad := ins.Rotation * math.Pi / 180
	sin, cos := math.Sin(rad), math.Cos(rad)
	cols, rows := ins.Columns, ins.Rows
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	out := entity.New()
	for c := 0; c < cols; c++ {
		for r := 0; r < rows; r++ {
			ax, ay := float64(c)*ins.ColumnSpacing, float64(r)*ins.RowSpacing
			t := entity.Transform{
				Base:  b.Coord,
				Angle: ins.Rotation,
				Scale: s,
				Offset: []float64{
					ins.Coord[0] - b.Coord[0] + ax*cos - ay*sin,
					ins.Coord[1] - b.Coord[1] + ax*sin + ay*cos,
					ins.Coord[2] - b.Coord[2],
				},
			}
			for _, be := range b.Entities {
				if _, ok := be.(*entity.AttributeDefinition); ok {
					continue
				}
				cp, err := Copy(d, be)
				if err != nil {
					return nil, err
				}
				inheritFromBlockReference(d, cp, ins)
				if err := d.TransformEntities(t, cp); err != nil {
					return nil, err
				}
				out = out.Add(cp)
			}
		}
	}
	for _, a := range ins.Attributes {
		if a.Flags&entity.ATTRIBUTE_INVISIBLE != 0 {
			continue
		}
		out = out.Add(attributeText(a))
	}
	if !insertAfter(d, ins, out...) {
		return nil, errors.New("explode: the block reference is not in the drawing")
	}
	d.RemoveEntity(ins)
	return out, nil
}

// attributeText returns a TEXT entity with the value and the text
// properties of an attribute.
func attributeText(a *entity.Attribute) *entity.Text {
	tx := entity.NewText()
	cp := func(v []float64) []float64 { return append([]float64(nil), v...) }
	tx.Coord1, tx.Coord2 = cp(a.Coord1), cp(a.Coord2)
	tx.Height, tx.Rotation = a.Height, a.Rotation
	tx.WidthFactor, tx.ObliqueAngle = a.WidthFactor, a.ObliqueAngle
	tx.Value, tx.Style = a.Value, a.Style
	tx.GenFlag, tx.HorizontalFlag, tx.VerticalFlag = a.GenFlag, a.HorizontalFlag, a.VerticalFlag
	tx.Thickness = a.Thickness
	if len(a.Extrusion) > 0 {
		tx.Extrusion = cp(a.Extrusion)
	}
	tx.SetLayer(a.Layer())
	entity.SetColorIndex(tx, a.ColorIndex())
	tx.SetLineType(a.LineType())
	tx.SetLineWeight(a.LineWeight())
	tx.SetLtscale(a.Ltscale())
	tx.SetPaperSpace(a.InPaperSpace())
	tx.SetVisible(a.IsVisible())
	return tx
}

// inheritFromBlockReference resolves layer 0 and BYBLOCK properties of an
// exploded entity from the block reference.
func inheritFromBlockReference(d *drawing.Drawing, e entity.Entity, ins *entity.Insert) {
	if e.Layer() == nil || e.Layer().Name() == "0" {
		e.SetLayer(ins.Layer())
	}
	if c, ok := e.(entity.Colored); ok && c.ColorIndex() == entity.ColorByBlock {
		entity.SetColorIndex(e, ins.ColorIndex())
		if tc, ok := e.(entity.TrueColorer); ok {
			if r, g, b, ok := ins.TrueColor(); ok {
				tc.SetTrueColor(r, g, b)
				entity.SetColorIndex(e, ins.ColorIndex())
			}
		}
	}
	if tc, ok := e.(entity.TrueColorer); ok {
		if _, isAlpha := tc.Transparency(); !isAlpha && isTransparencyByBlock(tc) {
			if t, ok := ins.Transparency(); ok {
				tc.SetTransparency(t)
			} else {
				tc.ClearTransparency()
			}
		}
	}
	if lp, ok := e.(entity.LineProperties); ok {
		if lp.LineType() != nil && strings.EqualFold(lp.LineType().Name(), "BYBLOCK") {
			lp.SetLineType(ins.LineType())
		}
		if lp.LineWeight() == table.LineWeightByBlock {
			lp.SetLineWeight(ins.LineWeight())
		}
	}
	if sp, ok := e.(entity.SpaceSelector); ok {
		sp.SetPaperSpace(ins.InPaperSpace())
	}
}

// isTransparencyByBlock reports whether the transparency is BYBLOCK.
func isTransparencyByBlock(e entity.TrueColorer) bool {
	type raw interface{ TransparencyValue() int }
	if r, ok := e.(raw); ok {
		return r.TransparencyValue() == entity.TransparencyByBlock
	}
	return false
}
