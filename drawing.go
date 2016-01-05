package dxf

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/yofu/dxf/block"
	"github.com/yofu/dxf/class"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
	"github.com/yofu/dxf/header"
	"github.com/yofu/dxf/object"
	"github.com/yofu/dxf/table"
	"os"
	"strings"
)

var (
	DefaultColor    = color.White
	DefaultLineType = table.LT_CONTINUOUS
)

type Drawing struct {
	FileName     string
	Layers       map[string]*table.Layer
	Groups       map[string]*object.Group
	Styles       map[string]*table.Style
	CurrentLayer *table.Layer
	CurrentStyle *table.Style
	formatter    *format.Formatter
	sections     []Section
	dictionary   *object.Dictionary
	groupdict    *object.Dictionary
	plotstyle    handle.Handler
}

func NewDrawing() *Drawing {
	d := new(Drawing)
	d.Layers = make(map[string]*table.Layer)
	d.Layers["0"] = table.LY_0
	d.Groups = make(map[string]*object.Group)
	d.CurrentLayer = d.Layers["0"]
	d.Styles = make(map[string]*table.Style)
	d.Styles["STANDARD"] = table.ST_STANDARD
	d.CurrentStyle = d.Styles["STANDARD"]
	d.formatter = format.New()
	d.formatter.SetPrecision(16)
	d.sections = []Section{
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
	d.plotstyle = ph
	d.Layers["0"].SetPlotStyle(d.plotstyle)
	return d
}

func Open(filename string) (*Drawing, error) {
	var err error
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	d := NewDrawing()
	var code, value string
	var reader Section
	data := make([][2]string, 0)
	setreader := false
	line := 0
	startline := 0
	for scanner.Scan() {
		line++
		if line%2 == 1 {
			code = strings.TrimSpace(scanner.Text())
			if err != nil {
				return d, err
			}
		} else {
			value = scanner.Text()
			if setreader {
				if code != "2" {
					return d, fmt.Errorf("line %d: invalid groupe code: %s", line, code)
				}
				ind := SectionTypeValue(strings.ToUpper(value))
				if ind < 0 {
					return d, fmt.Errorf("line %d: unknown section name: %s", line, value)
				}
				reader = d.sections[ind]
				startline = line + 1
				setreader = false
			} else {
				if code == "0" {
					switch strings.ToUpper(value) {
					case "EOF":
						return d, nil
					case "SECTION":
						setreader = true
					case "ENDSEC":
						err := reader.Read(startline, data)
						if err != nil {
							return d, err
						}
						data = make([][2]string, 0)
						startline = line + 1
					default:
						data = append(data, [2]string{code, scanner.Text()})
					}
				} else {
					data = append(data, [2]string{code, scanner.Text()})
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return d, err
	}
	if len(data) > 0 {
		err := reader.Read(startline, data)
		if err != nil {
			return d, err
		}
	}
	return d, nil
}

func (d *Drawing) saveFile(filename string) error {
	d.setHandle()
	d.formatter.Reset()
	for _, s := range d.sections {
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

func (d *Drawing) Save() error {
	if d.FileName == "" {
		return errors.New("filename is blank, use SaveAs(filename)")
	}
	return d.saveFile(d.FileName)
}

func (d *Drawing) SaveAs(filename string) error {
	d.FileName = filename
	return d.saveFile(filename)
}

func (d *Drawing) setHandle() {
	h := 1
	for _, s := range d.sections[1:] {
		s.SetHandle(&h)
	}
	d.sections[0].SetHandle(&h)
}

func (d *Drawing) Layer(name string, cl color.ColorNumber, lt *table.LineType, setcurrent bool) (*table.Layer, error) {
	if l, exist := d.Layers[name]; exist {
		if setcurrent {
			d.CurrentLayer = l
		}
		return l, errors.New(fmt.Sprintf("layer %s already exists", name))
	}
	l := table.NewLayer(name, cl, lt)
	l.SetPlotStyle(d.plotstyle)
	d.Layers[name] = l
	d.sections[2].(table.Tables).AddLayer(l)
	if setcurrent {
		d.CurrentLayer = l
	}
	return l, nil
}

func (d *Drawing) ChangeLayer(name string) error {
	if l, exist := d.Layers[name]; exist {
		d.CurrentLayer = l
		return nil
	}
	return errors.New(fmt.Sprintf("layer %s doesn't exist", name))
}

func (d *Drawing) addEntity(e entity.Entity) {
	d.sections[4] = d.sections[4].(entity.Entities).Add(e)
}

func (d *Drawing) Point(x, y, z float64) (*entity.Point, error) {
	p := entity.NewPoint()
	p.Coord = []float64{x, y, z}
	p.SetLayer(d.CurrentLayer)
	d.addEntity(p)
	return p, nil
}

func (d *Drawing) Line(x1, y1, z1, x2, y2, z2 float64) (*entity.Line, error) {
	l := entity.NewLine()
	l.Start = []float64{x1, y1, z1}
	l.End = []float64{x2, y2, z2}
	l.SetLayer(d.CurrentLayer)
	d.addEntity(l)
	return l, nil
}

func (d *Drawing) Circle(x, y, z, r float64) (*entity.Circle, error) {
	c := entity.NewCircle()
	c.Center = []float64{x, y, z}
	c.Radius = r
	c.SetLayer(d.CurrentLayer)
	d.addEntity(c)
	return c, nil
}

func (d *Drawing) Polyline(closed bool, vertices ...[]float64) (*entity.Polyline, error) {
	p := entity.NewPolyline()
	p.SetLayer(d.CurrentLayer)
	for _, v := range vertices {
		p.AddVertex(v[0], v[1], v[2])
	}
	if closed {
		p.Close()
	}
	d.addEntity(p)
	return p, nil
}

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
	d.addEntity(l)
	return l, nil
}

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
	d.addEntity(f)
	return f, nil
}

func (d *Drawing) Text(str string, x, y, z, height float64) (*entity.Text, error) {
	t := entity.NewText()
	t.Coord1 = []float64{x, y, z}
	t.Height = height
	t.Value = str
	t.SetLayer(d.CurrentLayer)
	t.Style = d.CurrentStyle
	d.addEntity(t)
	return t, nil
}

func (d *Drawing) addObject(o object.Object) {
	d.sections[5] = d.sections[5].(object.Objects).Add(o)
}

func (d *Drawing) Group(name, desc string, es ...entity.Entity) (*object.Group, error) {
	if g, exist := d.Groups[name]; exist {
		g.AddEntity(es...)
		return g, errors.New(fmt.Sprintf("group %s already exists", name))
	}
	g := object.NewGroup(name, desc, es...)
	d.Groups[name] = g
	g.SetOwner(d.groupdict)
	d.addObject(g)
	return g, nil
}

func (d *Drawing) AddToGroup(name string, es ...entity.Entity) error {
	if g, exist := d.Groups[name]; exist {
		g.AddEntity(es...)
	}
	return errors.New(fmt.Sprintf("group %s doesn't exist", name))
}

func ColorIndex(cl []int) color.ColorNumber {
	minind := 0
	minval := 1000000
	for i, c := range color.ColorRGB {
		tmpval := 0
		for j := 0; j < 3; j++ {
			tmpval += (cl[j] - int(c[j])) * (cl[j] - int(c[j]))
		}
		if tmpval < minval {
			minind = i
			minval = tmpval
			if minval == 0 {
				break
			}
		}
	}
	return color.ColorNumber(minind)
}

func IndexColor(index uint8) []uint8 {
	return color.ColorRGB[index]
}
