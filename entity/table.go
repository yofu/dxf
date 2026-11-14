package entity

import (
	"strconv"
	"strings"

	"github.com/yofu/dxf/format"
)

// AcadTable represents ACAD_TABLE entities (AC1018 or later). The graphics
// are in the anonymous block BlockName ("*T..."); the table data
// (AcDbTable) is kept as read. Content returns the texts of the cells.
type AcadTable struct {
	*entity
	BlockName string    // 2
	Insert    []float64 // 10, 20, 30
	// BlockReference holds the other group codes of AcDbBlockReference.
	BlockReference [][2]string
	// Data holds the group codes after the AcDbTable subclass marker.
	Data [][2]string
}

// IsEntity is for Entity interface.
func (t *AcadTable) IsEntity() bool { return true }

// NewAcadTable creates an ACAD_TABLE referring to the block name.
func NewAcadTable(block string, insert []float64) *AcadTable {
	return &AcadTable{entity: NewEntity(ACAD_TABLE), BlockName: block,
		Insert: []float64{coord(insert, 0), coord(insert, 1), coord(insert, 2)}}
}

// Format writes data to formatter.
func (t *AcadTable) Format(f format.Formatter) {
	t.entity.Format(f)
	f.WriteString(100, "AcDbBlockReference")
	f.WriteString(2, t.BlockName)
	writePoint3(f, 10, t.Insert)
	for _, p := range t.BlockReference {
		writeRawPair(f, p)
	}
	f.WriteString(100, "AcDbTable")
	for _, p := range t.Data {
		writeRawPair(f, p)
	}
}

// String outputs data using default formatter.
func (t *AcadTable) String() string { return t.FormatString(format.NewASCII()) }

// FormatString outputs data using given formatter.
func (t *AcadTable) FormatString(f format.Formatter) string {
	t.Format(f)
	return f.Output()
}

// BBox returns the insertion point (Drawing.SetExt uses the block).
func (t *AcadTable) BBox() ([]float64, []float64) { return pointsBBox(t.Insert) }

// ApplyTransform moves the table (only translations: the table data has
// sizes and directions that are not transformed).
func (t *AcadTable) ApplyTransform(tr Transform) error {
	sx, sy, sz := tr.Factors()
	if tr.Angle != 0 || sx != 1 || sy != 1 || sz != 1 {
		return ErrTransform
	}
	transformPoint(tr, t.Insert)
	return nil
}

func (t *AcadTable) first(code string) (string, bool) {
	for _, p := range t.Data {
		if p[0] == code {
			return strings.TrimSpace(p[1]), true
		}
	}
	return "", false
}

// Size returns the numbers of rows (91) and columns (92).
func (t *AcadTable) Size() (rows, cols int) {
	if v, ok := t.first("91"); ok {
		rows, _ = strconv.Atoi(v)
	}
	if v, ok := t.first("92"); ok {
		cols, _ = strconv.Atoi(v)
	}
	return
}

// Content returns the texts of the cells by rows (as ezdxf reads them:
// code 302 for AC1021 and later, chunks of codes 1 to 3 before). If the
// numbers of rows or columns are missing, all texts are in the first row.
func (t *AcadTable) Content() [][]string {
	split := "171" // AC1018
	if _, ok := t.first("302"); ok {
		split = "301" // AC1021 and later
	}
	var values []string
	var group [][2]string
	flush := func() {
		if group == nil {
			return
		}
		long, has302 := "", false
		var sb strings.Builder
		for _, p := range group {
			switch p[0] {
			case "302":
				if !has302 {
					long, has302 = p[1], true
				}
			case "1", "2", "3":
				sb.WriteString(p[1])
			}
		}
		if has302 {
			values = append(values, long)
		} else {
			values = append(values, sb.String())
		}
	}
	for _, p := range t.Data {
		if p[0] == split {
			flush()
			group = [][2]string{}
		}
		if group != nil {
			group = append(group, p)
		}
	}
	flush()
	rows, cols := t.Size()
	if rows*cols == 0 {
		return [][]string{values}
	}
	var out [][]string
	for r := 0; r < rows; r++ {
		start, end := r*cols, (r+1)*cols
		if start > len(values) {
			start = len(values)
		}
		if end > len(values) {
			end = len(values)
		}
		out = append(out, values[start:end])
	}
	return out
}
