package drawing

import (
	"errors"
	"math"

	"github.com/yofu/dxf/entity"
)

// Angular2LineDimension creates a 2-line angular DIMENSION of the angle
// between the lines l1a-l1b and l2a-l2b; the dimension arc passes through
// location, and the angle containing location is measured.
func (d *Drawing) Angular2LineDimension(l1a, l1b, l2a, l2b, location []float64) (*entity.Dimension, error) {
	dim := entity.NewDimension(entity.DIM_ANGULAR)
	return d.addDimension(dim, func() {
		dim.DefPoint2 = point3(l1a)
		dim.DefPoint3 = point3(l1b)
		dim.DefPoint4 = point3(l2a)
		dim.DefPoint = point3(l2b)
		dim.DefPoint5 = point3(location)
	})
}

// OrdinateDimension creates an ordinate DIMENSION of feature relative to
// origin: its X coordinate if xType, otherwise its Y coordinate; the
// leader ends at leaderEnd.
func (d *Drawing) OrdinateDimension(origin, feature, leaderEnd []float64, xType bool) (*entity.Dimension, error) {
	dim := entity.NewDimension(entity.DIM_ORDINATE)
	if xType {
		dim.Flags |= entity.DIM_ORDINATE_X
	}
	return d.addDimension(dim, func() {
		dim.DefPoint = point3(origin)
		dim.DefPoint2 = point3(feature)
		dim.DefPoint3 = point3(leaderEnd)
	})
}

// ArcDimension creates an arc length DIMENSION (ARC_DIMENSION) of the arc
// with center from p1 to p2 (counterclockwise); the dimension arc passes
// through location.
func (d *Drawing) ArcDimension(center, p1, p2, location []float64) (*entity.Dimension, error) {
	dim := entity.NewDimension(entity.DIM_ARC)
	return d.addDimension(dim, func() {
		dim.DefPoint4 = point3(center)
		dim.DefPoint2 = point3(p1)
		dim.DefPoint3 = point3(p2)
		dim.DefPoint = point3(location)
		dim.ArcLeader1, dim.ArcLeader2 = []float64{0, 0, 0}, []float64{0, 0, 0}
	})
}

// LinearDimensions creates continued linear DIMENSIONs between successive
// points on a common dimension line through location (angle in degrees).
func (d *Drawing) LinearDimensions(points [][]float64, location []float64, angle float64) ([]*entity.Dimension, error) {
	if len(points) < 2 {
		return nil, errors.New("continued dimensions need 2 or more points")
	}
	var dims []*entity.Dimension
	for i := 0; i+1 < len(points); i++ {
		dim, err := d.LinearDimension(points[i], points[i+1], location, angle)
		if err != nil {
			return dims, err
		}
		dims = append(dims, dim)
	}
	return dims, nil
}

// RadiusDimensionAt creates a radius DIMENSION of the circle with center
// and radius, at angle (degrees).
func (d *Drawing) RadiusDimensionAt(center []float64, radius, angle float64) (*entity.Dimension, error) {
	a := angle * math.Pi / 180
	c := point3(center)
	return d.RadiusDimension(c, []float64{c[0] + radius*math.Cos(a), c[1] + radius*math.Sin(a), c[2]})
}

// intersect returns the intersection of the lines a1-a2 and b1-b2 (XY).
func intersect(a1, a2, b1, b2 []float64) ([]float64, bool) {
	dx1, dy1 := a2[0]-a1[0], a2[1]-a1[1]
	dx2, dy2 := b2[0]-b1[0], b2[1]-b1[1]
	den := dx1*dy2 - dy1*dx2
	if math.Abs(den) < 1e-12*(math.Hypot(dx1, dy1)*math.Hypot(dx2, dy2)) {
		return nil, false
	}
	t := ((b1[0]-a1[0])*dy2 - (b1[1]-a1[1])*dx2) / den
	return []float64{a1[0] + t*dx1, a1[1] + t*dy1, a1[2]}, true
}

func farther(v, p, q []float64) []float64 {
	if math.Hypot(p[0]-v[0], p[1]-v[1]) >= math.Hypot(q[0]-v[0], q[1]-v[1]) {
		return p
	}
	return q
}

// angular2 draws a 2-line angular dimension: DefPoint2-DefPoint3 first
// line, DefPoint4-DefPoint second line, DefPoint5 on the dimension arc.
func (g *dimGraphics) angular2(dim *entity.Dimension) error {
	if len(dim.DefPoint5) < 2 {
		return errors.New("angular dimension: location (DefPoint5) required")
	}
	v, ok := intersect(dim.DefPoint2, dim.DefPoint3, dim.DefPoint4, dim.DefPoint)
	if !ok {
		return errors.New("angular dimension: lines are parallel")
	}
	p1 := farther(v, dim.DefPoint2, dim.DefPoint3)
	p2 := farther(v, dim.DefPoint4, dim.DefPoint)
	loc := dim.DefPoint5
	// the angle between the lines (not their extensions) containing loc
	al := math.Atan2(loc[1]-v[1], loc[0]-v[0])
	for _, flip := range [][2]bool{{false, false}, {true, false}, {false, true}, {true, true}} {
		q1, q2 := p1, p2
		if flip[0] {
			q1 = []float64{2*v[0] - p1[0], 2*v[1] - p1[1], v[2]}
		}
		if flip[1] {
			q2 = []float64{2*v[0] - p2[0], 2*v[1] - p2[1], v[2]}
		}
		a1 := math.Atan2(q1[1]-v[1], q1[0]-v[0])
		a2 := math.Atan2(q2[1]-v[1], q2[0]-v[0])
		sweep := math.Mod(a2-a1+4*math.Pi, 2*math.Pi)
		if sweep <= math.Pi && math.Mod(al-a1+4*math.Pi, 2*math.Pi) <= sweep {
			return g.angularArc(dim, v, q1, q2, loc)
		}
		if sweep > math.Pi && math.Mod(al-a2+4*math.Pi, 2*math.Pi) <= 2*math.Pi-sweep {
			return g.angularArc(dim, v, q2, q1, loc)
		}
	}
	return g.angularArc(dim, v, p1, p2, loc)
}

// ordinate draws an ordinate dimension: DefPoint origin, DefPoint2 feature,
// DefPoint3 leader end; the leader goes along the measured axis with a jog.
func (g *dimGraphics) ordinate(dim *entity.Dimension) error {
	o, f, e := dim.DefPoint, dim.DefPoint2, dim.DefPoint3
	xType := dim.Flags&entity.DIM_ORDINATE_X != 0
	// axis of the leader: perpendicular to the measured coordinate
	ax, ay := 0.0, 1.0 // X ordinate: vertical leader
	value := f[0] - o[0]
	if !xType {
		ax, ay = 1, 0
		value = f[1] - o[1]
	}
	along := (e[0]-f[0])*ax + (e[1]-f[1])*ay
	sign := 1.0
	if along < 0 {
		sign = -1
	}
	exo := g.size(g.style.ExtOffset)
	jog := 2 * g.size(g.style.ArrowSize)
	start := []float64{f[0] + sign*ax*exo, f[1] + sign*ay*exo}
	// the jog is in the last part of the leader
	cross := (e[0]-f[0])*ay - (e[1]-f[1])*ax
	if math.Abs(cross) < 1e-9 || math.Abs(along) <= exo+2*jog {
		g.line(start, []float64{e[0], e[1]})
	} else {
		k1 := along - sign*2*jog
		k2 := along - sign*jog
		p1 := []float64{f[0] + ax*k1, f[1] + ay*k1}
		p2 := []float64{e[0] - ax*sign*jog, e[1] - ay*sign*jog}
		_ = k2
		g.line(start, p1)
		g.line(p1, p2)
		g.line(p2, []float64{e[0], e[1]})
	}
	dim.Measurement = math.Abs(value)
	gap := g.size(g.style.TextGap)
	h := g.size(g.style.TextHeight)
	text := g.textValue(dim, math.Abs(value), "", "")
	w := float64(len([]rune(text))) * h * 0.7
	pos := []float64{e[0] + ax*sign*(gap+w/2), e[1] + ay*sign*(gap+h/2)}
	rot := 0.0
	if xType && !g.style.TextInsideHorizontal {
		rot = 90
		pos = []float64{e[0], e[1] + sign*(gap+w/2)}
	}
	g.text(dim, text, pos, rot)
	return nil
}

// arcLength draws an arc length dimension: DefPoint4 center, DefPoint2 and
// DefPoint3 ends of the arc (counterclockwise), DefPoint on the dimension arc.
func (g *dimGraphics) arcLength(dim *entity.Dimension) error {
	c, p1, p2, loc := dim.DefPoint4, dim.DefPoint2, dim.DefPoint3, dim.DefPoint
	r := math.Hypot(p1[0]-c[0], p1[1]-c[1])
	rd := math.Hypot(loc[0]-c[0], loc[1]-c[1])
	if r == 0 || rd == 0 {
		return errors.New("arc dimension: points at the center")
	}
	a1 := math.Atan2(p1[1]-c[1], p1[0]-c[0])
	a2 := math.Atan2(p2[1]-c[1], p2[0]-c[0])
	sweep := math.Mod(a2-a1+4*math.Pi, 2*math.Pi)
	if sweep == 0 {
		sweep = 2 * math.Pi
	}
	dim.ArcStart, dim.ArcEnd = a1, a1+sweep
	measurement := r * sweep
	dim.Measurement = measurement

	arc := entity.NewArc(nil)
	arc.Center = []float64{c[0], c[1], g.z}
	arc.Radius = rd
	arc.Angle = []float64{a1 * 180 / math.Pi, (a1 + sweep) * 180 / math.Pi}
	g.add(arc)
	exo, exe := g.size(g.style.ExtOffset), g.size(g.style.ExtExtend)
	for _, a := range []float64{a1, a1 + sweep} {
		e := []float64{math.Cos(a), math.Sin(a)}
		from, to := r+exo, rd+exe
		if rd < r {
			from, to = r-exo, rd-exe
		}
		g.line([]float64{c[0] + e[0]*from, c[1] + e[1]*from}, []float64{c[0] + e[0]*to, c[1] + e[1]*to})
	}
	start := []float64{c[0] + rd*math.Cos(a1), c[1] + rd*math.Sin(a1)}
	end := []float64{c[0] + rd*math.Cos(a1+sweep), c[1] + rd*math.Sin(a1+sweep)}
	g.arrow(start, []float64{-math.Sin(a1), math.Cos(a1)})
	g.arrow(end, []float64{math.Sin(a1 + sweep), -math.Cos(a1 + sweep)})
	am := a1 + sweep/2
	mid := []float64{c[0] + rd*math.Cos(am), c[1] + rd*math.Sin(am)}
	pos, rot := mid, readable(am-math.Pi/2)*180/math.Pi
	if g.style.TextAbove {
		off := g.size(g.style.TextGap) + g.size(g.style.TextHeight)/2
		pos = []float64{mid[0] + off*math.Cos(am), mid[1] + off*math.Sin(am)}
	}
	if g.style.TextInsideHorizontal {
		rot = 0
	}
	// the arc symbol precedes the text (DIMARCSYM 0)
	g.text(dim, g.textValue(dim, measurement, "\\U+2312", ""), pos, rot)
	return nil
}
