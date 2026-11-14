package dxf

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/yofu/dxf/format"
)

// pairSource reads group code/value pairs of an ASCII or a binary DXF file.
type pairSource interface {
	// next returns the next group code and value (as text; numbers of
	// binary files are formatted). It returns io.EOF at the end.
	next() (code, value string, err error)
	// line returns the position for error messages: the line of the value
	// (ASCII), or twice the number of pairs read (binary).
	line() int
}

// newPairSource returns a pairSource for r, detecting binary DXF by its sentinel.
func newPairSource(r io.Reader) pairSource {
	br := bufio.NewReader(r)
	head, _ := br.Peek(len(format.BinarySentinel) + 2)
	if len(head) >= len(format.BinarySentinel) && string(head[:len(format.BinarySentinel)]) == format.BinarySentinel {
		br.Discard(len(format.BinarySentinel))
		// R13 or later: 2-byte group codes ("0" is 00 00 before "SECTION")
		twoByte := len(head) < len(format.BinarySentinel)+2 || head[len(format.BinarySentinel)+1] == 0
		return &binarySource{r: br, twoByte: twoByte}
	}
	s := bufio.NewScanner(br)
	s.Buffer(make([]byte, 64*1024), 16*1024*1024)
	return &textSource{s: s}
}

// textSource reads ASCII DXF.
type textSource struct {
	s *bufio.Scanner
	n int
}

func (t *textSource) next() (string, string, error) {
	if !t.s.Scan() {
		if err := t.s.Err(); err != nil {
			return "", "", err
		}
		return "", "", io.EOF
	}
	t.n++
	code := strings.TrimSpace(t.s.Text())
	if t.n == 1 {
		code = strings.TrimPrefix(code, "\ufeff") // UTF-8 BOM
	}
	if !t.s.Scan() {
		if err := t.s.Err(); err != nil {
			return "", "", err
		}
		return "", "", io.EOF
	}
	t.n++
	return code, t.s.Text(), nil
}

func (t *textSource) line() int { return t.n }

// binarySource reads binary DXF.
type binarySource struct {
	r       *bufio.Reader
	twoByte bool
	n       int
}

func (b *binarySource) code() (int, error) {
	if b.twoByte {
		var c int16
		if err := binary.Read(b.r, binary.LittleEndian, &c); err != nil {
			return 0, err
		}
		return int(uint16(c)), nil
	}
	c, err := b.r.ReadByte()
	if err != nil {
		return 0, err
	}
	if c == 255 { // R12: extended group code
		var x int16
		if err := binary.Read(b.r, binary.LittleEndian, &x); err != nil {
			return 0, err
		}
		return int(x), nil
	}
	return int(c), nil
}

func (b *binarySource) next() (string, string, error) {
	c, err := b.code()
	if err == io.EOF {
		return "", "", io.EOF
	}
	if err != nil {
		return "", "", fmt.Errorf("binary dxf: %w", err)
	}
	b.n += 2
	unexpected := func(err error) error {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return fmt.Errorf("binary dxf: group code %d: %w", c, err)
	}
	var v string
	switch format.GroupCodeType(c) {
	case format.TypeDouble:
		var x float64
		if err := binary.Read(b.r, binary.LittleEndian, &x); err != nil {
			return "", "", unexpected(err)
		}
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return "", "", unexpected(fmt.Errorf("invalid double %v", x))
		}
		v = strconv.FormatFloat(x, 'g', -1, 64)
	case format.TypeInt16:
		var x int16
		if err := binary.Read(b.r, binary.LittleEndian, &x); err != nil {
			return "", "", unexpected(err)
		}
		v = strconv.Itoa(int(x))
	case format.TypeInt32:
		var x int32
		if err := binary.Read(b.r, binary.LittleEndian, &x); err != nil {
			return "", "", unexpected(err)
		}
		v = strconv.Itoa(int(x))
	case format.TypeInt64:
		var x int64
		if err := binary.Read(b.r, binary.LittleEndian, &x); err != nil {
			return "", "", unexpected(err)
		}
		v = strconv.FormatInt(x, 10)
	case format.TypeBool:
		x, err := b.r.ReadByte()
		if err != nil {
			return "", "", unexpected(err)
		}
		v = strconv.Itoa(int(x))
	case format.TypeBinary:
		n, err := b.r.ReadByte()
		if err != nil {
			return "", "", unexpected(err)
		}
		data := make([]byte, n)
		if _, err := io.ReadFull(b.r, data); err != nil {
			return "", "", unexpected(err)
		}
		v = strings.ToUpper(hex.EncodeToString(data))
	default:
		s, err := b.r.ReadBytes(0)
		if err != nil {
			return "", "", unexpected(err)
		}
		v = string(bytes.TrimSuffix(s, []byte{0}))
	}
	return strconv.Itoa(c), v, nil
}

func (b *binarySource) line() int { return b.n }
