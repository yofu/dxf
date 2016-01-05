package entity

import (
	"errors"
)

type EntityType int

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
	case POINT:
		return "POINT"
	case ARC:
		return "ARC"
	case TEXT:
		return "TEXT"
	default:
		return ""
	}
}

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
	default:
		return -1
	}
}

func ParseEntityFunc(t string) (func([][2]string)(Entity, error), error) {
	switch t {
	case "LINE":
		return ParseLine, nil
	case "3DFACE":
		return Parse3DFace, nil
	case "LWPOLYLINE":
		return ParseLwPolyline, nil
	case "CIRCLE":
		return ParseCircle, nil
	case "POLYLINE":
		return ParsePolyline, nil
	case "VERTEX":
		return ParseVertex, nil
	case "POINT":
		return ParsePoint, nil
	case "TEXT":
		return ParseText, nil
	default:
		return nil, errors.New("unknown entity type")
	}
}
