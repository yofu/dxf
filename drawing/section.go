package drawing

import (
	"strconv"

	"github.com/yofu/dxf/format"
)

// Section is interface for DXF sections.
type Section interface {
	Format(format.Formatter)
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

// RawSection is a section this package does not interpret (e.g.
// ACDSDATA with the ACIS data of solids for AC1027 and later, or
// THUMBNAILIMAGE). It is kept as read and written after the OBJECTS
// section.
type RawSection struct {
	Name    string      // 2
	Data    [][2]string // group codes between the name and ENDSEC
	Version string      // DXF version of the file read: not written for earlier versions
}

// Format writes the section (nothing for versions earlier than Version).
func (s *RawSection) Format(f format.Formatter) {
	if s.Version != "" && !format.AtLeast(f, s.Version) {
		return
	}
	f.WriteString(0, "SECTION")
	f.WriteString(2, s.Name)
	for _, p := range s.Data {
		if c, err := strconv.Atoi(p[0]); err == nil {
			f.WriteString(c, p[1])
		}
	}
	f.WriteString(0, "ENDSEC")
}

// SetHandle does nothing: the handles of a raw section are kept as read.
func (s *RawSection) SetHandle(*int) {}

// AddRawSection adds a section kept as read.
func (d *Drawing) AddRawSection(name string, data [][2]string) *RawSection {
	s := &RawSection{Name: name, Data: data, Version: d.Header().Version}
	d.Sections = append(d.Sections, s)
	return s
}

// RawSections returns the sections kept as read.
func (d *Drawing) RawSections() []*RawSection {
	var rs []*RawSection
	for _, s := range d.Sections {
		if r, ok := s.(*RawSection); ok {
			rs = append(rs, r)
		}
	}
	return rs
}
