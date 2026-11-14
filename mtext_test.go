package dxf

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/yofu/dxf/codepage"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
)

func TestMTextBackground(t *testing.T) {
	d := NewDrawing()
	m1, _ := d.MText("color", 0, 0, 0, 1, 10)
	m1.SetBackgroundColor(color.Yellow, 1.2)
	m2, _ := d.MText("true color", 0, 5, 0, 1, 10)
	m2.SetBackgroundTrueColor(10, 20, 30, 2)
	m2.SetTextFrame(true)
	m3, _ := d.MText("canvas", 0, 10, 0, 1, 10)
	m3.SetBackgroundCanvas(1.5)
	m4, _ := d.MText("frame", 0, 15, 0, 1, 10)
	m4.SetTextFrame(true)

	if out := writeString(t, d); strings.Contains(out, groups(45, 1.2)) || strings.Contains(out, groups(90, 16, 0, "ENDSEC")) {
		t.Error("background fill should not be written for AC1015")
	}
	d.SetVersion(format.R2004)
	out := writeString(t, d)
	for _, w := range []string{
		groups(44, 1.0, 45, 1.2, 90, 1, 63, 2, 0, "MTEXT"),
		groups(44, 1.0, 45, 2.0, 90, 17, 63, int(color.Nearest(10, 20, 30)), 421, 660510, 0, "MTEXT"),
		groups(44, 1.0, 45, 1.5, 90, 3, 63, 0, 0, "MTEXT"),
		groups(44, 1.0, 90, 16, 0, "ENDSEC"),
	} {
		if !strings.Contains(out, w) {
			t.Errorf("%q not found", w)
		}
	}
	checkStructure(t, d)

	d2 := readPreserve(t, out)
	es := d2.Entities()
	r1, r2, r3, r4 := es[0].(*entity.MText), es[1].(*entity.MText), es[2].(*entity.MText), es[3].(*entity.MText)
	if r1.BackgroundFill != 1 || r1.BackgroundColor != 2 || r1.BackgroundScale != 1.2 || r1.BackgroundTrueColor != -1 {
		t.Errorf("color: %+v", r1)
	}
	if r2.BackgroundFill != 17 || r2.BackgroundTrueColor != color.RGB(10, 20, 30) || r2.BackgroundColor == 0 {
		t.Errorf("true color: %+v", r2)
	}
	if r3.BackgroundFill != 3 || r4.BackgroundFill != 16 {
		t.Errorf("canvas %d, frame %d", r3.BackgroundFill, r4.BackgroundFill)
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
	r2.ClearBackground()
	if r2.BackgroundFill != 16 || r2.BackgroundTrueColor != -1 {
		t.Errorf("clear: %+v", r2)
	}
}

func TestMTextRoundTrip(t *testing.T) {
	d := NewDrawing()
	d.AddLayer("注記", color.Green, DefaultLineType, true)
	mt, err := d.MText("1行目\n2行目\r\n3行目", 10, 20, 0, 2.5, 50)
	if err != nil {
		t.Fatal(err)
	}
	if mt.Value != `1行目\P2行目\P3行目` {
		t.Fatalf("line breaks should be converted to \\P: %q", mt.Value)
	}
	mt.Attachment = entity.MTEXT_MIDDLE_CENTER
	mt.Rotation = 30
	mt.LineSpacingStyle = entity.MTEXT_EXACT
	mt.LineSpacingFactor = 1.5
	mt.SetColor(color.Blue)

	out := writeString(t, d)
	if !strings.Contains(out, "AcDbMText") {
		t.Error("MTEXT should be written with AcDbMText")
	}
	d2, err := FromStringData(out)
	if err != nil {
		t.Fatal(err)
	}
	got := d2.Entities()[0].(*entity.MText)
	if got.Value != mt.Value || got.PlainText() != "1行目\n2行目\n3行目" {
		t.Errorf("value %q plain %q", got.Value, got.PlainText())
	}
	if got.Coord[0] != 10 || got.Coord[1] != 20 || got.Height != 2.5 || got.Width != 50 ||
		got.Attachment != entity.MTEXT_MIDDLE_CENTER || !cmpF64(got.Rotation, 30) ||
		got.LineSpacingStyle != entity.MTEXT_EXACT || got.LineSpacingFactor != 1.5 ||
		got.ColorIndex() != int(color.Blue) || got.Layer().Name() != "注記" || got.Style.Name() != "Standard" {
		t.Errorf("round trip: %+v", got)
	}
}

func TestMTextLongTextChunks(t *testing.T) {
	long := strings.Repeat("あいうえおABCDE", 60) // 600 characters
	for _, cp := range []string{codepage.Default, codepage.Japanese} {
		d := NewDrawing()
		d.Header().CodePage = cp
		mt, _ := d.MText(long, 0, 0, 0, 1, 0)

		s := mt.String()
		if n := strings.Count(s, "\n3\n"); n != 2 {
			t.Errorf("%s: want 2 chunks in group 3, got %d", cp, n)
		}
		for _, line := range strings.Split(s, "\n") {
			if utf8.RuneCountInString(line) > 250 {
				t.Errorf("%s: chunk longer than 250 characters", cp)
			}
		}

		var b bytes.Buffer
		d.WriteTo(&b)
		d2, err := FromReader(&b)
		if err != nil {
			t.Fatal(err)
		}
		if v := d2.Entities()[0].(*entity.MText).Value; v != long {
			t.Errorf("%s: long text round trip failed (%d characters)", cp, utf8.RuneCountInString(v))
		}
	}

	// a chunk must not end with a single backslash (\P would be split)
	s := strings.Repeat("a", 249) + `\P` + strings.Repeat("b", 10)
	mt := entity.NewMText()
	mt.Value = s
	out := mt.String()
	i := strings.Index(out, "\n3\n")
	chunk := out[i+3:]
	chunk = chunk[:strings.Index(chunk, "\n")]
	if strings.HasSuffix(chunk, `\`) {
		t.Error("chunk ends with a backslash")
	}
	d := NewDrawing()
	d.AddEntity(mt)
	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	if v := d2.Entities()[0].(*entity.MText).Value; v != s {
		t.Errorf("split escape round trip: %q", v)
	}
}

func TestReadMText(t *testing.T) {
	src := entitiesSection("0\nMTEXT\n8\n0\n62\n1\n10\n1\n20\n2\n30\n0\n40\n2.5\n41\n100\n71\n5\n72\n1\n" +
		`3` + "\n" + `{\fArial|b1|i0|c0|p34;太字}\P\H2.5x;大きい\S1^2;` + "\n" +
		`1` + "\n" + `\~\Lunder\l \\end\U+00B1` + "\n" +
		"7\nSTANDARD\n50\n45\n73\n1\n44\n1\n" +
		"0\nMTEXT\n8\n0\n10\n0\n20\n0\n40\n1\n1\nrotated\n11\n0\n21\n1\n31\n0\n50\n10\n")
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	mt := d.Entities()[0].(*entity.MText)
	if want := "太字\n大きい1/2 under \\end±"; mt.PlainText() != want {
		t.Errorf("plain text: got %q want %q", mt.PlainText(), want)
	}
	if mt.Attachment != entity.MTEXT_MIDDLE_CENTER || mt.Width != 100 || mt.Rotation != 45 || mt.ColorIndex() != 1 {
		t.Errorf("mtext: %+v", mt)
	}
	// X-axis direction takes precedence over 50
	if r := d.Entities()[1].(*entity.MText).Rotation; !cmpF64(r, 90) {
		t.Errorf("rotation from 11/21: got %v", r)
	}
}

func TestMTextPlain(t *testing.T) {
	cases := map[string]string{
		`abc`:                    "abc",
		`a\Pb`:                   "a\nb",
		`{\C1;red} text`:         "red text",
		`\A1;\W0.8;\Q15;\T1.1;x`: "x",
		`\S1#2;`:                 "1/2",
		`\{braces\}`:             "{braces}",
		`\Oover\o`:               "over",
		`trailing\`:              `trailing\`,
		`\Xunknown`:              `\Xunknown`,
	}
	for in, want := range cases {
		if got := entity.MTextPlain(in); got != want {
			t.Errorf("MTextPlain(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMTextExtents(t *testing.T) {
	d := NewDrawing()
	d.MText("line1\nline2", 0, 10, 0, 2, 20) // top left at (0, 10), 2 lines
	d.SetExt()
	h := d.Header()
	if !cmpF64(h.ExtMin[0], 0) || !cmpF64(h.ExtMax[0], 20) || !cmpF64(h.ExtMin[1], 6) || !cmpF64(h.ExtMax[1], 10) {
		t.Errorf("extents: %v - %v", h.ExtMin, h.ExtMax)
	}
}
