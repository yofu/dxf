package dxf

import (
	"bytes"
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
