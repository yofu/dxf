package dxf

import (
	"strings"
	"testing"

	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/handle"
)

func TestHandlesAreKept(t *testing.T) {
	d := NewDrawing()
	d.Line(0, 0, 0, 1, 0, 0)
	d.Polyline(false, []float64{0, 0, 0}, []float64{1, 1, 1})
	d.BeginBlock("B", 0, 0, 0)
	d.AttributeDefinition("TAG", "", "v", 0, 0, 0, 1)
	d.EndBlock()
	d.Insert("B", 0, 0, 0)
	d.Group("G", "", d.Entities()[0])
	out := writeString(t, d)

	d2 := readPreserve(t, out)
	if writeString(t, d2) != out {
		t.Fatal("a drawing read and written again should be identical, including handles")
	}
	l := d2.Entities()[0].(*entity.Line)
	h := l.SourceHandle()
	if h == "" || !strings.Contains(out, "\n0\nLINE\n5\n"+h+"\n") {
		t.Errorf("source handle %q", h)
	}

	// new objects get handles above those of the file
	seed := 0
	if i := strings.Index(out, "$HANDSEED\n5\n"); i >= 0 {
		v := out[i+len("$HANDSEED\n5\n"):]
		seed = handle.Parse(v[:strings.Index(v, "\n")])
	}
	n, _ := d2.Circle(0, 0, 0, 1)
	c, _ := Copy(d2, l)
	d2.AddEntity(c)
	out2 := writeString(t, d2)
	checkStructure(t, d2)
	if n.Handle() < seed || c.Handle() < seed {
		t.Errorf("new handles %X %X should not be below $HANDSEED %X", n.Handle(), c.Handle(), seed)
	}
	if l.Handle() != handle.Parse(h) || !strings.Contains(out2, "\n0\nLINE\n5\n"+h+"\n") {
		t.Error("the handle of the line read should be kept")
	}
}
