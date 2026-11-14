package entity

import (
	"math"

	"github.com/yofu/dxf/format"
)

// MPolygon represents MPOLYGON entities: polygons with holes and a fill
// (solid, pattern or gradient), like HATCH with polyline boundaries.
type MPolygon struct {
	*entity
	Version      int                 // 70: 1
	Elevation    float64             // 30 (10, 20 are 0)
	Extrusion    []float64           // 210, 220, 230
	PatternName  string              // 2
	Solid        bool                // 71: solid fill
	Boundaries   []*HatchBoundary    // 91, ...: polyline paths
	PatternType  int                 // 76: HATCH_PATTERN_*
	PatternAngle float64             // 52 (degrees)
	PatternScale float64             // 41
	Double       bool                // 77
	PatternLines []*HatchPatternLine // 78, ...
	FillColor    int                 // 63: color number of the solid fill (-1: not written; AC1018 or later)
	Offset       []float64           // 11, 21: offset vector
	Degenerated  int                 // 99: number of degenerate boundary paths
	Gradient     *HatchGradient      // 450, ... (AC1018 or later)
}

// IsEntity is for Entity interface.
func (m *MPolygon) IsEntity() bool { return true }

// NewMPolygon creates a solid MPOLYGON without boundaries.
func NewMPolygon() *MPolygon {
	return &MPolygon{
		entity:       NewEntity(MPOLYGON),
		Version:      1,
		Extrusion:    []float64{0, 0, 1},
		PatternName:  "SOLID",
		Solid:        true,
		PatternType:  HATCH_PATTERN_PREDEFINED,
		PatternScale: 1,
		FillColor:    -1,
		Offset:       []float64{0, 0},
	}
}

// AddPath adds a closed polyline boundary path (bulges may be nil). The
// first path is external, the following are holes.
func (m *MPolygon) AddPath(vertices [][]float64, bulges []float64) *HatchBoundary {
	flag := HATCH_PATH_POLYLINE
	if len(m.Boundaries) == 0 {
		flag |= HATCH_PATH_EXTERNAL
	}
	b := &HatchBoundary{Flag: flag, Vertices: vertices, Bulges: bulges, Closed: true}
	m.Boundaries = append(m.Boundaries, b)
	return b
}

// SetPattern sets a predefined pattern (see PredefinedPatterns).
func (m *MPolygon) SetPattern(name string, angle, scale float64) error {
	h := NewHatch()
	if err := h.SetPattern(name, angle, scale); err != nil {
		return err
	}
	m.PatternName, m.Solid, m.PatternType = h.PatternName, false, h.PatternType
	m.PatternAngle, m.PatternScale, m.Double, m.PatternLines = angle, scale, false, h.PatternLines
	m.Gradient = nil
	return nil
}

// SetSolid sets solid fill with color number aci (-1: BYLAYER, not written).
func (m *MPolygon) SetSolid(aci int) {
	m.PatternName, m.Solid, m.PatternType = "SOLID", true, HATCH_PATTERN_PREDEFINED
	m.PatternLines, m.Gradient = nil, nil
	m.FillColor = aci
}

// Format writes data to formatter.
func (m *MPolygon) Format(f format.Formatter) {
	m.entity.Format(f)
	f.WriteString(100, "AcDbMPolygon")
	f.WriteInt(70, m.Version)
	f.WriteFloat(10, 0)
	f.WriteFloat(20, 0)
	f.WriteFloat(30, m.Elevation)
	writePoint3(f, 210, m.Extrusion)
	f.WriteString(2, m.PatternName)
	f.WriteInt(71, boolInt(m.Solid))
	f.WriteInt(91, len(m.Boundaries))
	for _, b := range m.Boundaries {
		hasBulge := nonZero(b.Bulges...)
		f.WriteInt(92, b.Flag|HATCH_PATH_POLYLINE)
		f.WriteInt(73, boolInt(b.Closed))
		f.WriteInt(72, boolInt(hasBulge))
		f.WriteInt(93, len(b.Vertices))
		for i, v := range b.Vertices {
			f.WriteFloat(10, v[0])
			f.WriteFloat(20, v[1])
			if hasBulge {
				f.WriteFloat(42, coord(b.Bulges, i))
			}
		}
	}
	f.WriteInt(76, m.PatternType)
	if !m.Solid {
		f.WriteFloat(52, m.PatternAngle)
		f.WriteFloat(41, m.PatternScale)
		f.WriteInt(77, boolInt(m.Double))
		f.WriteInt(78, len(m.PatternLines))
		for _, l := range m.PatternLines {
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
	r2004 := format.AtLeast(f, format.R2004)
	if r2004 && m.FillColor >= 0 {
		f.WriteInt(63, m.FillColor)
	}
	f.WriteFloat(11, coord(m.Offset, 0))
	f.WriteFloat(21, coord(m.Offset, 1))
	f.WriteInt(99, m.Degenerated)
	if r2004 {
		if m.Gradient != nil {
			m.Gradient.format(f)
		} else { // no gradient (written by AutoCAD)
			f.WriteInt(450, 0)
			f.WriteInt(451, 0)
			f.WriteFloat(460, 0)
			f.WriteFloat(461, 0)
			f.WriteInt(452, 0)
			f.WriteFloat(462, 0)
			f.WriteInt(453, 0)
			f.WriteString(470, "LINEAR")
		}
	}
}

// String outputs data using default formatter.
func (m *MPolygon) String() string { return m.FormatString(format.NewASCII()) }

// FormatString outputs data using given formatter.
func (m *MPolygon) FormatString(f format.Formatter) string {
	m.Format(f)
	return f.Output()
}

// BBox returns the bounding box of the boundaries (in OCS).
func (m *MPolygon) BBox() ([]float64, []float64) {
	mins := []float64{math.Inf(1), math.Inf(1), m.Elevation}
	maxs := []float64{math.Inf(-1), math.Inf(-1), m.Elevation}
	for _, b := range m.Boundaries {
		for _, p := range b.points() {
			for i := 0; i < 2; i++ {
				mins[i] = math.Min(mins[i], p[i])
				maxs[i] = math.Max(maxs[i], p[i])
			}
		}
	}
	if math.IsInf(mins[0], 1) {
		return []float64{0, 0, m.Elevation}, []float64{0, 0, m.Elevation}
	}
	return mins, maxs
}

// ApplyTransform transforms the boundaries and the pattern like HATCH.
func (m *MPolygon) ApplyTransform(t Transform) error {
	h := &Hatch{entity: m.entity, Elevation: m.Elevation, Extrusion: m.Extrusion, Solid: m.Solid,
		Boundaries: m.Boundaries, PatternAngle: m.PatternAngle, PatternScale: m.PatternScale,
		PatternLines: m.PatternLines, Gradient: m.Gradient}
	if err := h.ApplyTransform(t); err != nil {
		return err
	}
	m.Elevation, m.PatternAngle, m.PatternScale = h.Elevation, h.PatternAngle, h.PatternScale
	if len(m.Offset) >= 2 {
		copy(m.Offset, t.Vector(m.Offset[:2]))
	}
	return nil
}
