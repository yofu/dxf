package drawing

import (
	"errors"
	"fmt"
	"strings"

	"github.com/yofu/dxf/entity"
)

// AddMLeaderStyle adds an MLEADERSTYLE with the defaults of AutoCAD
// (dictionary ACAD_MLEADERSTYLE); its text style is the current one.
func (d *Drawing) AddMLeaderStyle(name string) (*entity.MLeaderStyle, error) {
	if s, err := d.MLeaderStyle(name); err == nil {
		return s, fmt.Errorf("mleader style %s already exists", name)
	}
	s := entity.NewMLeaderStyle(name)
	s.TextStyle = entity.Ref{Object: d.CurrentStyle}
	return s, d.AddMLeaderStyleObject(s)
}

// AddMLeaderStyleObject adds an MLEADERSTYLE (used when reading).
func (d *Drawing) AddMLeaderStyleObject(s *entity.MLeaderStyle) error {
	dict := d.namedDictionary("ACAD_MLEADERSTYLE")
	if err := dict.AddItem(s.Name(), s); err != nil {
		return err
	}
	s.SetOwner(dict)
	d.addObject(s)
	d.mleaderStyles = append(d.mleaderStyles, s)
	return nil
}

// MLeaderStyle returns the MLEADERSTYLE name (case-insensitive).
func (d *Drawing) MLeaderStyle(name string) (*entity.MLeaderStyle, error) {
	for _, s := range d.mleaderStyles {
		if strings.EqualFold(s.Name(), name) {
			return s, nil
		}
	}
	return nil, fmt.Errorf("mleader style %s doesn't exist", name)
}

// MLeaderStyles returns the MLEADERSTYLE objects.
func (d *Drawing) MLeaderStyles() []*entity.MLeaderStyle {
	return d.mleaderStyles
}

func (d *Drawing) standardMLeaderStyle() *entity.MLeaderStyle {
	if s, err := d.MLeaderStyle("Standard"); err == nil {
		return s
	}
	s, _ := d.AddMLeaderStyle("Standard")
	s.TextStyle = entity.Ref{Object: d.Styles["STANDARD"]}
	if st, ok := d.Styles["Standard"]; ok {
		s.TextStyle = entity.Ref{Object: st}
	}
	return s
}

// MultiLeaderText creates a MULTILEADER with MTEXT content: the text
// (left aligned) has its top left corner at insert and the char height
// (0: by style). Each leader is a polyline from its arrow head towards the
// text; it is connected to the left or right side of the text, whichever
// its last point is on, with a landing and a dogleg as ezdxf builds them.
// The size of the text is estimated (entity.EstimateTextSize).
func (d *Drawing) MultiLeaderText(content string, insert []float64, charHeight float64, leaders ...[][]float64) (*entity.MultiLeader, error) {
	if len(leaders) == 0 {
		return nil, errors.New("multileader needs a leader")
	}
	style := d.standardMLeaderStyle()
	m := entity.NewMultiLeader()
	m.Style = entity.Ref{Object: style}
	m.OverrideFlags = 0x7FFFFFFF
	m.TextStyle = style.TextStyle
	if lt, err := d.LineType("ByLayer"); err == nil {
		m.LeaderLineType = entity.Ref{Object: lt}
	}
	m.DoglegLength = style.DoglegLength
	m.ArrowSize = style.ArrowSize
	c := m.Context
	if charHeight <= 0 {
		charHeight = style.CharHeight
	}
	c.CharHeight = charHeight
	c.ArrowSize = style.ArrowSize
	c.LandingGap = style.LandingGap
	ins := []float64{coordAt(insert, 0), coordAt(insert, 1), coordAt(insert, 2)}
	c.MText = &entity.MLeaderMText{
		Content: content, Extrusion: []float64{0, 0, 1}, Style: style.TextStyle,
		Insert: ins, Direction: []float64{1, 0, 0}, LineSpacingFactor: 1, LineSpacingStyle: 1,
		Color: m.TextColor, Alignment: 1, FlowDirection: 1, BgColor: -939524096, BgScale: 1.5,
	}
	c.PlaneOrigin = append([]float64{}, ins...)
	// connection box (left aligned text, middle of the top line)
	width, height := entity.EstimateTextSize(content, charHeight, 1)
	_ = height
	gap := c.LandingGap
	y := ins[1] - charHeight*0.5
	left := []float64{ins[0] - gap, y, ins[2]}
	right := []float64{ins[0] + width + gap, y, ins[2]}
	c.BasePoint = append([]float64{}, left...)
	cx := ins[0] + width/2
	for i, pts := range leaders {
		if len(pts) == 0 {
			return nil, errors.New("multileader: empty leader")
		}
		last := pts[len(pts)-1]
		conn, dir := left, 1.0 // left side: the dogleg points to +X
		if coordAt(last, 0) > cx {
			conn, dir = right, -1.0
		}
		l := &entity.MLeaderLeader{
			HasLastLine: true, HasDogleg: true, Index: i, DoglegLength: m.DoglegLength,
			Dogleg:    []float64{dir, 0, 0},
			LastPoint: []float64{conn[0] - dir*m.DoglegLength, conn[1], conn[2]},
		}
		line := &entity.MLeaderLine{Index: 0, Color: m.LeaderColor}
		for _, p := range pts {
			line.Vertices = append(line.Vertices, []float64{coordAt(p, 0), coordAt(p, 1), coordAt(p, 2)})
		}
		l.Lines = []*entity.MLeaderLine{line}
		c.Leaders = append(c.Leaders, l)
	}
	d.add(m)
	return m, nil
}

// linkMLeaders gives MULTILEADER entities without references the
// "Standard" style, its text style and the linetype ByLayer.
func (d *Drawing) linkMLeaders() {
	d.forEachEntity(func(e entity.Entity) {
		m, ok := e.(*entity.MultiLeader)
		if !ok {
			return
		}
		if !m.Style.IsSet() {
			s := d.standardMLeaderStyle()
			m.Style = entity.Ref{Object: s}
			if !m.TextStyle.IsSet() {
				m.TextStyle = s.TextStyle
			}
		}
		if !m.LeaderLineType.IsSet() {
			if lt, err := d.LineType("ByLayer"); err == nil {
				m.LeaderLineType = entity.Ref{Object: lt}
			}
		}
		if c := m.Context; c != nil && c.MText != nil && !c.MText.Style.IsSet() {
			c.MText.Style = m.TextStyle
		}
	})
}
