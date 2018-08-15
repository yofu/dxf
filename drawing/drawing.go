// Package drawing defines Drawing struct for DXF.
package drawing

import (
	"errors"
	"fmt"
	"os"

	"github.com/yofu/dxf/block"
	"github.com/yofu/dxf/class"
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
	PlotStyle    handle.Handler
}

// New creates a new Drawing.
func New() *Drawing {
	d := new(Drawing)
	d.Layers = make(map[string]*table.Layer)
	d.Layers["0"] = table.LY_0
	d.Groups = make(map[string]*object.Group)
	d.CurrentLayer = d.Layers["0"]
	d.Styles = make(map[string]*table.Style)
	d.Styles["STANDARD"] = table.ST_STANDARD
	d.CurrentStyle = d.Styles["STANDARD"]
	d.formatter = format.NewASCII()
	d.formatter.SetPrecision(16)
	d.Sections = []Section{
		header.New(),
		class.New(),
		table.New(),
		block.New(),
		entity.New(),
		object.New(),
	}
	d.dictionary = object.NewDictionary()
	d.addObject(d.dictionary)
	wd, ph := object.NewAcDbDictionaryWDFLT(d.dictionary)
	d.dictionary.AddItem("ACAD_PLOTSTYLENAME", wd)
	d.addObject(wd)
	d.addObject(ph)
	d.groupdict = object.NewDictionary()
	d.addObject(d.groupdict)
	d.dictionary.AddItem("ACAD_GROUP", d.groupdict)
	d.PlotStyle = ph
	d.Layers["0"].SetPlotStyle(d.PlotStyle)
	return d
}

func (d *Drawing) saveFile(filename string) error {
	d.setHandle()
	d.formatter.Reset()
	for _, s := range d.Sections {
		s.WriteTo(d.formatter)
	}
	d.formatter.WriteString(0, "EOF")
	w, err := os.Create(filename)
	defer w.Close()
	if err != nil {
		return err
	}
	d.formatter.WriteTo(w)
	return nil
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

// setHandle sets all the handles contained in Drawing.
func (d *Drawing) setHandle() {
	h := 1
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

// Entities returns slice of all entities contained in Drawing.
func (d *Drawing) Entities() entity.Entities {
	return d.Sections[ENTITIES].(entity.Entities)
}

// AddEntity adds a new entity.
func (d *Drawing) AddEntity(e entity.Entity) {
	d.Sections[4] = d.Sections[4].(entity.Entities).Add(e)
}

// Point creates a new POINT at (x, y, z).
func (d *Drawing) Point(x, y, z float64) (*entity.Point, error) {
	p := entity.NewPoint()
	p.Coord = []float64{x, y, z}
	p.SetLayer(d.CurrentLayer)
	d.AddEntity(p)
	return p, nil
}

// Line creates a new LINE from (x1, y1, z1) to (x2, y2, z2).
func (d *Drawing) Line(x1, y1, z1, x2, y2, z2 float64) (*entity.Line, error) {
	l := entity.NewLine()
	l.Start = []float64{x1, y1, z1}
	l.End = []float64{x2, y2, z2}
	l.SetLayer(d.CurrentLayer)
	d.AddEntity(l)
	return l, nil
}

// Circle creates a new CIRCLE at (x, y, z) with radius r.
func (d *Drawing) Circle(x, y, z, r float64) (*entity.Circle, error) {
	c := entity.NewCircle()
	c.Center = []float64{x, y, z}
	c.Radius = r
	c.SetLayer(d.CurrentLayer)
	d.AddEntity(c)
	return c, nil
}

// Polyline creates a new POLYLINE with given vertices.
func (d *Drawing) Polyline(closed bool, vertices ...[]float64) (*entity.Polyline, error) {
	p := entity.NewPolyline()
	p.SetLayer(d.CurrentLayer)
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
	l.SetLayer(d.CurrentLayer)
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
	f.SetLayer(d.CurrentLayer)
	d.AddEntity(f)
	return f, nil
}

// Text creates a new TEXT str at (x, y, z) with given height.
func (d *Drawing) Text(str string, x, y, z, height float64) (*entity.Text, error) {
	t := entity.NewText()
	t.Coord1 = []float64{x, y, z}
	t.Height = height
	t.Value = str
	t.SetLayer(d.CurrentLayer)
	t.Style = d.CurrentStyle
	t.WidthFactor = t.Style.WidthFactor
	t.ObliqueAngle = t.Style.ObliqueAngle
	t.Style.LastHeightUsed = height
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

// AddGroup adds given entities to the named group.
// If the named group doesn't exist, returns error.
func (d *Drawing) AddToGroup(name string, es ...entity.Entity) error {
	if g, exist := d.Groups[name]; exist {
		g.AddEntity(es...)
	}
	return fmt.Errorf("group %s doesn't exist", name)
}

func (d *Drawing) SetExt() {
	mins := []float64{1e16, 1e16, 1e16}
	maxs := []float64{-1e16, -1e16, -1e16}
	for _, en := range d.Entities() {
		tmpmins, tmpmaxs := en.BBox()
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
