package dxf

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/geometry"
	"github.com/yofu/dxf/path"
)

// ErrNoVirtualEntities is returned by VirtualEntities for entities that
// are not made of other entities.
var ErrNoVirtualEntities = errors.New("entity has no virtual entities")

// VirtualEntities returns the entities an entity is made of, without
// changing the drawing (as ezdxf's virtual_entities): block references
// (as Explode), dimensions and tables (their graphics blocks), leaders,
// multileaders, multilines, polylines (lines and arcs), meshes and polyface
// meshes (3DFACEs).
func VirtualEntities(d *drawing.Drawing, e entity.Entity) (entity.Entities, error) {
	switch x := e.(type) {
	case *entity.Unknown:
		return ProxyGraphics(d, x)
	case *entity.Insert:
		return insertEntities(d, x)
	case *entity.Dimension:
		return blockCopies(d, x.BlockName, nil)
	case *entity.AcadTable:
		return blockCopies(d, x.BlockName, x.Insert)
	case *entity.Leader:
		var out entity.Entities
		for i := 0; i+1 < len(x.Vertices); i++ {
			out = append(out, lineOf(x.Vertices[i], x.Vertices[i+1], x))
		}
		if x.Arrow && len(x.Vertices) > 1 {
			out = append(out, arrowSolid(x.Vertices[0], x.Vertices[1], 2.5, x))
		}
		return out, nil
	case *entity.MultiLeader:
		return mleaderEntities(d, x)
	case *entity.MLine, *entity.LwPolyline, *entity.Polyline, *entity.Mesh:
		return pathEntities(e)
	}
	return nil, fmt.Errorf("%T: %w", e, ErrNoVirtualEntities)
}

// blockCopies returns copies of the entities of block moved by offset.
func blockCopies(d *drawing.Drawing, block string, offset []float64) (entity.Entities, error) {
	b, err := d.Block(block)
	if err != nil {
		return nil, err
	}
	var out entity.Entities
	for _, be := range b.Entities {
		c, err := Copy(d, be)
		if err != nil {
			return nil, err
		}
		if len(offset) >= 2 {
			t := entity.Move(offset[0]-b.Coord[0], offset[1]-b.Coord[1], coordOf(offset, 2)-coordOf(b.Coord, 2))
			if c, err = entity.TransformEntity(c, t); err != nil {
				return nil, err
			}
		}
		out = append(out, c)
	}
	return out, nil
}

func coordOf(v []float64, i int) float64 {
	if i < len(v) {
		return v[i]
	}
	return 0
}

func props(dst, src entity.Entity) {
	dst.SetLayer(src.Layer())
	if c, ok := src.(interface{ ColorIndex() int }); ok {
		if s, ok := dst.(interface{ SetColor(int) }); ok {
			s.SetColor(c.ColorIndex())
		}
	}
}

func lineOf(p, q []float64, src entity.Entity) *entity.Line {
	l := entity.NewLine()
	l.Start = []float64{coordOf(p, 0), coordOf(p, 1), coordOf(p, 2)}
	l.End = []float64{coordOf(q, 0), coordOf(q, 1), coordOf(q, 2)}
	props(l, src)
	return l
}

// arrowSolid returns a closed filled arrowhead at tip pointing away from
// toward.
func arrowSolid(tip, toward []float64, size float64, src entity.Entity) *entity.Solid {
	dx, dy := coordOf(toward, 0)-coordOf(tip, 0), coordOf(toward, 1)-coordOf(tip, 1)
	l := math.Hypot(dx, dy)
	if l == 0 {
		l = 1
	}
	dx, dy = dx/l, dy/l
	z := coordOf(tip, 2)
	bx, by := coordOf(tip, 0)+dx*size, coordOf(tip, 1)+dy*size
	w := size / 6
	s := entity.NewSolid()
	s.Points = [][]float64{{tip[0], tip[1], z}, {bx - dy*w, by + dx*w, z}, {bx + dy*w, by - dx*w, z}, {bx + dy*w, by - dx*w, z}}
	props(s, src)
	return s
}

func mleaderEntities(d *drawing.Drawing, m *entity.MultiLeader) (entity.Entities, error) {
	var out entity.Entities
	c := m.Context
	if c == nil {
		return nil, nil
	}
	for _, l := range c.Leaders {
		for _, ln := range l.Lines {
			pts := append(append([][]float64{}, ln.Vertices...), l.LastPoint)
			for i := 0; i+1 < len(pts); i++ {
				out = append(out, lineOf(pts[i], pts[i+1], m))
			}
			if len(pts) > 1 {
				out = append(out, arrowSolid(pts[0], pts[1], c.ArrowSize, m))
			}
		}
		if m.HasDogleg && len(l.Dogleg) >= 2 {
			end := []float64{l.LastPoint[0] + l.Dogleg[0]*l.DoglegLength, l.LastPoint[1] + l.Dogleg[1]*l.DoglegLength, coordOf(l.LastPoint, 2)}
			out = append(out, lineOf(l.LastPoint, end, m))
		}
	}
	if x := c.MText; x != nil {
		t := entity.NewMText()
		t.Coord = append([]float64{}, x.Insert...)
		t.Value = x.Content
		t.Height = c.CharHeight
		t.Width = x.Width
		t.Rotation = x.Rotation * 180 / math.Pi
		if x.Alignment >= 1 && x.Alignment <= 3 {
			t.Attachment = x.Alignment
		}
		if st, ok := x.Style.Object.(interface{ Name() string }); ok {
			if s, err := d.Style(st.Name(), false); err == nil {
				t.Style = s
			}
		}
		props(t, m)
		out = append(out, t)
	}
	return out, nil
}

// pathEntities returns the lines and arcs of the paths of e (bulges as
// arcs in the plane of e).
func pathEntities(e entity.Entity) (entity.Entities, error) {
	var out entity.Entities
	addBulges := func(vs [][]float64, bs []float64, closed bool, z float64, ext []float64) {
		n := len(vs)
		segs := n - 1
		if closed {
			segs = n
		}
		for i := 0; i < segs; i++ {
			s, en := vs[i], vs[(i+1)%n]
			b := 0.0
			if i < len(bs) {
				b = bs[i]
			}
			center, r, a1, a2, ok := geometry.BulgeToArc(s, en, b)
			if !ok {
				out = append(out, lineOf([]float64{s[0], s[1], z}, []float64{en[0], en[1], z}, e))
				continue
			}
			a := entity.NewArc(nil)
			a.Center = []float64{center[0], center[1], z}
			a.Radius = r
			a.Angle = []float64{a1 * 180 / math.Pi, a2 * 180 / math.Pi}
			if len(ext) >= 3 {
				a.Direction = append([]float64{}, ext...)
			}
			props(a, e)
			out = append(out, a)
		}
	}
	switch x := e.(type) {
	case *entity.LwPolyline:
		if ext := x.Extrusion; len(ext) >= 3 && !(ext[0] == 0 && ext[1] == 0 && ext[2] > 0) {
			return wcsLines(e)
		}
		addBulges(x.Vertices, x.Bulges, x.Closed, x.Elevation, x.Extrusion)
		return out, nil
	case *entity.Polyline:
		switch {
		case x.Flag&64 != 0:
			return polyfaceFaces(x), nil
		case x.Flag&(8|16) != 0:
			return wcsLines(e)
		}
		if ext := x.Extrusion; len(ext) >= 3 && !(ext[0] == 0 && ext[1] == 0 && ext[2] > 0) {
			return wcsLines(e)
		}
		var vs [][]float64
		var bs []float64
		for _, v := range x.Vertices {
			vs = append(vs, v.Coord)
			bs = append(bs, v.Bulge)
		}
		addBulges(vs, bs, x.Flag&1 != 0, x.Elevation, x.Extrusion)
		return out, nil
	case *entity.Mesh:
		for _, f := range x.Faces {
			for k := 1; k+1 < len(f); k += 2 {
				face := entity.New3DFace()
				ids := []int{f[0], f[k], f[k+1], f[k+1]}
				if k+2 < len(f) {
					ids[3] = f[k+2]
				}
				face.Points = nil
				for _, i := range ids {
					face.Points = append(face.Points, append([]float64{}, x.Vertices[i]...))
				}
				props(face, x)
				out = append(out, face)
			}
		}
		return out, nil
	}
	return wcsLines(e)
}

// wcsLines returns the lines of the flattened paths of e.
func wcsLines(e entity.Entity) (entity.Entities, error) {
	ps, err := path.FromEntity(e)
	if err != nil {
		return nil, err
	}
	var out entity.Entities
	for _, p := range ps {
		pts := p.Flatten(0.01)
		for i := 0; i+1 < len(pts); i++ {
			out = append(out, lineOf(pts[i], pts[i+1], e))
		}
	}
	return out, nil
}

func polyfaceFaces(x *entity.Polyline) entity.Entities {
	var verts [][]float64
	var out entity.Entities
	for _, v := range x.Vertices {
		if v.Flag&128 != 0 && v.Flag&64 == 0 {
			face := entity.New3DFace()
			face.Points = nil
			for _, i := range v.Indices {
				if i < 0 {
					i = -i
				}
				if i == 0 || i-1 >= len(verts) {
					continue
				}
				face.Points = append(face.Points, append([]float64{}, verts[i-1]...))
			}
			for len(face.Points) < 4 && len(face.Points) > 0 {
				face.Points = append(face.Points, face.Points[len(face.Points)-1])
			}
			props(face, x)
			out = append(out, face)
		} else {
			verts = append(verts, v.Coord)
		}
	}
	return out
}

// Extents returns the extents of entities (WCS): from their paths
// flattened with tolerance (arcs, ellipses and splines precisely), block
// references from their virtual entities, the others from BBox.
func Extents(d *drawing.Drawing, es entity.Entities, tolerance float64) ([]float64, []float64) {
	mins := []float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxs := []float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	add := func(lo, hi []float64) {
		for i := 0; i < 3; i++ {
			mins[i] = math.Min(mins[i], coordOf(lo, i))
			maxs[i] = math.Max(maxs[i], coordOf(hi, i))
		}
	}
	var visit func(es entity.Entities, depth int)
	visit = func(es entity.Entities, depth int) {
		for _, e := range es {
			if ins, ok := e.(*entity.Insert); ok && depth < 16 {
				if ves, err := VirtualEntities(d, ins); err == nil {
					visit(ves, depth+1)
					continue
				}
			}
			if dim, ok := e.(*entity.Dimension); ok && depth < 16 {
				if ves, err := VirtualEntities(d, dim); err == nil && len(ves) > 0 {
					visit(ves, depth+1)
					continue
				}
			}
			if ps, err := path.FromEntity(e); err == nil && len(ps) > 0 {
				if _, ok := e.(*entity.XLine); !ok {
					for _, p := range ps {
						add(p.BBox(tolerance))
					}
					continue
				}
			}
			add(e.BBox())
		}
	}
	visit(es, 0)
	if math.IsInf(mins[0], 1) {
		return []float64{0, 0, 0}, []float64{0, 0, 0}
	}
	return mins, maxs
}

// SelectWindow returns the entities inside the window min-max (XY), or
// also those crossing it if crossing is true (by their extents).
func SelectWindow(d *drawing.Drawing, es entity.Entities, min, max []float64, crossing bool) entity.Entities {
	var out entity.Entities
	for _, e := range es {
		lo, hi := Extents(d, entity.Entities{e}, 0.01)
		inside := lo[0] >= min[0] && lo[1] >= min[1] && hi[0] <= max[0] && hi[1] <= max[1]
		overlap := lo[0] <= max[0] && hi[0] >= min[0] && lo[1] <= max[1] && hi[1] >= min[1]
		if inside || (crossing && overlap && crossesWindow(e, min, max)) {
			out = append(out, e)
		}
	}
	return out
}

// crossesWindow reports whether a path of e has a point in the window or
// a segment crossing its sides.
func crossesWindow(e entity.Entity, min, max []float64) bool {
	ps, err := path.FromEntity(e)
	if err != nil {
		return true // by extents
	}
	corners := [][]float64{{min[0], min[1]}, {max[0], min[1]}, {max[0], max[1]}, {min[0], max[1]}}
	for _, p := range ps {
		pts := p.Flatten(0.01)
		for i, q := range pts {
			if q[0] >= min[0] && q[0] <= max[0] && q[1] >= min[1] && q[1] <= max[1] {
				return true
			}
			if i == 0 {
				continue
			}
			for k := 0; k < 4; k++ {
				if _, ok := geometry.LineIntersection(pts[i-1], q, corners[k], corners[(k+1)%4], true); ok {
					return true
				}
			}
		}
	}
	return false
}

// SelectPoint returns the entities whose paths pass within radius of p
// (XY).
func SelectPoint(es entity.Entities, p []float64, radius float64) entity.Entities {
	var out entity.Entities
	for _, e := range es {
		ps, err := path.FromEntity(e)
		if err != nil {
			lo, hi := e.BBox()
			if p[0] >= lo[0]-radius && p[0] <= hi[0]+radius && p[1] >= lo[1]-radius && p[1] <= hi[1]+radius {
				out = append(out, e)
			}
			continue
		}
	found:
		for _, pa := range ps {
			pts := pa.Flatten(radius / 10)
			for i := 0; i+1 < len(pts); i++ {
				if geometry.DistancePointLine(p, pts[i], pts[i+1]) <= radius {
					out = append(out, e)
					break found
				}
			}
		}
	}
	return out
}

// SelectPolygon returns the entities whose paths are inside the polygon
// (XY).
func SelectPolygon(es entity.Entities, polygon [][]float64) entity.Entities {
	var out entity.Entities
	for _, e := range es {
		ps, err := path.FromEntity(e)
		if err != nil {
			continue
		}
		inside := len(ps) > 0
		for _, p := range ps {
			for _, q := range p.Flatten(0.01) {
				inside = inside && geometry.PointInPolygon(q, polygon)
			}
		}
		if inside {
			out = append(out, e)
		}
	}
	return out
}

// Loops returns the closed loops formed by entities connected end to end
// (lines, arcs, open polylines, splines, elliptic arcs) within tolerance,
// as ezdxf's edgeminer: each loop is a sequence of entities.
func Loops(es entity.Entities, tolerance float64) [][]entity.Entity {
	type edge struct {
		e          entity.Entity
		start, end []float64
	}
	var edges []edge
	for _, e := range es {
		ps, err := path.FromEntity(e)
		if err != nil || len(ps) != 1 || ps[0].IsClosed() {
			continue
		}
		edges = append(edges, edge{e, ps[0].Start, ps[0].End()})
	}
	near := func(a, b []float64) bool { return math.Hypot(a[0]-b[0], a[1]-b[1]) <= tolerance }
	used := make([]bool, len(edges))
	var loops [][]entity.Entity
	var search func(chain []int, end []float64, start []float64) []int
	search = func(chain []int, end, start []float64) []int {
		if len(chain) > 1 && near(end, start) {
			return chain
		}
		for i, ed := range edges {
			if used[i] {
				continue
			}
			var next []float64
			switch {
			case near(ed.start, end):
				next = ed.end
			case near(ed.end, end):
				next = ed.start
			default:
				continue
			}
			used[i] = true
			if r := search(append(chain, i), next, start); r != nil {
				return r
			}
			used[i] = false
		}
		return nil
	}
	for i := range edges {
		if used[i] {
			continue
		}
		used[i] = true
		if r := search([]int{i}, edges[i].end, edges[i].start); r != nil {
			var loop []entity.Entity
			for _, k := range r {
				loop = append(loop, edges[k].e)
			}
			loops = append(loops, loop)
		} else {
			used[i] = false
		}
	}
	return loops
}

// GroupBy groups entities by key (as ezdxf's groupby).
func GroupBy(es entity.Entities, key func(entity.Entity) string) map[string]entity.Entities {
	out := map[string]entity.Entities{}
	for _, e := range es {
		k := key(e)
		out[k] = append(out[k], e)
	}
	return out
}

// TypeName returns the DXF type name of an entity (e.g. "LINE").
func TypeName(e entity.Entity) string {
	if t, ok := e.(interface{ TypeName() string }); ok && t.TypeName() != "" {
		return t.TypeName()
	}
	s := e.(interface{ String() string }).String()
	lines := strings.SplitN(s, "\n", 3)
	if len(lines) >= 2 {
		return strings.TrimSpace(lines[1])
	}
	return ""
}

// Query selects entities with a query string (as ezdxf's query): entity
// types separated by spaces ("*" for all, "!TYPE" excludes), optionally
// followed by conditions in brackets on layer, color, linetype,
// lineweight, handle and type: [layer=="WALL" & color!=1 | lineweight>=50],
// with == != < <= > >= and ? (regular expression), ! for not and
// parentheses; "i" after the brackets ignores the case of strings.
func Query(es entity.Entities, query string) (entity.Entities, error) {
	q := strings.TrimSpace(query)
	cond := ""
	ignoreCase := false
	if i := strings.Index(q, "["); i >= 0 {
		j := strings.LastIndex(q, "]")
		if j < i {
			return nil, errors.New("query: missing ]")
		}
		cond = q[i+1 : j]
		ignoreCase = strings.TrimSpace(q[j+1:]) == "i"
		q = strings.TrimSpace(q[:i])
	}
	include, exclude := map[string]bool{}, map[string]bool{}
	all := false
	for _, t := range strings.Fields(q) {
		t = strings.ToUpper(t)
		switch {
		case t == "*":
			all = true
		case strings.HasPrefix(t, "!"):
			exclude[t[1:]] = true
		default:
			include[t] = true
		}
	}
	if len(include) == 0 {
		all = true
	}
	var expr *qnode
	if strings.TrimSpace(cond) != "" {
		p := &qparser{toks: qtokens(cond)}
		var err error
		if expr, err = p.or(); err != nil {
			return nil, err
		}
		if p.pos < len(p.toks) {
			return nil, fmt.Errorf("query: unexpected %q", p.toks[p.pos])
		}
	}
	var out entity.Entities
	for _, e := range es {
		t := TypeName(e)
		if (!all && !include[t]) || exclude[t] {
			continue
		}
		if expr != nil {
			ok, err := expr.eval(e, ignoreCase)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
		}
		out = append(out, e)
	}
	return out, nil
}

type qnode struct {
	op          string // "&", "|", "!", or a comparison operator
	left, right *qnode
	attr, value string
	quoted      bool
}

func qtokens(s string) []string {
	re := regexp.MustCompile(`\s*("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'|==|!=|<=|>=|<|>|\?|&|\||!|\(|\)|[A-Za-z_][A-Za-z0-9_]*|-?[0-9.]+)`)
	var toks []string
	for _, m := range re.FindAllStringSubmatch(s, -1) {
		toks = append(toks, m[1])
	}
	return toks
}

type qparser struct {
	toks []string
	pos  int
}

func (p *qparser) peek() string {
	if p.pos < len(p.toks) {
		return p.toks[p.pos]
	}
	return ""
}

func (p *qparser) or() (*qnode, error) {
	l, err := p.and()
	for err == nil && p.peek() == "|" {
		p.pos++
		var r *qnode
		if r, err = p.and(); err == nil {
			l = &qnode{op: "|", left: l, right: r}
		}
	}
	return l, err
}

func (p *qparser) and() (*qnode, error) {
	l, err := p.not()
	for err == nil && p.peek() == "&" {
		p.pos++
		var r *qnode
		if r, err = p.not(); err == nil {
			l = &qnode{op: "&", left: l, right: r}
		}
	}
	return l, err
}

func (p *qparser) not() (*qnode, error) {
	if p.peek() == "!" {
		p.pos++
		n, err := p.not()
		return &qnode{op: "!", left: n}, err
	}
	if p.peek() == "(" {
		p.pos++
		n, err := p.or()
		if err == nil && p.peek() != ")" {
			err = errors.New("query: missing )")
		}
		p.pos++
		return n, err
	}
	if p.pos+2 >= len(p.toks) {
		return nil, errors.New("query: incomplete condition")
	}
	attr, op, val := p.toks[p.pos], p.toks[p.pos+1], p.toks[p.pos+2]
	p.pos += 3
	quoted := strings.HasPrefix(val, "\"") || strings.HasPrefix(val, "'")
	if quoted {
		val = val[1 : len(val)-1]
	}
	switch op {
	case "==", "!=", "<", "<=", ">", ">=", "?":
	default:
		return nil, fmt.Errorf("query: unknown operator %q", op)
	}
	return &qnode{op: op, attr: strings.ToLower(attr), value: val, quoted: quoted}, nil
}

func attrValue(e entity.Entity, attr string) (string, bool) {
	switch attr {
	case "layer":
		return e.Layer().Name(), false
	case "color":
		if c, ok := e.(interface{ ColorIndex() int }); ok {
			return strconv.Itoa(c.ColorIndex()), true
		}
	case "linetype":
		if l, ok := e.(interface{ LineTypeName() string }); ok {
			return l.LineTypeName(), false
		}
	case "lineweight":
		if l, ok := e.(interface{ LineWeight() int }); ok {
			return strconv.Itoa(l.LineWeight()), true
		}
	case "handle":
		if s, ok := e.(interface{ SourceHandle() string }); ok && s.SourceHandle() != "" {
			return s.SourceHandle(), false
		}
		return strings.ToUpper(strconv.FormatInt(int64(e.Handle()), 16)), false
	case "type":
		return TypeName(e), false
	}
	return "", false
}

func (n *qnode) eval(e entity.Entity, ignoreCase bool) (bool, error) {
	switch n.op {
	case "&":
		l, err := n.left.eval(e, ignoreCase)
		if err != nil || !l {
			return false, err
		}
		return n.right.eval(e, ignoreCase)
	case "|":
		l, err := n.left.eval(e, ignoreCase)
		if err != nil || l {
			return l, err
		}
		return n.right.eval(e, ignoreCase)
	case "!":
		l, err := n.left.eval(e, ignoreCase)
		return !l, err
	}
	v, numeric := attrValue(e, n.attr)
	if n.op == "?" {
		pat := n.value
		if ignoreCase {
			pat = "(?i)" + pat
		}
		re, err := regexp.Compile("^(?:" + pat + ")$")
		if err != nil {
			return false, err
		}
		return re.MatchString(v), nil
	}
	if numeric && !n.quoted {
		a, _ := strconv.ParseFloat(v, 64)
		b, err := strconv.ParseFloat(n.value, 64)
		if err != nil {
			return false, fmt.Errorf("query: %s needs a number", n.attr)
		}
		switch n.op {
		case "==":
			return a == b, nil
		case "!=":
			return a != b, nil
		case "<":
			return a < b, nil
		case "<=":
			return a <= b, nil
		case ">":
			return a > b, nil
		case ">=":
			return a >= b, nil
		}
	}
	a, b := v, n.value
	if ignoreCase {
		a, b = strings.ToLower(a), strings.ToLower(b)
	}
	switch n.op {
	case "==":
		return a == b, nil
	case "!=":
		return a != b, nil
	case "<":
		return a < b, nil
	case "<=":
		return a <= b, nil
	case ">":
		return a > b, nil
	case ">=":
		return a >= b, nil
	}
	return false, nil
}

// BlockReferenceCounts returns the number of references of each block: by
// INSERT entities (also in other blocks), dimensions (graphics blocks) and
// tables.
func BlockReferenceCounts(d *drawing.Drawing) map[string]int {
	counts := map[string]int{}
	d.ForEachEntity(func(e entity.Entity) {
		switch x := e.(type) {
		case *entity.Insert:
			counts[x.BlockName]++
		case *entity.Dimension:
			if x.BlockName != "" {
				counts[x.BlockName]++
			}
		case *entity.AcadTable:
			counts[x.BlockName]++
		}
	})
	return counts
}

// Purge removes the blocks without references (not the layouts), the
// layers without entities (not "0", "Defpoints" and the current layer),
// repeatedly while something is removed. It returns the names removed.
func Purge(d *drawing.Drawing) []string {
	var removed []string
	for {
		n := len(removed)
		counts := BlockReferenceCounts(d)
		for _, b := range d.Blocks() {
			name := b.Name
			if strings.HasPrefix(strings.ToUpper(name), "*MODEL_SPACE") || strings.HasPrefix(strings.ToUpper(name), "*PAPER_SPACE") || counts[name] > 0 {
				continue
			}
			if d.RemoveBlock(name) == nil {
				removed = append(removed, "BLOCK "+name)
			}
		}
		for name := range d.Layers {
			if name == "0" || strings.EqualFold(name, "Defpoints") || d.CurrentLayer == d.Layers[name] {
				continue
			}
			if d.RemoveLayer(name) == nil {
				removed = append(removed, "LAYER "+name)
			}
		}
		if len(removed) == n {
			return removed
		}
	}
}
