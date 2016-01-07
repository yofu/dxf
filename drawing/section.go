package drawing

import (
	"github.com/yofu/dxf/format"
)

// Section is interface for DXF sections.
type Section interface {
	WriteTo(format.Formatter)
	SetHandle(*int)
}

// SectionType represents Section names (code 2)
type SectionType int

// Section name: code 2
const (
	HEADER SectionType = iota
	CLASSES
	TABLES
	BLOCKS
	ENTITIES
	OBJECTS
)

// SectionTypeString converts SectionType to string.
// If SectionType is out of range, it returns empty string.
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

// SectionTypeValue converts string to SectionType.
// If string is unknown SectionType, it returns -1.
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
