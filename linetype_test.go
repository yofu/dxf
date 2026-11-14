package dxf

import (
	"strings"
	"testing"

	"github.com/yofu/dxf/table"
)

func TestComplexLineType(t *testing.T) {
	d := NewDrawing()
	st := d.CurrentStyle
	shx := d.AddShapeFile("ltypeshp.shx")
	if d.AddShapeFile("LTYPESHP.SHX") != shx {
		t.Error("the shape file should be reused")
	}
	gas, err := d.AddComplexLineType("GAS_LINE", "Gas line ----GAS----",
		&table.LineTypeElement{Length: 0.5},
		table.TextElement(-0.2, "GAS", st, 0.1, 0, -0.1, -0.05),
		&table.LineTypeElement{Length: -0.25})
	if err != nil {
		t.Fatal(err)
	}
	d.AddComplexLineType("TRACKS", "Tracks -|-|-|-",
		table.ShapeElement(0.25, 132, shx, 0.25, 0, 0, 0))
	if !gas.IsComplex() || gas.TotalLength() != 0.95 {
		t.Errorf("complex %v, total length %v", gas.IsComplex(), gas.TotalLength())
	}
	l, _ := d.Line(0, 0, 0, 10, 0, 0)
	l.SetLineType(gas)

	out := writeString(t, d)
	checkStructure(t, d)
	sth := "\n340\n" + hexHandle(st.Handle()) + "\n"
	if !strings.Contains(out, groups(49, -0.2, 74, 2, 75, 0)+"340") || !strings.Contains(out, sth) ||
		!strings.Contains(out, "\n9\nGAS\n") || !strings.Contains(out, groups(74, 4, 75, 132)) ||
		!strings.Contains(out, "\n340\n"+hexHandle(shx.Handle())+"\n") {
		t.Errorf("complex elements not written:\n%s", out)
	}

	d2 := readPreserve(t, out)
	g, err := d2.LineType("GAS_LINE")
	if err != nil {
		t.Fatal(err)
	}
	e := g.Elements()[1]
	st2 := d2.Styles["Standard"]
	if len(g.Elements()) != 3 || e.Text != "GAS" || e.Style != st2 || e.Scale != 0.1 || e.Offset[0] != -0.1 || e.Offset[1] != -0.05 || e.Flags != table.LT_TEXT {
		t.Errorf("read element: %+v", e)
	}
	tr, _ := d2.LineType("TRACKS")
	if s := tr.Elements()[0]; s.Shape != 132 || s.Style == nil || s.Style.FontName != "ltypeshp.shx" || s.Style.Flags != 1 {
		t.Errorf("read shape element: %+v", s)
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}

	// a new style gets a handle; the element refers to it
	d2.SetHandleBase(0x1000)
	ns, _ := d2.AddStyle("LT_TEXT", "romans.shx", "", false)
	e.Style = ns
	out2 := writeString(t, d2)
	if !strings.Contains(out2, "\n340\n"+hexHandle(ns.Handle())+"\n46\n") || ns.Handle() == 0 {
		t.Errorf("new style handle %X not referenced", ns.Handle())
	}
}

// unresolved style handles are written as read
func TestComplexLineTypeUnresolved(t *testing.T) {
	src := "0\nSECTION\n2\nTABLES\n0\nTABLE\n2\nLTYPE\n70\n1\n" +
		"0\nLTYPE\n5\n14\n2\nX\n70\n0\n3\nx\n72\n65\n73\n2\n40\n1.5\n49\n1.0\n74\n0\n49\n-0.5\n74\n3\n75\n0\n340\nABC\n46\n0.2\n50\n0.5\n44\n0\n45\n0\n9\nT\n" +
		"0\nENDTAB\n0\nENDSEC\n0\nEOF\n"
	d := readPreserve(t, src)
	lt, _ := d.LineType("X")
	e := lt.Elements()[1]
	if e.Style != nil || e.StyleHandle() != "ABC" || e.Rotation != 0.5 || e.Flags&table.LT_ABSOLUTE_ROTATION == 0 {
		t.Errorf("element: %+v", e)
	}
	if out := writeString(t, d); !strings.Contains(out, "\n340\nABC\n46\n") {
		t.Error("unresolved handle should be written as read")
	}
}
