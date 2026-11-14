// Package drawing defines Drawing struct for DXF.
package drawing

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/yofu/dxf/block"
	"github.com/yofu/dxf/class"
	"github.com/yofu/dxf/codepage"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
	"github.com/yofu/dxf/header"
	"github.com/yofu/dxf/object"
	"github.com/yofu/dxf/table"
)

// Drawing contains DXF drawing data.
type Drawing struct {
	FileName     string
	Layers       map[string]*table.Layer
	Groups       map[string]*object.Group
	Styles       map[string]*table.Style
	CurrentLayer *table.Layer
	CurrentStyle *table.Style
	formatter    format.Formatter
	Sections     []Section
	dictionary   *object.Dictionary
	groupdict    *object.Dictionary
	plotstyles   *object.AcDbDictionaryWDFLT
	layouts      []*object.Layout // "Model" (*Model_Space), "Layout1" (*Paper_Space), ...
	layoutdict   *object.Dictionary
	PlotStyle    handle.Handler

	underlayDicts map[string]*object.Dictionary // ACAD_PDFDEFINITIONS, ... by format
	underlayDefs  []*entity.UnderlayDefinition
	imageDict     *object.Dictionary // ACAD_IMAGE_DICT
	imageDefs     []*entity.ImageDefinition
	// savebuff is used internally for the io.Reader options.
	savebuff *bytes.Buffer

	currentBlock  *block.Block // block being defined between BeginBlock and EndBlock
	paperSpace    bool         // new entities are created in paper space
	handleBase    int          // first handle for new objects (above the handles of the file read)
	inLayoutBlock bool         // currentBlock is the block of a (non-active) layout
}

// New creates a new Drawing.
func New() *Drawing {
	d := new(Drawing)
	d.formatter = format.NewASCII()
	d.formatter.SetPrecision(16)
	tables := table.New()
	d.Sections = []Section{
		header.New(),
		class.New(),
		tables,
		block.New(),
		entity.New(),
		object.New(),
	}
	// layer "0" and style "Standard" belong to this drawing's tables
	// (not the package-level templates table.LY_0 / table.ST_STANDARD).
	ly0, _ := tables[table.LAYER].Contains("0")
	st, _ := tables[table.STYLE].Contains(table.ST_STANDARD.Name())
	d.Layers = make(map[string]*table.Layer)
	d.Layers["0"] = ly0.(*table.Layer)
	d.Groups = make(map[string]*object.Group)
	d.CurrentLayer = d.Layers["0"]
	d.Styles = make(map[string]*table.Style)
	d.Styles["STANDARD"] = st.(*table.Style)
	d.CurrentStyle = d.Styles["STANDARD"]
	d.dictionary = object.NewDictionary()
	d.addObject(d.dictionary)
	wd, ph := object.NewAcDbDictionaryWDFLT(d.dictionary)
	d.plotstyles = wd
	d.dictionary.AddItem("ACAD_PLOTSTYLENAME", wd)
	d.addObject(wd)
	d.addObject(ph)
	d.groupdict = object.NewDictionary()
	d.groupdict.SetOwner(d.dictionary)
	d.addObject(d.groupdict)
	d.dictionary.AddItem("ACAD_GROUP", d.groupdict)
	d.layoutdict = object.NewDictionary()
	d.layoutdict.SetOwner(d.dictionary)
	d.addObject(d.layoutdict)
	d.dictionary.AddItem("ACAD_LAYOUT", d.layoutdict)
	for i, name := range []string{"Model", "Layout1"} {
		l := object.NewLayout(name, i)
		l.Block = []string{"*Model_Space", "*Paper_Space"}[i]
		l.SetOwner(d.layoutdict)
		d.layoutdict.AddItem(name, l)
		d.addObject(l)
		d.layouts = append(d.layouts, l)
	}
	d.PlotStyle = ph
	d.Layers["0"].SetPlotStyle(d.PlotStyle)
	return d
}

func (d *Drawing) saveFile(filename string) error {
	w, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = d.WriteTo(w)
	return err
}

// Save saves the drawing file.
// If it is the first time, use SaveAs(filename).
func (d *Drawing) Save() error {
	if d.FileName == "" {
		return errors.New("filename is blank, use SaveAs(filename)")
	}
	return d.saveFile(d.FileName)
}

// SaveAs saves the drawing file as given filename.
func (d *Drawing) SaveAs(filename string) error {
	d.FileName = filename
	return d.saveFile(filename)
}

// linkOwners sets the owner references required by AC1015 before writing:
// BLOCK_RECORD <-> LAYOUT for model/paper space, BLOCK/ENDBLK -> BLOCK_RECORD,
// and entities -> BLOCK_RECORD of the block (or *Model_Space) containing them.
// A BLOCK_RECORD is added for blocks without one.
func (d *Drawing) linkOwners() {
	records := d.Sections[TABLES].(table.Tables)[table.BLOCK_RECORD]
	record := func(name string) *table.BlockRecord {
		r, err := records.Contains(name)
		if err != nil {
			br := table.NewBlockRecord(name)
			records.Add(br)
			return br
		}
		return r.(*table.BlockRecord)
	}
	for _, l := range d.layouts {
		r := record(l.Block)
		r.SetLayout(l)
		l.SetBlockRecord(r)
	}
	type owned interface{ SetOwner(handle.Handler) }
	for _, b := range d.Blocks() {
		r := record(b.Name)
		b.SetRecord(r)
		for _, e := range b.Entities {
			if o, ok := e.(owned); ok {
				o.SetOwner(r)
			}
		}
	}
	model := record("*Model_Space")
	paper := record("*Paper_Space")
	for _, e := range d.Entities() {
		if o, ok := e.(owned); ok {
			if sp, ok := e.(entity.SpaceSelector); ok && sp.InPaperSpace() {
				o.SetOwner(paper)
			} else {
				o.SetOwner(model)
			}
		}
	}
}

// setHandle sets all the handles contained in Drawing.
func (d *Drawing) setHandle() {
	// objects read from a file keep their handles; the others get handles
	// from handleBase, above the handles of the file
	h := d.handleBase
	if h < 1 {
		h = 1
	}
	for _, s := range d.Sections[1:] {
		s.SetHandle(&h)
	}
	d.Sections[0].SetHandle(&h)
}

func (d *Drawing) Header() *header.Header {
	return d.Sections[0].(*header.Header)
}

// Layer returns the named layer if exists.
// If setcurrent is true, set current layer to it.
func (d *Drawing) Layer(name string, setcurrent bool) (*table.Layer, error) {
	if l, exist := d.Layers[name]; exist {
		if setcurrent {
			d.CurrentLayer = l
		}
		return l, nil
	}
	return nil, fmt.Errorf("layer %s doesn't exist", name)
}

// AddLayer adds a new layer with given name, color and line type.
// If setcurrent is true, set current layer to it.
func (d *Drawing) AddLayer(name string, cl color.ColorNumber, lt *table.LineType, setcurrent bool) (*table.Layer, error) {
	if l, exist := d.Layers[name]; exist {
		if setcurrent {
			d.CurrentLayer = l
		}
		return l, fmt.Errorf("layer %s already exists", name)
	}
	l := table.NewLayer(name, cl, lt)
	l.SetPlotStyle(d.PlotStyle)
	d.Layers[name] = l
	d.Sections[2].(table.Tables).AddLayer(l)
	if setcurrent {
		d.CurrentLayer = l
	}
	return l, nil
}

// ChangeLayer changes current layer to the named layer.
func (d *Drawing) ChangeLayer(name string) error {
	if l, exist := d.Layers[name]; exist {
		d.CurrentLayer = l
		return nil
	}
	return fmt.Errorf("layer %s doesn't exist", name)
}

// Style returns the named text style if exists.
// If setcurrent is true, set current style to it.
func (d *Drawing) Style(name string, setcurrent bool) (*table.Style, error) {
	if s, exist := d.Styles[name]; exist {
		if setcurrent {
			d.CurrentStyle = s
		}
		return s, nil
	}
	return nil, fmt.Errorf("style %s doesn't exist", name)
}

// AddStyle adds a new text style.
// If setcurrent is true, set current style to it.
func (d *Drawing) AddStyle(name string, fontname, bigfontname string, setcurrent bool) (*table.Style, error) {
	if s, exist := d.Styles[name]; exist {
		if setcurrent {
			d.CurrentStyle = s
		}
		return s, fmt.Errorf("style %s already exists", name)
	}
	s := table.NewStyle(name)
	s.FontName = fontname
	s.BigFontName = bigfontname
	d.Styles[name] = s
	d.Sections[TABLES].(table.Tables)[table.STYLE].Add(s)
	if setcurrent {
		d.CurrentStyle = s
	}
	return s, nil
}

// LineType returns the named line type if exists.
func (d *Drawing) LineType(name string) (*table.LineType, error) {
	lt, err := d.Sections[TABLES].(table.Tables)[table.LTYPE].Contains(name)
	if err != nil {
		return nil, fmt.Errorf("linetype %s", err.Error())
	}
	return lt.(*table.LineType), nil
}

// AddLineType adds a new linetype.
func (d *Drawing) AddLineType(name string, desc string, ls ...float64) (*table.LineType, error) {
	lt, _ := d.Sections[TABLES].(table.Tables)[table.LTYPE].Contains(name)
	if lt != nil {
		return lt.(*table.LineType), fmt.Errorf("linetype %s already exists", name)
	}
	newlt := table.NewLineType(name, desc, ls...)
	d.Sections[TABLES].(table.Tables)[table.LTYPE].Add(newlt)
	return newlt, nil
}

// AddComplexLineType adds a new linetype with text strings or shapes
// (see table.TextElement and table.ShapeElement).
func (d *Drawing) AddComplexLineType(name, desc string, es ...*table.LineTypeElement) (*table.LineType, error) {
	lt, err := d.AddLineType(name, desc)
	if err != nil {
		return lt, err
	}
	lt.SetElements(es...)
	return lt, nil
}

// AddShapeFile returns the STYLE entry of shape file fontname (e.g.
// "ltypeshp.shx", flag 1 and no name), adding it if it does not exist.
// It is used by the shapes of complex linetypes.
func (d *Drawing) AddShapeFile(fontname string) *table.Style {
	t := d.Sections[TABLES].(table.Tables)[table.STYLE]
	for _, st := range t.Entries() {
		if s, ok := st.(*table.Style); ok && s.Flags&1 != 0 && strings.EqualFold(s.FontName, fontname) {
			return s
		}
	}
	s := table.NewStyle("")
	s.FontName = fontname
	s.Flags = 1
	s.LastHeightUsed = 0
	t.Add(s)
	return s
}

// Entities returns slice of all entities contained in Drawing.
func (d *Drawing) Entities() entity.Entities {
	return d.Sections[ENTITIES].(entity.Entities)
}

// AddEntity adds a new entity.
// If a block is being defined (between BeginBlock and EndBlock),
// the entity is added to that block instead of the ENTITIES section.
func (d *Drawing) AddEntity(e entity.Entity) {
	if d.currentBlock != nil {
		d.currentBlock.AddEntity(e)
		return
	}
	d.Sections[ENTITIES] = d.Sections[ENTITIES].(entity.Entities).Add(e)
}

// Blocks returns the blocks of the drawing (BLOCKS section).
func (d *Drawing) Blocks() block.Blocks {
	return d.Sections[BLOCKS].(block.Blocks)
}

// Block returns the block named name (case-insensitive).
func (d *Drawing) Block(name string) (*block.Block, error) {
	for _, b := range d.Blocks() {
		if strings.EqualFold(b.Name, name) {
			return b, nil
		}
	}
	return nil, fmt.Errorf("block %s doesn't exist", name)
}

// AddBlock defines a new block with base point (x, y, z).
// Its BLOCK_RECORD is added to the TABLES section as well.
// Entities are added with Block.AddEntity, or by drawing between
// BeginBlock and EndBlock.
func (d *Drawing) AddBlock(name string, x, y, z float64) (*block.Block, error) {
	if name == "" {
		return nil, errors.New("block name is blank")
	}
	if b, err := d.Block(name); err == nil {
		return b, fmt.Errorf("block %s already exists", name)
	}
	b := block.NewBlock(name, "")
	b.Coord = []float64{x, y, z}
	b.SetLayer(d.Layers["0"])
	d.Sections[BLOCKS] = d.Blocks().Add(b)
	records := d.Sections[TABLES].(table.Tables)[table.BLOCK_RECORD]
	if _, err := records.Contains(name); err != nil {
		records.Add(table.NewBlockRecord(name))
	}
	return b, nil
}

// BeginBlock defines a new block with base point (x, y, z) and starts adding
// entities created by Drawing methods (Line, Circle, ...) to it until EndBlock.
func (d *Drawing) BeginBlock(name string, x, y, z float64) (*block.Block, error) {
	if d.currentBlock != nil {
		return nil, fmt.Errorf("block %s is being defined", d.currentBlock.Name)
	}
	b, err := d.AddBlock(name, x, y, z)
	if err != nil {
		return b, err
	}
	d.currentBlock = b
	return b, nil
}

// EndBlock ends the block definition started by BeginBlock.
func (d *Drawing) EndBlock() error {
	if d.inLayoutBlock {
		return errors.New("no block is being defined (a layout is in use; call UseLayout or ModelSpace)")
	}
	if d.currentBlock == nil {
		return errors.New("no block is being defined")
	}
	d.currentBlock = nil
	return nil
}

// Insert creates a new INSERT (block reference) of the block named name at (x, y, z).
// The block must be defined beforehand.
func (d *Drawing) Insert(name string, x, y, z float64) (*entity.Insert, error) {
	b, err := d.Block(name)
	if err != nil {
		return nil, err
	}
	if d.currentBlock == b {
		return nil, fmt.Errorf("block %s cannot insert itself", name)
	}
	i := entity.NewInsert(b.Name)
	i.Coord = []float64{x, y, z}
	d.applyCurrent(i)
	if err := d.SyncAttributes(i); err != nil {
		return nil, err
	}
	d.AddEntity(i)
	return i, nil
}

// AttributeDefinition creates a new ATTDEF with tag, prompt and default value
// at (x, y, z) with given height. Create it between BeginBlock and EndBlock
// (or add it with Block.AddEntity) so that inserts of the block get the
// attribute (see Insert and SyncAttributes).
func (d *Drawing) AttributeDefinition(tag, prompt, defaultValue string, x, y, z, height float64) (*entity.AttributeDefinition, error) {
	if tag == "" || strings.ContainsAny(tag, " ") {
		return nil, fmt.Errorf("invalid attribute tag %q", tag)
	}
	a := entity.NewAttributeDefinition(tag, prompt, defaultValue)
	a.Coord1 = []float64{x, y, z}
	a.Coord2 = []float64{x, y, z}
	a.Height = height
	d.applyCurrent(a)
	a.Style = d.CurrentStyle
	a.WidthFactor = a.Style.WidthFactor
	d.AddEntity(a)
	return a, nil
}

// SyncAttributes creates the attributes (ATTRIB) of ins from the attribute
// definitions (ATTDEF, except constant ones) of its block, and updates the
// position, height and rotation of existing attributes according to the
// insertion point, scale and rotation of ins. Values of existing attributes
// are kept. Insert calls it; call it again after changing ins.
func (d *Drawing) SyncAttributes(ins *entity.Insert) error {
	b, err := d.Block(ins.BlockName)
	if err != nil {
		return err
	}
	for _, e := range b.Entities {
		def, ok := e.(*entity.AttributeDefinition)
		if !ok || def.IsConstant() {
			continue
		}
		a := ins.Attribute(def.Tag)
		if a == nil {
			a = entity.NewAttribute(def.Tag, def.Value)
			ins.AddAttribute(a)
		}
		a.Flags = def.Flags
		a.FieldLength = def.FieldLength
		a.Coord1 = ins.Transform(def.Coord1, b.Coord, 0, 0)
		align := def.Coord2
		if len(align) < 3 {
			align = def.Coord1
		}
		a.Coord2 = ins.Transform(align, b.Coord, 0, 0)
		a.Height = def.Height * math.Abs(ins.Scale[1])
		a.WidthFactor = def.WidthFactor
		if ins.Scale[1] != 0 {
			a.WidthFactor *= math.Abs(ins.Scale[0] / ins.Scale[1])
		}
		a.Rotation = def.Rotation + ins.Rotation
		a.ObliqueAngle = def.ObliqueAngle
		a.Style = def.Style
		a.GenFlag = def.GenFlag
		a.HorizontalFlag = def.HorizontalFlag
		a.VerticalFlag = def.VerticalFlag
		// attributes defined on layer "0" take the layer of the insert
		if def.Layer().Name() == "0" {
			a.SetLayer(ins.Layer())
		} else {
			a.SetLayer(def.Layer())
		}
	}
	return nil
}

// Point creates a new POINT at (x, y, z).
func (d *Drawing) Point(x, y, z float64) (*entity.Point, error) {
	p := entity.NewPoint()
	p.Coord = []float64{x, y, z}
	d.applyCurrent(p)
	d.AddEntity(p)
	return p, nil
}

// Line creates a new LINE from (x1, y1, z1) to (x2, y2, z2).
func (d *Drawing) Line(x1, y1, z1, x2, y2, z2 float64) (*entity.Line, error) {
	l := entity.NewLine()
	l.Start = []float64{x1, y1, z1}
	l.End = []float64{x2, y2, z2}
	d.applyCurrent(l)
	d.AddEntity(l)
	return l, nil
}

// Circle creates a new CIRCLE at (x, y, z) with radius r.
func (d *Drawing) Circle(x, y, z, r float64) (*entity.Circle, error) {
	c := entity.NewCircle()
	c.Center = []float64{x, y, z}
	c.Radius = r
	d.applyCurrent(c)
	d.AddEntity(c)
	return c, nil
}

// Arc creates a new ARC at (x, y, z) with radius r from start to end.
func (d *Drawing) Arc(x, y, z, r, start, end float64) (*entity.Arc, error) {
	c := entity.NewCircle()
	c.Center = []float64{x, y, z}
	c.Radius = r
	d.applyCurrent(c)
	a := entity.NewArc(c)
	a.Angle[0] = start
	a.Angle[1] = end
	d.AddEntity(a)
	return a, nil
}

// Polyline creates a new POLYLINE with given vertices.
func (d *Drawing) Polyline(closed bool, vertices ...[]float64) (*entity.Polyline, error) {
	p := entity.NewPolyline()
	d.applyCurrent(p)
	for _, v := range vertices {
		p.AddVertex(v[0], v[1], v[2])
	}
	if closed {
		p.Close()
	}
	d.AddEntity(p)
	return p, nil
}

// LwPolyline creates a new LWPOLYLINE with given vertices.
func (d *Drawing) LwPolyline(closed bool, vertices ...[]float64) (*entity.LwPolyline, error) {
	size := len(vertices)
	l := entity.NewLwPolyline(size)
	for i := 0; i < size; i++ {
		l.Vertices[i] = vertices[i]
	}
	if closed {
		l.Close()
	}
	d.applyCurrent(l)
	d.AddEntity(l)
	return l, nil
}

// ThreeDFace creates a new 3DFACE with given points.
func (d *Drawing) ThreeDFace(points [][]float64) (*entity.ThreeDFace, error) {
	f := entity.New3DFace()
	if len(points) < 3 {
		return nil, errors.New("3DFace needs 3 or more points")
	}
	for i := 0; i < 3; i++ {
		f.Points[i] = points[i]
	}
	if len(points) >= 4 {
		f.Points[3] = points[3]
	} else {
		f.Points[3] = points[2]
	}
	d.applyCurrent(f)
	d.AddEntity(f)
	return f, nil
}

// Text creates a new TEXT str at (x, y, z) with given height.
func (d *Drawing) Text(str string, x, y, z, height float64) (*entity.Text, error) {
	t := entity.NewText()
	t.Coord1 = []float64{x, y, z}
	t.Coord2 = []float64{x, y, z} // alignment point used when justified (see Text.Anchor)
	t.Height = height
	t.Value = str
	d.applyCurrent(t)
	t.Style = d.CurrentStyle
	t.WidthFactor = t.Style.WidthFactor
	t.ObliqueAngle = t.Style.ObliqueAngle
	t.Style.LastHeightUsed = height
	d.AddEntity(t)
	return t, nil
}

// MText creates a new MTEXT str at (x, y, z) with given height.
// The insertion point is the top left corner (see MText.Attachment).
// Line breaks ("\n", "\r\n") in str are converted to paragraph breaks (\P);
// other MTEXT formatting codes in str are written as they are.
// width is the reference rectangle width for word wrapping (0: no wrapping).
func (d *Drawing) MText(str string, x, y, z, height, width float64) (*entity.MText, error) {
	t := entity.NewMText()
	t.Coord = []float64{x, y, z}
	t.Height = height
	t.Width = width
	t.Value = strings.NewReplacer("\r\n", `\P`, "\n", `\P`).Replace(str)
	d.applyCurrent(t)
	t.Style = d.CurrentStyle
	d.AddEntity(t)
	return t, nil
}

func (d *Drawing) addObject(o object.Object) {
	d.Sections[5] = d.Sections[5].(object.Objects).Add(o)
}

// Group adds given entities to the named group.
// If the named group doesn't exist, create it.
func (d *Drawing) Group(name, desc string, es ...entity.Entity) (*object.Group, error) {
	if g, exist := d.Groups[name]; exist {
		g.AddEntity(es...)
		return g, fmt.Errorf("group %s already exists", name)
	}
	g := object.NewGroup(name, desc, es...)
	d.Groups[name] = g
	g.SetOwner(d.groupdict)
	d.addObject(g)
	return g, nil
}

// AddToGroup adds given entities to the named group.
// If the named group doesn't exist, returns error.
func (d *Drawing) AddToGroup(name string, es ...entity.Entity) error {
	if g, exist := d.Groups[name]; exist {
		g.AddEntity(es...)
		return nil
	}
	return fmt.Errorf("group %s doesn't exist", name)
}

// WriteTo write the dxf file data to the given writer until
// there is no more data to write or if an error occurs. The return
// value n is the number of bytes writer.
// This method full fills the io.WriterTo interface.
func (d *Drawing) WriteTo(w io.Writer) (n int64, err error) {
	if d == nil {
		return 0, nil
	}
	d.syncHeader()
	d.ensureDefaults()
	d.ensureClasses()
	d.ensureAppIDs()
	d.linkOwners()
	d.linkUnderlays()
	d.linkImages()
	d.ensureEntityClasses()
	d.setHandle()
	h := d.Header()
	unicode := codepage.IsUnicodeVersion(h.Version)
	// group codes of later versions are written only for those versions
	if vf, ok := d.formatter.(interface{ SetVersion(string) }); ok {
		vf.SetVersion(strings.ToUpper(strings.TrimSpace(h.Version)))
	}
	// formatters converting each string value (e.g. format.Binary)
	enc, perValue := d.formatter.(interface{ SetEncoder(func(string) string) })
	if perValue {
		if unicode {
			enc.SetEncoder(nil)
		} else {
			cp := h.CodePage
			enc.SetEncoder(func(s string) string { return codepage.Encode(s, cp) })
		}
	}
	d.formatter.Reset()
	for _, s := range d.Sections {
		s.Format(d.formatter)
	}
	d.formatter.WriteString(0, "EOF")
	if unicode || perValue {
		return d.formatter.WriteTo(w) // AC1021 or later: UTF-8
	}
	// Text is written in the code page of $DWGCODEPAGE, and characters
	// outside it as \U+XXXX escapes.
	n2, err := io.WriteString(w, codepage.Encode(d.formatter.Output(), h.CodePage))
	return int64(n2), err
}

// SetFormatter sets the formatter used for writing the drawing:
// format.NewASCII() (default) or format.NewBinary() for binary DXF.
func (d *Drawing) SetFormatter(f format.Formatter) {
	d.formatter = f
	d.savebuff = nil
}

// Formatter returns the formatter used for writing the drawing.
func (d *Drawing) Formatter() format.Formatter {
	return d.formatter
}

// SaveAsBinary saves the drawing as a binary DXF file.
// The formatter of the drawing is not changed.
func (d *Drawing) SaveAsBinary(filename string) error {
	f := d.formatter
	d.formatter = format.NewBinary()
	defer func() { d.formatter = f }()
	return d.SaveAs(filename)
}

var _ io.WriterTo = &Drawing{}

// Read implements the standard Read interface: it reads data from a buffer
// containing the drawing, the buffer in initialized on the first read or
// a read call after a close.  If the drawing is nil, read returns 0 for n,
// nil for the error.
func (d *Drawing) Read(p []byte) (n int, err error) {
	if d == nil {
		return 0, nil
	}
	if d.savebuff == nil {
		// We need to initilize our buffer, and write the contents of the
		// drawing into it.
		d.savebuff = new(bytes.Buffer)
		_, err := d.WriteTo(d.savebuff)
		if err != nil {
			return 0, err
		}
	}
	return d.savebuff.Read(p)
}

// Close implements the standard Close interface: it close the buffer used by
// the read command if one was initilized, and frees the memory held by it. Close
// will always return nil for the error.
func (d *Drawing) Close() error {
	if d != nil && d.savebuff != nil {
		d.savebuff = nil
	}
	return nil
}

var _ io.ReadCloser = &Drawing{}

// SetExt sets the extents of the drawing based on the entities
// in the drawing. If the drawing is nil, this function will panic.
func (d *Drawing) SetExt() {
	mins := []float64{1e16, 1e16, 1e16}
	maxs := []float64{-1e16, -1e16, -1e16}
	for _, en := range d.Entities() {
		if sp, ok := en.(entity.SpaceSelector); ok && sp.InPaperSpace() {
			continue // $EXTMIN/$EXTMAX are model space extents
		}
		tmpmins, tmpmaxs := d.bbox(en, 0)
		for i := 0; i < 3; i++ {
			if tmpmins[i] < mins[i] {
				mins[i] = tmpmins[i]
			}
			if tmpmaxs[i] > maxs[i] {
				maxs[i] = tmpmaxs[i]
			}
		}
	}
	h := d.Header()
	for i := 0; i < 3; i++ {
		h.ExtMin[i] = mins[i]
		h.ExtMax[i] = maxs[i]
	}
}

// maxInsertDepth limits the nesting of block references followed by bbox.
const maxInsertDepth = 16

// bbox returns the bounding box of en.
// For INSERT, the contents of the referenced block are transformed
// (scale, rotation, array; the extrusion direction is ignored).
func (d *Drawing) bbox(en entity.Entity, depth int) ([]float64, []float64) {
	if dim, ok := en.(*entity.Dimension); ok {
		// the dimension block is drawn in WCS without transformation
		b, err := d.Block(dim.BlockName)
		if err != nil || len(b.Entities) == 0 || depth >= maxInsertDepth {
			return dim.BBox()
		}
		mins := []float64{math.Inf(1), math.Inf(1), math.Inf(1)}
		maxs := []float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
		for _, e := range b.Entities {
			emins, emaxs := d.bbox(e, depth+1)
			for i := 0; i < 3; i++ {
				mins[i] = math.Min(mins[i], emins[i])
				maxs[i] = math.Max(maxs[i], emaxs[i])
			}
		}
		return mins, maxs
	}
	ins, ok := en.(*entity.Insert)
	if !ok {
		return en.BBox()
	}
	b, err := d.Block(ins.BlockName)
	if err != nil || len(b.Entities) == 0 || depth >= maxInsertDepth {
		return ins.BBox()
	}
	// each entity's box is transformed (tighter than transforming the box
	// of the whole block when the insert is rotated)
	var boxes [][2][]float64
	for _, e := range b.Entities {
		if _, ok := e.(*entity.AttributeDefinition); ok {
			continue // not displayed in block references (ATTRIBs are)
		}
		emins, emaxs := d.bbox(e, depth+1)
		if math.IsInf(emins[0], 1) {
			continue // empty
		}
		boxes = append(boxes, [2][]float64{emins, emaxs})
	}
	mins := []float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxs := []float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	// the transform is affine, so the first and last array elements bound
	// the whole array.
	cols := []int{0}
	if ins.Columns > 1 {
		cols = append(cols, ins.Columns-1)
	}
	rows := []int{0}
	if ins.Rows > 1 {
		rows = append(rows, ins.Rows-1)
	}
	for _, box := range boxes {
		for _, c := range cols {
			for _, r := range rows {
				for k := 0; k < 8; k++ {
					corner := []float64{box[0][0], box[0][1], box[0][2]}
					for i := 0; i < 3; i++ {
						if k&(1<<uint(i)) != 0 {
							corner[i] = box[1][i]
						}
					}
					p := ins.Transform(corner, b.Coord, c, r)
					for i := 0; i < 3; i++ {
						mins[i] = math.Min(mins[i], p[i])
						maxs[i] = math.Max(maxs[i], p[i])
					}
				}
			}
		}
	}
	for _, a := range ins.Attributes {
		amins, amaxs := a.BBox()
		for i := 0; i < 3; i++ {
			mins[i] = math.Min(mins[i], amins[i])
			maxs[i] = math.Max(maxs[i], amaxs[i])
		}
	}
	return mins, maxs
}

// Spline creates a clamped uniform B-spline from the control points p.
// The degree is 3, or len(p)-1 if fewer than 4 control points are given.
// The knot vector has len(p)+degree+1 values as required by the DXF reference.
func (d *Drawing) Spline(p [][]float64) (*entity.Spline, error) {
	if len(p) < 2 {
		return nil, errors.New("Spline needs 2 or more control points")
	}
	l := entity.NewSpline()
	d.applyCurrent(l)
	l.Controls = p
	if l.Degree > len(p)-1 {
		l.Degree = len(p) - 1
	}
	l.Knots = clampedKnots(len(p), l.Degree)
	d.AddEntity(l)
	return l, nil
}

// clampedKnots returns a clamped uniform knot vector on [0, 1].
func clampedKnots(n, degree int) []float64 {
	knots := make([]float64, n+degree+1)
	spans := n - degree
	for i := range knots {
		switch {
		case i <= degree:
			knots[i] = 0
		case i >= n:
			knots[i] = 1
		default:
			knots[i] = float64(i-degree) / float64(spans)
		}
	}
	return knots
}

// Ellipse creates a new full ELLIPSE with center, endpoint of the major axis
// relative to the center, and ratio of the minor axis to the major axis (0 < ratio <= 1).
// Set Start/End (radians) for an elliptical arc.
func (d *Drawing) Ellipse(center, majorAxis []float64, ratio float64) (*entity.Ellipse, error) {
	if len(center) < 3 || len(majorAxis) < 3 {
		return nil, errors.New("Ellipse needs 3D center and major axis")
	}
	if ratio <= 0 || ratio > 1 {
		return nil, fmt.Errorf("Ellipse ratio must be in (0, 1]: %v", ratio)
	}
	if majorAxis[0] == 0 && majorAxis[1] == 0 && majorAxis[2] == 0 {
		return nil, errors.New("Ellipse major axis is zero")
	}
	e := entity.NewEllipse()
	e.Center = []float64{center[0], center[1], center[2]}
	e.MajorAxis = []float64{majorAxis[0], majorAxis[1], majorAxis[2]}
	e.Ratio = ratio
	d.applyCurrent(e)
	d.AddEntity(e)
	return e, nil
}

// Hatch creates a new solid-filled HATCH bounded by the closed polygon
// of vertices (2D points). Use Hatch.SetPattern / SetUserPattern for
// patterns, and AddPolylinePath / AddEdgePath for more boundaries (holes).
func (d *Drawing) Hatch(vertices ...[]float64) (*entity.Hatch, error) {
	if len(vertices) < 3 {
		return nil, errors.New("Hatch needs 3 or more vertices")
	}
	vs := make([][]float64, len(vertices))
	for i, v := range vertices {
		if len(v) < 2 {
			return nil, errors.New("Hatch vertices need 2 coordinates")
		}
		vs[i] = []float64{v[0], v[1]}
	}
	h := entity.NewHatch()
	h.AddPolylinePath(vs, nil)
	d.applyCurrent(h)
	d.AddEntity(h)
	return h, nil
}

// applyCurrent sets the current layer and the current properties of the
// header ($CECOLOR, $CELTYPE, $CELWEIGHT, $CELTSCALE) to a new entity.
func (d *Drawing) applyCurrent(e entity.Entity) {
	e.SetLayer(d.CurrentLayer)
	if sp, ok := e.(entity.SpaceSelector); ok && (d.currentBlock == nil || d.inLayoutBlock) {
		sp.SetPaperSpace(d.paperSpace)
	}
	h := d.Header()
	if h.CEColor != entity.ColorByLayer && h.CEColor >= 0 && h.CEColor <= 256 {
		entity.SetColorIndex(e, h.CEColor)
	}
	if lp, ok := e.(entity.LineProperties); ok {
		if h.CELType != "" && !strings.EqualFold(h.CELType, "ByLayer") {
			if lt, err := d.LineType(h.CELType); err == nil {
				lp.SetLineType(lt)
			}
		}
		if h.CELWeight != table.LineWeightByLayer {
			lp.SetLineWeight(h.CELWeight)
		}
	}
	if h.CELTScale > 0 && h.CELTScale != 1 {
		e.SetLtscale(h.CELTScale)
	}
}

// syncHeader writes the current layer and text style to the header
// ($CLAYER, $TEXTSTYLE) before writing.
func (d *Drawing) syncHeader() {
	h := d.Header()
	if d.CurrentLayer != nil {
		h.CLayer = d.CurrentLayer.Name()
	}
	if d.CurrentStyle != nil {
		h.TextStyle = d.CurrentStyle.Name()
	}
	// the structure written is AC1015 (R2000): files read from older
	// versions (e.g. R12, AC1009) are written as AC1015
	if v := strings.ToUpper(strings.TrimSpace(h.Version)); len(v) == 6 && strings.HasPrefix(v, "AC") && v < "AC1015" {
		h.Version = "AC1015"
	}
}

// ensureAppIDs adds the applications of extended data (1001) of entities
// and table entries to the APPID table if they are not registered.
func (d *Drawing) ensureAppIDs() {
	apps := d.Sections[TABLES].(table.Tables)[table.APPID]
	register := func(x [][2]string) {
		for _, p := range x {
			if p[0] != "1001" {
				continue
			}
			name := strings.TrimSpace(p[1])
			if _, err := apps.Contains(name); err != nil && name != "" {
				apps.Add(table.NewAppID(name))
			}
		}
	}
	visit := func(es entity.Entities) {
		for _, e := range es {
			if x, ok := e.(entity.XDataHolder); ok {
				register(x.XData())
			}
			switch c := e.(type) {
			case *entity.Polyline:
				for _, v := range c.Vertices {
					register(v.XData())
				}
			case *entity.Insert:
				for _, a := range c.Attributes {
					register(a.XData())
				}
			}
		}
	}
	visit(d.Entities())
	for _, b := range d.Blocks() {
		visit(b.Entities)
	}
	for _, t := range d.Sections[TABLES].(table.Tables) {
		for _, st := range t.Entries() {
			switch r := st.(type) {
			case *table.Layer:
				register(r.Extra)
				if r.HasTransparency() {
					register([][2]string{{"1001", table.TransparencyApp}})
				}
			case *table.View:
				register(r.Extra)
			case *table.Ucs:
				register(r.Extra)
			case *table.DimStyle:
				register(r.Extra)
			case *table.BlockRecord:
				register(r.Extra)
			}
		}
	}
}

// ensureDefaults adds the dimension style of $DIMSTYLE if it is not in the
// DIMSTYLE table (e.g. a file without it was read).
func (d *Drawing) ensureDefaults() {
	h := d.Header()
	styles := d.Sections[TABLES].(table.Tables)[table.DIMSTYLE]
	if _, err := styles.Contains(h.DimStyle); err != nil && h.DimStyle != "" {
		styles.Add(table.NewDimStyle(h.DimStyle))
	}
}

// SetHandleBase sets the first handle for objects without a handle kept
// from the file they were read from. Readers set it above the largest
// handle of the file.
func (d *Drawing) SetHandleBase(h int) {
	if h > d.handleBase {
		d.handleBase = h
	}
}

// SetObjectSourceHandle sets the handle, in the file read from, of an
// object this package creates: the root dictionary (key ""), the
// dictionaries "ACAD_GROUP", "ACAD_LAYOUT", "ACAD_PLOTSTYLENAME",
// "ACAD_PDFDEFINITIONS" (also DWF and DGN), "ACAD_IMAGE_DICT" and the
// default plot style "ACAD_PLOTSTYLENAME/Normal". References to them by
// handle from objects kept as read stay valid.
func (d *Drawing) SetObjectSourceHandle(key, h string) {
	switch key {
	case "":
		d.dictionary.SetSourceHandle(h)
	case "ACAD_GROUP":
		d.groupdict.SetSourceHandle(h)
	case "ACAD_LAYOUT":
		d.layoutdict.SetSourceHandle(h)
	case "ACAD_PLOTSTYLENAME":
		d.plotstyles.SetSourceHandle(h)
	case "ACAD_IMAGE_DICT":
		d.imageDictionary().SetSourceHandle(h)
	case "ACAD_PDFDEFINITIONS", "ACAD_DWFDEFINITIONS", "ACAD_DGNDEFINITIONS":
		d.underlayDictionary(key[5:8]).SetSourceHandle(h)
	case "ACAD_PLOTSTYLENAME/Normal":
		if p, ok := d.PlotStyle.(interface{ SetSourceHandle(string) }); ok {
			p.SetSourceHandle(h)
		}
	}
}

// AddRawObject adds an object kept as read to the OBJECTS section; with
// a key, it is also an item of the root dictionary.
func (d *Drawing) AddRawObject(r *object.Raw, key string) {
	d.addObject(r)
	if key != "" {
		d.dictionary.AddItem(key, r)
	}
}

// RawObjects returns the objects kept as read.
func (d *Drawing) RawObjects() []*object.Raw {
	var rs []*object.Raw
	for _, o := range d.Sections[OBJECTS].(object.Objects) {
		if r, ok := o.(*object.Raw); ok {
			rs = append(rs, r)
		}
	}
	return rs
}

// SetVersion sets the DXF version written ($ACADVER): format.R2000
// (AC1015, default), format.R2004 (AC1018), format.R2007 (AC1021),
// format.R2010 (AC1024), format.R2013 (AC1027) or format.R2018 (AC1032).
// Group codes introduced in later versions (e.g. true color 420 in AC1018,
// transparency 440 in AC1018) are written only for those versions, and
// text is written in UTF-8 for AC1021 and later.
func (d *Drawing) SetVersion(v string) error {
	v = strings.ToUpper(strings.TrimSpace(v))
	for _, s := range format.Versions() {
		if v == s {
			d.Header().Version = v
			return nil
		}
	}
	return fmt.Errorf("unsupported DXF version %q (supported: %s)", v, strings.Join(format.Versions(), ", "))
}

// Version returns the DXF version written ($ACADVER).
func (d *Drawing) Version() string {
	return d.Header().Version
}

// ensureClasses adds the classes of the object types this package writes
// (class.Defaults) if they are not defined (e.g. a file without them was
// read).
func (d *Drawing) ensureClasses() {
	cs := d.Sections[CLASSES].(class.Classes)
	for _, def := range class.Defaults() {
		found := false
		for _, c := range cs {
			if strings.EqualFold(c.Name(), def.Name()) {
				found = true
				break
			}
		}
		if !found {
			cs = append(cs, def)
		}
	}
	d.Sections[CLASSES] = cs
}
