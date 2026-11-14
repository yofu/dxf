package dxf_test

import (
	"bytes"
	"fmt"
	"math"

	"github.com/yofu/dxf"
	"github.com/yofu/dxf/codepage"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
)

// The examples of README.md.

func Example() {
	d := dxf.NewDrawing()
	d.AddLayer("壁", color.Red, dxf.DefaultLineType, true) // current layer
	d.Line(0, 0, 0, 1000, 0, 0)
	d.Circle(500, 500, 0, 250)
	d.Text("平面図", 0, -100, 0, 50)
	d.SetExt() // $EXTMIN, $EXTMAX
	var b bytes.Buffer
	if _, err := d.WriteTo(&b); err != nil { // or d.SaveAs("plan.dxf")
		panic(err)
	}

	d2, err := dxf.FromReader(&b) // or dxf.FromFile("plan.dxf")
	if err != nil {
		panic(err)
	}
	for _, e := range d2.EntitiesOnLayer("壁") {
		fmt.Printf("%T\n", e)
	}
	// Output:
	// *entity.Line
	// *entity.Circle
	// *entity.Text
}

func Example_properties() {
	d := dxf.NewDrawing()
	l, _ := d.Line(0, 0, 0, 100, 0, 0)
	l.SetColor(color.Red) // 62 (SetColorByBlock / SetColorByLayer)
	lt, _ := d.LineType("HIDDEN")
	l.SetLineType(lt)   // 6 (nil: BYLAYER)
	l.SetLineWeight(35) // 370: 0.35 mm
	l.Thickness = 10    // 39
	fmt.Println(l.ColorIndex(), l.LineTypeName(), l.LineWeight())
	// Output: 1 HIDDEN 35
}

func Example_blocks() {
	d := dxf.NewDrawing()
	d.BeginBlock("TITLE", 0, 0, 0) // entities below are added to the block
	d.LwPolyline(true, []float64{0, 0}, []float64{180, 0}, []float64{180, 40}, []float64{0, 40})
	d.AttributeDefinition("NAME", "Drawing name", "untitled", 10, 15, 0, 5) // tag, prompt, default
	d.EndBlock()

	ins, _ := d.Insert("TITLE", 230, 10, 0) // attributes are created from the definitions
	ins.SetAttribute("NAME", "Floor plan")
	ins.Scale = []float64{2, 2, 2}
	d.SyncAttributes(ins) // after changing the insertion point, scale or rotation
	fmt.Println(ins.Attribute("NAME").Value, ins.Attribute("NAME").Coord1)
	// Output: Floor plan [250 40 0]
}

func Example_annotations() {
	d := dxf.NewDrawing()
	mt, _ := d.MText("first line\nsecond line", 0, 0, 0, 2.5, 0) // "\n" -> \P
	mt.Attachment = entity.MTEXT_MIDDLE_CENTER

	dim, _ := d.LinearDimension([]float64{0, 0, 0}, []float64{100, 0, 0}, []float64{50, 20, 0}, 0)
	dim.Text = "<> mm" // "<>" is the measurement
	d.RegenerateDimension(dim)

	d.LeaderText("Note", 0, []float64{0, 0, 0}, []float64{10, 10, 0})

	h, _ := d.Hatch([]float64{0, 0}, []float64{100, 0}, []float64{100, 50}) // solid fill
	h.SetPattern("ANSI31", 0, 1)
	e, _ := d.Ellipse([]float64{0, 0, 0}, []float64{10, 0, 0}, 0.5)
	e.Start, e.End = 0, math.Pi // elliptical arc

	fmt.Println(mt.PlainText(), "/", dim.Measurement)
	// Output:
	// first line
	// second line / 100
}

func Example_layouts() {
	d := dxf.NewDrawing()
	d.Line(0, 0, 0, 10000, 5000, 0)
	d.PaperSpace() // the active layout (Layout1)
	d.LwPolyline(true, []float64{10, 10}, []float64{410, 10}, []float64{410, 287}, []float64{10, 287})
	d.Viewport([]float64{210, 148.5, 0}, 380, 250, []float64{5000, 2500}, 6000)
	d.ModelSpace()

	d.RenameLayout("Layout1", "Plan")
	d.AddLayout("Section") // block *Paper_Space0
	d.UseLayout("Section")
	d.Text("Section", 20, 20, 0, 5)
	d.UseLayout("Model")
	for _, l := range d.Layouts() {
		fmt.Println(l.Name, l.Block)
	}
	// Output:
	// Model *Model_Space
	// Plan *Paper_Space
	// Section *Paper_Space0
}

func Example_editing() {
	d := dxf.NewDrawing()
	d.AddLayer("WALL", color.Red, dxf.DefaultLineType, true)
	d.Line(0, 0, 0, 100, 0, 0)
	d.Line(100, 0, 0, 100, 100, 0)
	d.ChangeLayer("0")

	walls := d.EntitiesOnLayer("WALL")
	d.Move(10, 0, 0, walls...) // also Rotate(base, angle, ...), Scale(base, factor, ...)
	copies, _ := dxf.CopyEntities(d, entity.Rotate([]float64{0, 0, 0}, 90), walls...)
	d.RemoveEntity(walls[0])
	d.RenameLayer("WALL", "A-WALL")
	fmt.Println(len(d.Entities()), len(copies), copies[0].Layer().Name())
	// Output: 3 2 A-WALL
}

func Example_encoding() {
	d := dxf.NewDrawing()
	d.Text("日本語", 0, 0, 0, 10)
	var b bytes.Buffer
	d.WriteTo(&b) // default ANSI_1252: \U+65E5\U+672C\U+8A9E
	fmt.Println(bytes.Contains(b.Bytes(), []byte(`\U+65E5`)))

	d.Header().CodePage = codepage.Japanese // Shift_JIS
	b.Reset()
	d.WriteTo(&b)
	d2, _ := dxf.FromReader(&b) // decoded by $DWGCODEPAGE
	fmt.Println(d2.Entities()[0].(*entity.Text).Value)
	// Output:
	// true
	// 日本語
}

func Example_binary() {
	d := dxf.NewDrawing()
	d.Line(0, 0, 0, 1, 1, 0)
	d.SetFormatter(format.NewBinary()) // or d.SaveAsBinary("drawing.dxf")
	var b bytes.Buffer
	d.WriteTo(&b)
	sentinel := b.String()[:18]
	d2, _ := dxf.FromReader(&b) // binary DXF is detected
	fmt.Println(sentinel, len(d2.Entities()))
	// Output: AutoCAD Binary DXF 1
}
