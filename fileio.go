package dxf

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/handle"
	"github.com/yofu/dxf/table"
)

// Problem is a problem found by Audit.
type Problem struct {
	Entity  entity.Entity // nil for problems of the drawing
	Message string
	Fixed   bool
}

func (p Problem) String() string {
	s := p.Message
	if p.Entity != nil {
		s = TypeName(p.Entity) + ": " + s
	}
	if p.Fixed {
		s += " (fixed)"
	}
	return s
}

// Audit checks a drawing (as ezdxf's audit): block references to undefined
// blocks, dimensions without graphics, coordinates that are not numbers,
// circles and arcs without radius, text without height, invalid color
// numbers and lineweights, images and underlays without definition, and
// duplicate handles of objects read. With fix, entities that cannot be
// repaired are removed and the others repaired.
func Audit(d *drawing.Drawing, fix bool) []Problem {
	var problems []Problem
	var remove []entity.Entity
	report := func(e entity.Entity, msg string, fixed bool) {
		problems = append(problems, Problem{e, msg, fixed})
	}
	badPoint := func(ps ...[]float64) bool {
		for _, p := range ps {
			for _, v := range p {
				if math.IsNaN(v) || math.IsInf(v, 0) {
					return true
				}
			}
		}
		return false
	}
	seen := map[int]entity.Entity{}
	d.ForEachEntity(func(e entity.Entity) {
		if s, ok := e.(interface{ SourceHandle() string }); ok && s.SourceHandle() != "" {
			h := handle.Parse(s.SourceHandle())
			if _, dup := seen[h]; dup {
				report(e, "duplicate handle "+s.SourceHandle(), false)
			}
			seen[h] = e
		}
		if c, ok := e.(interface{ ColorIndex() int }); ok {
			if v := c.ColorIndex(); v < 0 || v > 257 {
				if fix {
					e.(interface{ SetColor(int) }).SetColor(entity.ColorByLayer)
				}
				report(e, fmt.Sprintf("invalid color number %d", v), fix)
			}
		}
		if l, ok := e.(interface{ LineWeight() int }); ok {
			v := l.LineWeight()
			if v < -3 || v > 211 {
				if fix {
					e.(interface{ SetLineWeight(int) int }).SetLineWeight(table.LineWeightByLayer)
				}
				report(e, fmt.Sprintf("invalid lineweight %d", v), fix)
			}
		}
		lo, hi := e.BBox()
		if badPoint(lo, hi) {
			report(e, "coordinates are not numbers", fix)
			remove = append(remove, e)
			return
		}
		switch x := e.(type) {
		case *entity.Insert:
			if _, err := d.Block(x.BlockName); err != nil {
				report(e, "block "+x.BlockName+" is not defined", fix)
				remove = append(remove, e)
			}
		case *entity.Dimension:
			if _, err := d.Block(x.BlockName); x.BlockName == "" || err != nil {
				fixed := fix && d.RegenerateDimension(x) == nil
				report(e, "dimension without graphics block", fixed)
			}
		case *entity.Arc:
			if x.Radius <= 0 {
				report(e, "radius is not positive", fix)
				remove = append(remove, e)
			}
		case *entity.Circle:
			if x.Radius <= 0 {
				report(e, "radius is not positive", fix)
				remove = append(remove, e)
			}
		case *entity.Text:
			if x.Height <= 0 {
				if fix {
					x.Height = 1
				}
				report(e, "text height is not positive", fix)
			}
		case *entity.MText:
			if x.Height <= 0 {
				if fix {
					x.Height = 1
				}
				report(e, "text height is not positive", fix)
			}
		case *entity.Image:
			if !x.IsWipeout() && x.Definition == nil && x.DefinitionHandle() == "" {
				report(e, "image without definition", fix)
				remove = append(remove, e)
			}
		case *entity.Underlay:
			if x.Definition == nil && x.DefinitionHandle() == "" {
				report(e, "underlay without definition", fix)
				remove = append(remove, e)
			}
		}
	})
	if fix && len(remove) > 0 {
		d.RemoveEntity(remove...)
	}
	return problems
}

// copyResources makes the layers, linetypes, text styles, dimension
// styles and blocks used by e in src exist in dst. It returns the name of
// the block of e in dst (anonymous blocks get new names).
func copyResources(dst, src *drawing.Drawing, e entity.Entity, depth int) (string, error) {
	if depth > 32 {
		return "", errors.New("import: blocks nested too deeply")
	}
	if l := e.Layer(); l != nil {
		if _, err := dst.Layer(l.Name(), false); err != nil {
			nl, _ := dst.AddLayer(l.Name(), l.Color, dst.CurrentLayer.LineType, false)
			if nl != nil {
				if lt := l.LineType; lt != nil {
					nl.LineType = importLineType(dst, lt)
				}
				nl.SetLineWidth(l.LineWidth())
				if r, g, b, ok := l.TrueColor(); ok {
					nl.SetTrueColor(r, g, b)
				}
			}
		}
	}
	if lp, ok := e.(interface{ LineType() *table.LineType }); ok && lp.LineType() != nil {
		importLineType(dst, lp.LineType())
	}
	var style *table.Style
	switch x := e.(type) {
	case *entity.Text:
		style = x.Style
	case *entity.MText:
		style = x.Style
	case *entity.Attribute:
		style = x.Style
	case *entity.AttributeDefinition:
		style = x.Style
	}
	if style != nil {
		importStyle(dst, style)
	}
	block := ""
	switch x := e.(type) {
	case *entity.Insert:
		block = x.BlockName
	case *entity.AcadTable:
		block = x.BlockName
	case *entity.Dimension:
		block = x.BlockName
		if x.Style != "" {
			if ds, err := src.DimStyle(x.Style); err == nil {
				if _, err := dst.DimStyle(x.Style); err != nil {
					c := *ds
					dst.Sections[drawing.TABLES].(table.Tables)[table.DIMSTYLE].Add(&c)
				}
			}
		}
	}
	if block == "" {
		return "", nil
	}
	b, err := src.Block(block)
	if err != nil {
		return "", nil // kept as it is
	}
	anonymous := strings.HasPrefix(block, "*")
	if !anonymous {
		if _, err := dst.Block(block); err == nil {
			return block, nil
		}
	}
	var nb interface {
		AddEntity(...entity.Entity)
	}
	name := block
	if anonymous {
		prefix := block[:2]
		ab, err := dst.AddAnonymousBlock(prefix)
		if err != nil {
			return "", err
		}
		ab.Coord = append([]float64{}, b.Coord...)
		ab.Flag = b.Flag
		nb, name = ab, ab.Name
	} else {
		ab, err := dst.AddBlock(b.Name, b.Coord[0], b.Coord[1], b.Coord[2])
		if err != nil {
			return "", err
		}
		ab.Flag, ab.Description, ab.XrefPath = b.Flag, b.Description, b.XrefPath
		nb = ab
	}
	for _, be := range b.Entities {
		c, err := importEntity(dst, src, be, depth+1)
		if err != nil {
			return "", err
		}
		nb.AddEntity(c)
	}
	return name, nil
}

func importLineType(dst *drawing.Drawing, lt *table.LineType) *table.LineType {
	if x, err := dst.LineType(lt.Name()); err == nil {
		return x
	}
	var es []*table.LineTypeElement
	for _, el := range lt.Elements() {
		c := *el
		if c.Style != nil {
			c.Style = importStyle(dst, c.Style)
		}
		es = append(es, &c)
	}
	n, _ := dst.AddComplexLineType(lt.Name(), lt.Description, es...)
	return n
}

func importStyle(dst *drawing.Drawing, st *table.Style) *table.Style {
	if st.Flags&1 != 0 { // shape file
		return dst.AddShapeFile(st.FontName)
	}
	if s, ok := findStyle(dst, st.Name()); ok {
		return s
	}
	n, _ := dst.AddStyle(st.Name(), st.FontName, st.BigFontName, false)
	if n != nil {
		n.WidthFactor, n.ObliqueAngle, n.FixedTextHeight, n.Flags, n.GenFlag = st.WidthFactor, st.ObliqueAngle, st.FixedTextHeight, st.Flags, st.GenFlag
	}
	return n
}

// importEntity copies e of src into dst (not added).
func importEntity(dst, src *drawing.Drawing, e entity.Entity, depth int) (entity.Entity, error) {
	block, err := copyResources(dst, src, e, depth)
	if err != nil {
		return nil, err
	}
	c, err := Copy(dst, e)
	if err != nil {
		return nil, err
	}
	if block != "" {
		switch x := c.(type) {
		case *entity.Insert:
			x.BlockName = block
		case *entity.AcadTable:
			x.BlockName = block
		case *entity.Dimension:
			x.BlockName = block
		}
	}
	switch x := c.(type) {
	case *entity.Image:
		if def := e.(*entity.Image).Definition; def != nil {
			nd, err := dst.AddImageDefinition(def.FileName, int(def.Size[0]), int(def.Size[1]))
			if err == nil {
				nd.PixelSize = append([]float64{}, def.PixelSize...)
				x.Definition = nd
			}
		}
	case *entity.Underlay:
		if def := e.(*entity.Underlay).Definition; def != nil {
			if nd, err := dst.AddUnderlayDefinition(def.FileFormat, def.FileName, def.Name); err == nil {
				x.Definition = nd
			}
		}
	case *entity.MLine:
		if s := e.(*entity.MLine).Style; s != nil {
			ns, err := dst.MLineStyle(s.Name())
			if err != nil {
				ns, _ = dst.AddMLineStyle(s.Name(), s.Elements...)
				ns.Flags, ns.FillColor, ns.StartAngle, ns.EndAngle = s.Flags, s.FillColor, s.StartAngle, s.EndAngle
			}
			x.Style = ns
		}
	case *entity.MultiLeader:
		// references of src: dst gets its standard style
		x.Style, x.TextStyle, x.LeaderLineType, x.Arrow, x.Block = entity.Ref{}, entity.Ref{}, entity.Ref{}, entity.Ref{}, entity.Ref{}
		if x.Context != nil && x.Context.MText != nil {
			x.Context.MText.Style = entity.Ref{}
		}
		x.ArrowHeads, x.BlockAttribs = nil, nil
	case *entity.Hatch:
		for _, b := range x.Boundaries {
			b.Sources = nil // the boundary entities are other objects
		}
		x.Associative = false
	}
	if sh, ok := c.(interface{ SetSourceHandle(string) }); ok {
		sh.SetSourceHandle("")
	}
	return c, nil
}

// ImportEntities copies entities of drawing src into dst (as ezdxf's
// importer and xref modules) with the resources they use: layers,
// linetypes, text styles, dimension styles, blocks (recursively),
// multiline styles and image and underlay definitions; resources with the
// same name in dst are used. The copies are added to dst and returned.
func ImportEntities(dst, src *drawing.Drawing, es ...entity.Entity) (entity.Entities, error) {
	var out entity.Entities
	for _, e := range es {
		c, err := importEntity(dst, src, e, 0)
		if err != nil {
			return out, err
		}
		dst.AddEntity(c)
		out = append(out, c)
	}
	return out, nil
}

// Block flags of external references (BLOCK 70).
const (
	BlockXref        = 4
	BlockXrefOverlay = 8
	BlockXrefLoaded  = 32 // resolved
)

// AttachXref attaches the drawing file as external reference block name
// and inserts it at insert (as AutoCAD's XATTACH); overlay attaches it as
// overlay (not visible in drawings referring to this one).
func AttachXref(d *drawing.Drawing, name, file string, insert []float64, overlay bool) (*entity.Insert, error) {
	b, err := d.AddBlock(name, 0, 0, 0)
	if err != nil {
		return nil, err
	}
	b.Flag = BlockXref | BlockXrefLoaded
	if overlay {
		b.Flag |= BlockXrefOverlay
	}
	b.XrefPath = file
	if len(insert) < 3 {
		insert = append(append([]float64{}, insert...), 0, 0, 0)[:3]
	}
	return d.Insert(name, insert[0], insert[1], insert[2])
}

// BindXref binds the external reference block name: the model space
// entities of the referenced drawing (read by load, e.g. FromFile with the
// path of the block) become the entities of the block, with their
// resources, and the block is no longer an external reference.
func BindXref(d *drawing.Drawing, name string, load func(path string) (*drawing.Drawing, error)) error {
	b, err := d.Block(name)
	if err != nil {
		return err
	}
	if b.Flag&BlockXref == 0 {
		return fmt.Errorf("block %s is not an external reference", name)
	}
	src, err := load(b.XrefPath)
	if err != nil {
		return err
	}
	var es entity.Entities
	for _, e := range src.ModelSpaceEntities() {
		c, err := importEntity(d, src, e, 0)
		if err != nil {
			return err
		}
		es = append(es, c)
	}
	base := src.Header().InsBase
	if len(base) >= 3 {
		b.Coord = append([]float64{}, base...)
	}
	b.Entities = es
	b.Flag &^= BlockXref | BlockXrefOverlay | BlockXrefLoaded
	b.XrefPath = ""
	return nil
}
