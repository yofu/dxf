package drawing

import (
	"errors"
	"math"

	"github.com/yofu/dxf/entity"
)

// Leader creates a LEADER through points (the arrowhead is at the first
// point) without annotation.
func (d *Drawing) Leader(points ...[]float64) (*entity.Leader, error) {
	if len(points) < 2 {
		return nil, errors.New("Leader needs 2 or more points")
	}
	l := entity.NewLeader()
	for _, p := range points {
		l.Vertices = append(l.Vertices, point3(p))
	}
	d.applyCurrent(l)
	d.AddEntity(l)
	return l, nil
}

// LeaderText creates a LEADER through points with an MTEXT annotation.
// A horizontal hookline (of the dimension style's arrow size) is added at
// the last point, and the text is placed after it. height 0 uses the text
// height of the dimension style.
func (d *Drawing) LeaderText(text string, height float64, points ...[]float64) (*entity.Leader, *entity.MText, error) {
	if len(points) < 2 {
		return nil, nil, errors.New("Leader needs 2 or more points")
	}
	style, err := d.DimStyle("Standard")
	if err != nil {
		return nil, nil, err
	}
	scale := style.Scale
	if scale <= 0 {
		scale = 1
	}
	if height <= 0 {
		height = style.TextHeight * scale
	}
	l := entity.NewLeader()
	l.AnnotationType = entity.LEADER_MTEXT
	l.TextHeight = height
	for _, p := range points {
		l.Vertices = append(l.Vertices, point3(p))
	}
	n := len(points)
	last, prev := l.Vertices[n-1], l.Vertices[n-2]
	dir := 1.0 // the text is on the right if the last segment goes right
	if last[0] < prev[0] {
		dir = -1
	}
	if math.Abs(last[1]-prev[1]) > 1e-12 { // add a horizontal hookline
		hook := []float64{last[0] + dir*style.ArrowSize*scale, last[1], last[2]}
		l.Vertices = append(l.Vertices, hook)
		last = hook
	}
	if dir < 0 {
		l.HorizontalDirection = []float64{-1, 0, 0}
	}
	d.applyCurrent(l)

	t, err := d.MText(text, last[0]+dir*style.TextGap*scale, last[1], last[2], height, 0)
	if err != nil {
		return nil, nil, err
	}
	if dir > 0 {
		t.Attachment = entity.MTEXT_MIDDLE_LEFT
	} else {
		t.Attachment = entity.MTEXT_MIDDLE_RIGHT
	}
	l.Annotation = t
	d.AddEntity(l)
	return l, t, nil
}
