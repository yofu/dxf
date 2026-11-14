package entity

import (
	"math"
	"strings"

	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Raw color values of MULTILEADER and MLEADERSTYLE (AcCmColor).
const (
	RawColorByLayer = -1073741824 // 0xC0000000
	RawColorByBlock = -1056964608 // 0xC1000000
)

// RawColor returns the raw color value of color number aci.
func RawColor(aci int) int {
	switch aci {
	case ColorByLayer:
		return RawColorByLayer
	case ColorByBlock:
		return RawColorByBlock
	}
	return -1023410176 + aci // 0xC3000000 | aci
}

// Content types of MULTILEADER (code 172) and MLEADERSTYLE (170).
const (
	MLEADER_NONE  = 0
	MLEADER_BLOCK = 1
	MLEADER_MTEXT = 2
)

// Ref is a reference by handle to an object: the object (written with
// its handle) or the handle as read.
type Ref struct {
	Object handle.Handler
	Raw    string
}

func (r Ref) write(f format.Formatter, code int, always bool) {
	switch {
	case r.Object != nil:
		f.WriteHex(code, r.Object.Handle())
	case r.Raw != "":
		f.WriteString(code, r.Raw)
	case always:
		f.WriteString(code, "0")
	}
}

// IsSet reports whether the reference is set.
func (r Ref) IsSet() bool { return r.Object != nil || (r.Raw != "" && r.Raw != "0") }

// MLeaderLine is a leader line of a MULTILEADER leader.
type MLeaderLine struct {
	Vertices [][]float64 // 10, 20, 30: from the arrow head
	Breaks   [][2]string // 90, 11, 12: kept as read
	Index    int         // 91
	Color    int         // 92: raw color value
}

// MLeaderLeader is a leader (LEADER{ ... }) of a MULTILEADER: leader lines
// connected to the content at one side.
type MLeaderLeader struct {
	HasLastLine         bool        // 290
	HasDogleg           bool        // 291
	LastPoint           []float64   // 10, 20, 30: connection point of the leader lines
	Dogleg              []float64   // 11, 21, 31: dogleg direction
	Breaks              [][]float64 // 12, 13: alternating start and end points
	Index               int         // 90
	DoglegLength        float64     // 40
	Lines               []*MLeaderLine
	AttachmentDirection int // 271 (AC1024 or later): 0 horizontal, 1 vertical
}

// MLeaderMText is the MTEXT content of a MULTILEADER.
type MLeaderMText struct {
	Content            string    // 304
	Extrusion          []float64 // 11
	Style              Ref       // 340: text style
	Insert             []float64 // 12
	Direction          []float64 // 13
	Rotation           float64   // 42 (radians)
	Width              float64   // 43
	DefinedHeight      float64   // 44
	LineSpacingFactor  float64   // 45
	LineSpacingStyle   int       // 170
	Color              int       // 90: raw color value
	Alignment          int       // 171: 1 left, 2 center, 3 right
	FlowDirection      int       // 172
	BgColor            int       // 91: raw color value
	BgScale            float64   // 141
	BgTransparency     int       // 92
	UseWindowBgColor   bool      // 291
	HasBgFill          bool      // 292
	ColumnType         int       // 173
	UseAutoHeight      bool      // 293
	ColumnWidth        float64   // 142
	ColumnGutterWidth  float64   // 143
	ColumnFlowReversed bool      // 294
	ColumnSizes        []float64 // 144
	UseWordBreak       bool      // 295
}

// MLeaderBlock is the block content of a MULTILEADER.
type MLeaderBlock struct {
	Block     Ref       // 341: BLOCK_RECORD
	Extrusion []float64 // 14
	Insert    []float64 // 15
	Scale     []float64 // 16
	Rotation  float64   // 46 (radians)
	Color     int       // 93: raw color value
	Matrix    []float64 // 47 (16 values)
}

// MLeaderContext is the context data (CONTEXT_DATA{ ... }) of a MULTILEADER.
type MLeaderContext struct {
	Scale            float64   // 40
	BasePoint        []float64 // 10
	CharHeight       float64   // 41 (scaled)
	ArrowSize        float64   // 140
	LandingGap       float64   // 145
	LeftAttachment   int       // 174
	RightAttachment  int       // 175
	TextAlign        int       // 176
	AttachmentType   int       // 177
	MText            *MLeaderMText
	Block            *MLeaderBlock
	PlaneOrigin      []float64 // 110
	PlaneX           []float64 // 111
	PlaneY           []float64 // 112
	NormalReversed   bool      // 297
	Leaders          []*MLeaderLeader
	TopAttachment    int // 272 (AC1024 or later)
	BottomAttachment int // 273 (AC1024 or later)
}

// MLeaderAttrib is an attribute value of the block content.
type MLeaderAttrib struct {
	AttDef Ref     // 330: ATTDEF of the block
	Index  int     // 177
	Width  float64 // 44
	Text   string  // 302
}

// MultiLeader represents MULTILEADER (MLEADER) entities.
type MultiLeader struct {
	*entity
	Version               int // 270
	Context               *MLeaderContext
	Style                 Ref     // 340: MLEADERSTYLE
	OverrideFlags         int     // 90
	LeaderType            int     // 170: 0 invisible, 1 straight, 2 spline
	LeaderColor           int     // 91: raw color value
	LeaderLineType        Ref     // 341: LTYPE
	LeaderLineWeight      int     // 171
	HasLanding            bool    // 290
	HasDogleg             bool    // 291
	DoglegLength          float64 // 41
	Arrow                 Ref     // 342: BLOCK_RECORD of the arrow (none: closed filled)
	ArrowSize             float64 // 42
	ContentType           int     // 172: MLEADER_*
	TextStyle             Ref     // 343: STYLE
	TextLeftAttachment    int     // 173
	TextRightAttachment   int     // 95
	TextAngleType         int     // 174
	TextAlignmentType     int     // 175
	TextColor             int     // 92: raw color value
	HasTextFrame          bool    // 292
	Block                 Ref     // 344: BLOCK_RECORD of the block content
	BlockColor            int     // 93: raw color value
	BlockScale            []float64
	BlockRotation         float64 // 43 (radians)
	BlockConnectionType   int     // 176
	Annotative            bool    // 293
	ArrowHeads            []MLeaderArrowHead
	BlockAttribs          []MLeaderAttrib
	TextDirectionNegative bool    // 294
	TextIPEAlign          int     // 178
	TextAttachmentPoint   int     // 179
	OverallScale          float64 // 45
	TextAttachmentDir     int     // 271 (AC1024 or later)
	TextBottomAttachment  int     // 272 (AC1024 or later)
	TextTopAttachment     int     // 273 (AC1024 or later)
	ExtendToText          bool    // 295 (AC1027 or later)
}

// MLeaderArrowHead is an arrow head of a leader line (94, 345).
type MLeaderArrowHead struct {
	Index int
	Block Ref
}

// IsEntity is for Entity interface.
func (m *MultiLeader) IsEntity() bool { return true }

// NewMultiLeader creates a MULTILEADER with the defaults of AutoCAD.
func NewMultiLeader() *MultiLeader {
	return &MultiLeader{
		entity:               NewEntity(MULTILEADER),
		Version:              2,
		Context:              NewMLeaderContext(),
		OverrideFlags:        0,
		LeaderType:           1,
		LeaderColor:          RawColorByBlock,
		LeaderLineWeight:     -2,
		HasLanding:           true,
		HasDogleg:            true,
		DoglegLength:         8,
		ArrowSize:            4,
		ContentType:          MLEADER_MTEXT,
		TextLeftAttachment:   1,
		TextRightAttachment:  1,
		TextAngleType:        1,
		TextColor:            RawColorByBlock,
		BlockColor:           RawColorByBlock,
		BlockScale:           []float64{1, 1, 1},
		TextAttachmentPoint:  1,
		OverallScale:         1,
		TextBottomAttachment: 9,
		TextTopAttachment:    9,
	}
}

// NewMLeaderContext creates context data with the defaults of AutoCAD.
func NewMLeaderContext() *MLeaderContext {
	return &MLeaderContext{
		Scale: 1, BasePoint: []float64{0, 0, 0}, CharHeight: 4, ArrowSize: 4, LandingGap: 2,
		LeftAttachment: 1, RightAttachment: 1,
		PlaneOrigin: []float64{0, 0, 0}, PlaneX: []float64{1, 0, 0}, PlaneY: []float64{0, 1, 0},
		TopAttachment: 9, BottomAttachment: 9,
	}
}

func (c *MLeaderContext) format(f format.Formatter) {
	f.WriteString(300, "CONTEXT_DATA{")
	f.WriteFloat(40, c.Scale)
	writePoint3(f, 10, c.BasePoint)
	f.WriteFloat(41, c.CharHeight)
	f.WriteFloat(140, c.ArrowSize)
	f.WriteFloat(145, c.LandingGap)
	f.WriteInt(174, c.LeftAttachment)
	f.WriteInt(175, c.RightAttachment)
	f.WriteInt(176, c.TextAlign)
	f.WriteInt(177, c.AttachmentType)
	if t := c.MText; t != nil {
		f.WriteInt(290, 1)
		f.WriteString(304, t.Content)
		writePoint3(f, 11, t.Extrusion)
		t.Style.write(f, 340, true)
		writePoint3(f, 12, t.Insert)
		writePoint3(f, 13, t.Direction)
		f.WriteFloat(42, t.Rotation)
		f.WriteFloat(43, t.Width)
		f.WriteFloat(44, t.DefinedHeight)
		f.WriteFloat(45, t.LineSpacingFactor)
		f.WriteInt(170, t.LineSpacingStyle)
		f.WriteInt(90, t.Color)
		f.WriteInt(171, t.Alignment)
		f.WriteInt(172, t.FlowDirection)
		f.WriteInt(91, t.BgColor)
		f.WriteFloat(141, t.BgScale)
		f.WriteInt(92, t.BgTransparency)
		f.WriteInt(291, boolInt(t.UseWindowBgColor))
		f.WriteInt(292, boolInt(t.HasBgFill))
		f.WriteInt(173, t.ColumnType)
		f.WriteInt(293, boolInt(t.UseAutoHeight))
		f.WriteFloat(142, t.ColumnWidth)
		f.WriteFloat(143, t.ColumnGutterWidth)
		f.WriteInt(294, boolInt(t.ColumnFlowReversed))
		for _, s := range t.ColumnSizes {
			f.WriteFloat(144, s)
		}
		f.WriteInt(295, boolInt(t.UseWordBreak))
	} else {
		f.WriteInt(290, 0)
	}
	if b := c.Block; b != nil {
		f.WriteInt(296, 1)
		b.Block.write(f, 341, true)
		writePoint3(f, 14, b.Extrusion)
		writePoint3(f, 15, b.Insert)
		writePoint3(f, 16, b.Scale)
		f.WriteFloat(46, b.Rotation)
		f.WriteInt(93, b.Color)
		m := b.Matrix
		if len(m) != 16 {
			m = []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
		}
		for _, v := range m {
			f.WriteFloat(47, v)
		}
	} else {
		f.WriteInt(296, 0)
	}
	writePoint3(f, 110, c.PlaneOrigin)
	writePoint3(f, 111, c.PlaneX)
	writePoint3(f, 112, c.PlaneY)
	f.WriteInt(297, boolInt(c.NormalReversed))
	r2010 := format.AtLeast(f, format.R2010)
	for _, l := range c.Leaders {
		f.WriteString(302, "LEADER{")
		f.WriteInt(290, boolInt(l.HasLastLine))
		f.WriteInt(291, boolInt(l.HasDogleg))
		writePoint3(f, 10, l.LastPoint)
		writePoint3(f, 11, l.Dogleg)
		for i, b := range l.Breaks {
			writePoint3(f, 12+i%2, b)
		}
		f.WriteInt(90, l.Index)
		f.WriteFloat(40, l.DoglegLength)
		for _, ln := range l.Lines {
			f.WriteString(304, "LEADER_LINE{")
			for _, v := range ln.Vertices {
				writePoint3(f, 10, v)
			}
			for _, p := range ln.Breaks {
				writeRawPair(f, p)
			}
			f.WriteInt(91, ln.Index)
			f.WriteInt(92, ln.Color)
			f.WriteString(305, "}")
		}
		if r2010 {
			f.WriteInt(271, l.AttachmentDirection)
		}
		f.WriteString(303, "}")
	}
	if r2010 {
		f.WriteInt(272, c.TopAttachment)
		f.WriteInt(273, c.BottomAttachment)
	}
	f.WriteString(301, "}")
}

// Format writes data to formatter.
func (m *MultiLeader) Format(f format.Formatter) {
	m.entity.Format(f)
	f.WriteString(100, "AcDbMLeader")
	f.WriteInt(270, m.Version)
	ctx := m.Context
	if ctx == nil {
		ctx = NewMLeaderContext()
	}
	ctx.format(f)
	m.Style.write(f, 340, true)
	f.WriteInt(90, m.OverrideFlags)
	f.WriteInt(170, m.LeaderType)
	f.WriteInt(91, m.LeaderColor)
	m.LeaderLineType.write(f, 341, true)
	f.WriteInt(171, m.LeaderLineWeight)
	f.WriteInt(290, boolInt(m.HasLanding))
	f.WriteInt(291, boolInt(m.HasDogleg))
	f.WriteFloat(41, m.DoglegLength)
	m.Arrow.write(f, 342, false)
	f.WriteFloat(42, m.ArrowSize)
	f.WriteInt(172, m.ContentType)
	m.TextStyle.write(f, 343, true)
	f.WriteInt(173, m.TextLeftAttachment)
	f.WriteInt(95, m.TextRightAttachment)
	f.WriteInt(174, m.TextAngleType)
	f.WriteInt(175, m.TextAlignmentType)
	f.WriteInt(92, m.TextColor)
	f.WriteInt(292, boolInt(m.HasTextFrame))
	m.Block.write(f, 344, false)
	f.WriteInt(93, m.BlockColor)
	writePoint3(f, 10, m.BlockScale)
	f.WriteFloat(43, m.BlockRotation)
	f.WriteInt(176, m.BlockConnectionType)
	f.WriteInt(293, boolInt(m.Annotative))
	if format.AtLeast(f, format.R2007) {
		for _, a := range m.ArrowHeads {
			f.WriteInt(94, a.Index)
			a.Block.write(f, 345, true)
		}
		for _, a := range m.BlockAttribs {
			a.AttDef.write(f, 330, true)
			f.WriteInt(177, a.Index)
			f.WriteFloat(44, a.Width)
			f.WriteString(302, a.Text)
		}
		f.WriteInt(294, boolInt(m.TextDirectionNegative))
		f.WriteInt(178, m.TextIPEAlign)
		f.WriteInt(179, m.TextAttachmentPoint)
		f.WriteFloat(45, m.OverallScale)
	}
	if format.AtLeast(f, format.R2010) {
		f.WriteInt(271, m.TextAttachmentDir)
		f.WriteInt(272, m.TextBottomAttachment)
		f.WriteInt(273, m.TextTopAttachment)
	}
	if format.AtLeast(f, format.R2013) {
		f.WriteInt(295, boolInt(m.ExtendToText))
	}
}

// String outputs data using default formatter.
func (m *MultiLeader) String() string { return m.FormatString(format.NewASCII()) }

// FormatString outputs data using given formatter.
func (m *MultiLeader) FormatString(f format.Formatter) string {
	m.Format(f)
	return f.Output()
}

// points returns the points of the leaders and the content insertion points.
func (m *MultiLeader) points() [][]float64 {
	c := m.Context
	if c == nil {
		return nil
	}
	pts := [][]float64{c.BasePoint}
	if c.MText != nil {
		pts = append(pts, c.MText.Insert)
	}
	if c.Block != nil {
		pts = append(pts, c.Block.Insert)
	}
	for _, l := range c.Leaders {
		pts = append(pts, l.LastPoint)
		for _, ln := range l.Lines {
			pts = append(pts, ln.Vertices...)
		}
	}
	return pts
}

// BBox returns the bounding box of the leader points and the content
// insertion points (the content size is not known).
func (m *MultiLeader) BBox() ([]float64, []float64) { return pointsBBox(m.points()...) }

// ApplyTransform transforms the multileader (conformal transformations
// without mirror: the content would be mirrored).
func (m *MultiLeader) ApplyTransform(t Transform) error {
	if !t.Conformal() || t.Mirrored() {
		return ErrNonUniform
	}
	c := m.Context
	if c == nil {
		return nil
	}
	for _, p := range [][]float64{c.BasePoint, c.PlaneOrigin} {
		transformPoint(t, p)
	}
	c.PlaneX = normalize3(t.Vector(c.PlaneX))
	c.PlaneY = normalize3(t.Vector(c.PlaneY))
	k := t.Length(1)
	c.Scale *= k
	c.CharHeight *= k
	c.ArrowSize *= k
	c.LandingGap *= k
	m.DoglegLength *= k
	m.ArrowSize *= k
	rad := t.Angle * math.Pi / 180
	if x := c.MText; x != nil {
		transformPoint(t, x.Insert)
		x.Direction = normalize3(t.Vector(x.Direction))
		x.Rotation += rad
		x.Width *= k
		x.DefinedHeight *= k
	}
	if b := c.Block; b != nil {
		transformPoint(t, b.Insert)
		for i := range b.Scale {
			b.Scale[i] *= k
		}
		b.Rotation += rad
	}
	for _, l := range c.Leaders {
		transformPoint(t, l.LastPoint)
		l.Dogleg = normalize3(t.Vector(l.Dogleg))
		l.DoglegLength *= k
		for _, b := range l.Breaks {
			transformPoint(t, b)
		}
		for _, ln := range l.Lines {
			for _, v := range ln.Vertices {
				transformPoint(t, v)
			}
		}
	}
	return nil
}

// MLeaderStyle represents MLEADERSTYLE objects (dictionary
// ACAD_MLEADERSTYLE).
type MLeaderStyle struct {
	ContentType          int     // 170
	DrawMLeaderOrder     int     // 171
	DrawLeaderOrder      int     // 172
	MaxLeaderPoints      int     // 90
	FirstAngle           float64 // 40
	SecondAngle          float64 // 41
	LeaderType           int     // 173
	LeaderColor          int     // 91: raw color value
	LeaderLineType       Ref     // 340
	LeaderLineWeight     int     // 92
	HasLanding           bool    // 290
	LandingGap           float64 // 42
	HasDogleg            bool    // 291
	DoglegLength         float64 // 43
	Description          string  // 3
	Arrow                Ref     // 341
	ArrowSize            float64 // 44
	DefaultText          string  // 300
	TextStyle            Ref     // 342
	TextLeftAttachment   int     // 174
	TextAngleType        int     // 175
	TextAlignmentType    int     // 176
	TextRightAttachment  int     // 178
	TextColor            int     // 93: raw color value
	CharHeight           float64 // 45
	HasTextFrame         bool    // 292
	TextAlignAlwaysLeft  bool    // 297
	AlignSpace           float64 // 46
	HasBlockScaling      bool    // 293
	Block                Ref     // 343
	BlockColor           int     // 94: raw color value
	BlockScale           []float64
	HasBlockRotation     bool    // 294
	BlockRotation        float64 // 141
	BlockConnectionType  int     // 177
	Scale                float64 // 142
	OverwriteProperty    bool    // 295
	Annotative           bool    // 296
	BreakGapSize         float64 // 143
	TextAttachmentDir    int     // 271
	TextBottomAttachment int     // 272
	TextTopAttachment    int     // 273
	Unknown179           int     // 179

	name   string
	handle int
	src    int
	owner  handle.Handler
}

// NewMLeaderStyle creates an MLEADERSTYLE with the defaults of AutoCAD.
func NewMLeaderStyle(name string) *MLeaderStyle {
	return &MLeaderStyle{name: name, ContentType: MLEADER_MTEXT, DrawMLeaderOrder: 1, MaxLeaderPoints: 2,
		LeaderType: 1, LeaderColor: RawColorByBlock, LeaderLineWeight: -2, HasLanding: true, LandingGap: 2,
		HasDogleg: true, DoglegLength: 8, ArrowSize: 4, TextLeftAttachment: 1, TextAngleType: 1,
		TextRightAttachment: 1, TextColor: RawColorByBlock, CharHeight: 4, AlignSpace: 4, BlockColor: RawColorByBlock,
		BlockScale: []float64{1, 1, 1}, HasBlockRotation: true, Scale: 1, BreakGapSize: 3.75,
		TextBottomAttachment: 9, TextTopAttachment: 9, Unknown179: 2}
}

// Name returns the name of the style.
func (s *MLeaderStyle) Name() string { return s.name }

// IsObject is for object.Object interface.
func (s *MLeaderStyle) IsObject() bool { return true }

// Format writes data to formatter.
func (s *MLeaderStyle) Format(f format.Formatter) {
	f.WriteString(0, "MLEADERSTYLE")
	f.WriteHex(5, s.handle)
	writeOwner(f, s.owner, true)
	f.WriteString(100, "AcDbMLeaderStyle")
	f.WriteInt(179, s.Unknown179)
	f.WriteInt(170, s.ContentType)
	f.WriteInt(171, s.DrawMLeaderOrder)
	f.WriteInt(172, s.DrawLeaderOrder)
	f.WriteInt(90, s.MaxLeaderPoints)
	f.WriteFloat(40, s.FirstAngle)
	f.WriteFloat(41, s.SecondAngle)
	f.WriteInt(173, s.LeaderType)
	f.WriteInt(91, s.LeaderColor)
	s.LeaderLineType.write(f, 340, false)
	f.WriteInt(92, s.LeaderLineWeight)
	f.WriteInt(290, boolInt(s.HasLanding))
	f.WriteFloat(42, s.LandingGap)
	f.WriteInt(291, boolInt(s.HasDogleg))
	f.WriteFloat(43, s.DoglegLength)
	f.WriteString(3, s.Description)
	s.Arrow.write(f, 341, false)
	f.WriteFloat(44, s.ArrowSize)
	f.WriteString(300, s.DefaultText)
	s.TextStyle.write(f, 342, false)
	f.WriteInt(174, s.TextLeftAttachment)
	f.WriteInt(175, s.TextAngleType)
	f.WriteInt(176, s.TextAlignmentType)
	f.WriteInt(178, s.TextRightAttachment)
	f.WriteInt(93, s.TextColor)
	f.WriteFloat(45, s.CharHeight)
	f.WriteInt(292, boolInt(s.HasTextFrame))
	f.WriteInt(297, boolInt(s.TextAlignAlwaysLeft))
	f.WriteFloat(46, s.AlignSpace)
	f.WriteInt(293, boolInt(s.HasBlockScaling))
	s.Block.write(f, 343, false)
	f.WriteInt(94, s.BlockColor)
	f.WriteFloat(47, coord(s.BlockScale, 0))
	f.WriteFloat(49, coord(s.BlockScale, 1))
	f.WriteFloat(140, coord(s.BlockScale, 2))
	f.WriteInt(294, boolInt(s.HasBlockRotation))
	f.WriteFloat(141, s.BlockRotation)
	f.WriteInt(177, s.BlockConnectionType)
	f.WriteFloat(142, s.Scale)
	f.WriteInt(295, boolInt(s.OverwriteProperty))
	f.WriteInt(296, boolInt(s.Annotative))
	f.WriteFloat(143, s.BreakGapSize)
	if format.AtLeast(f, format.R2010) {
		f.WriteInt(271, s.TextAttachmentDir)
		f.WriteInt(272, s.TextBottomAttachment)
		f.WriteInt(273, s.TextTopAttachment)
	}
}

// Handle returns the handle.
func (s *MLeaderStyle) Handle() int { return s.handle }

// SetHandle sets the handle (the handle read is kept).
func (s *MLeaderStyle) SetHandle(v *int) { s.handle = handle.Next(s.src, v) }

// SetOwner sets the owner dictionary.
func (s *MLeaderStyle) SetOwner(h handle.Handler) { s.owner = h }

// SourceHandle returns the handle in the file it was read from ("" if new).
func (s *MLeaderStyle) SourceHandle() string { return handle.Format(s.src) }

// SetSourceHandle sets the handle in the file it was read from.
func (s *MLeaderStyle) SetSourceHandle(h string) { s.src = handle.Parse(h) }

// EstimateTextSize estimates the width and height of MTEXT content with
// char height: the width of a character is 0.7 of the height (1.0 for
// wide characters such as Japanese), the line spacing 5/3 of the height.
// Fonts are not measured.
func EstimateTextSize(content string, charHeight, lineSpacing float64) (float64, float64) {
	plain := (&MText{Value: content}).PlainText()
	lines := strings.Split(plain, "\n")
	width := 0.0
	for _, l := range lines {
		w := 0.0
		for _, r := range l {
			if r >= 0x1100 {
				w += 1.0
			} else {
				w += 0.7
			}
		}
		width = math.Max(width, w*charHeight)
	}
	if lineSpacing == 0 {
		lineSpacing = 1
	}
	height := charHeight + float64(len(lines)-1)*charHeight*5/3*lineSpacing
	return width, height
}
