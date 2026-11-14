package table

import (
	"strings"

	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// ExtGroups holds the 102 groups of a table entry ({ACAD_XDICTIONARY,
// {ACAD_REACTORS, ...} as read) and its extension dictionary.
type ExtGroups struct {
	groups [][2]string
	xdict  handle.Handler
}

// SetExtensionGroups sets the 102 groups as read.
func (x *ExtGroups) SetExtensionGroups(g [][2]string) { x.groups = g }

// ExtensionGroups returns the 102 groups as read.
func (x *ExtGroups) ExtensionGroups() [][2]string { return x.groups }

// SetExtensionDictionary sets the extension dictionary (written as
// 102 {ACAD_XDICTIONARY 360 ...).
func (x *ExtGroups) SetExtensionDictionary(h handle.Handler) { x.xdict = h }

// ExtensionDictionary returns the extension dictionary set by
// SetExtensionDictionary (nil if none).
func (x *ExtGroups) ExtensionDictionary() handle.Handler { return x.xdict }

// ExtensionDictionaryHandle returns the handle of the extension dictionary
// as read ("" if none).
func (x *ExtGroups) ExtensionDictionaryHandle() string {
	in := false
	for _, p := range x.groups {
		switch {
		case p[0] == "102" && strings.EqualFold(strings.TrimSpace(p[1]), "{ACAD_XDICTIONARY"):
			in = true
		case p[0] == "102":
			in = false
		case in && p[0] == "360":
			return strings.TrimSpace(p[1])
		}
	}
	return ""
}

// writeGroups writes the 102 groups; the extension dictionary set replaces
// the one read.
func (x *ExtGroups) writeGroups(f format.Formatter) {
	skip := false
	for _, p := range x.groups {
		if p[0] == "102" && strings.EqualFold(strings.TrimSpace(p[1]), "{ACAD_XDICTIONARY") && x.xdict != nil {
			skip = true
			continue
		}
		if skip {
			if p[0] == "102" {
				skip = false
			}
			continue
		}
		writeRaw(f, [][2]string{p})
	}
	if x.xdict != nil {
		f.WriteString(102, "{ACAD_XDICTIONARY")
		f.WriteHex(360, x.xdict.Handle())
		f.WriteString(102, "}")
	}
}
