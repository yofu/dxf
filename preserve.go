package dxf

import (
	"strconv"
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// isPointer reports whether code is a handle reference (soft/hard pointer
// or owner, 320-369 and 390, or an XDATA handle 1005). Such references
// cannot be kept because handles are renumbered when writing.
func isPointer(code string) bool {
	c, err := strconv.Atoi(strings.TrimSpace(code))
	if err != nil {
		return false
	}
	return (c >= 320 && c <= 369) || c == 390 || c == 1005
}

// isHardPointer reports whether code is a hard pointer or hard owner (340-349, 360-369).
func isHardPointer(code string) bool {
	c, err := strconv.Atoi(strings.TrimSpace(code))
	if err != nil {
		return false
	}
	return (c >= 340 && c <= 349) || (c >= 360 && c <= 369)
}

// rawExtra returns the group codes of data (a table entry or an entity,
// without its 0 group) that are not handled: structural codes (handles,
// owners, 102 groups, and subclass markers unless keepMarkers), handled
// codes and handle references are removed. XDATA (1001 and after) is kept
// as is, except handle references.
func rawExtra(data [][2]string, handled map[string]bool, keepMarkers bool) [][2]string {
	var extra [][2]string
	inGroup := false
	inXData := false
	for _, dt := range data {
		code := strings.TrimSpace(dt[0])
		if code == "1001" {
			inXData = true
		}
		if inXData {
			if code != "1005" {
				extra = append(extra, [2]string{code, dt[1]})
			}
			continue
		}
		switch {
		case code == "102":
			inGroup = strings.HasPrefix(strings.TrimSpace(dt[1]), "{")
			continue
		case inGroup:
			continue
		case code == "100" && keepMarkers:
		case code == "0", code == "5", code == "105", code == "100", isPointer(code), handled[code]:
			continue
		}
		extra = append(extra, [2]string{code, dt[1]})
	}
	return extra
}

// codes returns a set of group codes.
func codes(cs ...string) map[string]bool {
	m := make(map[string]bool, len(cs))
	for _, c := range cs {
		m[c] = true
	}
	return m
}

// splitXData splits the group codes of an entity at the first 1001
// (extended data).
func splitXData(data [][2]string) ([][2]string, [][2]string) {
	for i, dt := range data {
		if strings.TrimSpace(dt[0]) == "1001" {
			return data[:i], data[i:]
		}
	}
	return data, nil
}

// common entity group codes written by entity.Format
var commonEntityCodes = codes("8", "6", "62", "370", "48", "67", "60")

// parseUnknown creates an Unknown entity from the group codes of an entity
// type not supported by this package. Entities with hard pointers to
// other objects (e.g. IMAGE to IMAGEDEF) are not preserved, since the
// objects are not kept; ok is false then.
func parseUnknown(d *drawing.Drawing, data [][2]string) (*entity.Unknown, bool) {
	if len(data) == 0 {
		return nil, false
	}
	body, _ := splitXData(data[1:])
	for _, dt := range body {
		if isHardPointer(dt[0]) {
			return nil, false
		}
	}
	u := entity.NewUnknown(strings.TrimSpace(data[0][1]))
	if name, ok := value(body, "8"); ok {
		if l, err := d.Layer(name, false); err == nil {
			u.SetLayer(l)
		}
	}
	// the common properties are in the AcDbEntity part: before the first
	// other subclass marker (or everywhere without subclass markers)
	common := len(body)
	for i, dt := range body {
		if dt[0] == "100" && strings.TrimSpace(dt[1]) != "AcDbEntity" {
			common = i
			break
		}
	}
	u.Data = append(rawExtra(body[:common], commonEntityCodes, false), rawExtra(body[common:], nil, true)...)
	return u, true
}
