package dxf

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/yofu/dxf/block"
	"github.com/yofu/dxf/class"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/header"
	"github.com/yofu/dxf/object"
	"github.com/yofu/dxf/table"
	"os"
)

var (
	DefaultColor    = color.White
	DefaultLineType = table.LT_CONTINUOUS
)

type Drawing struct {
	FileName     string
	Layers       map[string]*table.Layer
	CurrentLayer *table.Layer
	sections     []Section
}

func NewDrawing() *Drawing {
	d := new(Drawing)
	d.Layers = make(map[string]*table.Layer)
	d.Layers["0"] = table.LY_0
	d.CurrentLayer = d.Layers["0"]
	d.sections = []Section{
		header.New(),
		class.New(),
		table.New(),
		block.New(),
		entity.New(),
		object.New(),
	}
	return d
}

func (d *Drawing) saveFile(filename string) error {
	d.setHandle()
	var otp bytes.Buffer
	for _, s := range d.sections {
		s.WriteTo(&otp)
	}
	otp.WriteString("0\nEOF\n")
	w, err := os.Create(filename)
	defer w.Close()
	if err != nil {
		return err
	}
	otp.WriteTo(w)
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
	h := 0
	for _, s := range d.sections {
		s.SetHandle(&h)
	}
}

func (d *Drawing) Layer(name string, cl color.ColorNumber, lt *table.LineType, setcurrent bool) (*table.Layer, error) {
	if l, exist := d.Layers[name]; exist {
		if setcurrent {
			d.CurrentLayer = l
		}
		return l, errors.New(fmt.Sprintf("layer %s already exists", name))
	}
	l := table.NewLayer(name, cl, lt)
	d.Layers[name] = l
	d.sections[2].(*table.Tables).AddLayer(l)
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

func (d *Drawing) Point(x, y, z float64) (*entity.Point, error) {
	p := entity.NewPoint()
	p.Coord = []float64{x, y, z}
	p.SetLayer(d.CurrentLayer)
	d.sections[4].(*entity.Entities).Add(p)
	return p, nil
}

func (d *Drawing) Line(x1, y1, z1, x2, y2, z2 float64) (*entity.Line, error) {
	l := entity.NewLine()
	l.Start = []float64{x1, y1, z1}
	l.End = []float64{x2, y2, z2}
	l.SetLayer(d.CurrentLayer)
	d.sections[4].(*entity.Entities).Add(l)
	return l, nil
}

func (d *Drawing) Circle(x, y, z, r float64) (*entity.Circle, error) {
	c := entity.NewCircle()
	c.Center = []float64{x, y, z}
	c.Radius = r
	c.SetLayer(d.CurrentLayer)
	d.sections[4].(*entity.Entities).Add(c)
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
	d.sections[4].(*entity.Entities).Add(p)
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
	d.sections[4].(*entity.Entities).Add(l)
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
	d.sections[4].(*entity.Entities).Add(f)
	return f, nil
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
