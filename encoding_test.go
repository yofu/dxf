package dxf

import (
	"bytes"
	"strings"
	"testing"

	"golang.org/x/text/encoding/japanese"

	"github.com/yofu/dxf/codepage"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/entity"
)

func sjis(t *testing.T, s string) string {
	t.Helper()
	b, err := japanese.ShiftJIS.NewEncoder().String(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func japaneseDrawing(t *testing.T) []byte {
	t.Helper()
	d := NewDrawing()
	if _, err := d.AddLayer("壁", color.Red, DefaultLineType, true); err != nil {
		t.Fatal(err)
	}
	d.Text("日本語テキスト😀", 0, 0, 0, 1)
	var b bytes.Buffer
	if _, err := d.WriteTo(&b); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func checkJapanese(t *testing.T, data []byte) {
	t.Helper()
	d, err := FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := d.Layers["壁"]; !ok {
		t.Error("layer 壁 not found")
	}
	tx := d.Entities()[0].(*entity.Text)
	if tx.Value != "日本語テキスト😀" || tx.Layer().Name() != "壁" {
		t.Errorf("text %q on layer %q", tx.Value, tx.Layer().Name())
	}
}

func TestWriteJapaneseDefaultCodePage(t *testing.T) {
	// default ANSI_1252: non-ASCII characters are written as \U+XXXX
	out := japaneseDrawing(t)
	for _, c := range out {
		if c >= 0x80 {
			t.Fatal("output should be ASCII only")
		}
	}
	if !bytes.Contains(out, []byte("$DWGCODEPAGE\n3\nANSI_1252\n")) {
		t.Error("$DWGCODEPAGE should be written")
	}
	if !bytes.Contains(out, []byte(`\U+58C1`)) {
		t.Error("layer name should be written as \\U+58C1")
	}
	checkJapanese(t, out)
}

func TestWriteShiftJIS(t *testing.T) {
	d := NewDrawing()
	d.Header().CodePage = codepage.Japanese
	d.AddLayer("壁", color.Red, DefaultLineType, true)
	d.Text("日本語テキスト😀", 0, 0, 0, 1)
	var b bytes.Buffer
	d.WriteTo(&b)
	out := b.Bytes()
	if !bytes.Contains(out, []byte("$DWGCODEPAGE\n3\nANSI_932\n")) {
		t.Error("$DWGCODEPAGE should be ANSI_932")
	}
	if !bytes.Contains(out, []byte("\n2\n"+sjis(t, "壁")+"\n")) {
		t.Error("layer name should be written in Shift_JIS")
	}
	// characters outside Shift_JIS are escaped
	if !bytes.Contains(out, []byte(sjis(t, "テキスト")+`\U+D83D\U+DE00`)) {
		t.Error("emoji should be written as \\U+ surrogate pair")
	}
	checkJapanese(t, out)

	d2, _ := FromReader(bytes.NewReader(out))
	if d2.Header().CodePage != codepage.Japanese {
		t.Errorf("code page should be kept: %s", d2.Header().CodePage)
	}
}

func TestWriteUnicodeVersion(t *testing.T) {
	d := NewDrawing()
	d.Header().Version = "AC1021"
	d.Text("日本語", 0, 0, 0, 1)
	var b bytes.Buffer
	d.WriteTo(&b)
	if !strings.Contains(b.String(), "\n1\n日本語\n") {
		t.Error("AC1021 or later should be written in UTF-8")
	}
	d2, err := FromReader(&b)
	if err != nil {
		t.Fatal(err)
	}
	if v := d2.Entities()[0].(*entity.Text).Value; v != "日本語" {
		t.Errorf("got %q", v)
	}
}

func TestReadShiftJIS(t *testing.T) {
	body := "0\nSECTION\n2\nTABLES\n0\nTABLE\n2\nLAYER\n70\n1\n0\nLAYER\n2\n" + sjis(t, "通り芯") + "\n70\n0\n62\n1\n6\nCONTINUOUS\n0\nENDTAB\n0\nENDSEC\n" +
		"0\nSECTION\n2\nENTITIES\n0\nTEXT\n8\n" + sjis(t, "通り芯") + "\n10\n0\n20\n0\n40\n1\n1\n" + sjis(t, "Ｘ１通り") + `\U+00B1` + `\M+193FA` + "\n0\nENDSEC\n0\nEOF\n"
	check := func(name string, tx *entity.Text) {
		if tx.Value != "Ｘ１通り±日" || tx.Layer().Name() != "通り芯" {
			t.Errorf("%s: text %q on layer %q", name, tx.Value, tx.Layer().Name())
		}
	}

	// with $DWGCODEPAGE
	withHeader := "0\nSECTION\n2\nHEADER\n9\n$ACADVER\n1\nAC1015\n9\n$DWGCODEPAGE\n3\nANSI_932\n0\nENDSEC\n" + body
	d, err := FromStringData(withHeader)
	if err != nil {
		t.Fatal(err)
	}
	check("declared", d.Entities()[0].(*entity.Text))

	// without $DWGCODEPAGE (e.g. files from applications omitting it)
	d, err = FromReaderWithCodePage(strings.NewReader(body), codepage.Japanese)
	if err != nil {
		t.Fatal(err)
	}
	check("option", d.Entities()[0].(*entity.Text))
	if d.Header().CodePage != codepage.Japanese {
		t.Errorf("code page option should be kept: %s", d.Header().CodePage)
	}

	// $DWGCODEPAGE in the file takes precedence over the option
	d, err = FromReaderWithCodePage(strings.NewReader(withHeader), "ANSI_1252")
	if err != nil {
		t.Fatal(err)
	}
	check("precedence", d.Entities()[0].(*entity.Text))
}

func TestReadUTF8WithoutCodePage(t *testing.T) {
	// files written by earlier versions of this package: UTF-8 without $DWGCODEPAGE
	d, err := FromStringData(entitiesSection("0\nTEXT\n8\n0\n10\n0\n20\n0\n40\n1\n1\n日本語\n"))
	if err != nil {
		t.Fatal(err)
	}
	if v := d.Entities()[0].(*entity.Text).Value; v != "日本語" {
		t.Errorf("got %q", v)
	}
}
