// Package proxygraphic decodes proxy graphics: the binary graphics that
// AutoCAD stores (group codes 92 and 310) for entities of applications
// that are not loaded (ACAD_PROXY_ENTITY and custom entities), as ezdxf's
// proxygraphic module.
//
// Supported commands: circles, arcs (also by 3 points), polylines,
// polygons, meshes, shells, text (ANSI and Unicode), xlines, rays, and the
// attributes color, true color, layer, linetype, lineweight and linetype
// scale, and model transformations. Other commands (markers, clipping,
// fill, LWPOLYLINE data, plot styles) are skipped.
package proxygraphic

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf16"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/geometry"
	"github.com/yofu/dxf/table"
)

// Command types.
const (
	Extents          = 1
	Circle           = 2
	Circle3P         = 3
	CircularArc      = 4
	CircularArc3P    = 5
	Polyline         = 6
	Polygon          = 7
	Mesh             = 8
	Shell            = 9
	Text             = 10
	Text2            = 11
	XLine            = 12
	Ray              = 13
	SubentColor      = 14
	SubentLayer      = 16
	SubentLineType   = 18
	SubentMarker     = 19
	SubentFill       = 20
	SubentTrueColor  = 22
	SubentLineWeight = 23
	SubentLtscale    = 24
	SubentThickness  = 25
	SubentPlotStyle  = 26
	PushClip         = 27
	PopClip          = 28
	PushMatrix       = 29
	PushMatrix2      = 30
	PopMatrix        = 31
	PolylineNormal   = 32
	LwPolyline       = 33
	UnicodeText      = 36
	UnicodeText2     = 38
)

// Options gives the tables the attribute commands refer to by index (in
// the order of the tables of the drawing); with nil tables the layer and
// linetype commands are ignored.
type Options struct {
	Layers    []*table.Layer
	LineTypes []*table.LineType
	Layer     *table.Layer // layer of entities drawn before a layer command (nil: 0)
}

// Data returns the proxy graphics of group codes read: the 310 groups
// after the first 92 (or 160) group.
func Data(codes [][2]string) ([]byte, error) {
	var out []byte
	found := false
	for _, p := range codes {
		c := strings.TrimSpace(p[0])
		switch {
		case !found && (c == "92" || c == "160"):
			found = true
		case found && c == "310":
			b, err := hex.DecodeString(strings.TrimSpace(p[1]))
			if err != nil {
				return nil, err
			}
			out = append(out, b...)
		case found && len(out) > 0:
			return out, nil
		}
	}
	if len(out) == 0 {
		return nil, errors.New("no proxy graphics")
	}
	return out, nil
}

type reader struct {
	b   []byte
	pos int
	err error
}

func (r *reader) need(n int) bool {
	if r.err != nil {
		return false
	}
	if r.pos+n > len(r.b) {
		r.err = fmt.Errorf("proxy graphics: data too short")
		return false
	}
	return true
}

func (r *reader) u32() uint32 {
	if !r.need(4) {
		return 0
	}
	v := binary.LittleEndian.Uint32(r.b[r.pos:])
	r.pos += 4
	return v
}

func (r *reader) i32() int { return int(int32(r.u32())) }

func (r *reader) f64() float64 {
	if !r.need(8) {
		return 0
	}
	v := math.Float64frombits(binary.LittleEndian.Uint64(r.b[r.pos:]))
	r.pos += 8
	return v
}

func (r *reader) vec() []float64 { return []float64{r.f64(), r.f64(), r.f64()} }

func (r *reader) vecs(n int) [][]float64 {
	if n < 0 || !r.need(24*n) {
		return nil
	}
	ps := make([][]float64, n)
	for i := range ps {
		ps[i] = r.vec()
	}
	return ps
}

// str reads a null terminated string padded to 4 bytes.
func (r *reader) str() string {
	if r.err != nil {
		return ""
	}
	end := r.pos
	for end < len(r.b) && r.b[end] != 0 {
		end++
	}
	s := string(r.b[r.pos:end])
	r.pos = (end + 4) &^ 3
	return s
}

// ustr reads a null terminated UTF-16 string padded to 4 bytes.
func (r *reader) ustr() string {
	if r.err != nil {
		return ""
	}
	var u []uint16
	end := r.pos
	for end+1 < len(r.b) {
		c := binary.LittleEndian.Uint16(r.b[end:])
		end += 2
		if c == 0 {
			break
		}
		u = append(u, c)
	}
	r.pos = (end + 3) &^ 3
	return string(utf16.Decode(u))
}

type matrix [16]float64 // row major, column vectors: translation in 3, 7, 11

var identity = matrix{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}

func (m matrix) mul(n matrix) matrix {
	var o matrix
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				o[4*i+j] += m[4*i+k] * n[4*k+j]
			}
		}
	}
	return o
}

func (m matrix) point(p []float64) []float64 {
	return []float64{
		m[0]*p[0] + m[1]*p[1] + m[2]*p[2] + m[3],
		m[4]*p[0] + m[5]*p[1] + m[6]*p[2] + m[7],
		m[8]*p[0] + m[9]*p[1] + m[10]*p[2] + m[11],
	}
}

func (m matrix) vector(p []float64) []float64 {
	return []float64{
		m[0]*p[0] + m[1]*p[1] + m[2]*p[2],
		m[4]*p[0] + m[5]*p[1] + m[6]*p[2],
		m[8]*p[0] + m[9]*p[1] + m[10]*p[2],
	}
}

func norm(v []float64) float64 { return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2]) }

func unit(v []float64) []float64 {
	l := norm(v)
	if l == 0 {
		return []float64{0, 0, 1}
	}
	return []float64{v[0] / l, v[1] / l, v[2] / l}
}

func dot(a, b []float64) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }

func cross(a, b []float64) []float64 {
	return []float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}

// ocs returns the axes of the object coordinate system of normal n.
func ocs(n []float64) (ax, ay, az []float64) {
	az = unit(n)
	ax, ay, err := geometry.ArbitraryAxis(az)
	if err != nil {
		return []float64{1, 0, 0}, []float64{0, 1, 0}, []float64{0, 0, 1}
	}
	return ax, ay, az
}

func toOCS(p, ax, ay, az []float64) []float64 {
	return []float64{dot(p, ax), dot(p, ay), dot(p, az)}
}

type state struct {
	opt       Options
	color     int     // ACI, 0: BYBLOCK, 256: BYLAYER
	trueColor []uint8 // nil: none
	layer     *table.Layer
	lineType  *table.LineType
	weight    int
	ltscale   float64
	matrices  []matrix
	out       entity.Entities
}

func (s *state) m() matrix { return s.matrices[len(s.matrices)-1] }

func (s *state) add(e entity.Entity) {
	if s.layer != nil {
		e.SetLayer(s.layer)
	}
	switch s.color {
	case entity.ColorByLayer:
	case entity.ColorByBlock:
		if c, ok := e.(interface{ SetColorByBlock() }); ok {
			c.SetColorByBlock()
		}
	default:
		if c, ok := e.(interface{ SetColor(color.ColorNumber) }); ok {
			c.SetColor(color.ColorNumber(s.color))
		}
	}
	if s.trueColor != nil {
		if c, ok := e.(interface{ SetTrueColor(r, g, b uint8) }); ok {
			c.SetTrueColor(s.trueColor[0], s.trueColor[1], s.trueColor[2])
		}
	}
	if s.lineType != nil {
		if c, ok := e.(interface{ SetLineType(*table.LineType) }); ok {
			c.SetLineType(s.lineType)
		}
	}
	if s.weight != table.LineWeightByLayer {
		if c, ok := e.(interface{ SetLineWeight(int) int }); ok {
			c.SetLineWeight(s.weight)
		}
	}
	if s.ltscale != 1 {
		if c, ok := e.(interface{ SetLtscale(float64) }); ok {
			c.SetLtscale(s.ltscale)
		}
	}
	s.out = append(s.out, e)
}

func (s *state) circle(center []float64, radius float64, normal []float64) {
	m := s.m()
	center, normal = m.point(center), m.vector(normal)
	radius *= math.Cbrt(math.Abs(det3(m)))
	ax, ay, az := ocs(normal)
	c := entity.NewCircle()
	c.Center = toOCS(center, ax, ay, az)
	c.Radius = radius
	c.Direction = az
	s.add(c)
}

func (s *state) arc(center []float64, radius float64, normal, startVec []float64, sweep float64) {
	if math.Abs(sweep) >= 2*math.Pi-1e-12 {
		s.circle(center, radius, normal)
		return
	}
	m := s.m()
	center, normal, startVec = m.point(center), m.vector(normal), m.vector(startVec)
	radius *= math.Cbrt(math.Abs(det3(m)))
	ax, ay, az := ocs(normal)
	a1 := math.Atan2(dot(startVec, ay), dot(startVec, ax))
	a2 := a1 + sweep
	if sweep < 0 {
		a1, a2 = a2, a1
	}
	a := entity.NewArc(nil)
	a.Center = toOCS(center, ax, ay, az)
	a.Radius = radius
	a.Direction = az
	a.Angle = []float64{a1 * 180 / math.Pi, a2 * 180 / math.Pi}
	s.add(a)
}

func det3(m matrix) float64 {
	return m[0]*(m[5]*m[10]-m[6]*m[9]) - m[1]*(m[4]*m[10]-m[6]*m[8]) + m[2]*(m[4]*m[9]-m[5]*m[8])
}

// circleFrom3 returns the circle through three points.
func circleFrom3(p1, p2, p3 []float64) (center []float64, radius float64, normal []float64, ok bool) {
	a := []float64{p2[0] - p1[0], p2[1] - p1[1], p2[2] - p1[2]}
	b := []float64{p3[0] - p1[0], p3[1] - p1[1], p3[2] - p1[2]}
	n := cross(a, b)
	nn := dot(n, n)
	if nn < 1e-24 {
		return nil, 0, nil, false
	}
	// center = p1 + (|a|²(b×n) + |b|²(n×a)) / (2|n|²)
	t1, t2 := cross(b, n), cross(n, a)
	aa, bb := dot(a, a), dot(b, b)
	c := make([]float64, 3)
	for i := range c {
		c[i] = p1[i] + (aa*t1[i]+bb*t2[i])/(2*nn)
	}
	d := []float64{p1[0] - c[0], p1[1] - c[1], p1[2] - c[2]}
	return c, norm(d), unit(n), true
}

func (s *state) polyline(ps [][]float64, closed bool, normal []float64) {
	m := s.m()
	tp := make([][]float64, len(ps))
	for i, p := range ps {
		tp[i] = m.point(p)
	}
	if len(tp) == 2 && !closed {
		l := entity.NewLine()
		l.Start, l.End = tp[0], tp[1]
		s.add(l)
		return
	}
	if normal != nil {
		normal = unit(m.vector(normal))
	}
	planar := normal != nil || func() bool {
		for _, p := range tp {
			if math.Abs(p[2]-tp[0][2]) > 1e-12 {
				return false
			}
		}
		return true
	}()
	if planar {
		if normal == nil {
			normal = []float64{0, 0, 1}
		}
		ax, ay, az := ocs(normal)
		lw := entity.NewLwPolyline(len(tp))
		for i, p := range tp {
			q := toOCS(p, ax, ay, az)
			lw.Vertices[i] = []float64{q[0], q[1]}
			lw.Elevation = q[2]
		}
		lw.Closed = closed
		if normal[2] != 1 {
			lw.Extrusion = az
		}
		s.add(lw)
		return
	}
	pl := entity.NewPolyline()
	pl.Flag = 8 // 3D polyline
	if closed {
		pl.Flag |= 1
	}
	for _, p := range tp {
		v := pl.AddVertex(p[0], p[1], p[2])
		v.Flag = 32
	}
	s.add(pl)
}

func (s *state) mesh(vs [][]float64, faces [][]int) {
	m := s.m()
	for i, v := range vs {
		vs[i] = m.point(v)
	}
	s.add(entity.NewMesh(vs, faces))
}

func (s *state) text(start, normal, dir []float64, height, width, oblique float64, str string) {
	m := s.m()
	start, normal, dir = m.point(start), m.vector(normal), m.vector(dir)
	height *= norm(m.vector(unit(cross(normal, dir))))
	ax, ay, az := ocs(normal)
	t := entity.NewText()
	t.Coord1 = toOCS(start, ax, ay, az)
	t.Height = height
	if width > 0 {
		t.WidthFactor = width
	}
	t.ObliqueAngle = oblique * 180 / math.Pi
	t.Rotation = math.Atan2(dot(dir, ay), dot(dir, ax)) * 180 / math.Pi
	t.Value = str
	if az[2] != 1 {
		t.Extrusion = az
	}
	s.add(t)
}

// Decode returns the entities drawn by proxy graphics data.
func Decode(data []byte, opt Options) (entity.Entities, error) {
	if len(data) < 8 {
		return nil, errors.New("proxy graphics: data too short")
	}
	s := &state{opt: opt, color: entity.ColorByLayer, weight: table.LineWeightByLayer, ltscale: 1, matrices: []matrix{identity}, layer: opt.Layer}
	pos := 8
	for pos+8 <= len(data) {
		size := int(binary.LittleEndian.Uint32(data[pos:]))
		typ := int(binary.LittleEndian.Uint32(data[pos+4:]))
		if size < 8 || pos+size > len(data) {
			return s.out, fmt.Errorf("proxy graphics: invalid record size %d at %d", size, pos)
		}
		r := &reader{b: data[pos+8 : pos+size]}
		pos += size
		if err := s.command(typ, r); err != nil {
			return s.out, err
		}
		if r.err != nil {
			return s.out, fmt.Errorf("command %d: %w", typ, r.err)
		}
	}
	return s.out, nil
}

func (s *state) command(typ int, r *reader) error {
	switch typ {
	case Circle:
		c, rad, n := r.vec(), r.f64(), r.vec()
		if r.err == nil {
			s.circle(c, rad, n)
		}
	case Circle3P:
		p1, p2, p3 := r.vec(), r.vec(), r.vec()
		if c, rad, n, ok := circleFrom3(p1, p2, p3); ok && r.err == nil {
			s.circle(c, rad, n)
		}
	case CircularArc:
		c, rad, n, sv, sweep := r.vec(), r.f64(), r.vec(), r.vec(), r.f64()
		if r.err == nil {
			s.arc(c, rad, n, sv, sweep)
		}
	case CircularArc3P:
		p1, p2, p3 := r.vec(), r.vec(), r.vec()
		if c, rad, n, ok := circleFrom3(p1, p2, p3); ok && r.err == nil {
			ax, ay, _ := ocs(n)
			ang := func(p []float64) float64 {
				v := []float64{p[0] - c[0], p[1] - c[1], p[2] - c[2]}
				return math.Atan2(dot(v, ay), dot(v, ax))
			}
			a1, a2, a3 := ang(p1), ang(p2), ang(p3)
			sweep := math.Mod(a3-a1+4*math.Pi, 2*math.Pi)
			if math.Mod(a2-a1+4*math.Pi, 2*math.Pi) > sweep { // clockwise through p2
				sweep -= 2 * math.Pi
			}
			s.arc(c, rad, n, []float64{p1[0] - c[0], p1[1] - c[1], p1[2] - c[2]}, sweep)
		}
	case Polyline, Polygon:
		n := r.i32()
		ps := r.vecs(n)
		if r.err == nil && n > 1 {
			s.polyline(ps, typ == Polygon, nil)
		}
	case PolylineNormal:
		n := r.i32()
		ps := r.vecs(n)
		normal := r.vec()
		if r.err == nil && n > 1 {
			s.polyline(ps, false, normal)
		}
	case Mesh:
		rows, cols := r.i32(), r.i32()
		vs := r.vecs(rows * cols)
		if r.err == nil && rows > 1 && cols > 1 {
			var faces [][]int
			for i := 0; i+1 < rows; i++ {
				for j := 0; j+1 < cols; j++ {
					k := i*cols + j
					faces = append(faces, []int{k, k + 1, k + cols + 1, k + cols})
				}
			}
			s.mesh(vs, faces)
		}
	case Shell:
		nv := r.i32()
		vs := r.vecs(nv)
		ne := r.i32()
		if ne < 0 || !r.need(4*ne) {
			return nil
		}
		ids := make([]int, ne)
		for i := range ids {
			ids[i] = r.i32()
		}
		var faces [][]int
		for k := 0; k < len(ids); {
			n := ids[k]
			if n < 0 {
				n = -n // a hole: kept as a face
			}
			k++
			if k+n > len(ids) {
				break
			}
			f := append([]int{}, ids[k:k+n]...)
			ok := true
			for _, i := range f {
				if i < 0 || i >= nv {
					ok = false
				}
			}
			if ok {
				faces = append(faces, f)
			}
			k += n
		}
		if r.err == nil {
			s.mesh(vs, faces)
		}
	case Text:
		start, normal, dir := r.vec(), r.vec(), r.vec()
		h, w, o := r.f64(), r.f64(), r.f64()
		str := r.str()
		if r.err == nil {
			s.text(start, normal, dir, h, w, o, str)
		}
	case Text2, UnicodeText, UnicodeText2:
		start, normal, dir := r.vec(), r.vec(), r.vec()
		var str string
		if typ == Text2 {
			str = r.str()
		} else {
			str = r.ustr()
		}
		r.i32() // ignore length of string
		r.i32() // raw
		h, w, o := r.f64(), r.f64(), r.f64()
		if r.err == nil {
			s.text(start, normal, dir, h, w, o, str)
		}
	case XLine, Ray:
		p1, p2 := r.vec(), r.vec()
		if r.err == nil {
			m := s.m()
			p1, p2 = m.point(p1), m.point(p2)
			d := unit([]float64{p2[0] - p1[0], p2[1] - p1[1], p2[2] - p1[2]})
			if typ == XLine {
				s.add(entity.NewXLine(p1, d))
			} else {
				s.add(entity.NewRay(p1, d))
			}
		}
	case SubentColor:
		c := r.u32()
		if r.err == nil {
			s.color, s.trueColor = int(c&0xff), nil
			if c == 256 {
				s.color = entity.ColorByLayer
			}
			if c == 0 {
				s.color = entity.ColorByBlock
			}
		}
	case SubentTrueColor:
		c := r.u32()
		switch c >> 24 {
		case 0xc0:
			s.color, s.trueColor = entity.ColorByBlock, nil
		case 0xc1:
			s.color, s.trueColor = entity.ColorByLayer, nil
		case 0xc2:
			s.trueColor = []uint8{uint8(c >> 16), uint8(c >> 8), uint8(c)}
		case 0xc3:
			s.color, s.trueColor = int(c&0xff), nil
		}
	case SubentLayer:
		i := r.i32()
		if i >= 0 && i < len(s.opt.Layers) {
			s.layer = s.opt.Layers[i]
		}
	case SubentLineType:
		i := r.i32()
		switch {
		case i == 32767:
			s.lineType = nil // BYLAYER
		case i >= 0 && i < len(s.opt.LineTypes):
			s.lineType = s.opt.LineTypes[i]
		}
	case SubentLineWeight:
		if w := r.i32(); r.err == nil {
			s.weight = w
		}
	case SubentLtscale:
		if v := r.f64(); r.err == nil {
			s.ltscale = v
		}
	case PushMatrix, PushMatrix2:
		var m matrix
		for i := range m {
			m[i] = r.f64()
		}
		if r.err == nil {
			s.matrices = append(s.matrices, s.m().mul(m))
		}
	case PopMatrix:
		if len(s.matrices) > 1 {
			s.matrices = s.matrices[:len(s.matrices)-1]
		}
	}
	return nil
}
