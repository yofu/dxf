package entity

// EntityType represents Entity names (code 2)
type EntityType int

// Entity name: code 2
const (
	LINE EntityType = iota
	THREEDFACE
	LWPOLYLINE
	CIRCLE
	POLYLINE
	VERTEX
	POINT
	ARC
	TEXT
	SPLINE
	INSERT
	MTEXT
	ELLIPSE
	HATCH
	SOLID
	DIMENSION
	ATTDEF
	ATTRIB
	VIEWPORT
	LEADER
	UNKNOWN // entity types not supported by this package (see Unknown)
)

// EntityTypeString converts EntityType to string.
// If EntityType is out of range, it returns empty string.
func EntityTypeString(t EntityType) string {
	switch t {
	case LINE:
		return "LINE"
	case THREEDFACE:
		return "3DFACE"
	case LWPOLYLINE:
		return "LWPOLYLINE"
	case CIRCLE:
		return "CIRCLE"
	case POLYLINE:
		return "POLYLINE"
	case VERTEX:
		return "VERTEX"
	case POINT:
		return "POINT"
	case ARC:
		return "ARC"
	case TEXT:
		return "TEXT"
	case SPLINE:
		return "SPLINE"
	case INSERT:
		return "INSERT"
	case MTEXT:
		return "MTEXT"
	case ELLIPSE:
		return "ELLIPSE"
	case HATCH:
		return "HATCH"
	case SOLID:
		return "SOLID"
	case DIMENSION:
		return "DIMENSION"
	case ATTDEF:
		return "ATTDEF"
	case ATTRIB:
		return "ATTRIB"
	case VIEWPORT:
		return "VIEWPORT"
	case LEADER:
		return "LEADER"
	default:
		return ""
	}
}

// EntityTypeValue converts string to EntityType.
// If string is unknown EntityType, it returns -1.
func EntityTypeValue(t string) EntityType {
	switch t {
	case "LINE":
		return LINE
	case "3DFACE":
		return THREEDFACE
	case "LWPOLYLINE":
		return LWPOLYLINE
	case "CIRCLE":
		return CIRCLE
	case "POLYLINE":
		return POLYLINE
	case "VERTEX":
		return VERTEX
	case "POINT":
		return POINT
	case "ARC":
		return ARC
	case "TEXT":
		return TEXT
	case "SPLINE":
		return SPLINE
	case "INSERT":
		return INSERT
	case "MTEXT":
		return MTEXT
	case "ELLIPSE":
		return ELLIPSE
	case "HATCH":
		return HATCH
	case "SOLID":
		return SOLID
	case "DIMENSION":
		return DIMENSION
	case "ATTDEF":
		return ATTDEF
	case "ATTRIB":
		return ATTRIB
	case "VIEWPORT":
		return VIEWPORT
	case "LEADER":
		return LEADER
	default:
		return -1
	}
}
