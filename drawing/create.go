package drawing

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/yofu/dxf/block"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/handle"
	"github.com/yofu/dxf/table"
)

// HatchFromEntities creates an associative solid HATCH bounded by
// entities: closed LWPOLYLINEs, CIRCLEs and full ELLIPSEs (one path each;
// the first is external, the others are holes). The entities refer to the
// hatch as reactor.
func (d *Drawing) HatchFromEntities(es ...entity.Entity) (*entity.Hatch, error) {
	if len(es) == 0 {
		return nil, errors.New("hatch: no boundary")
	}
	h := entity.NewHatch()
	h.Associative = true
	for _, e := range es {
		var b *entity.HatchBoundary
		switch x := e.(type) {
		case *entity.LwPolyline:
			if !x.Closed {
				return nil, errors.New("hatch: open polyline")
			}
			var vs [][]float64
			for _, v := range x.Vertices {
				vs = append(vs, []float64{v[0], v[1]})
			}
			b = h.AddPolylinePath(vs, append([]float64{}, x.Bulges...))
			h.Elevation = x.Elevation
		case *entity.Arc:
			return nil, errors.New("hatch: an arc is not closed")
		case *entity.Circle:
			b = h.AddEdgePath(&entity.HatchArc{Center: []float64{x.Center[0], x.Center[1]}, Radius: x.Radius, Start: 0, End: 360, CCW: true})
		case *entity.Ellipse:
			if x.End-x.Start < 2*math.Pi-1e-9 {
				return nil, errors.New("hatch: an elliptic arc is not closed")
			}
			b = h.AddEdgePath(&entity.HatchEllipse{Center: []float64{x.Center[0], x.Center[1]},
				MajorAxis: []float64{x.MajorAxis[0], x.MajorAxis[1]}, Ratio: x.Ratio, Start: 0, End: 360, CCW: true})
		default:
			return nil, fmt.Errorf("hatch: %T is not a supported boundary", e)
		}
		b.Sources = []entity.Ref{{Object: e}}
	}
	d.applyCurrent(h)
	d.AddEntity(h)
	for _, e := range es {
		if r, ok := e.(interface{ AddReactor(handle.Handler) }); ok {
			r.AddReactor(h)
		}
	}
	return h, nil
}

// PolygonMesh creates a polygon mesh (POLYLINE) of m x n vertices (row by
// row: m rows of n vertices), closed in the M and N directions if given.
func (d *Drawing) PolygonMesh(m, n int, vertices [][]float64, closedM, closedN bool) (*entity.Polyline, error) {
	if m < 2 || n < 2 || len(vertices) != m*n {
		return nil, errors.New("polygon mesh needs m x n (2 or more each) vertices")
	}
	p := entity.NewPolyline()
	p.Flag = 16
	if closedM {
		p.Flag |= 1
	}
	if closedN {
		p.Flag |= 32
	}
	p.MeshM, p.MeshN = m, n
	for _, v := range vertices {
		vx := p.AddVertex(v[0], v[1], coordAt(v, 2))
		vx.Flag = 64
	}
	d.add(p)
	return p, nil
}

// SplineFit creates a SPLINE defined by fit points (degree 3); CAD
// applications compute its control points.
func (d *Drawing) SplineFit(points [][]float64) (*entity.Spline, error) {
	if len(points) < 2 {
		return nil, errors.New("spline needs 2 or more fit points")
	}
	s := entity.NewSpline()
	s.Degree = 3
	s.Flag = 8 // planar
	for _, p := range points {
		s.Fits = append(s.Fits, []float64{coordAt(p, 0), coordAt(p, 1), coordAt(p, 2)})
	}
	s.Tolerance = []float64{1e-10, 1e-10, 1e-10}
	d.add(s)
	return s, nil
}

// SplineThrough creates a cubic SPLINE through points with its control
// points computed (global interpolation with chord length parameters, as
// ezdxf's add_cad_spline_control_frame); the fit points are kept.
func (d *Drawing) SplineThrough(points [][]float64) (*entity.Spline, error) {
	n := len(points)
	if n < 2 {
		return nil, errors.New("spline needs 2 or more points")
	}
	deg := 3
	if n-1 < deg {
		deg = n - 1
	}
	// chord length parameters
	t := make([]float64, n)
	total := 0.0
	for i := 1; i < n; i++ {
		total += dist3(points[i-1], points[i])
		t[i] = total
	}
	if total == 0 {
		return nil, errors.New("spline: all points are equal")
	}
	for i := range t {
		t[i] /= total
	}
	// averaged knots
	knots := make([]float64, n+deg+1)
	for i := 0; i <= deg; i++ {
		knots[i], knots[n+i] = 0, 1
	}
	for j := 1; j < n-deg; j++ {
		s := 0.0
		for i := j; i < j+deg; i++ {
			s += t[i]
		}
		knots[j+deg] = s / float64(deg)
	}
	// solve N * P = Q
	a := make([][]float64, n)
	for i := range a {
		a[i] = basisFuncs(knots, deg, n, t[i])
	}
	ctrl := make([][]float64, n)
	for k := 0; k < 3; k++ {
		b := make([]float64, n)
		for i := range b {
			b[i] = coordAt(points[i], k)
		}
		x, err := solve(a, b)
		if err != nil {
			return nil, err
		}
		for i := range x {
			if ctrl[i] == nil {
				ctrl[i] = make([]float64, 3)
			}
			ctrl[i][k] = x[i]
		}
	}
	s := entity.NewSpline()
	s.Degree = deg
	s.Flag = 8
	s.Knots = knots
	s.Controls = ctrl
	for _, p := range points {
		s.Fits = append(s.Fits, []float64{coordAt(p, 0), coordAt(p, 1), coordAt(p, 2)})
	}
	d.add(s)
	return s, nil
}

func dist3(p, q []float64) float64 {
	dx, dy, dz := coordAt(q, 0)-coordAt(p, 0), coordAt(q, 1)-coordAt(p, 1), coordAt(q, 2)-coordAt(p, 2)
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// basisFuncs returns the values of the n B-spline basis functions of
// degree deg at u (Cox-de Boor).
func basisFuncs(knots []float64, deg, n int, u float64) []float64 {
	N := make([]float64, len(knots)-1)
	for i := range N {
		if (knots[i] <= u && u < knots[i+1]) || (u == knots[len(knots)-1] && knots[i] < u && u <= knots[i+1]) {
			N[i] = 1
		}
	}
	for p := 1; p <= deg; p++ {
		for i := 0; i+p < len(knots)-1; i++ {
			v := 0.0
			if d := knots[i+p] - knots[i]; d > 0 {
				v += (u - knots[i]) / d * N[i]
			}
			if d := knots[i+p+1] - knots[i+1]; d > 0 {
				v += (knots[i+p+1] - u) / d * N[i+1]
			}
			N[i] = v
		}
	}
	return N[:n]
}

// solve solves a x = b by Gaussian elimination with partial pivoting.
func solve(a [][]float64, b []float64) ([]float64, error) {
	n := len(b)
	m := make([][]float64, n)
	for i := range m {
		m[i] = append(append([]float64{}, a[i]...), b[i])
	}
	for c := 0; c < n; c++ {
		p := c
		for r := c + 1; r < n; r++ {
			if math.Abs(m[r][c]) > math.Abs(m[p][c]) {
				p = r
			}
		}
		if math.Abs(m[p][c]) < 1e-14 {
			return nil, errors.New("spline: singular system")
		}
		m[c], m[p] = m[p], m[c]
		for r := c + 1; r < n; r++ {
			f := m[r][c] / m[c][c]
			for k := c; k <= n; k++ {
				m[r][k] -= f * m[c][k]
			}
		}
	}
	x := make([]float64, n)
	for r := n - 1; r >= 0; r-- {
		s := m[r][n]
		for k := r + 1; k < n; k++ {
			s -= m[r][k] * x[k]
		}
		x[r] = s / m[r][r]
	}
	return x, nil
}

// Arrow names of ArrowBlock (the blocks of AutoCAD's arrowheads, DIMBLK).
var ArrowNames = []string{"_CLOSEDFILLED", "_CLOSEDBLANK", "_CLOSED", "_DOT", "_ARCHTICK", "_OBLIQUE",
	"_OPEN", "_ORIGIN", "_ORIGIN2", "_OPEN90", "_OPEN30", "_DOTSMALL", "_DOTBLANK", "_SMALL",
	"_BOXBLANK", "_BOXFILLED", "_DATUMBLANK", "_DATUMFILLED", "_INTEGRAL", "_NONE"}

// ArrowBlock returns the block of an arrowhead (see ArrowNames), adding
// it if it does not exist. Arrowheads have size 1, the tip at the origin
// and point in +X.
func (d *Drawing) ArrowBlock(name string) (*block.Block, error) {
	name = strings.ToUpper(name)
	if name == "" {
		name = "_CLOSEDFILLED"
	}
	if b, err := d.Block(name); err == nil {
		return b, nil
	}
	known := false
	for _, n := range ArrowNames {
		known = known || n == name
	}
	if !known {
		return nil, fmt.Errorf("unknown arrow: %s", name)
	}
	b, err := d.AddBlock(name, 0, 0, 0)
	if err != nil {
		return nil, err
	}
	layer0 := d.Layers["0"]
	add := func(e entity.Entity) {
		e.SetLayer(layer0)
		if c, ok := e.(interface{ SetColorByBlock() }); ok {
			c.SetColorByBlock()
		}
		b.AddEntity(e)
	}
	line := func(x1, y1, x2, y2 float64) {
		l := entity.NewLine()
		l.Start, l.End = []float64{x1, y1, 0}, []float64{x2, y2, 0}
		add(l)
	}
	solid := func(pts ...[]float64) {
		s := entity.NewSolid()
		for i := range s.Points {
			p := pts[i%len(pts)]
			if i >= len(pts) {
				p = pts[len(pts)-1]
			}
			s.Points[i] = []float64{p[0], p[1], 0}
		}
		add(s)
	}
	circle := func(x, r float64) {
		c := entity.NewCircle()
		c.Center, c.Radius = []float64{x, 0, 0}, r
		add(c)
	}
	h := 1.0 / 6
	switch name {
	case "_CLOSEDFILLED":
		solid([]float64{0, 0}, []float64{-1, h}, []float64{-1, -h})
	case "_CLOSEDBLANK", "_CLOSED":
		line(0, 0, -1, h)
		line(-1, h, -1, -h)
		line(-1, -h, 0, 0)
		if name == "_CLOSED" {
			line(-1, 0, 0, 0)
		}
	case "_DOT":
		lw := entity.NewLwPolyline(2)
		lw.Vertices = [][]float64{{-0.25, 0}, {0.25, 0}}
		lw.Bulges = []float64{1, 1}
		lw.ConstantWidth = 0.5
		lw.Closed = true
		add(lw)
		line(-1, 0, -0.5, 0)
	case "_DOTSMALL":
		lw := entity.NewLwPolyline(2)
		lw.Vertices = [][]float64{{-0.0625, 0}, {0.0625, 0}}
		lw.Bulges = []float64{1, 1}
		lw.ConstantWidth = 0.125
		lw.Closed = true
		add(lw)
	case "_DOTBLANK":
		circle(0, 0.5)
		line(-1, 0, -0.5, 0)
	case "_SMALL":
		circle(0, 0.25)
	case "_ARCHTICK", "_OBLIQUE":
		line(-0.5, -0.5, 0.5, 0.5)
		if name == "_ARCHTICK" {
			lw := entity.NewLwPolyline(2)
			lw.Vertices = [][]float64{{-0.5, -0.5}, {0.5, 0.5}}
			lw.ConstantWidth = 0.15
			add(lw)
		}
	case "_OPEN":
		line(0, 0, -1, h)
		line(0, 0, -1, -h)
		line(-1, 0, 0, 0)
	case "_OPEN90":
		line(0, 0, -0.5, 0.5)
		line(0, 0, -0.5, -0.5)
		line(-1, 0, 0, 0)
	case "_OPEN30":
		t := math.Tan(15 * math.Pi / 180)
		line(0, 0, -1, t)
		line(0, 0, -1, -t)
		line(-1, 0, 0, 0)
	case "_ORIGIN":
		circle(0, 0.5)
		line(-1, 0, 0, 0)
	case "_ORIGIN2":
		circle(0, 0.5)
		circle(0, 0.25)
		line(-1, 0, -0.5, 0)
	case "_BOXBLANK":
		line(-0.5, -0.5, 0.5, -0.5)
		line(0.5, -0.5, 0.5, 0.5)
		line(0.5, 0.5, -0.5, 0.5)
		line(-0.5, 0.5, -0.5, -0.5)
		line(-1, 0, -0.5, 0)
	case "_BOXFILLED":
		solid([]float64{-0.5, -0.5}, []float64{0.5, -0.5}, []float64{-0.5, 0.5}, []float64{0.5, 0.5})
		line(-1, 0, -0.5, 0)
	case "_DATUMBLANK":
		line(0, 0.5, -1, 0)
		line(-1, 0, 0, -0.5)
		line(0, -0.5, 0, 0.5)
	case "_DATUMFILLED":
		solid([]float64{0, 0.5}, []float64{-1, 0}, []float64{0, -0.5})
	case "_INTEGRAL":
		a := entity.NewArc(nil)
		a.Center, a.Radius, a.Angle = []float64{0.44, -0.09, 0}, 0.45, []float64{102, 168}
		add(a)
		a2 := entity.NewArc(nil)
		a2.Center, a2.Radius, a2.Angle = []float64{-0.44, 0.09, 0}, 0.45, []float64{282, 348}
		add(a2)
	case "_NONE":
	}
	return b, nil
}

// InsertArrow inserts the arrowhead name (see ArrowBlock) with its tip at
// tip, size and rotation (degrees; the direction it points to).
func (d *Drawing) InsertArrow(name string, tip []float64, size, rotation float64) (*entity.Insert, error) {
	b, err := d.ArrowBlock(name)
	if err != nil {
		return nil, err
	}
	ins, err := d.Insert(b.Name, coordAt(tip, 0), coordAt(tip, 1), coordAt(tip, 2))
	if err != nil {
		return nil, err
	}
	ins.Scale = []float64{size, size, size}
	ins.Rotation = rotation
	return ins, nil
}

// LoadLineTypes adds the linetypes of a .lin file ("*NAME,description"
// followed by "A,dash,gap,..."; complex elements ["TEXT",STYLE,S=..,R=..,
// X=..,Y=..] use a text style of the drawing, [shape number,file.shx,...]
// the shape file). Linetypes that exist are not changed. It returns the
// linetypes added.
func (d *Drawing) LoadLineTypes(content string) ([]*table.LineType, error) {
	var added []*table.LineType
	name, desc := "", ""
	for n, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if i := strings.Index(line, ";;"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "*") {
			parts := strings.SplitN(line[1:], ",", 2)
			name = strings.TrimSpace(parts[0])
			desc = ""
			if len(parts) > 1 {
				desc = strings.TrimSpace(parts[1])
			}
			continue
		}
		if name == "" {
			continue
		}
		if !strings.HasPrefix(strings.ToUpper(line), "A") {
			return added, fmt.Errorf("line %d: alignment A expected", n+1)
		}
		es, err := d.linElements(line[strings.Index(line, ",")+1:])
		if err != nil {
			return added, fmt.Errorf("line %d: %w", n+1, err)
		}
		if lt, err := d.LineType(name); err == nil && lt != nil {
			name = ""
			continue
		}
		lt, err := d.AddComplexLineType(name, desc, es...)
		if err != nil {
			return added, err
		}
		added = append(added, lt)
		name = ""
	}
	return added, nil
}

// linElements parses the elements of a .lin pattern line.
func (d *Drawing) linElements(s string) ([]*table.LineTypeElement, error) {
	var es []*table.LineTypeElement
	for len(strings.TrimSpace(s)) > 0 {
		s = strings.TrimLeft(s, " ,")
		if strings.HasPrefix(s, "[") {
			end := strings.Index(s, "]")
			if end < 0 {
				return nil, errors.New("unclosed [")
			}
			el, err := d.linComplex(s[1:end])
			if err != nil {
				return nil, err
			}
			if len(es) == 0 {
				return nil, errors.New("a complex element must follow a dash or gap")
			}
			prev := es[len(es)-1]
			el.Length, prev.Length = prev.Length, 0
			es[len(es)-1] = el
			// the gap before the element keeps its length; the element is drawn at its start
			es = append(es[:len(es)-1], el)
			s = s[end+1:]
			continue
		}
		i := strings.Index(s, ",")
		f := s
		if i >= 0 {
			f, s = s[:i], s[i+1:]
		} else {
			s = ""
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(f), 64)
		if err != nil {
			return nil, err
		}
		es = append(es, &table.LineTypeElement{Length: v})
	}
	return es, nil
}

// linComplex parses a complex element ["TEXT",STYLE,...] or [num,file.shx,...].
func (d *Drawing) linComplex(s string) (*table.LineTypeElement, error) {
	parts := strings.Split(s, ",")
	if len(parts) < 2 {
		return nil, errors.New("complex element needs a text or shape and a style or file")
	}
	el := &table.LineTypeElement{Scale: 1, Offset: []float64{0, 0}}
	first := strings.TrimSpace(parts[0])
	ref := strings.TrimSpace(parts[1])
	if strings.HasPrefix(first, "\"") {
		el.Flags = table.LT_TEXT
		el.Text = strings.Trim(first, "\"")
		st, ok := d.Styles[ref]
		if !ok {
			for k, v := range d.Styles {
				if strings.EqualFold(k, ref) {
					st, ok = v, true
				}
			}
		}
		if !ok {
			return nil, fmt.Errorf("text style %s doesn't exist", ref)
		}
		el.Style = st
	} else {
		num, err := strconv.Atoi(first)
		if err != nil {
			return nil, fmt.Errorf("shape %s: only shape numbers are supported", first)
		}
		el.Flags, el.Shape, el.Style = table.LT_SHAPE, num, d.AddShapeFile(ref)
	}
	for _, p := range parts[2:] {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(kv) != 2 {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(kv[1]), 64)
		if err != nil {
			return nil, err
		}
		switch strings.ToUpper(strings.TrimSpace(kv[0])) {
		case "S":
			el.Scale = v
		case "R", "U":
			el.Rotation = v * math.Pi / 180
		case "A":
			el.Rotation = v * math.Pi / 180
			el.Flags |= table.LT_ABSOLUTE_ROTATION
		case "X":
			el.Offset[0] = v
		case "Y":
			el.Offset[1] = v
		}
	}
	return el, nil
}

// RevisionCloud creates a revision cloud (closed LWPOLYLINE of arcs, as
// AutoCAD's REVCLOUD and ezdxf): the polygon vertices are divided into arcs
// of about segmentLength; calligraphy gives the arcs widths.
func (d *Drawing) RevisionCloud(vertices [][]float64, segmentLength float64, calligraphy bool) (*entity.LwPolyline, error) {
	if segmentLength < 1e-6 {
		return nil, errors.New("segment length too small")
	}
	if len(vertices) < 3 {
		return nil, errors.New("revision cloud needs 3 or more vertices")
	}
	vs := append([][]float64{}, vertices...)
	if f, l := vs[0], vs[len(vs)-1]; f[0] != l[0] || f[1] != l[1] {
		vs = append(vs, f)
	}
	area := 0.0
	for i := 0; i+1 < len(vs); i++ {
		area += vs[i][0]*vs[i+1][1] - vs[i+1][0]*vs[i][1]
	}
	bulge := 0.520567050552
	if area < 0 { // clockwise
		bulge = -bulge
	}
	endWidth := 0.0
	if calligraphy {
		endWidth = segmentLength * 0.1
	}
	var pts [][]float64
	for i := 0; i+1 < len(vs); i++ {
		s, e := vs[i], vs[i+1]
		pts = append(pts, []float64{s[0], s[1]})
		l := math.Hypot(e[0]-s[0], e[1]-s[1])
		if l <= segmentLength {
			continue
		}
		count := int(math.Ceil(l / segmentLength))
		for k := 1; k < count; k++ {
			f := float64(k) / float64(count)
			pts = append(pts, []float64{s[0] + (e[0]-s[0])*f, s[1] + (e[1]-s[1])*f})
		}
	}
	lw := entity.NewLwPolyline(len(pts))
	lw.Vertices = pts
	lw.Closed = true
	for i := range pts {
		lw.Bulges[i] = bulge
		lw.Widths[i] = []float64{0, endWidth}
	}
	lw.SetXData([][2]string{{"1001", "RevcloudProps"}, {"1070", "0"}, {"1040", strconv.FormatFloat(segmentLength, 'f', -1, 64)}})
	d.add(lw)
	return lw, nil
}
