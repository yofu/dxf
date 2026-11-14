package dxf

import (
	"strconv"
	"strings"
	"testing"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/drawing"
)

type pair struct {
	code  int
	value string
}

type record struct {
	typ     string
	section string
	pairs   []pair
}

func (r record) get(code int) (string, bool) {
	for _, p := range r.pairs {
		if p.code == code {
			return p.value, true
		}
	}
	return "", false
}

// last returns the last value of code (e.g. the owner 330 at the end of LAYOUT).
func (r record) last(code int) string {
	v := ""
	for _, p := range r.pairs {
		if p.code == code {
			v = p.value
		}
	}
	return v
}

func (r record) handle() string {
	if h, ok := r.get(5); ok {
		return h
	}
	h, _ := r.get(105)
	return h
}

// owner returns the owner handle: the first 330 outside of the {ACAD_REACTORS} group.
func (r record) owner() string {
	inGroup := false
	for _, p := range r.pairs {
		switch {
		case p.code == 102 && strings.HasPrefix(p.value, "{"):
			inGroup = true
		case p.code == 102 && p.value == "}":
			inGroup = false
		case p.code == 330 && !inGroup:
			return p.value
		}
	}
	return ""
}

func splitRecords(t *testing.T, s string) []record {
	t.Helper()
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines)%2 != 0 {
		t.Fatal("odd number of lines")
	}
	var recs []record
	section := ""
	for i := 0; i < len(lines); i += 2 {
		code, err := strconv.Atoi(strings.TrimSpace(lines[i]))
		if err != nil {
			t.Fatalf("line %d: invalid group code %q", i+1, lines[i])
		}
		v := lines[i+1]
		if code == 0 {
			if v == "SECTION" {
				section = lines[i+3]
			}
			recs = append(recs, record{typ: v, section: section})
			continue
		}
		recs[len(recs)-1].pairs = append(recs[len(recs)-1].pairs, pair{code, v})
	}
	return recs
}

func checkStructure(t *testing.T, d *drawing.Drawing) {
	t.Helper()
	recs := splitRecords(t, writeString(t, d))

	handles := map[string]record{}
	maxHandle := int64(0)
	for _, r := range recs {
		h := r.handle()
		if h == "" || r.section == "HEADER" { // $HANDSEED is written with code 5
			continue
		}
		if _, dup := handles[h]; dup {
			t.Errorf("duplicate handle %s (%s)", h, r.typ)
		}
		handles[h] = r
		if v, _ := strconv.ParseInt(h, 16, 64); v > maxHandle {
			maxHandle = v
		}
	}
	// $HANDSEED must be larger than all handles
	for i, r := range recs {
		if r.section == "HEADER" && r.typ == "SECTION" {
			for j, p := range recs[i].pairs {
				if p.code == 9 && p.value == "$HANDSEED" {
					seed, _ := strconv.ParseInt(recs[i].pairs[j+1].value, 16, 64)
					if seed <= maxHandle {
						t.Errorf("$HANDSEED %X <= max handle %X", seed, maxHandle)
					}
				}
			}
		}
	}
	// all references must point to existing objects
	for _, r := range recs {
		for _, p := range r.pairs {
			switch p.code {
			case 330, 340, 350, 360, 390:
				if p.value == "0" {
					continue
				}
				if _, ok := handles[p.value]; !ok {
					t.Errorf("%s %s: code %d refers to undefined handle %s", r.typ, r.handle(), p.code, p.value)
				}
			}
		}
	}

	records := map[string]record{} // BLOCK_RECORD by upper-case name
	blocks := map[string]record{}
	var vportActive, dimStandard bool
	for _, r := range recs {
		name, _ := r.get(2)
		switch r.typ {
		case "BLOCK_RECORD":
			records[strings.ToUpper(name)] = r
		case "BLOCK":
			blocks[strings.ToUpper(name)] = r
		case "VPORT":
			vportActive = vportActive || strings.EqualFold(name, "*Active")
		case "DIMSTYLE":
			dimStandard = dimStandard || strings.EqualFold(name, "Standard")
		}
	}
	if !vportActive {
		t.Error("VPORT *Active is missing")
	}
	if !dimStandard {
		t.Error("DIMSTYLE Standard is missing")
	}
	for name, b := range blocks {
		rec, ok := records[name]
		if !ok {
			t.Errorf("BLOCK %s has no BLOCK_RECORD", name)
			continue
		}
		if b.owner() != rec.handle() {
			t.Errorf("BLOCK %s owner %s, want BLOCK_RECORD %s", name, b.owner(), rec.handle())
		}
	}
	for name := range records {
		if _, ok := blocks[name]; !ok {
			t.Errorf("BLOCK_RECORD %s has no BLOCK", name)
		}
	}
	for _, name := range []string{"*MODEL_SPACE", "*PAPER_SPACE"} {
		rec := records[name]
		lh, ok := rec.get(340)
		if !ok {
			t.Errorf("BLOCK_RECORD %s has no LAYOUT (340)", name)
			continue
		}
		l := handles[lh]
		if l.typ != "LAYOUT" || l.last(330) != rec.handle() {
			t.Errorf("BLOCK_RECORD %s: 340 -> %s %s which refers to %s", name, l.typ, lh, l.last(330))
		}
	}

	// entities: owner is the BLOCK_RECORD of the containing block
	currentBlock := ""
	for _, r := range recs {
		switch r.typ {
		case "SECTION", "ENDSEC", "EOF", "ENDBLK":
			continue
		case "BLOCK":
			name, _ := r.get(2)
			currentBlock = strings.ToUpper(name)
			continue
		}
		var want string
		switch r.section {
		case "ENTITIES":
			want = records["*MODEL_SPACE"].handle()
		case "BLOCKS":
			want = records[currentBlock].handle()
		default:
			continue
		}
		if r.typ == "VERTEX" || r.typ == "SEQEND" || r.typ == "ATTRIB" {
			continue // owned by POLYLINE or INSERT
		}
		if r.owner() != want {
			t.Errorf("%s %s in %s: owner %s, want %s", r.typ, r.handle(), r.section, r.owner(), want)
		}
	}
	for i, r := range recs {
		owner := handles[r.owner()].typ
		switch r.typ {
		case "VERTEX":
			if owner != "POLYLINE" {
				t.Errorf("VERTEX %s should be owned by POLYLINE, got %s", r.handle(), r.owner())
			}
		case "ATTRIB":
			if owner != "INSERT" {
				t.Errorf("ATTRIB %s should be owned by INSERT, got %s", r.handle(), r.owner())
			}
		case "SEQEND":
			if owner != "POLYLINE" && owner != "INSERT" {
				t.Errorf("SEQEND %s should be owned by POLYLINE or INSERT, got %s", r.handle(), r.owner())
			}
			// the owner is the entity that starts the sequence
			j := i - 1
			for j >= 0 && (recs[j].typ == "VERTEX" || recs[j].typ == "ATTRIB") {
				j--
			}
			if j < 0 || recs[j].handle() != r.owner() {
				t.Errorf("SEQEND %s: owner %s is not the preceding POLYLINE/INSERT", r.handle(), r.owner())
			}
		}
	}
}

func TestR2000Structure(t *testing.T) {
	d := NewDrawing()
	d.AddLayer("A", color.Red, DefaultLineType, true)
	d.BeginBlock("B", 0, 0, 0)
	d.Line(0, 0, 0, 1, 0, 0)
	d.Polyline(false, []float64{0, 0, 0}, []float64{1, 1, 1})
	d.EndBlock()
	l, _ := d.Line(0, 0, 0, 1, 1, 0)
	d.Insert("B", 5, 5, 0)
	d.Polyline(true, []float64{0, 0, 0}, []float64{1, 0, 0}, []float64{1, 1, 0})
	d.MText("text", 0, 0, 0, 1, 0)
	d.Group("G", "group", l)
	checkStructure(t, d)

	// written twice: owners and handles must stay consistent
	checkStructure(t, d)

	// read back and write again
	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	checkStructure(t, d2)

	// a file without BLOCK_RECORDs (e.g. R12 style)
	d3, err := FromStringData("0\nSECTION\n2\nBLOCKS\n0\nBLOCK\n8\n0\n2\nX\n70\n0\n10\n0\n20\n0\n0\nLINE\n8\n0\n10\n0\n20\n0\n11\n1\n21\n1\n0\nENDBLK\n8\n0\n0\nENDSEC\n" +
		"0\nSECTION\n2\nENTITIES\n0\nINSERT\n8\n0\n2\nX\n10\n0\n20\n0\n0\nENDSEC\n0\nEOF\n")
	if err != nil {
		t.Fatal(err)
	}
	checkStructure(t, d3)
}
