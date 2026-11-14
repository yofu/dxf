# dxf

DXF (Drawing Exchange Format by Autodesk, Inc.) library for Go.

- Writes drawings in the structure of AutoCAD 2000 (AC1015), as ASCII or
  binary DXF.
- Reads ASCII and binary DXF files of AutoCAD R12 and later. Data this
  package does not interpret is kept, and handles are kept, when a drawing
  read is written again.
- Entities: LINE, XLINE, RAY, CIRCLE, ARC, ELLIPSE, POINT, TEXT, MTEXT,
  SHAPE, 3DFACE, SOLID, TRACE, LWPOLYLINE, POLYLINE (2D, 3D, meshes), MESH,
  SPLINE, HELIX, INSERT, ATTDEF, ATTRIB, HATCH, MPOLYGON, DIMENSION
  (also ARC_DIMENSION and LARGE_RADIAL_DIMENSION), LEADER, MULTILEADER,
  MLINE, ACAD_TABLE (block and cell texts; the table data is kept), TOLERANCE,
  VIEWPORT, IMAGE, WIPEOUT, LIGHT, underlays (PDFUNDERLAY, DWFUNDERLAY,
  DGNUNDERLAY) and ACIS entities (3DSOLID, BODY, REGION and surfaces: the
  ACIS data as SAT text, the geometry is not interpreted). Other entity
  types are kept as read (`entity.Unknown`), as are sections this package
  does not interpret (e.g. ACDSDATA with the ACIS data of AC1027 and later,
  THUMBNAILIMAGE).
- Blocks with attributes, groups, layouts and viewports, extended data,
  text in code pages (e.g. Shift_JIS for Japanese), and editing (query,
  remove, move, rotate, scale, copy, explode).
- Files: R12 export, DWG through the ODA File Converter, streaming read,
  recovery of damaged files, audit, import between drawings, external
  references, proxy graphics of custom entities.
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

// linetypes: dashes (> 0), spaces (< 0) and dots (0)
d.AddLineType("DASHED2", "Dashed __ __ __", 0.5, -0.25)
// complex linetypes with text strings and shapes
gas, _ := d.AddComplexLineType("GAS_LINE", "Gas line ----GAS----",
	&table.LineTypeElement{Length: 0.5},
	table.TextElement(-0.2, "GAS", d.CurrentStyle, 0.1, 0, -0.1, -0.05), // scale, rotation (rad), offset
	&table.LineTypeElement{Length: -0.25})
shx := d.AddShapeFile("ltypeshp.shx")
d.AddComplexLineType("TRACKS", "Tracks -|-|-|-",
	table.ShapeElement(0.15, 132, shx, 0.25, 0, 0, 0), &table.LineTypeElement{Length: 0.15})
l.SetLineType(gas)
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

### Other entities

```go
d.XLine([]float64{0, 0, 0}, []float64{1, 1, 0}) // also Ray
d.Trace([][]float64{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}, {1, 1, 0}})
d.AddShapeFile("ltypeshp.shx")
d.Shape("TRACK1", []float64{5, 5, 0}, 2)
d.Tolerance(`{\Fgdt;j}%%v0.1`, []float64{0, 10, 0}) // feature control frame
d.Helix([]float64{0, 0, 0}, 2, 1, 3, true)            // radius, pitch, turns, ccw
d.Mesh(vertices, [][]int{{0, 1, 2}, {0, 2, 3}})
d.MPolygon(outer, hole) // polygon with holes, solid fill
d.Light("L1", entity.LIGHT_SPOT, []float64{0, 0, 10})
solid, _ := d.Body("3DSOLID", satLines) // ACIS data as SAT text
fmt.Println(solid.SAT)

// multilines: MLINESTYLE "Standard" (added when needed) or styles of your own
d.AddMLineStyle("WALL", entity.MLineStyleElement{Offset: 0.1, Color: 1, LineType: "BYLAYER"},
	entity.MLineStyleElement{Offset: -0.1, Color: 5, LineType: "BYLAYER"})
d.MLine("WALL", [][]float64{{0, 0}, {4, 0}, {4, 4}}, false, 1, entity.MLINE_ZERO) // closed, scale, justification
// multileader with MTEXT content at (10, 10); leaders from their arrow heads
// (the text size is estimated from the number of characters)
d.MultiLeaderText(`Note\Psecond line`, []float64{10, 10, 0}, 2.5, [][]float64{{0, 0, 0}, {5, 8, 0}})
// tables: cell texts and conversion to a block reference
content := table.Content()  // [][]string
d.TableToBlockReference(table)
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

// PDF underlay: the definition (OBJECTS, ACAD_PDFDEFINITIONS) and its reference
def, _ := d.AddUnderlayDefinition(entity.UNDERLAY_PDF, "plans/site.pdf", "2") // page 2
u, _ := d.Underlay(def, 0, 0, 0, 1) // insertion point, scale
u.Rotation = 30
u.SetClipping([]float64{0, 0}, []float64{100, 50}) // in underlay coordinates
u.Fade = 25                                        // also Contrast, Flags (UNDERLAY_MONOCHROME, ...)

// raster image: the definition (OBJECTS, ACAD_IMAGE_DICT) and its reference
img, _ := d.AddImageDefinition("images/plan.png", 640, 480) // size in pixels
im, _ := d.Image(img, 0, 0, 0, 64, 48, 0)                    // corner, width, height, rotation
im.SetBoundary([]float64{-0.5, -0.5}, []float64{319.5, 239.5}) // clipping in pixels
d.Wipeout([]float64{0, 0}, []float64{10, 0}, []float64{10, 5}) // masking polygon
```

### More creation

```go
// dimensions: also 2-line angular, ordinate, arc length and continued ones
d.Angular2LineDimension(l1a, l1b, l2a, l2b, location)
d.OrdinateDimension(origin, feature, leaderEnd, true) // X ordinate
d.ArcDimension(center, p1, p2, location)
d.LinearDimensions(points, location, 0)

// MTEXT content with formatting codes, and columns (AC1032)
ed := entity.NewMTextEditor("Title").Paragraph().Font("Arial", true, false, "bold").
	Height(2.5, "big").Stack("1", "2", entity.STACK_HORIZONTAL).ColumnBreak()
mt, _ := d.MText(ed.String(), 0, 0, 0, 1, 20)
mt.Columns = entity.NewStaticColumns(2, 20, 2, 30) // count, width, gutter, height
text.SetPlacement([]float64{5, 5}, nil, "MIDDLE_CENTER")

// hatch patterns of .pat files, associative hatches
pats, _ := entity.ParsePatterns(patFileContent)
hatch.SetPatternDefinition("MYPAT", pats["MYPAT"], 0, 1)
d.HatchFromEntities(closedPolyline, circle) // boundaries refer to the hatch

d.PolygonMesh(m, n, vertices, false, false)
d.SplineThrough(points) // control points computed; d.SplineFit(points): fit points only
d.InsertArrow("_DOT", tip, 2.5, 45) // arrowhead blocks (d.ArrowBlock)
d.LoadLineTypes(linFileContent)     // .lin definitions, also with text and shapes
dxf.SetHeaderVariable(d, "$MIRRTEXT", [2]string{"70", "1"}) // dxf.HeaderVariable
line.SetXDataValues("MYAPP", entity.XString("text"), entity.XReal(1.5)) // XDataValues
d.RevisionCloud(vertices, 2, true) // segment length, calligraphy
```

### Analysis

```go
ps, _ := path.FromEntity(e)          // lines and cubic Bezier curves (WCS)
pts := ps[0].Flatten(0.01)           // also Transform, ToLwPolyline, ToSpline, ToHatch
ves, _ := dxf.VirtualEntities(d, e)  // block references, dimensions, leaders, polylines, ...
lo, hi := dxf.Extents(d, es, 0.001)  // precise extents (arcs, splines, blocks)
dxf.SelectWindow(d, es, min, max, false) // also SelectPoint, SelectPolygon
dxf.Loops(es, 1e-6)                  // closed loops of connected edges
dxf.GroupBy(es, func(e entity.Entity) string { return e.Layer().Name() })
dxf.Query(es, `LINE CIRCLE[layer=="WALL" & color!=1]`)
dxf.BlockReferenceCounts(d)          // also dxf.Purge(d): unused blocks and layers
```

The `geometry` package has bulge/arc conversions, intersections, polygon
functions, convex hulls, Bezier curves and B-splines.

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

// page setup: A3 landscape in mm, margins (top, right, bottom, left), rotation, 1:50
d.Layouts()[1].PageSetup(420, 297, object.PLOT_MILLIMETERS, [4]float64{10, 10, 10, 20}, 0, 50, "DWG To PDF.pc3")
```

### Objects

```go
// XRECORD in the extension dictionary of an entity or a table entry
d.SetXRecord(line, "MYAPP", [][2]string{{"1", "data"}, {"40", "1.5"}})
x, _ := d.XRecord(line, "MYAPP") // x.Tags()
d.SetVariable("XCLIPFRAME", "2")  // DICTIONARYVAR of AcDbVariableDictionary
d.SetDrawOrder("", line)          // draw line on top (SORTENTSTABLE); d.DrawOrder("")
d.ClipBlockReference(insert, [][]float64{{0, 0}, {10, 5}}) // XCLIP (SPATIAL_FILTER)
d.SetRasterVariables(1, 1, 1)     // image frame, quality, units (ACAD_IMAGE_VARS)
d.SetWipeoutVariables(2)          // wipeout frame (ACAD_WIPEOUT_VARS)
for _, o := range d.Objects() {   // typed objects: *object.XRecord, *object.GeoData, ...
	_ = o
}
```

Objects read keep their group codes; the typed objects (XRECORD,
DICTIONARY, DICTIONARYVAR, SORTENTSTABLE, SPATIAL_FILTER, IDBUFFER,
LAYER_FILTER, FIELDLIST, FIELD, GEODATA, SUN, MATERIAL, VISUALSTYLE,
PLOTSETTINGS, RASTERVARIABLES, WIPEOUTVARIABLES) read and change them.

### Editing

```go
walls := d.EntitiesOnLayer("WALL") // or d.Query(func(e entity.Entity) bool { ... })
d.Move(10, 0, 0, walls...)         // also Rotate(base, angle, ...), Scale(base, factor, ...)
d.Mirror([]float64{0, 0}, []float64{0, 1}, walls...) // about the line through two points
d.ScaleXYZ([]float64{0, 0, 0}, 2, 1, 1, walls...)    // non-uniform; circles become ellipses
copies, _ := dxf.CopyEntities(d, entity.Rotate([]float64{0, 0, 0}, 90), walls...)
d.RemoveEntity(walls[0])
exploded, _ := dxf.Explode(d, insert) // block reference -> entities
d.RenameLayer("WALL", "A-WALL")       // also RemoveLayer, RenameBlock, RemoveBlock
```

Transformations (`entity.Transform`) are affine in the XY plane: moves,
rotations about the Z axis, scaling (also non-uniform, and along Z) and
mirrors (negative factors, `entity.MirrorBy`).

- Mirrored entities keep the Z axis as extrusion direction: arcs, bulges
  and hatch edges change their direction, block references get a negative
  Y scale. Text stays readable (its alignment is mirrored, like AutoCAD
  with MIRRTEXT = 0) unless `Transform.MirrorText` is set; `Explode`
  mirrors text as in the block reference.
- Under non-uniform scaling circles and arcs become ellipses
  (`entity.TransformEntity`, `Drawing.TransformEntities`); ellipses,
  splines, lines and hatch patterns are transformed exactly. Entities
  that cannot keep their shape give `entity.ErrNonUniform`: polylines
  with arcs or widths, hatches with arcs, rotated text and block
  references whose axes would not stay orthogonal, dimensions and leaders.
- Entities with an extrusion direction other than +Z are transformed in
  their object coordinate system (OCS); for directions other than +Z and
  -Z the scaling must be uniform. `entity.Upright` converts entities with
  extrusion -Z (made by mirroring in CAD applications) to +Z.

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
(`entity.Unknown`, e.g. MLINE, MULTILEADER), extended data (`XData()`),
extension dictionaries and reactors (102 groups), other values of table
entries (`Extra`), the CLASSES section and the objects of the OBJECTS
section (`RawObjects()`). Groups (`Groups`) and layouts (`Layouts()`) are
read from the OBJECTS section.

Files of versions older than AC1015 (e.g. R12) are written as AC1015
(see `WriteR12` below for R12 output). Comments (999 groups) of ASCII
files are read into `d.Comments` and written at the start of the file.

### R12

```go
err := dxf.SaveAsR12(d, "r12.dxf") // or dxf.WriteR12(d, w)
```

Entities R12 does not have are converted: LWPOLYLINE, ELLIPSE, SPLINE,
HELIX, MLINE and LEADER to POLYLINE, MTEXT to TEXT lines, HATCH and
MPOLYGON to their boundaries, MESH to a polyface mesh, MULTILEADER,
ACAD_TABLE and arc/large radial dimensions to their graphics. Lineweights,
true colors, objects and layouts other than model and paper space are not
written.

### DWG

DWG files are read and written with the
[ODA File Converter](https://www.opendesign.com/guestfiles/oda_file_converter)
(free, installed separately; found in PATH or the default install folder,
or set `dxf.ODAFileConverter`):

```go
d, err := dxf.ReadDWG("in.dwg")
err = dxf.SaveAsDWG(d, "out.dwg")                  // version of the drawing
err = dxf.ConvertDWG("in.dwg", "out.dxf", "R2000") // DWG/DXF to DWG/DXF
```

### Large and damaged files

```go
// stream the entities without keeping them
err := dxf.IterEntities(r, func(d *drawing.Drawing, e entity.Entity) error { ...; return nil })
// read what can be read; problems are returned
d, problems, err := dxf.Recover(r)
// check (and repair) a drawing: undefined blocks, invalid values, ...
for _, p := range dxf.Audit(d, true) { fmt.Println(p) }
```

### Import and external references

```go
copies, err := dxf.ImportEntities(dst, src, src.Entities()...) // with layers, styles, blocks, ...
ins, err := dxf.AttachXref(d, "PLAN", "plan.dxf", []float64{0, 0, 0}, false)
err = dxf.BindXref(d, "PLAN", dxf.FromFile) // the xref becomes a normal block
```

### Proxy graphics

Entities of applications that are not available (ACAD_PROXY_ENTITY and
custom entities, read as `entity.Unknown`) are kept as read; their proxy
graphics can be decoded into entities:

```go
es, err := dxf.ProxyGraphics(d, e) // also via dxf.VirtualEntities
```

Package `proxygraphic` decodes the commands for circles, arcs, polylines,
polygons, meshes, shells, text, xlines and rays with their colors, layers,
linetypes, lineweights and transformations.

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

## Package layout

| Path | Contents |
|---|---|
| `dxf.go` | `NewDrawing`, `FromFile`, `FromReader`, ... |
| `edit.go` | `Copy`, `CopyEntities`, `Explode`, `SetExtrusion` |
| `parse.go`, `parse_*.go` | reading: `parse_pairs.go` (ASCII/binary group codes), `parse_header.go`, `parse_tables.go`, `parse_blocks.go`, `parse_entities.go` (dispatch and common codes), one file per entity group (`parse_basic.go`, `parse_polyline.go`, `parse_text.go`, `parse_hatch.go`, `parse_dimension.go`, ...), `parse_objects.go` and `parse_preserve.go` (data kept as read) |
| `drawing/` | `Drawing`: sections, creation and editing API |
| `header/`, `class/`, `table/`, `block/`, `entity/`, `object/` | the sections and their records |
| `format/` | ASCII and binary formatters, DXF versions |
| `color/`, `codepage/`, `handle/`, `insunit/`, `geometry/` | color numbers and true colors, code pages, handles, units, geometry |
| `path/` | paths of entities (lines and Bezier curves) |
| `proxygraphic/` | decoding of proxy graphics |
| `r12.go`, `odafc.go`, `fileio.go`, `proxy.go` | R12 output, DWG via the ODA File Converter, audit/import/xref, proxy graphics |
| `testdata/` | expected output of the tests |
| `example/` | examples |
| `tools/ezdxfcheck/` | validation with ezdxf |

Tests are in `*_test.go` next to the code; helpers shared by them are in
`helpers_test.go`.

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
