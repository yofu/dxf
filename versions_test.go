package dxf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
)

func TestVersions(t *testing.T) {
	for _, v := range format.Versions() {
		d := NewDrawing()
		if err := d.SetVersion(v); err != nil {
			t.Fatal(err)
		}
		d.AddLayer("壁", 1, DefaultLineType, true)
		l, _ := d.Line(0, 0, 0, 1, 1, 0)
		l.SetTrueColor(1, 2, 3)
		l.SetTransparency(0.5)
		d.Text("日本語", 0, 0, 0, 1)
		d.BeginBlock("B", 0, 0, 0)
		d.EndBlock()
		out := writeString(t, d)
		checkStructure(t, d)

		if !strings.Contains(out, "$ACADVER\n1\n"+v+"\n") {
			t.Errorf("%s: $ACADVER", v)
		}
		unicode := v >= format.R2007
		if got := strings.Contains(out, "\n1\n日本語\n"); got != unicode {
			t.Errorf("%s: UTF-8 text %v, want %v", v, got, unicode)
		}
		if got := strings.Contains(out, `\U+65E5`); got == unicode {
			t.Errorf("%s: \\U+ escapes %v", v, got)
		}
		if got := strings.Contains(out, "\n420\n"); got != (v >= format.R2004) {
			t.Errorf("%s: true color written %v", v, got)
		}
		if got := strings.Contains(out, "\n440\n"); got != (v >= format.R2004) {
			t.Errorf("%s: transparency written %v", v, got)
		}
		if got := strings.Contains(out, "ObjectDBX Classes\n90\n0\n91\n"); got != (v >= format.R2004) {
			t.Errorf("%s: class instance count written %v", v, got)
		}
		if got := strings.Contains(out, "AcDbBlockTableRecord\n2\nB\n70\n"); got != (v >= format.R2007) {
			t.Errorf("%s: BLOCK_RECORD 70 written %v", v, got)
		}

		for _, bin := range []bool{false, true} {
			var b bytes.Buffer
			if bin {
				d.SetFormatter(format.NewBinary())
			} else {
				d.SetFormatter(format.NewASCII())
				d.Formatter().SetPrecision(16)
			}
			d.WriteTo(&b)
			d2, err := FromReader(&b)
			if err != nil {
				t.Fatal(err)
			}
			if d2.Version() != v {
				t.Errorf("%s: version read %s", v, d2.Version())
			}
			if tx := d2.Entities()[1].(*entity.Text); tx.Value != "日本語" || tx.Layer().Name() != "壁" {
				t.Errorf("%s (binary %v): text %q on %q", v, bin, tx.Value, tx.Layer().Name())
			}
		}
	}
}
