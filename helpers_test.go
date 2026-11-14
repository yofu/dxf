package dxf

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/yofu/dxf/drawing"
	"golang.org/x/text/encoding/japanese"
)

// TOLERANCE is the epsilon value used in comparing floats.
const TOLERANCE = 0.000001

// cmpF64 compares two floats to see if they are within the given tolerance.
func cmpF64(f1, f2 float64) bool {
	if math.IsInf(f1, 1) {
		return math.IsInf(f2, 1)
	}
	if math.IsInf(f2, 1) {
		return math.IsInf(f1, 1)
	}
	if math.IsInf(f1, -1) {
		return math.IsInf(f2, -1)
	}
	if math.IsInf(f2, -1) {
		return math.IsInf(f1, -1)
	}
	return math.Abs(f1-f2) < TOLERANCE
}

func entitiesSection(ents string) string {
	return "0\nSECTION\n2\nENTITIES\n" + ents + "0\nENDSEC\n0\nEOF\n"
}

func writeString(t *testing.T, d *drawing.Drawing) string {
	t.Helper()
	var b bytes.Buffer
	if _, err := d.WriteTo(&b); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// groupValue returns the value following the first "\n<code>\n" in s.
func groupValue(s, code string) (string, bool) {
	key := "\n" + code + "\n"
	i := strings.Index(s, key)
	if i < 0 {
		return "", false
	}
	v := s[i+len(key):]
	return v[:strings.Index(v, "\n")], true
}

func readPreserve(t *testing.T, src string) *drawing.Drawing {
	t.Helper()
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

type pair struct {
	code  int
	value string
}

type record struct {
	typ     string
	section string
	pairs   []pair
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
	// every LAYOUT refers to a BLOCK_RECORD which refers back to it
	for _, r := range recs {
		if r.typ != "LAYOUT" {
			continue
		}
		rec := handles[r.last(330)]
		if back, _ := rec.get(340); rec.typ != "BLOCK_RECORD" || back != r.handle() {
			t.Errorf("LAYOUT %s: 330 -> %s %s, which refers to %s", r.handle(), rec.typ, r.last(330), back)
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
			if ps, _ := r.get(67); ps == "1" {
				want = records["*PAPER_SPACE"].handle()
			}
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

// groups returns the group codes and values as written with precision 16.
func groups(cv ...interface{}) string {
	var b strings.Builder
	for i := 0; i+1 < len(cv); i += 2 {
		b.WriteString(fmt.Sprintf("\n%d\n", cv[i]))
		switch v := cv[i+1].(type) {
		case float64:
			b.WriteString(fmt.Sprintf("%.16f", v))
		default:
			b.WriteString(fmt.Sprint(v))
		}
	}
	return b.String() + "\n"
}

func hexHandle(h int) string {
	return fmt.Sprintf("%X", h)
}

func sjis(t *testing.T, s string) string {
	t.Helper()
	b, err := japanese.ShiftJIS.NewEncoder().String(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
