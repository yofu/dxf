package dxf

import (
	"reflect"
	"strings"
	"testing"

	"github.com/yofu/dxf/class"
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/table"
)

// a file with data this package does not interpret
const preserveSrc = "0\nSECTION\n2\nHEADER\n" +
	"9\n$ACADVER\n1\nAC1015\n9\n$ORTHOMODE\n70\n1\n9\n$FILLETRAD\n40\n5.0\n9\n$UCSORG\n10\n1.0\n20\n2.0\n30\n0.0\n" +
	"9\n$HANDSEED\n5\n100\n9\n$CLAYER\n8\nA\n9\n$PUCSBASE\n2\n\n9\n$SOMEHANDLE\n390\n2F\n" +
	"0\nENDSEC\n" +
	"0\nSECTION\n2\nCLASSES\n0\nCLASS\n1\nIMAGE\n2\nAcDbRasterImage\n3\nISM\n90\n127\n280\n0\n281\n1\n" +
	"0\nCLASS\n1\nIMAGEDEF\n2\nAcDbRasterImageDef\n3\nISM\n90\n0\n280\n0\n281\n0\n0\nENDSEC\n" +
	"0\nSECTION\n2\nTABLES\n" +
	"0\nTABLE\n2\nLAYER\n70\n1\n0\nLAYER\n5\n10\n330\n2\n100\nAcDbSymbolTableRecord\n100\nAcDbLayerTableRecord\n2\nA\n70\n0\n62\n1\n6\nCONTINUOUS\n290\n0\n370\n-3\n390\nF\n1001\nMYAPP\n1000\nlayer data\n0\nENDTAB\n" +
	"0\nTABLE\n2\nSTYLE\n70\n1\n0\nSTYLE\n5\n11\n2\nNARROW\n70\n0\n40\n0\n41\n0.7\n50\n15\n42\n3.5\n3\nromans.shx\n0\nENDTAB\n" +
	"0\nTABLE\n2\nVIEW\n70\n1\n0\nVIEW\n5\n20\n2\nV1\n70\n0\n40\n100\n10\n50\n20\n50\n41\n200\n11\n0\n21\n0\n31\n1\n12\n0\n22\n0\n32\n0\n0\nENDTAB\n" +
	"0\nTABLE\n2\nDIMSTYLE\n70\n1\n0\nDIMSTYLE\n105\n30\n2\nMM\n70\n0\n3\n<> mm\n41\n3\n178\n1\n340\n11\n0\nENDTAB\n" +
	"0\nENDSEC\n" +
	"0\nSECTION\n2\nENTITIES\n" +
	"0\nLINE\n5\n40\n330\n1F\n100\nAcDbEntity\n8\nA\n100\nAcDbLine\n10\n0\n20\n0\n30\n0\n11\n10\n21\n0\n31\n0\n" +
	"1001\nMYAPP\n1000\nline data\n1040\n1.5\n1005\n2F\n1002\n{\n1070\n7\n1002\n}\n" +
	"0\nXLINE\n5\n41\n102\n{ACAD_REACTORS\n330\n42\n102\n}\n330\n1F\n100\nAcDbEntity\n8\nA\n62\n3\n100\nAcDbXline\n10\n5\n20\n6\n30\n0\n11\n1\n21\n0\n31\n0\n" +
	"0\nIMAGE\n5\n42\n100\nAcDbEntity\n8\n0\n100\nAcDbRasterImage\n10\n0\n20\n0\n340\n60\n" +
	"0\nTRACE\n8\n0\n62\n2\n10\n0\n20\n0\n11\n1\n21\n0\n12\n0\n22\n1\n13\n1\n23\n1\n" +
	"0\nPOLYLINE\n5\n2F\n8\n0\n66\n1\n70\n0\n1001\nMYAPP\n1000\npolyline data\n" +
	"0\nVERTEX\n8\n0\n10\n0\n20\n0\n1001\nMYAPP\n1070\n1\n0\nVERTEX\n8\n0\n10\n1\n20\n1\n0\nSEQEND\n8\n0\n" +
	"0\nENDSEC\n" +
	"0\nSECTION\n2\nOBJECTS\n" +
	"0\nDICTIONARY\n5\nC\n330\n0\n100\nAcDbDictionary\n281\n1\n3\nACAD_IMAGE_DICT\n350\n61\n" +
	"0\nDICTIONARY\n5\n61\n102\n{ACAD_REACTORS\n330\nC\n102\n}\n330\nC\n100\nAcDbDictionary\n281\n1\n3\nplan\n350\n60\n" +
	"0\nIMAGEDEF\n5\n60\n102\n{ACAD_REACTORS\n330\n61\n102\n}\n330\n61\n100\nAcDbRasterImageDef\n90\n0\n1\nplan.png\n10\n640\n20\n480\n" +
	"0\nENDSEC\n0\nEOF\n"

func readPreserve(t *testing.T, src string) *drawing.Drawing {
	t.Helper()
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func checkPreserved(t *testing.T, d *drawing.Drawing) {
	t.Helper()
	h := d.Header()
	wantHeader := [][2]string{
		{"9", "$ORTHOMODE"}, {"70", "1"}, {"9", "$FILLETRAD"}, {"40", "5.0"},
		{"9", "$UCSORG"}, {"10", "1.0"}, {"20", "2.0"}, {"30", "0.0"}, {"9", "$PUCSBASE"}, {"2", ""},
		{"9", "$SOMEHANDLE"}, {"390", "2F"},
	}
	if !reflect.DeepEqual(h.Extra, wantHeader) {
		t.Errorf("header extra: %v", h.Extra)
	}
	if h.CLayer != "A" || d.CurrentLayer.Name() != "A" {
		t.Error("$CLAYER")
	}

	l := d.Layers["A"]
	if !reflect.DeepEqual(l.Extra, [][2]string{{"290", "0"}, {"1001", "MYAPP"}, {"1000", "layer data"}}) {
		t.Errorf("layer extra: %v", l.Extra)
	}
	st, _ := d.Style("NARROW", false)
	if st.WidthFactor != 0.7 || st.ObliqueAngle != 15 || st.LastHeightUsed != 3.5 || st.FontName != "romans.shx" {
		t.Errorf("style: %+v", st)
	}
	views := d.Sections[drawing.TABLES].(table.Tables)[table.VIEW].Entries()
	if len(views) != 1 || len(views[0].(*table.View).Extra) != 11 || views[0].(*table.View).Extra[0] != [2]string{"70", "0"} {
		t.Errorf("view extra: %v", views[0].(*table.View).Extra)
	}
	ds, _ := d.DimStyle("MM")
	if ds.ArrowSize != 3 || !reflect.DeepEqual(ds.Extra, [][2]string{{"3", "<> mm"}, {"178", "1"}, {"340", "11"}}) {
		t.Errorf("dimstyle: %v %v", ds.ArrowSize, ds.Extra)
	}

	es := d.Entities()
	if len(es) != 5 {
		t.Fatalf("want LINE, XLINE, IMAGE, TRACE, POLYLINE, got %d", len(es))
	}
	line := es[0].(*entity.Line)
	wantX := [][2]string{{"1001", "MYAPP"}, {"1000", "line data"}, {"1040", "1.5"}, {"1005", "2F"}, {"1002", "{"}, {"1070", "7"}, {"1002", "}"}}
	if !reflect.DeepEqual(line.XData(), wantX) {
		t.Errorf("line xdata: %v", line.XData())
	}
	x, ok := es[1].(*entity.Unknown)
	if !ok || x.TypeName() != "XLINE" || x.ColorIndex() != 3 || x.Layer().Name() != "A" {
		t.Fatalf("xline: %+v", es[1])
	}
	wantData := [][2]string{{"100", "AcDbXline"}, {"10", "5"}, {"20", "6"}, {"30", "0"}, {"11", "1"}, {"21", "0"}, {"31", "0"}}
	if !reflect.DeepEqual(x.Data, wantData) {
		t.Errorf("xline data: %v", x.Data)
	}
	if mins, maxs := x.BBox(); mins[0] != 1 || maxs[0] != 5 || maxs[1] != 6 {
		t.Errorf("xline bbox: %v %v", mins, maxs)
	}
	if g := x.ExtensionGroups(); len(g) != 3 || g[1] != [2]string{"330", "42"} {
		t.Errorf("xline reactors: %v", g)
	}
	img := es[2].(*entity.Unknown)
	if v, _ := value(img.Data, "340"); img.TypeName() != "IMAGE" || v != "60" || img.SourceHandle() != "42" {
		t.Errorf("image: %+v", img.Data)
	}
	classes := map[string]bool{}
	for _, c := range d.Sections[drawing.CLASSES].(class.Classes) {
		classes[c.Name()] = true
	}
	if !classes["IMAGE"] || !classes["IMAGEDEF"] {
		t.Errorf("CLASSES should be kept: %v", classes)
	}
	raws := d.RawObjects()
	if len(raws) != 2 || raws[0].Type != "DICTIONARY" || raws[1].Type != "IMAGEDEF" || raws[1].SourceHandle() != "60" {
		t.Errorf("raw objects: %v", raws)
	}
	tr := es[3].(*entity.Unknown)
	if tr.TypeName() != "TRACE" || tr.ColorIndex() != 2 || len(tr.Data) != 8 {
		t.Errorf("trace (R12 style, without subclass markers): %+v", tr.Data)
	}
	pl := es[4].(*entity.Polyline)
	if len(pl.XData()) != 2 || len(pl.Vertices) != 2 || len(pl.Vertices[0].XData()) != 2 || pl.Vertices[1].XData() != nil {
		t.Errorf("polyline xdata: %v %v", pl.XData(), pl.Vertices[0].XData())
	}
}

func TestRoundTripPreservation(t *testing.T) {
	d := readPreserve(t, preserveSrc)
	checkPreserved(t, d)

	out := writeString(t, d)
	checkStructure(t, d)
	// the application of the extended data is registered
	if !strings.Contains(out, "\n0\nAPPID\n") || !strings.Contains(out, "\n2\nMYAPP\n70\n0\n") {
		t.Error("APPID MYAPP should be added")
	}
	// handles and references are kept
	for _, s := range []string{"\n1005\n2F\n", "\n390\n2F\n", "\n340\n60\n", "{ACAD_REACTORS\n330\n42\n",
		"\n0\nLINE\n5\n40\n", "\n0\nIMAGE\n5\n42\n", "\n0\nPOLYLINE\n5\n2F\n", "\n0\nIMAGEDEF\n5\n60\n",
		"\n3\nACAD_IMAGE_DICT\n350\n61\n", "\n0\nDICTIONARY\n5\nC\n", "\n0\nCLASS\n1\nIMAGE\n"} {
		if !strings.Contains(out, s) {
			t.Errorf("output should contain %q", s)
		}
	}
	if strings.Count(out, "$ORTHOMODE") != 1 || strings.Count(out, "$CLAYER") != 1 || strings.Count(out, "$HANDSEED") != 1 {
		t.Error("header variables should be written once")
	}
	// the extended data follows the entity data, before VERTEX
	if !strings.Contains(out, "polyline data\n0\nVERTEX\n") {
		t.Error("POLYLINE extended data should be before VERTEX")
	}

	// read the output again: everything is kept
	d2 := readPreserve(t, out)
	checkPreserved(t, d2)
	if out2 := writeString(t, d2); out2 != out {
		t.Error("writing a read drawing again should give the same output")
	}
}

func TestSetXData(t *testing.T) {
	d := NewDrawing()
	l, _ := d.Line(0, 0, 0, 1, 1, 0)
	l.SetXData([][2]string{{"1001", "STRUCT"}, {"1000", "beam B1"}, {"1040", "300"}})
	out := writeString(t, d)
	if !strings.Contains(out, "\n1001\nSTRUCT\n1000\nbeam B1\n1040\n300\n") || !strings.Contains(out, "\n2\nSTRUCT\n") {
		t.Error("xdata and its APPID should be written")
	}
	checkStructure(t, d)
	d2 := readPreserve(t, out)
	if x := d2.Entities()[0].(*entity.Line).XData(); len(x) != 3 || x[1][1] != "beam B1" {
		t.Errorf("xdata round trip: %v", x)
	}
}
