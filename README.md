# dxf

DXF (Drawing Exchange Format by Autodesk, Inc.) library for Go.

- Writes drawings in the structure of AutoCAD 2000 (AC1015), as ASCII or
  binary DXF.
- Reads ASCII and binary DXF files of AutoCAD R12 and later. Data this
  package does not interpret is kept, and handles are kept, when a drawing
  read is written again.
- Entities: LINE, CIRCLE, ARC, ELLIPSE, POINT, TEXT, MTEXT, 3DFACE, SOLID,
  LWPOLYLINE, POLYLINE (2D, 3D, meshes), SPLINE, INSERT, ATTDEF, ATTRIB,
  HATCH, DIMENSION, LEADER and VIEWPORT. Other entity types are kept as read
  (`entity.Unknown`).
- Blocks with attributes, groups, layouts and viewports, extended data,
  text in code pages (e.g. Shift_JIS for Japanese), and editing (query,
  remove, move, rotate, scale, copy, explode).
- Validated with [ezdxf](https://ezdxf.mozman.at/) (see
  [tools/ezdxfcheck](tools/ezdxfcheck/README.md)).

## Installation

```
$ go get github.com/yofu/dxf
```

## Quick start

```go
d := dxf.NewDrawing()
d.AddLayer("WALL", color.Red, dxf.DefaultLineType, true) // and make it current
d.Line(0, 0, 0, 1000, 0, 0)
d.Circle(500, 500, 0, 250)
d.Text("Plan", 0, -100, 0, 50)
d.SetExt() // $EXTMIN, $EXTMAX
if err := d.SaveAs("plan.dxf"); err != nil {
	log.Fatal(err)
}

d2, err := dxf.FromFile("plan.dxf")
if err != nil {
	log.Fatal(err)
}
for _, e := range d2.EntitiesOnLayer("WALL") {
	fmt.Printf("%T\n", e)
}
```

The examples of this README are in [example_test.go](example_test.go).

![Torus](example/torus/torus.png)

The torus above is drawn by [example/torus](example/torus/torus.go).

## Drawing

### Entity properties

Color, linetype and lineweight are BYLAYER by default. New entities created
by `Drawing` methods are on the current layer (`ChangeLayer`) and get the
current properties of the header (`$CECOLOR`, `$CELTYPE`, `$CELWEIGHT`).

```go
l, _ := d.Line(0, 0, 0, 100, 0, 0)
l.SetColor(color.Red)      // 62 (SetColorByBlock / SetColorByLayer)
lt, _ := d.LineType("HIDDEN")
l.SetLineType(lt)          // 6 (nil: BYLAYER)
l.SetLineWeight(35)        // 370: 0.35 mm
l.Thickness = 10           // 39
l.SetVisible(false)        // 60
l.SetTrueColor(255, 128, 0) // 420 (AC1018 or later); 62 is set to the nearest color number
l.SetTransparency(0.5)      // 440 (AC1018 or later); also layers: layer.SetTransparency
l.SetXData([][2]string{{"1001", "MYAPP"}, {"1000", "data"}}) // APPID is registered when writing
```

### Blocks and attributes

```go
d.BeginBlock("TITLE", 0, 0, 0) // entities below are added to the block
d.LwPolyline(true, []float64{0, 0}, []float64{180, 0}, []float64{180, 40}, []float64{0, 40})
d.AttributeDefinition("NAME", "Drawing name", "untitled", 10, 15, 0, 5) // tag, prompt, default
d.EndBlock()

ins, _ := d.Insert("TITLE", 230, 10, 0) // attributes are created from the definitions
ins.SetAttribute("NAME", "Floor plan")
ins.Scale = []float64{2, 2, 2}
d.SyncAttributes(ins) // after changing the insertion point, scale or rotation
```

### Text, dimensions, leaders and hatches

```go
mt, _ := d.MText("first line\nsecond line", 0, 0, 0, 2.5, 0) // "\n" -> \P
mt.Attachment = entity.MTEXT_MIDDLE_CENTER
fmt.Println(mt.PlainText()) // without MTEXT formatting codes
mt.SetBackgroundColor(color.Yellow, 1.5) // background fill (AC1018 or later);
// also SetBackgroundTrueColor, SetBackgroundCanvas, SetTextFrame, ClearBackground

// the graphics are generated in an anonymous block *Dn with the
// dimension style (d.DimStyle("Standard"): arrow size, text height, ...)
dim, _ := d.LinearDimension([]float64{0, 0, 0}, []float64{100, 0, 0}, []float64{50, 20, 0}, 0)
dim.Text = "<> mm" // "<>" is the measurement
d.RegenerateDimension(dim)
// also AlignedDimension, RadiusDimension, DiameterDimension, AngularDimension

d.LeaderText("Note", 0, []float64{0, 0, 0}, []float64{10, 10, 0})

h, _ := d.Hatch([]float64{0, 0}, []float64{100, 0}, []float64{100, 50}) // solid fill
h.SetPattern("ANSI31", 0, 1) // or SetUserPattern(angle, spacing, double)
// gradient fill (AC1018 or later; a solid fill for older versions)
h.SetGradient(entity.GRADIENT_LINEAR, color.RGB(255, 0, 0), color.RGB(0, 0, 255), 30)
// or SetOneColorGradient(name, rgb, tint, rotation)

e, _ := d.Ellipse([]float64{0, 0, 0}, []float64{10, 0, 0}, 0.5)
e.Start, e.End = 0, math.Pi // elliptical arc
```

### Layouts and viewports

```go
d.PaperSpace() // following entities are in the active layout (Layout1)
d.LwPolyline(true, []float64{10, 10}, []float64{410, 10}, []float64{410, 287}, []float64{10, 287})
// a viewport of 380 x 250 on paper, showing model space around (5000, 2500)
// with 6000 units in its height
d.Viewport([]float64{210, 148.5, 0}, 380, 250, []float64{5000, 2500}, 6000)
d.ModelSpace()

d.RenameLayout("Layout1", "Plan")
d.AddLayout("Section") // its entities are in the block *Paper_Space0
d.UseLayout("Section")
d.Text("Section", 20, 20, 0, 5)
d.UseLayout("Model")
```

### Editing

```go
walls := d.EntitiesOnLayer("WALL") // or d.Query(func(e entity.Entity) bool { ... })
d.Move(10, 0, 0, walls...)         // also Rotate(base, angle, ...), Scale(base, factor, ...)
copies, _ := dxf.CopyEntities(d, entity.Rotate([]float64{0, 0, 0}, 90), walls...)
d.RemoveEntity(walls[0])
exploded, _ := dxf.Explode(d, insert) // block reference -> entities
d.RenameLayer("WALL", "A-WALL")       // also RemoveLayer, RenameBlock, RemoveBlock
```

Transformations are moves, rotations about the Z axis and uniform scaling
(`entity.Transform`); entities with an extrusion direction other than the
Z axis cannot be transformed.

## Files

### Versions

Drawings are written as AC1015 (AutoCAD 2000) by default. Other versions
can be chosen; group codes introduced in later versions (true color,
transparency, ...) are written only for those versions, and text is
written in UTF-8 for AC1021 and later.

```go
d.SetVersion(format.R2010) // R2000 (AC1015), R2004, R2007, R2010, R2013, R2018 (AC1032)
```

Files of later versions keep their version when read and written again.
The structure written is that of AC1015 with the additions of later
versions this package knows (true color, transparency, HATCH gradients
and MTEXT background fill from AC1018,
BLOCK_RECORD units/explodability from AC1021, class instance counts from
AC1018); output of each version is validated with ezdxf. Objects and
entities of later versions read from a file are kept as read.

### Text encoding

Up to AC1018, text in DXF files is stored in the code page of
`$DWGCODEPAGE`. By default (`ANSI_1252`), non-ASCII characters are written
as `\U+XXXX`, which AutoCAD reads in any locale. Set the code page to write
text in it, e.g. Shift_JIS:

```go
d.Header().CodePage = codepage.Japanese // "ANSI_932"
```

When reading, text is decoded according to `$DWGCODEPAGE` (UTF-8 for
AC1021 and later), and `\U+XXXX` / `\M+nXXXX` escapes are converted. For
files without `$DWGCODEPAGE`:

```go
d, err := dxf.FromFileWithCodePage("sjis.dxf", codepage.Japanese)
```

### Binary DXF

```go
d.SaveAsBinary("drawing.dxf") // or d.SetFormatter(format.NewBinary()) and d.WriteTo(w)
d, err := dxf.FromFile("drawing.dxf") // binary DXF is detected by its sentinel
```

Values are written in the type of their group code. Binary DXF of R12
(1-byte group codes) can also be read.

### Reading and writing back

When a file is read and written again, the handles of the objects read are
kept (new objects get handles above them), so references by handle stay
valid. Data this package does not interpret is kept as well: unknown header
variables (`Header().Extra`), entities of unsupported types
(`entity.Unknown`, e.g. IMAGE, WIPEOUT, MLINE), extended data (`XData()`),
extension dictionaries and reactors (102 groups), other values of table
entries (`Extra`), the CLASSES section and the objects of the OBJECTS
section (`RawObjects()`). Groups (`Groups`) and layouts (`Layouts()`) are
read from the OBJECTS section.

Not kept: the plot settings and other values of LAYOUT objects (they are
written by this package). Files of versions older than AC1015 (e.g. R12)
are written as AC1015.

## Compatibility notes

Changes from earlier versions of this package that may affect existing code
or output:

- `entity.Vertex.Buldge` is renamed to `Bulge`.
- `LwPolyline.Bulges[i]` is the bulge of the segment from vertex i (as in
  the DXF reference); `LwPolyline.Bulge(i)` is the arc to reach vertex i.
- The header has more variables; `$MEASUREMENT` is metric by default, and
  the dimension style "Standard" has metric (ISO-25 like) values.
- The default blocks are `*Model_Space` and `*Paper_Space` (no
  `*Paper_Space0`); the structure of AC1015 is completed (owners, layouts,
  VPORT `*Active`, DIMSTYLE `Standard`).
- `SetExt` uses the entities of model space only.
- Unsupported entity types are kept as `entity.Unknown` (they were skipped
  or made reading fail).
- `table.LY_0`, `table.ST_STANDARD` and `table.LT_*` are templates: each
  drawing has its own layer "0", style "Standard" and linetypes.

## References

- DXF reference of AutoCAD R12:
  [acad_r12_dxf.pdf](https://damassets.autodesk.net/content/dam/autodesk/www/developer-network/platform-technologies/autocad-dxf-archive/acad_r12_dxf.pdf)
- DXF references of later versions: Autodesk developer documentation
  ("DXF Reference")
- [ezdxf](https://ezdxf.mozman.at/) documentation

# License

MIT

# Author

Yoshihiro FUKUSHIMA
