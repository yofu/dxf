package entity

type EntityType int

const (
	LINE EntityType = iota
	ARC
	CIRCLE
	POLYLINE
	TEXT
)

func EntityTypeString(t EntityType) string {
	switch t {
	case LINE:
		return "LINE"
	case ARC:
		return "ARC"
	case CIRCLE:
		return "CIRCLE"
	case POLYLINE:
		return "POLYLINE"
	case TEXT:
		return "TEXT"
	default:
		return ""
	}
}
