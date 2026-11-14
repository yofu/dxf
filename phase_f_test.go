package dxf

import (
	"bytes"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

func dxfString(t *testing.T, d *drawing.Drawing) string {
	t.Helper()
	var b bytes.Buffer
	if _, err := d.WriteTo(&b); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func TestComments(t *testing.T) {
	d := NewDrawing()
	d.Comments = []string{"made by yofu/dxf", "second line"}
	d.Line(0, 0, 0, 1, 1, 0)
	s := dxfString(t, d)
	if !strings.HasPrefix(s, "999\nmade by yofu/dxf\n999\nsecond line\n") {
		t.Fatalf("comments not written: %q", s[:60])
	}
	d2, err := FromStringData(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(d2.Comments) != 2 || d2.Comments[1] != "second line" {
		t.Errorf("comments read: %v", d2.Comments)
	}
}

func TestIterEntities(t *testing.T) {
	d := NewDrawing()
	for i := 0; i < 5; i++ {
		d.Line(float64(i), 0, 0, float64(i), 1, 0)
	}
	d.Circle(0, 0, 0, 2)
	s := dxfString(t, d)
	var types []string
	err := IterEntities(strings.NewReader(s), func(_ *drawing.Drawing, e entity.Entity) error {
		types = append(types, TypeName(e))
		return nil
	})
	if err != nil || len(types) != 6 || types[5] != "CIRCLE" {
		t.Fatalf("iter: %v %v", err, types)
	}
	stop := errors.New("stop")
	n := 0
	err = IterEntities(strings.NewReader(s), func(_ *drawing.Drawing, e entity.Entity) error {
		n++
		if n == 2 {
			return stop
		}
		return nil
	})
	if !errors.Is(err, stop) || n != 2 {
		t.Errorf("stop: %v %d", err, n)
	}
}

func TestRecover(t *testing.T) {
	d := NewDrawing()
	d.Line(0, 0, 0, 1, 0, 0)
	d.Circle(5, 5, 0, 1)
	d.Line(2, 0, 0, 3, 0, 0)
	s := dxfString(t, d)
	// break the circle radius and cut the end of the file
	i := strings.Index(s, "CIRCLE")
	j := strings.Index(s[i:], "\n 40\n") + i
	if j < i {
		j = strings.Index(s[i:], "\n40\n") + i
	}
	broken := s[:j] + "\n 40\nabc" + s[strings.Index(s[j+5:], "\n")+j+5:]
	broken = broken[:strings.LastIndex(broken, "ENDSEC")]
	if _, err := FromStringData(broken); err == nil {
		t.Fatal("the damaged file should not be read without recovery")
	}
	r, problems, err := Recover(strings.NewReader(broken))
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) == 0 {
		t.Error("no problems reported")
	}
	lines := 0
	for _, e := range r.Entities() {
		if _, ok := e.(*entity.Line); ok {
			lines++
		}
	}
	if lines != 2 {
		t.Errorf("lines recovered: %d of %v", lines, r.Entities())
	}
}

func TestAudit(t *testing.T) {
	d := NewDrawing()
	d.Line(0, 0, 0, 1, 0, 0)
	c, _ := d.Circle(0, 0, 0, 1)
	c.Radius = 0
	l, _ := d.Line(0, 0, 0, math.NaN(), 0, 0)
	ins := entity.NewInsert("MISSING")
	d.AddEntity(ins)
	tx, _ := d.Text("x", 0, 0, 0, 1)
	tx.Height = 0
	problems := Audit(d, false)
	if len(problems) != 4 {
		t.Fatalf("problems: %v", problems)
	}
	if len(d.Entities()) != 5 {
		t.Error("audit without fix changed the drawing")
	}
	problems = Audit(d, true)
	for _, p := range problems {
		if !p.Fixed {
			t.Errorf("not fixed: %v", p)
		}
	}
	for _, e := range d.Entities() {
		if e == c || e == l || e == ins {
			t.Errorf("not removed: %v", e)
		}
	}
	if tx.Height != 1 || len(Audit(d, false)) != 0 {
		t.Error("problems remain")
	}
}

func TestImportEntities(t *testing.T) {
	src := NewDrawing()
	src.AddLayer("WALL", color.Red, DefaultLineType, true)
	src.BeginBlock("DOOR", 0, 0, 0)
	src.Line(0, 0, 0, 1, 0, 0)
	src.EndBlock()
	ins, _ := src.Insert("DOOR", 3, 0, 0)
	dim, err := src.LinearDimension([]float64{0, 0, 0}, []float64{10, 0, 0}, []float64{5, 2, 0}, 0)
	if err != nil {
		t.Fatal(err)
	}
	dst := NewDrawing()
	dst.BeginBlock("*D1", 0, 0, 0) // an anonymous block name already used
	dst.EndBlock()
	out, err := ImportEntities(dst, src, ins, dim)
	if err != nil || len(out) != 2 {
		t.Fatalf("import: %v %v", err, out)
	}
	if l, err := dst.Layer("WALL", false); err != nil || l.Color != color.Red {
		t.Errorf("layer: %v", err)
	}
	if b, err := dst.Block("DOOR"); err != nil || len(b.Entities) != 1 {
		t.Errorf("block: %v", err)
	}
	d2 := out[1].(*entity.Dimension)
	if d2 == dim || d2.BlockName == dim.BlockName && dim.BlockName == "*D1" {
		t.Errorf("dimension block not renamed: %s", d2.BlockName)
	}
	if _, err := dst.Block(d2.BlockName); err != nil {
		t.Errorf("dimension block: %v", err)
	}
	if _, err := FromStringData(dxfString(t, dst)); err != nil {
		t.Errorf("reread: %v", err)
	}
}

func TestXref(t *testing.T) {
	d := NewDrawing()
	ins, err := AttachXref(d, "PLAN", "plan.dxf", []float64{1, 2}, false)
	if err != nil || ins.BlockName != "PLAN" {
		t.Fatal(err)
	}
	d2, err := FromStringData(dxfString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	b, err := d2.Block("PLAN")
	if err != nil || b.Flag&BlockXref == 0 || b.XrefPath != "plan.dxf" {
		t.Fatalf("xref block: %v %+v", err, b)
	}
	err = BindXref(d2, "PLAN", func(path string) (*drawing.Drawing, error) {
		if path != "plan.dxf" {
			t.Errorf("path %s", path)
		}
		p := NewDrawing()
		p.AddLayer("P", color.Green, DefaultLineType, true)
		p.Line(0, 0, 0, 10, 0, 0)
		p.Circle(0, 0, 0, 3)
		return p, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	b, _ = d2.Block("PLAN")
	if b.Flag&BlockXref != 0 || len(b.Entities) != 2 {
		t.Errorf("bound block: %+v", b)
	}
	if _, err := d2.Layer("P", false); err != nil {
		t.Error("layer of the xref not imported")
	}
	if err := BindXref(d2, "PLAN", nil); err == nil {
		t.Error("binding a block twice should be an error")
	}
}

func TestWriteR12(t *testing.T) {
	d := NewDrawing()
	d.AddLayer("A", color.Red, DefaultLineType, true)
	lw, _ := d.LwPolyline(true, []float64{0, 0}, []float64{4, 0}, []float64{4, 4})
	lw.Bulges[1] = 0.5
	d.MText("hello\\Pworld", 0, 10, 0, 2.5, 0)
	d.Ellipse([]float64{0, 0, 0}, []float64{4, 0, 0}, 0.5)
	d.LinearDimension([]float64{0, 0, 0}, []float64{10, 0, 0}, []float64{5, 2, 0}, 0)
	d.Mesh([][]float64{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}, {0.5, 2, 0}}, [][]int{{0, 1, 2, 3}, {0, 1, 2, 3, 4}})
	var b bytes.Buffer
	if err := WriteR12(d, &b); err != nil {
		t.Fatal(err)
	}
	s := b.String()
	for _, want := range []string{"AC1009", "POLYLINE", "VERTEX", "SEQEND", "TEXT", "$MODEL_SPACE"} {
		if !strings.Contains(s, want) {
			t.Errorf("%s not written", want)
		}
	}
	for _, bad := range []string{"LWPOLYLINE", "MTEXT", "ELLIPSE", "AcDb", "\n330\n"} {
		if strings.Contains(s, bad) {
			t.Errorf("%q written in R12", bad)
		}
	}
	r, err := FromStringData(s)
	if err != nil {
		t.Fatalf("reread: %v", err)
	}
	if len(r.Entities()) == 0 {
		t.Error("no entities read")
	}
}

// fakeODA acts as the ODA File Converter when the test binary is run with
// DXF_FAKE_ODA set: it copies the input file to the output folder.
func fakeODA() {
	a := os.Args[1:]
	if len(a) != 7 {
		os.Exit(3)
	}
	in, out, outType, filter := a[0], a[1], a[3], a[6]
	b, err := os.ReadFile(filepath.Join(in, filter))
	if err != nil {
		os.Exit(4)
	}
	name := strings.TrimSuffix(filter, filepath.Ext(filter)) + "." + strings.ToLower(outType)
	if err := os.WriteFile(filepath.Join(out, name), b, 0o644); err != nil {
		os.Exit(5)
	}
	os.Exit(0)
}

func TestMain(m *testing.M) {
	if os.Getenv("DXF_FAKE_ODA") != "" {
		fakeODA()
	}
	os.Exit(m.Run())
}

func TestODAFileConverter(t *testing.T) {
	for v, want := range map[string]string{"": "ACAD2018", "R12": "ACAD12", "AC1018": "ACAD2004", "r2010": "ACAD2010", "ACAD2013": "ACAD2013"} {
		if got, err := odaVersion(v); err != nil || got != want {
			t.Errorf("version %q: %q %v", v, got, err)
		}
	}
	if _, err := odaVersion("R14"); err == nil {
		t.Error("R14 should be an error")
	}
	if a := odaArgs("in", "out", "ACAD2018", "DXF", "x.dwg"); strings.Join(a, " ") != "in out ACAD2018 DXF 0 1 x.dwg" {
		t.Errorf("args: %v", a)
	}
	old := ODAFileConverter
	defer func() { ODAFileConverter = old }()
	ODAFileConverter = filepath.Join(t.TempDir(), "missing.exe")
	if _, err := ReadDWG("x.dwg"); !errors.Is(err, ErrNoODAFileConverter) {
		t.Errorf("missing converter: %v", err)
	}
	// the test binary as converter
	exe, err := os.Executable()
	if err != nil {
		t.Skip(err)
	}
	ODAFileConverter = exe
	t.Setenv("DXF_FAKE_ODA", "1")
	dir := t.TempDir()
	d := NewDrawing()
	d.Line(0, 0, 0, 3, 4, 0)
	dwg := filepath.Join(dir, "図面.dwg")
	if err := SaveAsDWG(d, dwg); err != nil {
		t.Fatal(err)
	}
	r, err := ReadDWG(dwg)
	if err != nil || len(r.Entities()) != 1 {
		t.Fatalf("read: %v", err)
	}
	if err := ConvertDWG(dwg, filepath.Join(dir, "out.dxf"), "R2000"); err != nil {
		t.Fatal(err)
	}
	if _, err := FromFile(filepath.Join(dir, "out.dxf")); err != nil {
		t.Error(err)
	}
}

func TestProxyGraphics(t *testing.T) {
	// a circle of radius 5 at (1, 2) on the second layer
	g := "0000000002000000" + // size (patched), count
		"0C00000010000000" + "01000000" + // layer 1
		"4000000002000000" + // circle: 64 bytes
		"000000000000F03F" + "0000000000000040" + "0000000000000000" + // center
		"0000000000001440" + // radius
		"0000000000000000" + "0000000000000000" + "000000000000F03F" // normal
	g = "54000000" + g[8:] // 84 bytes
	d := NewDrawing()
	d.AddLayer("WALL", color.Red, DefaultLineType, false)
	d.Line(0, 0, 0, 1, 0, 0)
	s := dxfString(t, d)
	for _, typ := range []string{"ACAD_PROXY_ENTITY", "MY_CUSTOM"} {
		var ent string
		if typ == "ACAD_PROXY_ENTITY" {
			ent = "  0\nACAD_PROXY_ENTITY\n  5\nABC\n100\nAcDbEntity\n  8\n0\n100\nAcDbProxyEntity\n 90\n498\n 91\n500\n 95\n0\n 70\n0\n 92\n84\n310\n" + g + "\n 93\n0\n"
		} else {
			ent = "  0\nMY_CUSTOM\n  5\nABD\n100\nAcDbEntity\n  8\n0\n 92\n84\n310\n" + g + "\n100\nAcDbMyCustom\n 40\n1.0\n"
		}
		i := strings.Index(s, "ENTITIES")
		j := strings.Index(s[i:], "\n0\nENDSEC") + i + 1
		r, err := FromStringData(s[:j] + ent + s[j:])
		if err != nil {
			t.Fatalf("%s: %v", typ, err)
		}
		es := r.Entities()
		ves, err := VirtualEntities(r, es[len(es)-1])
		if err != nil || len(ves) != 1 {
			t.Fatalf("%s: %v %v", typ, err, ves)
		}
		c := ves[0].(*entity.Circle)
		if c.Radius != 5 || !near(c.Center, []float64{1, 2, 0}) || c.Layer().Name() != "WALL" {
			t.Errorf("%s: %+v", typ, c)
		}
		// written back unchanged
		if out := dxfString(t, r); !strings.Contains(out, g) {
			t.Errorf("%s: proxy graphics not written", typ)
		}
	}
}
