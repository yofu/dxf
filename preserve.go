package dxf

import (
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// rawExtra returns the group codes of data (a table entry or an entity,
// without its 0 group) that are not handled: handles (5, 105), 102 groups
// and handled codes are removed; for table entries and the common part of
// entities (keepMarkers false) also owners (330) and subclass markers.
// References by handle are kept: handles of objects read are kept when
// writing. XDATA (1001 and after) is kept as is.
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
			extra = append(extra, [2]string{code, dt[1]})
			continue
		}
		switch {
		case code == "102":
			inGroup = strings.HasPrefix(strings.TrimSpace(dt[1]), "{")
			continue
		case inGroup:
			continue
		case (code == "100" || code == "330") && keepMarkers:
		case code == "0", code == "5", code == "105", code == "100", code == "330", handled[code]:
			continue
		}
		extra = append(extra, [2]string{code, dt[1]})
	}
	return extra
}

// groups102 returns the 102 groups ({ACAD_REACTORS ...}, {ACAD_XDICTIONARY
// ...}) of data, as read.
func groups102(data [][2]string) [][2]string {
	var gs [][2]string
	inGroup := false
	for _, dt := range data {
		code := strings.TrimSpace(dt[0])
		if code == "102" {
			v := strings.TrimSpace(dt[1])
			gs = append(gs, [2]string{code, v})
			inGroup = strings.HasPrefix(v, "{")
			continue
		}
		if inGroup {
			gs = append(gs, [2]string{code, dt[1]})
		}
	}
	return gs
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
// type not supported by this package. References by handle (e.g. IMAGE to
// IMAGEDEF) are kept, as the handles and the objects read are kept.
func parseUnknown(d *drawing.Drawing, data [][2]string) (*entity.Unknown, bool) {
	if len(data) == 0 {
		return nil, false
	}
	body, _ := splitXData(data[1:])
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
