#dxf

DXF(Drawing Exchange Format by Autodesk, Inc) library for golang.

+ version: ACAD2000(AC1015)
+ format : ASCII

is only supported.

## Reference

[http://www.autodesk.com/techpubs/autocad/acad2000/dxf/index.htm](http://www.autodesk.com/techpubs/autocad/acad2000/dxf/index.htm)

## Example
```go
package main

import (
	"github.com/yofu/dxf"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/table"
	"log"
	"math"
)

func main() {
	d := dxf.NewDrawing()
	d.Header().LtScale = 100.0
	d.AddLayer("Toroidal", dxf.DefaultColor, dxf.DefaultLineType, true)
	d.AddLayer("Poloidal", color.Red, table.LT_HIDDEN, true)
	z := 0.0
	r1 := 200.0
	r2 := 500.0
	ndiv := 16
	dtheta := 2.0 * math.Pi / float64(ndiv)
	theta := 0.0
	for i := 0; i < ndiv; i++ {
		d.ChangeLayer("Toroidal")
		d.Circle(0.0, 0.0, z+r1*math.Cos(theta), r2-r1*math.Sin(theta))
		d.ChangeLayer("Poloidal")
		c, _ := d.Circle(r2*math.Cos(theta), r2*math.Sin(theta), 0.0, r1)
		dxf.SetExtrusion(c, []float64{-1.0 * math.Sin(theta), math.Cos(theta), 0.0})
		theta += dtheta
	}
	err := d.SaveAs("torus.dxf")
	if err != nil {
		log.Fatal(err)
	}
}
```

![Torus](example/torus/torus.png)

### Blocks

```go
d := dxf.NewDrawing()
d.BeginBlock("BOLT", 0.0, 0.0, 0.0) // entities below are added to the block
d.Circle(0.0, 0.0, 0.0, 5.0)
d.Line(-5.0, 0.0, 0.0, 5.0, 0.0, 0.0)
d.EndBlock()

ins, _ := d.Insert("BOLT", 100.0, 50.0, 0.0) // INSERT (block reference)
ins.Scale = []float64{2.0, 2.0, 1.0}
ins.Rotation = 45.0
```

### Entity properties

Color, linetype and lineweight are BYLAYER by default.

```go
l, _ := d.Line(0.0, 0.0, 0.0, 100.0, 0.0, 0.0)
l.SetColor(color.Red)       // 62 (SetColorByBlock / SetColorByLayer)
lt, _ := d.LineType("HIDDEN")
l.SetLineType(lt)           // 6 (nil: BYLAYER)
l.SetLineWeight(35)         // 370: 0.35 mm (table.LineWeightByBlock, ...)
```

### Ellipse and hatch

```go
// center, endpoint of the major axis (relative to the center), minor/major ratio
e, _ := d.Ellipse([]float64{0.0, 0.0, 0.0}, []float64{10.0, 0.0, 0.0}, 0.5)
e.Start, e.End = 0.0, math.Pi // elliptical arc (parameters in radians)

h, _ := d.Hatch([]float64{0.0, 0.0}, []float64{100.0, 0.0}, []float64{100.0, 50.0}) // solid fill
h.SetPattern("ANSI31", 0.0, 1.0)                   // or SetUserPattern(angle, spacing, double)
h.AddPolylinePath([][]float64{{10, 10}, {20, 10}, {20, 20}}, nil) // inner boundary (hole)
```

### Multiline text (MTEXT)

```go
d := dxf.NewDrawing()
// "\n" is converted to the paragraph break \P; width 0 means no wrapping.
mt, _ := d.MText("first line\nsecond line", 0.0, 0.0, 0.0, 2.5, 0.0)
mt.Attachment = entity.MTEXT_MIDDLE_CENTER
mt.Rotation = 30.0 // degrees

fmt.Println(mt.PlainText()) // text without MTEXT formatting codes
```

### Text encoding (non-ASCII text)

Up to AC1018, text in DXF files is stored in the code page of `$DWGCODEPAGE`.

```go
d := dxf.NewDrawing()
// default "ANSI_1252": non-ASCII characters are written as \U+XXXX escapes,
// which AutoCAD reads in any locale.
d.Text("日本語", 0.0, 0.0, 0.0, 10.0)

// write text in Shift_JIS (e.g. for applications that do not read \U+XXXX)
d.Header().CodePage = codepage.Japanese // "ANSI_932"
```

When reading, text is decoded according to `$DWGCODEPAGE` (or as UTF-8 for
AC1021 and later), and `\U+XXXX` / `\M+nXXXX` escapes are converted.
For files without `$DWGCODEPAGE`, the code page can be given:

```go
d, err := dxf.FromFileWithCodePage("sjis.dxf", codepage.Japanese)
```

## Installation

```
$ go get github.com/yofu/dxf
```

# License

MIT

# Author

Yoshihiro FUKUSHIMA
