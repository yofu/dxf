package dxf

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/format"
)

func TestMoreDimensions(t *testing.T) {
	d := NewDrawing()
	a2, err := d.Angular2LineDimension([]float64{0, 0}, []float64{10, 0}, []float64{0, 0}, []float64{10, 10}, []float64{8, 3})
	if err != nil || !cmpF64(a2.Measurement, 45) {
		t.Errorf("2-line angular: %v %v", err, a2)
	}
	ox, err := d.OrdinateDimension([]float64{0, 0}, []float64{30, 20}, []float64{35, 40}, true)
	if err != nil || !cmpF64(ox.Measurement, 30) {
		t.Errorf("ordinate X: %v %v", err, ox)
	}
	oy, _ := d.OrdinateDimension([]float64{0, 0}, []float64{30, 20}, []float64{50, 25}, false)
	if !cmpF64(oy.Measurement, 20) {
		t.Errorf("ordinate Y: %v", oy.Measurement)
	}
	arc, err := d.ArcDimension([]float64{0, 0}, []float64{5, 0}, []float64{0, 5}, []float64{4, 4})
	if err != nil || !cmpF64(arc.Measurement, 5*math.Pi/2) || arc.Type != entity.DIM_ARC {
		t.Errorf("arc length: %v %v", err, arc)
	}
	dims, _ := d.LinearDimensions([][]float64{{0, 0}, {10, 0}, {25, 0}}, []float64{0, -5}, 0)
	if len(dims) != 2 || !cmpF64(dims[1].Measurement, 15) {
		t.Errorf("continued: %v", dims)
	}
	if r, _ := d.RadiusDimensionAt([]float64{0, 0}, 3, 90); !cmpF64(r.Measurement, 3) {
		t.Errorf("radius at: %v", r.Measurement)
	}
	out := writeString(t, d)
	checkStructure(t, d)
	d2 := readPreserve(t, out)
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
}

func TestMTextColumnsAndEditor(t *testing.T) {
	d := NewDrawing()
	d.SetVersion(format.R2018)
	ed := entity.NewMTextEditor("Title").Paragraph().Font("Arial", true, false, "bold").Append(" ").
		Height(2.5, "big").Append(" ").Color(1, "red").Append(" ").Stack("1", "2", entity.STACK_HORIZONTAL).
		Append(" a{b}\\c").ColumnBreak().Underline("second column")
	want := `Title\P{\fArial|b1|i0;bold} {\H2.5;big} {\C1;red} \S1/2; a\{b\}\\c\N\Lsecond column\l`
	if ed.String() != want {
		t.Errorf("editor:\n%s\nwant\n%s", ed.String(), want)
	}
	m, _ := d.MText(ed.String(), 0, 0, 0, 1, 20)
	m.Columns = entity.NewStaticColumns(2, 20, 2, 30)
	out := writeString(t, d)
	if !strings.Contains(out, "\n101\nEmbedded Object\n") {
		t.Error("embedded object not written")
	}
	d2 := readPreserve(t, out)
	r := d2.Entities()[0].(*entity.MText)
	if r.Columns == nil || r.Columns.Count != 2 || r.Columns.Width != 20 || r.Columns.Gutter != 2 || r.Width != 20 || !near(r.Coord, []float64{0, 0, 0}) {
		t.Fatalf("columns: %+v %+v", r.Columns, r)
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
	d.SetVersion(format.R2013)
	if strings.Contains(writeString(t, d), "Embedded Object") {
		t.Error("the embedded object is written for AC1032 only")
	}
	// text placement
	tx, _ := d.Text("A", 0, 0, 0, 1)
	tx.SetPlacement([]float64{5, 5}, nil, "MIDDLE_CENTER")
	if tx.HorizontalFlag != 1 || tx.VerticalFlag != 2 || !near(tx.Coord2, []float64{5, 5, 0}) {
		t.Errorf("placement: %+v", tx)
	}
	if err := tx.SetPlacement([]float64{0, 0}, nil, "FIT"); err == nil {
		t.Error("FIT needs two points")
	}
}

func TestHatchPatternsAndAssociative(t *testing.T) {
	pats, err := entity.ParsePatterns("; comment\n*MYPAT, my pattern\n0, 0,0, 0,5, 3,-2\n90, 0,0, 0,5\n*OTHER\n45, 0,0, 0,1\n")
	if err != nil || len(pats["MYPAT"]) != 2 || len(pats["OTHER"]) != 1 || pats["MYPAT"][0][5] != 3 {
		t.Fatalf("patterns: %v %v", pats, err)
	}
	d := NewDrawing()
	h, _ := d.Hatch([]float64{0, 0}, []float64{10, 0}, []float64{10, 10})
	h.SetPatternDefinition("MYPAT", pats["MYPAT"], 0, 2)
	if len(h.PatternLines) != 2 || h.PatternLines[0].Dashes[0] != 6 {
		t.Errorf("pattern lines: %+v", h.PatternLines[0])
	}
	for _, name := range []string{"DASH", "DOTS", "CROSS", "BRICK", "SQUARE", "ZIGZAG"} {
		if err := h.SetPattern(name, 0, 1); err != nil {
			t.Error(err)
		}
	}
	lw, _ := d.LwPolyline(true, []float64{20, 0}, []float64{40, 0}, []float64{40, 20}, []float64{20, 20})
	c, _ := d.Circle(30, 10, 0, 3)
	ah, err := d.HatchFromEntities(lw, c)
	if err != nil {
		t.Fatal(err)
	}
	out := writeString(t, d)
	checkStructure(t, d)
	if !strings.Contains(out, "\n97\n1\n330\n"+hexHandle(lw.Handle())+"\n") || !strings.Contains(out, "\n102\n{ACAD_REACTORS\n330\n"+hexHandle(ah.Handle())+"\n102\n}\n") {
		t.Error("associative boundaries and reactors should be written")
	}
	d2 := readPreserve(t, out)
	if r := d2.Entities()[3].(*entity.Hatch); !r.Associative || len(r.Boundaries[0].Sources) != 1 || len(r.Boundaries[1].Edges) != 1 {
		t.Errorf("associative hatch read: %+v", r)
	}
	if writeString(t, d2) != out {
		t.Error("writing a read drawing again should give the same output")
	}
}

func TestMoreCreation(t *testing.T) {
	d := NewDrawing()
	mesh, err := d.PolygonMesh(2, 3, [][]float64{{0, 0, 0}, {1, 0, 0}, {2, 0, 1}, {0, 1, 0}, {1, 1, 1}, {2, 1, 0}}, false, true)
	if err != nil || mesh.MeshM != 2 || mesh.MeshN != 3 || mesh.Flag != 16|32 || mesh.Vertices[0].Flag != 64 {
		t.Errorf("mesh: %v %+v", err, mesh)
	}
	pts := [][]float64{{0, 0, 0}, {2, 3, 0}, {5, 1, 0}, {8, 4, 0}, {10, 0, 0}}
	s, err := d.SplineThrough(pts)
	if err != nil {
		t.Fatal(err)
	}
	// the spline passes through the points at their chord length parameters
	total := 0.0
	params := []float64{0}
	for i := 1; i < len(pts); i++ {
		total += math.Hypot(pts[i][0]-pts[i-1][0], pts[i][1]-pts[i-1][1])
		params = append(params, total)
	}
	for i, p := range pts {
		u := params[i] / total
		q := evalSpline(s, u)
		if !near(q, p) {
			t.Errorf("spline at %v: %v, want %v", u, q, p)
		}
	}
	if fit, _ := d.SplineFit(pts); len(fit.Fits) != 5 || len(fit.Controls) != 0 {
		t.Errorf("fit spline: %+v", fit)
	}
	for _, name := range []string{"_DOT", "_OPEN", "_ARCHTICK", "_BOXFILLED", "_NONE"} {
		if _, err := d.ArrowBlock(name); err != nil {
			t.Error(err)
		}
	}
	if _, err := d.ArrowBlock("_UNKNOWN"); err == nil {
		t.Error("unknown arrow should be an error")
	}
	ins, _ := d.InsertArrow("_DOT", []float64{5, 5}, 2.5, 45)
	if ins.BlockName != "_DOT" || ins.Scale[0] != 2.5 {
		t.Errorf("arrow: %+v", ins)
	}
	lts, err := d.LoadLineTypes("*DASH2,Dash __ __\nA,.5,-.25\n*GAS2,Gas ----GAS----\nA,.5,-.2,[\"GAS\",STANDARD,S=.1,R=0.0,X=-0.1,Y=-.05],-.25\n*TRK,Track\nA,.15,[132,ltypeshp.shx,S=.25],.15\n")
	if err != nil || len(lts) != 3 {
		t.Fatalf("linetypes: %v %v", lts, err)
	}
	if e := lts[1].Elements()[1]; e.Flags != entity.MTEXT_TOP_LEFT*0+2 || e.Text != "GAS" || e.Length != -0.2 || e.Offset[0] != -0.1 {
		t.Errorf("text element: %+v", e)
	}
	if e := lts[2].Elements()[0]; e.Shape != 132 || e.Scale != 0.25 {
		t.Errorf("shape element: %+v", e)
	}
	// header variables
	if err := SetHeaderVariable(d, "$PDMODE", [2]string{"70", "3"}); err != nil || d.Header().PdMode != 3 {
		t.Errorf("typed header variable: %v %v", err, d.Header().PdMode)
	}
	SetHeaderVariable(d, "$MIRRTEXT", [2]string{"70", "1"})
	if v, ok := HeaderVariable(d, "$MIRRTEXT"); !ok || !reflect.DeepEqual(v, [][2]string{{"70", "1"}}) {
		t.Errorf("header variable: %v %v", v, ok)
	}
	if v, ok := HeaderVariable(d, "$PDMODE"); !ok || v[0][1] != "3" {
		t.Errorf("typed header variable read: %v", v)
	}
	// typed XDATA
	l, _ := d.Line(0, 0, 0, 1, 1, 0)
	l.SetXDataValues("MYAPP", entity.XString("text"), entity.XReal(1.5), entity.XInt(7), entity.XPoint([]float64{1, 2, 3}))
	l.SetXDataValues("OTHER", entity.XString("other"))
	l.SetXDataValues("MYAPP", entity.XString("new"), entity.XPoint([]float64{4, 5, 6}))
	if v := l.XDataValues("MYAPP"); len(v) != 2 || v[0].Value != "new" || !reflect.DeepEqual(v[1].Value, []float64{4, 5, 6}) {
		t.Errorf("xdata: %+v", v)
	}
	if apps := l.XDataApps(); len(apps) != 2 || apps[0] != "OTHER" {
		t.Errorf("xdata apps: %v", apps)
	}
	rc, err := d.RevisionCloud([][]float64{{0, 0}, {10, 0}, {10, 5}, {0, 5}}, 2, true)
	if err != nil || len(rc.Vertices) != 16 || !rc.Closed || rc.Bulges[0] <= 0 {
		t.Errorf("revcloud: %v %d", err, len(rc.Vertices))
	}
	out := writeString(t, d)
	checkStructure(t, d)
	if !strings.Contains(out, "\n9\n$MIRRTEXT\n70\n1\n") || !strings.Contains(out, "\n2\nRevcloudProps\n") {
		t.Error("header variable and revcloud APPID should be written")
	}
}

// evalSpline evaluates a non-rational spline at u (de Boor).
func evalSpline(s *entity.Spline, u float64) []float64 {
	p := s.Degree
	k := p
	for k+1 < len(s.Knots)-p-1 && s.Knots[k+1] <= u {
		k++
	}
	dd := make([][]float64, p+1)
	for j := 0; j <= p; j++ {
		dd[j] = append([]float64{}, s.Controls[j+k-p]...)
	}
	for r := 1; r <= p; r++ {
		for j := p; j >= r; j-- {
			den := s.Knots[j+1+k-r] - s.Knots[j+k-p]
			a := 0.0
			if den != 0 {
				a = (u - s.Knots[j+k-p]) / den
			}
			for c := 0; c < 3; c++ {
				dd[j][c] = (1-a)*dd[j-1][c] + a*dd[j][c]
			}
		}
	}
	return dd[p]
}
