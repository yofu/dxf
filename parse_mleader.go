package dxf

import (
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// ParseMultiLeader parses MULTILEADER (MLEADER) entities. References by
// handle (styles, linetype, blocks) are kept as read.
func ParseMultiLeader(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	m := entity.NewMultiLeader()
	head, body := splitAt(data, "AcDbMLeader")
	for _, dt := range head {
		if err := setLayerAndLtscale(d, m, dt); err != nil {
			return m, err
		}
	}
	body, _ = splitXData(body)
	r := &groupReader{data: body}
	var err error
	var attrib *entity.MLeaderAttrib
	ref := func() entity.Ref { return entity.Ref{Raw: strings.TrimSpace(r.next()[1])} }
	b := func() bool { v, _ := r.int(r.peek(), 0); return v != 0 }
	for !r.done() && err == nil {
		switch r.peek() {
		case "270":
			m.Version, err = r.int("270", 2)
		case "300":
			r.next()
			m.Context, err = parseMLeaderContext(r)
		case "340":
			m.Style = ref()
		case "90":
			m.OverrideFlags, err = r.int("90", 0)
		case "170":
			m.LeaderType, err = r.int("170", 1)
		case "91":
			m.LeaderColor, err = r.int("91", 0)
		case "341":
			m.LeaderLineType = ref()
		case "171":
			m.LeaderLineWeight, err = r.int("171", -2)
		case "290":
			m.HasLanding = b()
		case "291":
			m.HasDogleg = b()
		case "41":
			m.DoglegLength, err = r.float("41", 8)
		case "342":
			m.Arrow = ref()
		case "42":
			m.ArrowSize, err = r.float("42", 4)
		case "172":
			m.ContentType, err = r.int("172", 2)
		case "343":
			m.TextStyle = ref()
		case "173":
			m.TextLeftAttachment, err = r.int("173", 1)
		case "95":
			m.TextRightAttachment, err = r.int("95", 1)
		case "174":
			m.TextAngleType, err = r.int("174", 1)
		case "175":
			m.TextAlignmentType, err = r.int("175", 0)
		case "92":
			m.TextColor, err = r.int("92", 0)
		case "292":
			m.HasTextFrame = b()
		case "344":
			m.Block = ref()
		case "93":
			m.BlockColor, err = r.int("93", 0)
		case "10", "20", "30":
			err = setPoint(r.next(), m.BlockScale)
		case "43":
			m.BlockRotation, err = r.float("43", 0)
		case "176":
			m.BlockConnectionType, err = r.int("176", 0)
		case "293":
			m.Annotative = b()
		case "94":
			var i int
			if i, err = r.int("94", 0); err == nil {
				var h entity.Ref
				if r.peek() == "345" {
					h = ref()
				}
				m.ArrowHeads = append(m.ArrowHeads, entity.MLeaderArrowHead{Index: i, Block: h})
			}
		case "330":
			m.BlockAttribs = append(m.BlockAttribs, entity.MLeaderAttrib{AttDef: ref()})
			attrib = &m.BlockAttribs[len(m.BlockAttribs)-1]
		case "177":
			var v int
			v, err = r.int("177", 0)
			if attrib != nil {
				attrib.Index = v
			}
		case "44":
			var v float64
			v, err = r.float("44", 0)
			if attrib != nil {
				attrib.Width = v
			}
		case "302":
			v := r.next()[1]
			if attrib != nil {
				attrib.Text = v
			}
		case "294":
			m.TextDirectionNegative = b()
		case "178":
			m.TextIPEAlign, err = r.int("178", 0)
		case "179":
			m.TextAttachmentPoint, err = r.int("179", 1)
		case "45":
			m.OverallScale, err = r.float("45", 1)
		case "271":
			m.TextAttachmentDir, err = r.int("271", 0)
		case "272":
			m.TextBottomAttachment, err = r.int("272", 9)
		case "273":
			m.TextTopAttachment, err = r.int("273", 9)
		case "295":
			m.ExtendToText = b()
		default:
			r.next()
		}
	}
	return m, err
}

// parseMLeaderContext parses CONTEXT_DATA{ ... } after 300 until 301.
func parseMLeaderContext(r *groupReader) (*entity.MLeaderContext, error) {
	c := entity.NewMLeaderContext()
	var err error
	var x *entity.MLeaderMText
	var blk *entity.MLeaderBlock
	b := func() bool { v, _ := r.int(r.peek(), 0); return v != 0 }
	mtext := func() *entity.MLeaderMText {
		if x == nil {
			x = &entity.MLeaderMText{Extrusion: []float64{0, 0, 1}, Insert: []float64{0, 0, 0}, Direction: []float64{1, 0, 0}, LineSpacingFactor: 1}
			c.MText = x
		}
		return x
	}
	block := func() *entity.MLeaderBlock {
		if blk == nil {
			blk = &entity.MLeaderBlock{Extrusion: []float64{0, 0, 1}, Insert: []float64{0, 0, 0}, Scale: []float64{1, 1, 1}}
			c.Block = blk
		}
		return blk
	}
	for !r.done() && err == nil {
		switch r.peek() {
		case "301":
			r.next()
			return c, nil
		case "40":
			c.Scale, err = r.float("40", 1)
		case "10", "20", "30":
			err = setPoint(r.next(), c.BasePoint)
		case "41":
			c.CharHeight, err = r.float("41", 4)
		case "140":
			c.ArrowSize, err = r.float("140", 4)
		case "145":
			c.LandingGap, err = r.float("145", 2)
		case "174":
			c.LeftAttachment, err = r.int("174", 1)
		case "175":
			c.RightAttachment, err = r.int("175", 1)
		case "176":
			c.TextAlign, err = r.int("176", 0)
		case "177":
			c.AttachmentType, err = r.int("177", 0)
		case "290":
			if b() {
				mtext()
			}
		case "296":
			if b() {
				block()
			}
		case "304":
			mtext().Content = r.next()[1]
		case "11", "21", "31":
			err = setPoint(r.next(), mtext().Extrusion)
		case "340":
			mtext().Style = entity.Ref{Raw: strings.TrimSpace(r.next()[1])}
		case "12", "22", "32":
			err = setPoint(r.next(), mtext().Insert)
		case "13", "23", "33":
			err = setPoint(r.next(), mtext().Direction)
		case "42":
			mtext().Rotation, err = r.float("42", 0)
		case "43":
			mtext().Width, err = r.float("43", 0)
		case "44":
			mtext().DefinedHeight, err = r.float("44", 0)
		case "45":
			mtext().LineSpacingFactor, err = r.float("45", 1)
		case "170":
			mtext().LineSpacingStyle, err = r.int("170", 1)
		case "90":
			mtext().Color, err = r.int("90", 0)
		case "171":
			mtext().Alignment, err = r.int("171", 1)
		case "172":
			mtext().FlowDirection, err = r.int("172", 1)
		case "91":
			mtext().BgColor, err = r.int("91", 0)
		case "141":
			mtext().BgScale, err = r.float("141", 1.5)
		case "92":
			mtext().BgTransparency, err = r.int("92", 0)
		case "291":
			mtext().UseWindowBgColor = b()
		case "292":
			mtext().HasBgFill = b()
		case "173":
			mtext().ColumnType, err = r.int("173", 0)
		case "293":
			mtext().UseAutoHeight = b()
		case "142":
			mtext().ColumnWidth, err = r.float("142", 0)
		case "143":
			mtext().ColumnGutterWidth, err = r.float("143", 0)
		case "294":
			mtext().ColumnFlowReversed = b()
		case "144":
			var v float64
			v, err = r.float("144", 0)
			mtext().ColumnSizes = append(mtext().ColumnSizes, v)
		case "295":
			mtext().UseWordBreak = b()
		case "341":
			block().Block = entity.Ref{Raw: strings.TrimSpace(r.next()[1])}
		case "14", "24", "34":
			err = setPoint(r.next(), block().Extrusion)
		case "15", "25", "35":
			err = setPoint(r.next(), block().Insert)
		case "16", "26", "36":
			err = setPoint(r.next(), block().Scale)
		case "46":
			block().Rotation, err = r.float("46", 0)
		case "93":
			block().Color, err = r.int("93", 0)
		case "47":
			var v float64
			v, err = r.float("47", 0)
			block().Matrix = append(block().Matrix, v)
		case "110", "120", "130":
			err = setPoint(r.next(), c.PlaneOrigin)
		case "111", "121", "131":
			err = setPoint(r.next(), c.PlaneX)
		case "112", "122", "132":
			err = setPoint(r.next(), c.PlaneY)
		case "297":
			c.NormalReversed = b()
		case "302":
			r.next()
			var l *entity.MLeaderLeader
			if l, err = parseMLeaderLeader(r); err == nil {
				c.Leaders = append(c.Leaders, l)
			}
		case "272":
			c.TopAttachment, err = r.int("272", 9)
		case "273":
			c.BottomAttachment, err = r.int("273", 9)
		default:
			r.next()
		}
	}
	return c, err
}

// parseMLeaderLeader parses LEADER{ ... } after 302 until 303.
func parseMLeaderLeader(r *groupReader) (*entity.MLeaderLeader, error) {
	l := &entity.MLeaderLeader{LastPoint: []float64{0, 0, 0}, Dogleg: []float64{1, 0, 0}}
	var err error
	b := func() bool { v, _ := r.int(r.peek(), 0); return v != 0 }
	for !r.done() && err == nil {
		switch r.peek() {
		case "303":
			r.next()
			return l, nil
		case "290":
			l.HasLastLine = b()
		case "291":
			l.HasDogleg = b()
		case "10", "20", "30":
			err = setPoint(r.next(), l.LastPoint)
		case "11", "21", "31":
			err = setPoint(r.next(), l.Dogleg)
		case "12", "13":
			l.Breaks = append(l.Breaks, []float64{0, 0, 0})
			err = setPoint(r.next(), l.Breaks[len(l.Breaks)-1])
		case "22", "32", "23", "33":
			if len(l.Breaks) > 0 {
				err = setPoint(r.next(), l.Breaks[len(l.Breaks)-1])
			} else {
				r.next()
			}
		case "90":
			l.Index, err = r.int("90", 0)
		case "40":
			l.DoglegLength, err = r.float("40", 0)
		case "304":
			r.next()
			var ln *entity.MLeaderLine
			if ln, err = parseMLeaderLine(r); err == nil {
				l.Lines = append(l.Lines, ln)
			}
		case "271":
			l.AttachmentDirection, err = r.int("271", 0)
		default:
			r.next()
		}
	}
	return l, err
}

// parseMLeaderLine parses LEADER_LINE{ ... } after 304 until 305.
func parseMLeaderLine(r *groupReader) (*entity.MLeaderLine, error) {
	ln := &entity.MLeaderLine{}
	var err error
	for !r.done() && err == nil {
		switch r.peek() {
		case "305":
			r.next()
			return ln, nil
		case "10":
			ln.Vertices = append(ln.Vertices, []float64{0, 0, 0})
			err = setPoint(r.next(), ln.Vertices[len(ln.Vertices)-1])
		case "20", "30":
			if len(ln.Vertices) > 0 {
				err = setPoint(r.next(), ln.Vertices[len(ln.Vertices)-1])
			} else {
				r.next()
			}
		case "91":
			ln.Index, err = r.int("91", 0)
		case "92":
			ln.Color, err = r.int("92", 0)
		default:
			ln.Breaks = append(ln.Breaks, r.next())
		}
	}
	return ln, err
}
