package entity

type EntityType int

const (
	LINE EntityType = iota
	THREEDFACE
	LWPOLYLINE
	CIRCLE
	ARC
	POLYLINE
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
	case ARC:
		return "ARC"
	case POLYLINE:
		return "POLYLINE"
	case TEXT:
		return "TEXT"
	default:
		return ""
	}
}
