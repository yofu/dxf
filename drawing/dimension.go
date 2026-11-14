package drawing

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/yofu/dxf/block"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/table"
)

// DimStyle returns the dimension style named name (case-insensitive).
func (d *Drawing) DimStyle(name string) (*table.DimStyle, error) {
	st, err := d.Sections[TABLES].(table.Tables)[table.DIMSTYLE].Contains(name)
	if err != nil {
		return nil, fmt.Errorf("dimstyle %s", err.Error())
	}
	return st.(*table.DimStyle), nil
}

// LinearDimension creates a linear DIMENSION between p1 and p2, measured
// along angle (degrees: 0 horizontal, 90 vertical). The dimension line
// passes through location. The graphics are generated in an anonymous block.
func (d *Drawing) LinearDimension(p1, p2, location []float64, angle float64) (*entity.Dimension, error) {
	dim := entity.NewDimension(entity.DIM_LINEAR)
	dim.Angle = angle
	return d.addDimension(dim, func() {
		dim.DefPoint2 = point3(p1)
		dim.DefPoint3 = point3(p2)
		dim.DefPoint = point3(location)
	})
}

// AlignedDimension creates an aligned DIMENSION between p1 and p2 whose
// dimension line, parallel to p1-p2, passes through location.
func (d *Drawing) AlignedDimension(p1, p2, location []float64) (*entity.Dimension, error) {
	dim := entity.NewDimension(entity.DIM_ALIGNED)
	return d.addDimension(dim, func() {
		dim.DefPoint2 = point3(p1)
		dim.DefPoint3 = point3(p2)
		dim.DefPoint = point3(location)
	})
}

// RadiusDimension creates a radius DIMENSION from center to point on the arc.
func (d *Drawing) RadiusDimension(center, point []float64) (*entity.Dimension, error) {
	dim := entity.NewDimension(entity.DIM_RADIUS)
	return d.addDimension(dim, func() {
		dim.DefPoint = point3(center)
		dim.DefPoint4 = point3(point)
	})
}

// DiameterDimension creates a diameter DIMENSION of the circle with center,
// through point (and the opposite point).
func (d *Drawing) DiameterDimension(center, point []float64) (*entity.Dimension, error) {
	dim := entity.NewDimension(entity.DIM_DIAMETER)
	return d.addDimension(dim, func() {
		c, p := point3(center), point3(point)
		dim.DefPoint4 = p
		dim.DefPoint = []float64{2*c[0] - p[0], 2*c[1] - p[1], 2*c[2] - p[2]}
	})
}

// AngularDimension creates a 3-point angular DIMENSION of the angle at
// vertex between the lines to p1 and p2. The dimension arc passes through
// location, and the angle containing location is measured.
func (d *Drawing) AngularDimension(vertex, p1, p2, location []float64) (*entity.Dimension, error) {
	dim := entity.NewDimension(entity.DIM_ANGULAR3P)
	return d.addDimension(dim, func() {
		dim.DefPoint4 = point3(vertex)
		dim.DefPoint2 = point3(p1)
		dim.DefPoint3 = point3(p2)
		dim.DefPoint = point3(location)
	})
}

func point3(p []float64) []float64 {
	q := []float64{0, 0, 0}
	copy(q, p)
	return q
}

func (d *Drawing) addDimension(dim *entity.Dimension, set func()) (*entity.Dimension, error) {
	set()
	d.applyCurrent(dim)
	if err := d.RegenerateDimension(dim); err != nil {
		return nil, err
	}
	d.AddEntity(dim)
	return dim, nil
}

// RegenerateDimension (re)creates the graphics block of dim from its
// definition points, type, text and dimension style, and updates
// Measurement and TextMid. Call it after changing a dimension.
// Linear, aligned, radius, diameter and 3-point angular dimensions are supported.
func (d *Drawing) RegenerateDimension(dim *entity.Dimension) error {
	style, err := d.DimStyle(dim.Style)
	if err != nil {
		if style, err = d.DimStyle("Standard"); err != nil {
			style = table.NewDimStyle("Standard")
		}
	}
	g := &dimGraphics{d: d, style: style, z: dim.DefPoint[2]}
	switch dim.Type {
	case entity.DIM_LINEAR, entity.DIM_ALIGNED:
		err = g.linear(dim)
	case entity.DIM_RADIUS, entity.DIM_DIAMETER:
		err = g.radial(dim)
	case entity.DIM_ANGULAR3P:
		err = g.angular(dim)
	default:
		return fmt.Errorf("dimension type %d is not supported", dim.Type)
	}
	if err != nil {
		return err
	}
	b, err := d.Block(dim.BlockName)
	if dim.BlockName == "" || err != nil {
		if b, err = d.anonymousDimensionBlock(); err != nil {
			return err
		}
		dim.BlockName = b.Name
	}
	b.Entities = g.entities
	dim.Flags |= entity.DIM_BLOCK_OWNED
	return nil
}

// anonymousDimensionBlock creates a new anonymous block "*Dn".
func (d *Drawing) anonymousDimensionBlock() (*block.Block, error) {
	for n := 1; ; n++ {
		name := "*D" + strconv.Itoa(n)
		if _, err := d.Block(name); err == nil {
			continue
		}
		b, err := d.AddBlock(name, 0, 0, 0)
		if err != nil {
			return nil, err
		}
		b.Flag = 1 // anonymous
		return b, nil
	}
}

// dimGraphics builds the entities of a dimension block.
type dimGraphics struct {
	d        *Drawing
	style    *table.DimStyle
	z        float64
	entities entity.Entities
}

func (g *dimGraphics) size(v float64) float64 {
	s := g.style.Scale
	if s <= 0 {
		s = 1
	}
	return v * s
}

// byBlock sets the properties of e to BYBLOCK so that the block follows the DIMENSION.
func (g *dimGraphics) add(e entity.Entity) {
	e.SetLayer(g.d.Layers["0"])
	if c, ok := e.(entity.Colored); ok {
		c.SetColorByBlock()
	}
	if lp, ok := e.(entity.LineProperties); ok {
		if lt, err := g.d.LineType("ByBlock"); err == nil {
			lp.SetLineType(lt)
		}
		lp.SetLineWeight(table.LineWeightByBlock)
	}
	g.entities = g.entities.Add(e)
}

func (g *dimGraphics) line(p, q []float64) {
	l := entity.NewLine()
	l.Start = []float64{p[0], p[1], g.z}
	l.End = []float64{q[0], q[1], g.z}
	g.add(l)
}

// arrow adds a closed filled arrowhead with its tip at tip, pointing
// opposite to dir (dir points from the tip into the dimension line).
func (g *dimGraphics) arrow(tip, dir []float64) {
	size := g.size(g.style.ArrowSize)
	if size <= 0 {
		return
	}
	base := []float64{tip[0] + dir[0]*size, tip[1] + dir[1]*size}
	w := size / 6
	n := []float64{-dir[1] * w, dir[0] * w}
	s := entity.NewSolid()
	s.Points = [][]float64{
		{tip[0], tip[1], g.z},
		{base[0] + n[0], base[1] + n[1], g.z},
		{base[0] - n[0], base[1] - n[1], g.z},
		{base[0] - n[0], base[1] - n[1], g.z},
	}
	g.add(s)
}

// text adds the dimension text centered at pos and returns pos.
func (g *dimGraphics) text(dim *entity.Dimension, value string, pos []float64, rotation float64) {
	t := entity.NewMText()
	t.Coord = []float64{pos[0], pos[1], g.z}
	t.Height = g.size(g.style.TextHeight)
	t.Attachment = entity.MTEXT_MIDDLE_CENTER
	t.Rotation = rotation
	t.Value = value
	if st, ok := g.d.Styles["STANDARD"]; ok {
		t.Style = st
	} else {
		t.Style = g.d.CurrentStyle
	}
	g.add(t)
	dim.TextMid = []float64{pos[0], pos[1], g.z}
}

// textValue returns the text: the measurement for "" or "<>",
// otherwise dim.Text with "<>" replaced by the measurement.
func (g *dimGraphics) textValue(dim *entity.Dimension, measurement float64, prefix, suffix string) string {
	m := prefix + strconv.FormatFloat(measurement, 'f', g.style.Decimals, 64) + suffix
	if dim.Text == "" {
		return m
	}
	return strings.ReplaceAll(dim.Text, "<>", m)
}

// readable returns the angle (radians) of direction u, turned so that
// text along it is readable (in (-90, 90] degrees).
func readable(angle float64) float64 {
	for angle > math.Pi/2+1e-9 {
		angle -= math.Pi
	}
	for angle <= -math.Pi/2+1e-9 {
		angle += math.Pi
	}
	return angle
}

// textPosition returns the text position near mid for a dimension line
// in direction angle (radians), and the text rotation (degrees).
func (g *dimGraphics) textPosition(mid []float64, angle float64) ([]float64, float64) {
	a := readable(angle)
	rotation := a * 180 / math.Pi
	if g.style.TextInsideHorizontal {
		rotation = 0
	}
	if !g.style.TextAbove {
		return mid, rotation
	}
	off := g.size(g.style.TextGap) + g.size(g.style.TextHeight)/2
	up := []float64{-math.Sin(a), math.Cos(a)}
	return []float64{mid[0] + up[0]*off, mid[1] + up[1]*off}, rotation
}

func (g *dimGraphics) linear(dim *entity.Dimension) error {
	p1, p2, loc := dim.DefPoint2, dim.DefPoint3, dim.DefPoint
	var u []float64
	if dim.Type == entity.DIM_LINEAR {
		a := dim.Angle * math.Pi / 180
		u = []float64{math.Cos(a), math.Sin(a)}
	} else {
		dx, dy := p2[0]-p1[0], p2[1]-p1[1]
		l := math.Hypot(dx, dy)
		if l == 0 {
			return errors.New("aligned dimension: p1 and p2 are the same point")
		}
		u = []float64{dx / l, dy / l}
	}
	n := []float64{-u[1], u[0]}
	foot := func(p []float64) []float64 {
		t := (loc[0]-p[0])*n[0] + (loc[1]-p[1])*n[1]
		return []float64{p[0] + t*n[0], p[1] + t*n[1]}
	}
	d1, d2 := foot(p1), foot(p2)
	measurement := math.Abs((p2[0]-p1[0])*u[0] + (p2[1]-p1[1])*u[1])
	dim.Measurement = measurement
	dim.DefPoint = []float64{d2[0], d2[1], loc[2]}

	exo, exe := g.size(g.style.ExtOffset), g.size(g.style.ExtExtend)
	for _, pair := range [][2][]float64{{p1, d1}, {p2, d2}} {
		p, q := pair[0], pair[1]
		vx, vy := q[0]-p[0], q[1]-p[1]
		l := math.Hypot(vx, vy)
		if l < 1e-12 {
			continue
		}
		e := []float64{vx / l, vy / l}
		g.line([]float64{p[0] + e[0]*exo, p[1] + e[1]*exo}, []float64{q[0] + e[0]*exe, q[1] + e[1]*exe})
	}
	g.line(d1, d2)
	if l := math.Hypot(d2[0]-d1[0], d2[1]-d1[1]); l > 0 {
		dir := []float64{(d2[0] - d1[0]) / l, (d2[1] - d1[1]) / l}
		g.arrow(d1, dir)
		g.arrow(d2, []float64{-dir[0], -dir[1]})
	}
	mid := []float64{(d1[0] + d2[0]) / 2, (d1[1] + d2[1]) / 2}
	pos, rot := g.textPosition(mid, math.Atan2(u[1], u[0]))
	g.text(dim, g.textValue(dim, measurement, "", ""), pos, rot)
	return nil
}

func (g *dimGraphics) radial(dim *entity.Dimension) error {
	p := dim.DefPoint4
	var c, q []float64 // center, other end of the dimension line
	if dim.Type == entity.DIM_RADIUS {
		c = dim.DefPoint
		q = c
	} else {
		q = dim.DefPoint
		c = []float64{(p[0] + q[0]) / 2, (p[1] + q[1]) / 2}
	}
	r := math.Hypot(p[0]-c[0], p[1]-c[1])
	if r == 0 {
		return errors.New("radial dimension: radius is zero")
	}
	dir := []float64{(p[0] - c[0]) / r, (p[1] - c[1]) / r}
	g.line(q, p)
	g.arrow(p, []float64{-dir[0], -dir[1]})
	var measurement float64
	var prefix string
	var mid []float64
	if dim.Type == entity.DIM_RADIUS {
		measurement, prefix = r, "R"
		mid = []float64{(c[0] + p[0]) / 2, (c[1] + p[1]) / 2}
	} else {
		measurement, prefix = 2*r, "∅"
		g.arrow(q, dir)
		mid = c
	}
	dim.Measurement = measurement
	pos, rot := g.textPosition(mid, math.Atan2(dir[1], dir[0]))
	g.text(dim, g.textValue(dim, measurement, prefix, ""), pos, rot)
	return nil
}

func (g *dimGraphics) angular(dim *entity.Dimension) error {
	v, p1, p2, loc := dim.DefPoint4, dim.DefPoint2, dim.DefPoint3, dim.DefPoint
	r := math.Hypot(loc[0]-v[0], loc[1]-v[1])
	if r == 0 {
		return errors.New("angular dimension: location is at the vertex")
	}
	norm := func(a float64) float64 {
		a = math.Mod(a, 2*math.Pi)
		if a < 0 {
			a += 2 * math.Pi
		}
		return a
	}
	a1 := math.Atan2(p1[1]-v[1], p1[0]-v[0])
	a2 := math.Atan2(p2[1]-v[1], p2[0]-v[0])
	al := math.Atan2(loc[1]-v[1], loc[0]-v[0])
	sweep := norm(a2 - a1)
	if sweep == 0 {
		return errors.New("angular dimension: lines are parallel")
	}
	if norm(al-a1) > sweep { // measure the other angle, which contains location
		p1, p2 = p2, p1
		a1, a2 = a2, a1
		sweep = 2*math.Pi - sweep
	}
	_ = a2
	measurement := sweep * 180 / math.Pi
	dim.Measurement = measurement

	arc := entity.NewArc(nil)
	arc.Center = []float64{v[0], v[1], g.z}
	arc.Radius = r
	arc.Angle = []float64{a1 * 180 / math.Pi, (a1 + sweep) * 180 / math.Pi}
	g.add(arc)

	exo, exe := g.size(g.style.ExtOffset), g.size(g.style.ExtExtend)
	for _, pa := range []struct {
		p []float64
		a float64
	}{{p1, a1}, {p2, a1 + sweep}} {
		ri := math.Hypot(pa.p[0]-v[0], pa.p[1]-v[1])
		e := []float64{math.Cos(pa.a), math.Sin(pa.a)}
		var from, to float64
		switch {
		case r > ri+exo:
			from, to = ri+exo, r+exe
		case r < ri-exo:
			from, to = ri-exo, r-exe
		default:
			continue
		}
		g.line([]float64{v[0] + e[0]*from, v[1] + e[1]*from}, []float64{v[0] + e[0]*to, v[1] + e[1]*to})
	}
	start := []float64{v[0] + r*math.Cos(a1), v[1] + r*math.Sin(a1)}
	end := []float64{v[0] + r*math.Cos(a1+sweep), v[1] + r*math.Sin(a1+sweep)}
	g.arrow(start, []float64{-math.Sin(a1), math.Cos(a1)})
	g.arrow(end, []float64{math.Sin(a1 + sweep), -math.Cos(a1 + sweep)})

	am := a1 + sweep/2
	mid := []float64{v[0] + r*math.Cos(am), v[1] + r*math.Sin(am)}
	pos, rot := mid, readable(am-math.Pi/2)*180/math.Pi
	if g.style.TextAbove {
		off := g.size(g.style.TextGap) + g.size(g.style.TextHeight)/2
		pos = []float64{mid[0] + off*math.Cos(am), mid[1] + off*math.Sin(am)}
	}
	if g.style.TextInsideHorizontal {
		rot = 0
	}
	g.text(dim, g.textValue(dim, measurement, "", "°"), pos, rot)
	return nil
}
