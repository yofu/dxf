package drawing

import (
	"github.com/yofu/dxf/format"
)

type Section interface {
	WriteTo(*format.Formatter)
	SetHandle(*int)
	Read(int, [][2]string) error
}

type SectionType int

const (
	HEADER SectionType = iota
	CLASSES
	TABLES
	BLOCKS
	ENTITIES
	OBJECTS
)

func SectionTypeString(s SectionType) string {
	switch s {
	case HEADER:
		return "HEADER"
	case CLASSES:
		return "CLASSES"
	case TABLES:
		return "TABLES"
	case BLOCKS:
		return "BLOCKS"
	case ENTITIES:
		return "ENTITIES"
	case OBJECTS:
		return "OBJECTS"
	default:
		return ""
	}
}

func SectionTypeValue(s string) SectionType {
	switch s {
	case "HEADER":
		return HEADER
	case "CLASSES":
		return CLASSES
	case "TABLES":
		return TABLES
	case "BLOCKS":
		return BLOCKS
	case "ENTITIES":
		return ENTITIES
	case "OBJECTS":
		return OBJECTS
	default:
		return -1
	}
}
