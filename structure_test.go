package dxf

import (
	"strings"
	"testing"

	"github.com/yofu/dxf/color"
)

func (r record) get(code int) (string, bool) {
	for _, p := range r.pairs {
		if p.code == code {
			return p.value, true
		}
	}
	return "", false
}

// last returns the last value of code (e.g. the owner 330 at the end of LAYOUT).
func (r record) last(code int) string {
	v := ""
	for _, p := range r.pairs {
		if p.code == code {
			v = p.value
		}
	}
	return v
}

func (r record) handle() string {
	if h, ok := r.get(5); ok {
		return h
	}
	h, _ := r.get(105)
	return h
}

// owner returns the owner handle: the first 330 outside of the {ACAD_REACTORS} group.
func (r record) owner() string {
	inGroup := false
	for _, p := range r.pairs {
		switch {
		case p.code == 102 && strings.HasPrefix(p.value, "{"):
			inGroup = true
		case p.code == 102 && p.value == "}":
			inGroup = false
		case p.code == 330 && !inGroup:
			return p.value
		}
	}
	return ""
}

func TestR2000Structure(t *testing.T) {
	d := NewDrawing()
	d.AddLayer("A", color.Red, DefaultLineType, true)
	d.BeginBlock("B", 0, 0, 0)
	d.Line(0, 0, 0, 1, 0, 0)
	d.Polyline(false, []float64{0, 0, 0}, []float64{1, 1, 1})
	d.EndBlock()
	l, _ := d.Line(0, 0, 0, 1, 1, 0)
	d.Insert("B", 5, 5, 0)
	d.Polyline(true, []float64{0, 0, 0}, []float64{1, 0, 0}, []float64{1, 1, 0})
	d.MText("text", 0, 0, 0, 1, 0)
	d.Group("G", "group", l)
	checkStructure(t, d)

	// written twice: owners and handles must stay consistent
	checkStructure(t, d)

	// read back and write again
	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	checkStructure(t, d2)

	// a file without BLOCK_RECORDs (e.g. R12 style)
	d3, err := FromStringData("0\nSECTION\n2\nBLOCKS\n0\nBLOCK\n8\n0\n2\nX\n70\n0\n10\n0\n20\n0\n0\nLINE\n8\n0\n10\n0\n20\n0\n11\n1\n21\n1\n0\nENDBLK\n8\n0\n0\nENDSEC\n" +
		"0\nSECTION\n2\nENTITIES\n0\nINSERT\n8\n0\n2\nX\n10\n0\n20\n0\n0\nENDSEC\n0\nEOF\n")
	if err != nil {
		t.Fatal(err)
	}
	checkStructure(t, d3)
}
