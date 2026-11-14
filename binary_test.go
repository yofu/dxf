package dxf

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yofu/dxf/codepage"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
)

// richDrawing creates a drawing using many kinds of entities and objects.
func richDrawing(t *testing.T) *drawing.Drawing {
	t.Helper()
	d := NewDrawing()
	d.AddLayer("壁", color.Red, DefaultLineType, true)
	d.BeginBlock("TITLE", 0, 0, 0)
	d.Line(0, 0, 0, 100, 0, 0)
	d.AttributeDefinition("NAME", "図面名", "未定", 10, 5, 0, 3)
	d.EndBlock()
	ins, _ := d.Insert("TITLE", 10, 20, 0)
	ins.SetAttribute("NAME", "平面図")
	l, _ := d.Line(0, 0, 0, 1.25, -3.5, 0)
	l.SetXData([][2]string{{"1001", "APP"}, {"1000", "data"}, {"1040", "0.1"}, {"1070", "-2"}, {"1071", "70000"}})
	lw, _ := d.LwPolyline(true, []float64{0, 0}, []float64{10, 0}, []float64{10, 10})
	lw.Bulges[0] = 0.5
	lw.SetWidth(1, 0.2, 0.4)
	d.Text("日本語テキスト", 0, 0, 0, 2.5)
	d.MText("1行目\n2行目", 0, 10, 0, 2.5, 50)
	d.Spline([][]float64{{0, 0, 0}, {1, 1, 0}, {2, 0, 0}, {3, 1, 0}})
	d.Ellipse([]float64{0, 0, 0}, []float64{5, 0, 0}, 0.5)
	h, _ := d.Hatch([]float64{0, 0}, []float64{10, 0}, []float64{10, 10})
	h.SetPattern("ANSI31", 0, 1)
	d.LinearDimension([]float64{0, 0, 0}, []float64{100, 0, 0}, []float64{50, 20, 0}, 0)
	d.LeaderText("注記", 0, []float64{0, 0, 0}, []float64{10, 10, 0})
	d.Group("G", "group", l)
	d.PaperSpace()
	d.Viewport([]float64{210, 148.5, 0}, 380, 250, []float64{50, 0}, 200)
	d.ModelSpace()
	return d
}

func ascii(t *testing.T, d *drawing.Drawing) string {
	t.Helper()
	d.SetFormatter(format.NewASCII())
	d.Formatter().SetPrecision(16)
	return writeString(t, d)
}

func TestBinaryRoundTrip(t *testing.T) {
	d := richDrawing(t)
	want := ascii(t, d)

	d.SetFormatter(format.NewBinary())
	var b bytes.Buffer
	if _, err := d.WriteTo(&b); err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(b.Bytes(), []byte(format.BinarySentinel)) {
		t.Fatal("binary DXF should start with the sentinel")
	}
	if !bytes.HasSuffix(b.Bytes(), append(append([]byte{0, 0}, "EOF"...), 0)) {
		t.Error("binary DXF should end with 0 EOF")
	}
	// non-ASCII text is written as \U+XXXX in the default code page
	if !bytes.Contains(b.Bytes(), []byte(`\U+58C1`)) {
		t.Error("layer name should be written as \\U+58C1")
	}

	d2, err := FromReader(&b)
	if err != nil {
		t.Fatal(err)
	}
	checkStructure(t, d2)
	if got := ascii(t, d2); got != want {
		t.Error("a drawing read from binary DXF should be the same as the original")
		gl, wl := strings.Split(got, "\n"), strings.Split(want, "\n")
		for i := 0; i < len(gl) && i < len(wl); i++ {
			if gl[i] != wl[i] {
				t.Logf("line %d: got %q want %q", i+1, gl[i], wl[i])
				break
			}
		}
	}
}

func TestBinaryShiftJIS(t *testing.T) {
	d := NewDrawing()
	d.Header().CodePage = codepage.Japanese
	d.Text("壁", 0, 0, 0, 1)
	d.SetFormatter(format.NewBinary())
	var b bytes.Buffer
	d.WriteTo(&b)
	if !bytes.Contains(b.Bytes(), []byte(sjis(t, "壁")+"\x00")) {
		t.Error("text should be written in Shift_JIS")
	}
	d2, err := FromReader(&b)
	if err != nil {
		t.Fatal(err)
	}
	if v := d2.Entities()[0].(*entity.Text).Value; v != "壁" {
		t.Errorf("text: %q", v)
	}
}

func TestSaveAsBinary(t *testing.T) {
	d := richDrawing(t)
	fn := filepath.Join(t.TempDir(), "binary.dxf")
	if err := d.SaveAsBinary(fn); err != nil {
		t.Fatal(err)
	}
	if _, ok := d.Formatter().(*format.ASCII); !ok {
		t.Error("SaveAsBinary should not change the formatter")
	}
	data, _ := os.ReadFile(fn)
	if !bytes.HasPrefix(data, []byte(format.BinarySentinel)) {
		t.Error("file should be binary DXF")
	}
	d2, err := FromFile(fn)
	if err != nil {
		t.Fatal(err)
	}
	if len(d2.Entities()) != len(d.Entities()) {
		t.Errorf("entities: %d, want %d", len(d2.Entities()), len(d.Entities()))
	}
}

func TestBinaryChunkRoundTrip(t *testing.T) {
	src := entitiesSection("0\nACAD_PROXY_ENTITY\n8\n0\n100\nAcDbProxyEntity\n90\n498\n92\n4\n310\n0A0B0C0D\n")
	d := readPreserve(t, src)
	d.SetFormatter(format.NewBinary())
	var b bytes.Buffer
	d.WriteTo(&b)
	d2, err := FromReader(&b)
	if err != nil {
		t.Fatal(err)
	}
	u := d2.Entities()[0].(*entity.Unknown)
	if v, _ := value(u.Data, "310"); v != "0A0B0C0D" {
		t.Errorf("binary chunk: %v", u.Data)
	}
}

func TestReadBinaryR12(t *testing.T) {
	// R12 binary DXF: 1-byte group codes
	var b bytes.Buffer
	b.WriteString(format.BinarySentinel)
	str := func(code byte, s string) { b.WriteByte(code); b.WriteString(s); b.WriteByte(0) }
	dbl := func(code byte, v float64) { b.WriteByte(code); binary.Write(&b, binary.LittleEndian, v) }
	str(0, "SECTION")
	str(2, "ENTITIES")
	str(0, "LINE")
	str(8, "0")
	b.WriteByte(62)
	binary.Write(&b, binary.LittleEndian, int16(3))
	dbl(10, 1)
	dbl(20, 2)
	dbl(11, 3)
	dbl(21, 4)
	str(0, "ENDSEC")
	str(0, "EOF")
	d, err := FromReader(&b)
	if err != nil {
		t.Fatal(err)
	}
	l := d.Entities()[0].(*entity.Line)
	if l.Start[1] != 2 || l.End[0] != 3 || l.ColorIndex() != 3 {
		t.Errorf("line: %v %v color %d", l.Start, l.End, l.ColorIndex())
	}
}

func TestReadTruncatedBinary(t *testing.T) {
	d := NewDrawing()
	d.Line(0, 0, 0, 1, 1, 0)
	d.SetFormatter(format.NewBinary())
	var b bytes.Buffer
	d.WriteTo(&b)
	data := b.Bytes()
	if _, err := FromReader(bytes.NewReader(data[:len(data)/2+3])); err == nil {
		t.Error("truncated binary DXF should be an error")
	}
}
