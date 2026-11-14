// Package codepage converts DXF text between Go strings (UTF-8) and the
// encodings used in DXF files.
//
// Up to AutoCAD 2004 (AC1018), text in DXF files is stored in the code page
// given by the header variable $DWGCODEPAGE (e.g. ANSI_932 for Japanese
// Shift_JIS). Characters outside that code page are written as \U+XXXX
// (Unicode) or \M+nXXXX (multibyte, n = code page index) escapes.
// From AutoCAD 2007 (AC1021), DXF files are UTF-8.
package codepage

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

// Default is the default code page of DXF files.
const Default = "ANSI_1252"

// Japanese is the code page for Japanese (Shift_JIS).
const Japanese = "ANSI_932"

var encodings = map[string]encoding.Encoding{
	"ANSI_874":  charmap.Windows874,
	"ANSI_932":  japanese.ShiftJIS,
	"ANSI_936":  simplifiedchinese.GBK,
	"ANSI_949":  korean.EUCKR,
	"ANSI_950":  traditionalchinese.Big5,
	"ANSI_1250": charmap.Windows1250,
	"ANSI_1251": charmap.Windows1251,
	"ANSI_1252": charmap.Windows1252,
	"ANSI_1253": charmap.Windows1253,
	"ANSI_1254": charmap.Windows1254,
	"ANSI_1255": charmap.Windows1255,
	"ANSI_1256": charmap.Windows1256,
	"ANSI_1257": charmap.Windows1257,
	"ANSI_1258": charmap.Windows1258,
}

// code pages used by \M+n escapes (MIF).
var mifEncodings = map[byte]encoding.Encoding{
	'1': japanese.ShiftJIS,
	'2': traditionalchinese.Big5,
	'3': korean.EUCKR,
	'5': simplifiedchinese.GBK,
}

// Lookup returns the encoding for a $DWGCODEPAGE value (case-insensitive).
func Lookup(name string) (encoding.Encoding, bool) {
	e, ok := encodings[strings.ToUpper(strings.TrimSpace(name))]
	return e, ok
}

// IsUnicodeVersion reports whether DXF files of the $ACADVER version are UTF-8 (AC1021 or later).
func IsUnicodeVersion(version string) bool {
	version = strings.ToUpper(strings.TrimSpace(version))
	if !strings.HasPrefix(version, "AC") {
		return false
	}
	n, err := strconv.Atoi(version[2:])
	return err == nil && n >= 1021
}

// Encode converts a UTF-8 string to the code page name.
// Characters that cannot be represented in the code page (or all non-ASCII
// characters if the code page is unknown) are written as \U+XXXX escapes,
// so that the output can be read regardless of the reader's locale.
func Encode(s, name string) string {
	if isASCII(s) {
		return s
	}
	enc, ok := Lookup(name)
	var encoder *encoding.Encoder
	if ok {
		encoder = enc.NewEncoder()
	}
	var b strings.Builder
	for _, r := range s {
		if r < utf8.RuneSelf {
			b.WriteRune(r)
			continue
		}
		if encoder != nil {
			if bs, err := encoder.String(string(r)); err == nil {
				b.WriteString(bs)
				continue
			}
		}
		writeEscape(&b, r)
	}
	return b.String()
}

func writeEscape(b *strings.Builder, r rune) {
	if r > 0xFFFF {
		r1, r2 := utf16.EncodeRune(r)
		fmt.Fprintf(b, "\\U+%04X\\U+%04X", r1, r2)
		return
	}
	fmt.Fprintf(b, "\\U+%04X", r)
}

// Decode converts a string stored in the code page name to UTF-8.
// If name is empty, unknown or ANSI_1252, s is returned as is when it is
// valid UTF-8 (some applications write UTF-8 text with the default
// $DWGCODEPAGE ANSI_1252), otherwise it is decoded as ANSI_1252.
// Escapes (\U+XXXX, \M+nXXXX) are not converted; use Unescape.
func Decode(s, name string) string {
	if isASCII(s) {
		return s
	}
	enc, ok := Lookup(name)
	if !ok || enc == charmap.Windows1252 {
		if utf8.ValidString(s) {
			return s
		}
		enc = charmap.Windows1252
	}
	d, err := enc.NewDecoder().String(s)
	if err != nil {
		return s
	}
	return d
}

// Unescape converts \U+XXXX and \M+nXXXX escapes to UTF-8 characters.
// UTF-16 surrogate pairs written as two \U+ escapes are combined.
// Unknown or malformed escapes are left as they are.
func Unescape(s string) string {
	if !strings.Contains(s, "\\U+") && !strings.Contains(s, "\\M+") &&
		!strings.Contains(s, "\\u+") && !strings.Contains(s, "\\m+") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if r, n := unicodeEscape(s[i:]); n > 0 {
			if utf16.IsSurrogate(r) {
				if r2, n2 := unicodeEscape(s[i+n:]); n2 > 0 {
					if c := utf16.DecodeRune(r, r2); c != utf8.RuneError {
						b.WriteRune(c)
						i += n + n2
						continue
					}
				}
			} else {
				b.WriteRune(r)
				i += n
				continue
			}
		}
		if str, n := mifEscape(s[i:]); n > 0 {
			b.WriteString(str)
			i += n
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// unicodeEscape parses \U+XXXX at the beginning of s.
func unicodeEscape(s string) (rune, int) {
	if len(s) < 7 || s[0] != '\\' || (s[1] != 'U' && s[1] != 'u') || s[2] != '+' {
		return 0, 0
	}
	v, err := strconv.ParseUint(s[3:7], 16, 32)
	if err != nil {
		return 0, 0
	}
	return rune(v), 7
}

// mifEscape parses \M+nXXXX at the beginning of s.
func mifEscape(s string) (string, int) {
	if len(s) < 8 || s[0] != '\\' || (s[1] != 'M' && s[1] != 'm') || s[2] != '+' {
		return "", 0
	}
	enc, ok := mifEncodings[s[3]]
	if !ok {
		return "", 0
	}
	v, err := strconv.ParseUint(s[4:8], 16, 16)
	if err != nil {
		return "", 0
	}
	var raw []byte
	if v > 0xFF {
		raw = []byte{byte(v >> 8), byte(v)}
	} else {
		raw = []byte{byte(v)}
	}
	d, err := enc.NewDecoder().Bytes(raw)
	if err != nil || !utf8.Valid(d) || strings.ContainsRune(string(d), utf8.RuneError) {
		return "", 0
	}
	return string(d), 8
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
