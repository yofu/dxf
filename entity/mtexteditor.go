package entity

import (
	"strconv"
	"strings"
)

// MTextEditor builds MTEXT content with formatting codes (as ezdxf's
// MTextEditor): text, paragraphs, fonts, heights, colors, stacking and
// so on. String returns the content for MText.Value.
type MTextEditor struct {
	sb strings.Builder
}

// NewMTextEditor returns an editor with text.
func NewMTextEditor(text string) *MTextEditor {
	e := &MTextEditor{}
	return e.Append(text)
}

// String returns the content.
func (e *MTextEditor) String() string { return e.sb.String() }

func ftoa(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// escapeMText escapes the characters with a meaning in MTEXT ("\", "{",
// "}") and converts "\n" to paragraph breaks.
func escapeMText(s string) string {
	r := strings.NewReplacer(`\`, `\\`, "{", `\{`, "}", `\}`, "\n", `\P`)
	return r.Replace(s)
}

// Append appends text (escaped; "\n" is a paragraph break).
func (e *MTextEditor) Append(text string) *MTextEditor {
	e.sb.WriteString(escapeMText(text))
	return e
}

// Raw appends content with formatting codes as is.
func (e *MTextEditor) Raw(content string) *MTextEditor {
	e.sb.WriteString(content)
	return e
}

// Paragraph appends a paragraph break (\P).
func (e *MTextEditor) Paragraph() *MTextEditor { return e.Raw(`\P`) }

// LineBreak appends a line break without a new paragraph (\X is not used:
// MTEXT has no line break; a paragraph break is written).
func (e *MTextEditor) LineBreak() *MTextEditor { return e.Paragraph() }

// ColumnBreak appends a column break (\N).
func (e *MTextEditor) ColumnBreak() *MTextEditor { return e.Raw(`\N`) }

// group appends {code text}: the formatting applies to text only.
func (e *MTextEditor) group(code, text string) *MTextEditor {
	return e.Raw("{" + code + escapeMText(text) + "}")
}

// Font appends text in font name (a TrueType family, e.g. "Arial"), bold
// and italic.
func (e *MTextEditor) Font(name string, bold, italic bool, text string) *MTextEditor {
	b, i := "0", "0"
	if bold {
		b = "1"
	}
	if italic {
		i = "1"
	}
	return e.group(`\f`+name+"|b"+b+"|i"+i+";", text)
}

// Height appends text of height (absolute drawing units).
func (e *MTextEditor) Height(h float64, text string) *MTextEditor {
	return e.group(`\H`+ftoa(h)+";", text)
}

// ScaleHeight appends text with height factor k of the current height.
func (e *MTextEditor) ScaleHeight(k float64, text string) *MTextEditor {
	return e.group(`\H`+ftoa(k)+"x;", text)
}

// Width appends text with width factor w.
func (e *MTextEditor) Width(w float64, text string) *MTextEditor {
	return e.group(`\W`+ftoa(w)+";", text)
}

// Oblique appends text with oblique angle (degrees).
func (e *MTextEditor) Oblique(angle float64, text string) *MTextEditor {
	return e.group(`\Q`+ftoa(angle)+";", text)
}

// Tracking appends text with character spacing factor k.
func (e *MTextEditor) Tracking(k float64, text string) *MTextEditor {
	return e.group(`\T`+ftoa(k)+";", text)
}

// Color appends text of color number aci.
func (e *MTextEditor) Color(aci int, text string) *MTextEditor {
	return e.group(`\C`+strconv.Itoa(aci)+";", text)
}

// TrueColor appends text of an RGB color.
func (e *MTextEditor) TrueColor(r, g, b uint8, text string) *MTextEditor {
	return e.group(`\c`+strconv.Itoa(int(b)<<16|int(g)<<8|int(r))+";", text)
}

// Underline appends underlined text.
func (e *MTextEditor) Underline(text string) *MTextEditor {
	return e.Raw(`\L` + escapeMText(text) + `\l`)
}

// Overline appends overlined text.
func (e *MTextEditor) Overline(text string) *MTextEditor {
	return e.Raw(`\O` + escapeMText(text) + `\o`)
}

// Strike appends struck through text.
func (e *MTextEditor) Strike(text string) *MTextEditor { return e.Raw(`\K` + escapeMText(text) + `\k`) }

// Stack kinds of Stack.
const (
	STACK_HORIZONTAL = "/" // horizontal fraction bar
	STACK_TOLERANCE  = "^" // tolerance (no bar)
	STACK_DIAGONAL   = "#" // diagonal fraction bar
)

// Stack appends stacked text: upper over lower with kind (STACK_*).
func (e *MTextEditor) Stack(upper, lower, kind string) *MTextEditor {
	esc := func(s string) string {
		return strings.NewReplacer("/", `\/`, "^", `\^`, "#", `\#`, ";", `\;`).Replace(escapeMText(s))
	}
	return e.Raw(`\S` + esc(upper) + kind + esc(lower) + ";")
}

// Alignment of MTEXT lines (\A): bottom 0, center 1, top 2.
func (e *MTextEditor) Alignment(a int) *MTextEditor { return e.Raw(`\A` + strconv.Itoa(a) + ";") }

// Paragraph properties (\p): indent of the first line, left and right
// indents, alignment ("l", "r", "c", "j", "d"; "" for none) and tab
// stops.
func (e *MTextEditor) ParagraphProperties(indent, left, right float64, align string, tabs ...float64) *MTextEditor {
	parts := []string{"i" + ftoa(indent), "l" + ftoa(left), "r" + ftoa(right)}
	if align != "" {
		parts = append(parts, "q"+align)
	}
	if len(tabs) > 0 {
		var ts []string
		for _, t := range tabs {
			ts = append(ts, ftoa(t))
		}
		parts = append(parts, "t"+strings.Join(ts, ","))
	}
	return e.Raw(`\p` + strings.Join(parts, ",") + ";")
}

// Bullets appends a bullet list: items as paragraphs with bullet (e.g.
// "•", or numbers when bullet is "1."), indented by indent.
func (e *MTextEditor) Bullets(bullet string, indent float64, items ...string) *MTextEditor {
	e.Raw("{")
	e.ParagraphProperties(-indent, indent, 0, "", indent)
	for i, it := range items {
		b := bullet
		if bullet == "1." {
			b = strconv.Itoa(i+1) + "."
		}
		e.Append(b).Raw("^I").Append(it)
		if i < len(items)-1 {
			e.Paragraph()
		}
	}
	return e.Raw("}")
}
