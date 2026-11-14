package entity

import (
	"math"

	"github.com/yofu/dxf/format"
)

// Dimension types (code 70, lower bits).
const (
	DIM_LINEAR      = 0 // rotated, horizontal or vertical
	DIM_ALIGNED     = 1
	DIM_ANGULAR     = 2 // 2 lines
	DIM_DIAMETER    = 3
	DIM_RADIUS      = 4
	DIM_ANGULAR3P   = 5 // 3 points
	DIM_ORDINATE    = 6
	DIM_BLOCK_OWNED = 32  // the block is referenced by this dimension only
	DIM_ORDINATE_X  = 64  // ordinate type is X (otherwise Y)
	DIM_USER_TEXT   = 128 // text is at a user-defined location
)

// Dimension represents DIMENSION Entity.
//
// The meaning of the definition points depends on Type
// (DXF reference, DIMENSION group codes):
//
//	linear, aligned: DefPoint (10) on the dimension line at the second extension line,
//	                 DefPoint2 (13) / DefPoint3 (14) origins of extension lines
//	radius:          DefPoint (10) center, DefPoint4 (15) point on the curve
//	diameter:        DefPoint (10) and DefPoint4 (15) opposite points on the curve
//	angular 3-point: DefPoint (10) on the dimension arc, DefPoint2 (13) / DefPoint3 (14)
//	                 points on the extension lines, DefPoint4 (15) vertex
//	angular 2-line:  DefPoint2 (13) - DefPoint3 (14) first line, DefPoint4 (15) - DefPoint (10)
//	                 second line, DefPoint5 (16) on the dimension arc
//	ordinate:        DefPoint (10) origin, DefPoint2 (13) feature, DefPoint3 (14) leader end
//
// The graphics are in the block BlockName (an anonymous block "*Dn").
type Dimension struct {
	*entity
	BlockName    string    // 2
	DefPoint     []float64 // 10, 20, 30 (WCS)
	TextMid      []float64 // 11, 21, 31: middle point of the text (OCS)
	Type         int       // 70 & 7: DIM_LINEAR, ...
	Flags        int       // 70 &^ 7: DIM_BLOCK_OWNED, DIM_ORDINATE_X, DIM_USER_TEXT
	Attachment   int       // 71: attachment point of the text (MTEXT_*)
	Measurement  float64   // 42: actual measurement
	Text         string    // 1: "" or "<>" is the measurement; "<>" in text is replaced
	TextRotation float64   // 53 (degrees)
	Extrusion    []float64 // 210, 220, 230
	Style        string    // 3: dimension style name
	DefPoint2    []float64 // 13, 23, 33 (WCS)
	DefPoint3    []float64 // 14, 24, 34 (WCS)
	DefPoint4    []float64 // 15, 25, 35 (WCS)
	DefPoint5    []float64 // 16, 26, 36 (OCS)
	Angle        float64   // 50: rotation of linear dimensions (degrees)
	Oblique      float64   // 52: oblique angle of extension lines (degrees)
	LeaderLength float64   // 40: radius/diameter
	// HorizontalDirection (51) is the angle of the horizontal direction
	// of the dimension (degrees), written if not 0.
	HorizontalDirection float64
	ClonePoint          []float64 // 12, 22, 32: insertion point for clones (nil if not given)
}

// IsEntity is for Entity interface.
func (d *Dimension) IsEntity() bool {
	return true
}

// NewDimension creates a new Dimension of type typ.
func NewDimension(typ int) *Dimension {
	return &Dimension{
		entity:     NewEntity(DIMENSION),
		DefPoint:   []float64{0.0, 0.0, 0.0},
		TextMid:    []float64{0.0, 0.0, 0.0},
		Type:       typ,
		Flags:      DIM_BLOCK_OWNED,
		Attachment: MTEXT_MIDDLE_CENTER,
		Extrusion:  []float64{0.0, 0.0, 1.0},
		Style:      "Standard",
		DefPoint2:  []float64{0.0, 0.0, 0.0},
		DefPoint3:  []float64{0.0, 0.0, 0.0},
		DefPoint4:  []float64{0.0, 0.0, 0.0},
		DefPoint5:  []float64{0.0, 0.0, 0.0},
	}
}

func writePoint(f format.Formatter, code int, p []float64) {
	for i := 0; i < 3; i++ {
		f.WriteFloat(code+i*10, p[i])
	}
}

// Format writes data to formatter.
func (d *Dimension) Format(f format.Formatter) {
	d.entity.Format(f)
	f.WriteString(100, "AcDbDimension")
	f.WriteString(2, d.BlockName)
	writePoint(f, 10, d.DefPoint)
	writePoint(f, 11, d.TextMid)
	f.WriteInt(70, d.Type|d.Flags)
	f.WriteInt(71, d.Attachment)
	f.WriteFloat(42, d.Measurement)
	f.WriteString(1, d.Text)
	if d.TextRotation != 0 {
		f.WriteFloat(53, d.TextRotation)
	}
	if d.HorizontalDirection != 0 {
		f.WriteFloat(51, d.HorizontalDirection)
	}
	if d.Extrusion[0] != 0 || d.Extrusion[1] != 0 || d.Extrusion[2] != 1 {
		writePoint(f, 210, d.Extrusion)
	}
	f.WriteString(3, d.Style)
	switch d.Type {
	case DIM_LINEAR, DIM_ALIGNED:
		f.WriteString(100, "AcDbAlignedDimension")
		if len(d.ClonePoint) >= 3 {
			writePoint(f, 12, d.ClonePoint)
		}
		writePoint(f, 13, d.DefPoint2)
		writePoint(f, 14, d.DefPoint3)
		if d.Type == DIM_LINEAR {
			f.WriteFloat(50, d.Angle)
		}
		if d.Oblique != 0 {
			f.WriteFloat(52, d.Oblique)
		}
		if d.Type == DIM_LINEAR {
			f.WriteString(100, "AcDbRotatedDimension")
		}
	case DIM_ANGULAR:
		f.WriteString(100, "AcDb2LineAngularDimension")
		writePoint(f, 13, d.DefPoint2)
		writePoint(f, 14, d.DefPoint3)
		writePoint(f, 15, d.DefPoint4)
		writePoint(f, 16, d.DefPoint5)
	case DIM_DIAMETER:
		f.WriteString(100, "AcDbDiametricDimension")
		writePoint(f, 15, d.DefPoint4)
		f.WriteFloat(40, d.LeaderLength)
	case DIM_RADIUS:
		f.WriteString(100, "AcDbRadialDimension")
		writePoint(f, 15, d.DefPoint4)
		f.WriteFloat(40, d.LeaderLength)
	case DIM_ANGULAR3P:
		f.WriteString(100, "AcDb3PointAngularDimension")
		writePoint(f, 13, d.DefPoint2)
		writePoint(f, 14, d.DefPoint3)
		writePoint(f, 15, d.DefPoint4)
	case DIM_ORDINATE:
		f.WriteString(100, "AcDbOrdinateDimension")
		writePoint(f, 13, d.DefPoint2)
		writePoint(f, 14, d.DefPoint3)
	}
}

// String outputs data using default formatter.
func (d *Dimension) String() string {
	f := format.NewASCII()
	return d.FormatString(f)
}

// FormatString outputs data using given formatter.
func (d *Dimension) FormatString(f format.Formatter) string {
	d.Format(f)
	return f.Output()
}

// BBox returns the bounding box of the definition points and the text
// position. Drawing.SetExt uses the dimension block if it exists.
func (d *Dimension) BBox() ([]float64, []float64) {
	pts := [][]float64{d.DefPoint, d.TextMid}
	switch d.Type {
	case DIM_LINEAR, DIM_ALIGNED, DIM_ORDINATE:
		pts = append(pts, d.DefPoint2, d.DefPoint3)
	case DIM_RADIUS, DIM_DIAMETER:
		pts = append(pts, d.DefPoint4)
	case DIM_ANGULAR3P:
		pts = append(pts, d.DefPoint2, d.DefPoint3, d.DefPoint4)
	case DIM_ANGULAR:
		pts = append(pts, d.DefPoint2, d.DefPoint3, d.DefPoint4, d.DefPoint5)
	}
	mins := []float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxs := []float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, p := range pts {
		for i := 0; i < 3; i++ {
			mins[i] = math.Min(mins[i], p[i])
			maxs[i] = math.Max(maxs[i], p[i])
		}
	}
	return mins, maxs
}
