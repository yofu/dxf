package entity

import (
	"strconv"
	"strings"

	"github.com/yofu/dxf/format"
)

// Body represents the entities with ACIS data: BODY, REGION, 3DSOLID,
// SURFACE, PLANESURFACE, EXTRUDEDSURFACE, LOFTEDSURFACE, REVOLVEDSURFACE
// and SWEPTSURFACE (see TypeName). The geometry is not interpreted: the ACIS
// data can be read and replaced as SAT text (AC1015 to AC1024; AC1027 and
// later keep it as SAB in the ACDSDATA section, kept as read).
type Body struct {
	*entity
	Version int      // 70: modeler format version (1)
	SAT     []string // 1, 3: ACIS data as SAT text lines (decoded), AC1024 or earlier
	Flags   int      // 290 (AC1027 or later)
	UID     string   // 2 (AC1027 or later)
	// Extra holds the group codes after the modeler geometry (e.g.
	// AcDb3dSolid history, AcDbSurface and the data of the surface
	// subclasses), kept as read.
	Extra [][2]string

	hasVersion bool
}

// IsEntity is for Entity interface.
func (b *Body) IsEntity() bool { return true }

// NewBody creates an entity of ACIS type name (e.g. "3DSOLID") with SAT
// text lines.
func NewBody(name string, sat []string) *Body {
	b := &Body{entity: NewEntity(BODY), Version: 1, SAT: sat, hasVersion: true}
	b.typeName = strings.ToUpper(name)
	return b
}

// TypeName returns the entity type (e.g. "3DSOLID").
func (b *Body) TypeName() string { return b.typeName }

// SetVersionGiven sets whether 70 is written for AC1027 and later.
func (b *Body) SetVersionGiven(v bool) { b.hasVersion = v }

// Format writes data to formatter.
func (b *Body) Format(f format.Formatter) {
	b.entity.Format(f)
	f.WriteString(100, "AcDbModelerGeometry")
	if format.AtLeast(f, format.R2013) {
		if b.hasVersion {
			f.WriteInt(70, b.Version)
		}
		f.WriteInt(290, b.Flags)
		f.WriteString(2, b.UID)
	} else {
		f.WriteInt(70, b.Version)
		for _, line := range EncodeSAT(b.SAT) {
			f.WriteString(1, line[:min(len(line), 255)])
			for rest := line[min(len(line), 255):]; len(rest) > 0; rest = rest[min(len(rest), 255):] {
				f.WriteString(3, rest[:min(len(rest), 255)])
			}
		}
	}
	for _, p := range b.Extra {
		writeRawPair(f, p)
	}
	if b.typeName == "3DSOLID" && len(b.Extra) == 0 && format.AtLeast(f, format.R2007) {
		f.WriteString(100, "AcDb3dSolid")
		f.WriteString(350, "0")
	}
}

// String outputs data using default formatter.
func (b *Body) String() string { return b.FormatString(format.NewASCII()) }

// FormatString outputs data using given formatter.
func (b *Body) FormatString(f format.Formatter) string {
	b.Format(f)
	return f.Output()
}

// BBox returns zero boxes: the geometry is not interpreted.
func (b *Body) BBox() ([]float64, []float64) { return []float64{0, 0, 0}, []float64{0, 0, 0} }

// ApplyTransform returns ErrTransform: the ACIS geometry is not interpreted.
func (b *Body) ApplyTransform(t Transform) error { return ErrTransform }

// DecodeSAT decodes SAT text lines as stored in DXF (AutoCAD's simple
// "encryption").
func DecodeSAT(lines []string) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		var sb strings.Builder
		skip := false
		for j := 0; j < len(line); j++ {
			c := line[j]
			if skip {
				skip = false
				continue
			}
			switch {
			case c == 0x20:
				sb.WriteByte(' ')
			case c == 0x40:
				sb.WriteByte('_')
			case c == 0x5F:
				sb.WriteByte('@')
			case c >= 0x41 && c < 0x5F:
				sb.WriteByte(byte(0x41 + (0x5E - int(c))))
				skip = c == 0x5E // a space follows 'A'
			default:
				sb.WriteByte(c ^ 0x5F)
			}
		}
		out[i] = sb.String()
	}
	return out
}

// EncodeSAT encodes SAT text lines for DXF (inverse of DecodeSAT).
func EncodeSAT(lines []string) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		var sb strings.Builder
		for j := 0; j < len(line); j++ {
			c := line[j]
			switch {
			case c == ' ':
				sb.WriteByte(' ')
			case c == '_':
				sb.WriteByte('@')
			case c == '@':
				sb.WriteByte('_')
			case c >= 0x41 && c < 0x5F:
				e := byte(0x5E - (int(c) - 0x41))
				sb.WriteByte(e)
				if e == 0x5E {
					sb.WriteByte(' ')
				}
			default:
				sb.WriteByte(c ^ 0x5F)
			}
		}
		out[i] = sb.String()
	}
	return out
}

// writeRawPair writes a group code/value pair kept as read.
func writeRawPair(f format.Formatter, p [2]string) {
	if c, err := strconv.Atoi(strings.TrimSpace(p[0])); err == nil {
		f.WriteString(c, p[1])
	}
}
