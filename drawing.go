package dxf

import (
	"bytes"
	"errors"
	"os"
	"github.com/yofu/dxf/header"
	"github.com/yofu/dxf/class"
	"github.com/yofu/dxf/table"
	"github.com/yofu/dxf/block"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/object"
)

type Drawing struct {
	FileName string
	sections []Section
}

func NewDrawing() *Drawing {
	d := new(Drawing)
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

func (d *Drawing) Line(x1, y1, z1, x2, y2, z2 float64) (*entity.Line, error) {
	l := entity.NewLine()
	l.Start = []float64{x1, y1, z1}
	l.End = []float64{x2, y2, z2}
	d.sections[4].(*entity.Entities).Add(l)
	return l, nil
}

func (d *Drawing) ThreeDFace(points [][]float64) (*entity.ThreeDFace, error) {
	f := entity.New3DFace()
	if len(points) < 3 {
		return nil, errors.New("3DFace needs 3 or more points")
	}
	for i:=0; i<3; i++ {
		f.Points[i] = points[i]
	}
	if len(points) >= 4 {
		f.Points[3] = points[3]
	} else {
		f.Points[3] = points[2]
	}
	d.sections[4].(*entity.Entities).Add(f)
	return f, nil
}
