package entity

import (
	"fmt"
	"math"
	"strings"

	"github.com/yofu/dxf/format"
)

// Boundary path type flags of HATCH (code 92).
const (
	HATCH_PATH_EXTERNAL  = 1
	HATCH_PATH_POLYLINE  = 2
	HATCH_PATH_DERIVED   = 4
	HATCH_PATH_TEXTBOX   = 8
	HATCH_PATH_OUTERMOST = 16
)

// Hatch styles (code 75).
const (
	HATCH_STYLE_NORMAL    = 0 // odd parity
	HATCH_STYLE_OUTERMOST = 1
	HATCH_STYLE_IGNORE    = 2
)

// Hatch pattern types (code 76).
const (
	HATCH_PATTERN_USER       = 0
	HATCH_PATTERN_PREDEFINED = 1
	HATCH_PATTERN_CUSTOM     = 2
)

// Edge types of HATCH boundary paths (code 72).
const (
	HATCH_EDGE_LINE    = 1
	HATCH_EDGE_ARC     = 2
	HATCH_EDGE_ELLIPSE = 3
	HATCH_EDGE_SPLINE  = 4
)

// HatchEdge is an edge of a HATCH boundary path (in OCS, 2D).
// It is one of *HatchLine, *HatchArc, *HatchEllipse and *HatchSpline.
type HatchEdge interface {
	EdgeType() int
	format(format.Formatter)
	points() [][]float64 // points bounding the edge, for BBox
}

// HatchLine is a line edge.
type HatchLine struct {
	Start []float64 // 10, 20
	End   []float64 // 11, 21
}

// EdgeType returns HATCH_EDGE_LINE.
func (e *HatchLine) EdgeType() int { return HATCH_EDGE_LINE }

func (e *HatchLine) format(f format.Formatter) {
	f.WriteFloat(10, e.Start[0])
	f.WriteFloat(20, e.Start[1])
	f.WriteFloat(11, e.End[0])
	f.WriteFloat(21, e.End[1])
}

func (e *HatchLine) points() [][]float64 { return [][]float64{e.Start, e.End} }

// HatchArc is a circular arc edge. Angles are in degrees.
type HatchArc struct {
	Center []float64 // 10, 20
	Radius float64   // 40
	Start  float64   // 50
	End    float64   // 51
	CCW    bool      // 73
}

// EdgeType returns HATCH_EDGE_ARC.
func (e *HatchArc) EdgeType() int { return HATCH_EDGE_ARC }

func (e *HatchArc) format(f format.Formatter) {
	f.WriteFloat(10, e.Center[0])
	f.WriteFloat(20, e.Center[1])
	f.WriteFloat(40, e.Radius)
	f.WriteFloat(50, e.Start)
	f.WriteFloat(51, e.End)
	f.WriteInt(73, boolInt(e.CCW))
}

func (e *HatchArc) points() [][]float64 {
	return [][]float64{
		{e.Center[0] - e.Radius, e.Center[1] - e.Radius},
		{e.Center[0] + e.Radius, e.Center[1] + e.Radius},
	}
}

// HatchEllipse is an elliptic arc edge. Angles are in degrees.
type HatchEllipse struct {
	Center    []float64 // 10, 20
	MajorAxis []float64 // 11, 21: endpoint of major axis relative to Center
	Ratio     float64   // 40: minor axis / major axis
	Start     float64   // 50
	End       float64   // 51
	CCW       bool      // 73
}

// EdgeType returns HATCH_EDGE_ELLIPSE.
func (e *HatchEllipse) EdgeType() int { return HATCH_EDGE_ELLIPSE }

func (e *HatchEllipse) format(f format.Formatter) {
	f.WriteFloat(10, e.Center[0])
	f.WriteFloat(20, e.Center[1])
	f.WriteFloat(11, e.MajorAxis[0])
	f.WriteFloat(21, e.MajorAxis[1])
	f.WriteFloat(40, e.Ratio)
	f.WriteFloat(50, e.Start)
	f.WriteFloat(51, e.End)
	f.WriteInt(73, boolInt(e.CCW))
}

func (e *HatchEllipse) points() [][]float64 {
	r := math.Hypot(e.MajorAxis[0], e.MajorAxis[1])
	return [][]float64{
		{e.Center[0] - r, e.Center[1] - r},
		{e.Center[0] + r, e.Center[1] + r},
	}
}

// HatchSpline is a spline edge.
type HatchSpline struct {
	Degree   int         // 94
	Rational bool        // 73
	Periodic bool        // 74
	Knots    []float64   // 95, 40
	Controls [][]float64 // 96, 10, 20
	Weights  []float64   // 42 (rational splines)
	Fits     [][]float64 // 97, 11, 21 (AC1024 or later; not written)
}

// EdgeType returns HATCH_EDGE_SPLINE.
func (e *HatchSpline) EdgeType() int { return HATCH_EDGE_SPLINE }

func (e *HatchSpline) format(f format.Formatter) {
	f.WriteInt(94, e.Degree)
	f.WriteInt(73, boolInt(e.Rational))
	f.WriteInt(74, boolInt(e.Periodic))
	f.WriteInt(95, len(e.Knots))
	f.WriteInt(96, len(e.Controls))
	for _, k := range e.Knots {
		f.WriteFloat(40, k)
	}
	for i, c := range e.Controls {
		f.WriteFloat(10, c[0])
		f.WriteFloat(20, c[1])
		if e.Rational && i < len(e.Weights) {
			f.WriteFloat(42, e.Weights[i])
		}
	}
}

func (e *HatchSpline) points() [][]float64 { return e.Controls }

// HatchBoundary is a boundary path of HATCH.
// If Flag has HATCH_PATH_POLYLINE, Vertices/Bulges/Closed are used,
// otherwise Edges.
type HatchBoundary struct {
	Flag     int         // 92
	Vertices [][]float64 // 10, 20 (polyline)
	Bulges   []float64   // 42 (polyline)
	Closed   bool        // 73 (polyline)
	Edges    []HatchEdge // 72 ... (edges)
}

// IsPolyline reports whether the path is a polyline path.
func (b *HatchBoundary) IsPolyline() bool {
	return b.Flag&HATCH_PATH_POLYLINE != 0
}

func (b *HatchBoundary) format(f format.Formatter) {
	f.WriteInt(92, b.Flag)
	if b.IsPolyline() {
		hasBulge := false
		for _, v := range b.Bulges {
			if v != 0 {
				hasBulge = true
			}
		}
		f.WriteInt(72, boolInt(hasBulge))
		f.WriteInt(73, boolInt(b.Closed))
		f.WriteInt(93, len(b.Vertices))
		for i, v := range b.Vertices {
			f.WriteFloat(10, v[0])
			f.WriteFloat(20, v[1])
			if hasBulge {
				bulge := 0.0
				if i < len(b.Bulges) {
					bulge = b.Bulges[i]
				}
				f.WriteFloat(42, bulge)
			}
		}
	} else {
		f.WriteInt(93, len(b.Edges))
		for _, e := range b.Edges {
			f.WriteInt(72, e.EdgeType())
			e.format(f)
		}
	}
	f.WriteInt(97, 0) // number of source boundary objects (not associative)
}

func (b *HatchBoundary) points() [][]float64 {
	if b.IsPolyline() {
		pts := append([][]float64{}, b.Vertices...)
		// bulged segments may extend beyond the vertices
		for i := range b.Vertices {
			if i < len(b.Bulges) && b.Bulges[i] != 0 && len(b.Vertices) > 1 {
				j := (i + 1) % len(b.Vertices)
				if c, r := bulge(b.Bulges[i], b.Vertices[i], b.Vertices[j]); c != nil {
					pts = append(pts, []float64{c[0] - r, c[1] - r}, []float64{c[0] + r, c[1] + r})
				}
			}
		}
		return pts
	}
	var pts [][]float64
	for _, e := range b.Edges {
		pts = append(pts, e.points()...)
	}
	return pts
}

// HatchPatternLine is a pattern definition line of HATCH, as stored in DXF:
// the angle includes the hatch angle, and base point, offset and dashes
// are rotated and scaled.
type HatchPatternLine struct {
	Angle  float64   // 53 (degrees)
	Base   []float64 // 43, 44
	Offset []float64 // 45, 46
	Dashes []float64 // 79, 49
}

// Hatch represents HATCH Entity.
type Hatch struct {
	*entity
	Elevation    float64             // 30 (10, 20 are 0)
	Extrusion    []float64           // 210, 220, 230
	PatternName  string              // 2
	Solid        bool                // 70
	Associative  bool                // 71 (source boundary objects are not written)
	Boundaries   []*HatchBoundary    // 91, ...
	Style        int                 // 75: HATCH_STYLE_*
	PatternType  int                 // 76: HATCH_PATTERN_*
	PatternAngle float64             // 52 (degrees)
	PatternScale float64             // 41
	Double       bool                // 77
	PatternLines []*HatchPatternLine // 78, ...
	Seeds        [][]float64         // 98, 10, 20
}

// IsEntity is for Entity interface.
func (h *Hatch) IsEntity() bool {
	return true
}

// NewHatch creates a new solid-filled Hatch without boundaries.
func NewHatch() *Hatch {
	return &Hatch{
		entity:       NewEntity(HATCH),
		Extrusion:    []float64{0.0, 0.0, 1.0},
		PatternName:  "SOLID",
		Solid:        true,
		Style:        HATCH_STYLE_NORMAL,
		PatternType:  HATCH_PATTERN_PREDEFINED,
		PatternScale: 1.0,
	}
}

// AddPolylinePath adds a closed polyline boundary path.
// vertices are 2D points (OCS); bulges may be nil.
// The first path is flagged as external; following paths (e.g. holes) are not.
func (h *Hatch) AddPolylinePath(vertices [][]float64, bulges []float64) *HatchBoundary {
	b := &HatchBoundary{
		Flag:     h.externalFlag() | HATCH_PATH_POLYLINE,
		Vertices: vertices,
		Bulges:   bulges,
		Closed:   true,
	}
	h.Boundaries = append(h.Boundaries, b)
	return b
}

// AddEdgePath adds a boundary path made of edges.
// The first path is flagged as external; following paths (e.g. holes) are not.
func (h *Hatch) AddEdgePath(edges ...HatchEdge) *HatchBoundary {
	b := &HatchBoundary{
		Flag:  h.externalFlag(),
		Edges: edges,
	}
	h.Boundaries = append(h.Boundaries, b)
	return b
}

func (h *Hatch) externalFlag() int {
	if len(h.Boundaries) == 0 {
		return HATCH_PATH_EXTERNAL
	}
	return 0
}

// SetSolid sets solid fill.
func (h *Hatch) SetSolid() {
	h.PatternName = "SOLID"
	h.Solid = true
	h.PatternType = HATCH_PATTERN_PREDEFINED
	h.PatternAngle = 0
	h.PatternScale = 1
	h.Double = false
	h.PatternLines = nil
}

// predefined patterns (acadiso.pat): angle, base x, base y, offset x, offset y, dashes...
var predefinedPatterns = map[string][][]float64{
	"ANSI31": {{45, 0, 0, 0, 3.175}},
	"ANSI37": {{45, 0, 0, 0, 3.175}, {135, 0, 0, 0, 3.175}},
	"LINE":   {{0, 0, 0, 0, 3.175}},
	"NET":    {{0, 0, 0, 0, 3.175}, {90, 0, 0, 0, 3.175}},
}

// PredefinedPatterns returns the names of the predefined patterns
// available for SetPattern.
func PredefinedPatterns() []string {
	names := make([]string, 0, len(predefinedPatterns))
	for k := range predefinedPatterns {
		names = append(names, k)
	}
	return names
}

// SetPattern sets a predefined pattern (see PredefinedPatterns) with
// angle (degrees) and scale.
func (h *Hatch) SetPattern(name string, angle, scale float64) error {
	def, ok := predefinedPatterns[strings.ToUpper(name)]
	if !ok {
		return fmt.Errorf("unknown hatch pattern: %s", name)
	}
	h.PatternName = strings.ToUpper(name)
	h.Solid = false
	h.PatternType = HATCH_PATTERN_PREDEFINED
	h.PatternAngle = angle
	h.PatternScale = scale
	h.Double = false
	h.PatternLines = patternLines(def, angle, scale)
	return nil
}

// SetUserPattern sets a user-defined pattern of parallel lines with
// angle (degrees) and spacing; double adds crossing lines at 90 degrees.
func (h *Hatch) SetUserPattern(angle, spacing float64, double bool) {
	h.PatternName = "_USER"
	h.Solid = false
	h.PatternType = HATCH_PATTERN_USER
	h.PatternAngle = angle
	h.PatternScale = spacing
	h.Double = double
	def := [][]float64{{0, 0, 0, 0, 1}}
	if double {
		def = append(def, []float64{90, 0, 0, 0, 1})
	}
	h.PatternLines = patternLines(def, angle, spacing)
}

// patternLines transforms pattern definitions (.pat format: offset in the
// line's coordinate system) to DXF pattern lines (offset in OCS).
func patternLines(def [][]float64, angle, scale float64) []*HatchPatternLine {
	lines := make([]*HatchPatternLine, 0, len(def))
	for _, d := range def {
		total := d[0] + angle
		base := rotate([]float64{d[1] * scale, d[2] * scale}, angle)
		offset := rotate([]float64{d[3] * scale, d[4] * scale}, total)
		dashes := make([]float64, 0, len(d)-5)
		for _, v := range d[5:] {
			dashes = append(dashes, v*scale)
		}
		lines = append(lines, &HatchPatternLine{Angle: total, Base: base, Offset: offset, Dashes: dashes})
	}
	return lines
}

func rotate(p []float64, deg float64) []float64 {
	rad := deg * math.Pi / 180.0
	s, c := math.Sin(rad), math.Cos(rad)
	return []float64{p[0]*c - p[1]*s, p[0]*s + p[1]*c}
}

// Format writes data to formatter.
func (h *Hatch) Format(f format.Formatter) {
	h.entity.Format(f)
	f.WriteString(100, "AcDbHatch")
	f.WriteFloat(10, 0.0)
	f.WriteFloat(20, 0.0)
	f.WriteFloat(30, h.Elevation)
	for i := 0; i < 3; i++ {
		f.WriteFloat(200+(i+1)*10, h.Extrusion[i])
	}
	f.WriteString(2, h.PatternName)
	f.WriteInt(70, boolInt(h.Solid))
	f.WriteInt(71, boolInt(h.Associative))
	f.WriteInt(91, len(h.Boundaries))
	for _, b := range h.Boundaries {
		b.format(f)
	}
	f.WriteInt(75, h.Style)
	f.WriteInt(76, h.PatternType)
	if !h.Solid {
		f.WriteFloat(52, h.PatternAngle)
		f.WriteFloat(41, h.PatternScale)
		f.WriteInt(77, boolInt(h.Double))
		f.WriteInt(78, len(h.PatternLines))
		for _, l := range h.PatternLines {
			f.WriteFloat(53, l.Angle)
			f.WriteFloat(43, l.Base[0])
			f.WriteFloat(44, l.Base[1])
			f.WriteFloat(45, l.Offset[0])
			f.WriteFloat(46, l.Offset[1])
			f.WriteInt(79, len(l.Dashes))
			for _, d := range l.Dashes {
				f.WriteFloat(49, d)
			}
		}
	}
	f.WriteInt(98, len(h.Seeds))
	for _, s := range h.Seeds {
		f.WriteFloat(10, s[0])
		f.WriteFloat(20, s[1])
	}
}

// String outputs data using default formatter.
func (h *Hatch) String() string {
	f := format.NewASCII()
	return h.FormatString(f)
}

// FormatString outputs data using given formatter.
func (h *Hatch) FormatString(f format.Formatter) string {
	h.Format(f)
	return f.Output()
}

// BBox returns the bounding box of the boundary paths (in OCS; arcs are
// bounded by their full circles).
func (h *Hatch) BBox() ([]float64, []float64) {
	mins := []float64{math.Inf(1), math.Inf(1), h.Elevation}
	maxs := []float64{math.Inf(-1), math.Inf(-1), h.Elevation}
	for _, b := range h.Boundaries {
		for _, p := range b.points() {
			for i := 0; i < 2; i++ {
				mins[i] = math.Min(mins[i], p[i])
				maxs[i] = math.Max(maxs[i], p[i])
			}
		}
	}
	if math.IsInf(mins[0], 1) {
		return []float64{0, 0, h.Elevation}, []float64{0, 0, h.Elevation}
	}
	return mins, maxs
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
