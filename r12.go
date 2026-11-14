package dxf

import (
	"io"
	"math"
	"os"
	"strings"

	"github.com/yofu/dxf/codepage"
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/path"
	"github.com/yofu/dxf/table"
)

// SaveAsR12 saves the drawing as DXF R12 (AC1009), see WriteR12.
func SaveAsR12(d *drawing.Drawing, filename string) error {
	w, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer w.Close()
	return WriteR12(d, w)
}

// WriteR12 writes the drawing as DXF R12 (AC1009, as ezdxf's r12export):
// header, tables (linetypes, layers, styles, application ids, dimension
// style), blocks and entities, without handles and objects. Entities of
// R12 are written as they are; the others are converted: LWPOLYLINE to
// POLYLINE, ELLIPSE, SPLINE and HELIX to flattened POLYLINE, MTEXT to TEXT
// lines, HATCH, MPOLYGON, IMAGE, WIPEOUT and underlays to their boundary
// polylines, MESH to a polyface mesh, LEADER to a polyline, MLINE,
// MULTILEADER and ACAD_TABLE to their virtual entities. Entities R12 cannot
// show (XLINE, RAY, LIGHT, ACIS, ...) are not written.
func WriteR12(d *drawing.Drawing, w io.Writer) error {
	f := format.NewASCII()
	f.SetPrecision(16)
	h := d.Header()
	r := &r12{d: d, f: f, apps: map[string]bool{}}
	// header
	f.WriteString(0, "SECTION")
	f.WriteString(2, "HEADER")
	f.WriteString(9, "$ACADVER")
	f.WriteString(1, "AC1009")
	f.WriteString(9, "$DWGCODEPAGE")
	f.WriteString(3, h.CodePage)
	f.WriteString(9, "$INSBASE")
	writeP(f, 10, h.InsBase)
	f.WriteString(9, "$EXTMIN")
	writeP(f, 10, h.ExtMin)
	f.WriteString(9, "$EXTMAX")
	writeP(f, 10, h.ExtMax)
	f.WriteString(9, "$LIMMIN")
	f.WriteFloat(10, coordOf(h.LimMin, 0))
	f.WriteFloat(20, coordOf(h.LimMin, 1))
	f.WriteString(9, "$LIMMAX")
	f.WriteFloat(10, coordOf(h.LimMax, 0))
	f.WriteFloat(20, coordOf(h.LimMax, 1))
	f.WriteString(9, "$HANDLING")
	f.WriteInt(70, 0)
	f.WriteString(9, "$CLAYER")
	f.WriteString(8, d.CurrentLayer.Name())
	f.WriteString(0, "ENDSEC")
	// entities first (to know the application ids of extended data)
	ent := format.NewASCII()
	ent.SetPrecision(16)
	r.f = ent
	for _, e := range d.Entities() {
		r.entity(e, false)
	}
	blk := format.NewASCII()
	blk.SetPrecision(16)
	r.f = blk
	for _, b := range d.Blocks() {
		name := b.Name
		up := strings.ToUpper(name)
		if strings.HasPrefix(up, "*MODEL_SPACE") || strings.HasPrefix(up, "*PAPER_SPACE") {
			if up != "*MODEL_SPACE" && up != "*PAPER_SPACE" {
				continue // other layouts do not exist in R12
			}
		}
		r12name := strings.ToUpper(name)
		switch r12name {
		case "*MODEL_SPACE":
			r12name = "$MODEL_SPACE"
		case "*PAPER_SPACE":
			r12name = "$PAPER_SPACE"
		}
		blk.WriteString(0, "BLOCK")
		blk.WriteString(8, "0")
		blk.WriteString(2, r12name)
		blk.WriteInt(70, b.Flag)
		writeP(blk, 10, b.Coord)
		blk.WriteString(3, r12name)
		if b.XrefPath != "" {
			blk.WriteString(1, b.XrefPath)
		}
		for _, e := range b.Entities {
			r.entity(e, false)
		}
		blk.WriteString(0, "ENDBLK")
		blk.WriteString(8, "0")
	}
	r.f = f
	r.tables()
	f.WriteString(0, "SECTION")
	f.WriteString(2, "BLOCKS")
	out := f.Output() + blk.Output()
	f.Reset()
	f.WriteString(0, "ENDSEC")
	f.WriteString(0, "SECTION")
	f.WriteString(2, "ENTITIES")
	out += f.Output() + ent.Output()
	f.Reset()
	f.WriteString(0, "ENDSEC")
	f.WriteString(0, "EOF")
	out += f.Output()
	_, err := io.WriteString(w, codepage.Encode(out, h.CodePage))
	return err
}

func writeP(f format.Formatter, code int, p []float64) {
	for i := 0; i < 3; i++ {
		f.WriteFloat(code+10*i, coordOf(p, i))
	}
}

type r12 struct {
	d    *drawing.Drawing
	f    format.Formatter
	apps map[string]bool
}

func (r *r12) tables() {
	f := r.f
	tables := r.d.Sections[drawing.TABLES].(table.Tables)
	begin := func(name string, n int) {
		f.WriteString(0, "TABLE")
		f.WriteString(2, name)
		f.WriteInt(70, n)
	}
	f.WriteString(0, "SECTION")
	f.WriteString(2, "TABLES")
	lts := tables[table.LTYPE].Entries()
	begin("LTYPE", len(lts))
	for _, st := range lts {
		lt := st.(*table.LineType)
		f.WriteString(0, "LTYPE")
		f.WriteString(2, strings.ToUpper(lt.Name()))
		f.WriteInt(70, 0)
		f.WriteString(3, lt.Description)
		f.WriteInt(72, 65)
		f.WriteInt(73, len(lt.Elements()))
		f.WriteFloat(40, lt.TotalLength())
		for _, e := range lt.Elements() {
			f.WriteFloat(49, e.Length) // complex elements do not exist in R12
		}
	}
	f.WriteString(0, "ENDTAB")
	layers := tables[table.LAYER].Entries()
	begin("LAYER", len(layers))
	for _, st := range layers {
		l := st.(*table.Layer)
		f.WriteString(0, "LAYER")
		f.WriteString(2, l.Name())
		f.WriteInt(70, 0)
		c := int(l.Color)
		if !l.IsOn() {
			c = -c
		}
		f.WriteInt(62, c)
		lt := "CONTINUOUS"
		if l.LineType != nil {
			lt = strings.ToUpper(l.LineType.Name())
		}
		f.WriteString(6, lt)
	}
	f.WriteString(0, "ENDTAB")
	styles := tables[table.STYLE].Entries()
	begin("STYLE", len(styles))
	for _, st := range styles {
		s := st.(*table.Style)
		f.WriteString(0, "STYLE")
		f.WriteString(2, strings.ToUpper(s.Name()))
		f.WriteInt(70, s.Flags)
		f.WriteFloat(40, s.FixedTextHeight)
		f.WriteFloat(41, s.WidthFactor)
		f.WriteFloat(50, s.ObliqueAngle)
		f.WriteInt(71, s.GenFlag)
		f.WriteFloat(42, s.LastHeightUsed)
		f.WriteString(3, s.FontName)
		f.WriteString(4, s.BigFontName)
	}
	f.WriteString(0, "ENDTAB")
	apps := []string{"ACAD"}
	for a := range r.apps {
		if !strings.EqualFold(a, "ACAD") {
			apps = append(apps, a)
		}
	}
	begin("APPID", len(apps))
	for _, a := range apps {
		f.WriteString(0, "APPID")
		f.WriteString(2, strings.ToUpper(a))
		f.WriteInt(70, 0)
	}
	f.WriteString(0, "ENDTAB")
	begin("DIMSTYLE", 1)
	f.WriteString(0, "DIMSTYLE")
	f.WriteString(2, "STANDARD")
	f.WriteInt(70, 0)
	f.WriteString(0, "ENDTAB")
	f.WriteString(0, "ENDSEC")
}

// common writes the type and the common group codes of e.
func (r *r12) common(typ string, e entity.Entity) {
	f := r.f
	f.WriteString(0, typ)
	if sp, ok := e.(entity.SpaceSelector); ok && sp.InPaperSpace() {
		f.WriteInt(67, 1)
	}
	f.WriteString(8, e.Layer().Name())
	if lp, ok := e.(interface{ LineType() *table.LineType }); ok && lp.LineType() != nil && !strings.EqualFold(lp.LineType().Name(), "BYLAYER") {
		f.WriteString(6, strings.ToUpper(lp.LineType().Name()))
	}
	if c, ok := e.(interface{ ColorIndex() int }); ok && c.ColorIndex() != 256 {
		f.WriteInt(62, c.ColorIndex())
	}
}

func (r *r12) extrusion(n []float64) {
	if len(n) >= 3 && !(n[0] == 0 && n[1] == 0 && n[2] == 1) {
		writeP(r.f, 210, n)
	}
}

func (r *r12) xdata(e entity.Entity) {
	if x, ok := e.(interface{ XData() [][2]string }); ok {
		for _, p := range x.XData() {
			if p[0] == "1001" {
				r.apps[strings.TrimSpace(p[1])] = true
			}
			writeRawPairs(r.f, p)
		}
	}
}

func writeRawPairs(f format.Formatter, p [2]string) {
	if c, err := parseCode(p[0]); err == nil {
		f.WriteString(c, p[1])
	}
}

func parseCode(s string) (int, error) {
	n := 0
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, io.EOF
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, io.EOF
		}
		n = n*10 + int(ch-'0')
	}
	return n, nil
}

// entity writes e in R12 form (converted if needed).
func (r *r12) entity(e entity.Entity, nested bool) {
	f := r.f
	switch x := e.(type) {
	case *entity.Line:
		r.common("LINE", e)
		if x.Thickness != 0 {
			f.WriteFloat(39, x.Thickness)
		}
		writeP(f, 10, x.Start)
		writeP(f, 11, x.End)
		r.extrusion(x.Extrusion)
	case *entity.Point:
		r.common("POINT", e)
		writeP(f, 10, x.Coord)
		if x.Thickness != 0 {
			f.WriteFloat(39, x.Thickness)
		}
		r.extrusion(x.Extrusion)
	case *entity.Arc:
		r.common("ARC", e)
		if x.Thickness != 0 {
			f.WriteFloat(39, x.Thickness)
		}
		writeP(f, 10, x.Center)
		f.WriteFloat(40, x.Radius)
		f.WriteFloat(50, x.Angle[0])
		f.WriteFloat(51, x.Angle[1])
		r.extrusion(x.Direction)
	case *entity.Circle:
		r.common("CIRCLE", e)
		if x.Thickness != 0 {
			f.WriteFloat(39, x.Thickness)
		}
		writeP(f, 10, x.Center)
		f.WriteFloat(40, x.Radius)
		r.extrusion(x.Direction)
	case *entity.Solid:
		typ := "SOLID"
		if x.Type == entity.TRACE {
			typ = "TRACE"
		}
		r.common(typ, e)
		for i, p := range x.Points {
			writeP(f, 10+i, p)
		}
		r.extrusion(x.Extrusion)
	case *entity.ThreeDFace:
		r.common("3DFACE", e)
		for i, p := range x.Points {
			writeP(f, 10+i, p)
		}
		f.WriteInt(70, x.Flag)
	case *entity.AttributeDefinition:
		r.common("ATTDEF", e)
		r.text(x.Text)
		f.WriteString(3, x.Prompt)
		f.WriteString(2, x.Tag)
		f.WriteInt(70, x.Flags)
		f.WriteInt(74, x.VerticalFlag)
	case *entity.Text:
		r.common("TEXT", e)
		r.text(x)
		if x.VerticalFlag != 0 {
			f.WriteInt(73, x.VerticalFlag)
		}
	case *entity.Shape:
		r.common("SHAPE", e)
		writeP(f, 10, x.Insert)
		f.WriteFloat(40, x.Size)
		f.WriteString(2, x.Name)
		f.WriteFloat(50, x.Rotation)
		f.WriteFloat(41, x.XScale)
		f.WriteFloat(51, x.Oblique)
	case *entity.Insert:
		r.common("INSERT", e)
		if len(x.Attributes) > 0 {
			f.WriteInt(66, 1)
		}
		f.WriteString(2, strings.ToUpper(x.BlockName))
		writeP(f, 10, x.Coord)
		f.WriteFloat(41, coordOf(x.Scale, 0))
		f.WriteFloat(42, coordOf(x.Scale, 1))
		f.WriteFloat(43, coordOf(x.Scale, 2))
		f.WriteFloat(50, x.Rotation)
		if x.Columns > 1 || x.Rows > 1 {
			f.WriteInt(70, x.Columns)
			f.WriteInt(71, x.Rows)
			f.WriteFloat(44, x.ColumnSpacing)
			f.WriteFloat(45, x.RowSpacing)
		}
		r.extrusion(x.Direction)
		r.xdata(e)
		if len(x.Attributes) > 0 {
			for _, a := range x.Attributes {
				r.common("ATTRIB", a)
				r.text(a.Text)
				f.WriteString(2, a.Tag)
				f.WriteInt(70, a.Flags)
				f.WriteInt(73, a.FieldLength)
				if a.VerticalFlag != 0 {
					f.WriteInt(74, a.VerticalFlag)
				}
			}
			f.WriteString(0, "SEQEND")
			f.WriteString(8, x.Layer().Name())
		}
		return
	case *entity.Polyline:
		r.common("POLYLINE", e)
		f.WriteInt(66, 1)
		f.WriteFloat(10, 0)
		f.WriteFloat(20, 0)
		f.WriteFloat(30, x.Elevation)
		if x.Thickness != 0 {
			f.WriteFloat(39, x.Thickness)
		}
		f.WriteInt(70, x.Flag)
		if x.StartWidth != 0 || x.EndWidth != 0 {
			f.WriteFloat(40, x.StartWidth)
			f.WriteFloat(41, x.EndWidth)
		}
		if x.MeshM != 0 || x.MeshN != 0 {
			f.WriteInt(71, x.MeshM)
			f.WriteInt(72, x.MeshN)
		}
		r.extrusion(x.Extrusion)
		r.xdata(e)
		for _, v := range x.Vertices {
			r.common("VERTEX", e)
			writeP(f, 10, v.Coord)
			if v.StartWidth != 0 || v.EndWidth != 0 {
				f.WriteFloat(40, v.StartWidth)
				f.WriteFloat(41, v.EndWidth)
			}
			if v.Bulge != 0 {
				f.WriteFloat(42, v.Bulge)
			}
			f.WriteInt(70, v.Flag)
			for i, idx := range v.Indices {
				f.WriteInt(71+i, idx)
			}
		}
		f.WriteString(0, "SEQEND")
		f.WriteString(8, x.Layer().Name())
		return
	case *entity.Dimension:
		if x.Type == entity.DIM_ARC || x.Type == entity.DIM_LARGE_RADIAL {
			// not in R12: the graphics block is inserted
			ins := entity.NewInsert(x.BlockName)
			ins.SetLayer(x.Layer())
			r.entity(ins, nested)
			return
		}
		r.common("DIMENSION", e)
		f.WriteString(2, strings.ToUpper(x.BlockName))
		writeP(f, 10, x.DefPoint)
		writeP(f, 11, x.TextMid)
		f.WriteInt(70, x.Type|x.Flags&^entity.DIM_BLOCK_OWNED)
		f.WriteString(1, x.Text)
		for _, p := range []struct {
			c int
			v []float64
		}{{13, x.DefPoint2}, {14, x.DefPoint3}, {15, x.DefPoint4}, {16, x.DefPoint5}} {
			if len(p.v) >= 3 {
				writeP(f, p.c, p.v)
			}
		}
		if x.LeaderLength != 0 {
			f.WriteFloat(40, x.LeaderLength)
		}
		if x.Type == entity.DIM_LINEAR {
			f.WriteFloat(50, x.Angle)
		}
		r.extrusion(x.Extrusion)
	case *entity.LwPolyline:
		p := entity.NewPolyline()
		p.SetLayer(x.Layer())
		copyProps(p, x)
		p.Elevation = x.Elevation
		p.Thickness = x.Thickness
		p.Extrusion = x.Extrusion
		if x.Closed {
			p.Flag |= 1
		}
		for i, v := range x.Vertices {
			vx := p.AddVertex(v[0], v[1], x.Elevation)
			if i < len(x.Bulges) {
				vx.Bulge = x.Bulges[i]
			}
			if i < len(x.Widths) {
				vx.StartWidth, vx.EndWidth = x.Widths[i][0], x.Widths[i][1]
			}
			if x.ConstantWidth != 0 && vx.StartWidth == 0 && vx.EndWidth == 0 {
				vx.StartWidth, vx.EndWidth = x.ConstantWidth, x.ConstantWidth
			}
		}
		r.entity(p, nested)
		return
	case *entity.MText:
		lines := strings.Split(x.PlainText(), "\n")
		rad := x.Rotation * math.Pi / 180
		step := x.Height * 5 / 3 * x.LineSpacingFactor
		for i, l := range lines {
			if l == "" {
				continue
			}
			t := entity.NewText()
			t.SetLayer(x.Layer())
			copyProps(t, x)
			t.Value = l
			t.Height = x.Height
			t.Style = x.Style
			t.Rotation = x.Rotation
			off := float64(i)*step + x.Height // baseline of line i below the top
			t.Coord1 = []float64{x.Coord[0] + off*math.Sin(rad), x.Coord[1] - off*math.Cos(rad), coordOf(x.Coord, 2)}
			r.entity(t, nested)
		}
		return
	case *entity.Ellipse, *entity.Spline, *entity.Helix, *entity.Hatch, *entity.MPolygon, *entity.Image, *entity.Leader:
		ps, err := path.FromEntity(e)
		if err != nil {
			return
		}
		for _, p := range ps {
			pl := entity.NewPolyline()
			pl.SetLayer(e.Layer())
			copyProps(pl, e)
			pts := p.Flatten(0.01)
			closed := p.IsClosed() && len(pts) > 2
			if closed {
				pts = pts[:len(pts)-1]
				pl.Flag |= 1
			}
			planar := true
			for _, q := range pts {
				planar = planar && q[2] == pts[0][2]
			}
			if !planar {
				pl.Flag |= 8
			} else {
				pl.Elevation = pts[0][2]
			}
			for _, q := range pts {
				v := pl.AddVertex(q[0], q[1], q[2])
				if !planar {
					v.Flag = 32
				}
			}
			r.entity(pl, nested)
		}
		if l, ok := e.(*entity.Leader); ok && l.Arrow && len(l.Vertices) > 1 {
			r.entity(arrowSolid(l.Vertices[0], l.Vertices[1], 2.5, l), nested)
		}
		return
	case *entity.Mesh:
		pl := entity.NewPolyline()
		pl.SetLayer(x.Layer())
		copyProps(pl, x)
		pl.Flag = 64
		pl.MeshM, pl.MeshN = len(x.Vertices), len(x.Faces)
		for _, v := range x.Vertices {
			vx := pl.AddVertex(v[0], v[1], coordOf(v, 2))
			vx.Flag = 128 | 64
		}
		addFace := func(ids ...int) {
			vx := pl.AddVertex(0, 0, 0)
			vx.Flag = 128
			for _, i := range ids {
				vx.Indices = append(vx.Indices, i+1)
			}
		}
		for _, face := range x.Faces {
			switch {
			case len(face) < 3:
			case len(face) <= 4:
				addFace(face...)
			default: // triangle fan
				for k := 1; k+1 < len(face); k++ {
					addFace(face[0], face[k], face[k+1])
				}
			}
		}
		r.entity(pl, nested)
		return
	case *entity.MLine, *entity.MultiLeader, *entity.AcadTable:
		if ves, err := VirtualEntities(r.d, e); err == nil {
			for _, v := range ves {
				r.entity(v, nested)
			}
		}
		return
	default:
		return // not in R12
	}
	r.xdata(e)
}

func copyProps(dst, src entity.Entity) {
	if c, ok := src.(interface{ ColorIndex() int }); ok {
		if s, ok := dst.(interface{ SetColor(int) }); ok {
			s.SetColor(c.ColorIndex())
		}
	}
	if lp, ok := src.(interface{ LineType() *table.LineType }); ok && lp.LineType() != nil {
		if s, ok := dst.(interface{ SetLineType(*table.LineType) }); ok {
			s.SetLineType(lp.LineType())
		}
	}
	if sp, ok := src.(entity.SpaceSelector); ok {
		if dp, ok := dst.(entity.SpaceSelector); ok {
			dp.SetPaperSpace(sp.InPaperSpace())
		}
	}
}

// text writes the group codes of TEXT, ATTRIB and ATTDEF.
func (r *r12) text(t *entity.Text) {
	f := r.f
	if t.Thickness != 0 {
		f.WriteFloat(39, t.Thickness)
	}
	writeP(f, 10, t.Coord1)
	f.WriteFloat(40, t.Height)
	f.WriteString(1, t.Value)
	if t.Rotation != 0 {
		f.WriteFloat(50, t.Rotation)
	}
	if t.WidthFactor != 0 && t.WidthFactor != 1 {
		f.WriteFloat(41, t.WidthFactor)
	}
	if t.ObliqueAngle != 0 {
		f.WriteFloat(51, t.ObliqueAngle)
	}
	if t.Style != nil {
		f.WriteString(7, strings.ToUpper(t.Style.Name()))
	}
	if t.GenFlag != 0 {
		f.WriteInt(71, t.GenFlag)
	}
	if t.HorizontalFlag != 0 {
		f.WriteInt(72, t.HorizontalFlag)
	}
	if (t.HorizontalFlag != 0 || t.VerticalFlag != 0) && len(t.Coord2) >= 2 {
		writeP(f, 11, t.Coord2)
	}
	r.extrusion(t.Extrusion)
}
