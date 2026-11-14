package entity

import (
	"math"
	"strings"
	"unicode/utf8"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/table"
)

// Attachment points of MTEXT (code 71).
const (
	MTEXT_TOP_LEFT = iota + 1
	MTEXT_TOP_CENTER
	MTEXT_TOP_RIGHT
	MTEXT_MIDDLE_LEFT
	MTEXT_MIDDLE_CENTER
	MTEXT_MIDDLE_RIGHT
	MTEXT_BOTTOM_LEFT
	MTEXT_BOTTOM_CENTER
	MTEXT_BOTTOM_RIGHT
)

// Drawing directions of MTEXT (code 72).
const (
	MTEXT_LEFT_TO_RIGHT = 1
	MTEXT_TOP_TO_BOTTOM = 3
	MTEXT_BY_STYLE      = 5
)

// Line spacing styles of MTEXT (code 73).
const (
	MTEXT_AT_LEAST = 1
	MTEXT_EXACT    = 2
)

// Background fill flags of MTEXT (code 90).
const (
	MTEXT_BG_COLOR      = 1  // fill with the background color (63, 421, 431)
	MTEXT_BG_CANVAS     = 2  // fill with the drawing window color
	MTEXT_BG_TEXT_FRAME = 16 // draw a text frame
)

// mtextChunk is the maximum length of a text chunk in group codes 1 and 3.
const mtextChunk = 250

// MText represents MTEXT Entity (multiline text).
//
// Value may contain MTEXT formatting codes such as \P (new paragraph);
// PlainText returns the text without them.
type MText struct {
	*entity
	Coord             []float64    // 10, 20, 30: insertion point
	Height            float64      // 40: nominal (initial) text height
	Width             float64      // 41: reference rectangle width (0: no wrapping)
	Attachment        int          // 71: attachment point (MTEXT_TOP_LEFT, ...)
	Direction         int          // 72: drawing direction (MTEXT_LEFT_TO_RIGHT, ...)
	Value             string       // 1, 3: text string with formatting codes
	Style             *table.Style // 7
	Rotation          float64      // rotation angle in degrees, written as X-axis direction (11, 21, 31)
	LineSpacingStyle  int          // 73: MTEXT_AT_LEAST or MTEXT_EXACT
	LineSpacingFactor float64      // 44: 0.25 to 4.00
	Extrusion         []float64    // 210, 220, 230

	// Background fill (AC1018 or later)
	BackgroundFill         int     // 90: MTEXT_BG_* flags (0: off)
	BackgroundScale        float64 // 45: fill box scale (border around the text, 1 to 5)
	BackgroundColor        int     // 63: color number
	BackgroundTrueColor    int     // 421: true color value (-1: not set)
	BackgroundColorName    string  // 431: color name
	BackgroundTransparency int     // 441: raw transparency value (0: not set)
}

// IsEntity is for Entity interface.
func (t *MText) IsEntity() bool {
	return true
}

// NewMText creates a new MText.
func NewMText() *MText {
	return &MText{
		entity:            NewEntity(MTEXT),
		Coord:             []float64{0.0, 0.0, 0.0},
		Height:            1.0,
		Attachment:        MTEXT_TOP_LEFT,
		Direction:         MTEXT_LEFT_TO_RIGHT,
		Style:             table.ST_STANDARD,
		LineSpacingStyle:  MTEXT_AT_LEAST,
		LineSpacingFactor: 1.0,
		Extrusion:         []float64{0.0, 0.0, 1.0},

		BackgroundScale:     1.5,
		BackgroundTrueColor: -1,
	}
}

// Format writes data to formatter.
func (t *MText) Format(f format.Formatter) {
	t.entity.Format(f)
	f.WriteString(100, "AcDbMText")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, t.Coord[i])
	}
	f.WriteFloat(40, t.Height)
	f.WriteFloat(41, t.Width)
	f.WriteInt(71, t.Attachment)
	f.WriteInt(72, t.Direction)
	// Text longer than 250 characters is divided into 250-character chunks
	// in group 3, and the rest in group 1 (DXF reference, MTEXT).
	chunks := splitText(t.Value, mtextChunk)
	for _, c := range chunks[:len(chunks)-1] {
		f.WriteString(3, c)
	}
	f.WriteString(1, chunks[len(chunks)-1])
	f.WriteString(7, t.Style.Name())
	if t.Extrusion[0] != 0.0 || t.Extrusion[1] != 0.0 || t.Extrusion[2] != 1.0 {
		for i := 0; i < 3; i++ {
			f.WriteFloat(200+(i+1)*10, t.Extrusion[i])
		}
	}
	if t.Rotation != 0.0 {
		rad := t.Rotation * math.Pi / 180.0
		f.WriteFloat(11, math.Cos(rad))
		f.WriteFloat(21, math.Sin(rad))
		f.WriteFloat(31, 0.0)
	}
	f.WriteInt(73, t.LineSpacingStyle)
	f.WriteFloat(44, t.LineSpacingFactor)
	if t.BackgroundFill != 0 && format.AtLeast(f, format.R2004) {
		if t.BackgroundFill&(MTEXT_BG_COLOR|MTEXT_BG_CANVAS) == 0 {
			// text frame only: 45 and 63 are not written
			f.WriteInt(90, t.BackgroundFill)
			return
		}
		f.WriteFloat(45, t.BackgroundScale)
		f.WriteInt(90, t.BackgroundFill)
		f.WriteInt(63, t.BackgroundColor)
		if t.BackgroundTrueColor >= 0 {
			f.WriteInt(421, t.BackgroundTrueColor)
		}
		if t.BackgroundColorName != "" {
			f.WriteString(431, t.BackgroundColorName)
		}
		if t.BackgroundTransparency != 0 {
			f.WriteInt(441, t.BackgroundTransparency)
		}
	}
}

// SetBackgroundColor fills the background with color number aci.
// scale is the fill box scale (border around the text based on the text
// height, 1 to 5; AutoCAD's default is 1.5).
func (t *MText) SetBackgroundColor(aci color.ColorNumber, scale float64) {
	t.setBackground(MTEXT_BG_COLOR, int(aci), -1, scale)
}

// SetBackgroundTrueColor fills the background with a true color.
// The nearest color number is also written (63).
func (t *MText) SetBackgroundTrueColor(r, g, b uint8, scale float64) {
	t.setBackground(MTEXT_BG_COLOR, int(color.Nearest(r, g, b)), color.RGB(r, g, b), scale)
}

// SetBackgroundCanvas fills the background with the drawing window color
// (a background mask; 90 = MTEXT_BG_COLOR|MTEXT_BG_CANVAS as written by
// AutoCAD for "Use drawing background color").
func (t *MText) SetBackgroundCanvas(scale float64) {
	t.setBackground(MTEXT_BG_COLOR|MTEXT_BG_CANVAS, 0, -1, scale)
}

// SetTextFrame sets or clears the text frame flag.
func (t *MText) SetTextFrame(frame bool) {
	if frame {
		t.BackgroundFill |= MTEXT_BG_TEXT_FRAME
	} else {
		t.BackgroundFill &^= MTEXT_BG_TEXT_FRAME
	}
}

// ClearBackground removes the background fill (the text frame is kept).
func (t *MText) ClearBackground() {
	t.BackgroundFill &= MTEXT_BG_TEXT_FRAME
	t.BackgroundColor = 0
	t.BackgroundTrueColor = -1
	t.BackgroundColorName = ""
	t.BackgroundTransparency = 0
	t.BackgroundScale = 1.5
}

func (t *MText) setBackground(flag, aci, rgb int, scale float64) {
	t.ClearBackground()
	t.BackgroundFill |= flag
	t.BackgroundColor = aci
	t.BackgroundTrueColor = rgb
	t.BackgroundScale = scale
}

// splitText splits s into chunks of at most n characters.
// The last chunk has fewer than n characters (possibly empty), so that it can
// be written in group 1. A backslash escape is not split from its character.
func splitText(s string, n int) []string {
	chunks := make([]string, 0, utf8.RuneCountInString(s)/n+1)
	for utf8.RuneCountInString(s) >= n {
		end, count := 0, 0
		for count < n {
			_, size := utf8.DecodeRuneInString(s[end:])
			end += size
			count++
		}
		// do not end a chunk with a single backslash
		if s[end-1] == '\\' && !(end >= 2 && s[end-2] == '\\') {
			end--
		}
		chunks = append(chunks, s[:end])
		s = s[end:]
	}
	return append(chunks, s)
}

// String outputs data using default formatter.
func (t *MText) String() string {
	f := format.NewASCII()
	return t.FormatString(f)
}

// FormatString outputs data using given formatter.
func (t *MText) FormatString(f format.Formatter) string {
	t.Format(f)
	return f.Output()
}

// PlainText returns the text without MTEXT formatting codes.
// Paragraph breaks (\P) are converted to "\n".
func (t *MText) PlainText() string {
	return MTextPlain(t.Value)
}

// MTextPlain removes MTEXT formatting codes from s.
// \P becomes "\n", \~ a space, stacked text (\S1^2;) "1/2",
// and \\, \{, \} the characters themselves.
func MTextPlain(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '{', '}':
			continue
		case '\\':
		default:
			b.WriteByte(c)
			continue
		}
		if i+1 >= len(s) {
			b.WriteByte(c)
			break
		}
		i++
		switch s[i] {
		case 'P':
			b.WriteByte('\n')
		case '~':
			b.WriteByte(' ')
		case '\\', '{', '}':
			b.WriteByte(s[i])
		case 'L', 'l', 'O', 'o', 'K', 'k', 'N':
			// underline, overline, strike-through, column break: no text
		case 'S':
			// stacked text: \S<upper>^<lower>; (also / and #)
			end := strings.IndexByte(s[i+1:], ';')
			if end < 0 {
				end = len(s) - i - 1
			}
			stack := s[i+1 : i+1+end]
			stack = strings.NewReplacer("^", "/", "#", "/").Replace(stack)
			b.WriteString(stack)
			i += end + 1
		case 'f', 'F', 'H', 'h', 'W', 'w', 'Q', 'q', 'T', 't', 'A', 'a', 'C', 'c', 'p':
			// codes with an argument terminated by ';'
			end := strings.IndexByte(s[i+1:], ';')
			if end < 0 {
				i = len(s)
			} else {
				i += end + 1
			}
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// BBox returns an approximate bounding box.
// The width is the reference rectangle width, or estimated from the
// longest line if it is 0; each line is assumed to be Height tall.
func (t *MText) BBox() ([]float64, []float64) {
	lines := strings.Split(t.PlainText(), "\n")
	w := t.Width
	if w <= 0 {
		for _, l := range lines {
			w = math.Max(w, float64(utf8.RuneCountInString(l))*t.Height)
		}
	}
	h := float64(len(lines)) * t.Height * t.LineSpacingFactor
	att := t.Attachment
	if att < 1 || att > 9 {
		att = MTEXT_TOP_LEFT
	}
	col := float64((att - 1) % 3) // 0: left, 1: center, 2: right
	row := float64((att - 1) / 3) // 0: top, 1: middle, 2: bottom
	x0 := -col * w / 2
	y1 := row * h / 2
	rad := t.Rotation * math.Pi / 180.0
	sin, cos := math.Sin(rad), math.Cos(rad)
	mins := []float64{math.Inf(1), math.Inf(1), t.Coord[2]}
	maxs := []float64{math.Inf(-1), math.Inf(-1), t.Coord[2]}
	for _, p := range [][2]float64{{x0, y1}, {x0 + w, y1}, {x0, y1 - h}, {x0 + w, y1 - h}} {
		x := t.Coord[0] + p[0]*cos - p[1]*sin
		y := t.Coord[1] + p[0]*sin + p[1]*cos
		mins[0], maxs[0] = math.Min(mins[0], x), math.Max(maxs[0], x)
		mins[1], maxs[1] = math.Min(mins[1], y), math.Max(maxs[1], y)
	}
	return mins, maxs
}
