package dxf

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// groupReader reads group code/value pairs sequentially.
// HATCH needs it because the same group codes appear in different contexts.
type groupReader struct {
	data [][2]string
	pos  int
}

func (r *groupReader) done() bool { return r.pos >= len(r.data) }

func (r *groupReader) peek() string {
	if r.done() {
		return ""
	}
	return r.data[r.pos][0]
}

// peekAt returns the group code at pos+n.
func (r *groupReader) peekAt(n int) string {
	if r.pos+n >= len(r.data) {
		return ""
	}
	return r.data[r.pos+n][0]
}

func (r *groupReader) next() [2]string {
	dt := r.data[r.pos]
	r.pos++
	return dt
}

// float reads a float value if the next code is code; otherwise it returns def.
func (r *groupReader) float(code string, def float64) (float64, error) {
	if r.peek() != code {
		return def, nil
	}
	dt := r.next()
	v, err := strconv.ParseFloat(strings.TrimSpace(dt[1]), 64)
	if err != nil {
		return def, fmt.Errorf("code %s: %w", dt[0], err)
	}
	return v, nil
}

// int reads an int value if the next code is code; otherwise it returns def.
func (r *groupReader) int(code string, def int) (int, error) {
	if r.peek() != code {
		return def, nil
	}
	dt := r.next()
	v, err := strconv.Atoi(strings.TrimSpace(dt[1]))
	if err != nil {
		return def, fmt.Errorf("code %s: %w", dt[0], err)
	}
	return v, nil
}

// point2 reads a 2D point (xcode, ycode).
func (r *groupReader) point2(xcode, ycode string) ([]float64, error) {
	x, err := r.float(xcode, 0)
	if err != nil {
		return nil, err
	}
	y, err := r.float(ycode, 0)
	if err != nil {
		return nil, err
	}
	return []float64{x, y}, nil
}

// ParseHatch parses HATCH entities.
// Source boundary objects (330 in paths) are skipped.
func ParseHatch(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	h := entity.NewHatch()
	h.PatternName = ""
	h.Solid = false
	r := &groupReader{data: data}
	afterPaths := false
	for !r.done() {
		var err error
		switch r.peek() {
		case "8":
			if layer, lerr := d.Layer(r.next()[1], false); lerr == nil {
				h.SetLayer(layer)
			}
		case "48":
			var v float64
			if v, err = r.float("48", 1); err == nil {
				h.SetLtscale(v)
			}
		case "30":
			if afterPaths {
				r.next()
			} else {
				h.Elevation, err = r.float("30", 0)
			}
		case "210":
			h.Extrusion[0], err = r.float("210", 0)
		case "220":
			h.Extrusion[1], err = r.float("220", 0)
		case "230":
			h.Extrusion[2], err = r.float("230", 1)
		case "2":
			h.PatternName = r.next()[1]
		case "70":
			var v int
			v, err = r.int("70", 0)
			h.Solid = v == 1
		case "71":
			var v int
			v, err = r.int("71", 0)
			h.Associative = v == 1
		case "91":
			var n int
			if n, err = r.int("91", 0); err != nil {
				break
			}
			for i := 0; i < n && err == nil; i++ {
				var b *entity.HatchBoundary
				if b, err = parseHatchBoundary(r); err == nil {
					h.Boundaries = append(h.Boundaries, b)
				}
			}
			afterPaths = true
		case "75":
			h.Style, err = r.int("75", 0)
		case "76":
			h.PatternType, err = r.int("76", 0)
		case "52":
			h.PatternAngle, err = r.float("52", 0)
		case "41":
			h.PatternScale, err = r.float("41", 1)
		case "77":
			var v int
			v, err = r.int("77", 0)
			h.Double = v == 1
		case "78":
			var n int
			if n, err = r.int("78", 0); err != nil {
				break
			}
			for i := 0; i < n && err == nil; i++ {
				var l *entity.HatchPatternLine
				if l, err = parseHatchPatternLine(r); err == nil {
					h.PatternLines = append(h.PatternLines, l)
				}
			}
		case "98":
			var n int
			if n, err = r.int("98", 0); err != nil {
				break
			}
			for i := 0; i < n && err == nil; i++ {
				var p []float64
				if p, err = r.point2("10", "20"); err == nil {
					h.Seeds = append(h.Seeds, p)
				}
			}
		case "450":
			h.Gradient, err = parseHatchGradient(r)
		default:
			r.next()
		}
		if err != nil {
			return h, err
		}
	}
	return h, nil
}

// parseHatchGradient parses gradient data starting with 450.
// It returns nil if 450 is 0 (not a gradient fill).
func parseHatchGradient(r *groupReader) (*entity.HatchGradient, error) {
	kind, err := r.int("450", 0)
	if err != nil {
		return nil, err
	}
	g := &entity.HatchGradient{}
	for !r.done() && err == nil {
		switch r.peek() {
		case "451":
			r.next()
		case "452":
			var v int
			v, err = r.int("452", 0)
			g.OneColor = v == 1
		case "453":
			var n int
			if n, err = r.int("453", 0); err != nil {
				break
			}
			for i := 0; i < n && err == nil; i++ {
				c := entity.GradientColor{}
				if c.Value, err = r.float("463", 0); err != nil {
					break
				}
				if c.ACI, err = r.int("63", 0); err != nil {
					break
				}
				c.RGB, err = r.int("421", 0)
				g.Colors = append(g.Colors, c)
			}
		case "460":
			var v float64
			v, err = r.float("460", 0)
			g.Rotation = v * 180.0 / math.Pi
		case "461":
			g.Centered, err = r.float("461", 0)
		case "462":
			g.Tint, err = r.float("462", 0)
		case "470":
			g.Name = r.next()[1]
		default:
			if kind != 1 {
				return nil, nil
			}
			return g, nil
		}
	}
	if err != nil || kind != 1 {
		return nil, err
	}
	return g, nil
}

func parseHatchBoundary(r *groupReader) (*entity.HatchBoundary, error) {
	if r.peek() != "92" {
		return nil, fmt.Errorf("HATCH: boundary path should start with 92, got %s", r.peek())
	}
	flag, err := r.int("92", 0)
	if err != nil {
		return nil, err
	}
	b := &entity.HatchBoundary{Flag: flag}
	if b.IsPolyline() {
		hasBulge, err := r.int("72", 0)
		if err != nil {
			return nil, err
		}
		closed, err := r.int("73", 0)
		if err != nil {
			return nil, err
		}
		b.Closed = closed == 1
		n, err := r.int("93", 0)
		if err != nil {
			return nil, err
		}
		for i := 0; i < n; i++ {
			p, err := r.point2("10", "20")
			if err != nil {
				return nil, err
			}
			bulge := 0.0
			if hasBulge == 1 || r.peek() == "42" {
				if bulge, err = r.float("42", 0); err != nil {
					return nil, err
				}
			}
			b.Vertices = append(b.Vertices, p)
			b.Bulges = append(b.Bulges, bulge)
		}
	} else {
		n, err := r.int("93", 0)
		if err != nil {
			return nil, err
		}
		for i := 0; i < n; i++ {
			e, err := parseHatchEdge(r)
			if err != nil {
				return nil, err
			}
			b.Edges = append(b.Edges, e)
		}
	}
	// source boundary objects
	n, err := r.int("97", 0)
	if err != nil {
		return nil, err
	}
	for i := 0; i < n && r.peek() == "330"; i++ {
		r.next()
	}
	return b, nil
}

func parseHatchEdge(r *groupReader) (entity.HatchEdge, error) {
	typ, err := r.int("72", 0)
	if err != nil {
		return nil, err
	}
	switch typ {
	case entity.HATCH_EDGE_LINE:
		s, err := r.point2("10", "20")
		if err != nil {
			return nil, err
		}
		e, err := r.point2("11", "21")
		if err != nil {
			return nil, err
		}
		return &entity.HatchLine{Start: s, End: e}, nil
	case entity.HATCH_EDGE_ARC:
		a := &entity.HatchArc{}
		if a.Center, err = r.point2("10", "20"); err != nil {
			return nil, err
		}
		if a.Radius, err = r.float("40", 0); err != nil {
			return nil, err
		}
		if a.Start, err = r.float("50", 0); err != nil {
			return nil, err
		}
		if a.End, err = r.float("51", 360); err != nil {
			return nil, err
		}
		ccw, err := r.int("73", 1)
		a.CCW = ccw == 1
		return a, err
	case entity.HATCH_EDGE_ELLIPSE:
		e := &entity.HatchEllipse{}
		if e.Center, err = r.point2("10", "20"); err != nil {
			return nil, err
		}
		if e.MajorAxis, err = r.point2("11", "21"); err != nil {
			return nil, err
		}
		if e.Ratio, err = r.float("40", 1); err != nil {
			return nil, err
		}
		if e.Start, err = r.float("50", 0); err != nil {
			return nil, err
		}
		if e.End, err = r.float("51", 360); err != nil {
			return nil, err
		}
		ccw, err := r.int("73", 1)
		e.CCW = ccw == 1
		return e, err
	case entity.HATCH_EDGE_SPLINE:
		return parseHatchSpline(r)
	default:
		return nil, fmt.Errorf("HATCH: unknown edge type %d", typ)
	}
}

func parseHatchSpline(r *groupReader) (*entity.HatchSpline, error) {
	s := &entity.HatchSpline{}
	var err error
	if s.Degree, err = r.int("94", 3); err != nil {
		return nil, err
	}
	rational, err := r.int("73", 0)
	if err != nil {
		return nil, err
	}
	s.Rational = rational == 1
	periodic, err := r.int("74", 0)
	if err != nil {
		return nil, err
	}
	s.Periodic = periodic == 1
	nk, err := r.int("95", 0)
	if err != nil {
		return nil, err
	}
	nc, err := r.int("96", 0)
	if err != nil {
		return nil, err
	}
	for i := 0; i < nk; i++ {
		k, err := r.float("40", 0)
		if err != nil {
			return nil, err
		}
		s.Knots = append(s.Knots, k)
	}
	for i := 0; i < nc; i++ {
		p, err := r.point2("10", "20")
		if err != nil {
			return nil, err
		}
		s.Controls = append(s.Controls, p)
		if r.peek() == "42" {
			w, err := r.float("42", 1)
			if err != nil {
				return nil, err
			}
			s.Weights = append(s.Weights, w)
		}
	}
	// weights may also follow all control points
	for r.peek() == "42" {
		w, err := r.float("42", 1)
		if err != nil {
			return nil, err
		}
		s.Weights = append(s.Weights, w)
	}
	// AC1024 or later: fit data (97 count, 11/21 points, 12/22 and 13/23 tangents).
	// It is distinguished from the source boundary count (97) of the path by what follows.
	if r.peek() == "97" {
		after := r.peekAt(1)
		if after == "11" || after == "12" || after == "97" {
			nf, err := r.int("97", 0)
			if err != nil {
				return nil, err
			}
			for i := 0; i < nf; i++ {
				p, err := r.point2("11", "21")
				if err != nil {
					return nil, err
				}
				s.Fits = append(s.Fits, p)
			}
			for _, c := range [][2]string{{"12", "22"}, {"13", "23"}} {
				if r.peek() == c[0] {
					if _, err := r.point2(c[0], c[1]); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return s, nil
}

func parseHatchPatternLine(r *groupReader) (*entity.HatchPatternLine, error) {
	l := &entity.HatchPatternLine{}
	var err error
	if l.Angle, err = r.float("53", 0); err != nil {
		return nil, err
	}
	if l.Base, err = r.point2("43", "44"); err != nil {
		return nil, err
	}
	if l.Offset, err = r.point2("45", "46"); err != nil {
		return nil, err
	}
	n, err := r.int("79", 0)
	if err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		v, err := r.float("49", 0)
		if err != nil {
			return nil, err
		}
		l.Dashes = append(l.Dashes, v)
	}
	return l, nil
}
