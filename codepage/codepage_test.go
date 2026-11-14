package codepage

import "testing"

func TestEncode(t *testing.T) {
	cases := []struct {
		s, cp, want string
	}{
		{"abc", Default, "abc"},
		{"壁A", Default, `\U+58C1A`},
		{"café", Default, "caf\xe9"},
		{"壁", Japanese, "\x95\xc7"},
		{"壁😀", Japanese, "\x95\xc7" + `\U+D83D\U+DE00`},
		{"壁", "", `\U+58C1`},
		{"壁", "ansi_932", "\x95\xc7"},
	}
	for _, c := range cases {
		if got := Encode(c.s, c.cp); got != c.want {
			t.Errorf("Encode(%q, %q) = %q, want %q", c.s, c.cp, got, c.want)
		}
	}
}

func TestDecode(t *testing.T) {
	cases := []struct {
		s, cp, want string
	}{
		{"abc", Default, "abc"},
		{"\x95\xc7", Japanese, "壁"},
		{"caf\xe9", Default, "café"},
		{"壁", "", "壁"},          // valid UTF-8, no code page
		{"caf\xe9", "", "café"}, // invalid UTF-8, no code page: ANSI_1252
		{"日本語", Default, "日本語"}, // UTF-8 text with the default code page
	}
	for _, c := range cases {
		if got := Decode(c.s, c.cp); got != c.want {
			t.Errorf("Decode(%q, %q) = %q, want %q", c.s, c.cp, got, c.want)
		}
	}
}

func TestUnescape(t *testing.T) {
	cases := []struct {
		s, want string
	}{
		{`\U+58C1`, "壁"},
		{`a\u+58c1b`, "a壁b"},
		{`\U+D83D\U+DE00`, "😀"},
		{`\M+193FA`, "日"},
		{`\U+ZZZZ`, `\U+ZZZZ`},
		{`\M+9FFFF`, `\M+9FFFF`},
		{`C:\Users`, `C:\Users`},
		{`\U+58`, `\U+58`},
	}
	for _, c := range cases {
		if got := Unescape(c.s); got != c.want {
			t.Errorf("Unescape(%q) = %q, want %q", c.s, got, c.want)
		}
	}
}

func TestIsUnicodeVersion(t *testing.T) {
	for v, want := range map[string]bool{"AC1015": false, "AC1018": false, "AC1021": true, "AC1032": true, "": false} {
		if got := IsUnicodeVersion(v); got != want {
			t.Errorf("IsUnicodeVersion(%q) = %v", v, got)
		}
	}
}
