package entity

type EntityType int

const (
	LINE EntityType = iota
	THREEDFACE
	LWPOLYLINE
	CIRCLE
	POLYLINE
	VERTEX
	ARC
	TEXT
)

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
	case ARC:
		return "ARC"
	case TEXT:
		return "TEXT"
	default:
		return ""
	}
}
