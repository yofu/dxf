package dxf

import (
	"strings"
	"testing"

	"github.com/yofu/dxf/entity"
)

func TestLeader(t *testing.T) {
	d := NewDrawing()
	l, err := d.Leader([]float64{0, 0, 0}, []float64{10, 10, 0}, []float64{20, 10, 0})
	if err != nil {
		t.Fatal(err)
	}
	if !l.Arrow || l.AnnotationType != entity.LEADER_NONE || len(l.Vertices) != 3 {
		t.Errorf("leader: %+v", l)
	}
	if strings.Contains(l.String(), "\n340\n") {
		t.Error("leader without annotation should not have 340")
	}
	if _, err := d.Leader([]float64{0, 0, 0}); err == nil {
		t.Error("leader with 1 point should be an error")
	}

	lt, mt, err := d.LeaderText("注記テキスト", 0, []float64{100, 0, 0}, []float64{110, 10, 0})
	if err != nil {
		t.Fatal(err)
	}
	// a hookline of arrow size 2.5 is added, and the text follows after the gap 0.625
	if len(lt.Vertices) != 3 || lt.Vertices[2][0] != 112.5 || lt.Vertices[2][1] != 10 {
		t.Errorf("hookline: %v", lt.Vertices)
	}
	if mt.Coord[0] != 113.125 || mt.Coord[1] != 10 || mt.Attachment != entity.MTEXT_MIDDLE_LEFT || mt.Height != 2.5 {
		t.Errorf("text: %v attachment %d height %v", mt.Coord, mt.Attachment, mt.Height)
	}
	if lt.Annotation != mt || lt.AnnotationType != entity.LEADER_MTEXT || lt.TextHeight != 2.5 {
		t.Errorf("annotation: %+v", lt)
	}
	left, lmt, _ := d.LeaderText("left", 3, []float64{0, 0, 0}, []float64{-10, 0, 0})
	if len(left.Vertices) != 2 || lmt.Attachment != entity.MTEXT_MIDDLE_RIGHT || lmt.Coord[0] != -10.625 || left.HorizontalDirection[0] != -1 {
		t.Errorf("leftward leader: %v, text %v", left.Vertices, lmt.Coord)
	}
	checkStructure(t, d) // 340 refers to the MTEXT

	d2, err := FromStringData(writeString(t, d))
	if err != nil {
		t.Fatal(err)
	}
	checkStructure(t, d2)
	var leaders []*entity.Leader
	for _, e := range d2.Entities() {
		if l, ok := e.(*entity.Leader); ok {
			leaders = append(leaders, l)
		}
	}
	if len(leaders) != 3 || leaders[0].Annotation != nil || len(leaders[1].Vertices) != 3 {
		t.Fatalf("leaders round trip: %d", len(leaders))
	}
	a, ok := leaders[1].Annotation.(*entity.MText)
	if !ok || a.Value != "注記テキスト" {
		t.Errorf("annotation round trip: %+v", leaders[1].Annotation)
	}
}

func TestReadLeader(t *testing.T) {
	// the annotation MTEXT may follow the LEADER
	src := entitiesSection("0\nLEADER\n5\n2A\n8\n0\n100\nAcDbLeader\n3\nISO-25\n71\n1\n72\n1\n73\n0\n74\n1\n75\n1\n40\n3.5\n41\n20\n76\n2\n" +
		"10\n0\n20\n0\n30\n0\n10\n5\n20\n5\n30\n0\n340\n2B\n211\n1\n221\n0\n231\n0\n" +
		"0\nMTEXT\n5\n2B\n8\n0\n10\n6\n20\n5\n40\n3.5\n1\nNOTE\n")
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	l := d.Entities()[0].(*entity.Leader)
	if l.Style != "ISO-25" || !l.Arrow || !l.Spline || l.AnnotationType != 0 || l.HooklineDirection != 1 || !l.Hookline ||
		l.TextHeight != 3.5 || l.TextWidth != 20 || len(l.Vertices) != 2 || l.Vertices[1][0] != 5 {
		t.Errorf("leader: %+v", l)
	}
	if l.Annotation != d.Entities()[1] {
		t.Error("annotation should be resolved to the MTEXT")
	}
}
