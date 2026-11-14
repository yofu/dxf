package proxygraphic

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/table"
)

// builder writes proxy graphics commands.
type builder struct {
	recs  [][]byte
	cur   []byte
	count int
}

func (b *builder) cmd(typ int) *builder {
	b.end()
	b.cur = binary.LittleEndian.AppendUint32(nil, 0)
	b.cur = binary.LittleEndian.AppendUint32(b.cur, uint32(typ))
	b.count++
	return b
}

func (b *builder) end() {
	if b.cur != nil {
		binary.LittleEndian.PutUint32(b.cur, uint32(len(b.cur)))
		b.recs = append(b.recs, b.cur)
		b.cur = nil
	}
}

func (b *builder) u(v ...int) *builder {
	for _, x := range v {
		b.cur = binary.LittleEndian.AppendUint32(b.cur, uint32(int32(x)))
	}
	return b
}

func (b *builder) f(v ...float64) *builder {
	for _, x := range v {
		b.cur = binary.LittleEndian.AppendUint64(b.cur, math.Float64bits(x))
	}
	return b
}

func (b *builder) s(v string) *builder {
	b.cur = append(b.cur, v...)
	b.cur = append(b.cur, 0)
	for len(b.cur)%4 != 0 {
		b.cur = append(b.cur, 0)
	}
	return b
}

func (b *builder) bytes() []byte {
	b.end()
	var body []byte
	for _, r := range b.recs {
		body = append(body, r...)
	}
	out := binary.LittleEndian.AppendUint32(nil, uint32(len(body)+8))
	out = binary.LittleEndian.AppendUint32(out, uint32(b.count))
	return append(out, body...)
}

func near(a, b []float64) bool {
	for i := range a {
		if math.Abs(a[i]-b[i]) > 1e-9 {
			return false
		}
	}
	return len(a) == len(b)
}

func TestDecode(t *testing.T) {
	l0, l1 := table.NewLayer("0", 7, nil), table.NewLayer("WALL", 1, nil)
	b := &builder{}
	b.cmd(SubentLayer).u(1)
	b.cmd(SubentColor).u(3)
	b.cmd(Circle).f(1, 2, 0, 5, 0, 0, 1)
	b.cmd(SubentTrueColor).u(0xc2102030)
	b.cmd(CircularArc).f(0, 0, 0, 2, 0, 0, 1, 0, 1, 0, math.Pi/2) // 90 to 180
	b.cmd(Polyline).u(2).f(0, 0, 0, 3, 4, 0)
	b.cmd(Polygon).u(3).f(0, 0, 0, 1, 0, 0, 1, 1, 0)
	b.cmd(Polyline).u(3).f(0, 0, 0, 1, 0, 1, 1, 1, 2)
	b.cmd(Text).f(1, 1, 0, 0, 0, 1, 0, 1, 0, 2.5, 1, 0).s("ABC")
	b.cmd(PushMatrix).f(1, 0, 0, 10, 0, 1, 0, 20, 0, 0, 1, 0, 0, 0, 0, 1)
	b.cmd(Circle).f(0, 0, 0, 1, 0, 0, 1)
	b.cmd(PopMatrix)
	b.cmd(Shell).u(4).f(0, 0, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0).u(5, 4, 0, 1, 2, 3)
	b.cmd(SubentMarker).u(7) // skipped
	b.cmd(Ray).f(0, 0, 0, 2, 0, 0)
	es, err := Decode(b.bytes(), Options{Layers: []*table.Layer{l0, l1}})
	if err != nil {
		t.Fatal(err)
	}
	if len(es) != 9 {
		t.Fatalf("entities: %d %v", len(es), es)
	}
	c := es[0].(*entity.Circle)
	if !near(c.Center, []float64{1, 2, 0}) || c.Radius != 5 || c.Layer() != l1 || c.ColorIndex() != 3 {
		t.Errorf("circle: %+v", c)
	}
	a := es[1].(*entity.Arc)
	if a.Radius != 2 || math.Abs(a.Angle[0]-90) > 1e-9 || math.Abs(a.Angle[1]-180) > 1e-9 {
		t.Errorf("arc: %v", a.Angle)
	}
	if l := es[2].(*entity.Line); !near(l.End, []float64{3, 4, 0}) {
		t.Errorf("line: %v", l.End)
	}
	if lw := es[3].(*entity.LwPolyline); !lw.Closed || len(lw.Vertices) != 3 {
		t.Errorf("polygon: %+v", lw)
	}
	if pl := es[4].(*entity.Polyline); pl.Flag&8 == 0 || len(pl.Vertices) != 3 {
		t.Errorf("3d polyline: %+v", pl)
	}
	if tx := es[5].(*entity.Text); tx.Value != "ABC" || tx.Height != 2.5 || math.Abs(tx.Rotation-90) > 1e-9 {
		t.Errorf("text: %+v", tx)
	}
	if c := es[6].(*entity.Circle); !near(c.Center, []float64{10, 20, 0}) {
		t.Errorf("transformed circle: %v", c.Center)
	}
	if m := es[7].(*entity.Mesh); len(m.Faces) != 1 || len(m.Faces[0]) != 4 {
		t.Errorf("shell: %+v", m.Faces)
	}
	if r := es[8].(*entity.XLine); !near(r.Direction, []float64{1, 0, 0}) {
		t.Errorf("ray: %+v", r)
	}
	// truncated data
	data := b.bytes()
	if _, err := Decode(data[:len(data)-5], Options{}); err == nil {
		t.Error("truncated data should be an error")
	}
}

func TestData(t *testing.T) {
	codes := [][2]string{{"100", "AcDbProxyEntity"}, {"90", "498"}, {"92", "8"}, {"310", "0800"}, {"310", "0000"}, {"93", "4"}, {"310", "FFFF"}}
	b, err := Data(codes)
	if err != nil || len(b) != 4 || b[0] != 8 {
		t.Errorf("data: %x %v", b, err)
	}
	if _, err := Data(codes[:2]); err == nil {
		t.Error("no data should be an error")
	}
}
